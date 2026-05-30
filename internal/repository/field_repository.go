package repository

import (
	"context"
	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// ListTicketFields 获取所有自定义字段
func (r *TicketRepository) ListTicketFields(ctx context.Context) ([]model.TicketField, error) {
	var list []model.TicketField
	err := r.db.WithContext(ctx).Order("sort_order ASC, id ASC").Find(&list).Error
	return list, err
}

// GetTicketField 获取单个自定义字段
func (r *TicketRepository) GetTicketField(ctx context.Context, id uint) (*model.TicketField, error) {
	var f model.TicketField
	err := r.db.WithContext(ctx).First(&f, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}

// CreateTicketField 创建自定义字段
func (r *TicketRepository) CreateTicketField(ctx context.Context, f *model.TicketField) error {
	return r.db.WithContext(ctx).Create(f).Error
}

// UpdateTicketField 更新自定义字段
func (r *TicketRepository) UpdateTicketField(ctx context.Context, f *model.TicketField) error {
	return r.db.WithContext(ctx).Save(f).Error
}

// DeleteTicketField 删除自定义字段
func (r *TicketRepository) DeleteTicketField(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("field_id = ?", id).Delete(&model.TicketFieldValue{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.TicketField{}, id).Error
	})
}

// SetTicketFieldValues 设置工单自定义字段值
func (r *TicketRepository) SetTicketFieldValues(ctx context.Context, ticketID uint, values []model.TicketFieldValue) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("ticket_id = ?", ticketID).Delete(&model.TicketFieldValue{}).Error; err != nil {
			return err
		}
		for _, v := range values {
			v.TicketID = ticketID
			if err := tx.Create(&v).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetTicketFieldValues 获取工单自定义字段值
func (r *TicketRepository) GetTicketFieldValues(ctx context.Context, ticketID uint) ([]model.TicketFieldValue, error) {
	var list []model.TicketFieldValue
	err := r.db.WithContext(ctx).Preload("Field").
		Where("ticket_id = ?", ticketID).
		Order("field_id ASC").
		Find(&list).Error
	return list, err
}
