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
	DBHost            string
	DBPort            int
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	DBMaxIdleConns    int
	DBMaxOpenConns    int
	DBConnMaxLifetime int

	// Redis 配置
	RedisHost         string
	RedisPort         string
	RedisPassword     string
	RedisDB           int
	RedisPoolSize     int
	RedisMinIdleConns int

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

	// 飞书配置
	FeishuAppID     string
	FeishuAppSecret string

	// 管理后台配置
	AdminMode string // embedded, external; default embedded
	AdminURL string  // external URL for admin UI, e.g. http://localhost:5173

	// LDAP 配置
	LDAPHost     string
	LDAPPort     int
	LDAPBindDN   string
	LDAPPassword string
	LDAPBaseDN   string
	LDAPFilter   string
	LDAPFieldMap string // JSON: {"cn":"username","mail":"email","telephoneNumber":"phone"}
}

// LoadConfig 从环境变量加载配置
func LoadConfig() (*Config, error) {
	cfg := &Config{
		// 服务器配置
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		ServerMode:   getEnv("SERVER_MODE", "debug"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		LogFormat:    getEnv("LOG_FORMAT", "json"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnvAsInt("DB_PORT", 5432),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", "postgres"),
		DBName:            getEnv("DB_NAME", "husky"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		DBMaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
		DBMaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
		DBConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 3600),
		RedisHost:         getEnv("REDIS_HOST", "localhost"),
		RedisPort:         getEnv("REDIS_PORT", "6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		RedisDB:           getEnvAsInt("REDIS_DB", 0),
		RedisPoolSize:     getEnvAsInt("REDIS_POOL_SIZE", 100),
		RedisMinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5),
		JWTSecret:    getEnv("JWT_SECRET", "husky-secret-key-change-in-production"),
		JWTExpireHour: getEnvAsInt("JWT_EXPIRE_HOUR", 24),
		LLMProvider:   getEnv("LLM_PROVIDER", "openai"),
		LLMAPIKey:     getEnv("LLM_API_KEY", ""),
		LLMBaseURL:    getEnv("LLM_BASE_URL", "https://api.openai.com/v1"),
		VectorDBProvider: getEnv("VECTOR_DB_PROVIDER", "pgvector"),
		VectorDBHost:     getEnv("VECTOR_DB_HOST", "localhost"),
		VectorDBPort:     getEnv("VECTOR_DB_PORT", "5432"),
		FeishuAppID:     getEnv("FEISHU_APP_ID", ""),
		FeishuAppSecret: getEnv("FEISHU_APP_SECRET", ""),
		AdminMode:       getEnv("ADMIN_MODE", "embedded"),
		AdminURL:        getEnv("ADMIN_URL", "http://localhost:5173"),
		LDAPHost:        getEnv("LDAP_HOST", ""),
		LDAPPort:        getEnvAsInt("LDAP_PORT", 389),
		LDAPBindDN:      getEnv("LDAP_BIND_DN", ""),
		LDAPPassword:    getEnv("LDAP_PASSWORD", ""),
		LDAPBaseDN:      getEnv("LDAP_BASE_DN", ""),
		LDAPFilter:      getEnv("LDAP_FILTER", "(objectClass=person)"),
		LDAPFieldMap:    getEnv("LDAP_FIELD_MAP", `{"cn":"username","mail":"email"}`),
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
