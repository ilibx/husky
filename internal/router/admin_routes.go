package router

import (
	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/handler"
	"github.com/husky/husky/internal/middleware/auth"
)

func setupDepartmentRoutes(group *gin.RouterGroup, h *handler.DepartmentHandler) {
	group.GET("", h.ListDepartments)
	group.GET("/:id", h.GetDepartment)
	group.POST("", auth.RBACMiddleware("admin"), h.CreateDepartment)
	group.PUT("/:id", auth.RBACMiddleware("admin"), h.UpdateDepartment)
	group.DELETE("/:id", auth.RBACMiddleware("admin"), h.DeleteDepartment)
}

func setupCategoryRoutes(group *gin.RouterGroup, h *handler.CategoryHandler) {
	group.GET("", h.ListCategories)
	group.GET("/:id", h.GetCategory)
	group.POST("", auth.RBACMiddleware("admin"), h.CreateCategory)
	group.PUT("/:id", auth.RBACMiddleware("admin"), h.UpdateCategory)
	group.DELETE("/:id", auth.RBACMiddleware("admin"), h.DeleteCategory)
}

func setupStatsRoutes(group *gin.RouterGroup, h *handler.StatsHandler) {
	group.GET("/overview", h.Overview)
	group.GET("/tickets", h.Tickets)
	group.GET("/performance", h.Performance)
	group.GET("/agents", h.AgentPerformance)
}
