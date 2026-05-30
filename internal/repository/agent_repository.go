package repository

import (
	"context"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// AgentRepository Agent Repository
type AgentRepository struct {
	*BaseRepository
}

func NewAgentRepository(db *gorm.DB) *AgentRepository {
	return &AgentRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *AgentRepository) Create(ctx context.Context, agent *model.Agent) error {
	return r.db.WithContext(ctx).Create(agent).Error
}

func (r *AgentRepository) GetByID(ctx context.Context, id uint) (*model.Agent, error) {
	var agent model.Agent
	err := r.db.WithContext(ctx).Preload("Creator").First(&agent, id).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *AgentRepository) List(ctx context.Context, offset, limit int) ([]model.Agent, int64, error) {
	var list []model.Agent
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Agent{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// ListByTrigger 按触发类型查询已启用的 Agent
func (r *AgentRepository) ListByTrigger(ctx context.Context, triggerOn string) ([]model.Agent, error) {
	var list []model.Agent
	err := r.db.WithContext(ctx).Model(&model.Agent{}).
		Where("enabled = ?", true).
		Where("config->>'trigger_on' IN (?, 'any')", triggerOn).
		Order("id ASC").
		Find(&list).Error
	return list, err
}

func (r *AgentRepository) Update(ctx context.Context, agent *model.Agent) error {
	return r.db.WithContext(ctx).Save(agent).Error
}

func (r *AgentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Agent{}, id).Error
}
