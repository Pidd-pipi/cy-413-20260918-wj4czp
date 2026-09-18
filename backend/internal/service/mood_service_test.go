package service

import (
	"errors"
	"github.com/blueship581/mindgarden/backend/internal/constants"
	"github.com/blueship581/mindgarden/backend/internal/dto"
	"github.com/blueship581/mindgarden/backend/internal/model"
	"github.com/blueship581/mindgarden/backend/internal/repository"
	"github.com/blueship581/mindgarden/backend/internal/util"
	"gorm.io/gorm"
	"io"
	"log/slog"
	"testing"
	"time"
)

type fakeMoodRepo struct {
	rows map[uint]*model.Mood
	next uint
	now  func() time.Time
}

func newFakeMoodRepo(now func() time.Time) *fakeMoodRepo {
	return &fakeMoodRepo{rows: map[uint]*model.Mood{}, now: now}
}
func (f *fakeMoodRepo) Create(v *model.Mood) error {
	f.next++
	v.ID = f.next
	f.rows[v.ID] = v
	return nil
}
func (f *fakeMoodRepo) List(uid uint, date *time.Time) (out []model.Mood, _ error) {
	for _, v := range f.rows {
		if v.UserID == uid && !v.DeletedAt.Valid {
			if date != nil && !sameDay(v.RecordDate, *date) {
				continue
			}
			out = append(out, *v)
		}
	}
	return
}
func sameDay(a, b time.Time) bool {
	return a.Truncate(24 * time.Hour).Equal(b.Truncate(24 * time.Hour))
}
func (f *fakeMoodRepo) ByID(id, uid uint) (*model.Mood, error) {
	v, ok := f.rows[id]
	if !ok || v.UserID != uid || v.DeletedAt.Valid {
		return nil, repository.ErrNotFound
	}
	return v, nil
}
func (f *fakeMoodRepo) ByIDAnyState(id, uid uint) (*model.Mood, error) {
	v, ok := f.rows[id]
	if !ok || v.UserID != uid {
		return nil, repository.ErrNotFound
	}
	return v, nil
}
func (f *fakeMoodRepo) ListTrash(uid uint) (out []model.Mood, _ error) {
	for _, v := range f.rows {
		if v.UserID == uid && v.DeletedAt.Valid {
			out = append(out, *v)
		}
	}
	return
}
func (f *fakeMoodRepo) Update(v *model.Mood) error { f.rows[v.ID] = v; return nil }
func (f *fakeMoodRepo) SoftDelete(v *model.Mood) error {
	v.DeletedAt = gorm.DeletedAt{Time: f.now(), Valid: true}
	return nil
}
func (f *fakeMoodRepo) Restore(v *model.Mood) error {
	v.DeletedAt = gorm.DeletedAt{}
	return nil
}
func (f *fakeMoodRepo) PurgeExpiredForUser(uid uint, cutoff time.Time) (int64, error) {
	var n int64
	for id, v := range f.rows {
		if v.UserID == uid && v.DeletedAt.Valid && v.DeletedAt.Time.Before(cutoff) {
			delete(f.rows, id)
			n++
		}
	}
	return n, nil
}
func (f *fakeMoodRepo) PurgeExpired(cutoff time.Time) (int64, error) {
	var n int64
	for id, v := range f.rows {
		if v.DeletedAt.Valid && v.DeletedAt.Time.Before(cutoff) {
			delete(f.rows, id)
			n++
		}
	}
	return n, nil
}

func newMoodServiceWithClock() (*MoodService, *fakeMoodRepo, *time.Time) {
	clock := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)
	repo := newFakeMoodRepo(func() time.Time { return clock })
	s := NewMoodService(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	s.now = func() time.Time { return clock }
	return s, repo, &clock
}

func seedMood(t *testing.T, s *MoodService, uid uint, date string) *model.Mood {
	t.Helper()
	v, e := s.Create(uid, dto.MoodRequest{MoodLevel: 7, MoodTags: []string{constants.MoodCalm}, Note: "保留我", RecordDate: date})
	if e != nil {
		t.Fatalf("seed mood: %v", e)
	}
	return v
}

func TestMoodDeleteRemovesFromListFilterAndTrend(t *testing.T) {
	s, repo, _ := newMoodServiceWithClock()
	m := seedMood(t, s, 1, "2026-09-17")

	if e := s.Delete(1, m.ID); e != nil {
		t.Fatalf("delete: %v", e)
	}
	if got, _ := s.List(1, ""); len(got) != 0 {
		t.Fatalf("active list must exclude removed records, got %d", len(got))
	}
	if got, _ := s.List(1, "2026-09-17"); len(got) != 0 {
		t.Fatalf("date-filtered list must exclude removed records, got %d", len(got))
	}
	trash, e := s.ListTrash(1)
	if e != nil || len(trash) != 1 {
		t.Fatalf("trash should hold the record, got %d err=%v", len(trash), e)
	}
	if repo.rows[m.ID].DeletedAt.Time.IsZero() {
		t.Fatal("deleted_at must be set")
	}
}

func TestMoodRepeatedDeleteKeepsSameResultAndWindow(t *testing.T) {
	s, repo, _ := newMoodServiceWithClock()
	m := seedMood(t, s, 1, "2026-09-17")

	if e := s.Delete(1, m.ID); e != nil {
		t.Fatalf("first delete: %v", e)
	}
	first := repo.rows[m.ID].DeletedAt.Time
	repo.now = func() time.Time { return first.Add(2 * 24 * time.Hour) }
	s.now = repo.now
	if e := s.Delete(1, m.ID); e != nil {
		t.Fatalf("repeated delete must stay successful: %v", e)
	}
	if got := repo.rows[m.ID].DeletedAt.Time; !got.Equal(first) {
		t.Fatalf("repeated delete must not extend the window: first=%v got=%v", first, got)
	}
	if got, _ := s.ListTrash(1); len(got) != 1 {
		t.Fatalf("trash must still contain exactly one record, got %d", len(got))
	}
}

func TestMoodRestorePreservesOriginalFields(t *testing.T) {
	s, _, _ := newMoodServiceWithClock()
	m := seedMood(t, s, 1, "2026-09-17")
	if e := s.Delete(1, m.ID); e != nil {
		t.Fatalf("delete: %v", e)
	}
	got, e := s.Restore(1, m.ID)
	if e != nil {
		t.Fatalf("restore: %v", e)
	}
	if got.RecordDate.Format("2006-01-02") != "2026-09-17" || got.MoodLevel != 7 || got.Note != "保留我" || got.MoodTags != `["calm"]` {
		t.Fatalf("restore must keep date/level/tags/note: %+v", got)
	}
	if got.DeletedAt.Valid {
		t.Fatal("restored record must clear deleted_at")
	}
	if list, _ := s.List(1, ""); len(list) != 1 {
		t.Fatalf("restored record must return to the active list, got %d", len(list))
	}
	if trash, _ := s.ListTrash(1); len(trash) != 0 {
		t.Fatalf("restored record must leave the trash, got %d", len(trash))
	}
}

func TestMoodRepeatedRestoreIsIdempotent(t *testing.T) {
	s, _, _ := newMoodServiceWithClock()
	m := seedMood(t, s, 1, "2026-09-17")
	if e := s.Delete(1, m.ID); e != nil {
		t.Fatalf("delete: %v", e)
	}
	if _, e := s.Restore(1, m.ID); e != nil {
		t.Fatalf("first restore: %v", e)
	}
	if got, e := s.Restore(1, m.ID); e != nil || got.DeletedAt.Valid {
		t.Fatalf("repeated restore must stay a successful no-op, got %+v err=%v", got, e)
	}
}

func TestMoodRestoreNonExistentReturnsNotFound(t *testing.T) {
	s, _, _ := newMoodServiceWithClock()
	_, e := s.Restore(1, 999)
	if !errors.Is(e, repository.ErrNotFound) {
		t.Fatalf("missing record must map to ErrNotFound, got %v", e)
	}
}

func TestMoodExpiredRecordsArePurgedAndNeverReturn(t *testing.T) {
	s, repo, clock := newMoodServiceWithClock()
	m := seedMood(t, s, 1, "2026-09-01")
	if e := s.Delete(1, m.ID); e != nil {
		t.Fatalf("delete: %v", e)
	}
	// Advance beyond the seven-day retention window.
	*clock = clock.Add(constants.MoodTrashRetention + time.Hour)

	if trash, e := s.ListTrash(1); e != nil || len(trash) != 0 {
		t.Fatalf("expired record must not appear in trash, got %d err=%v", len(trash), e)
	}
	if _, ok := repo.rows[m.ID]; ok {
		t.Fatal("expired record must be hard-deleted")
	}
	if list, _ := s.List(1, ""); len(list) != 0 {
		t.Fatalf("expired record must not return to the active list, got %d", len(list))
	}
	if _, e := s.Restore(1, m.ID); !errors.Is(e, repository.ErrNotFound) {
		t.Fatalf("expired restore must report not found, got %v", e)
	}
	if e := s.Delete(1, m.ID); !errors.Is(e, repository.ErrNotFound) {
		t.Fatalf("expired delete must report not found, got %v", e)
	}
}

func TestMoodTrashIsScopedToCurrentUser(t *testing.T) {
	s, _, _ := newMoodServiceWithClock()
	a := seedMood(t, s, 1, "2026-09-17")
	_ = seedMood(t, s, 2, "2026-09-17")
	if e := s.Delete(1, a.ID); e != nil {
		t.Fatalf("delete: %v", e)
	}
	if trash, _ := s.ListTrash(2); len(trash) != 0 {
		t.Fatalf("user 2 must not see user 1 trashed records, got %d", len(trash))
	}
	if _, e := s.Restore(2, a.ID); !errors.Is(e, repository.ErrNotFound) {
		t.Fatalf("user 2 must not restore user 1 record, got %v", e)
	}
}

func TestMoodDeleteMissingReturnsWrappedNotFound(t *testing.T) {
	s, _, _ := newMoodServiceWithClock()
	e := s.Delete(1, 4242)
	if !errors.Is(e, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", e)
	}
	var app *util.AppError
	if errors.As(e, &app) {
		t.Fatal("plain delete-not-found should stay the sentinel error")
	}
}
