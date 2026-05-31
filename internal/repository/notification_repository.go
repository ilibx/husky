package repository

import (
	"context"
	"github.com/husky/husky/internal/model"
)

// CreateNotification 创建通知
func (r *TicketRepository) CreateNotification(ctx context.Context, n *model.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

// ListNotifications 获取通知列表
// userID=0 返回所有通知（管理员）, 否则返回指定用户的通知
func (r *TicketRepository) ListNotifications(ctx context.Context, userID uint, offset, limit int, keyword string) ([]model.Notification, int64, error) {
	var list []model.Notification
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Notification{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if keyword != "" {
		query = query.Where("title ILIKE ?", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// GetUnreadNotificationCount 获取未读通知数
func (r *TicketRepository) GetUnreadNotificationCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("is_read = ?", false)
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	err := query.Count(&count).Error
	return count, err
}

// MarkNotificationRead 标记通知为已读
func (r *TicketRepository) MarkNotificationRead(ctx context.Context, id, userID uint) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

// MarkAllNotificationsRead 标记所有通知为已读
func (r *TicketRepository) MarkAllNotificationsRead(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}
