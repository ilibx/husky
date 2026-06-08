package model

// MCP 工具配置（支持内置工具和 MCP 服务）
type MCP struct {
	Base
	Name        string `gorm:"size:100;not null" json:"name"`
	Type        string `gorm:"size:20;default:'mcp'" json:"type"`          // "builtin" | "mcp"
	Key         string `gorm:"size:100;index" json:"key"`                  // 内置工具唯一标识（如 web_search）
	Description string `gorm:"type:text" json:"description,omitempty"`
	Endpoint    string `gorm:"size:255" json:"endpoint,omitempty"`        // MCP 服务地址（内置工具可选）
	Enabled     bool   `gorm:"default:true" json:"enabled"`
	CreatedBy   uint   `gorm:"not null" json:"created_by"`

	Creator User  `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	SkillID *uint `gorm:"index" json:"skill_id,omitempty"`
}

func (MCP) TableName() string { return "mcps" }
