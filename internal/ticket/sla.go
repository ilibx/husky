package ticket

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
)

// SLAEscalator monitors ticket SLA deadlines and escalates breached tickets.
type SLAEscalator struct {
	ticketRepo *repository.TicketRepository
	interval   time.Duration
}

func NewSLAEscalator(ticketRepo *repository.TicketRepository) *SLAEscalator {
	return &SLAEscalator{
		ticketRepo: ticketRepo,
		interval:   5 * time.Minute,
	}
}

func (e *SLAEscalator) SetInterval(d time.Duration) {
	e.interval = d
}

// Start begins the SLA monitoring loop in a background goroutine.
func (e *SLAEscalator) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(e.interval)
		defer ticker.Stop()

		log.Printf("SLA escalator started (interval: %v)", e.interval)

		for {
			select {
			case <-ctx.Done():
				log.Println("SLA escalator stopped")
				return
			case <-ticker.C:
				e.checkAndEscalate(context.Background())
			}
		}
	}()
}

func (e *SLAEscalator) checkAndEscalate(ctx context.Context) {
	overdue, err := e.ticketRepo.FindOverdueTickets(ctx, time.Now())
	if err != nil {
		log.Printf("SLA escalator: failed to query overdue tickets: %v", err)
		return
	}

	for _, t := range overdue {
		if err := e.escalateTicket(ctx, &t); err != nil {
			log.Printf("SLA escalator: failed to escalate ticket %d: %v", t.ID, err)
		}
	}
}

func (e *SLAEscalator) escalateTicket(ctx context.Context, ticket *model.Ticket) error {
	if ticket.DueAt == nil {
		return nil
	}

	now := time.Now()
	isWarning := ticket.SLAStatus == "normal" && now.After(ticket.DueAt.Add(-30*time.Minute))
	isBreached := now.After(*ticket.DueAt)

	var newStatus string
	switch {
	case isBreached:
		newStatus = "breached"
	case isWarning:
		newStatus = "warning"
	default:
		return nil
	}

	if err := e.ticketRepo.UpdateFields(ctx, ticket.ID, map[string]interface{}{
		"sla_status": newStatus,
	}); err != nil {
		return fmt.Errorf("failed to update SLA status: %w", err)
	}

	ticket.SLAStatus = newStatus

	if isBreached || isWarning {
		e.createSLAAlert(ctx, ticket, newStatus)
	}

	return nil
}

func (e *SLAEscalator) createSLAAlert(ctx context.Context, ticket *model.Ticket, slaType string) {
	title := fmt.Sprintf("SLA %s: %s", slaType, ticket.TicketNo)
	content := fmt.Sprintf("工单 #%s（%s）SLA %s，请及时处理。", ticket.TicketNo, ticket.Title, slaType)

	if slaType == "breached" {
		content = fmt.Sprintf("工单 #%s（%s）SLA 已超期，请立即处理。", ticket.TicketNo, ticket.Title)
	}

	if ticket.AssigneeID != nil {
		e.createNotification(ctx, *ticket.AssigneeID, title, content, ticket.ID)
	}

	e.createNotification(ctx, ticket.RequesterID, title,
		fmt.Sprintf("工单 #%s SLA %s", ticket.TicketNo, slaType), ticket.ID)
}

func (e *SLAEscalator) createNotification(ctx context.Context, userID uint, title, content string, ticketID uint) {
	if err := e.ticketRepo.CreateNotification(ctx, &model.Notification{
		UserID:        userID,
		Type:          "sla",
		Title:         title,
		Content:       content,
		ReferenceID:   ticketID,
		ReferenceType: "ticket",
		Status:        "unread",
	}); err != nil {
		log.Printf("SLA escalator: failed to create notification: %v", err)
	}
}
