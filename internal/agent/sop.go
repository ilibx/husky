package agent

import (
	"context"
	"fmt"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
)

type SOPService interface {
	Create(ctx context.Context, userID uint, req *model.CreateSOPRequest) (*model.SOP, error)
	GetByID(ctx context.Context, id uint) (*model.SOP, error)
	List(ctx context.Context, offset, limit int, keyword string) ([]model.SOP, int64, error)
	Update(ctx context.Context, id uint, req *model.UpdateSOPRequest) (*model.SOP, error)
	Delete(ctx context.Context, id uint) error
}

type sopService struct {
	sopRepo *repository.SOPRepository
}

func NewSOPService(sopRepo *repository.SOPRepository) SOPService {
	return &sopService{sopRepo: sopRepo}
}

func (s *sopService) Create(ctx context.Context, userID uint, req *model.CreateSOPRequest) (*model.SOP, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	version := req.Version
	if version == "" {
		version = "1.0"
	}

	riskLevel := req.RiskLevel
	if riskLevel == "" {
		riskLevel = "low"
	}

	sop := &model.SOP{
		Name:              req.Name,
		Description:       req.Description,
		Version:           version,
		Steps:             req.Steps,
		TriggerType:       req.TriggerType,
		TriggerConfig:     req.TriggerConfig,
		RiskLevel:         riskLevel,
		NotificationConfig: req.NotificationConfig,
		CreatedBy:         userID,
	}

	if err := s.sopRepo.Create(ctx, sop); err != nil {
		return nil, err
	}
	return sop, nil
}

func (s *sopService) GetByID(ctx context.Context, id uint) (*model.SOP, error) {
	sop, err := s.sopRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("sop not found: %w", err)
	}
	return sop, nil
}

func (s *sopService) List(ctx context.Context, offset, limit int, keyword string) ([]model.SOP, int64, error) {
	return s.sopRepo.List(ctx, offset, limit, keyword)
}

func (s *sopService) Update(ctx context.Context, id uint, req *model.UpdateSOPRequest) (*model.SOP, error) {
	sop, err := s.sopRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("sop not found: %w", err)
	}

	if req.Name != nil {
		sop.Name = *req.Name
	}
	if req.Description != nil {
		sop.Description = *req.Description
	}
	if req.Version != nil {
		sop.Version = *req.Version
	}
	if req.Steps != nil {
		sop.Steps = *req.Steps
	}
	if req.TriggerType != nil {
		sop.TriggerType = *req.TriggerType
	}
	if req.TriggerConfig != nil {
		sop.TriggerConfig = *req.TriggerConfig
	}
	if req.RiskLevel != nil {
		sop.RiskLevel = *req.RiskLevel
	}
	if req.NotificationConfig != nil {
		sop.NotificationConfig = *req.NotificationConfig
	}
	if req.Status != nil {
		sop.Status = *req.Status
	}

	if err := s.sopRepo.Update(ctx, sop); err != nil {
		return nil, err
	}
	return sop, nil
}

func (s *sopService) Delete(ctx context.Context, id uint) error {
	return s.sopRepo.Delete(ctx, id)
}
