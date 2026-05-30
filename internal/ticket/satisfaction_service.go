package ticket

import (
	"context"
	"fmt"

	"github.com/husky/husky/internal/model"
)

func (s *service) RateTicket(ctx context.Context, ticketID, userID uint, score int, comment string) (*model.Satisfaction, error) {
	if score < 1 || score > 5 {
		return nil, fmt.Errorf("score must be between 1 and 5")
	}

	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket.Status != string(model.TicketStatusResolved) && ticket.Status != string(model.TicketStatusClosed) {
		return nil, fmt.Errorf("can only rate resolved or closed tickets")
	}

	existing, _ := s.ticketRepo.GetSatisfactionByTicket(ctx, ticketID)
	if existing != nil {
		return nil, fmt.Errorf("ticket already rated")
	}

	sat := &model.Satisfaction{
		TicketID:  ticketID,
		Score:     score,
		Comment:   comment,
		CreatedBy: userID,
	}
	if err := s.ticketRepo.CreateSatisfaction(ctx, sat); err != nil {
		return nil, err
	}
	return sat, nil
}

func (s *service) GetSatisfaction(ctx context.Context, ticketID uint) (*model.Satisfaction, error) {
	return s.ticketRepo.GetSatisfactionByTicket(ctx, ticketID)
}

func (s *service) GetSatisfactionStats(ctx context.Context) (map[string]interface{}, error) {
	return s.ticketRepo.GetSatisfactionStats(ctx)
}
