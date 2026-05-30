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

type apiError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *apiError) Error() string {
	return fmt.Sprintf("feishu API error [%d]: %s", e.Code, e.Msg)
}

func (c *Client) doPost(ctx context.Context, path, token string, body interface{}) ([]byte, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("feishu marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("feishu create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("feishu request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("feishu read response: %w", err)
	}

	var apiErr apiError
	if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Code != 0 {
		return nil, &apiErr
	}
	return respBody, nil
}

func (c *Client) doGet(ctx context.Context, path, token string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("feishu create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("feishu request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("feishu read response: %w", err)
	}

	var apiErr apiError
	if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Code != 0 {
		return nil, &apiErr
	}
	return respBody, nil
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
	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("feishu marshal token request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/auth/v3/tenant_access_token/internal", "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to get tenant token: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}
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
