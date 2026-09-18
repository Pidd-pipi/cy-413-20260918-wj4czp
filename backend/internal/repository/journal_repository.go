package repository

import (
	"errors"
	"github.com/blueship581/mindgarden/backend/internal/model"
	"gorm.io/gorm"
)

type JournalRepository interface {
	Create(*model.Journal) error
	List(uint, int) ([]model.Journal, error)
	ByID(uint, uint) (*model.Journal, error)
	Update(*model.Journal) error
	Delete(*model.Journal) error
}
type journalRepository struct{ db *gorm.DB }

func NewJournalRepository(db *gorm.DB) JournalRepository   { return &journalRepository{db} }
func (r *journalRepository) Create(v *model.Journal) error { return r.db.Create(v).Error }
func (r *journalRepository) List(uid uint, level int) (out []model.Journal, e error) {
	q := r.db.Where("user_id = ?", uid)
	if level > 0 {
		q = q.Where("mood_level = ?", level)
	}
	e = q.Order("created_at desc").Find(&out).Error
	return
}
func (r *journalRepository) ByID(id, uid uint) (*model.Journal, error) {
	var v model.Journal
	e := r.db.Where("id=? AND user_id=?", id, uid).First(&v).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &v, e
}
func (r *journalRepository) Update(v *model.Journal) error { return r.db.Save(v).Error }
func (r *journalRepository) Delete(v *model.Journal) error { return r.db.Delete(v).Error }
