package service

import (
	"context"
	"fmt"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
)

// CategoryService 分类服务接口
type CategoryService interface {
	Create(ctx context.Context, req *model.CreateCategoryRequest) (*model.Category, error)
	GetByID(ctx context.Context, id uint) (*model.Category, error)
	List(ctx context.Context, keyword string) ([]model.Category, error)
	Update(ctx context.Context, id uint, req *model.UpdateCategoryRequest) (*model.Category, error)
	Delete(ctx context.Context, id uint) error
}

type categoryService struct {
	catRepo *repository.CategoryRepository
}

// NewCategoryService 创建分类服务
func NewCategoryService(catRepo *repository.CategoryRepository) CategoryService {
	return &categoryService{catRepo: catRepo}
}

func (s *categoryService) Create(ctx context.Context, req *model.CreateCategoryRequest) (*model.Category, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	exists, err := s.catRepo.ExistsByNameAndParent(ctx, req.Name, req.ParentID, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("category name already exists under the same parent")
	}

	cat := &model.Category{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
		ManagerID:   req.ManagerID,
		Status:      1,
		Type:        "system",
	}

	// 构建 Path
	if req.ParentID != nil {
		parent, err := s.catRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent category not found: %w", err)
		}
		cat.Path = parent.Path + "/" + req.Name
	} else {
		cat.Path = "/" + req.Name
	}

	if err := s.catRepo.Create(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *categoryService) GetByID(ctx context.Context, id uint) (*model.Category, error) {
	cat, err := s.catRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}
	return cat, nil
}

func (s *categoryService) List(ctx context.Context, keyword string) ([]model.Category, error) {
	return s.catRepo.List(ctx, keyword)
}

func (s *categoryService) Update(ctx context.Context, id uint, req *model.UpdateCategoryRequest) (*model.Category, error) {
	cat, err := s.catRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	if req.Name != "" {
		cat.Name = req.Name
	}
	if req.Description != "" {
		cat.Description = req.Description
	}
	if req.ParentID != nil {
		cat.ParentID = req.ParentID
		parent, err := s.catRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent category not found: %w", err)
		}
		cat.Path = parent.Path + "/" + cat.Name
	}
	if req.SortOrder != 0 {
		cat.SortOrder = req.SortOrder
	}
	if req.ManagerID != nil {
		cat.ManagerID = req.ManagerID
	}

	exists, err := s.catRepo.ExistsByNameAndParent(ctx, cat.Name, cat.ParentID, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("category name already exists under the same parent")
	}

	if err := s.catRepo.Update(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *categoryService) Delete(ctx context.Context, id uint) error {
	return s.catRepo.Delete(ctx, id)
}
