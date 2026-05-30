package repository

import (
	"context"
	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// CreateTicketRelation 创建工单关联
func (r *TicketRepository) CreateTicketRelation(ctx context.Context, rel *model.TicketRelation) error {
	return r.db.WithContext(ctx).Create(rel).Error
}

// DeleteTicketRelation 删除工单关联
func (r *TicketRepository) DeleteTicketRelation(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.TicketRelation{}, id).Error
}

// ListTicketRelations 获取工单关联列表
func (r *TicketRepository) ListTicketRelations(ctx context.Context, ticketID uint) ([]model.TicketRelation, error) {
	var list []model.TicketRelation
	err := r.db.WithContext(ctx).Where("ticket_id = ? OR related_id = ?", ticketID, ticketID).
		Find(&list).Error
	return list, err
}

// GetTicketRelation 获取单个关联
func (r *TicketRepository) GetTicketRelation(ctx context.Context, id uint) (*model.TicketRelation, error) {
	var rel model.TicketRelation
	err := r.db.WithContext(ctx).First(&rel, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &rel, nil
}
