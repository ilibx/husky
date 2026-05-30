package service

import (
	"context"

	"github.com/husky/husky/internal/repository"
)

// StatsService 统计服务接口
type StatsService interface {
	GetOverview(ctx context.Context) (map[string]interface{}, error)
	GetTicketStats(ctx context.Context) (map[string]interface{}, error)
	GetSatisfactionStats(ctx context.Context) (map[string]interface{}, error)
	GetAgentPerformance(ctx context.Context) ([]repository.AgentPerformance, error)
}

type statsService struct {
	ticketRepo *repository.TicketRepository
}

// NewStatsService 创建统计服务实例
func NewStatsService(ticketRepo *repository.TicketRepository) StatsService {
	return &statsService{ticketRepo: ticketRepo}
}

func (s *statsService) GetOverview(ctx context.Context) (map[string]interface{}, error) {
	return s.ticketRepo.GetStatsOverview(ctx)
}

func (s *statsService) GetTicketStats(ctx context.Context) (map[string]interface{}, error) {
	return s.ticketRepo.GetTicketStats(ctx)
}

func (s *statsService) GetSatisfactionStats(ctx context.Context) (map[string]interface{}, error) {
	return s.ticketRepo.GetSatisfactionStats(ctx)
}

func (s *statsService) GetAgentPerformance(ctx context.Context) ([]repository.AgentPerformance, error) {
	return s.ticketRepo.GetAgentPerformance(ctx)
}
