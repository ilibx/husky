package gateway

import (
	"context"

	"github.com/husky/husky/internal/intent"
)

func (g *Gateway) tryBotAnswer(ctx context.Context, msg *IncomingMessage) bool {
	if g.bot == nil || msg.ChatID == "" {
		return false
	}
	answer, err := g.bot.Ask(ctx, msg.Content)
	if err != nil {
		g.log.Warn("Gateway: bot Ask failed, creating ticket", "error", err)
		return false
	}
	if answer == nil || answer.Answer == "" || answer.Answer == "未找到相关知识" {
		return false
	}
	out := &OutboundMessage{
		Channel:  msg.Channel,
		TargetID: msg.ChatID,
		Content:  answer.Answer,
		MsgType:  "text",
	}
	if sendErr := g.SendToChannel(ctx, out); sendErr != nil {
		g.log.Error("Gateway: failed to send bot reply", "error", sendErr)
	}
	return true
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

func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return s
}
