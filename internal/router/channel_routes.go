package router

import (
	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/channel"
	"github.com/husky/husky/internal/middleware/auth"
)

func setupChannelConfigRoutes(group *gin.RouterGroup, h *channel.Handler) {
	group.GET("", h.ListChannelConfigs)
	group.GET("/:id", h.GetChannelConfig)
	group.POST("", auth.RBACMiddleware("admin"), h.CreateChannelConfig)
	group.PUT("/:id", auth.RBACMiddleware("admin"), h.UpdateChannelConfig)
	group.DELETE("/:id", auth.RBACMiddleware("admin"), h.DeleteChannelConfig)
}

func setupNotificationRoutes(group *gin.RouterGroup, h *channel.Handler) {
	group.GET("", h.ListNotifications)
	group.GET("/unread-count", h.GetUnreadCount)
	group.PUT("/:id/read", h.MarkRead)
}

func setupMessageRoutes(group *gin.RouterGroup, h *channel.Handler) {
	group.GET("", auth.RBACMiddleware("admin"), h.ListWebhookMessages)
}

func setupWebhookRoutes(group *gin.RouterGroup, h *channel.Handler) {
	group.POST("/lark", h.LarkWebhook)
	group.POST("/dingtalk", h.DingTalkWebhook)
	group.POST("/wecom", h.WeComWebhook)
}

func setupFeishuGroupRoutes(group *gin.RouterGroup, h *channel.Handler) {
	group.POST("/join-group", h.JoinGroup)
}
