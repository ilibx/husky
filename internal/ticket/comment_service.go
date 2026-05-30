package ticket

import (
	"context"
	"fmt"

	"github.com/husky/husky/internal/model"
)

func (s *service) AddComment(ctx context.Context, ticketID, userID uint, content string, isInternal bool) (*model.Comment, error) {
	if content == "" {
		return nil, fmt.Errorf("comment content is required")
	}

	comment := &model.Comment{
		TicketID:   ticketID,
		UserID:     userID,
		Content:    content,
		IsInternal: isInternal,
	}

	if err := s.ticketRepo.AddComment(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}

	ticket, _ := s.ticketRepo.GetByID(ctx, ticketID)
	if ticket != nil {
		if ticket.AssigneeID != nil && *ticket.AssigneeID != userID {
			s.CreateNotification(ctx, *ticket.AssigneeID, "ticket_comment",
				"New comment on: "+ticket.Title,
				content,
				ticket.ID, "ticket")
		}
		watchers, _ := s.ticketRepo.GetWatchers(ctx, ticketID)
		for _, w := range watchers {
			if w.UserID != userID && (ticket.AssigneeID == nil || w.UserID != *ticket.AssigneeID) {
				s.CreateNotification(ctx, w.UserID, "ticket_comment",
					"New comment on: "+ticket.Title,
					content,
					ticket.ID, "ticket")
			}
		}
	}

	return comment, nil
}

func (s *service) ListComments(ctx context.Context, ticketID uint, offset, limit int) ([]model.Comment, int64, error) {
	return s.ticketRepo.ListComments(ctx, ticketID, offset, limit)
}
