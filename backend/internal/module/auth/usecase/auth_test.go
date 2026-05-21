package usecase_test

import (
	"testing"
	"time"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	"github.com/durianpay/fullstack-boilerplate/internal/module/auth/usecase"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	user *entity.User
	err  error
}

func (m *mockUserRepo) GetUserByEmail(email string) (*entity.User, error) {
	return m.user, m.err
}

func hashPassword(t *testing.T, pw string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	return string(h)
}

func TestLogin_Success(t *testing.T) {
	repo := &mockUserRepo{user: &entity.User{
		ID: "1", Email: "cs@test.com",
		PasswordHash: hashPassword(t, "password"), Role: "cs",
	}}
	uc := usecase.NewAuthUsecase(repo, []byte("secret"), time.Hour)
	token, user, err := uc.Login("cs@test.com", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if user.Email != "cs@test.com" {
		t.Errorf("expected email cs@test.com, got %s", user.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := &mockUserRepo{user: &entity.User{
		ID: "1", Email: "cs@test.com",
		PasswordHash: hashPassword(t, "password"), Role: "cs",
	}}
	uc := usecase.NewAuthUsecase(repo, []byte("secret"), time.Hour)
	_, _, err := uc.Login("cs@test.com", "wrong")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{err: entity.ErrorNotFound("user not found")}
	uc := usecase.NewAuthUsecase(repo, []byte("secret"), time.Hour)
	_, _, err := uc.Login("nobody@test.com", "password")
	if err == nil {
		t.Fatal("expected error for missing user")
	}
}
