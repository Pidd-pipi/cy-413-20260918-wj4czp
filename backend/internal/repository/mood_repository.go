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
	ByIDAnyState(uint, uint) (*model.Mood, error)
	ListTrash(uint) ([]model.Mood, error)
	Update(*model.Mood) error
	SoftDelete(*model.Mood) error
	Restore(*model.Mood) error
	PurgeExpiredForUser(uint, time.Time) (int64, error)
	PurgeExpired(time.Time) (int64, error)
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
func (r *moodRepository) ByIDAnyState(id, uid uint) (*model.Mood, error) {
	var v model.Mood
	e := r.db.Unscoped().Where("id = ? AND user_id = ?", id, uid).First(&v).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &v, e
}
func (r *moodRepository) ListTrash(uid uint) (out []model.Mood, e error) {
	e = r.db.Unscoped().Where("user_id = ? AND deleted_at IS NOT NULL", uid).
		Order("deleted_at desc, id desc").Find(&out).Error
	return
}
func (r *moodRepository) Update(v *model.Mood) error { return r.db.Save(v).Error }
func (r *moodRepository) SoftDelete(v *model.Mood) error {
	return r.db.Delete(v).Error
}
func (r *moodRepository) Restore(v *model.Mood) error {
	if e := r.db.Unscoped().Model(v).Update("deleted_at", nil).Error; e != nil {
		return e
	}
	v.DeletedAt = gorm.DeletedAt{}
	return nil
}
func (r *moodRepository) PurgeExpiredForUser(uid uint, cutoff time.Time) (n int64, e error) {
	res := r.db.Unscoped().
		Where("user_id = ? AND deleted_at IS NOT NULL AND deleted_at < ?", uid, cutoff).
		Delete(&model.Mood{})
	return res.RowsAffected, res.Error
}
func (r *moodRepository) PurgeExpired(cutoff time.Time) (n int64, e error) {
	res := r.db.Unscoped().Where("deleted_at IS NOT NULL AND deleted_at < ?", cutoff).
		Delete(&model.Mood{})
	return res.RowsAffected, res.Error
}
