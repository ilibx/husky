package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type GeminiProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewGeminiProvider(apiKey, baseURL, model string) *GeminiProvider {
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}
	if model == "" {
		model = "gemini-2.0-flash"
	}
	return &GeminiProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiChatRequest struct {
	Contents         []geminiContent `json:"contents"`
	SystemInstruction *geminiContent  `json:"system_instruction,omitempty"`
	GenerationConfig struct {
		Temperature   float32 `json:"temperature,omitempty"`
		MaxOutputTokens int   `json:"maxOutputTokens,omitempty"`
	} `json:"generation_config,omitempty"`
}

type geminiChatResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type geminiEmbedRequest struct {
	Content geminiContent `json:"content"`
}

type geminiEmbedResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *GeminiProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	body := geminiEmbedRequest{
		Content: geminiContent{Parts: []geminiPart{{Text: text}}},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gemini embed request: %w", err)
	}

	u := fmt.Sprintf("%s/v1/models/%s:embedContent?key=%s", p.baseURL, p.model, url.QueryEscape(p.apiKey))
	req, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini embed request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read gemini embed response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini embed API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result geminiEmbedResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse gemini embed response: %w", err)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("gemini API error: %s", result.Error.Message)
	}

	return result.Embedding.Values, nil
}

func (p *GeminiProvider) BatchEmbed(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, text := range texts {
		vec, err := p.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("batch embed item %d: %w", i, err)
		}
		results[i] = vec
	}
	return results, nil
}

func (p *GeminiProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	contents := make([]geminiContent, 0, len(req.Messages))
	var systemContent *geminiContent

	for _, m := range req.Messages {
		role := m.Role
		if role == "system" {
			systemContent = &geminiContent{Parts: []geminiPart{{Text: m.Content}}}
			continue
		}
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: m.Content}},
		})
	}

	chatReq := geminiChatRequest{Contents: contents}
	if systemContent != nil {
		chatReq.SystemInstruction = systemContent
	}
	if req.Temperature > 0 {
		chatReq.GenerationConfig.Temperature = req.Temperature
	}
	if req.MaxTokens > 0 {
		chatReq.GenerationConfig.MaxOutputTokens = req.MaxTokens
	}

	model := req.Model
	if model == "" {
		model = p.model
	}

	payload, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gemini chat request: %w", err)
	}

	u := fmt.Sprintf("%s/v1/models/%s:generateContent?key=%s", p.baseURL, model, url.QueryEscape(p.apiKey))
	httpReq, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini chat request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read gemini response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result geminiChatResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse gemini response: %w", err)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("gemini API error: %s", result.Error.Message)
	}
	if len(result.Candidates) == 0 {
		return nil, fmt.Errorf("empty gemini response")
	}

	var text string
	for _, p := range result.Candidates[0].Content.Parts {
		text += p.Text
	}

	return &ChatResponse{
		Content: text,
		Model:   model,
	}, nil
}

func (p *GeminiProvider) Name() string {
	return ProviderGemini
}
