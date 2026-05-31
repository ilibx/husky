package router

import (
    "github.com/gin-gonic/gin"
    "github.com/husky/husky/internal/handler"
    "github.com/husky/husky/internal/repository"
    "gorm.io/gorm"
)

func RegisterMCPRoutes(r *gin.RouterGroup, db *gorm.DB) {
    repo := repository.NewMCPRepository(db)
    h := handler.NewMCPHandler(*repo)
    
    mcpGroup := r.Group("/mcps")
    {
        mcpGroup.GET("", h.List)
        mcpGroup.GET("/:id", h.Get)
        mcpGroup.POST("", h.Create)
        mcpGroup.PUT("/:id", h.Update)
        mcpGroup.DELETE("/:id", h.Delete)
    }
}
