package service

import (
	"context"
	"errors"
	"time"

	"github.com/husky/husky/internal/middleware/auth"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务接口
type AuthService interface {
	Login(ctx context.Context, email, password string) (string, *model.User, error)
	Register(ctx context.Context, email, username, password string) (*model.User, error)
	RefreshToken(ctx context.Context, tokenString string) (string, error)
	GetCurrentUser(ctx context.Context, userID uint) (*model.User, error)
	UpdateProfile(ctx context.Context, userID uint, req *model.UpdateUserRequest) (*model.User, error)
}

type authService struct {
	userRepo *repository.UserRepository
}

// NewAuthService 创建认证服务实例
func NewAuthService(userRepo *repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (s *authService) Login(ctx context.Context, email, password string) (string, *model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("invalid email or password")
		}
		return "", nil, err
	}

	if user.Status == 0 {
		return "", nil, errors.New("account is disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("invalid email or password")
	}

	now := time.Now()
	user.LastLoginAt = &now
	s.userRepo.Update(ctx, user)

	token, err := auth.GenerateToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *authService) Register(ctx context.Context, email, username, password string) (*model.User, error) {
	if email == "" || username == "" || password == "" {
		return nil, errors.New("email, username and password are required")
	}

	existing, _ := s.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:    email,
		Username: username,
		Password: string(hashed),
		Status:   1,
		Role:     "user",
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) RefreshToken(ctx context.Context, tokenString string) (string, error) {
	return auth.RefreshToken(tokenString)
}

func (s *authService) GetCurrentUser(ctx context.Context, userID uint) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *authService) UpdateProfile(ctx context.Context, userID uint, req *model.UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Department != "" {
		user.Department = req.Department
	}
	if req.Title != "" {
		user.Title = req.Title
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}
