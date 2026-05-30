package ticket

import (
	"context"
	"log"

	"github.com/husky/husky/internal/model"
)

func (s *service) WatchTicket(ctx context.Context, ticketID, userID uint) error {
	return s.ticketRepo.AddWatcher(ctx, ticketID, userID)
}

func (s *service) UnwatchTicket(ctx context.Context, ticketID, userID uint) error {
	return s.ticketRepo.RemoveWatcher(ctx, ticketID, userID)
}

func (s *service) GetWatchers(ctx context.Context, ticketID uint) ([]model.TicketWatcher, error) {
	return s.ticketRepo.GetWatchers(ctx, ticketID)
}

func (s *service) SaveWebhookMessage(ctx context.Context, msg *model.WebhookMessageRecord) error {
	return s.ticketRepo.CreateWebhookMessage(ctx, msg)
}

func (s *service) ListAuditLogs(ctx context.Context, resourceType string, resourceID uint, offset, limit int) ([]model.AuditLog, int64, error) {
	return s.ticketRepo.ListAuditLogs(ctx, resourceType, resourceID, offset, limit)
}

func (s *service) ListWebhookMessages(ctx context.Context, offset, limit int, channel string) ([]model.WebhookMessageRecord, int64, error) {
	return s.ticketRepo.ListWebhookMessages(ctx, offset, limit, channel)
}

func (s *service) CreateNotification(ctx context.Context, userID uint, nType, title, content string, refID uint, refType string) error {
	n := &model.Notification{
		UserID:        userID,
		Type:          nType,
		Title:         title,
		Content:       content,
		ReferenceID:   refID,
		ReferenceType: refType,
		Status:        "unread",
	}
	if err := s.ticketRepo.CreateNotification(ctx, n); err != nil {
		log.Printf("failed to create notification (user=%d, type=%s): %v", userID, nType, err)
		return err
	}
	return nil
}

func (s *service) ListNotifications(ctx context.Context, userID uint, offset, limit int) ([]model.Notification, int64, error) {
	return s.ticketRepo.ListNotifications(ctx, userID, offset, limit)
}

func (s *service) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	return s.ticketRepo.GetUnreadNotificationCount(ctx, userID)
}

func (s *service) MarkNotificationRead(ctx context.Context, id, userID uint) error {
	return s.ticketRepo.MarkNotificationRead(ctx, id, userID)
}

func (s *service) MarkAllNotificationsRead(ctx context.Context, userID uint) error {
	return s.ticketRepo.MarkAllNotificationsRead(ctx, userID)
}
