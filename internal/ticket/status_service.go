package ticket

import (
	"context"
	"fmt"
	"time"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
)

func (s *service) UpdateStatus(ctx context.Context, id uint, status string) error {
	ticket, err := s.ticketRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !model.IsValidTransition(model.TicketStatus(ticket.Status), model.TicketStatus(status)) {
		return errors.New(errors.ErrInvalidTransition, fmt.Sprintf("cannot transition from %s to %s", ticket.Status, status))
	}

	oldStatus := ticket.Status
	now := time.Now()

	switch model.TicketStatus(status) {
	case model.TicketStatusResolved:
		ticket.ResolvedAt = &now
	case model.TicketStatusClosed:
		ticket.ClosedAt = &now
	}
	if model.TicketStatus(oldStatus) == model.TicketStatusResolved && status != string(model.TicketStatusResolved) {
		ticket.ResolvedAt = nil
	}

	ticket.Status = status
	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return err
	}

	s.logAudit(ctx, 0, "update_status", "ticket", ticket.ID,
		map[string]string{"status": oldStatus},
		map[string]string{"status": status})

	if ticket.AssigneeID != nil {
		s.CreateNotification(ctx, *ticket.AssigneeID, "ticket_status",
			"Status updated: "+ticket.Title,
			"#"+ticket.TicketNo+" status changed from "+oldStatus+" to "+status,
			ticket.ID, "ticket")
	}

	watchers, _ := s.ticketRepo.GetWatchers(ctx, ticket.ID)
	for _, w := range watchers {
		if ticket.AssigneeID != nil && w.UserID == *ticket.AssigneeID {
			continue
		}
		s.CreateNotification(ctx, w.UserID, "ticket_status",
			"Ticket status updated: "+ticket.Title,
			"#"+ticket.TicketNo+" status changed from "+oldStatus+" to "+status,
			ticket.ID, "ticket")
	}

	return nil
}

func (s *service) SetDueAt(ctx context.Context, id uint, dueAt time.Time) error {
	if err := s.ticketRepo.SetTicketDueAt(ctx, id, dueAt); err != nil {
		return err
	}

	ticket, _ := s.ticketRepo.GetByID(ctx, id)
	if ticket != nil && ticket.AssigneeID != nil {
		s.CreateNotification(ctx, *ticket.AssigneeID, "ticket_status",
			"SLA deadline set: "+ticket.Title,
			"#"+ticket.TicketNo+" must be resolved by "+dueAt.Format("2006-01-02 15:04"),
			ticket.ID, "ticket")
	}
	return nil
}

func (s *service) ListOverdue(ctx context.Context) ([]model.Ticket, error) {
	return s.ticketRepo.FindOverdueTickets(ctx, time.Now())
}
