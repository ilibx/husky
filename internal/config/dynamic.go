package config

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
	"github.com/husky/husky/pkg/llm"
	"github.com/husky/husky/pkg/vectorstore"
	"gorm.io/gorm"
)

type DynamicConfig struct {
	mu          sync.RWMutex
	cfgRepo     *repository.SystemConfigRepository
	chatSvc     *llm.ChatService
	embedSvc    *llm.EmbeddingService
	provider    llm.EmbeddingProvider
	initialized atomic.Bool
	vectorCfg   model.VectorDBConfig
	vectorStore vectorstore.Store
	gormDB      *gorm.DB
}

func NewDynamicConfig(cfgRepo *repository.SystemConfigRepository) *DynamicConfig {
	return &DynamicConfig{
		cfgRepo: cfgRepo,
	}
}

func (d *DynamicConfig) InitLLM(ctx context.Context) error {
	dbCfg, err := d.cfgRepo.GetLLMConfig(ctx)
	if err != nil {
		return fmt.Errorf("get LLM config from DB: %w", err)
	}

	// Try fallback to the first enabled platform preset if no individual LLM config is set
	if dbCfg.APIKey == "" {
		presetCfg, presetErr := d.loadFirstPreset(ctx)
		if presetErr == nil && presetCfg != nil {
			dbCfg = presetCfg
		}
	}

	if dbCfg.APIKey == "" {
		return fmt.Errorf("LLM API key not configured in system_configs, please set via admin UI")
	}

	providerName := dbCfg.Provider
	if providerName == "" {
		providerName = "openai"
	}
	baseURL := dbCfg.BaseURL
	if baseURL == "" {
		switch providerName {
		case "claude":
			baseURL = "https://api.anthropic.com"
		case "gemini":
			baseURL = "https://generativelanguage.googleapis.com"
		case "dashscope":
			baseURL = "https://dashscope.aliyuncs.com/api/v1"
		default:
			baseURL = "https://api.openai.com/v1"
		}
	}

	provider, err := llm.NewProviderFromConfig(providerName, dbCfg.APIKey, baseURL)
	if err != nil {
		return fmt.Errorf("init LLM provider: %w", err)
	}

	d.mu.Lock()
	d.provider = provider
	d.chatSvc = llm.NewChatService(provider, llm.WithRateLimit(10, 20), llm.WithDefaultModel(dbCfg.Model))
	d.embedSvc = llm.NewEmbeddingService(provider)
	d.mu.Unlock()

	d.initialized.Store(true)
	return nil
}

func (d *DynamicConfig) GetChatService() *llm.ChatService {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.chatSvc
}

func (d *DynamicConfig) GetEmbeddingService() *llm.EmbeddingService {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.embedSvc
}

func (d *DynamicConfig) IsReady() bool {
	return d.initialized.Load()
}

func (d *DynamicConfig) ReloadLLM(ctx context.Context) error {
	return d.InitLLM(ctx)
}

func (d *DynamicConfig) InitVector(ctx context.Context, gormDB *gorm.DB) error {
	dbCfg, err := d.cfgRepo.GetVectorConfig(ctx)
	if err != nil {
		return fmt.Errorf("get vector config from DB: %w", err)
	}

	store, err := vectorstore.NewFromConfig(vectorstore.Config{
		Provider: dbCfg.Provider,
		Host:     dbCfg.Host,
		Port:     dbCfg.Port,
		User:     dbCfg.User,
		Password: dbCfg.Password,
		Database: dbCfg.Database,
		SSLMode:  dbCfg.SSLMode,
	}, gormDB)
	if err != nil {
		return fmt.Errorf("init vector store: %w", err)
	}

	d.mu.Lock()
	d.vectorCfg = *dbCfg
	d.vectorStore = store
	d.gormDB = gormDB
	d.mu.Unlock()

	return nil
}

func (d *DynamicConfig) GetVectorConfig() model.VectorDBConfig {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.vectorCfg
}

func (d *DynamicConfig) GetVectorStore() vectorstore.Store {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.vectorStore
}

func (d *DynamicConfig) ReloadVector(ctx context.Context) error {
	d.mu.RLock()
	gormDB := d.gormDB
	d.mu.RUnlock()
	return d.InitVector(ctx, gormDB)
}

// loadFirstPreset 尝试从首个启用的平台 preset 中读取 LLM 配置作为 fallback
func (d *DynamicConfig) loadFirstPreset(ctx context.Context) (*model.LLMConfig, error) {
	configs, err := d.cfgRepo.ListByCategory(ctx, model.SysCfgCategoryLLM)
	if err != nil {
		return nil, err
	}
	for _, c := range configs {
		if !strings.HasPrefix(c.Key, "preset:") || !c.Enabled {
			continue
		}
		var preset struct {
			Name    string   `json:"name"`
			Type    string   `json:"type"`
			APIKey  string   `json:"api_key"`
			BaseURL string   `json:"base_url"`
			Enabled bool     `json:"enabled"`
			Models  []string `json:"models"`
		}
		if err := json.Unmarshal([]byte(c.Value), &preset); err != nil {
			continue
		}
		if preset.APIKey == "" {
			continue
		}
		cfg := model.DefaultLLMConfig()
		cfg.Provider = preset.Type
		cfg.APIKey = preset.APIKey
		if preset.BaseURL != "" {
			cfg.BaseURL = preset.BaseURL
		}
		if len(preset.Models) > 0 {
			cfg.Model = preset.Models[0]
		}
		return &cfg, nil
	}
	return nil, fmt.Errorf("no enabled platform preset with api_key found")
}
