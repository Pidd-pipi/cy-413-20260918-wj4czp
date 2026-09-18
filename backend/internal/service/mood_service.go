package service

import (
	"encoding/json"
	"errors"
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

// Delete 把记录移入七天回收站：记录立即退出列表、筛选与本周曲线（软删除默认作用域已排除）。
// 重复移除保持同一结果且不重置七天计时；已过期或不存在返回 not found。
func (s *MoodService) Delete(uid, id uint) error {
	if _, e := s.PurgeExpired(); e != nil {
		return fmt.Errorf("Mood[id=%d] purge expired failed: %w", id, e)
	}
	if trashed, e := s.repo.TrashByID(id, uid); e == nil {
		// 已在回收站中：重复移除幂等，且不刷新 deleted_at，过期时刻保持不变。
		s.logger.Info(constants.LogMoodDeleted, "mood_id", trashed.ID, "already_trashed", true)
		return nil
	} else if !errors.Is(e, repository.ErrNotFound) {
		return fmt.Errorf("Mood[id=%d] trash fetch failed: %w", id, e)
	}
	v, e := s.repo.ByID(id, uid)
	if e != nil {
		return fmt.Errorf("Mood[id=%d] fetch failed: %w", id, e)
	}
	if e = s.repo.Delete(v); e != nil {
		return util.WrapEntity("Mood", "id", id, constants.CodeInternal, e)
	}
	s.logger.Info(constants.LogMoodDeleted, "mood_id", id, "already_trashed", false)
	return nil
}

// TrashList 只返回当前账号回收站中尚未过期的记录。
func (s *MoodService) TrashList(uid uint) ([]dto.MoodTrashItem, error) {
	if _, e := s.PurgeExpired(); e != nil {
		return nil, fmt.Errorf("Mood[user_id] trash list purge failed: %w", e)
	}
	vs, e := s.repo.TrashList(uid)
	if e != nil {
		return nil, fmt.Errorf("Mood[user_id] trash list failed: %w", e)
	}
	cutoff := s.now().Add(-constants.MoodTrashRetention)
	items := make([]dto.MoodTrashItem, 0, len(vs))
	for _, v := range vs {
		// 双保险：只保留仍在七天窗口内的记录，过期记录不能再出现。
		if v.DeletedAt.Time.Before(cutoff) {
			continue
		}
		items = append(items, dto.MoodTrashItem{Mood: v, ExpiresAt: util.MoodTrashExpiresAt(v.DeletedAt.Time)})
	}
	s.logger.Info(constants.LogMoodTrashListed, "user_id", uid, "count", len(items))
	return items, nil
}

// Restore 按原日期还原记录（record_date、等级、标签、备注均不变）。
// 重复还原保持同一结果：已还原的活动记录直接返回；已过期或不存在返回 not found。
func (s *MoodService) Restore(uid, id uint) (*model.Mood, error) {
	if _, e := s.PurgeExpired(); e != nil {
		return nil, fmt.Errorf("Mood[id=%d] purge expired failed: %w", id, e)
	}
	v, e := s.repo.TrashByID(id, uid)
	if e == nil {
		if e = s.repo.Restore(v); e != nil {
			return nil, util.WrapEntity("Mood", "deleted_at", id, constants.CodeInternal, e)
		}
		v.DeletedAt.Valid = false
		s.logger.Info(constants.LogMoodRestored, "mood_id", id)
		return v, nil
	}
	if !errors.Is(e, repository.ErrNotFound) {
		return nil, fmt.Errorf("Mood[id=%d] trash fetch failed: %w", id, e)
	}
	// 已还原（重复还原）或从未移除：活动记录直接返回，保持同一结果。
	active, err := s.repo.ByID(id, uid)
	if err != nil {
		return nil, fmt.Errorf("Mood[id=%d] restore fetch failed: %w", id, err)
	}
	s.logger.Info(constants.LogMoodRestored, "mood_id", id, "already_restored", true)
	return active, nil
}

// PurgeExpired 硬删除超过七天的回收站记录，过期记录此后不能再出现。
func (s *MoodService) PurgeExpired() (int64, error) {
	cutoff := s.now().Add(-constants.MoodTrashRetention)
	n, e := s.repo.PurgeExpired(cutoff)
	if e != nil {
		return 0, fmt.Errorf("Mood[deleted_at] purge expired failed: %w", e)
	}
	if n > 0 {
		s.logger.Info(constants.LogMoodTrashPurged, "purged", n, "cutoff", cutoff)
	}
	return n, nil
}
