package config

import (
	"context"
	"fmt"
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

	if dbCfg.APIKey == "" {
		return fmt.Errorf("LLM API key not configured in system_configs, please set via admin UI")
	}

	providerName := dbCfg.Provider
	if providerName == "" {
		providerName = "openai"
	}
	baseURL := dbCfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
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
