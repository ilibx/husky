package llm

import "context"

// ChatService LLM 对话服务
type ChatService struct {
	provider ChatProvider
}

// NewChatService 创建对话服务，若 provider 不支持 Chat 则返回 nil
func NewChatService(provider EmbeddingProvider) *ChatService {
	chatProvider, ok := provider.(ChatProvider)
	if !ok {
		return nil
	}
	return &ChatService{provider: chatProvider}
}

// Chat 发送对话请求
func (s *ChatService) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	return s.provider.Chat(ctx, req)
}
