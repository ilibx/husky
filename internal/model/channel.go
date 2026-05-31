package model

// ChannelConfig 渠道配置
type ChannelConfig struct {
	Base
	Name    string `gorm:"size:100;not null" json:"name"`
	Type    string `gorm:"size:50;not null;index" json:"type"` // lark, dingtalk, wecom, email
	Config  string `gorm:"type:text" json:"config"`            // JSON 格式配置
	Enabled bool   `gorm:"default:true" json:"enabled"`
	Status  int    `gorm:"default:1" json:"status"`

	// 渠道凭据（敏感信息，前端展示时脱敏）
	AppID       string `gorm:"size:255" json:"app_id,omitempty"`       // 飞书 AppID / 钉钉 AppKey / 企微 CorpID
	AppSecret   string `gorm:"size:500" json:"app_secret,omitempty"`   // 飞书 AppSecret / 钉钉 AppSecret / 企微 CorpSecret
	AgentID     string `gorm:"size:100" json:"agent_id,omitempty"`     // 钉钉/企微 AgentID
	WebhookURL  string `gorm:"size:500" json:"webhook_url,omitempty"`  // 自定义 Webhook 地址
	VerifyToken string `gorm:"size:255" json:"verify_token,omitempty"` // 飞书验证令牌
	EncryptKey  string `gorm:"size:255" json:"encrypt_key,omitempty"`  // 飞书加密 Key
}

// CreateChannelConfigRequest 创建渠道配置请求
type CreateChannelConfigRequest struct {
	Name        string  `json:"name" binding:"required"`
	Type        string  `json:"type" binding:"required"`
	Config      string  `json:"config,omitempty"`
	AppID       *string `json:"app_id,omitempty"`
	AppSecret   *string `json:"app_secret,omitempty"`
	AgentID     *string `json:"agent_id,omitempty"`
	WebhookURL  *string `json:"webhook_url,omitempty"`
	VerifyToken *string `json:"verify_token,omitempty"`
	EncryptKey  *string `json:"encrypt_key,omitempty"`
}

// UpdateChannelConfigRequest 更新渠道配置请求
type UpdateChannelConfigRequest struct {
	Name        *string `json:"name,omitempty"`
	Config      *string `json:"config,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
	Status      *int    `json:"status,omitempty"`
	AppID       *string `json:"app_id,omitempty"`
	AppSecret   *string `json:"app_secret,omitempty"`
	AgentID     *string `json:"agent_id,omitempty"`
	WebhookURL  *string `json:"webhook_url,omitempty"`
	VerifyToken *string `json:"verify_token,omitempty"`
	EncryptKey  *string `json:"encrypt_key,omitempty"`
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
