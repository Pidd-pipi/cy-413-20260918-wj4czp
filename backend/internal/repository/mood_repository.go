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
