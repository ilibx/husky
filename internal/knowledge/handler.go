package knowledge

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
)

type Handler struct {
	knowledgeService Service
}

func NewHandler(knowledgeService Service) *Handler {
	return &Handler{
		knowledgeService: knowledgeService,
	}
}

func (h *Handler) CreateKnowledge(c *gin.Context) {
	var req model.KnowledgeBase
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	kb, err := h.knowledgeService.CreateKnowledge(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, kb)
}

func (h *Handler) SearchKnowledge(c *gin.Context) {
	var query model.KnowledgeQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	results, err := h.knowledgeService.SearchKnowledge(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  results,
		"count": len(results),
		"query": query.Query,
	})
}

func (h *Handler) GetKnowledge(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "id is required"))
		return
	}

	kb, err := h.knowledgeService.GetKnowledge(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	if kb == nil {
		c.JSON(http.StatusNotFound, errors.NewErrorResponse(errors.ErrNotFound, "knowledge not found"))
		return
	}

	c.JSON(http.StatusOK, kb)
}

func (h *Handler) UpdateKnowledge(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "id is required"))
		return
	}

	var req model.KnowledgeBase
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	req.ID = id

	if err := h.knowledgeService.UpdateKnowledge(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, req)
}

func (h *Handler) ListKnowledge(c *gin.Context) {
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "20")
	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)

	list, total, err := h.knowledgeService.ListKnowledge(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  list,
		"total": total,
	})
}

func (h *Handler) DeleteKnowledge(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "id is required"))
		return
	}

	if err := h.knowledgeService.DeleteKnowledge(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ImportKnowledge(c *gin.Context) {
	var items []model.KnowledgeBase
	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	count, err := h.knowledgeService.ImportKnowledge(c.Request.Context(), items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "import complete", "imported": count})
}

func (h *Handler) ExportKnowledge(c *gin.Context) {
	items, err := h.knowledgeService.ExportKnowledge(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items, "total": len(items)})
}

func (h *Handler) QueryKnowledge(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "query parameter 'q' is required"))
		return
	}

	category := c.Query("category")
	tagsStr := c.Query("tags")
	limitStr := c.Query("limit")

	var tags []string
	if tagsStr != "" {
		for _, tag := range strings.Split(tagsStr, ",") {
			if tag != "" {
				tags = append(tags, strings.TrimSpace(tag))
			}
		}
	}

	limit := 5
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	query := model.KnowledgeQuery{
		Query:    q,
		Category: category,
		Tags:     tags,
		Limit:    limit,
	}

	results, err := h.knowledgeService.SearchKnowledge(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  results,
		"count": len(results),
		"query": q,
	})
}
