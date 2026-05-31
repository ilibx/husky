package ticket

import (
	"context"
	"log"
	"time"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
)

// SLAEscalator monitors ticket SLA deadlines and escalates breached tickets.
type SLAEscalator struct {
	ticketRepo *repository.TicketRepository
	slaSvc     *SLAConfigService
	interval   time.Duration
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewSLAEscalator(ticketRepo *repository.TicketRepository, slaSvc *SLAConfigService) *SLAEscalator {
	return &SLAEscalator{
		ticketRepo: ticketRepo,
		slaSvc:     slaSvc,
		interval:   5 * time.Minute,
	}
}

func (e *SLAEscalator) SetInterval(d time.Duration) {
	e.interval = d
}

// Start begins the SLA monitoring loop in a background goroutine.
func (e *SLAEscalator) Start(ctx context.Context) {
	e.ctx, e.cancel = context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(e.interval)
		defer ticker.Stop()

		log.Printf("SLA escalator started (interval: %v)", e.interval)

		for {
			select {
			case <-e.ctx.Done():
				log.Println("SLA escalator stopped")
				return
			case <-ticker.C:
				e.checkAndEscalate(e.ctx)
			}
		}
	}()
}

// Stop gracefully stops the SLA escalator.
func (e *SLAEscalator) Stop() {
	if e.cancel != nil {
		e.cancel()
	}
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
	isBreached := now.After(*ticket.DueAt)

	var resolutionMinutes float64 = 480
	var warningThreshold float64 = 0.8
	if e.slaSvc != nil {
		if cfg, err := e.slaSvc.FindMatch(ctx, ticket.Priority, &ticket.CategoryID); err == nil && cfg != nil {
			resolutionMinutes = float64(cfg.ResolutionMinutes)
			warningThreshold = cfg.WarningThreshold
		}
	}

	warningMinutes := resolutionMinutes * (1 - warningThreshold)
	isWarning := ticket.SLAStatus == "normal" && now.After(ticket.DueAt.Add(-time.Duration(warningMinutes)*time.Minute))

	var newStatus string
	var eventType string
	switch {
	case isBreached:
		newStatus = "breached"
		eventType = "sla_breach"
	case isWarning:
		newStatus = "warning"
		eventType = "sla_warning"
	default:
		return nil
	}

	if err := e.ticketRepo.UpdateFields(ctx, ticket.ID, map[string]interface{}{
		"sla_status": newStatus,
	}); err != nil {
		return err
	}

	ticket.SLAStatus = newStatus

	// Fire SLA event via service (notification + webhook + channel)
	if e.slaSvc != nil {
		e.slaSvc.FireSLAEvent(ctx, ticket, eventType)
	}

	return nil
}

// ComputeAndSetDueAt 根据 SLA 配置计算并设置工单 DueAt
func (e *SLAEscalator) ComputeAndSetDueAt(ctx context.Context, ticket *model.Ticket) {
	if e.slaSvc == nil {
		return
	}
	dueAt := e.slaSvc.ComputeDueAt(ctx, ticket)
	if dueAt != nil {
		if err := e.ticketRepo.UpdateFields(ctx, ticket.ID, map[string]interface{}{
			"due_at": dueAt,
		}); err != nil {
			log.Printf("SLA: failed to set due_at for ticket %d: %v", ticket.ID, err)
		}
	}
}
