package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/husky/husky/internal/config"
	"github.com/husky/husky/internal/database"
	"github.com/husky/husky/internal/repository"
	"github.com/husky/husky/internal/router"
	"github.com/husky/husky/pkg/cache"
	"github.com/husky/husky/pkg/logger"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	config.Conf = cfg

	// 初始化日志
	logInstance, err := logger.NewLogger(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logInstance.Info("Starting Husky API server...", "config", cfg)

	// 初始化数据库连接
	dbCfg := database.Config{
		Host:            cfg.DBHost,
		Port:            cfg.DBPort,
		User:            cfg.DBUser,
		Password:        cfg.DBPassword,
		DBName:          cfg.DBName,
		SSLMode:         cfg.DBSSLMode,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		MaxOpenConns:    cfg.DBMaxOpenConns,
		ConnMaxLifetime: time.Duration(cfg.DBConnMaxLifetime) * time.Second,
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

	// 初始化缓存连接（可选，Redis 不可用时降级为无缓存模式）
	var cacheClient cache.CacheInterface
	redisClient, err := cache.NewRedis(cfg)
	if err != nil {
		logInstance.Warn("Redis unavailable, running without cache", "error", err)
		cacheClient = &cache.NoopCache{}
	} else {
		logInstance.Info("Redis connected successfully")
		cacheClient = cache.NewCache(redisClient)
	}

	// 创建数据库连接包装
	dbConn := &repository.DatabaseConnection{
		DB: db,
	}

	// 将 cacheClient 注入到上下文 (后续使用)
	_ = cacheClient

	// 设置路由
	r := router.SetupRouter(cfg, dbConn, logInstance)

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器在 goroutine 中
	go func() {
		logInstance.Info("Server starting", "port", cfg.ServerPort)
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
