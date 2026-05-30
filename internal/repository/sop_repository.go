package repository

import (
	"context"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// SOPRepository SOP Repository
type SOPRepository struct {
	*BaseRepository
}

func NewSOPRepository(db *gorm.DB) *SOPRepository {
	return &SOPRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *SOPRepository) Create(ctx context.Context, sop *model.SOP) error {
	return r.db.WithContext(ctx).Create(sop).Error
}

func (r *SOPRepository) GetByID(ctx context.Context, id uint) (*model.SOP, error) {
	var sop model.SOP
	err := r.db.WithContext(ctx).Preload("Creator").First(&sop, id).Error
	if err != nil {
		return nil, err
	}
	return &sop, nil
}

func (r *SOPRepository) List(ctx context.Context, offset, limit int) ([]model.SOP, int64, error) {
	var list []model.SOP
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.SOP{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// ListActive 查询所有已启用的 SOP
func (r *SOPRepository) ListActive(ctx context.Context) ([]model.SOP, error) {
	var list []model.SOP
	err := r.db.WithContext(ctx).Where("status = ?", 1).Order("id ASC").Find(&list).Error
	return list, err
}

func (r *SOPRepository) Update(ctx context.Context, sop *model.SOP) error {
	return r.db.WithContext(ctx).Save(sop).Error
}

func (r *SOPRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.SOP{}, id).Error
}
