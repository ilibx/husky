package agent

import (
	"context"

	"github.com/husky/husky/internal/model"
)

func (s *WorkflowService) createNotification(ctx context.Context, userID uint, nType, title, content string, refID uint, refType string) {
	if userID == 0 {
		return
	}
	s.ticketRepo.CreateNotification(ctx, &model.Notification{
		UserID:        userID,
		Type:          nType,
		Title:         title,
		Content:       content,
		ReferenceID:   refID,
		ReferenceType: refType,
		Status:        "unread",
	})
}
