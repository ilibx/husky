package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/husky/husky/internal/agent"
	"github.com/husky/husky/internal/channel/feishu"
	"github.com/husky/husky/internal/intent"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/ticket"
	"github.com/husky/husky/pkg/logger"
)

var enrichers = map[ChannelType]UserEnricher{}

func RegisterEnricher(ch ChannelType, e UserEnricher) {
	enrichers[ch] = e
}

type Gateway struct {
	ticketSvc   ticket.Service
	agentEng    *agent.Engine
	feishuCli   *feishu.Client
	ticketGrp   TicketGroupHandler
	bot         BotHandler
	botCfgRepo  BotConfigProvider
	intentSvc   *intent.Service
	log         *logger.Logger
}

type TicketGroupHandler interface {
	CreateGroupForTicket(ctx context.Context, ticket *model.Ticket) (*model.TicketGroup, error)
	OnGroupMessage(ctx context.Context, chatID, userID, content string) error
}

type BotConfigProvider interface {
	GetBotConfig(ctx context.Context, channel string) (*model.BotConfig, error)
}

func NewGateway(
	ticketSvc ticket.Service,
	agentEng *agent.Engine,
	feishuCli *feishu.Client,
	ticketGrp TicketGroupHandler,
	bot BotHandler,
	botCfgRepo BotConfigProvider,
	intentSvc *intent.Service,
	log *logger.Logger,
) *Gateway {
	return &Gateway{
		ticketSvc:  ticketSvc,
		agentEng:   agentEng,
		feishuCli:  feishuCli,
		ticketGrp:  ticketGrp,
		bot:        bot,
		botCfgRepo: botCfgRepo,
		intentSvc:  intentSvc,
		log:        log,
	}
}

func (g *Gateway) HandleIncoming(ctx context.Context, msg *IncomingMessage) (*model.TicketResponse, error) {
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	if err := g.enrichUser(ctx, msg); err != nil {
		g.log.Warn("Gateway: failed to enrich user", "error", err, "channel", msg.Channel, "user_id", msg.UserID)
	}

	// 1. Intent classification
	intentResult, _ := g.classifyIntent(ctx, msg.Content)

	switch intentResult.Intent {
	case intent.IntentGreeting:
		// 2a. Greeting → send welcome
		g.sendWelcome(ctx, msg)
		return nil, nil

	case intent.IntentCreateTicket:
		// 2b. Create ticket → build ticket with AI-extracted metadata
		g.log.Info("Gateway: intent=create_ticket, creating ticket", "channel", msg.Channel, "title", intentResult.Title)
		return g.createTicketFromIntent(ctx, msg, intentResult)

	default:
		// 2c. ask_knowledge or unknown → try bot first
		if g.bot != nil && msg.ChatID != "" {
			answer, err := g.bot.Ask(ctx, msg.Content)
			if err == nil && answer != nil && answer.Answer != "" && answer.Answer != "未找到相关知识" {
				g.log.Info("Gateway: bot answered, skipping ticket creation", "channel", msg.Channel, "chat_id", msg.ChatID)
				out := &OutboundMessage{
					Channel:  msg.Channel,
					TargetID: msg.ChatID,
					Content:  answer.Answer,
					MsgType:  "text",
				}
				if sendErr := g.SendToChannel(ctx, out); sendErr != nil {
					g.log.Error("Gateway: failed to send bot reply", "error", sendErr)
				}
				return nil, nil
			}
		}

		// Welcome if configured
		g.sendWelcome(ctx, msg)

		// Fallback: create ticket
		ticketReq := g.buildTicketRequest(msg)
		resp, err := g.ticketSvc.CreateTicket(ctx, ticketReq)
		if err != nil {
			return nil, fmt.Errorf("gateway: create ticket: %w", err)
		}
		go g.onTicketCreated(context.Background(), resp.ID, msg)
		return resp, nil
	}
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

func (g *Gateway) classifyIntent(ctx context.Context, message string) (*intent.IntentResult, error) {
	if g.intentSvc == nil {
		return &intent.IntentResult{Intent: intent.IntentUnknown}, nil
	}
	result, err := g.intentSvc.Classify(ctx, message)
	if err != nil {
		g.log.Debug("Gateway: intent classification failed", "error", err)
		return &intent.IntentResult{Intent: intent.IntentUnknown}, nil
	}
	g.log.Info("Gateway: intent classified", "intent", result.Intent, "title", result.Title)
	return result, nil
}

func (g *Gateway) sendWelcome(ctx context.Context, msg *IncomingMessage) {
	if g.botCfgRepo == nil || msg.ChatID == "" {
		return
	}
	botCfg, cfgErr := g.botCfgRepo.GetBotConfig(ctx, string(msg.Channel))
	if cfgErr == nil && botCfg != nil && botCfg.Enabled && botCfg.WelcomeMsg != "" {
		content := botCfg.WelcomeMsg
		if botCfg.Signature != "" {
			content += "\n\n" + botCfg.Signature
		}
		out := &OutboundMessage{
			Channel:  msg.Channel,
			TargetID: msg.ChatID,
			Content:  content,
			MsgType:  "text",
		}
		if sendErr := g.SendToChannel(ctx, out); sendErr != nil {
			g.log.Error("Gateway: failed to send welcome", "error", sendErr)
		}
	}
}

func (g *Gateway) createTicketFromIntent(ctx context.Context, msg *IncomingMessage, ir *intent.IntentResult) (*model.TicketResponse, error) {
	title := ir.Title
	if title == "" {
		title = truncateString(msg.Content, 100)
	}

	priority := ir.Priority
	if priority == "" || !model.IsValidPriority(priority) {
		priority = "medium"
	}

	ticketReq := &model.CreateTicketRequest{
		Title:       title,
		Description: msg.Content,
		Priority:    priority,
		RequesterID: msg.UserID,
		Channel:     string(msg.Channel),
		Metadata: map[string]interface{}{
			"channel":           string(msg.Channel),
			"channel_msg_id":    msg.MessageID,
			"channel_user_id":   msg.UserID,
			"raw_content":       msg.Content,
			"intent":            string(ir.Intent),
			"intent_category":   ir.Category,
			"intent_summary":    ir.Summary,
			"auto_created":      true,
		},
	}

	resp, err := g.ticketSvc.CreateTicket(ctx, ticketReq)
	if err != nil {
		return nil, fmt.Errorf("gateway: create ticket from intent: %w", err)
	}

	go g.onTicketCreated(context.Background(), resp.ID, msg)
	return resp, nil
}
