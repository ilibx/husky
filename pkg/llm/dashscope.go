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

// DashScopeProvider 阿里云通义千问 Embedding 提供商
type DashScopeProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewDashScopeProvider 创建 DashScope 提供商
func NewDashScopeProvider(apiKey, baseURL, model string) *DashScopeProvider {
	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com/api/v1"
	}
	if model == "" {
		model = "text-embedding-v2"
	}
	return &DashScopeProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type dashScopeEmbedRequest struct {
	Model string            `json:"model"`
	Input dashScopeInput    `json:"input"`
	Params map[string]interface{} `json:"params,omitempty"`
}

type dashScopeInput struct {
	Texts []string `json:"texts"`
}

type dashScopeEmbedResponse struct {
	Output struct {
		Embeddings []struct {
			TextIndex int       `json:"text_index"`
			Embedding []float32 `json:"embedding"`
		} `json:"embeddings"`
	} `json:"output"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (p *DashScopeProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	results, err := p.BatchEmbed(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("empty embedding result")
	}
	return results[0], nil
}

func (p *DashScopeProvider) BatchEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts to embed")
	}

	body := dashScopeEmbedRequest{
		Model: p.model,
		Input: dashScopeInput{Texts: texts},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/services/embeddings/text-embedding/text-embedding", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedding response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result dashScopeEmbedResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}
	if result.Code != "" {
		return nil, fmt.Errorf("dashscope API error [%s]: %s", result.Code, result.Message)
	}

	embeddings := make([][]float32, len(result.Output.Embeddings))
	for _, e := range result.Output.Embeddings {
		embeddings[e.TextIndex] = e.Embedding
	}
	return embeddings, nil
}

func (p *DashScopeProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	return nil, fmt.Errorf("DashScope does not support chat")
}

func (p *DashScopeProvider) ChatStream(ctx context.Context, req *ChatRequest, callback ChatStreamCallback) (*ChatResponse, error) {
	return nil, fmt.Errorf("DashScope does not support chat")
}

func (p *DashScopeProvider) Name() string {
	return ProviderDashScope
}
