package feishu

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) CreateChat(ctx context.Context, name, description string) (string, string, error) {
	token, err := c.getTenantToken(ctx)
	if err != nil {
		return "", "", err
	}

	body := map[string]interface{}{
		"name":        name,
		"description": description,
		"chat_type":   "group",
	}
	respBody, err := c.doPost(ctx, "/im/v1/chats", token, body)
	if err != nil {
		return "", "", fmt.Errorf("feishu create chat: %w", err)
	}

	var result struct {
		Data struct {
			ChatID string `json:"chat_id"`
			Name   string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", fmt.Errorf("feishu parse create chat response: %w", err)
	}
	return result.Data.ChatID, result.Data.Name, nil
}

func (c *Client) AddChatMembers(ctx context.Context, chatID string, openIDs []string) error {
	token, err := c.getTenantToken(ctx)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"id_list": openIDs,
	}
	_, err = c.doPost(ctx, "/im/v1/chats/"+chatID+"/members", token, body)
	if err != nil {
		return fmt.Errorf("feishu add members: %w", err)
	}
	return nil
}

func (c *Client) GetChatMembers(ctx context.Context, chatID string) ([]string, error) {
	token, err := c.getTenantToken(ctx)
	if err != nil {
		return nil, err
	}

	respBody, err := c.doGet(ctx, "/im/v1/chats/"+chatID+"/members?page_size=50", token)
	if err != nil {
		return nil, fmt.Errorf("feishu get members: %w", err)
	}

	var result struct {
		Data struct {
			Items []struct {
				OpenID string `json:"member_id"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("feishu parse get members response: %w", err)
	}
	var members []string
	for _, item := range result.Data.Items {
		members = append(members, item.OpenID)
	}
	return members, nil
}
