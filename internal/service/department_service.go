package service

import (
	"context"
	"fmt"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
)

// DepartmentService 部门服务接口
type DepartmentService interface {
	Create(ctx context.Context, req *model.CreateDepartmentRequest) (*model.Department, error)
	GetByID(ctx context.Context, id uint) (*model.Department, error)
	List(ctx context.Context) ([]model.Department, error)
	Update(ctx context.Context, id uint, req *model.UpdateDepartmentRequest) (*model.Department, error)
	Delete(ctx context.Context, id uint) error
}

type departmentService struct {
	deptRepo *repository.DepartmentRepository
}

func NewDepartmentService(deptRepo *repository.DepartmentRepository) DepartmentService {
	return &departmentService{deptRepo: deptRepo}
}

func (s *departmentService) Create(ctx context.Context, req *model.CreateDepartmentRequest) (*model.Department, error) {
	if req.Name == "" || req.Code == "" {
		return nil, fmt.Errorf("name and code are required")
	}

	dept := &model.Department{
		Name:      req.Name,
		Code:      req.Code,
		ParentID:  req.ParentID,
		ManagerID: req.ManagerID,
		Status:    1,
	}

	if req.ParentID != nil {
		parent, err := s.deptRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent department not found: %w", err)
		}
		dept.Path = parent.Path + "/" + req.Code
	} else {
		dept.Path = "/" + req.Code
	}

	if err := s.deptRepo.Create(ctx, dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *departmentService) GetByID(ctx context.Context, id uint) (*model.Department, error) {
	dept, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("department not found: %w", err)
	}
	return dept, nil
}

func (s *departmentService) List(ctx context.Context) ([]model.Department, error) {
	return s.deptRepo.List(ctx)
}

func (s *departmentService) Update(ctx context.Context, id uint, req *model.UpdateDepartmentRequest) (*model.Department, error) {
	dept, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("department not found: %w", err)
	}

	if req.Name != nil {
		dept.Name = *req.Name
	}
	if req.Code != nil {
		dept.Code = *req.Code
	}
	if req.ParentID != nil {
		dept.ParentID = req.ParentID
		parent, err := s.deptRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent department not found: %w", err)
		}
		dept.Path = parent.Path + "/" + dept.Code
	}
	if req.ManagerID != nil {
		dept.ManagerID = req.ManagerID
	}
	if req.Status != nil {
		dept.Status = *req.Status
	}

	if err := s.deptRepo.Update(ctx, dept); err != nil {
		return nil, err
	}
	return dept, nil
}

func (s *departmentService) Delete(ctx context.Context, id uint) error {
	return s.deptRepo.Delete(ctx, id)
}
