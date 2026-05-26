package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LarkWebhookEvent 飞书 Webhook 事件
type LarkWebhookEvent struct {
	Challenge string          `json:"challenge,omitempty"` // URL 验证挑战码
	Token     string          `json:"token,omitempty"`     // 验证 Token
	Type      string          `json:"type,omitempty"`      // 事件类型：url_verification, event_callback
	Event     *LarkEventData  `json:"event,omitempty"`
	Schema    string          `json:"schema,omitempty"`
	Header    *LarkEventHeader `json:"header,omitempty"`
}

// LarkEventHeader 飞书事件头
type LarkEventHeader struct {
	EventID    string `json:"event_id,omitempty"`
	EventType  string `json:"event_type,omitempty"`
	CreateTime string `json:"create_time,omitempty"`
	Token      string `json:"token,omitempty"`
	AppID      string `json:"app_id,omitempty"`
	TenantKey  string `json:"tenant_key,omitempty"`
}

// LarkEventData 飞书事件数据
type LarkEventData struct {
	MessageID string        `json:"message_id,omitempty"`
	UserID    string        `json:"user_id,omitempty"`
	OpenID    string        `json:"open_id,omitempty"`
	UnionID   string        `json:"union_id,omitempty"`
	TenantKey string        `json:"tenant_key,omitempty"`
	Content   string        `json:"content,omitempty"`
	Text      LarkTextContent `json:"text,omitempty"`
	MessageType string      `json:"message_type,omitempty"`
	ChatID    string        `json:"chat_id,omitempty"`
}

// LarkTextContent 飞书文本内容
type LarkTextContent struct {
	Content string `json:"content"`
}

// LarkTicketRequest 从飞书消息创建的工单请求
type LarkTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	RequesterID string `json:"requester_id"` // 飞书用户 ID
	Channel     string `json:"channel"`      // 渠道标识：lark
	RawMessage  string `json:"raw_message"`  // 原始消息内容
	MessageID   string `json:"message_id"`   // 飞书消息 ID
}

// LarkCallbackResponse 飞书回调响应
type LarkCallbackResponse struct {
	Challenge string `json:"challenge,omitempty"` // URL 验证时返回
	Success   bool   `json:"success,omitempty"`
	Message   string `json:"message,omitempty"`
}

// DingTalkWebhookEvent 钉钉 Webhook 事件
type DingTalkWebhookEvent struct {
	MsgType     string           `json:"msgtype"`
	Text        *DingTalkText    `json:"text,omitempty"`
	At          *DingTalkAt      `json:"at,omitempty"`
	RobotCode   string           `json:"robotCode"`
	ConversationID string        `json:"conversationId"`
	ChatbotUserID string         `json:"chatbotUserId"`
	SenderID    string           `json:"senderId"`
	SenderNick  string           `json:"senderNick"`
	Webhook     string           `json:"webhook"`
	Timestamp   string           `json:"timestamp"`
}

// DingTalkText 钉钉文本内容
type DingTalkText struct {
	Content string `json:"content"`
}

// DingTalkAt @信息
type DingTalkAt struct {
	IsInCname bool     `json:"isInCname"`
	RobotCode string   `json:"robotCode"`
	AtMobiles []string `json:"atMobiles"`
	AtUserIds []string `json:"atUserIds"`
}

// WeComWebhookEvent 企业微信 Webhook 事件
type WeComWebhookEvent struct {
	MsgType   string         `json:"msgtype"`
	Text      *WeComText     `json:"text,omitempty"`
	FromUserName string      `json:"FromUserName"`
	ToUserName   string      `json:"ToUserName"`
	CreateTime   int64       `json:"CreateTime"`
	MsgID        string      `json:"MsgId"`
	AgentID      int         `json:"AgentID"`
	EnterpriseCorpID string `json:"EnterpriseCorpID"`
}

// WeComText 企业微信文本内容
type WeComText struct {
	Content string `json:"content"`
}

// ChannelType 渠道类型
type ChannelType string

const (
	ChannelLark       ChannelType = "lark"
	ChannelDingTalk   ChannelType = "dingtalk"
	ChannelWeCom      ChannelType = "wecom"
	ChannelEmail      ChannelType = "email"
	ChannelWeb        ChannelType = "web"
)

// WebhookMessage 通用 Webhook 消息结构
type WebhookMessage struct {
	Channel     ChannelType `json:"channel"`
	MessageID   string      `json:"message_id"`
	UserID      string      `json:"user_id"`
	Content     string      `json:"content"`
	RawPayload  interface{} `json:"raw_payload"`
	CreatedAt   time.Time   `json:"created_at"`
}

// BeforeCreate 钩子：生成 UUID（如果需要）
func (w *WebhookMessage) BeforeCreate(tx *gorm.DB) error {
	if w.MessageID == "" {
		w.MessageID = uuid.New().String()
	}
	if w.CreatedAt.IsZero() {
		w.CreatedAt = time.Now()
	}
	return nil
}
