package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "embed"

	"github.com/husky/husky/internal/config"
	"github.com/husky/husky/internal/database"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
	"github.com/husky/husky/internal/router"
	"github.com/husky/husky/pkg/logger"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

//go:embed default_menus.yaml
var defaultMenusData []byte

func main() {
	configPath := flag.String("config", "config.yaml", "Path to config YAML file")
	flag.Parse()

	// 从 YAML 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logInstance, err := logger.NewLogger(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logInstance.Info("Starting Husky API server...", "config_path", *configPath)

	// 初始化数据库连接
	dbCfg := database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.Name,
		SSLMode:         cfg.Database.SSLMode,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		ConnMaxLifetime: time.Duration(cfg.Database.ConnMaxLifetime) * time.Second,
	}
	db, err := database.NewDatabase(dbCfg)
	if err != nil {
		logInstance.Fatal("Failed to initialize database", "error", err)
	}
	logInstance.Info("Database connected successfully")

	// 自动迁移数据库 Schema
	if err := database.AutoMigrate(db); err != nil {
		logInstance.Fatal("Failed to migrate database", "error", err)
	}
	logInstance.Info("Database migration completed")

	// 初始化默认管理员账号
	seedDefaultAdmin(db, logInstance)
	seedDefaultMenus(db, logInstance)

	// 创建数据库连接包装
	dbConn := &repository.DatabaseConnection{
		DB: db,
	}

	// 设置路由
	r, cleanup := router.SetupRouter(cfg, dbConn, logInstance)
	defer cleanup()

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器在 goroutine 中
	go func() {
		logInstance.Info("Server starting", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logInstance.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logInstance.Info("Shutting down server...")

	// 优雅关闭，给予 5 秒超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logInstance.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logInstance.Info("Server exited")
}

const defaultAdminEmail = "admin@husky.local"
const defaultAdminPassword = "admin123"

func seedDefaultAdmin(db *gorm.DB, log *logger.Logger) {
	var count int64
	db.Model(&model.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(defaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Warn("Failed to hash default admin password", "error", err)
		return
	}

	user := &model.User{
		Email:    defaultAdminEmail,
		Username: "admin",
		Password: string(hash),
		Role:     "admin",
		Status:   1,
	}

	if err := db.Create(user).Error; err != nil {
		log.Warn("Failed to create default admin user", "error", err)
		return
	}

	log.Warn("Default admin account created",
		"email", defaultAdminEmail,
		"password", defaultAdminPassword,
		"role", "admin",
		"IMPORTANT", "Change this password immediately!")
	}

func seedDefaultMenus(db *gorm.DB, log *logger.Logger) {
	var count int64
	db.Model(&model.Menu{}).Count(&count)
	if count > 0 {
		return
	}

	type menuDef struct {
		Name     string    `yaml:"name"`
		Path     string    `yaml:"path"`
		Icon     string    `yaml:"icon"`
		Roles    string    `yaml:"roles"`
		Sort     int       `yaml:"sort"`
		Children []menuDef `yaml:"children"`
	}
	type menuFile struct {
		Menus []menuDef `yaml:"menus"`
	}

	data := defaultMenusData

	var mf menuFile
	if err := yaml.Unmarshal(data, &mf); err != nil {
		log.Warn("Failed to parse default menus YAML", "error", err)
		return
	}

	var create func(parentID *uint, defs []menuDef)
	create = func(parentID *uint, defs []menuDef) {
		for _, d := range defs {
			m := model.Menu{
				Name:     d.Name,
				Path:     d.Path,
				Icon:     d.Icon,
				Roles:    d.Roles,
				Sort:     d.Sort,
				ParentID: parentID,
			}
			if err := db.Create(&m).Error; err != nil {
				log.Warn("Failed to create menu", "name", d.Name, "error", err)
				continue
			}
			if len(d.Children) > 0 {
				pid := m.ID
				create(&pid, d.Children)
			}
		}
	}
	create(nil, mf.Menus)

	log.Info("Default menus seeded successfully")
}
