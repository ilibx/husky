package router

import (
	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/agent"
	"github.com/husky/husky/internal/middleware/auth"
)

func setupSOPRoutes(group *gin.RouterGroup, h *agent.CRUDHandler) {
	group.GET("", h.ListSOPs)
	group.GET("/:id", h.GetSOP)
	group.POST("", h.CreateSOP)
	group.PUT("/:id", h.UpdateSOP)
	group.DELETE("/:id", h.DeleteSOP)
}

func setupWorkflowRoutes(group *gin.RouterGroup, h *agent.CRUDHandler) {
	group.GET("", h.ListWorkflows)
	group.GET("/:id", h.GetWorkflow)
	group.POST("/steps/:stepId/complete", h.CompleteStep)
	group.POST("/steps/:stepId/approve", h.ApproveStep)
	group.POST("/steps/:stepId/reject", h.RejectStep)
	group.POST("/steps/:stepId/revise", h.ReviseStep)
	group.GET("/tasks", h.PendingSteps)
}

func setupAgentRoutes(group *gin.RouterGroup, h *agent.CRUDHandler) {
	group.GET("", h.ListAgents)
	group.GET("/:id", h.GetAgent)
	group.POST("", auth.RBACMiddleware("admin"), h.CreateAgent)
	group.PUT("/:id", auth.RBACMiddleware("admin"), h.UpdateAgent)
	group.DELETE("/:id", auth.RBACMiddleware("admin"), h.DeleteAgent)
}
