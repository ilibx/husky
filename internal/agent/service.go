package agent

import (
	"context"
	"fmt"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
)

type Service interface {
	Create(ctx context.Context, userID uint, req *model.CreateAgentRequest) (*model.Agent, error)
	GetByID(ctx context.Context, id uint) (*model.Agent, error)
	List(ctx context.Context, offset, limit int, keyword string) ([]model.Agent, int64, error)
	Update(ctx context.Context, id uint, req *model.UpdateAgentRequest) (*model.Agent, error)
	Delete(ctx context.Context, id uint) error
}

type service struct {
	agentRepo *repository.AgentRepository
}

func NewService(agentRepo *repository.AgentRepository) Service {
	return &service{agentRepo: agentRepo}
}

func (s *service) Create(ctx context.Context, userID uint, req *model.CreateAgentRequest) (*model.Agent, error) {
	if req.Name == "" || req.Type == "" {
		return nil, fmt.Errorf("name and type are required")
	}

	validTypes := map[string]bool{"llm": true, "rule": true, "hybrid": true}
	if !validTypes[req.Type] {
		return nil, fmt.Errorf("invalid agent type: %s (must be llm, rule, or hybrid)", req.Type)
	}

	agent := &model.Agent{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Config:      req.Config,
		Model:       req.Model,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Enabled:     true,
		CreatedBy:   userID,
	}

	if err := s.agentRepo.Create(ctx, agent); err != nil {
		return nil, err
	}
	return agent, nil
}

func (s *service) GetByID(ctx context.Context, id uint) (*model.Agent, error) {
	agent, err := s.agentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}
	return agent, nil
}

func (s *service) List(ctx context.Context, offset, limit int, keyword string) ([]model.Agent, int64, error) {
	return s.agentRepo.List(ctx, offset, limit, keyword)
}

func (s *service) Update(ctx context.Context, id uint, req *model.UpdateAgentRequest) (*model.Agent, error) {
	agent, err := s.agentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("agent not found: %w", err)
	}

	if req.Name != nil {
		agent.Name = *req.Name
	}
	if req.Description != nil {
		agent.Description = *req.Description
	}
	if req.Type != nil {
		agent.Type = *req.Type
	}
	if req.Config != nil {
		agent.Config = *req.Config
	}
	if req.Model != nil {
		agent.Model = *req.Model
	}
	if req.Temperature != nil {
		agent.Temperature = *req.Temperature
	}
	if req.MaxTokens != nil {
		agent.MaxTokens = *req.MaxTokens
	}
	if req.Enabled != nil {
		agent.Enabled = *req.Enabled
	}

	if err := s.agentRepo.Update(ctx, agent); err != nil {
		return nil, err
	}
	return agent, nil
}

func (s *service) Delete(ctx context.Context, id uint) error {
	return s.agentRepo.Delete(ctx, id)
}
