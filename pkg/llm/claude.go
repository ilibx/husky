package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ClaudeProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewClaudeProvider(apiKey, baseURL, _ string) *ClaudeProvider {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	return &ClaudeProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeRequest struct {
	Model         string          `json:"model"`
	MaxTokens     int             `json:"max_tokens"`
	Messages      []claudeMessage `json:"messages"`
	Temperature   float32         `json:"temperature,omitempty"`
	StopSequences []string        `json:"stop_sequences,omitempty"`
}

type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *ClaudeProvider) Embed(_ context.Context, _ string) ([]float32, error) {
	return nil, fmt.Errorf("Claude does not support embeddings")
}

func (p *ClaudeProvider) BatchEmbed(_ context.Context, _ []string) ([][]float32, error) {
	return nil, fmt.Errorf("Claude does not support embeddings")
}

func (p *ClaudeProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	msgs := make([]claudeMessage, len(req.Messages))
	for i, m := range req.Messages {
		role := m.Role
		if role == "system" {
			role = "user"
		}
		msgs[i] = claudeMessage{Role: role, Content: m.Content}
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	body := claudeRequest{
		Model:     req.Model,
		MaxTokens: maxTokens,
		Messages:  msgs,
	}
	if req.Temperature > 0 {
		body.Temperature = req.Temperature
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal claude request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("claude chat request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read claude response: %w", err)
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("claude API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result claudeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse claude response: %w", err)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("claude API error: %s", result.Error.Message)
	}
	if len(result.Content) == 0 {
		return nil, fmt.Errorf("empty claude response")
	}

	return &ChatResponse{
		Content: result.Content[0].Text,
		Model:   req.Model,
	}, nil
}

func (p *ClaudeProvider) Name() string {
	return ProviderClaude
}
