package ticket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
	"github.com/husky/husky/pkg/errors"
)

type service struct {
	ticketRepo    *repository.TicketRepository
	auditSvc      *AuditService
	eventHandlers []TicketCreatedHandler
}

type TicketCreatedHandler func(ctx context.Context, ticket *model.Ticket)

func NewService(ticketRepo *repository.TicketRepository) Service {
	return &service{
		ticketRepo: ticketRepo,
		auditSvc:   NewAuditService(ticketRepo),
	}
}

// OnTicketCreated registers a post-creation handler.
// The handler runs in a goroutine after each successful ticket creation.
func (s *service) OnTicketCreated(h TicketCreatedHandler) {
	s.eventHandlers = append(s.eventHandlers, h)
}

func (s *service) CreateTicket(ctx context.Context, req *model.CreateTicketRequest) (*model.TicketResponse, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	if req.Priority != "" && !model.IsValidPriority(req.Priority) {
		return nil, fmt.Errorf("invalid priority: %s", req.Priority)
	}

	ticketNo := fmt.Sprintf("TK%s%s", time.Now().Format("20060102"), strings.ToUpper(uuid.New().String()[:6]))

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	ticket := &model.Ticket{
		TicketNo:    ticketNo,
		Title:       req.Title,
		Description: req.Description,
		Priority:    priority,
		Status:      string(model.TicketStatusOpen),
		Source:      req.Channel,
	}

	if req.RequesterID != "" {
		if id, err := strconv.ParseUint(req.RequesterID, 10, 64); err == nil {
			ticket.RequesterID = uint(id)
		}
	}

	if err := s.ticketRepo.Create(ctx, ticket); err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	s.logAudit(ctx, ticket.RequesterID, "create", "ticket", ticket.ID, nil, ticket)

	for _, h := range s.eventHandlers {
		handler := h
		go handler(context.WithoutCancel(ctx), ticket)
	}

	return &model.TicketResponse{
		ID:          ticket.ID,
		TicketNo:    ticket.TicketNo,
		Title:       ticket.Title,
		Status:      ticket.Status,
		Priority:    ticket.Priority,
		Source:      ticket.Source,
		CreatedAt:   ticket.CreatedAt,
		UpdatedAt:   ticket.UpdatedAt,
		RequesterID: ticket.RequesterID,
		AssigneeID:  ticket.AssigneeID,
	}, nil
}

func (s *service) GetTicket(ctx context.Context, id uint) (*model.Ticket, error) {
	return s.ticketRepo.GetByID(ctx, id)
}

func (s *service) ListTickets(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]model.Ticket, int64, error) {
	return s.ticketRepo.List(ctx, offset, limit, filters)
}

func (s *service) UpdateTicket(ctx context.Context, ticket *model.Ticket) error {
	existing, err := s.ticketRepo.GetByID(ctx, ticket.ID)
	if err != nil {
		return err
	}

	if ticket.Status != "" && ticket.Status != existing.Status {
		if !model.IsValidTransition(model.TicketStatus(existing.Status), model.TicketStatus(ticket.Status)) {
			return errors.New(errors.ErrInvalidTransition, fmt.Sprintf("cannot transition from %s to %s", existing.Status, ticket.Status))
		}
	}
	if ticket.Priority != "" && !model.IsValidPriority(ticket.Priority) {
		return fmt.Errorf("invalid priority: %s", ticket.Priority)
	}

	return s.ticketRepo.Update(ctx, ticket)
}

func (s *service) DeleteTicket(ctx context.Context, id uint) error {
	return s.ticketRepo.Delete(ctx, id)
}













func (s *service) logAudit(ctx context.Context, userID uint, action, resource string, resourceID uint, oldVal, newVal interface{}) {
	oldJSON, err := json.Marshal(oldVal)
	if err != nil {
		log.Printf("logAudit: failed to marshal oldVal: %v", err)
	}
	newJSON, err := json.Marshal(newVal)
	if err != nil {
		log.Printf("logAudit: failed to marshal newVal: %v", err)
	}
	audit := &model.AuditLog{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		OldValue:   string(oldJSON),
		NewValue:   string(newJSON),
	}
	if err := s.ticketRepo.CreateAuditLog(ctx, audit); err != nil {
		log.Printf("logAudit: failed to create audit log: %v", err)
	}
}



