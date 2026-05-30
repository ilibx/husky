package model

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

// Agent Agent 配置
type Agent struct {
	Base
	Name        string  `gorm:"size:100;not null" json:"name"`
	Description string  `gorm:"type:text" json:"description"`
	Type        string  `gorm:"size:50;not null" json:"type"` // llm, rule, hybrid
	Config      string  `gorm:"type:text" json:"config"`      // JSON 格式存储配置
	Model       string  `gorm:"size:100" json:"model"`        // LLM 模型名称
	Temperature float32 `gorm:"default:0.7" json:"temperature"`
	MaxTokens   int     `gorm:"default:2048" json:"max_tokens"`
	Enabled     bool    `gorm:"default:true" json:"enabled"`
	CreatedBy   uint    `gorm:"not null" json:"created_by"`

	Creator User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (Agent) TableName() string {
	return "agents"
}
