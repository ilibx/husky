package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	appID         string
	appSecret     string
	baseURL       string
	httpClient    *http.Client
	mu            sync.RWMutex
	tenantToken   string
	tokenExpireAt time.Time
}

func NewClient(appID, appSecret string) *Client {
	return &Client{
		appID:      appID,
		appSecret:  appSecret,
		baseURL:    "https://open.feishu.cn/open-apis",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type tenantTokenResp struct {
	Code              int    `json:"code"`
	Msg               string `json:"msg"`
	TenantAccessToken string `json:"tenant_access_token"`
	Expire            int    `json:"expire"`
}

func (c *Client) getTenantToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	if c.tenantToken != "" && time.Now().Before(c.tokenExpireAt) {
		c.mu.RUnlock()
		return c.tenantToken, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.tenantToken != "" && time.Now().Before(c.tokenExpireAt) {
		return c.tenantToken, nil
	}

	body := map[string]string{
		"app_id":     c.appID,
		"app_secret": c.appSecret,
	}
	payload, _ := json.Marshal(body)

	resp, err := c.httpClient.Post(c.baseURL+"/auth/v3/tenant_access_token/internal", "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to get tenant token: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result tenantTokenResp
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}
	if result.Code != 0 {
		return "", fmt.Errorf("feishu API error [%d]: %s", result.Code, result.Msg)
	}

	c.tenantToken = result.TenantAccessToken
	c.tokenExpireAt = time.Now().Add(time.Duration(result.Expire-60) * time.Second)
	return c.tenantToken, nil
}

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
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/im/v1/messages?receive_id_type=open_id", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	json.Unmarshal(respBody, &result)
	if result.Code != 0 {
		return fmt.Errorf("feishu send message error [%d]: %s", result.Code, result.Msg)
	}
	return nil
}

func (c *Client) GetUserInfo(ctx context.Context, userID string) (map[string]interface{}, error) {
	token, err := c.getTenantToken(ctx)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/contact/v3/users/"+userID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result, nil
}

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
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/im/v1/chats", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to create chat: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			ChatID string `json:"chat_id"`
			Name   string `json:"name"`
		} `json:"data"`
	}
	json.Unmarshal(respBody, &result)
	if result.Code != 0 {
		return "", "", fmt.Errorf("feishu create chat error [%d]: %s", result.Code, result.Msg)
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
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/im/v1/chats/"+chatID+"/members", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add chat members: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	json.Unmarshal(respBody, &result)
	if result.Code != 0 {
		return fmt.Errorf("feishu add members error [%d]: %s", result.Code, result.Msg)
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
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/im/v1/messages?receive_id_type=chat_id", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send chat message: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	json.Unmarshal(respBody, &result)
	if result.Code != 0 {
		return fmt.Errorf("feishu send chat message error [%d]: %s", result.Code, result.Msg)
	}
	return nil
}

func (c *Client) GetChatMembers(ctx context.Context, chatID string) ([]string, error) {
	token, err := c.getTenantToken(ctx)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/im/v1/chats/"+chatID+"/members?page_size=50", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat members: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Items []struct {
				OpenID string `json:"member_id"`
			} `json:"items"`
		} `json:"data"`
	}
	json.Unmarshal(respBody, &result)
	if result.Code != 0 {
		return nil, fmt.Errorf("feishu get members error [%d]: %s", result.Code, result.Msg)
	}
	var members []string
	for _, item := range result.Data.Items {
		members = append(members, item.OpenID)
	}
	return members, nil
}

func (c *Client) SendCardMessage(ctx context.Context, chatID, cardJSON string) error {
	return c.SendMessageToChat(ctx, chatID, "interactive", cardJSON)
}

func VerifyWebhook(token, challenge string) map[string]string {
	return map[string]string{"challenge": challenge}
}
