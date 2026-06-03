package repository

import (
	"context"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

type MCPRepository struct {
	*BaseRepository
}

func NewMCPRepository(db *gorm.DB) *MCPRepository {
	return &MCPRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *MCPRepository) Create(ctx context.Context, mcp *model.MCP) error {
	return r.db.WithContext(ctx).Create(mcp).Error
}

func (r *MCPRepository) GetByID(ctx context.Context, id uint) (*model.MCP, error) {
	var m model.MCP
	err := r.db.WithContext(ctx).Preload("Creator").First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MCPRepository) List(ctx context.Context, offset, limit int) ([]model.MCP, int64, error) {
	var list []model.MCP
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.MCP{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *MCPRepository) ListBySkill(ctx context.Context, skillID uint) ([]model.MCP, error) {
	var list []model.MCP
	err := r.db.WithContext(ctx).
		Where("skill_id = ?", skillID).
		Order("created_at ASC").
		Find(&list).Error
	return list, err
}

func (r *MCPRepository) Update(ctx context.Context, mcp *model.MCP) error {
	return r.db.WithContext(ctx).Save(mcp).Error
}

func (r *MCPRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.MCP{}, id).Error
}
