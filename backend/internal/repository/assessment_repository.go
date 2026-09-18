package repository

import (
	"errors"
	"github.com/blueship581/mindgarden/backend/internal/model"
	"gorm.io/gorm"
)

type AssessmentRepository interface {
	List() ([]model.Assessment, error)
	ByID(uint) (*model.Assessment, error)
	Create(*model.Assessment) error
	CreateUserAssessment(*model.UserAssessment) error
	Report(uint) ([]model.UserAssessment, error)
	Count() (int64, error)
}
type assessmentRepository struct{ db *gorm.DB }

func NewAssessmentRepository(db *gorm.DB) AssessmentRepository { return &assessmentRepository{db} }
func (r *assessmentRepository) List() (out []model.Assessment, e error) {
	e = r.db.Order("id desc").Find(&out).Error
	return
}
func (r *assessmentRepository) ByID(id uint) (*model.Assessment, error) {
	var a model.Assessment
	e := r.db.First(&a, id).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &a, e
}
func (r *assessmentRepository) Create(v *model.Assessment) error { return r.db.Create(v).Error }
func (r *assessmentRepository) CreateUserAssessment(v *model.UserAssessment) error {
	return r.db.Create(v).Error
}
func (r *assessmentRepository) Report(uid uint) (out []model.UserAssessment, e error) {
	e = r.db.Where("user_id = ?", uid).Order("created_at desc").Find(&out).Error
	return
}
func (r *assessmentRepository) Count() (int64, error) {
	var n int64
	e := r.db.Model(&model.Assessment{}).Count(&n).Error
	return n, e
}
