package repository

import (
	"context"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// --- Webhook Config ---

type WebhookConfigRepository struct {
	*BaseRepository
}

func NewWebhookConfigRepository(db *gorm.DB) *WebhookConfigRepository {
	return &WebhookConfigRepository{NewBaseRepository(db)}
}

func (r *WebhookConfigRepository) List(ctx context.Context) ([]model.WebhookConfig, error) {
	var list []model.WebhookConfig
	err := r.db.WithContext(ctx).Order("name").Find(&list).Error
	return list, err
}

func (r *WebhookConfigRepository) GetByID(ctx context.Context, id uint) (*model.WebhookConfig, error) {
	var c model.WebhookConfig
	err := r.db.WithContext(ctx).First(&c, id).Error
	return &c, err
}

func (r *WebhookConfigRepository) Create(ctx context.Context, c *model.WebhookConfig) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *WebhookConfigRepository) Update(ctx context.Context, c *model.WebhookConfig) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *WebhookConfigRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.WebhookConfig{}, id).Error
}

func (r *WebhookConfigRepository) FindByEvent(ctx context.Context, event string) ([]model.WebhookConfig, error) {
	var list []model.WebhookConfig
	err := r.db.WithContext(ctx).
		Where("events LIKE ? AND enabled = ?", "%"+event+"%", true).
		Find(&list).Error
	return list, err
}
