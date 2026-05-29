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
	"github.com/husky/husky/internal/knowledge"
	"github.com/husky/husky/internal/middleware"
	"github.com/husky/husky/internal/middleware/auth"
	"github.com/husky/husky/internal/repository"
	"github.com/husky/husky/internal/service"
	"github.com/husky/husky/internal/ticket"
	"github.com/husky/husky/pkg/llm"
	"github.com/husky/husky/pkg/logger"
)

func SetupRouter(cfg *config.Config, dbConn *repository.DatabaseConnection, log *logger.Logger) http.Handler {
	if cfg.ServerMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(middleware.CORS())

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
			chatSvc = llm.NewChatService(provider)
			log.Info("LLM embedding service initialized", "provider", provider.Name())
		}
	} else {
		log.Warn("LLM API key not set, knowledge embedding disabled (text search fallback)")
	}

	// --- Core services ---
	authService := service.NewAuthService(userRepo)
	ticketSvc := ticket.NewService(ticketRepo)
	knowledgeSvc := knowledge.NewService(kbRepo, embedService)
	statsService := service.NewStatsService(ticketRepo)
	userService := service.NewUserService(userRepo)
	categoryService := service.NewCategoryService(
		repository.NewCategoryRepository(dbConn.DB),
	)
	deptService := service.NewDepartmentService(
		repository.NewDepartmentRepository(dbConn.DB),
	)

	// --- Agent layer (SOP + Workflow + Agent engine + RAG) ---
	sopRepo := repository.NewSOPRepository(dbConn.DB)
	agentRepo := repository.NewAgentRepository(dbConn.DB)
	wfRepo := repository.NewWorkflowRepository(dbConn.DB)
	workflowSvc := agent.NewWorkflowService(wfRepo, ticketRepo, agentRepo, kbRepo, chatSvc)
	agentEngine := agent.NewEngine(agentRepo, ticketRepo, kbRepo, chatSvc, sopRepo, workflowSvc)
	agentSvc := agent.NewService(agentRepo)
	sopSvc := agent.NewSOPService(sopRepo)

	// --- SLA escalation (background goroutine, 5min interval) ---
	slaEscalator := ticket.NewSLAEscalator(ticketRepo)
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

	// Register channel user enrichers for gateway
	if feishuCli != nil {
		gateway.RegisterEnricher(gateway.ChannelLark, gateway.NewFeishuEnricher(feishuCli))
	}

	// --- Gateway: channel ↔ gateway ↔ ticket | gateway ↔ agent ---
	gw := gateway.NewGateway(ticketSvc, agentEngine, feishuCli, ticketGroupSvc, log)

	// --- Handlers ---
	authHandler := handler.NewAuthHandler(authService, log)
	ticketHandler := ticket.NewTicketHandler(ticketSvc, ticketGroupSvc, agentEngine, log)
	knowledgeHandler := knowledge.NewHandler(knowledgeSvc)
	statsHandler := handler.NewStatsHandler(statsService, log)
	userHandler := handler.NewUserHandler(userService, log)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	deptHandler := handler.NewDepartmentHandler(deptService)
	channelHandler := channel.NewHandler(gw, channelCfgSvc, ticketSvc, log)
	agentHandler := agent.NewCRUDHandler(agentSvc, sopSvc, workflowSvc, log)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	adminHandler := admin.NewHandler(admin.Option{
		Mode: config.Conf.AdminMode,
		URL:  config.Conf.AdminURL,
	})
	r.Any("/admin/*path", gin.WrapH(adminHandler))
	r.Any("/admin", gin.WrapH(adminHandler))

	v1 := r.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/refresh", authHandler.Refresh)
		}

		users := v1.Group("/users")
		users.Use(auth.AuthMiddleware())
		{
			users.GET("/me", authHandler.Me)
			users.GET("", auth.RBACMiddleware("admin"), userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", auth.RBACMiddleware("admin"), userHandler.DeleteUser)
			users.PUT("/:id/role", auth.RBACMiddleware("admin"), userHandler.ChangeRole)
		}

		tickets := v1.Group("/tickets")
		tickets.Use(auth.AuthMiddleware())
		{
			tickets.POST("", ticketHandler.CreateTicket)
			tickets.GET("", ticketHandler.ListTickets)
			tickets.GET("/overdue", ticketHandler.ListOverdue)
			tickets.GET("/:id", ticketHandler.GetTicket)
			tickets.PUT("/:id", ticketHandler.UpdateTicket)
			tickets.DELETE("/:id", ticketHandler.DeleteTicket)
			tickets.POST("/:id/assign", ticketHandler.AssignTicket)
			tickets.POST("/:id/auto-assign", ticketHandler.AutoAssignTicket)
			tickets.POST("/:id/claim", ticketHandler.ClaimTicket)
			tickets.POST("/:id/watch", ticketHandler.WatchTicket)
			tickets.DELETE("/:id/watch", ticketHandler.UnwatchTicket)
			tickets.POST("/:id/status", ticketHandler.UpdateStatus)
			tickets.POST("/:id/due", ticketHandler.SetDueAt)
			tickets.POST("/:id/rate", ticketHandler.RateTicket)
			tickets.GET("/:id/rate", ticketHandler.GetSatisfaction)
			tickets.POST("/:id/comments", ticketHandler.AddComment)
			tickets.GET("/:id/comments", ticketHandler.GetComments)
			tickets.GET("/:id/audit-logs", ticketHandler.ListAuditLogs)
			tickets.POST("/:id/attachments", ticketHandler.UploadAttachment)
			tickets.GET("/:id/attachments", ticketHandler.ListAttachments)
			tickets.DELETE("/:id/attachments/:attachmentId", ticketHandler.DeleteAttachment)
		}

		if ticketGroupSvc != nil {
			feishuGroup := v1.Group("/feishu")
			feishuGroup.Use(auth.AuthMiddleware())
			{
				feishuGroup.POST("/join-group", channelHandler.JoinGroup)
			}
		}

		know := v1.Group("/knowledge")
		know.Use(auth.AuthMiddleware())
		{
			know.GET("", knowledgeHandler.ListKnowledge)
			know.POST("", knowledgeHandler.CreateKnowledge)
			know.POST("/import", knowledgeHandler.ImportKnowledge)
			know.GET("/export", knowledgeHandler.ExportKnowledge)
			know.POST("/search", knowledgeHandler.SearchKnowledge)
			know.GET("/query", knowledgeHandler.QueryKnowledge)
			know.GET("/:id", knowledgeHandler.GetKnowledge)
			know.PUT("/:id", knowledgeHandler.UpdateKnowledge)
			know.DELETE("/:id", knowledgeHandler.DeleteKnowledge)
		}

		sop := v1.Group("/sop")
		sop.Use(auth.AuthMiddleware())
		{
			sop.GET("", agentHandler.ListSOPs)
			sop.GET("/:id", agentHandler.GetSOP)
			sop.POST("", agentHandler.CreateSOP)
			sop.PUT("/:id", agentHandler.UpdateSOP)
			sop.DELETE("/:id", agentHandler.DeleteSOP)
		}

		workflows := v1.Group("/workflows")
		workflows.Use(auth.AuthMiddleware())
		{
			workflows.GET("", agentHandler.ListWorkflows)
			workflows.GET("/:id", agentHandler.GetWorkflow)
			workflows.POST("/steps/:stepId/complete", agentHandler.CompleteStep)
			workflows.GET("/tasks", agentHandler.PendingSteps)
		}

		departments := v1.Group("/departments")
		departments.Use(auth.AuthMiddleware())
		{
			departments.GET("", deptHandler.ListDepartments)
			departments.GET("/:id", deptHandler.GetDepartment)
			departments.POST("", auth.RBACMiddleware("admin"), deptHandler.CreateDepartment)
			departments.PUT("/:id", auth.RBACMiddleware("admin"), deptHandler.UpdateDepartment)
			departments.DELETE("/:id", auth.RBACMiddleware("admin"), deptHandler.DeleteDepartment)
		}

		agents := v1.Group("/agents")
		agents.Use(auth.AuthMiddleware())
		{
			agents.GET("", agentHandler.ListAgents)
			agents.GET("/:id", agentHandler.GetAgent)
			agents.POST("", auth.RBACMiddleware("admin"), agentHandler.CreateAgent)
			agents.PUT("/:id", auth.RBACMiddleware("admin"), agentHandler.UpdateAgent)
			agents.DELETE("/:id", auth.RBACMiddleware("admin"), agentHandler.DeleteAgent)
		}

		channels := v1.Group("/channels")
		channels.Use(auth.AuthMiddleware())
		{
			channels.GET("", channelHandler.ListChannelConfigs)
			channels.GET("/:id", channelHandler.GetChannelConfig)
			channels.POST("", auth.RBACMiddleware("admin"), channelHandler.CreateChannelConfig)
			channels.PUT("/:id", auth.RBACMiddleware("admin"), channelHandler.UpdateChannelConfig)
			channels.DELETE("/:id", auth.RBACMiddleware("admin"), channelHandler.DeleteChannelConfig)
		}

		categories := v1.Group("/categories")
		categories.Use(auth.AuthMiddleware())
		{
			categories.GET("", categoryHandler.ListCategories)
			categories.GET("/:id", categoryHandler.GetCategory)
			categories.POST("", auth.RBACMiddleware("admin"), categoryHandler.CreateCategory)
			categories.PUT("/:id", auth.RBACMiddleware("admin"), categoryHandler.UpdateCategory)
			categories.DELETE("/:id", auth.RBACMiddleware("admin"), categoryHandler.DeleteCategory)
		}

		notifications := v1.Group("/notifications")
		notifications.Use(auth.AuthMiddleware())
		{
			notifications.GET("", channelHandler.ListNotifications)
			notifications.GET("/unread-count", channelHandler.GetUnreadCount)
			notifications.PUT("/:id/read", channelHandler.MarkRead)
		}

		messages := v1.Group("/messages")
		messages.Use(auth.AuthMiddleware())
		{
			messages.GET("", auth.RBACMiddleware("admin"), channelHandler.ListWebhookMessages)
		}

		stats := v1.Group("/stats")
		stats.Use(auth.AuthMiddleware())
		{
			stats.GET("/overview", statsHandler.Overview)
			stats.GET("/tickets", statsHandler.Tickets)
			stats.GET("/performance", statsHandler.Performance)
		}
	}

	webhooks := r.Group("/webhooks")
	{
		webhooks.POST("/lark", channelHandler.LarkWebhook)
		webhooks.POST("/dingtalk", channelHandler.DingTalkWebhook)
		webhooks.POST("/wecom", channelHandler.WeComWebhook)
	}

	return r
}
