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
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/ldap"
	"github.com/husky/husky/internal/middleware"
	"github.com/husky/husky/internal/middleware/auth"
	"github.com/husky/husky/internal/repository"
	"github.com/husky/husky/internal/service"
	"github.com/husky/husky/internal/ticket"
	"github.com/husky/husky/pkg/llm"
	"github.com/husky/husky/pkg/logger"
)

func SetupRouter(cfg *config.Config, dbConn *repository.DatabaseConnection, log *logger.Logger) (http.Handler, func()) {
	if cfg.ServerMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimit(100, 200))

	auth.SetJWTConfig(cfg.JWTSecret, cfg.JWTExpireHour)

	ticketRepo := repository.NewTicketRepository(dbConn.DB)
	userRepo := repository.NewUserRepository(dbConn.DB)
	kbRepo := repository.NewVectorStoreRepository(dbConn.DB)

	var embedService *llm.EmbeddingService
	var chatSvc *llm.ChatService
	if cfg.LLMAPIKey != "" {
		provider, err := llm.NewProviderFromConfig(cfg.LLMProvider, cfg.LLMAPIKey, cfg.LLMBaseURL)
		if err != nil {
			log.Warn("Failed to initialize LLM provider", "error", err)
		} else {
			embedService = llm.NewEmbeddingService(provider)
			chatSvc = llm.NewChatService(provider, llm.WithRateLimit(10, 20))
			log.Info("LLM embedding service initialized", "provider", provider.Name())
		}
	} else {
		log.Warn("LLM API key not set, knowledge embedding disabled (text search fallback)")
	}

	// --- Core services ---
	authService := service.NewAuthService(userRepo)
	ticketSvc := ticket.NewService(ticketRepo)
	knowledgeSvc := knowledge.NewService(kbRepo, embedService, repository.NewCategoryRepository(dbConn.DB), chatSvc)
	statsService := service.NewStatsService(ticketRepo)
	userService := service.NewUserService(userRepo)
	categoryService := service.NewCategoryService(
		repository.NewCategoryRepository(dbConn.DB),
	)
	deptService := service.NewDepartmentService(
		repository.NewDepartmentRepository(dbConn.DB),
	)

	// --- SLA config + escalation ---
	slaConfigRepo := repository.NewSLAConfigRepository(dbConn.DB)
	webhookRepo := repository.NewWebhookConfigRepository(dbConn.DB)
	slaConfigSvc := ticket.NewSLAConfigService(slaConfigRepo, webhookRepo, nil, ticketRepo)
	slaEscalator := ticket.NewSLAEscalator(ticketRepo, slaConfigSvc)
	slaEscalator.Start(context.Background())

	// --- Channel layer (Feishu, Lark, DingTalk, WeCom) ---
	channelCfgRepo := repository.NewChannelConfigRepository(dbConn.DB)
	channelCfgSvc := channel.NewConfigService(channelCfgRepo)

	var feishuCli *feishu.Client
	var ticketGroupSvc *channel.TicketGroupService
	if cfg.FeishuAppID != "" && cfg.FeishuAppSecret != "" {
		feishuCli = feishu.NewClient(cfg.FeishuAppID, cfg.FeishuAppSecret)
		ticketGroupSvc = channel.NewTicketGroupService(ticketRepo, feishuCli, kbRepo)
		log.Info("Feishu ticket group service initialized")
	} else {
		log.Warn("Feishu AppID/Secret not set, ticket group auto-creation disabled")
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

	// --- Register post-creation handler on ticket service ---
	// Consolidates SLA, group creation, agent execution, and SOP matching
	// into a single path. Each concern is a separate handler.
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
	gw := gateway.NewGateway(ticketSvc, feishuCli, ticketGroupSvc, knowledgeSvc, ticketSvc, intent.NewService(chatSvc), enrichers, log)

	// --- Handlers ---
	authHandler := handler.NewAuthHandler(authService, log)
	ticketHandler := ticket.NewTicketHandler(ticketSvc, slaConfigSvc, log)
	knowledgeHandler := knowledge.NewHandler(knowledgeSvc)
	statsHandler := handler.NewStatsHandler(statsService, log)
	userHandler := handler.NewUserHandler(userService, ldap.NewService(cfg, userRepo, log), log)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	deptHandler := handler.NewDepartmentHandler(deptService)
	channelHandler := channel.NewHandler(gw, channelCfgSvc, ticketSvc, log)
	agentHandler := agent.NewCRUDHandler(agentSvc, sopSvc, workflowSvc, log)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Setup permissions loader for fine-grained RBAC
	auth.SetPermissionsLoader(ticketRepo)

	adminHandler := admin.NewHandler(admin.Option{
		Mode: cfg.AdminMode,
		URL:  cfg.AdminURL,
	})
	r.Any("/admin/*path", gin.WrapH(adminHandler))
	r.Any("/admin", gin.WrapH(adminHandler))

	v1 := r.Group("/api/v1")
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

	webhooks := r.Group("/webhooks")
	setupWebhookRoutes(webhooks, channelHandler)

	return r, func() {
		slaEscalator.Stop()
	}
}
