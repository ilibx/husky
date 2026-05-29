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
	knowledgeSvc := knowledge.NewService(kbRepo, embedService, repository.NewCategoryRepository(dbConn.DB), chatSvc)
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

	// Register channel user enrichers for gateway
	if feishuCli != nil {
		gateway.RegisterEnricher(gateway.ChannelLark, gateway.NewFeishuEnricher(feishuCli))
	}

	// --- Gateway: channel ↔ gateway ↔ ticket | gateway ↔ agent ---
	gw := gateway.NewGateway(ticketSvc, agentEngine, feishuCli, ticketGroupSvc, knowledgeSvc, ticketSvc, intent.NewService(chatSvc), log)

	// --- Handlers ---
	authHandler := handler.NewAuthHandler(authService, log)
	ticketHandler := ticket.NewTicketHandler(ticketSvc, ticketGroupSvc, agentEngine, slaConfigSvc, log)
	knowledgeHandler := knowledge.NewHandler(knowledgeSvc)
	statsHandler := handler.NewStatsHandler(statsService, log)
	userHandler := handler.NewUserHandler(userService, ldap.NewService(userRepo, log), log)
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
		Mode: config.Conf.AdminMode,
		URL:  config.Conf.AdminURL,
	})
	r.Any("/admin/*path", gin.WrapH(adminHandler))
	r.Any("/admin", gin.WrapH(adminHandler))

	v1 := r.Group("/api/v1")
	v1.Use(auth.LoadPermissionsMiddleware())
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
			users.POST("/sync-ldap", auth.RBACMiddleware("admin"), userHandler.SyncLDAPUsers)
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
			// Ticket tags
			tickets.GET("/:id/tags", ticketHandler.GetTicketTags)
			tickets.PUT("/:id/tags", ticketHandler.UpdateTicketTags)
			tickets.POST("/:id/tags", ticketHandler.AddTicketTags)
			tickets.DELETE("/:id/tags/:tagId", ticketHandler.RemoveTicketTag)
			// Ticket custom fields
			tickets.GET("/fields/definitions", ticketHandler.ListTicketFields)
			tickets.POST("/fields/definitions", ticketHandler.CreateTicketField)
			tickets.PUT("/fields/definitions/:id", ticketHandler.UpdateTicketField)
			tickets.DELETE("/fields/definitions/:id", ticketHandler.DeleteTicketField)
			tickets.GET("/:id/fields", ticketHandler.GetTicketFieldValues)
			tickets.PUT("/:id/fields", ticketHandler.UpdateTicketFieldValues)
		}
		// Global tags
		tags := v1.Group("/tags")
		tags.Use(auth.AuthMiddleware())
		{
			tags.POST("", ticketHandler.CreateTag)
			tags.GET("", ticketHandler.ListTags)
			tags.GET("/:id", ticketHandler.GetTag)
			tags.PUT("/:id", ticketHandler.UpdateTag)
			tags.DELETE("/:id", ticketHandler.DeleteTag)
		}
		// Ticket relations
		tickets.POST("/:id/relations", ticketHandler.CreateTicketRelation)
		tickets.GET("/:id/relations", ticketHandler.ListTicketRelations)
		tickets.DELETE("/:id/relations/:relationId", ticketHandler.DeleteTicketRelation)
		// Assign config
		assign := v1.Group("/assign-config")
		assign.Use(auth.AuthMiddleware())
		{
			assign.GET("", ticketHandler.GetAssignConfig)
			assign.POST("", ticketHandler.SetAssignConfig)
		}
		// Bot config
		bot := v1.Group("/bot-config")
		bot.Use(auth.AuthMiddleware())
		{
			bot.GET("/:channel", ticketHandler.GetBotConfig)
			bot.POST("", ticketHandler.SetBotConfig)
		}

		sla := v1.Group("/sla-configs")
		sla.Use(auth.AuthMiddleware())
		{
			sla.GET("", auth.RBACMiddleware("admin"), ticketHandler.ListSLAConfigs)
			sla.GET("/:id", auth.RBACMiddleware("admin"), ticketHandler.GetSLAConfig)
			sla.POST("", auth.RBACMiddleware("admin"), ticketHandler.CreateSLAConfig)
			sla.PUT("/:id", auth.RBACMiddleware("admin"), ticketHandler.UpdateSLAConfig)
			sla.DELETE("/:id", auth.RBACMiddleware("admin"), ticketHandler.DeleteSLAConfig)
		}

		webhooksCfg := v1.Group("/webhook-configs")
		webhooksCfg.Use(auth.AuthMiddleware())
		{
			webhooksCfg.GET("", auth.RBACMiddleware("admin"), ticketHandler.ListWebhookConfigs)
			webhooksCfg.GET("/:id", auth.RBACMiddleware("admin"), ticketHandler.GetWebhookConfig)
			webhooksCfg.POST("", auth.RBACMiddleware("admin"), ticketHandler.CreateWebhookConfig)
			webhooksCfg.PUT("/:id", auth.RBACMiddleware("admin"), ticketHandler.UpdateWebhookConfig)
			webhooksCfg.DELETE("/:id", auth.RBACMiddleware("admin"), ticketHandler.DeleteWebhookConfig)
		}

		if ticketGroupSvc != nil {
			feishuGroup := v1.Group("/feishu")
			feishuGroup.Use(auth.AuthMiddleware())
			{
				feishuGroup.POST("/join-group", channelHandler.JoinGroup)
		}
		// Roles
		roles := v1.Group("/roles")
		roles.Use(auth.AuthMiddleware())
		{
			roles.GET("", auth.RBACMiddleware("admin"), ticketHandler.ListRoles)
			roles.GET("/:id", auth.RBACMiddleware("admin"), ticketHandler.GetRole)
			roles.POST("", auth.RBACMiddleware("admin"), ticketHandler.CreateRole)
			roles.PUT("/:id", auth.RBACMiddleware("admin"), ticketHandler.UpdateRole)
			roles.DELETE("/:id", auth.RBACMiddleware("admin"), ticketHandler.DeleteRole)
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
			know.GET("/categories", knowledgeHandler.ListKnowledgeCategories)
			know.GET("/categories/tree", knowledgeHandler.KnowledgeCategoryTree)
			know.POST("/ask", knowledgeHandler.Ask)
			know.GET("/recommend", knowledgeHandler.RecommendKnowledge)
			know.POST("/:id/view", knowledgeHandler.RecordKnowledgeView)
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
			workflows.POST("/steps/:stepId/approve", agentHandler.ApproveStep)
			workflows.POST("/steps/:stepId/reject", agentHandler.RejectStep)
			workflows.POST("/steps/:stepId/revise", agentHandler.ReviseStep)
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
