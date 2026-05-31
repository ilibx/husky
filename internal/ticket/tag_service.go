package ticket

import (
	"context"
	"fmt"

	"github.com/husky/husky/internal/model"
)

func (s *service) CreateTag(ctx context.Context, name, color string) (*model.Tag, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	tag := &model.Tag{Name: name, Color: color}
	if err := s.ticketRepo.CreateTag(ctx, tag); err != nil {
		return nil, err
	}
	return tag, nil
}

func (s *service) ListTags(ctx context.Context, keyword string) ([]model.Tag, error) {
	return s.ticketRepo.ListTags(ctx, keyword)
}

func (s *service) GetTag(ctx context.Context, id uint) (*model.Tag, error) {
	return s.ticketRepo.GetTag(ctx, id)
}

func (s *service) UpdateTag(ctx context.Context, tag *model.Tag) error {
	if tag.ID == 0 {
		return fmt.Errorf("id is required")
	}
	return s.ticketRepo.UpdateTag(ctx, tag)
}

func (s *service) DeleteTag(ctx context.Context, id uint) error {
	return s.ticketRepo.DeleteTag(ctx, id)
}

func (s *service) AddTagsToTicket(ctx context.Context, ticketID uint, tagIDs []uint) error {
	return s.ticketRepo.AddTagsToTicket(ctx, ticketID, tagIDs)
}

func (s *service) RemoveTagFromTicket(ctx context.Context, ticketID, tagID uint) error {
	return s.ticketRepo.RemoveTagFromTicket(ctx, ticketID, tagID)
}

func (s *service) GetTicketTags(ctx context.Context, ticketID uint) ([]model.Tag, error) {
	return s.ticketRepo.GetTicketTags(ctx, ticketID)
}

func (s *service) UpdateTicketTags(ctx context.Context, ticketID uint, tagIDs []uint) error {
	return s.ticketRepo.UpdateTicketTags(ctx, ticketID, tagIDs)
}
