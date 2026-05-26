package repository

import (
	"context"
	"fmt"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// BaseRepository 基础 Repository
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository 创建基础 Repository
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// UserRepository 用户 Repository
type UserRepository struct {
	*BaseRepository
}

// NewUserRepository 创建用户 Repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete 删除用户（软删除）
func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

// List 获取用户列表
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}

// TicketRepository 工单 Repository
type TicketRepository struct {
	*BaseRepository
}

// NewTicketRepository 创建工单 Repository
func NewTicketRepository(db *gorm.DB) *TicketRepository {
	return &TicketRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

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

// Update 更新工单
func (r *TicketRepository) Update(ctx context.Context, ticket *model.Ticket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
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

	// 应用过滤条件
	for key, value := range filters {
		if value != nil && value != "" {
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

// KnowledgeRepository 知识库 Repository
type KnowledgeRepository struct {
	*BaseRepository
}

// NewKnowledgeRepository 创建知识库 Repository
func NewKnowledgeRepository(db *gorm.DB) *KnowledgeRepository {
	return &KnowledgeRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建知识文章
func (r *KnowledgeRepository) Create(ctx context.Context, knowledge *model.Knowledge) error {
	return r.db.WithContext(ctx).Create(knowledge).Error
}

// GetByID 根据 ID 获取知识文章
func (r *KnowledgeRepository) GetByID(ctx context.Context, id uint) (*model.Knowledge, error) {
	var knowledge model.Knowledge
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Author").
		First(&knowledge, id).Error
	if err != nil {
		return nil, err
	}
	return &knowledge, nil
}

// Update 更新知识文章
func (r *KnowledgeRepository) Update(ctx context.Context, knowledge *model.Knowledge) error {
	return r.db.WithContext(ctx).Save(knowledge).Error
}

// Delete 删除知识文章（软删除）
func (r *KnowledgeRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Knowledge{}, id).Error
}

// List 获取知识文章列表
func (r *KnowledgeRepository) List(ctx context.Context, offset, limit int, status *int) ([]model.Knowledge, int64, error) {
	var knowledgeList []model.Knowledge
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Knowledge{})

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Category").
		Preload("Author").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&knowledgeList).Error

	return knowledgeList, total, err
}

// Search 搜索知识文章
func (r *KnowledgeRepository) Search(ctx context.Context, keyword string, offset, limit int) ([]model.Knowledge, int64, error) {
	var knowledgeList []model.Knowledge
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Knowledge{}).
		Where("status = ? AND (title LIKE ? OR content LIKE ? OR tags LIKE ?)",
			1, "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Category").
		Preload("Author").
		Offset(offset).
		Limit(limit).
		Order("view_count DESC").
		Find(&knowledgeList).Error

	return knowledgeList, total, err
}

// AgentRepository Agent Repository
type AgentRepository struct {
	*BaseRepository
}

// NewAgentRepository 创建 Agent Repository
func NewAgentRepository(db *gorm.DB) *AgentRepository {
	return &AgentRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create 创建 Agent
func (r *AgentRepository) Create(ctx context.Context, agent *model.Agent) error {
	return r.db.WithContext(ctx).Create(agent).Error
}

// GetByID 根据 ID 获取 Agent
func (r *AgentRepository) GetByID(ctx context.Context, id uint) (*model.Agent, error) {
	var agent model.Agent
	err := r.db.WithContext(ctx).First(&agent, id).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// Update 更新 Agent
func (r *AgentRepository) Update(ctx context.Context, agent *model.Agent) error {
	return r.db.WithContext(ctx).Save(agent).Error
}

// Delete 删除 Agent（软删除）
func (r *AgentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Agent{}, id).Error
}

// List 获取 Agent 列表
func (r *AgentRepository) List(ctx context.Context, offset, limit int, enabled *bool) ([]model.Agent, int64, error) {
	var agents []model.Agent
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Agent{})

	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&agents).Error

	return agents, total, err
}
