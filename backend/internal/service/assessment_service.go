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
)

type AssessmentService struct {
	repo   repository.AssessmentRepository
	logger *slog.Logger
}

func NewAssessmentService(r repository.AssessmentRepository, l *slog.Logger) *AssessmentService {
	return &AssessmentService{r, l}
}
func validCategory(v string) bool {
	for _, x := range constants.AssessmentCategories {
		if x == v {
			return true
		}
	}
	return false
}
func (s *AssessmentService) List() ([]model.Assessment, error) {
	v, e := s.repo.List()
	if e != nil {
		return nil, fmt.Errorf("Assessment list failed: %w", e)
	}
	s.logger.Info(constants.LogAssessmentListed)
	return v, nil
}
func (s *AssessmentService) Create(req dto.AssessmentRequest) (*model.Assessment, error) {
	if !validCategory(req.Category) {
		return nil, util.NewAppError(constants.CodeValidation, "Assessment[category] create failed: unsupported category", nil)
	}
	v := &model.Assessment{Title: req.Title, Description: req.Description, Category: req.Category, Questions: req.Questions, ScoringRule: req.ScoringRule}
	if e := s.repo.Create(v); e != nil {
		return nil, fmt.Errorf("Assessment[title] create failed: %w", e)
	}
	s.logger.Info(constants.LogAssessmentCreated, "assessment_id", v.ID)
	return v, nil
}
func (s *AssessmentService) Take(uid, id uint, req dto.TakeAssessmentRequest) (*model.UserAssessment, error) {
	a, e := s.repo.ByID(id)
	if e != nil {
		return nil, fmt.Errorf("Assessment[id=%d] fetch failed: %w", id, e)
	}
	score := 0
	for _, v := range req.Answers {
		score += v
	}
	result := "状态平稳"
	suggestion := "继续保持每天记录与规律作息。"
	if score >= len(req.Answers)*4 {
		result = "需要更多关照"
		suggestion = "尝试短暂呼吸练习；若困扰持续，请联系专业心理健康服务。"
	}
	b, _ := json.Marshal(req.Answers)
	v := &model.UserAssessment{UserID: uid, AssessmentID: a.ID, Answers: string(b), Score: score, Result: result, Suggestion: suggestion}
	if e = s.repo.CreateUserAssessment(v); e != nil {
		return nil, fmt.Errorf("UserAssessment[assessment_id] create failed: %w", e)
	}
	s.logger.Info(constants.LogAssessmentTaken, "user_id", uid, "assessment_id", id)
	return v, nil
}
func (s *AssessmentService) Report(uid uint) ([]model.UserAssessment, error) {
	v, e := s.repo.Report(uid)
	if e != nil {
		return nil, fmt.Errorf("UserAssessment[user_id] report failed: %w", e)
	}
	return v, nil
}
func (s *AssessmentService) Seed() error {
	n, e := s.repo.Count()
	if e != nil {
		return e
	}
	if n > 0 {
		return nil
	}
	q := "[{\"id\":1,\"text\":\"过去一周，我能平静地面对日常任务。\"},{\"id\":2,\"text\":\"我有足够的休息与恢复时间。\"},{\"id\":3,\"text\":\"我能觉察自己的情绪变化。\"}]"
	for _, v := range []model.Assessment{{Title: "三分钟压力自检", Description: "快速了解近期压力感受", Category: constants.AssessmentStress, Questions: q, ScoringRule: "sum"}, {Title: "睡眠关怀小测", Description: "回顾睡眠质量与休息节律", Category: constants.AssessmentSleep, Questions: q, ScoringRule: "sum"}} {
		if e = s.repo.Create(&v); e != nil {
			return e
		}
	}
	s.logger.Info(constants.LogSeeded)
	return nil
}
