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

func (s *service) ListRoles(ctx context.Context) ([]model.Role, error) {
	return s.ticketRepo.ListRoles(ctx)
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

func (s *service) ListTicketFields(ctx context.Context) ([]model.TicketField, error) {
	return s.ticketRepo.ListTicketFields(ctx)
}

func (s *service) CreateTicketField(ctx context.Context, req *model.CreateTicketFieldRequest) (*model.TicketField, error) {
	field := &model.TicketField{
		Name:        req.Name,
		FieldKey:    req.FieldKey,
		FieldType:   req.FieldType,
		Options:     req.Options,
		Required:    req.Required,
		SortOrder:   req.SortOrder,
		Placeholder: req.Placeholder,
	}
	if err := s.ticketRepo.CreateTicketField(ctx, field); err != nil {
		return nil, err
	}
	return field, nil
}

func (s *service) UpdateTicketField(ctx context.Context, id uint, req *model.UpdateTicketFieldRequest) (*model.TicketField, error) {
	existing, err := s.ticketRepo.GetTicketField(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("ticket field not found")
	}
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.FieldType != nil {
		existing.FieldType = *req.FieldType
	}
	if req.Options != nil {
		existing.Options = *req.Options
	}
	if req.Required != nil {
		existing.Required = *req.Required
	}
	if req.SortOrder != nil {
		existing.SortOrder = *req.SortOrder
	}
	if req.Placeholder != nil {
		existing.Placeholder = *req.Placeholder
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}
	if err := s.ticketRepo.UpdateTicketField(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *service) DeleteTicketField(ctx context.Context, id uint) error {
	return s.ticketRepo.DeleteTicketField(ctx, id)
}

func (s *service) UpdateTicketFieldValues(ctx context.Context, ticketID uint, values []model.TicketFieldValue) error {
	return s.ticketRepo.SetTicketFieldValues(ctx, ticketID, values)
}

func (s *service) GetTicketFieldValues(ctx context.Context, ticketID uint) ([]model.TicketFieldValue, error) {
	return s.ticketRepo.GetTicketFieldValues(ctx, ticketID)
}

func (s *service) GetBotConfig(ctx context.Context, channel string) (*model.BotConfig, error) {
	return s.ticketRepo.GetBotConfig(ctx, channel)
}

func (s *service) SetBotConfig(ctx context.Context, cfg *model.BotConfig) error {
	return s.ticketRepo.SetBotConfig(ctx, cfg)
}
