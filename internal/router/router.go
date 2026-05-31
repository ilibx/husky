package router

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/admin"
	"github.com/husky/husky/internal/agent"
	"github.com/husky/husky/internal/channel"
	"github.com/husky/husky/internal/channel/feishu"
	"github.com/husky/husky/internal/config"
	"github.com/husky/husky/internal/gateway"
	"github.com/husky/husky/internal/handler"
	"github.com/husky/husky/internal/intent"
	"github.com/husky/husky/internal/knowledge"
	"github.com/husky/husky/internal/ldap"
	"github.com/husky/husky/internal/middleware"
	"github.com/husky/husky/internal/middleware/auth"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
	"github.com/husky/husky/internal/service"
	"github.com/husky/husky/internal/ticket"
	"github.com/husky/husky/pkg/llm"
	"github.com/husky/husky/pkg/logger"
)

func SetupRouter(cfg *config.Config, dbConn *repository.DatabaseConnection, log *logger.Logger) (http.Handler, func()) {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(100, 200))

	auth.SetJWTConfig(cfg.JWT.Secret, cfg.JWT.ExpireHour)

	ticketRepo := repository.NewTicketRepository(dbConn.DB)
	userRepo := repository.NewUserRepository(dbConn.DB)
	kbRepo := repository.NewVectorStoreRepository(dbConn.DB)
	sysCfgRepo := repository.NewSystemConfigRepository(dbConn.DB)

	// --- Dynamic runtime config (DB-backed, hot-reloadable via API) ---
	dynamicCfg := config.NewDynamicConfig(sysCfgRepo)
	if err := dynamicCfg.InitLLM(context.Background()); err != nil {
		log.Warn("Dynamic LLM config not initialized, use admin UI to configure LLM",
			"error", err)
	}
	if err := dynamicCfg.InitVector(context.Background(), dbConn.DB); err != nil {
		log.Warn("Dynamic vector store not initialized, configure via admin UI",
			"error", err)
	}

	// --- Core services ---
	authService := service.NewAuthService(userRepo)
	ticketSvc := ticket.NewService(ticketRepo)

	var chatSvc *llm.ChatService
	var embedService *llm.EmbeddingService
	if dynamicCfg.IsReady() {
		chatSvc = dynamicCfg.GetChatService()
		embedService = dynamicCfg.GetEmbeddingService()
	}

	knowledgeSvc := knowledge.NewService(kbRepo, embedService, repository.NewCategoryRepository(dbConn.DB), chatSvc)
	statsService := service.NewStatsService(ticketRepo)
	userService := service.NewUserService(userRepo)
	categoryService := service.NewCategoryService(
		repository.NewCategoryRepository(dbConn.DB),
	)
	deptService := service.NewDepartmentService(
		repository.NewDepartmentRepository(dbConn.DB),
	)

	// --- SLA repos + channel layer ---
	slaConfigRepo := repository.NewSLAConfigRepository(dbConn.DB)
	webhookRepo := repository.NewWebhookConfigRepository(dbConn.DB)

	// --- Channel layer (Feishu, DingTalk, WeCom) ---
	channelCfgRepo := repository.NewChannelConfigRepository(dbConn.DB)
	channelCfgSvc := channel.NewConfigService(channelCfgRepo)

	// 从 DB 加载飞书凭据（由管理页面 ChannelConfig 配置）
	var feishuCli *feishu.Client
	var ticketGroupSvc *channel.TicketGroupService
	if feishuCfg := loadFeishuConfig(context.Background(), channelCfgRepo); feishuCfg != nil {
		feishuCli = feishu.NewClient(feishuCfg.AppID, feishuCfg.AppSecret)
		ticketGroupSvc = channel.NewTicketGroupService(ticketRepo, feishuCli, kbRepo)
		log.Info("Feishu client initialized from DB channel config")
	} else {
		log.Warn("Feishu not configured, ticket group auto-creation disabled (configure via Channels admin page)")
	}

	// --- Agent layer (SOP + Workflow + Agent engine + RAG) ---
	sopRepo := repository.NewSOPRepository(dbConn.DB)
	agentRepo := repository.NewAgentRepository(dbConn.DB)
	wfRepo := repository.NewWorkflowRepository(dbConn.DB)
	workflowSvc := agent.NewWorkflowService(wfRepo, ticketRepo, agentRepo, kbRepo, chatSvc)
	workflowSvc.SetAgentUserID(1)
	sopMatcher := agent.NewSOPMatcher(sopRepo, workflowSvc, chatSvc)
	agentEngine := agent.NewEngine(agentRepo, ticketRepo, kbRepo, chatSvc, ticketSvc)
	agentSvc := agent.NewService(agentRepo)
	sopSvc := agent.NewSOPService(sopRepo)

	var slaConfigSvc *ticket.SLAConfigService

	// --- Register post-creation handler on ticket service ---
	ticketSvc.OnTicketCreated(func(ctx context.Context, t *model.Ticket) {
		if slaConfigSvc != nil {
			if dueAt := slaConfigSvc.ComputeDueAt(ctx, t); dueAt != nil {
				ticketSvc.SetDueAt(ctx, t.ID, *dueAt)
			}
		}
		if ticketGroupSvc != nil {
			if _, err := ticketGroupSvc.CreateGroupForTicket(ctx, t); err != nil {
				log.Error("Failed to create group", "error", err, "ticket_id", t.ID)
			}
		}
	})
	ticketSvc.OnTicketCreated(func(ctx context.Context, t *model.Ticket) {
		if agentEngine != nil {
			agentEngine.OnTicketCreated(ctx, t)
		}
	})
	ticketSvc.OnTicketCreated(func(ctx context.Context, t *model.Ticket) {
		if sopMatcher != nil {
			sopMatcher.MatchAndCreateWorkflow(ctx, t)
		}
	})

	// Build channel user enrichers for gateway
	enrichers := make(map[gateway.ChannelType]gateway.UserEnricher)
	if feishuCli != nil {
		enrichers[gateway.ChannelLark] = gateway.NewFeishuEnricher(feishuCli)
	}

	// --- Gateway: channel ↔ gateway ↔ ticket | gateway ↔ agent ---
	gw := gateway.NewGateway(ticketSvc, feishuCli, ticketGroupSvc, knowledgeSvc, ticketSvc, intent.NewService(chatSvc), enrichers, userRepo, log)

	// --- SLA config + escalation (after gw is available for channel notifications) ---
	slaEventHandler := ticket.SLAEventAdapterFunc(func(ctx context.Context, channel, targetID, content string) error {
		gw.SendToChannel(ctx, &gateway.OutboundMessage{
			Channel:  gateway.ChannelType(channel),
			TargetID: targetID,
			Content:  content,
			MsgType:  "text",
		})
		return nil
	})
	slaConfigSvc = ticket.NewSLAConfigService(slaConfigRepo, webhookRepo, sysCfgRepo, slaEventHandler, ticketRepo)
	slaEscalator := ticket.NewSLAEscalator(ticketRepo, slaConfigSvc)
	slaEscalator.Start(context.Background())

	// --- Handlers ---
	authHandler := handler.NewAuthHandler(authService, log)
	ticketHandler := ticket.NewTicketHandler(ticketSvc, slaConfigSvc, log)
	storageRoot := cfg.Storage.Options["root"]
	if storageRoot == "" {
		storageRoot = "./data/knowledge"
	}
	knowledgeHandler := knowledge.NewHandler(knowledgeSvc, storageRoot)
	statsHandler := handler.NewStatsHandler(statsService, log)
	userHandler := handler.NewUserHandler(userService, ldap.NewService(cfg, userRepo, log), log)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	deptHandler := handler.NewDepartmentHandler(deptService)
	channelHandler := channel.NewHandler(gw, channelCfgSvc, ticketSvc, log)
	channelUserRepo := repository.NewChannelUserRepository(dbConn.DB)
	channelGroupRepo := repository.NewChannelGroupRepository(dbConn.DB)
	channelUserHandler := channel.NewChannelUserHandler(channelUserRepo)
	channelGroupHandler := channel.NewChannelGroupHandler(channelGroupRepo)
	agentHandler := agent.NewCRUDHandler(agentSvc, sopSvc, workflowSvc, log)
	systemCfgHandler := handler.NewSystemConfigHandler(sysCfgRepo, dynamicCfg, log)
	menuRepo := repository.NewMenuRepository(dbConn.DB)
	menuHandler := handler.NewMenuHandler(menuRepo)

	adminHandler := admin.NewHandler(admin.Option{
		BasePath: cfg.Server.BasePath,
	})

	if cfg.Server.BasePath == "" {
		r.Any("/admin/*path", adminHandler)
		r.Any("/admin", adminHandler)
		r.NoRoute(adminHandler)
	} else {
		r.Any(cfg.Server.BasePath+"/*path", adminHandler)
		r.Any(cfg.Server.BasePath, adminHandler)
	}

	r.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "husky-api", "status": "running"})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Setup permissions loader for fine-grained RBAC
	auth.SetPermissionsLoader(ticketRepo)

	v1 := r.Group("/api")
	v1.Use(auth.LoadPermissionsMiddleware())

	setupAuthRoutes(v1.Group("/auth"), authHandler)

	users := v1.Group("/users")
	users.Use(auth.AuthMiddleware())
	setupUserRoutes(users, userHandler, authHandler)

	tickets := v1.Group("/tickets")
	tickets.Use(auth.AuthMiddleware())
	setupTicketRoutes(tickets, ticketHandler)

	tagRoutes := v1.Group("/tags")
	tagRoutes.Use(auth.AuthMiddleware())
	setupTagRoutes(tagRoutes, ticketHandler)

	tickets.POST("/:id/relations", ticketHandler.CreateTicketRelation)
	tickets.GET("/:id/relations", ticketHandler.ListTicketRelations)
	tickets.DELETE("/:id/relations/:relationId", ticketHandler.DeleteTicketRelation)

	assignRoutes := v1.Group("/assign-config")
	assignRoutes.Use(auth.AuthMiddleware())
	setupAssignConfigRoutes(assignRoutes, ticketHandler)

	botRoutes := v1.Group("/bot-config")
	botRoutes.Use(auth.AuthMiddleware())
	setupBotConfigRoutes(botRoutes, ticketHandler)

	sla := v1.Group("/sla-configs")
	sla.Use(auth.AuthMiddleware())
	setupSLAConfigRoutes(sla, ticketHandler)

	webhookCfgRoutes := v1.Group("/webhook-configs")
	webhookCfgRoutes.Use(auth.AuthMiddleware())
	setupWebhookConfigRoutes(webhookCfgRoutes, ticketHandler)

	if ticketGroupSvc != nil {
		feishuGroup := v1.Group("/feishu")
		feishuGroup.Use(auth.AuthMiddleware())
		setupFeishuGroupRoutes(feishuGroup, channelHandler)
	}

	roles := v1.Group("/roles")
	roles.Use(auth.AuthMiddleware())
	setupRoleRoutes(roles, ticketHandler)

	know := v1.Group("/knowledge")
	know.Use(auth.AuthMiddleware())
	setupKnowledgeRoutes(know, knowledgeHandler)

	sop := v1.Group("/sop")
	sop.Use(auth.AuthMiddleware())
	setupSOPRoutes(sop, agentHandler)

	workflows := v1.Group("/workflows")
	workflows.Use(auth.AuthMiddleware())
	setupWorkflowRoutes(workflows, agentHandler)

	departments := v1.Group("/departments")
	departments.Use(auth.AuthMiddleware())
	setupDepartmentRoutes(departments, deptHandler)

	agents := v1.Group("/agents")
	agents.Use(auth.AuthMiddleware())
	setupAgentRoutes(agents, agentHandler)

	channels := v1.Group("/channels")
	channels.Use(auth.AuthMiddleware())
	setupChannelConfigRoutes(channels, channelHandler)

	channelUsers := v1.Group("/channel-users")
	channelUsers.Use(auth.AuthMiddleware())
	setupChannelUserRoutes(channelUsers, channelUserHandler)

	channelGroups := v1.Group("/channel-groups")
	channelGroups.Use(auth.AuthMiddleware())
	setupChannelGroupRoutes(channelGroups, channelGroupHandler)

	categories := v1.Group("/categories")
	categories.Use(auth.AuthMiddleware())
	setupCategoryRoutes(categories, categoryHandler)

	notifications := v1.Group("/notifications")
	notifications.Use(auth.AuthMiddleware())
	setupNotificationRoutes(notifications, channelHandler)

	messages := v1.Group("/messages")
	messages.Use(auth.AuthMiddleware())
	setupMessageRoutes(messages, channelHandler)

	stats := v1.Group("/stats")
	stats.Use(auth.AuthMiddleware())
	setupStatsRoutes(stats, statsHandler)

	// --- System Config (runtime dynamic config API, admin only) ---
	cfgRoutes := v1.Group("/system-config")
	cfgRoutes.Use(auth.AuthMiddleware(), auth.RBACMiddleware("admin"))
	{
		cfgRoutes.GET("", systemCfgHandler.ListSystemConfigs)
		cfgRoutes.GET("/:id", systemCfgHandler.GetSystemConfig)
		cfgRoutes.GET("/llm", systemCfgHandler.GetLLMConfig)
		cfgRoutes.GET("/vector", systemCfgHandler.GetVectorConfig)
		cfgRoutes.GET("/lookup", systemCfgHandler.GetSystemConfigByKey)
		cfgRoutes.POST("", systemCfgHandler.CreateSystemConfig)
		cfgRoutes.PUT("/:id", systemCfgHandler.UpdateSystemConfig)
		cfgRoutes.POST("/upsert", systemCfgHandler.UpsertSystemConfig)
		cfgRoutes.DELETE("/:id", systemCfgHandler.DeleteSystemConfig)
	}

	// --- Menu (dynamic menu system, role-filtered) ---
	menus := v1.Group("/menus")
	menus.Use(auth.AuthMiddleware())
	{
		menus.GET("", menuHandler.GetMenus)          // returns menus for current role (all auth users)
		menus.GET("/all", auth.RBACMiddleware("admin"), menuHandler.ListAllMenus)
		menus.POST("", auth.RBACMiddleware("admin"), menuHandler.CreateMenu)
		menus.PUT("/:id", auth.RBACMiddleware("admin"), menuHandler.UpdateMenu)
		menus.DELETE("/:id", auth.RBACMiddleware("admin"), menuHandler.DeleteMenu)
	}

	// --- Skill (AI skill management) ---
	skills := v1.Group("/skills")
	skills.Use(auth.AuthMiddleware(), auth.RBACMiddleware("admin"))
	RegisterSkillRoutes(skills, dbConn.DB)

	// --- MCP (Managed Control Plane services) ---
	mcps := v1.Group("/mcps")
	mcps.Use(auth.AuthMiddleware(), auth.RBACMiddleware("admin"))
	RegisterMCPRoutes(mcps, dbConn.DB)

	webhooks := r.Group("/webhooks")
	setupWebhookRoutes(webhooks, channelHandler)

	return r, func() {
		slaEscalator.Stop()
	}
}

// loadFeishuConfig 从 DB 加载飞书凭据
func loadFeishuConfig(ctx context.Context, repo *repository.ChannelConfigRepository) *struct {
	AppID     string
	AppSecret string
} {
	configs, err := repo.List(ctx, "")
	if err != nil {
		return nil
	}
	for _, c := range configs {
		if c.Type == "lark" && c.Enabled && c.AppID != "" && c.AppSecret != "" {
			return &struct {
				AppID     string
				AppSecret string
			}{AppID: c.AppID, AppSecret: c.AppSecret}
		}
	}
	return nil
}
