package knowledge

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

func (h *Handler) ListKnowledgeCategories(c *gin.Context) {
	categories, err := h.knowledgeService.ListCategories(c.Request.Context())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"data": categories})
}

func (h *Handler) KnowledgeCategoryTree(c *gin.Context) {
	tree, err := h.knowledgeService.CategoryTree(c.Request.Context())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"data": tree})
}

func (h *Handler) Ask(c *gin.Context) {
	var req struct {
		Question string `json:"question" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "question is required")
		return
	}

	resp, err := h.knowledgeService.Ask(c.Request.Context(), req.Question)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, resp)
}

func (h *Handler) RecommendKnowledge(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	results, err := h.knowledgeService.RecommendKnowledge(c.Request.Context(), limit)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"data": results})
}

func (h *Handler) RecordKnowledgeView(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "id is required")
		return
	}

	if err := h.knowledgeService.RecordView(c.Request.Context(), id); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"success": true})
}
