package repository

import (
	"errors"
	"github.com/blueship581/mindgarden/backend/internal/model"
	"gorm.io/gorm"
	"time"
)

type MoodRepository interface {
	Create(*model.Mood) error
	List(uint, *time.Time) ([]model.Mood, error)
	ByID(uint, uint) (*model.Mood, error)
	Update(*model.Mood) error
	Delete(*model.Mood) error
	// TrashList 只返回当前账号回收站中尚未到期的记录（deleted_at 非空）。
	TrashList(uid uint) ([]model.Mood, error)
	// TrashByID 绕过软删除作用域，按账号读取任意状态的记录。
	TrashByID(id, uid uint) (*model.Mood, error)
	// Restore 清除 deleted_at，把记录还原回正常列表。
	Restore(v *model.Mood) error
	// PurgeExpired 硬删除所有用户在 cutoff 之前移除的记录，返回删除条数。
	PurgeExpired(cutoff time.Time) (int64, error)
}
type moodRepository struct{ db *gorm.DB }

func NewMoodRepository(db *gorm.DB) MoodRepository   { return &moodRepository{db} }
func (r *moodRepository) Create(v *model.Mood) error { return r.db.Create(v).Error }
func (r *moodRepository) List(uid uint, date *time.Time) (out []model.Mood, e error) {
	q := r.db.Where("user_id = ?", uid)
	if date != nil {
		q = q.Where("record_date >= ? AND record_date < ?", date.Truncate(24*time.Hour), date.Truncate(24*time.Hour).AddDate(0, 0, 1))
	}
	e = q.Order("record_date desc, id desc").Find(&out).Error
	return
}
func (r *moodRepository) ByID(id, uid uint) (*model.Mood, error) {
	var v model.Mood
	e := r.db.Where("id = ? AND user_id = ?", id, uid).First(&v).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &v, e
}
func (r *moodRepository) Update(v *model.Mood) error { return r.db.Save(v).Error }
func (r *moodRepository) Delete(v *model.Mood) error { return r.db.Delete(v).Error }
func (r *moodRepository) TrashByID(id, uid uint) (*model.Mood, error) {
	var v model.Mood
	e := r.db.Unscoped().Where("id = ? AND user_id = ? AND deleted_at IS NOT NULL", id, uid).First(&v).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &v, e
}
func (r *moodRepository) TrashList(uid uint) (out []model.Mood, e error) {
	e = r.db.Unscoped().Where("user_id = ? AND deleted_at IS NOT NULL", uid).Order("deleted_at desc, id desc").Find(&out).Error
	return
}
func (r *moodRepository) Restore(v *model.Mood) error {
	return r.db.Unscoped().Model(&model.Mood{}).Where("id = ?", v.ID).Update("deleted_at", nil).Error
}
func (r *moodRepository) PurgeExpired(cutoff time.Time) (n int64, e error) {
	res := r.db.Unscoped().Where("deleted_at IS NOT NULL AND deleted_at < ?", cutoff).Delete(&model.Mood{})
	return res.RowsAffected, res.Error
}
