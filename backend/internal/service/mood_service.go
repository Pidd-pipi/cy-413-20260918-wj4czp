package service

import (
	"encoding/json"
	"fmt"
	"github.com/blueship581/mindgarden/backend/internal/constants"
	"github.com/blueship581/mindgarden/backend/internal/dto"
	"github.com/blueship581/mindgarden/backend/internal/model"
	"github.com/blueship581/mindgarden/backend/internal/repository"
	"github.com/blueship581/mindgarden/backend/internal/util"
	"log/slog"
	"strings"
	"time"
)

type MoodService struct {
	repo   repository.MoodRepository
	logger *slog.Logger
	now    func() time.Time
}

func NewMoodService(r repository.MoodRepository, l *slog.Logger) *MoodService {
	return &MoodService{repo: r, logger: l, now: time.Now}
}
func validTags(tags []string) bool {
	allowed := map[string]bool{}
	for _, v := range constants.MoodTags {
		allowed[v] = true
	}
	for _, v := range tags {
		if !allowed[v] {
			return false
		}
	}
	return true
}
func (s *MoodService) Create(uid uint, req dto.MoodRequest) (*model.Mood, error) {
	if !validTags(req.MoodTags) {
		return nil, util.NewAppError(constants.CodeValidation, "Mood[mood_tags] create failed: unsupported tag", nil)
	}
	d, e := time.Parse("2006-01-02", req.RecordDate)
	if e != nil {
		return nil, util.NewAppError(constants.CodeValidation, "Mood[record_date] create failed: invalid date", e)
	}
	b, _ := json.Marshal(req.MoodTags)
	v := &model.Mood{UserID: uid, MoodLevel: req.MoodLevel, MoodTags: string(b), Note: strings.TrimSpace(req.Note), RecordDate: d}
	if e = s.repo.Create(v); e != nil {
		return nil, fmt.Errorf("Mood[user_id] create failed: %w", e)
	}
	s.logger.Info(constants.LogMoodCreated, "user_id", uid, "mood_level", v.MoodLevel)
	return v, nil
}
func (s *MoodService) List(uid uint, date string) ([]model.Mood, error) {
	var d *time.Time
	if date != "" {
		x, e := time.Parse("2006-01-02", date)
		if e != nil {
			return nil, util.NewAppError(constants.CodeValidation, "Mood[record_date] list failed: invalid date", e)
		}
		d = &x
	}
	vs, e := s.repo.List(uid, d)
	if e != nil {
		return nil, fmt.Errorf("Mood[user_id] list failed: %w", e)
	}
	s.logger.Info(constants.LogMoodListed, "user_id", uid)
	return vs, nil
}
func (s *MoodService) Update(uid, id uint, req dto.MoodRequest) (*model.Mood, error) {
	v, e := s.repo.ByID(id, uid)
	if e != nil {
		return nil, fmt.Errorf("Mood[id=%d] fetch failed: %w", id, e)
	}
	if !validTags(req.MoodTags) {
		return nil, util.WrapEntity("Mood", "mood_tags", id, constants.CodeValidation, nil)
	}
	d, e := time.Parse("2006-01-02", req.RecordDate)
	if e != nil {
		return nil, util.WrapEntity("Mood", "record_date", id, constants.CodeValidation, e)
	}
	b, _ := json.Marshal(req.MoodTags)
	v.MoodLevel = req.MoodLevel
	v.MoodTags = string(b)
	v.Note = req.Note
	v.RecordDate = d
	if e = s.repo.Update(v); e != nil {
		return nil, util.WrapEntity("Mood", "mood_level", id, constants.CodeInternal, e)
	}
	s.logger.Info(constants.LogMoodUpdated, "mood_id", id)
	return v, nil
}

// Delete moves a mood into the seven-day recycle bin. Removing an already
// removed record is a no-op that keeps the original deletion time, so the
// seven-day window is never extended.
func (s *MoodService) Delete(uid, id uint) error {
	v, e := s.repo.ByIDAnyState(id, uid)
	if e != nil {
		return fmt.Errorf("Mood[id=%d] fetch failed: %w", id, e)
	}
	if v.DeletedAt.Valid {
		if s.expired(v.DeletedAt.Time) {
			if _, e = s.repo.PurgeExpiredForUser(uid, s.cutoff()); e != nil {
				return util.WrapEntity("Mood", "deleted_at", id, constants.CodeInternal, e)
			}
			return fmt.Errorf("Mood[id=%d] fetch failed: %w", id, repository.ErrNotFound)
		}
		s.logger.Info(constants.LogMoodDeleted, "mood_id", id, "already_removed", true)
		return nil
	}
	if e = s.repo.SoftDelete(v); e != nil {
		return util.WrapEntity("Mood", "id", id, constants.CodeInternal, e)
	}
	s.logger.Info(constants.LogMoodDeleted, "mood_id", id, "already_removed", false)
	return nil
}

// ListTrash permanently removes records whose seven-day window elapsed before
// listing them, so expired records can never reappear.
func (s *MoodService) ListTrash(uid uint) ([]model.Mood, error) {
	if _, e := s.repo.PurgeExpiredForUser(uid, s.cutoff()); e != nil {
		return nil, fmt.Errorf("Mood[deleted_at] purge failed: %w", e)
	}
	vs, e := s.repo.ListTrash(uid)
	if e != nil {
		return nil, fmt.Errorf("Mood[deleted_at] list failed: %w", e)
	}
	s.logger.Info(constants.LogMoodTrashListed, "user_id", uid, "count", len(vs))
	return vs, nil
}

// Restore brings a removed mood back to the active list. The original date,
// level, tags and note are untouched. Restoring an already active record is a
// no-op; expired records are purged and treated as not found.
func (s *MoodService) Restore(uid, id uint) (*model.Mood, error) {
	v, e := s.repo.ByIDAnyState(id, uid)
	if e != nil {
		return nil, fmt.Errorf("Mood[id=%d] fetch failed: %w", id, e)
	}
	if !v.DeletedAt.Valid {
		s.logger.Info(constants.LogMoodRestored, "mood_id", id, "already_active", true)
		return v, nil
	}
	if s.expired(v.DeletedAt.Time) {
		if _, e = s.repo.PurgeExpiredForUser(uid, s.cutoff()); e != nil {
			return nil, util.WrapEntity("Mood", "deleted_at", id, constants.CodeInternal, e)
		}
		return nil, fmt.Errorf("Mood[id=%d] fetch failed: %w", id, repository.ErrNotFound)
	}
	if e = s.repo.Restore(v); e != nil {
		return nil, util.WrapEntity("Mood", "deleted_at", id, constants.CodeInternal, e)
	}
	s.logger.Info(constants.LogMoodRestored, "mood_id", id, "already_active", false)
	return v, nil
}

// PurgeExpired permanently removes every record past its seven-day window. It
// is safe to run repeatedly and returns the number of purged rows.
func (s *MoodService) PurgeExpired() (int64, error) {
	n, e := s.repo.PurgeExpired(s.cutoff())
	if e != nil {
		return 0, fmt.Errorf("Mood[deleted_at] purge failed: %w", e)
	}
	if n > 0 {
		s.logger.Info(constants.LogMoodPurged, "count", n)
	}
	return n, nil
}

func (s *MoodService) cutoff() time.Time         { return s.now().Add(-constants.MoodTrashRetention) }
func (s *MoodService) expired(at time.Time) bool { return at.Before(s.cutoff()) }
