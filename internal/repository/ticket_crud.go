package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// Create 创建工单
func (r *TicketRepository) Create(ctx context.Context, ticket *model.Ticket) error {
	return r.db.WithContext(ctx).Create(ticket).Error
}

// GetByID 根据 ID 获取工单
func (r *TicketRepository) GetByID(ctx context.Context, id uint) (*model.Ticket, error) {
	var ticket model.Ticket
	err := r.db.WithContext(ctx).
		Preload("Requester").
		Preload("Assignee").
		Preload("Category").
		Preload("Comments", func(db *gorm.DB) *gorm.DB {
			return db.Order("comments.created_at ASC")
		}).
		Preload("Attachments").
		First(&ticket, id).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

// GetByTicketNo 根据工单号获取工单
func (r *TicketRepository) GetByTicketNo(ctx context.Context, ticketNo string) (*model.Ticket, error) {
	var ticket model.Ticket
	err := r.db.WithContext(ctx).Where("ticket_no = ?", ticketNo).First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

// FindOverdueTickets 查询已超期的工单（DueAt < now 且未完成）
func (r *TicketRepository) FindOverdueTickets(ctx context.Context, now time.Time) ([]model.Ticket, error) {
	var tickets []model.Ticket
	err := r.db.WithContext(ctx).
		Where("due_at IS NOT NULL AND due_at < ?", now).
		Where("status NOT IN (?)", []string{"resolved", "closed"}).
		Order("due_at ASC").
		Find(&tickets).Error
	return tickets, err
}

// Update 更新工单（带乐观锁）
func (r *TicketRepository) Update(ctx context.Context, ticket *model.Ticket) error {
	result := r.db.WithContext(ctx).
		Model(&model.Ticket{}).
		Where("id = ? AND version = ?", ticket.ID, ticket.Version).
		Updates(map[string]interface{}{
			"title":       ticket.Title,
			"description": ticket.Description,
			"status":      ticket.Status,
			"priority":    ticket.Priority,
			"assignee_id": ticket.AssigneeID,
			"category_id": ticket.CategoryID,
			"due_at":      ticket.DueAt,
			"resolved_at": ticket.ResolvedAt,
			"closed_at":   ticket.ClosedAt,
			"sla_status":  ticket.SLAStatus,
			"source":      ticket.Source,
			"version":     ticket.Version + 1,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("ticket %d was modified by another process, retry", ticket.ID)
	}
	ticket.Version++
	return nil
}

// UpdateFields 更新工单的指定字段，避免 Save 覆盖零值
func (r *TicketRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.Ticket{}).Where("id = ?", id).Updates(fields).Error
}

// Delete 删除工单（软删除）
func (r *TicketRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Ticket{}, id).Error
}

// List 获取工单列表
func (r *TicketRepository) List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]model.Ticket, int64, error) {
	var tickets []model.Ticket
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Ticket{})

	var userIDs []uint
	for key, value := range filters {
		if value == nil {
			continue
		}
		switch key {
		case "keyword":
			q := "%" + value.(string) + "%"
			query = query.Where("title ILIKE ? OR description ILIKE ?", q, q)
		case "assignee_id":
			query = query.Where("assignee_id = ?", value)
		case "scope":
			scope := value.(string)
			uid := filters["scope_user_id"].(uint)
			roleName := filters["scope_role"].(string)
			if scope == "my" {
				query = query.Where("assignee_id = ?", uid)
			} else if scope == "my_team" && roleName != "" {
				var userLevel int
				r.db.WithContext(ctx).Model(&model.Role{}).Where("name = ?", roleName).Pluck("level", &userLevel)
				r.db.WithContext(ctx).Model(&model.User{}).
					Joins("JOIN roles ON roles.name = users.role").
					Where("roles.level <= ?", userLevel).
					Pluck("users.id", &userIDs)
				if len(userIDs) > 0 {
					query = query.Where("assignee_id IN ?", userIDs)
				} else {
					query = query.Where("assignee_id = ?", uid)
				}
			}
		case "scope_user_id", "scope_role":
			// skip — already handled in scope case
		default:
			if !isSafeColumn(key) {
				continue
			}
			query = query.Where(fmt.Sprintf("%s = ?", key), value)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Requester").
		Preload("Assignee").
		Preload("Category").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&tickets).Error

	return tickets, total, err
}
