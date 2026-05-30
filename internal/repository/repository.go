package repository

import (
	"gorm.io/gorm"
)

// DatabaseConnection 数据库连接包装，解耦 Redis 依赖
type DatabaseConnection struct {
	DB *gorm.DB
}

var safeColumns = map[string]bool{
	"id":              true,
	"ticket_no":       true,
	"title":           true,
	"status":          true,
	"priority":        true,
	"category_id":     true,
	"assignee_id":     true,
	"requester_id":    true,
	"created_by":      true,
	"department_id":   true,
	"source":          true,
	"created_at":      true,
	"updated_at":      true,
	"due_at":          true,
	"resolved_at":     true,
	"closed_at":       true,
	"role":            true,
	"email":           true,
	"username":        true,
	"is_active":       true,
}

func isSafeColumn(name string) bool {
	return safeColumns[name]
}

// BaseRepository 基础 Repository
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository 创建基础 Repository
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// TicketRepository 工单 Repository
type TicketRepository struct {
	*BaseRepository
}

// GetDB 获取底层数据库连接
func (r *TicketRepository) GetDB() *gorm.DB {
	return r.db
}

// NewTicketRepository 创建工单 Repository
func NewTicketRepository(db *gorm.DB) *TicketRepository {
	return &TicketRepository{
		BaseRepository: NewBaseRepository(db),
	}
}
