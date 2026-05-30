package model

import "time"

// Workflow 工作流实例
type Workflow struct {
	Base
	TicketID    uint   `gorm:"not null;index" json:"ticket_id"`
	SOPID       uint   `gorm:"not null" json:"sop_id"`
	Status      string `gorm:"size:20;default:'pending'" json:"status"` // pending, running, completed, failed, cancelled
	CurrentStep int    `gorm:"default:0" json:"current_step"`
	TotalSteps  int    `gorm:"default:0" json:"total_steps"`

	Ticket Ticket         `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
	SOP    SOP            `gorm:"foreignKey:SOPID" json:"sop,omitempty"`
	Steps  []WorkflowStep `gorm:"foreignKey:WorkflowID" json:"steps,omitempty"`
}

// WorkflowStep 工作流步骤
type WorkflowStep struct {
	Base
	WorkflowID    uint       `gorm:"not null;index" json:"workflow_id"`
	StepIndex     int        `gorm:"not null" json:"step_index"`
	Name          string     `gorm:"size:255" json:"name"`
	Type          string     `gorm:"size:50;not null" json:"type"` // agent, human, condition, notification
	AgentID       *uint      `json:"agent_id,omitempty"`           // 执行的 Agent
	AssigneeID    *uint      `json:"assignee_id,omitempty"`        // 人工步骤负责人
	Config        string     `gorm:"type:text" json:"config"`      // 步骤配置 JSON
	Status        string     `gorm:"size:20;default:'pending'" json:"status"` // pending, running, completed, failed, skipped
	Result        string     `gorm:"type:text" json:"result,omitempty"`       // 执行结果

	// HITL 字段
	Suggestion      string `gorm:"type:text" json:"suggestion,omitempty"`                  // AI 建议内容 (JSON)
	SuggestedBy     *uint  `json:"suggested_by,omitempty"`                                 // 建议来源 Agent ID
	Decision        string `gorm:"size:20" json:"decision,omitempty"`                      // approve, reject, revise
	Feedback        string `gorm:"type:text" json:"feedback,omitempty"`                    // 人工反馈
	ApprovedBy      *uint  `json:"approved_by,omitempty"`                                  // 审核人 User ID
	RejectionReason string `gorm:"type:text" json:"rejection_reason,omitempty"`            // 拒绝原因

	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	Assignee *User  `gorm:"foreignKey:AssigneeID" json:"assignee,omitempty"`
	Agent    *Agent `gorm:"foreignKey:AgentID" json:"agent,omitempty"`
}

func (Workflow) TableName() string { return "workflows" }

func (WorkflowStep) TableName() string { return "workflow_steps" }
