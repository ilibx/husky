package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIProvider OpenAI 兼容 API 的 Embedding 提供商
type OpenAIProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewOpenAIProvider 创建 OpenAI 提供商
func NewOpenAIProvider(apiKey, baseURL, model string) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasSuffix(baseURL, "/v1") {
		baseURL += "/v1"
	}
	if model == "" {
		model = "text-embedding-ada-002"
	}
	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

type openAIEmbedRequest struct {
	Model string `json:"model"`
	Input interface{} `json:"input"` // string or []string
}

type openAIEmbedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *OpenAIProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	results, err := p.BatchEmbed(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("empty embedding result")
	}
	return results[0], nil
}

func (p *OpenAIProvider) BatchEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts to embed")
	}

	body := openAIEmbedRequest{
		Model: p.model,
	}
	if len(texts) == 1 {
		body.Input = texts[0]
	} else {
		body.Input = texts
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/embeddings", bytes.NewReader(payload))
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

	var result openAIEmbedResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse embedding response: %w", err)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("embedding API error: %s", result.Error.Message)
	}

	embeddings := make([][]float32, 0, len(result.Data))
	for _, d := range result.Data {
		embeddings = append(embeddings, d.Embedding)
	}
	return embeddings, nil
}

func buildMessages(req *ChatRequest) []map[string]interface{} {
	messages := make([]map[string]interface{}, len(req.Messages))
	for i, msg := range req.Messages {
		m := map[string]interface{}{
			"role": msg.Role,
		}
		if len(msg.Images) > 0 {
			content := []map[string]interface{}{
				{"type": "text", "text": msg.Content},
			}
			for _, img := range msg.Images {
				content = append(content, map[string]interface{}{
					"type": "image_url",
					"image_url": map[string]string{
						"url": img,
					},
				})
			}
			m["content"] = content
		} else {
			m["content"] = msg.Content
		}
		messages[i] = m
	}
	return messages
}

func (p *OpenAIProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	messages := buildMessages(req)
	body := map[string]interface{}{
		"model":    req.Model,
		"messages": messages,
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	}
	if req.ReasoningEffort != "" {
		body["reasoning_effort"] = req.ReasoningEffort
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chat request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("chat request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read chat response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chat API returned %d (URL: %s): %s", resp.StatusCode, p.baseURL+"/chat/completions", string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse chat response (URL: %s, body: %s): %w", p.baseURL+"/chat/completions", string(respBody), err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("empty chat response")
	}

	return &ChatResponse{
		Content: result.Choices[0].Message.Content,
		Model:   req.Model,
	}, nil
}

func (p *OpenAIProvider) ChatStream(ctx context.Context, req *ChatRequest, callback ChatStreamCallback) (*ChatResponse, error) {
	messages := buildMessages(req)
	body := map[string]interface{}{
		"model":    req.Model,
		"messages": messages,
		"stream":   true,
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stream request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create stream request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("stream request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("stream API returned %d (URL: %s): %s", resp.StatusCode, p.baseURL+"/chat/completions", string(respBody))
	}

	fullContent := ""
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		fullContent += delta
		if err := callback(delta); err != nil {
			return nil, err
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("stream read error: %w", err)
	}

	return &ChatResponse{
		Content: fullContent,
		Model:   req.Model,
	}, nil
}

func (p *OpenAIProvider) Name() string {
	return ProviderOpenAI
}
