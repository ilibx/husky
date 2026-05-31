package router

import (
	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/handler"
)

func setupMenuRoutes(group *gin.RouterGroup, h *handler.MenuHandler) {
	group.GET("", h.GetMenus)
	group.GET("/all", h.ListAllMenus)
	group.POST("", h.CreateMenu)
	group.PUT("/:id", h.UpdateMenu)
	group.DELETE("/:id", h.DeleteMenu)
}
