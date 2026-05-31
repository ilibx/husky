package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port     string `yaml:"port"`
	BasePath string `yaml:"base_path"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type DatabaseConfig struct {
	Provider        string `yaml:"provider"`
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"username"`
	Password        string `yaml:"password"`
	Name            string `yaml:"name"`
	SSLMode         string `yaml:"sslmode"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

type CacheConfig struct {
	Provider     string `yaml:"provider"`
	Enable       bool   `yaml:"enable"`
	Host         string `yaml:"host"`
	Port         string `yaml:"port"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret"`
	ExpireHour int    `yaml:"expire_hour"`
}

type LDAPConfig struct {
	Enable   bool   `yaml:"enable"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	BindDN   string `yaml:"bind_dn"`
	Password string `yaml:"password"`
	BaseDN   string `yaml:"base_dn"`
	Filter   string `yaml:"filter"`
	FieldMap string `yaml:"field_map"`
}

type CORSConfig struct {
	AllowedOrigins string `yaml:"allowed_origins"`
	AllowedMethods string `yaml:"allowed_methods"`
	AllowedHeaders string `yaml:"allowed_headers"`
}

type StorageConfig struct {
	Provider string            `yaml:"provider"`
	Options  map[string]string `yaml:"options"`
}

func DefaultStorageConfig() StorageConfig {
	return StorageConfig{
		Provider: "local",
		Options:  map[string]string{"root": "./data/knowledge"},
	}
}

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Log      LogConfig      `yaml:"log"`
	Database DatabaseConfig `yaml:"database"`
	Cache    CacheConfig    `yaml:"cache"`
	JWT      JWTConfig      `yaml:"jwt"`
	LDAP     LDAPConfig     `yaml:"ldap"`
	Storage  StorageConfig  `yaml:"storage"`
	CORS     CORSConfig     `yaml:"cors"`
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = "config.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file %s: %w", path, err)
	}

	cfg := &Config{
		Cache:   CacheConfig{Enable: true},
		LDAP:    LDAPConfig{Enable: true},
		Storage: DefaultStorageConfig(),
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file %s: %w", path, err)
	}

	setDefaults(cfg)

	if cfg.JWT.Secret == "husky-secret-key-change-in-production" {
		fmt.Println("WARNING: Using default JWT secret. Please change in production!")
	}

	return cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = "console"
	}
	if cfg.Database.Provider == "" {
		cfg.Database.Provider = "postgresql"
	}
	if cfg.Database.Host == "" {
		cfg.Database.Host = "localhost"
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 5432
	}
	if cfg.Database.User == "" {
		cfg.Database.User = "postgres"
	}
	if cfg.Database.Password == "" {
		cfg.Database.Password = "postgres"
	}
	if cfg.Database.Name == "" {
		cfg.Database.Name = "husky"
	}
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}
	if cfg.Database.MaxIdleConns <= 0 {
		cfg.Database.MaxIdleConns = 10
	}
	if cfg.Database.MaxOpenConns <= 0 {
		cfg.Database.MaxOpenConns = 100
	}
	if cfg.Database.ConnMaxLifetime <= 0 {
		cfg.Database.ConnMaxLifetime = 3600
	}
	if cfg.Cache.Provider == "" {
		cfg.Cache.Provider = "redis"
	}
	if cfg.Cache.Host == "" {
		cfg.Cache.Host = "localhost"
	}
	if cfg.Cache.Port == "" {
		cfg.Cache.Port = "6379"
	}
	if cfg.Cache.PoolSize <= 0 {
		cfg.Cache.PoolSize = 100
	}
	if cfg.Cache.MinIdleConns <= 0 {
		cfg.Cache.MinIdleConns = 5
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "husky-secret-key-change-in-production"
	}
	if cfg.JWT.ExpireHour <= 0 {
		cfg.JWT.ExpireHour = 24
	}
	if cfg.LDAP.Port <= 0 {
		cfg.LDAP.Port = 389
	}
	if cfg.LDAP.Filter == "" {
		cfg.LDAP.Filter = "(objectClass=person)"
	}
	if cfg.LDAP.FieldMap == "" {
		cfg.LDAP.FieldMap = `{"cn":"username","mail":"email"}`
	}
	if cfg.CORS.AllowedOrigins == "" {
		cfg.CORS.AllowedOrigins = "*"
	}
	if cfg.CORS.AllowedMethods == "" {
		cfg.CORS.AllowedMethods = "GET,POST,PUT,DELETE,OPTIONS,PATCH"
	}
	if cfg.CORS.AllowedHeaders == "" {
		cfg.CORS.AllowedHeaders = "Origin,Content-Type,Accept,Authorization,X-Requested-With"
	}
}

func LoadConfigFromEnv() (*Config, error) {
	cfg := &Config{
		Cache:  CacheConfig{Enable: true},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			Name:            getEnv("DB_NAME", "husky"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 3600),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "husky-secret-key-change-in-production"),
			ExpireHour: getEnvAsInt("JWT_EXPIRE_HOUR", 24),
		},
		LDAP: LDAPConfig{
			Enable:   true,
			Host:     getEnv("LDAP_HOST", ""),
			Port:     getEnvAsInt("LDAP_PORT", 389),
			BindDN:   getEnv("LDAP_BIND_DN", ""),
			Password: getEnv("LDAP_PASSWORD", ""),
			BaseDN:   getEnv("LDAP_BASE_DN", ""),
			Filter:   getEnv("LDAP_FILTER", "(objectClass=person)"),
			FieldMap: getEnv("LDAP_FIELD_MAP", `{"cn":"username","mail":"email"}`),
		},
	}

	setDefaults(cfg)

	if cfg.JWT.Secret == "husky-secret-key-change-in-production" {
		fmt.Println("WARNING: Using default JWT secret. Please change in production!")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

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
