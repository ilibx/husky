package router

import (
	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/middleware/auth"
	"github.com/husky/husky/internal/ticket"
)

func setupTicketRoutes(group *gin.RouterGroup, h *ticket.TicketHandler) {
	group.POST("", h.CreateTicket)
	group.GET("", h.ListTickets)
	group.GET("/overdue", h.ListOverdue)
	group.GET("/:id", h.GetTicket)
	group.PUT("/:id", h.UpdateTicket)
	group.DELETE("/:id", h.DeleteTicket)
	group.POST("/:id/assign", h.AssignTicket)
	group.POST("/:id/auto-assign", h.AutoAssignTicket)
	group.POST("/:id/claim", h.ClaimTicket)
	group.POST("/:id/watch", h.WatchTicket)
	group.DELETE("/:id/watch", h.UnwatchTicket)
	group.POST("/:id/status", h.UpdateStatus)
	group.POST("/:id/due", h.SetDueAt)
	group.POST("/:id/rate", h.RateTicket)
	group.GET("/:id/rate", h.GetSatisfaction)
	group.POST("/:id/comments", h.AddComment)
	group.GET("/:id/comments", h.GetComments)
	group.GET("/:id/audit-logs", h.ListAuditLogs)
	group.POST("/:id/attachments", h.UploadAttachment)
	group.GET("/:id/attachments", h.ListAttachments)
	group.DELETE("/:id/attachments/:attachmentId", h.DeleteAttachment)
	group.GET("/:id/tags", h.GetTicketTags)
	group.PUT("/:id/tags", h.UpdateTicketTags)
	group.POST("/:id/tags", h.AddTicketTags)
	group.DELETE("/:id/tags/:tagId", h.RemoveTicketTag)

}

func setupSLAConfigRoutes(group *gin.RouterGroup, h *ticket.TicketHandler) {
	group.GET("", auth.RBACMiddleware("admin"), h.ListSLAConfigs)
	group.GET("/:id", auth.RBACMiddleware("admin"), h.GetSLAConfig)
	group.POST("", auth.RBACMiddleware("admin"), h.CreateSLAConfig)
	group.PUT("/:id", auth.RBACMiddleware("admin"), h.UpdateSLAConfig)
	group.DELETE("/:id", auth.RBACMiddleware("admin"), h.DeleteSLAConfig)
}

func setupWebhookConfigRoutes(group *gin.RouterGroup, h *ticket.TicketHandler) {
	group.GET("", auth.RBACMiddleware("admin"), h.ListWebhookConfigs)
	group.GET("/:id", auth.RBACMiddleware("admin"), h.GetWebhookConfig)
	group.POST("", auth.RBACMiddleware("admin"), h.CreateWebhookConfig)
	group.PUT("/:id", auth.RBACMiddleware("admin"), h.UpdateWebhookConfig)
	group.DELETE("/:id", auth.RBACMiddleware("admin"), h.DeleteWebhookConfig)
}

func setupRoleRoutes(group *gin.RouterGroup, h *ticket.TicketHandler) {
	group.GET("", auth.RBACMiddleware("admin"), h.ListRoles)
	group.GET("/:id", auth.RBACMiddleware("admin"), h.GetRole)
	group.POST("", auth.RBACMiddleware("admin"), h.CreateRole)
	group.PUT("/:id", auth.RBACMiddleware("admin"), h.UpdateRole)
	group.DELETE("/:id", auth.RBACMiddleware("admin"), h.DeleteRole)
}

func setupTagRoutes(group *gin.RouterGroup, h *ticket.TicketHandler) {
	group.POST("", h.CreateTag)
	group.GET("", h.ListTags)
	group.GET("/:id", h.GetTag)
	group.PUT("/:id", h.UpdateTag)
	group.DELETE("/:id", h.DeleteTag)
}

func setupAssignConfigRoutes(group *gin.RouterGroup, h *ticket.TicketHandler) {
	group.GET("", h.GetAssignConfig)
	group.POST("", h.SetAssignConfig)
}

func setupBotConfigRoutes(group *gin.RouterGroup, h *ticket.TicketHandler) {
	group.GET("/:channel", h.GetBotConfig)
	group.POST("", h.SetBotConfig)
}
