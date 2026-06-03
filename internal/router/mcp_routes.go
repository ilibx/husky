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

	r.GET("", h.List)
	r.GET("/:id", h.Get)
	r.POST("", h.Create)
	r.PUT("/:id", h.Update)
	r.DELETE("/:id", h.Delete)
}
