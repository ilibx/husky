package ticket

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// TicketRepository 工单数据访问接口
type TicketRepository interface {
	// Create 创建工单
	Create(ctx context.Context, ticket *model.Ticket) error
	// GetByID 根据 ID 获取工单
	GetByID(ctx context.Context, id string) (*model.Ticket, error)
	// List 查询工单列表
	List(ctx context.Context, opts *ListTicketsOptions) ([]*model.Ticket, int64, error)
	// Update 更新工单
	Update(ctx context.Context, ticket *model.Ticket) error
	// Delete 删除工单（软删除）
	Delete(ctx context.Context, id string) error
	// AddComment 添加评论
	AddComment(ctx context.Context, comment *model.Comment) error
	// ListComments 获取评论列表
	ListComments(ctx context.Context, ticketID string, offset, limit int) ([]*model.Comment, int64, error)
}

// ListTicketsOptions 查询工单选项
type ListTicketsOptions struct {
	Status      []model.TicketStatus
	Priority    []model.Priority
	AssigneeID  string
	SubmitterID string
	CategoryID  string
	Keyword     string
	Offset      int
	Limit       int
	SortBy      string
	SortOrder   string
}

type ticketRepository struct {
	db    *gorm.DB
	cache *redis.Client
}

// NewTicketRepository 创建工单仓库实例
func NewTicketRepository(db *gorm.DB, cache *redis.Client) TicketRepository {
	return &ticketRepository{
		db:    db,
		cache: cache,
	}
}

func (r *ticketRepository) Create(ctx context.Context, ticket *model.Ticket) error {
	return r.db.WithContext(ctx).Create(ticket).Error
}

func (r *ticketRepository) GetByID(ctx context.Context, id string) (*model.Ticket, error) {
	// 尝试从缓存获取
	cacheKey := "ticket:" + id
	cached := &model.Ticket{}
	if err := r.cache.Get(ctx, cacheKey).Scan(cached); err == nil {
		return cached, nil
	}

	// 从数据库获取
	var ticket model.Ticket
	if err := r.db.WithContext(ctx).
		Preload("Comments").
		Preload("Attachments").
		Preload("OperationLogs").
		First(&ticket, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	// 写入缓存
	r.cache.Set(ctx, cacheKey, ticket, 5*time.Minute)

	return &ticket, nil
}

func (r *ticketRepository) List(ctx context.Context, opts *ListTicketsOptions) ([]*model.Ticket, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Ticket{})

	// 应用筛选条件
	if len(opts.Status) > 0 {
		query = query.Where("status IN ?", opts.Status)
	}
	if len(opts.Priority) > 0 {
		query = query.Where("priority IN ?", opts.Priority)
	}
	if opts.AssigneeID != "" {
		query = query.Where("assignee_id = ?", opts.AssigneeID)
	}
	if opts.SubmitterID != "" {
		query = query.Where("submitter_id = ?", opts.SubmitterID)
	}
	if opts.CategoryID != "" {
		query = query.Where("category_id = ?", opts.CategoryID)
	}
	if opts.Keyword != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", 
			"%"+opts.Keyword+"%", "%"+opts.Keyword+"%")
	}

	// 计数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	sortBy := "created_at"
	if opts.SortBy != "" {
		sortBy = opts.SortBy
	}
	sortOrder := "DESC"
	if opts.SortOrder == "asc" {
		sortOrder = "ASC"
	}
	query = query.Order(sortBy + " " + sortOrder)

	// 分页
	var tickets []*model.Ticket
	if err := query.Offset(opts.Offset).Limit(opts.Limit).Find(&tickets).Error; err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

func (r *ticketRepository) Update(ctx context.Context, ticket *model.Ticket) error {
	// 删除缓存
	cacheKey := fmt.Sprintf("ticket:%d", ticket.ID)
	r.cache.Del(ctx, cacheKey)

	return r.db.WithContext(ctx).Save(ticket).Error
}

func (r *ticketRepository) Delete(ctx context.Context, id string) error {
	// 软删除
	result := r.db.WithContext(ctx).
		Model(&model.Ticket{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.Expr("NOW()"))

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	// 删除缓存
	cacheKey := "ticket:" + id
	r.cache.Del(ctx, cacheKey)

	return result.Error
}

func (r *ticketRepository) AddComment(ctx context.Context, comment *model.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *ticketRepository) ListComments(ctx context.Context, ticketID string, offset, limit int) ([]*model.Comment, int64, error) {
	var comments []*model.Comment
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
