package service

import (
	"context"
	"fmt"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
)

// UserService 用户管理服务接口
type UserService interface {
	List(ctx context.Context, offset, limit int, keyword string) ([]model.User, int64, error)
	GetByID(ctx context.Context, id uint) (*model.User, error)
	Update(ctx context.Context, id uint, req *model.UpdateUserRequest) (*model.User, error)
	Delete(ctx context.Context, id uint) error
	ChangeRole(ctx context.Context, id uint, role string) error
}

type userService struct {
	userRepo *repository.UserRepository
}

// NewUserService 创建用户管理服务
func NewUserService(userRepo *repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) List(ctx context.Context, offset, limit int, keyword string) ([]model.User, int64, error) {
	return s.userRepo.List(ctx, offset, limit, keyword)
}

func (s *userService) GetByID(ctx context.Context, id uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

func (s *userService) Update(ctx context.Context, id uint, req *model.UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
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

func (s *userService) Delete(ctx context.Context, id uint) error {
	return s.userRepo.Delete(ctx, id)
}

func (s *userService) ChangeRole(ctx context.Context, id uint, role string) error {
	validRoles := map[string]bool{"user": true, "admin": true, "agent": true}
	if !validRoles[role] {
		return fmt.Errorf("invalid role: %s", role)
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	user.Role = role
	return s.userRepo.Update(ctx, user)
}
