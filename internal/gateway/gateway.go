package gateway

import (
	"context"
	"fmt"
	"time"

	"github.com/husky/husky/internal/channel/feishu"
	"github.com/husky/husky/internal/intent"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/ticket"
	"github.com/husky/husky/pkg/logger"
)

type Gateway struct {
	ticketSvc  ticket.TicketCreator
	feishuCli  *feishu.Client
	ticketGrp  TicketGroupHandler
	bot        BotHandler
	botCfgRepo BotConfigProvider
	intentSvc  *intent.Service
	enrichers  map[ChannelType]UserEnricher
	userLookup UserLookup
	log        *logger.Logger
}

type TicketGroupHandler interface {
	CreateGroupForTicket(ctx context.Context, ticket *model.Ticket) (*model.TicketGroup, error)
	OnGroupMessage(ctx context.Context, chatID, userID, content string) error
}

type BotConfigProvider interface {
	GetBotConfig(ctx context.Context, channel string) (*model.BotConfig, error)
}

// UserLookup resolves a channel user (by email or channel_id) to a local user ID.
type UserLookup interface {
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

func NewGateway(
	ticketSvc ticket.TicketCreator,
	feishuCli *feishu.Client,
	ticketGrp TicketGroupHandler,
	bot BotHandler,
	botCfgRepo BotConfigProvider,
	intentSvc *intent.Service,
	enrichers map[ChannelType]UserEnricher,
	userLookup UserLookup,
	log *logger.Logger,
) *Gateway {
	return &Gateway{
		ticketSvc:  ticketSvc,
		feishuCli:  feishuCli,
		ticketGrp:  ticketGrp,
		bot:        bot,
		botCfgRepo: botCfgRepo,
		intentSvc:  intentSvc,
		enrichers:  enrichers,
		userLookup: userLookup,
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
	intentResult, intentErr := g.classifyIntent(ctx, msg.Content)
	if intentErr != nil {
		g.log.Warn("Gateway: intent classification failed, falling back to default",
			"error", intentErr, "channel", msg.Channel)
	}

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
		if g.tryBotAnswer(ctx, msg) {
			return nil, nil
		}
		g.sendWelcome(ctx, msg)
		ticketReq := g.buildTicketRequest(ctx, msg)
		resp, err := g.ticketSvc.CreateTicket(ctx, ticketReq)
		if err != nil {
			return nil, fmt.Errorf("gateway: create ticket: %w", err)
		}
		return resp, nil
	}
}

func (g *Gateway) HandleGroupMessage(ctx context.Context, msg *IncomingMessage) error {
	if g.ticketGrp == nil {
		return nil
	}
	return g.ticketGrp.OnGroupMessage(ctx, msg.ChatID, msg.UserID, msg.Content)
}
