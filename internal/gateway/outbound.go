package gateway

import (
	"context"
	"fmt"
)



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
	g.log.Warn("Gateway: dingtalk outgoing not implemented yet, message dropped",
		"target", out.TargetID, "content_length", len(out.Content))
	return fmt.Errorf("gateway: dingtalk outgoing not implemented")
}

func (g *Gateway) sendWeCom(ctx context.Context, out *OutboundMessage) error {
	g.log.Warn("Gateway: wecom outgoing not implemented yet, message dropped",
		"target", out.TargetID, "content_length", len(out.Content))
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
