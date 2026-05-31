package model

// CreateSOPRequest 创建 SOP 请求
type CreateSOPRequest struct {
	Name              string `json:"name" binding:"required"`
	Description       string `json:"description,omitempty"`
	Version           string `json:"version,omitempty"`
	Steps             string `json:"steps,omitempty"`
	TriggerType       string `json:"trigger_type,omitempty"`
	TriggerConfig     string `json:"trigger_config,omitempty"`
	RiskLevel         string `json:"risk_level,omitempty"`
	NotificationConfig string `json:"notification_config,omitempty"`
}

// UpdateSOPRequest 更新 SOP 请求
type UpdateSOPRequest struct {
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	Version           *string `json:"version,omitempty"`
	Steps             *string `json:"steps,omitempty"`
	TriggerType       *string `json:"trigger_type,omitempty"`
	TriggerConfig     *string `json:"trigger_config,omitempty"`
	RiskLevel         *string `json:"risk_level,omitempty"`
	NotificationConfig *string `json:"notification_config,omitempty"`
	Status            *int    `json:"status,omitempty"`
}

// SOP 标准操作流程
type SOP struct {
	Base
	Name              string `gorm:"size:255;not null" json:"name"`
	Description       string `gorm:"type:text" json:"description"`
	Version           string `gorm:"size:50;not null" json:"version"`
	Steps             string `gorm:"type:text" json:"steps"`                        // JSON 格式存储步骤
	TriggerType       string `gorm:"size:50" json:"trigger_type"`                   // manual, auto, event
	TriggerConfig     string `gorm:"type:text" json:"trigger_config"`               // JSON 格式存储触发配置
	RiskLevel         string `gorm:"size:20;default:'low'" json:"risk_level"`       // low, medium, high, critical
	NotificationConfig string `gorm:"type:text" json:"notification_config"`          // JSON: 通知方式配置
	Status            int    `gorm:"default:0" json:"status"`                       // 0: 草稿，1: 激活，2: 停用
	CreatedBy         uint   `gorm:"not null" json:"created_by"`

	Creator User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (SOP) TableName() string {
	return "sops"
}
