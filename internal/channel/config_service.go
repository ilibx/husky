package channel

import (
	"context"
	"fmt"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
)

type ConfigService interface {
	Create(ctx context.Context, req *model.CreateChannelConfigRequest) (*model.ChannelConfig, error)
	GetByID(ctx context.Context, id uint) (*model.ChannelConfig, error)
	List(ctx context.Context) ([]model.ChannelConfig, error)
	Update(ctx context.Context, id uint, req *model.UpdateChannelConfigRequest) (*model.ChannelConfig, error)
	Delete(ctx context.Context, id uint) error
}

type configService struct {
	cfgRepo *repository.ChannelConfigRepository
}

func NewConfigService(cfgRepo *repository.ChannelConfigRepository) ConfigService {
	return &configService{cfgRepo: cfgRepo}
}

func (s *configService) Create(ctx context.Context, req *model.CreateChannelConfigRequest) (*model.ChannelConfig, error) {
	if req.Name == "" || req.Type == "" {
		return nil, fmt.Errorf("name and type are required")
	}

	validTypes := map[string]bool{"lark": true, "dingtalk": true, "wecom": true, "email": true}
	if !validTypes[req.Type] {
		return nil, fmt.Errorf("invalid channel type: %s", req.Type)
	}

	cfg := &model.ChannelConfig{
		Name:    req.Name,
		Type:    req.Type,
		Config:  req.Config,
		Enabled: true,
		Status:  1,
	}

	if err := s.cfgRepo.Create(ctx, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *configService) GetByID(ctx context.Context, id uint) (*model.ChannelConfig, error) {
	cfg, err := s.cfgRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("channel config not found: %w", err)
	}
	return cfg, nil
}

func (s *configService) List(ctx context.Context) ([]model.ChannelConfig, error) {
	return s.cfgRepo.List(ctx)
}

func (s *configService) Update(ctx context.Context, id uint, req *model.UpdateChannelConfigRequest) (*model.ChannelConfig, error) {
	cfg, err := s.cfgRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("channel config not found: %w", err)
	}

	if req.Name != nil {
		cfg.Name = *req.Name
	}
	if req.Config != nil {
		cfg.Config = *req.Config
	}
	if req.Enabled != nil {
		cfg.Enabled = *req.Enabled
	}
	if req.Status != nil {
		cfg.Status = *req.Status
	}

	if err := s.cfgRepo.Update(ctx, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *configService) Delete(ctx context.Context, id uint) error {
	return s.cfgRepo.Delete(ctx, id)
}
