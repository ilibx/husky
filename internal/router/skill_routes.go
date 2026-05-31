package router

import (
    "github.com/gin-gonic/gin"
    "github.com/husky/husky/internal/handler"
    "github.com/husky/husky/internal/repository"
    "gorm.io/gorm"
)

func RegisterSkillRoutes(r *gin.RouterGroup, db *gorm.DB) {
    repo := repository.NewSkillRepository(db)
    h := handler.NewSkillHandler(*repo)
    
    skillGroup := r.Group("/skills")
    {
        skillGroup.GET("", h.List)
        skillGroup.GET("/:id", h.Get)
        skillGroup.POST("", h.Create)
        skillGroup.PUT("/:id", h.Update)
        skillGroup.DELETE("/:id", h.Delete)
    }
}
