package database

import (
	"fmt"
	"log"
	"time"

	"github.com/husky/husky/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config 数据库配置
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

// NewDatabase 创建数据库连接
func NewDatabase(cfg Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying DB object: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established successfully")

	return db, nil
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate(db *gorm.DB) error {
	log.Println("Starting database migration...")

	err := db.AutoMigrate(
		&model.User{},
		&model.Ticket{},
		&model.Category{},
		&model.Comment{},
		&model.Attachment{},
		&model.KnowledgeBase{},
		&model.Satisfaction{},
		&model.SOP{},
		&model.Agent{},
		&model.Role{},
		&model.Department{},
		&model.Notification{},
		&model.AuditLog{},
		&model.WebhookMessageRecord{},
		&model.ChannelConfig{},
		&model.TicketWatcher{},
		&model.TicketGroup{},
		&model.Tag{},
		&model.AssignConfig{},
		&model.TicketRelation{},

		&model.BotConfig{},
		&model.SystemConfig{},
		&model.Menu{},
		&model.Workflow{},
		&model.WorkflowStep{},
		&model.SLAConfig{},
		&model.WebhookConfig{},
		&model.Skill{},
		&model.MCP{},
		&model.ChannelUser{},
		&model.ChannelGroup{},
	)

	if err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("Database migration completed successfully")

	if err := SeedBuiltinTools(db); err != nil {
		return fmt.Errorf("failed to seed built-in tools: %w", err)
	}

	return nil
}

// SeedBuiltinTools 初始化系统内置工具
func SeedBuiltinTools(db *gorm.DB) error {
	builtins := []model.MCP{
		{
			Name:        "网页搜索",
			Type:        "builtin",
			Key:         "web_search",
			Description: "通过 SearXNG 搜索引擎检索互联网信息",
			Enabled:     true,
			CreatedBy:   1,
		},
		{
			Name:        "Miniflux",
			Type:        "builtin",
			Key:         "miniflux",
			Description: "Miniflux RSS 阅读器集成，订阅和获取文章更新",
			Enabled:     true,
			CreatedBy:   1,
		},
	}

	for _, t := range builtins {
		var existing model.MCP
		result := db.Where("key = ?", t.Key).First(&existing)
		if result.Error != nil {
			if err := db.Create(&t).Error; err != nil {
				return fmt.Errorf("failed to create builtin tool %s: %w", t.Key, err)
			}
			log.Printf("Builtin tool created: %s (%s)", t.Name, t.Key)
		}
	}
	return nil
}

// Close 关闭数据库连接
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
