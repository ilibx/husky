package repository

import (
	"context"
	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// GetBotConfig 获取机器人配置
func (r *TicketRepository) GetBotConfig(ctx context.Context, channel string) (*model.BotConfig, error) {
	var cfg model.BotConfig
	err := r.db.WithContext(ctx).Where("channel = ?", channel).First(&cfg).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

// SetBotConfig 设置机器人配置
func (r *TicketRepository) SetBotConfig(ctx context.Context, cfg *model.BotConfig) error {
	var existing model.BotConfig
	if err := r.db.WithContext(ctx).Where("channel = ?", cfg.Channel).First(&existing).Error; err == nil {
		cfg.ID = existing.ID
		cfg.CreatedAt = existing.CreatedAt
		return r.db.WithContext(ctx).Save(cfg).Error
	}
	return r.db.WithContext(ctx).Create(cfg).Error
}
