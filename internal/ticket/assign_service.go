package ticket

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
)

func (s *service) AutoAssignTicket(ctx context.Context, id uint) (*uint, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	cfg, err := s.ticketRepo.GetAssignConfig(ctx, &ticket.CategoryID)
	if err != nil || cfg == nil {
		cfg = &model.AssignConfig{Strategy: model.AssignStrategyLeastBusy}
	}

	var agentID *uint
	switch cfg.Strategy {
	case model.AssignStrategyRoundRobin:
		ids, err := s.ticketRepo.ListAgentIDs(ctx)
		if err != nil || len(ids) == 0 {
			return nil, fmt.Errorf("no available agents found")
		}
		idx, _ := s.ticketRepo.IncrementRoundRobin(ctx, cfg.ID, len(ids))
		agentID = &ids[idx]

	case model.AssignStrategySkillBased:
		if ticket.CategoryID > 0 {
			ids, err := s.ticketRepo.FindAgentsBySkill(ctx, ticket.CategoryID)
			if err == nil && len(ids) > 0 {
				agentID = &ids[0]
				break
			}
		}
		fallthrough

	case model.AssignStrategyRandom:
		ids, err := s.ticketRepo.ListAgentIDs(ctx)
		if err != nil || len(ids) == 0 {
			return nil, fmt.Errorf("no available agents found")
		}
		agentID = &ids[rand.Intn(len(ids))]

	default: // least_busy
		id, err := s.ticketRepo.FindLeastBusyAgent(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to find available agent: %w", err)
		}
		agentID = id
	}

	if agentID == nil {
		return nil, fmt.Errorf("no available agents found")
	}
	if err := s.AssignTicket(ctx, id, *agentID); err != nil {
		return nil, err
	}
	return agentID, nil
}

func (s *service) AssignTicket(ctx context.Context, id, assigneeID uint) error {
	ticket, err := s.ticketRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	previousAssignee := ticket.AssigneeID
	ticket.AssigneeID = &assigneeID
	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return err
	}

	s.logAudit(ctx, assigneeID, "assign", "ticket", ticket.ID,
		map[string]interface{}{"assignee_id": previousAssignee},
		map[string]interface{}{"assignee_id": assigneeID})

	s.CreateNotification(ctx, assigneeID, "ticket_assigned",
		"New ticket assigned: "+ticket.Title,
		"#"+ticket.TicketNo+" has been assigned to you.\nPriority: "+ticket.Priority,
		ticket.ID, "ticket")
	return nil
}

func (s *service) ClaimTicket(ctx context.Context, id, userID uint) error {
	ticket, err := s.ticketRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if ticket.AssigneeID != nil && *ticket.AssigneeID != userID {
		return errors.New(errors.ErrForbidden, "ticket already assigned to another user")
	}
	previousAssignee := ticket.AssigneeID
	ticket.AssigneeID = &userID
	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return err
	}

	if previousAssignee != nil && *previousAssignee != userID {
		s.CreateNotification(ctx, *previousAssignee, "ticket_assigned",
			"Ticket re-assigned: "+ticket.Title,
			"#"+ticket.TicketNo+" has been re-assigned to another user.",
			ticket.ID, "ticket")
	}
	s.CreateNotification(ctx, userID, "ticket_assigned",
		"Ticket claimed: "+ticket.Title,
		"#"+ticket.TicketNo+" has been claimed by you.\nPriority: "+ticket.Priority,
		ticket.ID, "ticket")
	return nil
}
