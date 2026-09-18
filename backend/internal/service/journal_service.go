package service

import (
	"fmt"
	"github.com/blueship581/mindgarden/backend/internal/constants"
	"github.com/blueship581/mindgarden/backend/internal/dto"
	"github.com/blueship581/mindgarden/backend/internal/model"
	"github.com/blueship581/mindgarden/backend/internal/repository"
	"github.com/blueship581/mindgarden/backend/internal/util"
	"log/slog"
)

type JournalService struct {
	repo   repository.JournalRepository
	logger *slog.Logger
}

func NewJournalService(r repository.JournalRepository, l *slog.Logger) *JournalService {
	return &JournalService{r, l}
}
func (s *JournalService) Create(uid uint, req dto.JournalRequest) (*model.Journal, error) {
	v := &model.Journal{UserID: uid, Title: req.Title, Content: req.Content, MoodLevel: req.MoodLevel, Weather: req.Weather, IsPrivate: req.IsPrivate}
	if e := s.repo.Create(v); e != nil {
		return nil, fmt.Errorf("Journal[user_id] create failed: %w", e)
	}
	s.logger.Info(constants.LogJournalCreated, "user_id", uid)
	return v, nil
}
func (s *JournalService) List(uid uint, level int) ([]model.Journal, error) {
	v, e := s.repo.List(uid, level)
	if e != nil {
		return nil, fmt.Errorf("Journal[user_id] list failed: %w", e)
	}
	s.logger.Info(constants.LogJournalListed, "user_id", uid)
	return v, nil
}
func (s *JournalService) Update(uid, id uint, req dto.JournalRequest) (*model.Journal, error) {
	v, e := s.repo.ByID(id, uid)
	if e != nil {
		return nil, fmt.Errorf("Journal[id=%d] read failed: %w", id, e)
	}
	v.Title = req.Title
	v.Content = req.Content
	v.MoodLevel = req.MoodLevel
	v.Weather = req.Weather
	v.IsPrivate = req.IsPrivate
	if e = s.repo.Update(v); e != nil {
		return nil, util.WrapEntity("Journal", "content", id, constants.CodeInternal, e)
	}
	s.logger.Info(constants.LogJournalUpdated, "journal_id", id)
	return v, nil
}
func (s *JournalService) Delete(uid, id uint) error {
	v, e := s.repo.ByID(id, uid)
	if e != nil {
		return fmt.Errorf("Journal[id=%d] read failed: %w", id, e)
	}
	if e = s.repo.Delete(v); e != nil {
		return util.WrapEntity("Journal", "id", id, constants.CodeInternal, e)
	}
	s.logger.Info(constants.LogJournalDeleted, "journal_id", id)
	return nil
}
