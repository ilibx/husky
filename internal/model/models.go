package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

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
	Skills      string `gorm:"type:text" json:"skills"` // JSON array, e.g. ["network","hardware","account"]
	MaxLoad     int    `gorm:"default:10" json:"max_load"` // 最大并发处理工单数
	Status      int    `gorm:"default:1" json:"status"` // 1: 激活，0: 禁用
	Role        string `gorm:"size:50;default:'user'" json:"role"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// UpdateUserRequest 用户资料更新请求
type UpdateUserRequest struct {
	Username   string `json:"username,omitempty"`
	Avatar     string `json:"avatar,omitempty"`
	Department string `json:"department,omitempty"`
	Title      string `json:"title,omitempty"`
	Phone      string `json:"phone,omitempty"`
}

// ChangeRoleRequest 角色变更请求
type ChangeRoleRequest struct {
	Role string `json:"role" binding:"required"`
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
	Requester     User          `gorm:"foreignKey:RequesterID" json:"requester,omitempty"`
	Assignee      *User         `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Category      *Category     `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Comments      []Comment     `gorm:"foreignKey:TicketID" json:"comments,omitempty"`
	Attachments   []Attachment  `gorm:"foreignKey:TicketID" json:"attachments,omitempty"`
	OperationLogs []AuditLog    `gorm:"foreignKey:ResourceID" json:"operation_logs,omitempty"`
	Tags          []Tag         `gorm:"many2many:ticket_tags;" json:"tags,omitempty"`
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

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	SortOrder   int    `json:"sort_order,omitempty"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	SortOrder   int    `json:"sort_order,omitempty"`
}

// Category 工单分类
type Category struct {
	Base
	Name        string `gorm:"size:100;not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	Path        string `gorm:"size:500" json:"path"`
	SortOrder   int    `gorm:"default:0" json:"sort_order"`
	Status      int    `gorm:"default:1" json:"status"` // 1: 激活，0: 禁用
	Type        string `gorm:"size:50;default:'ticket'" json:"type"` // ticket, knowledge, both
	
	Parent *Category `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// CategoryTreeNode 分类树节点
type CategoryTreeNode struct {
	ID        uint                `json:"id"`
	Name      string              `json:"name"`
	ParentID  *uint               `json:"parent_id,omitempty"`
	SortOrder int                 `json:"sort_order"`
	Children  []CategoryTreeNode  `json:"children,omitempty"`
}

// BuildCategoryTree 将扁平分类列表构建为树
func BuildCategoryTree(categories []Category) []CategoryTreeNode {
	byParent := make(map[uint][]Category)
	var roots []Category
	for _, c := range categories {
		if c.ParentID == nil {
			roots = append(roots, c)
		} else {
			byParent[*c.ParentID] = append(byParent[*c.ParentID], c)
		}
	}

	var build func(parentID uint) []CategoryTreeNode
	build = func(parentID uint) []CategoryTreeNode {
		var nodes []CategoryTreeNode
		for _, c := range byParent[parentID] {
			node := CategoryTreeNode{
				ID:        c.ID,
				Name:      c.Name,
				ParentID:  c.ParentID,
				SortOrder: c.SortOrder,
				Children:  build(c.ID),
			}
			nodes = append(nodes, node)
		}
		return nodes
	}

	var tree []CategoryTreeNode
	for _, r := range roots {
		node := CategoryTreeNode{
			ID:        r.ID,
			Name:      r.Name,
			ParentID:  r.ParentID,
			SortOrder: r.SortOrder,
			Children:  build(r.ID),
		}
		tree = append(tree, node)
	}
	return tree
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

// TicketWatcher 工单关注者
type TicketWatcher struct {
	Base
	TicketID uint `gorm:"uniqueIndex:idx_ticket_user;not null" json:"ticket_id"`
	UserID   uint `gorm:"uniqueIndex:idx_ticket_user;not null" json:"user_id"`

	Ticket Ticket `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
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

// CreateSOPRequest 创建 SOP 请求
type CreateSOPRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description,omitempty"`
	Version       string `json:"version,omitempty"`
	Steps         string `json:"steps,omitempty"`
	TriggerType   string `json:"trigger_type,omitempty"`
	TriggerConfig string `json:"trigger_config,omitempty"`
}

// UpdateSOPRequest 更新 SOP 请求
type UpdateSOPRequest struct {
	Name          *string `json:"name,omitempty"`
	Description   *string `json:"description,omitempty"`
	Version       *string `json:"version,omitempty"`
	Steps         *string `json:"steps,omitempty"`
	TriggerType   *string `json:"trigger_type,omitempty"`
	TriggerConfig *string `json:"trigger_config,omitempty"`
	Status        *int    `json:"status,omitempty"`
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

// CreateAgentRequest 创建 Agent 请求
type CreateAgentRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description,omitempty"`
	Type        string  `json:"type" binding:"required"` // llm, rule, hybrid
	Config      string  `json:"config,omitempty"`
	Model       string  `json:"model,omitempty"`
	Temperature float32 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
}

// UpdateAgentRequest 更新 Agent 请求
type UpdateAgentRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Type        *string  `json:"type,omitempty"`
	Config      *string  `json:"config,omitempty"`
	Model       *string  `json:"model,omitempty"`
	Temperature *float32 `json:"temperature,omitempty"`
	MaxTokens   *int     `json:"max_tokens,omitempty"`
	Enabled     *bool    `json:"enabled,omitempty"`
}

// ChannelConfig 渠道配置
type ChannelConfig struct {
	Base
	Name    string `gorm:"size:100;not null" json:"name"`
	Type    string `gorm:"size:50;not null;index" json:"type"` // lark, dingtalk, wecom, email
	Config  string `gorm:"type:text" json:"config"`            // JSON 格式配置
	Enabled bool   `gorm:"default:true" json:"enabled"`
	Status  int    `gorm:"default:1" json:"status"`
}

// CreateChannelConfigRequest 创建渠道配置请求
type CreateChannelConfigRequest struct {
	Name   string `json:"name" binding:"required"`
	Type   string `json:"type" binding:"required"`
	Config string `json:"config,omitempty"`
}

// UpdateChannelConfigRequest 更新渠道配置请求
type UpdateChannelConfigRequest struct {
	Name    *string `json:"name,omitempty"`
	Config  *string `json:"config,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
	Status  *int    `json:"status,omitempty"`
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
	Permissions string `gorm:"type:text" json:"permissions"` // JSON array of permission keys, e.g. ["ticket:create","ticket:read"]
	Status      int    `gorm:"default:1" json:"status"`
}

// ScanPermissions 解析权限 JSON 到字符串切片
func (r *Role) ScanPermissions(out *[]string) error {
	if r.Permissions == "" {
		*out = nil
		return nil
	}
	return json.Unmarshal([]byte(r.Permissions), out)
}

// DefaultAdminPermissions 返回 admin 角色的默认权限列表
func DefaultAdminPermissions() []string {
	return []string{
		"ticket:manage", "user:manage", "knowledge:manage",
		"agent:manage", "sop:manage", "category:manage",
		"department:manage", "channel:manage", "role:manage",
		"stats:manage", "webhook:manage",
	}
}

// DefaultAgentPermissions 返回 agent 角色的默认权限列表
func DefaultAgentPermissions() []string {
	return []string{
		"ticket:create", "ticket:read", "ticket:update", "ticket:assign",
		"knowledge:read", "stats:read",
	}
}

// CreateDepartmentRequest 创建部门请求
type CreateDepartmentRequest struct {
	Name      string `json:"name" binding:"required"`
	Code      string `json:"code" binding:"required"`
	ParentID  *uint  `json:"parent_id,omitempty"`
	ManagerID *uint  `json:"manager_id,omitempty"`
}

// UpdateDepartmentRequest 更新部门请求
type UpdateDepartmentRequest struct {
	Name      *string `json:"name,omitempty"`
	Code      *string `json:"code,omitempty"`
	ParentID  *uint   `json:"parent_id,omitempty"`
	ManagerID *uint   `json:"manager_id,omitempty"`
	Status    *int    `json:"status,omitempty"`
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
	UserID       uint       `gorm:"not null;index" json:"user_id"`
	Type         string     `gorm:"size:50;not null" json:"type"` // ticket_assigned, ticket_status, ticket_comment, system
	Title        string     `gorm:"size:255;not null" json:"title"`
	Content      string     `gorm:"type:text;not null" json:"content"`
	ReferenceID  uint       `json:"reference_id,omitempty"`   // 关联工单/资源 ID
	ReferenceType string    `gorm:"size:50" json:"reference_type,omitempty"` // ticket, knowledge, etc.
	IsRead       bool       `gorm:"default:false;index" json:"is_read"`
	Status       string     `gorm:"size:20;default:'unread'" json:"status"` // unread, read, sent, failed
	SentAt       *time.Time `json:"sent_at,omitempty"`
	Error        string     `gorm:"type:text" json:"error,omitempty"`
	
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Satisfaction 满意度评价
type Satisfaction struct {
	Base
	TicketID  uint   `gorm:"uniqueIndex;not null" json:"ticket_id"`
	Score     int    `gorm:"not null" json:"score"` // 1-5
	Comment   string `gorm:"type:text" json:"comment,omitempty"`
	CreatedBy uint   `gorm:"not null" json:"created_by"`

	Ticket Ticket `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
	User   User   `gorm:"foreignKey:CreatedBy" json:"user,omitempty"`
}

// WebhookMessageRecord 渠道消息记录
type WebhookMessageRecord struct {
	Base
	Channel     string `gorm:"size:50;not null;index" json:"channel"`
	MessageID   string `gorm:"size:255;index" json:"message_id"`
	UserID      string `gorm:"size:255;index" json:"user_id"`
	Content     string `gorm:"type:text" json:"content"`
	TicketID    *uint  `json:"ticket_id,omitempty"`
	RawPayload  string `gorm:"type:text" json:"raw_payload"`
	Status      string `gorm:"size:20;default:'received'" json:"status"` // received, processed, failed
}

// TicketGroup 工单-飞书群绑定
type TicketGroup struct {
	Base
	TicketID uint   `gorm:"uniqueIndex;not null" json:"ticket_id"`
	GroupID  string `gorm:"size:255;not null" json:"group_id"` // feishu chat_id
	GroupName string `gorm:"size:255" json:"group_name"`
	JoinLink string `gorm:"type:text" json:"join_link"`
	Status   string `gorm:"size:20;default:'active'" json:"status"` // active, archived, deleted
}

func (TicketGroup) TableName() string { return "ticket_groups" }

// Workflow 工作流实例
type Workflow struct {
	Base
	TicketID  uint   `gorm:"not null;index" json:"ticket_id"`
	SOPID     uint   `gorm:"not null" json:"sop_id"`
	Status    string `gorm:"size:20;default:'pending'" json:"status"` // pending, running, completed, failed, cancelled
	CurrentStep int  `gorm:"default:0" json:"current_step"`
	TotalSteps  int  `gorm:"default:0" json:"total_steps"`

	Ticket Ticket `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
	SOP    SOP    `gorm:"foreignKey:SOPID" json:"sop,omitempty"`
	Steps  []WorkflowStep `gorm:"foreignKey:WorkflowID" json:"steps,omitempty"`
}

func (Workflow) TableName() string { return "workflows" }

// WorkflowStep 工作流步骤
type WorkflowStep struct {
	Base
	WorkflowID uint   `gorm:"not null;index" json:"workflow_id"`
	StepIndex  int    `gorm:"not null" json:"step_index"`
	Name       string `gorm:"size:255" json:"name"`
	Type       string `gorm:"size:50;not null" json:"type"` // agent, human, condition, notification
	AgentID    *uint  `json:"agent_id,omitempty"`            // 执行的 Agent
	AssigneeID *uint  `json:"assignee_id,omitempty"`         // 人工步骤负责人
	Config     string `gorm:"type:text" json:"config"`       // 步骤配置 JSON
	Status     string `gorm:"size:20;default:'pending'" json:"status"` // pending, running, completed, failed, skipped
	Result     string `gorm:"type:text" json:"result,omitempty"`       // 执行结果

	// HITL 字段
	Suggestion       string `gorm:"type:text" json:"suggestion,omitempty"`                  // AI 建议内容 (JSON)
	SuggestedBy      *uint  `json:"suggested_by,omitempty"`                                 // 建议来源 Agent ID
	Decision         string `gorm:"size:20" json:"decision,omitempty"`                      // approve, reject, revise
	Feedback         string `gorm:"type:text" json:"feedback,omitempty"`                    // 人工反馈
	ApprovedBy       *uint  `json:"approved_by,omitempty"`                                  // 审核人 User ID
	RejectionReason  string `gorm:"type:text" json:"rejection_reason,omitempty"`            // 拒绝原因

	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	Assignee *User  `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Agent    *Agent `gorm:"foreignKey:AgentID" json:"agent,omitempty"`
}

func (WorkflowStep) TableName() string { return "workflow_steps" }

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

func (Satisfaction) TableName() string {
	return "satisfactions"
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
	CategoryID *uint         `json:"category_id,omitempty"` // 为空则全局配置
	Strategy   AssignStrategy `gorm:"size:50;default:'least_busy'" json:"strategy"`
	RoundRobinIndex int       `gorm:"default:0" json:"round_robin_index"` // 轮询计数器
}

func (AssignConfig) TableName() string { return "assign_configs" }

// AssignConfigRequest 分配配置请求
type AssignConfigRequest struct {
	CategoryID *uint         `json:"category_id,omitempty"`
	Strategy   AssignStrategy `json:"strategy" binding:"required"`
}

// TicketRelation 工单关联关系
type TicketRelation struct {
	Base
	TicketID    uint   `gorm:"uniqueIndex:idx_ticket_rel;not null" json:"ticket_id"`
	RelatedID   uint   `gorm:"uniqueIndex:idx_ticket_rel;not null" json:"related_id"`
	RelationType string `gorm:"size:50;not null" json:"relation_type"` // parent, child, related, duplicate, blocks, blocked_by
}

func (TicketRelation) TableName() string { return "ticket_relations" }

// TicketField 工单自定义字段定义
type TicketField struct {
	Base
	Name        string `gorm:"size:100;not null" json:"name"`
	FieldKey    string `gorm:"size:100;uniqueIndex;not null" json:"field_key"`
	FieldType   string `gorm:"size:50;not null;default:'text'" json:"field_type"` // text, textarea, select, multi_select, number, date, boolean
	Options     string `gorm:"type:text" json:"options,omitempty"` // JSON array for select/multi_select
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

// BotConfig 机器人配置
type BotConfig struct {
	Base
	Channel       string `gorm:"size:50;uniqueIndex;not null" json:"channel"` // lark, dingtalk, wecom
	WelcomeMsg    string `gorm:"type:text" json:"welcome_msg,omitempty"`      // 欢迎语
	Signature     string `gorm:"type:text" json:"signature,omitempty"`        // 签名
	Enabled       bool   `gorm:"default:true" json:"enabled"`
}

func (BotConfig) TableName() string { return "bot_configs" }

// SLAConfig SLA 规则配置
type SLAConfig struct {
	Base
	Priority          string `gorm:"size:20;not null" json:"priority"`                    // low, medium, high, urgent
	CategoryID        *uint  `json:"category_id,omitempty"`                                // nil = 全局规则
	ResponseMinutes   int    `gorm:"default:60" json:"response_minutes"`                   // 首次响应时限（分钟）
	ResolutionMinutes int    `gorm:"default:480" json:"resolution_minutes"`                // 解决时限（分钟）
	WarningThreshold  float64 `gorm:"default:0.8" json:"warning_threshold"`                // 预警阈值 0.0-1.0
	Enabled           bool   `gorm:"default:true" json:"enabled"`
}

func (SLAConfig) TableName() string { return "sla_configs" }

// WebhookConfig Webhook 配置
type WebhookConfig struct {
	Base
	Name    string `gorm:"size:100;not null" json:"name"`
	URL     string `gorm:"size:500;not null" json:"url"`
	Secret  string `gorm:"size:255" json:"secret,omitempty"`
	Events  string `gorm:"type:text" json:"events"` // JSON 数组: ["sla_breach","sla_warning","ticket_created"]
	Enabled bool   `gorm:"default:true" json:"enabled"`
}

func (WebhookConfig) TableName() string { return "webhook_configs" }
