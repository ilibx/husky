package router

import (
	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/knowledge"
)

func setupKnowledgeRoutes(group *gin.RouterGroup, h *knowledge.Handler) {
	group.GET("", h.ListKnowledge)
	group.POST("", h.CreateKnowledge)
	group.POST("/import", h.ImportKnowledge)
	group.GET("/export", h.ExportKnowledge)
	group.POST("/search", h.SearchKnowledge)
	group.GET("/query", h.QueryKnowledge)
	group.GET("/:id", h.GetKnowledge)
	group.PUT("/:id", h.UpdateKnowledge)
	group.DELETE("/:id", h.DeleteKnowledge)
	group.GET("/categories", h.ListKnowledgeCategories)
	group.GET("/categories/tree", h.KnowledgeCategoryTree)
	group.POST("/ask", h.Ask)
	group.GET("/recommend", h.RecommendKnowledge)
	group.POST("/:id/view", h.RecordKnowledgeView)
	group.POST("/:id/feedback", h.RecordFeedback)
	group.GET("/:id/feedback", h.GetFeedbackStats)
}
