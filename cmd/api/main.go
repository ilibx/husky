package main

import (
	"context"
	"encoding/json"
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

	// 初始化默认数据
	seedDefaultRoles(db, logInstance)
	seedDefaultAdmin(db, logInstance)
	seedDefaultMenus(db, logInstance)
	seedDefaultAgentAssets(db, logInstance)

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

func seedDefaultRoles(db *gorm.DB, log *logger.Logger) {
	permissions, err := json.Marshal(model.DefaultAdminPermissions())
	if err != nil {
		log.Warn("Failed to marshal default admin permissions", "error", err)
		return
	}

	role := model.Role{
		Name:        "admin",
		Description: "系统管理员",
		Permissions: string(permissions),
		Status:      1,
	}

	if err := db.Where("name = ?", role.Name).FirstOrCreate(&role).Error; err != nil {
		log.Warn("Failed to create default admin role", "error", err)
	}
}

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

func seedDefaultAgentAssets(db *gorm.DB, log *logger.Logger) {
	adminID := defaultCreatorID(db)
	seedDefaultCustomerServiceAgent(db, log, adminID)
	seedDefaultCustomerServiceSkill(db, log, adminID)
	seedDefaultCustomerServiceSOP(db, log, adminID)
}

func defaultCreatorID(db *gorm.DB) uint {
	var user model.User
	if err := db.Where("email = ?", defaultAdminEmail).First(&user).Error; err == nil {
		return user.ID
	}
	return 1
}

func seedDefaultCustomerServiceAgent(db *gorm.DB, log *logger.Logger, createdBy uint) {
	agent := model.Agent{
		Name:        "默认客服 Agent",
		Description: "面向客户咨询、故障排查、配置核对和工单回复的默认客服 Agent。",
		Type:        "llm",
		Config:      `{"trigger_on":"ticket_created","role":"customer_service","language":"zh-CN","principles":["先确认问题影响范围和紧急程度","基于知识库、SOP和工单上下文回复","涉及生产数据、权限、重启、部署、删除、回滚时必须人工确认","不暴露密钥、令牌、内部账号和敏感信息"]}`,
		Model:       "",
		Temperature: 0.3,
		MaxTokens:   2048,
		Enabled:     true,
		CreatedBy:   createdBy,
	}

	if err := db.Where("name = ?", agent.Name).FirstOrCreate(&agent).Error; err != nil {
		log.Warn("Failed to seed default customer service agent", "error", err)
	}
}

func seedDefaultCustomerServiceSkill(db *gorm.DB, log *logger.Logger, createdBy uint) {
	skill := model.Skill{
		Name:        "客服基础技能",
		Description: "用于客户咨询、工单回复、故障排查、配置核对、需求反馈和问题升级的基础客服技能。",
		Category:    "客服",
		Enabled:     true,
		CreatedBy:   createdBy,
	}

	if err := db.Where("name = ?", skill.Name).FirstOrCreate(&skill).Error; err != nil {
		log.Warn("Failed to seed default customer service skill", "error", err)
	}
}

func seedDefaultCustomerServiceSOP(db *gorm.DB, log *logger.Logger, createdBy uint) {
	var agent model.Agent
	var agentID *uint
	if err := db.Where("name = ?", "默认客服 Agent").First(&agent).Error; err == nil {
		agentID = &agent.ID
	}

	type stepDef struct {
		Name    string                 `json:"name"`
		Type    string                 `json:"type"`
		AgentID *uint                  `json:"agent_id,omitempty"`
		Config  map[string]interface{} `json:"config,omitempty"`
	}

	steps, err := json.Marshal([]stepDef{
		{
			Name:    "识别问题并收集信息",
			Type:    "react",
			AgentID: agentID,
			Config: map[string]interface{}{
				"goal": "判断工单类型、紧急程度和影响范围，必要时向用户补充询问环境、时间、复现步骤、报错信息、账号或租户标识。",
			},
		},
		{
			Name:    "检索知识库并给出初步回复",
			Type:    "react",
			AgentID: agentID,
			Config: map[string]interface{}{
				"goal": "基于知识库和工单上下文检索相关资料，先给结论，再给可执行排查步骤或使用指导。",
			},
		},
		{
			Name: "人工确认高风险操作",
			Type: "human",
			Config: map[string]interface{}{
				"goal": "涉及生产数据、权限、重启、部署、删除、回滚、批量更新或安全处置时，由人工客服确认后再继续。",
			},
		},
	})
	if err != nil {
		log.Warn("Failed to marshal default customer service SOP steps", "error", err)
		return
	}

	sop := model.SOP{
		Name:          "默认客服处理 SOP",
		Description:   "客户咨询、故障排查、配置核对、需求反馈和工单回复的默认处理流程。",
		Version:       "1.0",
		Steps:         string(steps),
		TriggerType:   "ticket_created",
		TriggerConfig: `{}`,
		RiskLevel:     "medium",
		Status:        1,
		CreatedBy:     createdBy,
	}

	if err := db.Where("name = ?", sop.Name).FirstOrCreate(&sop).Error; err != nil {
		log.Warn("Failed to seed default customer service SOP", "error", err)
	}
}
