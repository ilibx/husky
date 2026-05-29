package gateway

import "time"

type ChannelType string

const (
	ChannelLark     ChannelType = "lark"
	ChannelDingTalk ChannelType = "dingtalk"
	ChannelWeCom    ChannelType = "wecom"
	ChannelEmail    ChannelType = "email"
	ChannelWeb      ChannelType = "web"
	ChannelAPI      ChannelType = "api"
)

type UserContext struct {
	ChannelUserID string
	OpenID        string
	UnionID       string
	Username      string
	DisplayName   string
	Department    string
	Title         string
	AvatarURL     string
	Email         string
	Phone         string
	Extra         map[string]interface{}
}

type IncomingMessage struct {
	Channel        ChannelType
	MessageID      string
	UserID         string
	Content        string
	ChatID         string
	Raw            interface{}
	User           *UserContext
	Timestamp      time.Time
}

type OutboundMessage struct {
	Channel   ChannelType
	TargetID  string
	Title     string
	Content   string
	MsgType   string
	CardJSON  string
	TicketID  uint
}

type TicketEvent struct {
	TicketID uint
	Action   string
	Message  string
	Channel  ChannelType
	User     *UserContext
}

type UserEnricher interface {
	Enrich(ctx *UserContext) error
}
