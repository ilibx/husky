package repository

import (
	"context"

	"github.com/husky/husky/internal/model"
)

// CreateAuditLog 创建审计日志
func (r *TicketRepository) CreateAuditLog(ctx context.Context, log *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// ListAuditLogs 查询审计日志
func (r *TicketRepository) ListAuditLogs(ctx context.Context, resource string, resourceID uint, offset, limit int) ([]model.AuditLog, int64, error) {
	var list []model.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&model.AuditLog{})
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if resourceID > 0 {
		query = query.Where("resource_id = ?", resourceID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Preload("User").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// CreateWebhookMessage 保存渠道消息记录
func (r *TicketRepository) CreateWebhookMessage(ctx context.Context, msg *model.WebhookMessageRecord) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

// ListWebhookMessages 获取渠道消息记录
func (r *TicketRepository) ListWebhookMessages(ctx context.Context, offset, limit int, channel string) ([]model.WebhookMessageRecord, int64, error) {
	var list []model.WebhookMessageRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&model.WebhookMessageRecord{})
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
