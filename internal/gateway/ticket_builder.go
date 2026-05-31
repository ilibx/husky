package gateway

import (
	"context"
	"fmt"
	"strconv"

	"github.com/husky/husky/internal/intent"
	"github.com/husky/husky/internal/model"
)

func (g *Gateway) enrichUser(ctx context.Context, msg *IncomingMessage) error {
	if msg.User == nil {
		msg.User = &UserContext{ChannelUserID: msg.UserID}
	}

	e, ok := g.enrichers[msg.Channel]
	if !ok {
		g.log.Debug("Gateway: no enricher for channel", "channel", msg.Channel)
		return nil
	}

	if err := e.Enrich(msg.User); err != nil {
		return fmt.Errorf("enrich user: %w", err)
	}
	return nil
}

func (g *Gateway) resolveRequesterID(ctx context.Context, msg *IncomingMessage) string {
	if msg.User != nil && msg.User.Email != "" && g.userLookup != nil {
		user, err := g.userLookup.GetByEmail(ctx, msg.User.Email)
		if err == nil && user != nil {
			return strconv.FormatUint(uint64(user.ID), 10)
		}
	}
	_, err := strconv.ParseUint(msg.UserID, 10, 64)
	if err == nil {
		return msg.UserID
	}
	return ""
}

func (g *Gateway) buildTicketRequest(ctx context.Context, msg *IncomingMessage) *model.CreateTicketRequest {
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
		RequesterID: g.resolveRequesterID(ctx, msg),
		Channel:     string(msg.Channel),
		Metadata:    metadata,
	}
	return req
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
		RequesterID: g.resolveRequesterID(ctx, msg),
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
	return resp, nil
}
