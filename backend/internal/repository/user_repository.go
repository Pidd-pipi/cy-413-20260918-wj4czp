package repository

import (
	"errors"
	"github.com/blueship581/mindgarden/backend/internal/model"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type UserRepository interface {
	Create(*model.User) error
	ByEmail(string) (*model.User, error)
	ByID(uint) (*model.User, error)
	Update(*model.User) error
}
type userRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository   { return &userRepository{db} }
func (r *userRepository) Create(u *model.User) error { return r.db.Create(u).Error }
func (r *userRepository) ByEmail(email string) (*model.User, error) {
	var u model.User
	e := r.db.Where("email = ?", email).First(&u).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &u, e
}
func (r *userRepository) ByID(id uint) (*model.User, error) {
	var u model.User
	e := r.db.First(&u, id).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &u, e
}
func (r *userRepository) Update(u *model.User) error { return r.db.Save(u).Error }
