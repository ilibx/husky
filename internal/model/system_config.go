package model

import "time"

// SystemConfig 系统动态配置，存储 LLM、渠道凭据等可在管理页面运行时修改的参数
type SystemConfig struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Category  string    `gorm:"size:50;not null;uniqueIndex:idx_category_key" json:"category"` // llm, channel, notification, general
	Key       string    `gorm:"size:100;not null;uniqueIndex:idx_category_key" json:"key"`
	Value     string    `gorm:"type:text;not null" json:"value"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (SystemConfig) TableName() string { return "system_configs" }

// SystemConfigCategory 配置分类常量
const (
	SysCfgCategoryLLM          = "llm"
	SysCfgCategoryVector       = "vector"
	SysCfgCategoryChannel      = "channel"
	SysCfgCategoryNotification = "notification"
	SysCfgCategoryGeneral      = "general"
)

// SysCfg keys  LLM
const (
	SysCfgLLMProvider    = "llm_provider"
	SysCfgLLMAPIKey      = "llm_api_key"
	SysCfgLLMBaseURL     = "llm_base_url"
	SysCfgLLMModel       = "llm_model"
	SysCfgLLMMaxTokens   = "llm_max_tokens"
	SysCfgLLMTemperature = "llm_temperature"
)

// SysCfg keys  Vector
const (
	SysCfgVectorProvider = "vector_provider"
	SysCfgVectorHost     = "vector_host"
	SysCfgVectorPort     = "vector_port"
	SysCfgVectorUser     = "vector_user"
	SysCfgVectorPassword = "vector_password"
	SysCfgVectorDatabase = "vector_database"
	SysCfgVectorSSLMode  = "vector_sslmode"
	SysCfgVectorSources  = "sources"
)

type KnowledgeStoreConfig struct {
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Provider    string `json:"provider"`
	Host        string `json:"host"`
	Port        string `json:"port"`
	User        string `json:"user"`
	Password    string `json:"password"`
	Database    string `json:"database"`
	SSLMode     string `json:"sslmode"`
	Enabled     bool   `json:"enabled"`
	IsDefault   bool   `json:"is_default"`
	Description string `json:"description"`
}

// SystemConfigRequest 创建/更新系统配置请求
type SystemConfigRequest struct {
	Category string `json:"category" binding:"required"`
	Key      string `json:"key" binding:"required"`
	Value    string `json:"value" binding:"required"`
	Enabled  *bool  `json:"enabled,omitempty"`
}

// SystemConfigResponse 系统配置响应
type SystemConfigResponse struct {
	ID        uint      `json:"id"`
	Category  string    `json:"category"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LLMConfig 从 SystemConfig 聚合得到的 LLM 配置快照
type LLMConfig struct {
	Provider    string  `json:"provider"`
	APIKey      string  `json:"api_key"`
	BaseURL     string  `json:"base_url"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// DefaultLLMConfig 默认 LLM 配置（与 env 默认值对齐）
func DefaultLLMConfig() LLMConfig {
	return LLMConfig{
		Provider:    "openai",
		APIKey:      "",
		BaseURL:     "https://api.openai.com/v1",
		Model:       "gpt-4o",
		MaxTokens:   4096,
		Temperature: 0.7,
	}
}

// VectorDBConfig 从 SystemConfig 聚合得到的向量库配置快照
type VectorDBConfig struct {
	Provider string `json:"provider"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"sslmode"`
}

func DefaultVectorDBConfig() VectorDBConfig {
	return VectorDBConfig{
		Provider: "postgresql",
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "",
		Database: "husky",
		SSLMode:  "disable",
	}
}
