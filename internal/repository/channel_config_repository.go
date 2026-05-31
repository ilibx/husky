package repository

import (
	"context"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// ChannelConfigRepository 渠道配置 Repository
type ChannelConfigRepository struct {
	*BaseRepository
}

func NewChannelConfigRepository(db *gorm.DB) *ChannelConfigRepository {
	return &ChannelConfigRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *ChannelConfigRepository) Create(ctx context.Context, cfg *model.ChannelConfig) error {
	return r.db.WithContext(ctx).Create(cfg).Error
}

func (r *ChannelConfigRepository) GetByID(ctx context.Context, id uint) (*model.ChannelConfig, error) {
	var cfg model.ChannelConfig
	err := r.db.WithContext(ctx).First(&cfg, id).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *ChannelConfigRepository) List(ctx context.Context, keyword string) ([]model.ChannelConfig, error) {
	var list []model.ChannelConfig
	query := r.db.WithContext(ctx).Order("created_at DESC")
	if keyword != "" {
		query = query.Where("name ILIKE ?", "%"+keyword+"%")
	}
	err := query.Find(&list).Error
	return list, err
}

func (r *ChannelConfigRepository) Update(ctx context.Context, cfg *model.ChannelConfig) error {
	return r.db.WithContext(ctx).Save(cfg).Error
}

func (r *ChannelConfigRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.ChannelConfig{}, id).Error
}
