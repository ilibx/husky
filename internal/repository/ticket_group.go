package repository

import (
	"context"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// CreateTicketGroup 创建工单群绑定
func (r *TicketRepository) CreateTicketGroup(ctx context.Context, tg *model.TicketGroup) error {
	return r.db.WithContext(ctx).Create(tg).Error
}

// GetTicketGroupByTicket 根据工单获取群信息
func (r *TicketRepository) GetTicketGroupByTicket(ctx context.Context, ticketID uint) (*model.TicketGroup, error) {
	var tg model.TicketGroup
	err := r.db.WithContext(ctx).Where("ticket_id = ? AND status = 'active'", ticketID).First(&tg).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tg, nil
}

// GetTicketGroupByGroupID 根据群ID获取工单
func (r *TicketRepository) GetTicketGroupByGroupID(ctx context.Context, groupID string) (*model.TicketGroup, error) {
	var tg model.TicketGroup
	err := r.db.WithContext(ctx).Where("group_id = ? AND status = 'active'", groupID).First(&tg).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tg, nil
}
