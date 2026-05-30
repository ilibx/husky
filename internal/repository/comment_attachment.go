package repository

import (
	"context"

	"github.com/husky/husky/internal/model"
)

// AddComment 添加工单评论
func (r *TicketRepository) AddComment(ctx context.Context, comment *model.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

// CreateAttachment 创建附件记录
func (r *TicketRepository) CreateAttachment(ctx context.Context, att *model.Attachment) error {
	return r.db.WithContext(ctx).Create(att).Error
}

// GetAttachment 获取附件
func (r *TicketRepository) GetAttachment(ctx context.Context, id uint) (*model.Attachment, error) {
	var att model.Attachment
	err := r.db.WithContext(ctx).Preload("Uploader").First(&att, id).Error
	if err != nil {
		return nil, err
	}
	return &att, nil
}

// ListAttachments 获取工单附件列表
func (r *TicketRepository) ListAttachments(ctx context.Context, ticketID uint) ([]model.Attachment, error) {
	var attachments []model.Attachment
	err := r.db.WithContext(ctx).
		Preload("Uploader").
		Where("ticket_id = ?", ticketID).
		Order("created_at DESC").
		Find(&attachments).Error
	return attachments, err
}

// DeleteAttachment 删除附件
func (r *TicketRepository) DeleteAttachment(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Attachment{}, id).Error
}

// ListComments 获取工单评论列表
func (r *TicketRepository) ListComments(ctx context.Context, ticketID uint, offset, limit int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Comment{}).Where("ticket_id = ?", ticketID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at ASC").Offset(offset).Limit(limit).Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}
