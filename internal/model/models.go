package model

import (
	"time"

	"gorm.io/gorm"
)

// TicketStatus 工单状态
type TicketStatus string

const (
	TicketStatusOpen       TicketStatus = "open"
	TicketStatusInProgress TicketStatus = "in_progress"
	TicketStatusPending    TicketStatus = "pending"
	TicketStatusResolved   TicketStatus = "resolved"
	TicketStatusClosed     TicketStatus = "closed"
)

// Priority 优先级
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityUrgent   Priority = "urgent"
)

// Base 基础模型，包含通用字段
type Base struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// User 用户模型
type User struct {
	Base
	Email       string `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Phone       string `gorm:"size:20" json:"phone"`
	Username    string `gorm:"size:100;not null" json:"username"`
	Password    string `gorm:"size:255;not null" json:"-"` // 不返回密码
	Avatar      string `gorm:"size:500" json:"avatar"`
	Department  string `gorm:"size:100" json:"department"`
	Title       string `gorm:"size:100" json:"title"`
	Status      int    `gorm:"default:1" json:"status"` // 1: 激活，0: 禁用
	Role        string `gorm:"size:50;default:'user'" json:"role"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// Ticket 工单模型
type Ticket struct {
	Base
	TicketNo    string `gorm:"uniqueIndex;size:50;not null" json:"ticket_no"`
	Title       string `gorm:"size:255;not null" json:"title"`
	Description string `gorm:"type:text;not null" json:"description"`
	Status      string `gorm:"size:50;default:'open'" json:"status"` // open, in_progress, pending, resolved, closed
	Priority    string `gorm:"size:20;default:'medium'" json:"priority"` // low, medium, high, urgent
	Type        string `gorm:"size:50" json:"type"` // incident, service_request, problem, change
	CategoryID  uint   `json:"category_id,omitempty"`
	AssigneeID  *uint  `json:"assignee_id,omitempty"`
	RequesterID uint   `json:"requester_id"`
	Source      string `gorm:"size:50;default:'web'" json:"source"` // web, email, im, api
	SLAStatus   string `gorm:"size:50;default:'normal'" json:"sla_status"` // normal, warning, breached
	DueAt       *time.Time `json:"due_at,omitempty"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	ClosedAt    *time.Time `json:"closed_at,omitempty"`
	
	// 关联
	Requester User          `gorm:"foreignKey:RequesterID" json:"requester,omitempty"`
	Assignee  *User         `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Category  *Category     `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Comments  []Comment     `gorm:"foreignKey:TicketID" json:"comments,omitempty"`
	Attachments []Attachment `gorm:"foreignKey:TicketID" json:"attachments,omitempty"`
}

// Category 工单分类
type Category struct {
	Base
	Name        string `gorm:"size:100;not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	Path        string `gorm:"size:500" json:"path"` // 分类路径，用于层级结构
	SortOrder   int    `gorm:"default:0" json:"sort_order"`
	Status      int    `gorm:"default:1" json:"status"` // 1: 激活，0: 禁用
	
	Parent *Category `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
}

// Comment 工单评论
type Comment struct {
	Base
	TicketID  uint   `gorm:"not null;index" json:"ticket_id"`
	UserID    uint   `gorm:"not null" json:"user_id"`
	Content   string `gorm:"type:text;not null" json:"content"`
	IsInternal bool  `gorm:"default:false" json:"is_internal"` // 是否内部评论
	Visibility string `gorm:"size:20;default:'public'" json:"visibility"` // public, internal, private
	
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Attachment 附件
type Attachment struct {
	Base
	TicketID  uint   `gorm:"not null;index" json:"ticket_id"`
	FileName  string `gorm:"size:255;not null" json:"file_name"`
	FileSize  int64  `gorm:"not null" json:"file_size"`
	FileType  string `gorm:"size:100" json:"file_type"`
	FileURL   string `gorm:"size:500;not null" json:"file_url"`
	UploadedBy uint   `gorm:"not null" json:"uploaded_by"`
	
	Uploader User `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
}

// Knowledge 知识库文章
type Knowledge struct {
	Base
	Title     string `gorm:"size:255;not null" json:"title"`
	Content   string `gorm:"type:text;not null" json:"content"`
	Summary   string `gorm:"type:text" json:"summary"`
	CategoryID uint  `json:"category_id,omitempty"`
	Tags      string `gorm:"size:500" json:"tags"` // 逗号分隔的标签
	Status    int    `gorm:"default:0" json:"status"` // 0: 草稿，1: 发布，2: 归档
	ViewCount int    `gorm:"default:0" json:"view_count"`
	UsefulCount int  `gorm:"default:0" json:"useful_count"`
	CreatedBy uint   `gorm:"not null" json:"created_by"`
	
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Author   User      `gorm:"foreignKey:CreatedBy" json:"author,omitempty"`
}

// SOP 标准操作流程
type SOP struct {
	Base
	Name        string `gorm:"size:255;not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Version     string `gorm:"size:50;not null" json:"version"`
	Steps       string `gorm:"type:text" json:"steps"` // JSON 格式存储步骤
	TriggerType string `gorm:"size:50" json:"trigger_type"` // manual, auto, event
	TriggerConfig string `gorm:"type:text" json:"trigger_config"` // JSON 格式存储触发配置
	Status      int    `gorm:"default:0" json:"status"` // 0: 草稿，1: 激活，2: 停用
	CreatedBy   uint   `gorm:"not null" json:"created_by"`
	
	Creator User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// Agent Agent 配置
type Agent struct {
	Base
	Name        string `gorm:"size:100;not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Type        string `gorm:"size:50;not null" json:"type"` // llm, rule, hybrid
	Config      string `gorm:"type:text" json:"config"` // JSON 格式存储配置
	Model       string `gorm:"size:100" json:"model"` // LLM 模型名称
	Temperature float32 `gorm:"default:0.7" json:"temperature"`
	MaxTokens   int    `gorm:"default:2048" json:"max_tokens"`
	Enabled     bool   `gorm:"default:true" json:"enabled"`
	CreatedBy   uint   `gorm:"not null" json:"created_by"`
	
	Creator User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// Role 角色
type Role struct {
	Base
	Name        string `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Permissions string `gorm:"type:text" json:"permissions"` // JSON 格式存储权限列表
	Status      int    `gorm:"default:1" json:"status"`
}

// Department 部门
type Department struct {
	Base
	Name        string `gorm:"size:100;not null" json:"name"`
	Code        string `gorm:"size:50;uniqueIndex;not null" json:"code"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	Path        string `gorm:"size:500" json:"path"`
	ManagerID   *uint  `json:"manager_id,omitempty"`
	Status      int    `gorm:"default:1" json:"status"`
	
	Parent *Department `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Manager *User       `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
}

// Notification 通知记录
type Notification struct {
	Base
	UserID      uint   `gorm:"not null;index" json:"user_id"`
	Type        string `gorm:"size:50;not null" json:"type"` // email, sms, im, push
	Title       string `gorm:"size:255;not null" json:"title"`
	Content     string `gorm:"type:text;not null" json:"content"`
	Status      string `gorm:"size:20;default:'pending'" json:"status"` // pending, sent, failed
	SentAt      *time.Time `json:"sent_at,omitempty"`
	Error       string `gorm:"type:text" json:"error,omitempty"`
	
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// AuditLog 审计日志
type AuditLog struct {
	Base
	UserID      uint   `gorm:"index" json:"user_id"`
	Action      string `gorm:"size:100;not null" json:"action"`
	Resource    string `gorm:"size:100;not null" json:"resource"`
	ResourceID  uint   `json:"resource_id"`
	OldValue    string `gorm:"type:text" json:"old_value,omitempty"` // JSON 格式
	NewValue    string `gorm:"type:text" json:"new_value,omitempty"` // JSON 格式
	IP          string `gorm:"size:50" json:"ip"`
	UserAgent   string `gorm:"size:500" json:"user_agent"`
	
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

func (Ticket) TableName() string {
	return "tickets"
}

func (Category) TableName() string {
	return "categories"
}

func (Comment) TableName() string {
	return "comments"
}

func (Attachment) TableName() string {
	return "attachments"
}

func (Knowledge) TableName() string {
	return "knowledge"
}

func (SOP) TableName() string {
	return "sops"
}

func (Agent) TableName() string {
	return "agents"
}

func (Role) TableName() string {
	return "roles"
}

func (Department) TableName() string {
	return "departments"
}

func (Notification) TableName() string {
	return "notifications"
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
