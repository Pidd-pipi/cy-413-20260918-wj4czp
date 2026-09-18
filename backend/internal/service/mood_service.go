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
}

func NewMoodService(r repository.MoodRepository, l *slog.Logger) *MoodService {
	return &MoodService{r, l}
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
func (s *MoodService) Delete(uid, id uint) error {
	v, e := s.repo.ByID(id, uid)
	if e != nil {
		return fmt.Errorf("Mood[id=%d] fetch failed: %w", id, e)
	}
	if e = s.repo.Delete(v); e != nil {
		return util.WrapEntity("Mood", "id", id, constants.CodeInternal, e)
	}
	s.logger.Info(constants.LogMoodDeleted, "mood_id", id)
	return nil
}
