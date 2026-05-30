package llm

import (
	"context"
	"sync"
	"time"
)

// ChatService LLM 对话服务
type ChatService struct {
	provider   ChatProvider
	limiter    *TokenBucket
}

// ChatServiceOption 配置 ChatService
type ChatServiceOption func(*ChatService)

// WithRateLimit 设置 LLM 请求速率限制（rps: 每秒请求数, burst: 突发上限）
func WithRateLimit(rps int, burst int) ChatServiceOption {
	return func(s *ChatService) {
		if rps > 0 {
			s.limiter = NewTokenBucket(rps, burst)
		}
	}
}

// NewChatService 创建对话服务，若 provider 不支持 Chat 则返回 nil
func NewChatService(provider EmbeddingProvider, opts ...ChatServiceOption) *ChatService {
	chatProvider, ok := provider.(ChatProvider)
	if !ok {
		return nil
	}
	s := &ChatService{provider: chatProvider}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Chat 发送对话请求
func (s *ChatService) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if s.limiter != nil {
		if err := s.limiter.Wait(ctx); err != nil {
			return nil, err
		}
	}
	return s.provider.Chat(ctx, req)
}

// TokenBucket 简单的令牌桶限速器
type TokenBucket struct {
	mu       sync.Mutex
	tokens   int
	capacity int
	rate     time.Duration
	lastFill time.Time
}

// NewTokenBucket 创建令牌桶（rps: 每秒生成令牌数, burst: 最大令牌数）
func NewTokenBucket(rps int, burst int) *TokenBucket {
	if burst <= 0 {
		burst = rps
	}
	return &TokenBucket{
		tokens:   burst,
		capacity: burst,
		rate:     time.Second / time.Duration(rps),
		lastFill: time.Now(),
	}
}

// Wait 阻塞直到获取到令牌或上下文取消
func (tb *TokenBucket) Wait(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		tb.mu.Lock()
		tb.fill()
		if tb.tokens > 0 {
			tb.tokens--
			tb.mu.Unlock()
			return nil
		}
		tb.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(tb.rate):
		}
	}
}

func (tb *TokenBucket) fill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastFill)
	tokensToAdd := int(elapsed / tb.rate)
	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastFill = now
	}
}
