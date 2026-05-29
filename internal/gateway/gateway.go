package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/husky/husky/internal/agent"
	"github.com/husky/husky/internal/channel/feishu"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/ticket"
	"github.com/husky/husky/pkg/logger"
)

var enrichers = map[ChannelType]UserEnricher{}

func RegisterEnricher(ch ChannelType, e UserEnricher) {
	enrichers[ch] = e
}

type Gateway struct {
	ticketSvc  ticket.Service
	agentEng   *agent.Engine
	feishuCli  *feishu.Client
	ticketGrp  TicketGroupHandler
	log        *logger.Logger
}

type TicketGroupHandler interface {
	CreateGroupForTicket(ctx context.Context, ticket *model.Ticket) (*model.TicketGroup, error)
	OnGroupMessage(ctx context.Context, chatID, userID, content string) error
}

func NewGateway(
	ticketSvc ticket.Service,
	agentEng *agent.Engine,
	feishuCli *feishu.Client,
	ticketGrp TicketGroupHandler,
	log *logger.Logger,
) *Gateway {
	return &Gateway{
		ticketSvc: ticketSvc,
		agentEng:  agentEng,
		feishuCli: feishuCli,
		ticketGrp: ticketGrp,
		log:       log,
	}
}

func (g *Gateway) HandleIncoming(ctx context.Context, msg *IncomingMessage) (*model.TicketResponse, error) {
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	if err := g.enrichUser(ctx, msg); err != nil {
		g.log.Warn("Gateway: failed to enrich user", "error", err, "channel", msg.Channel, "user_id", msg.UserID)
	}

	ticketReq := g.buildTicketRequest(msg)

	resp, err := g.ticketSvc.CreateTicket(ctx, ticketReq)
	if err != nil {
		return nil, fmt.Errorf("gateway: create ticket: %w", err)
	}

	go g.onTicketCreated(context.Background(), resp.ID, msg)

	return resp, nil
}

func (g *Gateway) HandleGroupMessage(ctx context.Context, msg *IncomingMessage) error {
	if g.ticketGrp == nil {
		return nil
	}
	return g.ticketGrp.OnGroupMessage(ctx, msg.ChatID, msg.UserID, msg.Content)
}

func (g *Gateway) enrichUser(ctx context.Context, msg *IncomingMessage) error {
	if msg.User == nil {
		msg.User = &UserContext{ChannelUserID: msg.UserID}
	}

	e, ok := enrichers[msg.Channel]
	if !ok {
		g.log.Debug("Gateway: no enricher for channel", "channel", msg.Channel)
		return nil
	}

	if err := e.Enrich(msg.User); err != nil {
		return fmt.Errorf("enrich user: %w", err)
	}
	return nil
}

func (g *Gateway) onTicketCreated(ctx context.Context, ticketID uint, msg *IncomingMessage) {
	t, err := g.ticketSvc.GetTicket(ctx, ticketID)
	if err != nil {
		return
	}

	if g.ticketGrp != nil && msg.ChatID != "" {
		if _, err := g.ticketGrp.CreateGroupForTicket(ctx, t); err != nil {
			g.log.Error("Gateway: failed to create group", "error", err, "ticket_id", ticketID)
		}
	}

	if g.agentEng != nil {
		g.agentEng.OnTicketCreated(ctx, t)
	}
}

func (g *Gateway) buildTicketRequest(msg *IncomingMessage) *model.CreateTicketRequest {
	prefix := string(msg.Channel)
	title := fmt.Sprintf("%s ticket - %s", prefix, truncateString(msg.Content, 50))

	metadata := map[string]interface{}{
		"channel":          string(msg.Channel),
		"channel_msg_id":   msg.MessageID,
		"channel_user_id":  msg.UserID,
		"raw_content":      msg.Content,
		"timestamp":        msg.Timestamp,
	}

	if msg.User != nil {
		userMeta := map[string]interface{}{
			"open_id":      msg.User.OpenID,
			"union_id":     msg.User.UnionID,
			"username":     msg.User.Username,
			"display_name": msg.User.DisplayName,
			"department":   msg.User.Department,
			"title":        msg.User.Title,
			"avatar_url":   msg.User.AvatarURL,
			"email":        msg.User.Email,
		}
		metadata["user_context"] = userMeta

		if msg.User.Department != "" {
			metadata["requester_dept"] = msg.User.Department
		}
		if msg.User.Title != "" {
			metadata["requester_title"] = msg.User.Title
		}
	}

	req := &model.CreateTicketRequest{
		Title:       title,
		Description: msg.Content,
		Priority:    "medium",
		RequesterID: msg.UserID,
		Channel:     string(msg.Channel),
		Metadata:    metadata,
	}
	return req
}

func (g *Gateway) SendToChannel(ctx context.Context, out *OutboundMessage) error {
	switch out.Channel {
	case ChannelLark:
		return g.sendLark(ctx, out)
	case ChannelDingTalk:
		return g.sendDingTalk(ctx, out)
	case ChannelWeCom:
		return g.sendWeCom(ctx, out)
	default:
		return fmt.Errorf("gateway: unsupported channel: %s", out.Channel)
	}
}

func (g *Gateway) sendLark(ctx context.Context, out *OutboundMessage) error {
	if g.feishuCli == nil {
		return fmt.Errorf("gateway: feishu client not configured")
	}
	if out.CardJSON != "" {
		return g.feishuCli.SendCardMessage(ctx, out.TargetID, out.CardJSON)
	}
	return g.feishuCli.SendMessageToChat(ctx, out.TargetID, "text", out.Content)
}

func (g *Gateway) sendDingTalk(ctx context.Context, out *OutboundMessage) error {
	return fmt.Errorf("gateway: dingtalk outgoing not implemented")
}

func (g *Gateway) sendWeCom(ctx context.Context, out *OutboundMessage) error {
	return fmt.Errorf("gateway: wecom outgoing not implemented")
}

func (g *Gateway) RequestHumanConfirmation(ctx context.Context, ticketID uint, channel ChannelType, targetID string, question string) error {
	out := &OutboundMessage{
		Channel:  channel,
		TargetID: targetID,
		Title:    "Confirmation required",
		Content:  question,
		MsgType:  "text",
		TicketID: ticketID,
	}
	return g.SendToChannel(ctx, out)
}

func (g *Gateway) NotifyTicketUpdate(ctx context.Context, ticketID uint, channel ChannelType, targetID string, message string) error {
	out := &OutboundMessage{
		Channel:  channel,
		TargetID: targetID,
		Title:    "Ticket update",
		Content:  message,
		MsgType:  "text",
		TicketID: ticketID,
	}
	return g.SendToChannel(ctx, out)
}

func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return s
}
