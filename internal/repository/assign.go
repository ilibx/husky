package repository

import (
	"context"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// GetAssignConfig 获取分配配置（按分类，无则返回全局）
func (r *TicketRepository) GetAssignConfig(ctx context.Context, categoryID *uint) (*model.AssignConfig, error) {
	var cfg model.AssignConfig
	query := r.db.WithContext(ctx).Model(&model.AssignConfig{})
	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	} else {
		query = query.Where("category_id IS NULL")
	}
	err := query.First(&cfg).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

// SetAssignConfig 设置分配配置
func (r *TicketRepository) SetAssignConfig(ctx context.Context, cfg *model.AssignConfig) error {
	query := r.db.WithContext(ctx).Model(&model.AssignConfig{})
	if cfg.CategoryID != nil {
		query = query.Where("category_id = ?", *cfg.CategoryID)
	} else {
		query = query.Where("category_id IS NULL")
	}
	var existing model.AssignConfig
	if err := query.First(&existing).Error; err == nil {
		cfg.ID = existing.ID
		cfg.CreatedAt = existing.CreatedAt
		return r.db.WithContext(ctx).Save(cfg).Error
	}
	return r.db.WithContext(ctx).Create(cfg).Error
}
