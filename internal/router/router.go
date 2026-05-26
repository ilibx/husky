package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/config"
)

// SetupRouter 设置路由
func SetupRouter(cfg *config.Config) http.Handler {
	// 设置 Gin 模式
	if cfg.ServerMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

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
		auth := v1.Group("/auth")
		{
			// TODO: 实现认证路由
			auth.POST("/login", nil)
			auth.POST("/logout", nil)
			auth.POST("/refresh", nil)
		}

		// 用户管理路由
		users := v1.Group("/users")
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
		{
			// TODO: 实现工单管理路由
			tickets.GET("", nil)
			tickets.GET("/:id", nil)
			tickets.POST("", nil)
			tickets.PUT("/:id", nil)
			tickets.DELETE("/:id", nil)
			tickets.POST("/:id/comments", nil)
			tickets.POST("/:id/attachments", nil)
		}

		// 知识库路由
		knowledge := v1.Group("/knowledge")
		{
			// TODO: 实现知识库路由
			knowledge.GET("", nil)
			knowledge.GET("/:id", nil)
			knowledge.POST("", nil)
			knowledge.PUT("/:id", nil)
			knowledge.DELETE("/:id", nil)
		}

		// SOP 流程路由
		sop := v1.Group("/sop")
		{
			// TODO: 实现 SOP 路由
			sop.GET("", nil)
			sop.GET("/:id", nil)
			sop.POST("", nil)
			sop.PUT("/:id", nil)
			sop.DELETE("/:id", nil)
		}

		// Agent 管理路由
		agents := v1.Group("/agents")
		{
			// TODO: 实现 Agent 路由
			agents.GET("", nil)
			agents.GET("/:id", nil)
			agents.POST("", nil)
			agents.PUT("/:id", nil)
			agents.DELETE("/:id", nil)
		}

		// 渠道配置路由
		channels := v1.Group("/channels")
		{
			// TODO: 实现渠道路由
			channels.GET("", nil)
			channels.GET("/:id", nil)
			channels.POST("", nil)
			channels.PUT("/:id", nil)
			channels.DELETE("/:id", nil)
		}

		// 统计报表路由
		stats := v1.Group("/stats")
		{
			// TODO: 实现统计路由
			stats.GET("/overview", nil)
			stats.GET("/tickets", nil)
			stats.GET("/performance", nil)
		}
	}

	// Webhook 路由（用于渠道事件接收）
	webhooks := r.Group("/webhooks")
	{
		// 飞书 webhook
		webhooks.POST("/feishu/:channelId", nil)
		// Lark webhook
		webhooks.POST("/lark/:channelId", nil)
		// 钉钉 webhook
		webhooks.POST("/dingtalk/:channelId", nil)
		// 企业微信 webhook
		webhooks.POST("/wecom/:channelId", nil)
	}

	return r
}
