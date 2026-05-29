package channel

import (
	"encoding/json"
	"strings"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
)

func ParseWebhookMessage(channelType model.ChannelType, rawJSON []byte) (*model.WebhookMessage, error) {
	msg := &model.WebhookMessage{
		Channel: channelType,
	}

	switch channelType {
	case model.ChannelLark:
		var event model.LarkWebhookEvent
		if err := unmarshal(rawJSON, &event); err != nil {
			return nil, err
		}
		if event.Event != nil {
			msg.UserID = event.Event.UserID
			msg.MessageID = event.Event.MessageID
			content := event.Event.Content
			if content == "" && event.Event.Text.Content != "" {
				content = event.Event.Text.Content
			}
			msg.Content = cleanLarkContent(content)
		}
		msg.RawPayload = event

	case model.ChannelDingTalk:
		var event model.DingTalkWebhookEvent
		if err := unmarshal(rawJSON, &event); err != nil {
			return nil, err
		}
		msg.UserID = event.SenderID
		msg.MessageID = event.ConversationID
		if event.Text != nil {
			msg.Content = strings.TrimSpace(event.Text.Content)
		}
		msg.RawPayload = event

	case model.ChannelWeCom:
		var event model.WeComWebhookEvent
		if err := unmarshal(rawJSON, &event); err != nil {
			return nil, err
		}
		msg.UserID = event.FromUserName
		msg.MessageID = event.MsgID
		if event.Text != nil {
			msg.Content = strings.TrimSpace(event.Text.Content)
		}
		msg.RawPayload = event

	default:
		return nil, errors.New(errors.ErrInvalidParams, "unsupported channel type")
	}

	return msg, nil
}

func cleanLarkContent(content string) string {
	result := content
	result = strings.ReplaceAll(result, "<at>", "")
	result = strings.ReplaceAll(result, "</at>", "")
	result = strings.ReplaceAll(result, "<a>", "")
	result = strings.ReplaceAll(result, "</a>", "")
	result = strings.ReplaceAll(result, "<b>", "")
	result = strings.ReplaceAll(result, "</b>", "")
	result = strings.ReplaceAll(result, "<i>", "")
	result = strings.ReplaceAll(result, "</i>", "")
	result = strings.ReplaceAll(result, "<br/>", "\n")
	result = strings.ReplaceAll(result, "<br>", "\n")
	return strings.TrimSpace(result)
}

func unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
