package router

import (
	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/handler"
	"github.com/husky/husky/internal/middleware/auth"
)

func setupAuthRoutes(group *gin.RouterGroup, h *handler.AuthHandler) {
	group.POST("/login", h.Login)
	group.POST("/register", h.Register)
	group.POST("/refresh", h.Refresh)
}

func setupUserRoutes(group *gin.RouterGroup, h *handler.UserHandler, ah *handler.AuthHandler) {
	group.GET("/me", ah.Me)
	group.PUT("/me", ah.UpdateProfile)
	group.GET("", auth.RBACMiddleware("admin"), h.ListUsers)
	group.GET("/:id", h.GetUser)
	group.PUT("/:id", h.UpdateUser)
	group.DELETE("/:id", auth.RBACMiddleware("admin"), h.DeleteUser)
	group.PUT("/:id/role", auth.RBACMiddleware("admin"), h.ChangeRole)
	group.POST("/sync-ldap", auth.RBACMiddleware("admin"), h.SyncLDAPUsers)
}
