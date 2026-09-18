package service

import (
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

type UserService struct {
	repo   repository.UserRepository
	logger *slog.Logger
}

func NewUserService(r repository.UserRepository, l *slog.Logger) *UserService {
	return &UserService{r, l}
}
func (s *UserService) Register(req dto.RegisterRequest) (*model.User, error) {
	_, e := s.repo.ByEmail(strings.ToLower(req.Email))
	if e == nil {
		return nil, util.NewAppError(constants.CodeConflict, "User[email] register failed: already exists", nil)
	}
	if !errors.Is(e, repository.ErrNotFound) {
		return nil, fmt.Errorf("User[email] check failed: %w", e)
	}
	h, e := util.HashPassword(req.Password)
	if e != nil {
		return nil, fmt.Errorf("User[password_hash] hash failed: %w", e)
	}
	u := &model.User{Email: strings.ToLower(req.Email), PasswordHash: h, Nickname: req.Nickname, Role: constants.RoleUser}
	if e = s.repo.Create(u); e != nil {
		return nil, fmt.Errorf("User[email] create failed: %w", e)
	}
	s.logger.Info(constants.LogAuthRegister, "user_id", u.ID)
	return u, nil
}
func (s *UserService) Login(req dto.LoginRequest) (*model.User, error) {
	u, e := s.repo.ByEmail(strings.ToLower(req.Email))
	if errors.Is(e, repository.ErrNotFound) {
		return nil, util.NewAppError(constants.CodeUnauthorized, "User[email] login failed: invalid credentials", e)
	}
	if e != nil {
		return nil, fmt.Errorf("User[email] fetch failed: %w", e)
	}
	if e = util.ComparePassword(u.PasswordHash, req.Password); e != nil {
		return nil, util.NewAppError(constants.CodeUnauthorized, "User[password] login failed: invalid credentials", e)
	}
	s.logger.Info(constants.LogAuthLogin, "user_id", u.ID)
	return u, nil
}
func (s *UserService) Get(id uint) (*model.User, error) {
	u, e := s.repo.ByID(id)
	if e != nil {
		return nil, fmt.Errorf("User[id] read failed: %w", e)
	}
	return u, nil
}
func (s *UserService) Update(id uint, req dto.UpdateProfileRequest) (*model.User, error) {
	u, e := s.Get(id)
	if e != nil {
		return nil, e
	}
	u.Nickname = req.Nickname
	u.Avatar = req.Avatar
	u.Gender = req.Gender
	if req.BirthDate != "" {
		d, p := time.Parse("2006-01-02", req.BirthDate)
		if p != nil {
			return nil, util.WrapEntity("User", "birth_date", id, constants.CodeValidation, p)
		}
		u.BirthDate = &d
	}
	if e = s.repo.Update(u); e != nil {
		return nil, util.WrapEntity("User", "profile", id, constants.CodeInternal, e)
	}
	s.logger.Info(constants.LogUserProfileUpdated, "user_id", id)
	return u, nil
}
