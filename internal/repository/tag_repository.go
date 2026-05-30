package repository

import (
	"context"
	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreateTag 创建标签
func (r *TicketRepository) CreateTag(ctx context.Context, tag *model.Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

// ListTags 获取所有标签
func (r *TicketRepository) ListTags(ctx context.Context) ([]model.Tag, error) {
	var list []model.Tag
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetTag 获取单个标签
func (r *TicketRepository) GetTag(ctx context.Context, id uint) (*model.Tag, error) {
	var tag model.Tag
	if err := r.db.WithContext(ctx).First(&tag, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &tag, nil
}

// UpdateTag 更新标签
func (r *TicketRepository) UpdateTag(ctx context.Context, tag *model.Tag) error {
	return r.db.WithContext(ctx).Save(tag).Error
}

// DeleteTag 删除标签（自动清理关联）
func (r *TicketRepository) DeleteTag(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tag_id = ?", id).Delete(&model.TicketTag{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Tag{}, id).Error
	})
}

// AddTagsToTicket 为工单添加标签
func (r *TicketRepository) AddTagsToTicket(ctx context.Context, ticketID uint, tagIDs []uint) error {
	if len(tagIDs) == 0 {
		return nil
	}
	var records []model.TicketTag
	for _, tagID := range tagIDs {
		records = append(records, model.TicketTag{TicketID: ticketID, TagID: tagID})
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&records).Error
}

// RemoveTagFromTicket 移除工单标签
func (r *TicketRepository) RemoveTagFromTicket(ctx context.Context, ticketID, tagID uint) error {
	return r.db.WithContext(ctx).Where("ticket_id = ? AND tag_id = ?", ticketID, tagID).
		Delete(&model.TicketTag{}).Error
}

// GetTicketTags 获取工单标签列表
func (r *TicketRepository) GetTicketTags(ctx context.Context, ticketID uint) ([]model.Tag, error) {
	var tagIDs []uint
	if err := r.db.WithContext(ctx).Model(&model.TicketTag{}).
		Where("ticket_id = ?", ticketID).Pluck("tag_id", &tagIDs).Error; err != nil {
		return nil, err
	}
	if len(tagIDs) == 0 {
		return nil, nil
	}
	var tags []model.Tag
	if err := r.db.WithContext(ctx).Find(&tags, tagIDs).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

// UpdateTicketTags 替换工单的所有标签
func (r *TicketRepository) UpdateTicketTags(ctx context.Context, ticketID uint, tagIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("ticket_id = ?", ticketID).Delete(&model.TicketTag{}).Error; err != nil {
			return err
		}
		for _, tagID := range tagIDs {
			if err := tx.Create(&model.TicketTag{TicketID: ticketID, TagID: tagID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
