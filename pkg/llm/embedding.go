package llm

import (
	"context"
	"fmt"
)

// EmbeddingService Embedding 服务
type EmbeddingService struct {
	provider EmbeddingProvider
	dimension int
}

// NewEmbeddingService 创建 Embedding 服务
func NewEmbeddingService(provider EmbeddingProvider) *EmbeddingService {
	dim := 1536 // text-embedding-ada-002 default
	if provider.Name() == ProviderDashScope {
		dim = 1536 // text-embedding-v2 default
	}
	return &EmbeddingService{
		provider:  provider,
		dimension: dim,
	}
}

// Embed 生成单段文本的向量
func (s *EmbeddingService) Embed(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return make([]float32, s.dimension), nil
	}
	vec, err := s.provider.Embed(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}
	if len(vec) == 0 {
		return make([]float32, s.dimension), nil
	}
	return vec, nil
}

// Dimension 返回向量维度
func (s *EmbeddingService) Dimension() int {
	return s.dimension
}

// NewProviderFromConfig 根据配置创建 Embedding 提供商
func NewProviderFromConfig(providerName, apiKey, baseURL string) (EmbeddingProvider, error) {
	switch providerName {
	case ProviderOpenAI:
		return NewOpenAIProvider(apiKey, baseURL, ""), nil
	case ProviderDashScope:
		return NewDashScopeProvider(apiKey, baseURL, ""), nil
	case ProviderClaude:
		return NewClaudeProvider(apiKey, baseURL, ""), nil
	case ProviderGemini:
		return NewGeminiProvider(apiKey, baseURL, ""), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", providerName)
	}
}
