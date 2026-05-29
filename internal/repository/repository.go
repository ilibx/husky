package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// DatabaseConnection 数据库连接包装，解耦 Redis 依赖
type DatabaseConnection struct {
	DB *gorm.DB
}

// BaseRepository 基础 Repository
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository 创建基础 Repository
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// CategoryRepository 分类 Repository
type CategoryRepository struct {
	*BaseRepository
}

// NewCategoryRepository 创建分类 Repository
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

func (r *CategoryRepository) Create(ctx context.Context, cat *model.Category) error {
	return r.db.WithContext(ctx).Create(cat).Error
}

func (r *CategoryRepository) GetByID(ctx context.Context, id uint) (*model.Category, error) {
	var cat model.Category
	err := r.db.WithContext(ctx).Preload("Parent").First(&cat, id).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

func (r *CategoryRepository) List(ctx context.Context) ([]model.Category, error) {
	var list []model.Category
	err := r.db.WithContext(ctx).Order("sort_order ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *CategoryRepository) Update(ctx context.Context, cat *model.Category) error {
	return r.db.WithContext(ctx).Save(cat).Error
}

func (r *CategoryRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Category{}, id).Error
}

// SOPRepository SOP Repository
type SOPRepository struct {
	*BaseRepository
}

func NewSOPRepository(db *gorm.DB) *SOPRepository {
	return &SOPRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *SOPRepository) Create(ctx context.Context, sop *model.SOP) error {
	return r.db.WithContext(ctx).Create(sop).Error
}

func (r *SOPRepository) GetByID(ctx context.Context, id uint) (*model.SOP, error) {
	var sop model.SOP
	err := r.db.WithContext(ctx).Preload("Creator").First(&sop, id).Error
	if err != nil {
		return nil, err
	}
	return &sop, nil
}

func (r *SOPRepository) List(ctx context.Context, offset, limit int) ([]model.SOP, int64, error) {
	var list []model.SOP
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.SOP{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *SOPRepository) Update(ctx context.Context, sop *model.SOP) error {
	return r.db.WithContext(ctx).Save(sop).Error
}

func (r *SOPRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.SOP{}, id).Error
}

// DepartmentRepository 部门 Repository
type DepartmentRepository struct {
	*BaseRepository
}

func NewDepartmentRepository(db *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *DepartmentRepository) Create(ctx context.Context, dept *model.Department) error {
	return r.db.WithContext(ctx).Create(dept).Error
}

func (r *DepartmentRepository) GetByID(ctx context.Context, id uint) (*model.Department, error) {
	var dept model.Department
	err := r.db.WithContext(ctx).Preload("Parent").Preload("Manager").First(&dept, id).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

func (r *DepartmentRepository) List(ctx context.Context) ([]model.Department, error) {
	var list []model.Department
	err := r.db.WithContext(ctx).Preload("Manager").Order("code ASC").Find(&list).Error
	return list, err
}

func (r *DepartmentRepository) Update(ctx context.Context, dept *model.Department) error {
	return r.db.WithContext(ctx).Save(dept).Error
}

func (r *DepartmentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Department{}, id).Error
}

// AgentRepository Agent Repository
type AgentRepository struct {
	*BaseRepository
}

func NewAgentRepository(db *gorm.DB) *AgentRepository {
	return &AgentRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *AgentRepository) Create(ctx context.Context, agent *model.Agent) error {
	return r.db.WithContext(ctx).Create(agent).Error
}

func (r *AgentRepository) GetByID(ctx context.Context, id uint) (*model.Agent, error) {
	var agent model.Agent
	err := r.db.WithContext(ctx).Preload("Creator").First(&agent, id).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *AgentRepository) List(ctx context.Context, offset, limit int) ([]model.Agent, int64, error) {
	var list []model.Agent
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Agent{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *AgentRepository) Update(ctx context.Context, agent *model.Agent) error {
	return r.db.WithContext(ctx).Save(agent).Error
}

func (r *AgentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Agent{}, id).Error
}

// ChannelConfigRepository 渠道配置 Repository
type ChannelConfigRepository struct {
	*BaseRepository
}

func NewChannelConfigRepository(db *gorm.DB) *ChannelConfigRepository {
	return &ChannelConfigRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *ChannelConfigRepository) Create(ctx context.Context, cfg *model.ChannelConfig) error {
	return r.db.WithContext(ctx).Create(cfg).Error
}

func (r *ChannelConfigRepository) GetByID(ctx context.Context, id uint) (*model.ChannelConfig, error) {
	var cfg model.ChannelConfig
	err := r.db.WithContext(ctx).First(&cfg, id).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *ChannelConfigRepository) List(ctx context.Context) ([]model.ChannelConfig, error) {
	var list []model.ChannelConfig
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *ChannelConfigRepository) Update(ctx context.Context, cfg *model.ChannelConfig) error {
	return r.db.WithContext(ctx).Save(cfg).Error
}

func (r *ChannelConfigRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.ChannelConfig{}, id).Error
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

// Update 更新工单（全量 Save）
func (r *TicketRepository) Update(ctx context.Context, ticket *model.Ticket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

// UpdateFields 更新工单的指定字段，避免 Save 覆盖零值
func (r *TicketRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.Ticket{}).Where("id = ?", id).Updates(fields).Error
}

// Delete 删除工单（软删除）
func (r *TicketRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Ticket{}, id).Error
}

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

// List 获取工单列表
func (r *TicketRepository) List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]model.Ticket, int64, error) {
	var tickets []model.Ticket
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Ticket{})

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
		default:
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

// GetStatsOverview 获取工单统计概览
func (r *TicketRepository) GetStatsOverview(ctx context.Context) (map[string]interface{}, error) {
	var total int64
	r.db.WithContext(ctx).Model(&model.Ticket{}).Count(&total)

	var openCount int64
	r.db.WithContext(ctx).Model(&model.Ticket{}).Where("status = ?", "open").Count(&openCount)

	var inProgressCount int64
	r.db.WithContext(ctx).Model(&model.Ticket{}).Where("status = ?", "in_progress").Count(&inProgressCount)

	var resolvedCount int64
	r.db.WithContext(ctx).Model(&model.Ticket{}).Where("status = ?", "resolved").Count(&resolvedCount)

	var closedCount int64
	r.db.WithContext(ctx).Model(&model.Ticket{}).Where("status = ?", "closed").Count(&closedCount)

	var overdueCount int64
	r.db.WithContext(ctx).Model(&model.Ticket{}).Where("due_at IS NOT NULL AND due_at < NOW() AND status NOT IN ('resolved', 'closed')").Count(&overdueCount)

	return map[string]interface{}{
		"total":    total,
		"open":     openCount,
		"in_progress": inProgressCount,
		"resolved": resolvedCount,
		"closed":   closedCount,
		"overdue":  overdueCount,
	}, nil
}

// GetTicketStats 获取工单按渠道/优先级分布统计
func (r *TicketRepository) GetTicketStats(ctx context.Context) (map[string]interface{}, error) {
	type SourceCount struct {
		Source string
		Count  int64
	}
	var bySource []SourceCount
	r.db.WithContext(ctx).Model(&model.Ticket{}).Select("source, count(*) as count").Group("source").Find(&bySource)

	type PriorityCount struct {
		Priority string
		Count    int64
	}
	var byPriority []PriorityCount
	r.db.WithContext(ctx).Model(&model.Ticket{}).Select("priority, count(*) as count").Group("priority").Find(&byPriority)

	sourceMap := make(map[string]int64)
	for _, s := range bySource {
		sourceMap[s.Source] = s.Count
	}
	priorityMap := make(map[string]int64)
	for _, p := range byPriority {
		priorityMap[p.Priority] = p.Count
	}

	return map[string]interface{}{
		"by_source":   sourceMap,
		"by_priority": priorityMap,
	}, nil
}

// CreateSatisfaction 创建满意度评价
func (r *TicketRepository) CreateSatisfaction(ctx context.Context, s *model.Satisfaction) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// GetSatisfactionByTicket 获取工单满意度评价
func (r *TicketRepository) GetSatisfactionByTicket(ctx context.Context, ticketID uint) (*model.Satisfaction, error) {
	var s model.Satisfaction
	err := r.db.WithContext(ctx).Preload("User").Where("ticket_id = ?", ticketID).First(&s).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// GetSatisfactionStats 获取满意度统计
func (r *TicketRepository) GetSatisfactionStats(ctx context.Context) (map[string]interface{}, error) {
	var total int64
	r.db.WithContext(ctx).Model(&model.Satisfaction{}).Count(&total)

	var avg float64
	r.db.WithContext(ctx).Model(&model.Satisfaction{}).Select("COALESCE(AVG(score), 0)").Scan(&avg)

	var distribution []struct {
		Score int
		Count int64
	}
	r.db.WithContext(ctx).Model(&model.Satisfaction{}).Select("score, count(*) as count").Group("score").Order("score").Find(&distribution)

	distMap := make(map[int]int64)
	for _, d := range distribution {
		distMap[d.Score] = d.Count
	}

	return map[string]interface{}{
		"total":        total,
		"average":      avg,
		"distribution": distMap,
	}, nil
}

// SetTicketDueAt 设置工单 SLA 截止时间
func (r *TicketRepository) SetTicketDueAt(ctx context.Context, id uint, dueAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Ticket{}).Where("id = ?", id).Update("due_at", dueAt).Error
}

// AddWatcher 添加工单关注者
func (r *TicketRepository) AddWatcher(ctx context.Context, ticketID, userID uint) error {
	return r.db.WithContext(ctx).FirstOrCreate(&model.TicketWatcher{}, map[string]interface{}{
		"ticket_id": ticketID,
		"user_id":   userID,
	}).Error
}

// RemoveWatcher 移除工单关注者
func (r *TicketRepository) RemoveWatcher(ctx context.Context, ticketID, userID uint) error {
	return r.db.WithContext(ctx).Where("ticket_id = ? AND user_id = ?", ticketID, userID).
		Delete(&model.TicketWatcher{}).Error
}

// GetWatchers 获取工单关注者列表
func (r *TicketRepository) GetWatchers(ctx context.Context, ticketID uint) ([]model.TicketWatcher, error) {
	var watchers []model.TicketWatcher
	err := r.db.WithContext(ctx).Preload("User").
		Where("ticket_id = ?", ticketID).Find(&watchers).Error
	return watchers, err
}

// FindLeastBusyAgent 查找当前工单最少的客服
func (r *TicketRepository) FindLeastBusyAgent(ctx context.Context) (*uint, error) {
	type agentLoad struct {
		AssigneeID uint
		Count      int64
	}
	var result agentLoad
	err := r.db.WithContext(ctx).Model(&model.Ticket{}).
		Select("assignee_id, COUNT(*) as count").
		Where("assignee_id IS NOT NULL AND status NOT IN ('resolved', 'closed')").
		Group("assignee_id").
		Order("count ASC").
		Limit(1).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	if result.AssigneeID == 0 {
		return nil, nil
	}
	return &result.AssigneeID, nil
}

// CreateAuditLog 创建审计日志
func (r *TicketRepository) CreateAuditLog(ctx context.Context, log *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// ListAuditLogs 查询审计日志
func (r *TicketRepository) ListAuditLogs(ctx context.Context, resource string, resourceID uint, offset, limit int) ([]model.AuditLog, int64, error) {
	var list []model.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&model.AuditLog{})
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if resourceID > 0 {
		query = query.Where("resource_id = ?", resourceID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Preload("User").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// CreateWebhookMessage 保存渠道消息记录
func (r *TicketRepository) CreateWebhookMessage(ctx context.Context, msg *model.WebhookMessageRecord) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

// ListWebhookMessages 获取渠道消息记录
func (r *TicketRepository) ListWebhookMessages(ctx context.Context, offset, limit int, channel string) ([]model.WebhookMessageRecord, int64, error) {
	var list []model.WebhookMessageRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&model.WebhookMessageRecord{})
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// CreateNotification 创建通知
func (r *TicketRepository) CreateNotification(ctx context.Context, n *model.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

// ListNotifications 获取用户通知列表
func (r *TicketRepository) ListNotifications(ctx context.Context, userID uint, offset, limit int) ([]model.Notification, int64, error) {
	var list []model.Notification
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Notification{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// GetUnreadNotificationCount 获取未读通知数
func (r *TicketRepository) GetUnreadNotificationCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// MarkNotificationRead 标记通知为已读
func (r *TicketRepository) MarkNotificationRead(ctx context.Context, id, userID uint) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

// MarkAllNotificationsRead 标记所有通知为已读
func (r *TicketRepository) MarkAllNotificationsRead(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

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


