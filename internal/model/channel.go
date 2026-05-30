package model

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

// WebhookMessageRecord 渠道消息记录
type WebhookMessageRecord struct {
	Base
	Channel    string `gorm:"size:50;not null;index" json:"channel"`
	MessageID  string `gorm:"size:255;index" json:"message_id"`
	UserID     string `gorm:"size:255;index" json:"user_id"`
	Content    string `gorm:"type:text" json:"content"`
	TicketID   *uint  `json:"ticket_id,omitempty"`
	RawPayload string `gorm:"type:text" json:"raw_payload"`
	Status     string `gorm:"size:20;default:'received'" json:"status"` // received, processed, failed
}

// BotConfig 机器人配置
type BotConfig struct {
	Base
	Channel    string `gorm:"size:50;uniqueIndex;not null" json:"channel"` // lark, dingtalk, wecom
	WelcomeMsg string `gorm:"type:text" json:"welcome_msg,omitempty"`      // 欢迎语
	Signature  string `gorm:"type:text" json:"signature,omitempty"`        // 签名
	Enabled    bool   `gorm:"default:true" json:"enabled"`
}

func (BotConfig) TableName() string { return "bot_configs" }

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
