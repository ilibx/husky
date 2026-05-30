package repository

import (
	"context"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// --- SLA Config ---

type SLAConfigRepository struct {
	*BaseRepository
}

func NewSLAConfigRepository(db *gorm.DB) *SLAConfigRepository {
	return &SLAConfigRepository{NewBaseRepository(db)}
}

func (r *SLAConfigRepository) List(ctx context.Context) ([]model.SLAConfig, error) {
	var list []model.SLAConfig
	err := r.db.WithContext(ctx).Order("priority, category_id").Find(&list).Error
	return list, err
}

func (r *SLAConfigRepository) GetByID(ctx context.Context, id uint) (*model.SLAConfig, error) {
	var c model.SLAConfig
	err := r.db.WithContext(ctx).First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *SLAConfigRepository) Create(ctx context.Context, c *model.SLAConfig) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *SLAConfigRepository) Update(ctx context.Context, c *model.SLAConfig) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *SLAConfigRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.SLAConfig{}, id).Error
}

func (r *SLAConfigRepository) FindMatch(ctx context.Context, priority string, categoryID *uint) (*model.SLAConfig, error) {
	// Exact match: priority + category
	var c model.SLAConfig
	err := r.db.WithContext(ctx).
		Where("priority = ? AND category_id = ? AND enabled = ?", priority, categoryID, true).
		First(&c).Error
	if err == nil {
		return &c, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	// Fallback: priority only (global)
	err = r.db.WithContext(ctx).
		Where("priority = ? AND category_id IS NULL AND enabled = ?", priority, true).
		First(&c).Error
	if err == nil {
		return &c, nil
	}
	return nil, nil
}
