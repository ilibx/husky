package llm

import "context"

// EmbeddingProvider 文本向量化服务接口
type EmbeddingProvider interface {
	// Embed 将文本转换为向量
	Embed(ctx context.Context, text string) ([]float32, error)
	// BatchEmbed 批量文本向量化
	BatchEmbed(ctx context.Context, texts []string) ([][]float32, error)
	// Name 返回提供商名称
	Name() string
}

// ChatProvider LLM 对话接口
type ChatProvider interface {
	// Chat 发送对话请求
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}

// ChatRequest 对话请求
type ChatRequest struct {
	Model           string        `json:"model"`
	Messages        []ChatMessage `json:"messages"`
	Temperature     float32       `json:"temperature,omitempty"`
	MaxTokens       int           `json:"max_tokens,omitempty"`
	ReasoningEffort string        `json:"reasoning_effort,omitempty"`
}

// ChatMessage 对话消息
type ChatMessage struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"-"` // base64 encoded images (not serialized directly)
}

// ChatResponse 对话响应
type ChatResponse struct {
	Content string `json:"content"`
	Model   string `json:"model"`
}

const (
	ProviderOpenAI    = "openai"
	ProviderDashScope = "dashscope"
	ProviderClaude    = "claude"
	ProviderGemini    = "gemini"
)
