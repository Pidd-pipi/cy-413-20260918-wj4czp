package service

import (
	"errors"
	"github.com/blueship581/mindgarden/backend/internal/dto"
	"github.com/blueship581/mindgarden/backend/internal/model"
	"github.com/blueship581/mindgarden/backend/internal/repository"
	"io"
	"log/slog"
	"testing"
)

type fakeUserRepo struct {
	users map[string]*model.User
	next  uint
}

func (f *fakeUserRepo) Create(u *model.User) error {
	f.next++
	u.ID = f.next
	f.users[u.Email] = u
	return nil
}
func (f *fakeUserRepo) ByEmail(e string) (*model.User, error) {
	u, ok := f.users[e]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return u, nil
}
func (f *fakeUserRepo) ByID(id uint) (*model.User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, repository.ErrNotFound
}
func (f *fakeUserRepo) Update(u *model.User) error { return nil }
func TestUserServiceRegisterAndLogin(t *testing.T) {
	repo := &fakeUserRepo{users: map[string]*model.User{}}
	s := NewUserService(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	cases := []struct {
		name, email, password string
		wantErr               bool
	}{{"register", "a@example.com", "password1", false}, {"duplicate", "a@example.com", "password1", true}}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, e := s.Register(dto.RegisterRequest{Email: tt.email, Password: tt.password, Nickname: "Garden"})
			if (e != nil) != tt.wantErr {
				t.Fatalf("err=%v", e)
			}
		})
	}
	u, e := s.Login(dto.LoginRequest{Email: "a@example.com", Password: "password1"})
	if e != nil || u.Email != "a@example.com" {
		t.Fatalf("login got %v %v", u, e)
	}
	_, e = s.Login(dto.LoginRequest{Email: "a@example.com", Password: "wrong"})
	if e == nil || errors.Is(e, repository.ErrNotFound) {
		t.Fatalf("expected credential error")
	}
}
