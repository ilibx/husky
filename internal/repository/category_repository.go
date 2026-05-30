package repository

import (
	"context"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// CategoryRepository 分类 Repository
type CategoryRepository struct {
	*BaseRepository
}

// NewCategoryRepository 创建分类 Repository
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *CategoryRepository) Create(ctx context.Context, cat *model.Category) error {
	return r.db.WithContext(ctx).Create(cat).Error
}

func (r *CategoryRepository) GetByID(ctx context.Context, id uint) (*model.Category, error) {
	var cat model.Category
	err := r.db.WithContext(ctx).Preload("Parent").First(&cat, id).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *CategoryRepository) List(ctx context.Context) ([]model.Category, error) {
	var list []model.Category
	err := r.db.WithContext(ctx).Order("sort_order ASC, id ASC").Find(&list).Error
	return list, err
}

// ListByType 按类型查询分类
func (r *CategoryRepository) ListByType(ctx context.Context, categoryType string) ([]model.Category, error) {
	var list []model.Category
	query := r.db.WithContext(ctx).Order("sort_order ASC, id ASC")
	if categoryType != "" {
		query = query.Where("type = ? OR type = 'both'", categoryType)
	}
	err := query.Find(&list).Error
	return list, err
}

func (r *CategoryRepository) Update(ctx context.Context, cat *model.Category) error {
	return r.db.WithContext(ctx).Save(cat).Error
}

func (r *CategoryRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Category{}, id).Error
}
