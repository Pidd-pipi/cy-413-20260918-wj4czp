package service

import (
	"errors"
	"github.com/blueship581/mindgarden/backend/internal/constants"
	"github.com/blueship581/mindgarden/backend/internal/dto"
	"github.com/blueship581/mindgarden/backend/internal/model"
	"github.com/blueship581/mindgarden/backend/internal/repository"
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
	v.DeletedAt = gorm.DeletedAt{}
	cp := *v
	f.rows[v.ID] = &cp
	return nil
}
func (f *fakeMoodRepo) List(uid uint, date *time.Time) ([]model.Mood, error) {
	out := []model.Mood{}
	for _, v := range f.rows {
		if v.UserID != uid || v.DeletedAt.Valid {
			continue
		}
		if date != nil {
			start := date.Truncate(24 * time.Hour)
			if v.RecordDate.Before(start) || !v.RecordDate.Before(start.AddDate(0, 0, 1)) {
				continue
			}
		}
		out = append(out, *v)
	}
	return out, nil
}
func (f *fakeMoodRepo) ByID(id, uid uint) (*model.Mood, error) {
	v, ok := f.rows[id]
	if !ok || v.UserID != uid || v.DeletedAt.Valid {
		return nil, repository.ErrNotFound
	}
	cp := *v
	return &cp, nil
}
func (f *fakeMoodRepo) Update(v *model.Mood) error {
	if _, ok := f.rows[v.ID]; !ok {
		return repository.ErrNotFound
	}
	cp := *v
	f.rows[v.ID] = &cp
	return nil
}
func (f *fakeMoodRepo) Delete(v *model.Mood) error {
	r, ok := f.rows[v.ID]
	if !ok {
		return repository.ErrNotFound
	}
	r.DeletedAt = gorm.DeletedAt{Time: f.now(), Valid: true}
	return nil
}
func (f *fakeMoodRepo) TrashByID(id, uid uint) (*model.Mood, error) {
	v, ok := f.rows[id]
	if !ok || v.UserID != uid || !v.DeletedAt.Valid {
		return nil, repository.ErrNotFound
	}
	cp := *v
	return &cp, nil
}
func (f *fakeMoodRepo) TrashList(uid uint) ([]model.Mood, error) {
	out := []model.Mood{}
	for _, v := range f.rows {
		if v.UserID == uid && v.DeletedAt.Valid {
			out = append(out, *v)
		}
	}
	return out, nil
}
func (f *fakeMoodRepo) Restore(v *model.Mood) error {
	r, ok := f.rows[v.ID]
	if !ok || !r.DeletedAt.Valid {
		return repository.ErrNotFound
	}
	r.DeletedAt = gorm.DeletedAt{}
	return nil
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
	now := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)
	repo := newFakeMoodRepo(func() time.Time { return now })
	s := NewMoodService(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	s.now = repo.now
	return s, repo, &now
}

func sampleMoodRequest(note string) dto.MoodRequest {
	return dto.MoodRequest{MoodLevel: 8, MoodTags: []string{constants.MoodHappy, constants.MoodCalm}, Note: note, RecordDate: "2026-09-17"}
}

// 移除立即退出列表/筛选；七天内出现在回收站；还原后字段不变。
func TestMoodTrashMoveListRestoreLifecycle(t *testing.T) {
	s, _, _ := newMoodServiceWithClock()
	const uid uint = 1
	v, e := s.Create(uid, sampleMoodRequest("温柔地看见自己"))
	if e != nil {
		t.Fatalf("create: %v", e)
	}
	if got, _ := s.List(uid, ""); len(got) != 1 {
		t.Fatalf("active list before delete = %d, want 1", len(got))
	}
	if e := s.Delete(uid, v.ID); e != nil {
		t.Fatalf("delete: %v", e)
	}
	if got, _ := s.List(uid, ""); len(got) != 0 {
		t.Fatalf("active list after delete = %d, want 0 (record must leave list immediately)", len(got))
	}
	if got, _ := s.List(uid, "2026-09-17"); len(got) != 0 {
		t.Fatalf("date-filtered list after delete = %d, want 0", len(got))
	}
	trash, e := s.TrashList(uid)
	if e != nil || len(trash) != 1 {
		t.Fatalf("trash list = %d items, err=%v; want 1", len(trash), e)
	}
	it := trash[0]
	if it.MoodLevel != 8 || it.MoodTags != `["happy","calm"]` || it.Note != "温柔地看见自己" {
		t.Fatalf("trash item fields altered: %+v", it)
	}
	if it.RecordDate.Format("2006-01-02") != "2026-09-17" {
		t.Fatalf("record_date altered: %s", it.RecordDate)
	}
	if !it.ExpiresAt.Equal(it.DeletedAt.Time.Add(constants.MoodTrashRetention)) {
		t.Fatalf("expires_at not 7 days after deleted_at: %v vs %v", it.ExpiresAt, it.DeletedAt.Time)
	}
	restored, e := s.Restore(uid, v.ID)
	if e != nil {
		t.Fatalf("restore: %v", e)
	}
	if restored.RecordDate.Format("2006-01-02") != "2026-09-17" || restored.MoodLevel != 8 || restored.MoodTags != `["happy","calm"]` || restored.Note != "温柔地看见自己" {
		t.Fatalf("restored fields altered: %+v", restored)
	}
	if got, _ := s.List(uid, ""); len(got) != 1 {
		t.Fatalf("active list after restore = %d, want 1", len(got))
	}
	if got, _ := s.TrashList(uid); len(got) != 0 {
		t.Fatalf("trash list after restore = %d, want 0", len(got))
	}
}

// 重复移除与重复还原必须保持同一结果；重复移除不得重置七天计时。
func TestMoodTrashIdempotentRepeatedActions(t *testing.T) {
	s, _, _ := newMoodServiceWithClock()
	const uid uint = 1
	v, _ := s.Create(uid, sampleMoodRequest("first"))

	if e := s.Delete(uid, v.ID); e != nil {
		t.Fatalf("first delete: %v", e)
	}
	firstTrash, _ := s.TrashList(uid)
	firstDeletedAt := firstTrash[0].DeletedAt.Time

	if e := s.Delete(uid, v.ID); e != nil {
		t.Fatalf("repeated delete must stay idempotent, got %v", e)
	}
	secondTrash, _ := s.TrashList(uid)
	if len(secondTrash) != 1 || !secondTrash[0].DeletedAt.Time.Equal(firstDeletedAt) {
		t.Fatalf("repeated delete must not reset the 7-day window or duplicate items: %+v", secondTrash)
	}

	if _, e := s.Restore(uid, v.ID); e != nil {
		t.Fatalf("first restore: %v", e)
	}
	again, e := s.Restore(uid, v.ID)
	if e != nil || again.ID != v.ID {
		t.Fatalf("repeated restore must return the same active record, got id=%d err=%v", again.ID, e)
	}
	if got, _ := s.TrashList(uid); len(got) != 0 {
		t.Fatalf("trash after repeated restore = %d, want 0", len(got))
	}
	if got, _ := s.List(uid, ""); len(got) != 1 {
		t.Fatalf("active list after repeated restore = %d, want 1", len(got))
	}
}

// 超过七天：回收站不再展示，不能还原、不能再次移除，记录永久消失。
func TestMoodTrashExpiry(t *testing.T) {
	s, _, now := newMoodServiceWithClock()
	const uid uint = 1
	v, _ := s.Create(uid, sampleMoodRequest("expiring"))
	if e := s.Delete(uid, v.ID); e != nil {
		t.Fatalf("delete: %v", e)
	}

	cases := []struct {
		name string
		step time.Duration
	}{
		{"within retention appears in trash", 6 * 24 * time.Hour},
		{"after 7 days gone forever", constants.MoodTrashRetention + time.Hour},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			*now = now.Add(tt.step)
			trash, e := s.TrashList(uid)
			if e != nil {
				t.Fatalf("trash list: %v", e)
			}
			within := tt.step < constants.MoodTrashRetention
			if len(trash) != boolToInt(within) {
				t.Fatalf("trash size = %d, want %v", len(trash), within)
			}
			if within {
				return
			}
			if _, e := s.Restore(uid, v.ID); !errors.Is(e, repository.ErrNotFound) {
				t.Fatalf("restore expired must yield ErrNotFound, got %v", e)
			}
			if e := s.Delete(uid, v.ID); !errors.Is(e, repository.ErrNotFound) {
				t.Fatalf("re-delete expired must yield ErrNotFound, got %v", e)
			}
			if got, _ := s.List(uid, ""); len(got) != 0 {
				t.Fatalf("expired record leaked back into active list: %d", len(got))
			}
		})
	}
}

// 回收站与还原只作用于当前账号。
func TestMoodTrashAccountIsolation(t *testing.T) {
	s, _, _ := newMoodServiceWithClock()
	const alice, bob uint = 1, 2
	av, _ := s.Create(alice, sampleMoodRequest("alice"))
	if _, e := s.Create(bob, sampleMoodRequest("bob")); e != nil {
		t.Fatalf("bob create: %v", e)
	}
	if e := s.Delete(alice, av.ID); e != nil {
		t.Fatalf("alice delete: %v", e)
	}

	cases := []struct {
		name string
		uid  uint
		want int
	}{
		{"owner sees own trashed record", alice, 1},
		{"other account trash is empty", bob, 0},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, e := s.TrashList(tt.uid)
			if e != nil || len(got) != tt.want {
				t.Fatalf("trash list for uid=%d = %d items, err=%v; want %d", tt.uid, len(got), e, tt.want)
			}
		})
	}
	if _, e := s.Restore(bob, av.ID); !errors.Is(e, repository.ErrNotFound) {
		t.Fatalf("other account must not restore foreign record, got %v", e)
	}
	if got, _ := s.TrashList(alice); len(got) != 1 {
		t.Fatalf("owner record must remain trashed after foreign restore attempt: %d", len(got))
	}
}

// 回收站中的记录不能被正常更新接口改动，还原后仍是原日期的内容。
func TestMoodTrashUpdateSkipsTrashed(t *testing.T) {
	s, _, _ := newMoodServiceWithClock()
	const uid uint = 1
	v, _ := s.Create(uid, sampleMoodRequest("keep"))
	if e := s.Delete(uid, v.ID); e != nil {
		t.Fatalf("delete: %v", e)
	}
	changed := sampleMoodRequest("changed")
	changed.RecordDate = "2026-09-01"
	if _, e := s.Update(uid, v.ID, changed); !errors.Is(e, repository.ErrNotFound) {
		t.Fatalf("update on trashed record must yield ErrNotFound, got %v", e)
	}
	restored, e := s.Restore(uid, v.ID)
	if e != nil {
		t.Fatalf("restore: %v", e)
	}
	if restored.Note != "keep" || restored.RecordDate.Format("2006-01-02") != "2026-09-17" {
		t.Fatalf("trashed record was mutated while in recycle bin: %+v", restored)
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
