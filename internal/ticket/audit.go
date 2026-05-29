package ticket

import (
	"context"
	"encoding/json"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
)

type AuditService struct {
	repo *repository.TicketRepository
}

func NewAuditService(repo *repository.TicketRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Log(ctx context.Context, userID uint, action, resource string, resourceID uint, oldVal, newVal interface{}) {
	var oldStr, newStr string
	if oldVal != nil {
		if b, err := json.Marshal(oldVal); err == nil {
			oldStr = string(b)
		}
	}
	if newVal != nil {
		if b, err := json.Marshal(newVal); err == nil {
			newStr = string(b)
		}
	}

	s.repo.CreateAuditLog(ctx, &model.AuditLog{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		OldValue:   oldStr,
		NewValue:   newStr,
	})
}

func (s *AuditService) LogTicketCreated(ctx context.Context, userID uint, ticket *model.Ticket) {
	s.Log(ctx, userID, "ticket.created", "ticket", ticket.ID, nil, map[string]interface{}{
		"title":    ticket.Title,
		"priority": ticket.Priority,
		"source":   ticket.Source,
	})
}

func (s *AuditService) LogStatusChange(ctx context.Context, userID uint, ticketID uint, oldStatus, newStatus string) {
	s.Log(ctx, userID, "ticket.status_changed", "ticket", ticketID,
		map[string]string{"status": oldStatus},
		map[string]string{"status": newStatus},
	)
}

func (s *AuditService) LogAssignChange(ctx context.Context, userID uint, ticketID uint, oldAssigneeID, newAssigneeID *uint) {
	var oldVal, newVal interface{}
	if oldAssigneeID != nil {
		oldVal = map[string]uint{"assignee_id": *oldAssigneeID}
	}
	if newAssigneeID != nil {
		newVal = map[string]uint{"assignee_id": *newAssigneeID}
	}
	s.Log(ctx, userID, "ticket.assigned", "ticket", ticketID, oldVal, newVal)
}

func (s *AuditService) LogCommentAdded(ctx context.Context, userID uint, ticketID uint, commentID uint, isInternal bool) {
	s.Log(ctx, userID, "ticket.comment_added", "ticket", ticketID, nil, map[string]interface{}{
		"comment_id":  commentID,
		"is_internal": isInternal,
	})
}

func (s *AuditService) LogPriorityChange(ctx context.Context, userID uint, ticketID uint, oldPriority, newPriority string) {
	s.Log(ctx, userID, "ticket.priority_changed", "ticket", ticketID,
		map[string]string{"priority": oldPriority},
		map[string]string{"priority": newPriority},
	)
}
