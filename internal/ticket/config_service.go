package ticket

import (
	"context"
	"fmt"

	"github.com/husky/husky/internal/model"
)

func (s *service) GetAssignConfig(ctx context.Context, categoryID *uint) (*model.AssignConfig, error) {
	return s.ticketRepo.GetAssignConfig(ctx, categoryID)
}

func (s *service) SetAssignConfig(ctx context.Context, cfg *model.AssignConfig) error {
	return s.ticketRepo.SetAssignConfig(ctx, cfg)
}

func (s *service) CreateTicketRelation(ctx context.Context, ticketID, relatedID uint, relationType string) (*model.TicketRelation, error) {
	if ticketID == relatedID {
		return nil, fmt.Errorf("cannot relate a ticket to itself")
	}
	rel := &model.TicketRelation{
		TicketID:     ticketID,
		RelatedID:    relatedID,
		RelationType: relationType,
	}
	if err := s.ticketRepo.CreateTicketRelation(ctx, rel); err != nil {
		return nil, err
	}
	return rel, nil
}

func (s *service) DeleteTicketRelation(ctx context.Context, id uint) error {
	return s.ticketRepo.DeleteTicketRelation(ctx, id)
}

func (s *service) ListTicketRelations(ctx context.Context, ticketID uint) ([]model.TicketRelation, error) {
	return s.ticketRepo.ListTicketRelations(ctx, ticketID)
}

func (s *service) ListRoles(ctx context.Context, keyword string) ([]model.Role, error) {
	return s.ticketRepo.ListRoles(ctx, keyword)
}

func (s *service) GetRole(ctx context.Context, id uint) (*model.Role, error) {
	return s.ticketRepo.GetRole(ctx, id)
}

func (s *service) CreateRole(ctx context.Context, role *model.Role) error {
	return s.ticketRepo.CreateRole(ctx, role)
}

func (s *service) UpdateRole(ctx context.Context, role *model.Role) error {
	return s.ticketRepo.UpdateRole(ctx, role)
}

func (s *service) DeleteRole(ctx context.Context, id uint) error {
	return s.ticketRepo.DeleteRole(ctx, id)
}




func (s *service) GetBotConfig(ctx context.Context, channel string) (*model.BotConfig, error) {
	return s.ticketRepo.GetBotConfig(ctx, channel)
}

func (s *service) SetBotConfig(ctx context.Context, cfg *model.BotConfig) error {
	return s.ticketRepo.SetBotConfig(ctx, cfg)
}
