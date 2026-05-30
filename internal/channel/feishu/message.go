package feishu

import (
	"context"
	"fmt"
)

func (c *Client) SendMessage(ctx context.Context, receiveID, msgType, content string) error {
	token, err := c.getTenantToken(ctx)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"receive_id": receiveID,
		"msg_type":   msgType,
		"content":    content,
	}
	_, err = c.doPost(ctx, "/im/v1/messages?receive_id_type=open_id", token, body)
	if err != nil {
		return fmt.Errorf("feishu send message: %w", err)
	}
	return nil
}

func (c *Client) SendMessageToChat(ctx context.Context, chatID, msgType, content string) error {
	token, err := c.getTenantToken(ctx)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"receive_id": chatID,
		"msg_type":   msgType,
		"content":    content,
	}
	_, err = c.doPost(ctx, "/im/v1/messages?receive_id_type=chat_id", token, body)
	if err != nil {
		return fmt.Errorf("feishu send chat message: %w", err)
	}
	return nil
}

func (c *Client) SendCardMessage(ctx context.Context, chatID, cardJSON string) error {
	token, err := c.getTenantToken(ctx)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"receive_id": chatID,
		"msg_type":   "interactive",
		"content":    cardJSON,
	}
	_, err = c.doPost(ctx, "/im/v1/messages?receive_id_type=chat_id", token, body)
	if err != nil {
		return fmt.Errorf("feishu card message: %w", err)
	}
	return nil
}
