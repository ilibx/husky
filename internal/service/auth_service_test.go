package service

import (
	"context"
	"testing"

	"github.com/husky/husky/internal/middleware/auth"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	auth.SetJWTConfig("test-secret", 24)
}

func setupAuthTest(t *testing.T) (*repository.UserRepository, AuthService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	db.AutoMigrate(&model.User{})
	userRepo := repository.NewUserRepository(db)
	return userRepo, NewAuthService(userRepo)
}

func TestRegister_Success(t *testing.T) {
	_, svc := setupAuthTest(t)

	user, err := svc.Register(context.Background(), "test@example.com", "testuser", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "test@example.com" {
		t.Errorf("got email %s, want test@example.com", user.Email)
	}
	if user.Role != "user" {
		t.Errorf("got role %s, want user", user.Role)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	_, svc := setupAuthTest(t)

	svc.Register(context.Background(), "dup@example.com", "user1", "password123")
	_, err := svc.Register(context.Background(), "dup@example.com", "user2", "password456")
	if err == nil {
		t.Error("expected error for duplicate email")
	}
}

func TestRegister_EmptyFields(t *testing.T) {
	_, svc := setupAuthTest(t)

	tests := []struct {
		email, username, password string
	}{
		{"", "user", "pass"},
		{"email@test.com", "", "pass"},
		{"email@test.com", "user", ""},
	}

	for _, tt := range tests {
		_, err := svc.Register(context.Background(), tt.email, tt.username, tt.password)
		if err == nil {
			t.Errorf("expected error for empty field: %+v", tt)
		}
	}
}

func TestLogin_Success(t *testing.T) {
	_, svc := setupAuthTest(t)

	svc.Register(context.Background(), "login@test.com", "loginuser", "mypassword")
	token, user, err := svc.Login(context.Background(), "login@test.com", "mypassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if user.Email != "login@test.com" {
		t.Errorf("got email %s, want login@test.com", user.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	_, svc := setupAuthTest(t)

	svc.Register(context.Background(), "wrongpw@test.com", "user", "correctpassword")
	_, _, err := svc.Login(context.Background(), "wrongpw@test.com", "wrongpassword")
	if err == nil {
		t.Error("expected error for wrong password")
	}
}

func TestLogin_NotFound(t *testing.T) {
	_, svc := setupAuthTest(t)

	_, _, err := svc.Login(context.Background(), "nonexistent@test.com", "password")
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}
