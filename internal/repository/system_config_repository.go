package repository

import (
	"context"
	"fmt"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// SystemConfigRepository 系统动态配置 Repository
type SystemConfigRepository struct {
	db *gorm.DB
}

func NewSystemConfigRepository(db *gorm.DB) *SystemConfigRepository {
	return &SystemConfigRepository{db: db}
}

func (r *SystemConfigRepository) Create(ctx context.Context, cfg *model.SystemConfig) error {
	return r.db.WithContext(ctx).Create(cfg).Error
}

func (r *SystemConfigRepository) GetByID(ctx context.Context, id uint) (*model.SystemConfig, error) {
	var cfg model.SystemConfig
	err := r.db.WithContext(ctx).First(&cfg, id).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *SystemConfigRepository) GetByKey(ctx context.Context, category, key string) (*model.SystemConfig, error) {
	var cfg model.SystemConfig
	err := r.db.WithContext(ctx).
		Where("category = ? AND key = ?", category, key).
		First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ListByCategory 按分类获取所有启用配置
func (r *SystemConfigRepository) ListByCategory(ctx context.Context, category string) ([]model.SystemConfig, error) {
	var configs []model.SystemConfig
	err := r.db.WithContext(ctx).
		Where("category = ? AND enabled = ?", category, true).
		Find(&configs).Error
	return configs, err
}

// ListAll 获取所有配置
func (r *SystemConfigRepository) ListAll(ctx context.Context, offset, limit int) ([]model.SystemConfig, int64, error) {
	var configs []model.SystemConfig
	var total int64

	if err := r.db.WithContext(ctx).Model(&model.SystemConfig{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Offset(offset).Limit(limit).
		Order("category ASC, key ASC").
		Find(&configs).Error
	return configs, total, err
}

func (r *SystemConfigRepository) Upsert(ctx context.Context, cfg *model.SystemConfig) error {
	var existing model.SystemConfig
	result := r.db.WithContext(ctx).
		Where("category = ? AND key = ?", cfg.Category, cfg.Key).
		First(&existing)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return r.db.WithContext(ctx).Create(cfg).Error
		}
		return result.Error
	}

	cfg.ID = existing.ID
	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"value":   cfg.Value,
		"enabled": cfg.Enabled,
	}).Error
}

func (r *SystemConfigRepository) Update(ctx context.Context, cfg *model.SystemConfig) error {
	return r.db.WithContext(ctx).Save(cfg).Error
}

func (r *SystemConfigRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.SystemConfig{}, id).Error
}

// GetLLMConfig 从 DB 聚合获取运行时 LLM 配置
func (r *SystemConfigRepository) GetLLMConfig(ctx context.Context) (*model.LLMConfig, error) {
	configs, err := r.ListByCategory(ctx, model.SysCfgCategoryLLM)
	if err != nil {
		return nil, fmt.Errorf("get LLM config: %w", err)
	}

	cfg := model.DefaultLLMConfig()
	for _, c := range configs {
		switch c.Key {
		case model.SysCfgLLMProvider:
			cfg.Provider = c.Value
		case model.SysCfgLLMAPIKey:
			cfg.APIKey = c.Value
		case model.SysCfgLLMBaseURL:
			cfg.BaseURL = c.Value
		case model.SysCfgLLMModel:
			cfg.Model = c.Value
		case model.SysCfgLLMMaxTokens:
			if v, err := fmt.Sscanf(c.Value, "%d", &cfg.MaxTokens); err != nil || v != 1 {
				cfg.MaxTokens = 4096
			}
		case model.SysCfgLLMTemperature:
			if v, err := fmt.Sscanf(c.Value, "%f", &cfg.Temperature); err != nil || v != 1 {
				cfg.Temperature = 0.7
			}
		}
	}
	return &cfg, nil
}

func (r *SystemConfigRepository) GetVectorConfig(ctx context.Context) (*model.VectorDBConfig, error) {
	configs, err := r.ListByCategory(ctx, model.SysCfgCategoryVector)
	if err != nil {
		return nil, fmt.Errorf("get vector config: %w", err)
	}

	cfg := model.DefaultVectorDBConfig()
	for _, c := range configs {
		switch c.Key {
		case model.SysCfgVectorProvider:
			cfg.Provider = c.Value
		case model.SysCfgVectorHost:
			cfg.Host = c.Value
		case model.SysCfgVectorPort:
			cfg.Port = c.Value
		case model.SysCfgVectorUser:
			cfg.User = c.Value
		case model.SysCfgVectorPassword:
			cfg.Password = c.Value
		case model.SysCfgVectorDatabase:
			cfg.Database = c.Value
		case model.SysCfgVectorSSLMode:
			cfg.SSLMode = c.Value
		}
	}
	return &cfg, nil
}
