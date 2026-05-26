package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/config"
	"github.com/husky/husky/internal/handler"
	"github.com/husky/husky/internal/middleware/auth"
	"github.com/husky/husky/internal/repository"
	"github.com/husky/husky/internal/service"
	"github.com/husky/husky/pkg/logger"
)

// SetupRouter 设置路由
func SetupRouter(cfg *config.Config, dbConn *repository.DatabaseConnection, log *logger.Logger) http.Handler {
	// 设置 Gin 模式
	if cfg.ServerMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 初始化仓库
	ticketRepo := repository.NewTicketRepository(dbConn.DB)
	kbRepo := repository.NewVectorStoreRepository(dbConn.DB)

	// 初始化服务
	ticketService := service.NewTicketService(ticketRepo)
	knowledgeService := service.NewKnowledgeService(kbRepo)

	// 初始化处理器
	ticketHandler := handler.NewTicketHandler(ticketService, log)
	knowledgeHandler := handler.NewKnowledgeHandler(knowledgeService)
	webhookHandler := handler.NewWebhookHandler(ticketService, log)

	// 健康检查端点
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 认证相关路由
		authGroup := v1.Group("/auth")
		{
			// TODO: 实现认证路由
			authGroup.POST("/login", nil)
			authGroup.POST("/logout", nil)
			authGroup.POST("/refresh", nil)
		}

		// 用户管理路由
		users := v1.Group("/users")
		users.Use(auth.AuthMiddleware())
		{
			// TODO: 实现用户管理路由
			users.GET("", nil)
			users.GET("/:id", nil)
			users.POST("", nil)
			users.PUT("/:id", nil)
			users.DELETE("/:id", nil)
		}

		// 工单管理路由
		tickets := v1.Group("/tickets")
		tickets.Use(auth.AuthMiddleware())
		{
			tickets.POST("", ticketHandler.CreateTicket)
			tickets.GET("", ticketHandler.ListTickets)
			tickets.GET("/:id", ticketHandler.GetTicket)
			tickets.PUT("/:id", ticketHandler.UpdateTicket)
			tickets.DELETE("/:id", ticketHandler.DeleteTicket)
			tickets.POST("/:id/assign", ticketHandler.AssignTicket)
			tickets.POST("/:id/status", ticketHandler.UpdateStatus)
			tickets.POST("/:id/comments", ticketHandler.AddComment)
			tickets.GET("/:id/comments", ticketHandler.GetComments)
		}

		// 知识库路由
		knowledge := v1.Group("/knowledge")
		knowledge.Use(auth.AuthMiddleware())
		{
			knowledge.POST("", knowledgeHandler.CreateKnowledge)
			knowledge.POST("/search", knowledgeHandler.SearchKnowledge)
			knowledge.GET("/query", knowledgeHandler.QueryKnowledge)
			knowledge.GET("/:id", knowledgeHandler.GetKnowledge)
			knowledge.PUT("/:id", knowledgeHandler.UpdateKnowledge)
			knowledge.DELETE("/:id", knowledgeHandler.DeleteKnowledge)
		}

		// SOP 流程路由
		sop := v1.Group("/sop")
		sop.Use(auth.AuthMiddleware())
		{
			sop.GET("", nil)
			sop.GET("/:id", nil)
			sop.POST("", nil)
			sop.PUT("/:id", nil)
			sop.DELETE("/:id", nil)
		}

		// Agent 管理路由
		agents := v1.Group("/agents")
		agents.Use(auth.AuthMiddleware())
		{
			agents.GET("", nil)
			agents.GET("/:id", nil)
			agents.POST("", nil)
			agents.PUT("/:id", nil)
			agents.DELETE("/:id", nil)
		}

		// 渠道配置路由
		channels := v1.Group("/channels")
		channels.Use(auth.AuthMiddleware())
		{
			channels.GET("", nil)
			channels.GET("/:id", nil)
			channels.POST("", nil)
			channels.PUT("/:id", nil)
			channels.DELETE("/:id", nil)
		}

		// 统计报表路由
		stats := v1.Group("/stats")
		stats.Use(auth.AuthMiddleware())
		{
			stats.GET("/overview", nil)
			stats.GET("/tickets", nil)
			stats.GET("/performance", nil)
		}
	}

	// Webhook 路由（用于渠道事件接收）- 不需要认证
	webhooks := r.Group("/webhooks")
	{
		webhooks.POST("/lark", webhookHandler.LarkWebhook)
		webhooks.POST("/dingtalk", webhookHandler.DingTalkWebhook)
		webhooks.POST("/wecom", webhookHandler.WeComWebhook)
	}

	return r
}
