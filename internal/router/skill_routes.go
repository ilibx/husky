package router

import (
	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/handler"
	"github.com/husky/husky/internal/repository"
	"gorm.io/gorm"
)

func RegisterSkillRoutes(r *gin.RouterGroup, db *gorm.DB) {
	skillRepo := repository.NewSkillRepository(db)
	mcpRepo := repository.NewMCPRepository(db)
	skillH := handler.NewSkillHandler(*skillRepo)
	mcpH := handler.NewSkillMCPHandler(*mcpRepo)

	r.GET("", skillH.List)
	r.GET("/:id", skillH.Get)
	r.POST("", skillH.Create)
	r.PUT("/:id", skillH.Update)
	r.DELETE("/:id", skillH.Delete)

	r.GET("/:id/mcps", mcpH.List)
	r.POST("/:id/mcps", mcpH.Create)
	r.PUT("/:id/mcps/:mcpId", mcpH.Update)
	r.DELETE("/:id/mcps/:mcpId", mcpH.Delete)
}
