package model

import "time"

// CreateTicketRequest 创建工单请求
type CreateTicketRequest struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Priority    string                 `json:"priority"`
	RequesterID string                 `json:"requester_id"`
	Channel     string                 `json:"channel"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TicketResponse 工单响应
type TicketResponse struct {
	ID          uint       `json:"id"`
	TicketNo    string     `json:"ticket_no"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	Source      string     `json:"source"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	RequesterID uint       `json:"requester_id"`
	AssigneeID  *uint      `json:"assignee_id,omitempty"`
}

// TicketStatus 工单状态
type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "open"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusPending    TicketStatus = "pending"
	TicketStatusResolved   TicketStatus = "resolved"
	TicketStatusClosed     TicketStatus = "closed"
)

// ValidTransitions 状态合法转换表
var ValidTransitions = map[TicketStatus][]TicketStatus{
	TicketStatusOpen:       {TicketStatusInProgress, TicketStatusClosed},
	TicketStatusInProgress: {TicketStatusPending, TicketStatusResolved, TicketStatusOpen},
	TicketStatusPending:    {TicketStatusInProgress, TicketStatusResolved},
	TicketStatusResolved:   {TicketStatusClosed, TicketStatusOpen},
	TicketStatusClosed:     {},
}

// IsValidTransition 检查状态转换是否合法
func IsValidTransition(from, to TicketStatus) bool {
	allowed, ok := ValidTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// ValidPriorities 有效优先级列表
var ValidPriorities = []Priority{PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent}

// IsValidPriority 检查优先级是否有效
func IsValidPriority(p string) bool {
	for _, v := range ValidPriorities {
		if string(v) == p {
			return true
		}
	}
	return false
}

// Priority 优先级
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityUrgent   Priority = "urgent"
)

// Ticket 工单模型
type Ticket struct {
	Base
	TicketNo    string `gorm:"uniqueIndex;size:50;not null" json:"ticket_no"`
	Title       string `gorm:"size:255;not null" json:"title"`
	Description string `gorm:"type:text;not null" json:"description"`
	Status      string `gorm:"size:50;default:'open'" json:"status"`          // open, in_progress, pending, resolved, closed
	Priority    string `gorm:"size:20;default:'medium'" json:"priority"`      // low, medium, high, urgent
	Type        string `gorm:"size:50" json:"type"`                           // incident, service_request, problem, change
	CategoryID  uint   `json:"category_id,omitempty"`
	AssigneeID  *uint  `json:"assignee_id,omitempty"`
	RequesterID uint   `json:"requester_id"`
	Source      string `gorm:"size:50;default:'web'" json:"source"`           // web, email, im, api
	SLAStatus   string `gorm:"size:50;default:'normal'" json:"sla_status"`    // normal, warning, breached
	DueAt       *time.Time `json:"due_at,omitempty"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	ClosedAt    *time.Time `json:"closed_at,omitempty"`

	// 关联
	Requester     User         `gorm:"foreignKey:RequesterID" json:"requester,omitempty"`
	Assignee      *User        `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Category      *Category    `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Comments      []Comment    `gorm:"foreignKey:TicketID" json:"comments,omitempty"`
	Attachments   []Attachment `gorm:"foreignKey:TicketID" json:"attachments,omitempty"`
	OperationLogs []AuditLog   `gorm:"foreignKey:ResourceID" json:"operation_logs,omitempty"`
	Tags          []Tag        `gorm:"many2many:ticket_tags;" json:"tags,omitempty"`
}

// Tag 标签
type Tag struct {
	Base
	Name  string `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Color string `gorm:"size:20" json:"color"` // 标签颜色，用于前端展示
}

func (Tag) TableName() string { return "tags" }

// TicketTag 工单-标签关联
type TicketTag struct {
	TicketID uint `gorm:"primaryKey;autoIncrement:false"`
	TagID    uint `gorm:"primaryKey;autoIncrement:false"`
}

// TicketWatcher 工单关注者
type TicketWatcher struct {
	Base
	TicketID uint `gorm:"uniqueIndex:idx_ticket_user;not null" json:"ticket_id"`
	UserID   uint `gorm:"uniqueIndex:idx_ticket_user;not null" json:"user_id"`

	Ticket Ticket `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TicketGroup 工单-飞书群绑定
type TicketGroup struct {
	Base
	TicketID  uint   `gorm:"uniqueIndex;not null" json:"ticket_id"`
	GroupID   string `gorm:"size:255;not null" json:"group_id"` // feishu chat_id
	GroupName string `gorm:"size:255" json:"group_name"`
	JoinLink  string `gorm:"type:text" json:"join_link"`
	Status    string `gorm:"size:20;default:'active'" json:"status"` // active, archived, deleted
}

func (TicketGroup) TableName() string { return "ticket_groups" }

// TicketRelation 工单关联关系
type TicketRelation struct {
	Base
	TicketID     uint   `gorm:"uniqueIndex:idx_ticket_rel;not null" json:"ticket_id"`
	RelatedID    uint   `gorm:"uniqueIndex:idx_ticket_rel;not null" json:"related_id"`
	RelationType string `gorm:"size:50;not null" json:"relation_type"` // parent, child, related, duplicate, blocks, blocked_by
}

func (TicketRelation) TableName() string { return "ticket_relations" }

// TicketField 工单自定义字段定义
type TicketField struct {
	Base
	Name        string `gorm:"size:100;not null" json:"name"`
	FieldKey    string `gorm:"size:100;uniqueIndex;not null" json:"field_key"`
	FieldType   string `gorm:"size:50;not null;default:'text'" json:"field_type"` // text, textarea, select, multi_select, number, date, boolean
	Options     string `gorm:"type:text" json:"options,omitempty"`                // JSON array for select/multi_select
	Required    bool   `gorm:"default:false" json:"required"`
	SortOrder   int    `gorm:"default:0" json:"sort_order"`
	Placeholder string `gorm:"size:255" json:"placeholder,omitempty"`
	Enabled     bool   `gorm:"default:true" json:"enabled"`
}

func (TicketField) TableName() string { return "ticket_fields" }

// TicketFieldValue 工单自定义字段值
type TicketFieldValue struct {
	TicketID uint   `gorm:"primaryKey;autoIncrement:false" json:"ticket_id"`
	FieldID  uint   `gorm:"primaryKey;autoIncrement:false" json:"field_id"`
	Value    string `gorm:"type:text" json:"value"`

	Field TicketField `gorm:"foreignKey:FieldID" json:"field,omitempty"`
}

func (TicketFieldValue) TableName() string { return "ticket_field_values" }

// CreateTicketFieldRequest 创建自定义字段请求
type CreateTicketFieldRequest struct {
	Name        string `json:"name" binding:"required"`
	FieldKey    string `json:"field_key" binding:"required"`
	FieldType   string `json:"field_type" binding:"required"`
	Options     string `json:"options,omitempty"`
	Required    bool   `json:"required"`
	SortOrder   int    `json:"sort_order"`
	Placeholder string `json:"placeholder,omitempty"`
}

// UpdateTicketFieldRequest 更新自定义字段请求
type UpdateTicketFieldRequest struct {
	Name        *string `json:"name,omitempty"`
	FieldType   *string `json:"field_type,omitempty"`
	Options     *string `json:"options,omitempty"`
	Required    *bool   `json:"required,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
	Placeholder *string `json:"placeholder,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
}

// AssignStrategy 分配策略
type AssignStrategy string

const (
	AssignStrategyLeastBusy  AssignStrategy = "least_busy"
	AssignStrategyRoundRobin AssignStrategy = "round_robin"
	AssignStrategySkillBased AssignStrategy = "skill_based"
	AssignStrategyRandom     AssignStrategy = "random"
)

// AssignConfig 自动分配配置
type AssignConfig struct {
	Base
	CategoryID      *uint         `json:"category_id,omitempty"` // 为空则全局配置
	Strategy        AssignStrategy `gorm:"size:50;default:'least_busy'" json:"strategy"`
	RoundRobinIndex int            `gorm:"default:0" json:"round_robin_index"` // 轮询计数器
}

func (AssignConfig) TableName() string { return "assign_configs" }

// AssignConfigRequest 分配配置请求
type AssignConfigRequest struct {
	CategoryID *uint         `json:"category_id,omitempty"`
	Strategy   AssignStrategy `json:"strategy" binding:"required"`
}

// SLAConfig SLA 规则配置
type SLAConfig struct {
	Base
	Priority          string  `gorm:"size:20;not null" json:"priority"`                       // low, medium, high, urgent
	CategoryID        *uint   `json:"category_id,omitempty"`                                   // nil = 全局规则
	ResponseMinutes   int     `gorm:"default:60" json:"response_minutes"`                      // 首次响应时限（分钟）
	ResolutionMinutes int     `gorm:"default:480" json:"resolution_minutes"`                   // 解决时限（分钟）
	WarningThreshold  float64 `gorm:"default:0.8" json:"warning_threshold"`                    // 预警阈值 0.0-1.0
	Enabled           bool    `gorm:"default:true" json:"enabled"`
}

func (SLAConfig) TableName() string { return "sla_configs" }

func (Ticket) TableName() string {
	return "tickets"
}
