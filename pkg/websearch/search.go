package websearch

import (
	"context"
	"fmt"
)

// Result 网页搜索结果
type Result struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

// Searcher 网页搜索接口
type Searcher interface {
	Search(ctx context.Context, query string, limit int) ([]Result, error)
}

// Config 网页搜索配置
type Config struct {
	Provider string `json:"provider"` // "searxng" | "custom"
	Endpoint string `json:"endpoint"`
	APIKey   string `json:"api_key"`
}

// NewSearcher 根据配置创建 Searcher
func NewSearcher(cfg Config) (Searcher, error) {
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("websearch: endpoint is required")
	}
	switch cfg.Provider {
	case "searxng", "":
		return NewSearXNG(cfg.Endpoint, cfg.APIKey), nil
	default:
		return nil, fmt.Errorf("websearch: unsupported provider: %s", cfg.Provider)
	}
}
