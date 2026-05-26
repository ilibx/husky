package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config 应用配置结构体
type Config struct {
	// 服务器配置
	ServerPort   string
	ServerMode   string // debug, release, test
	LogLevel     string
	LogFormat    string // json, console

	// 数据库配置
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Redis 配置
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	// JWT 配置
	JWTSecret     string
	JWTExpireHour int

	// 第三方服务配置
	LLMProvider      string
	LLMAPIKey        string
	LLMBaseURL       string
	VectorDBProvider string
	VectorDBHost     string
	VectorDBPort     string
}

// LoadConfig 从环境变量加载配置
func LoadConfig() (*Config, error) {
	cfg := &Config{
		// 服务器配置
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		ServerMode:   getEnv("SERVER_MODE", "debug"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		LogFormat:    getEnv("LOG_FORMAT", "json"),
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "postgres"),
		DBPassword:   getEnv("DB_PASSWORD", "postgres"),
		DBName:       getEnv("DB_NAME", "husky"),
		DBSSLMode:    getEnv("DB_SSLMODE", "disable"),
		RedisHost:    getEnv("REDIS_HOST", "localhost"),
		RedisPort:    getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:      getEnvAsInt("REDIS_DB", 0),
		JWTSecret:    getEnv("JWT_SECRET", "husky-secret-key-change-in-production"),
		JWTExpireHour: getEnvAsInt("JWT_EXPIRE_HOUR", 24),
		LLMProvider:   getEnv("LLM_PROVIDER", "openai"),
		LLMAPIKey:     getEnv("LLM_API_KEY", ""),
		LLMBaseURL:    getEnv("LLM_BASE_URL", "https://api.openai.com/v1"),
		VectorDBProvider: getEnv("VECTOR_DB_PROVIDER", "pgvector"),
		VectorDBHost:     getEnv("VECTOR_DB_HOST", "localhost"),
		VectorDBPort:     getEnv("VECTOR_DB_PORT", "5432"),
	}

	// 验证必要配置
	if cfg.JWTSecret == "husky-secret-key-change-in-production" {
		fmt.Println("WARNING: Using default JWT secret. Please change in production!")
	}

	return cfg, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}
