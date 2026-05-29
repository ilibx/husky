package knowledge

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
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
	category := c.Query("category")
	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)

	list, total, err := h.knowledgeService.ListKnowledge(c.Request.Context(), offset, limit, category)
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
	format := c.DefaultQuery("format", "json")

	switch format {
	case "csv":
		h.importCSV(c)
	default:
		h.importJSON(c)
	}
}

func (h *Handler) importJSON(c *gin.Context) {
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

func (h *Handler) importCSV(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "file is required"))
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, "failed to open file"))
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "failed to parse CSV: "+err.Error()))
		return
	}

	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "CSV must have header and at least one row"))
		return
	}

	headers := make([]string, len(records[0]))
	for i, h := range records[0] {
		headers[i] = strings.TrimSpace(strings.ToLower(h))
	}

	var items []model.KnowledgeBase
	for _, row := range records[1:] {
		if len(row) != len(headers) {
			continue
		}
		item := model.KnowledgeBase{Status: "active"}
		for i, val := range row {
			val = strings.TrimSpace(val)
			switch headers[i] {
			case "title":
				item.Title = val
			case "content":
				item.Content = val
			case "category":
				item.Category = val
			case "tags":
				if val != "" {
					tags := strings.Split(val, ";")
					tagJSON, _ := json.Marshal(tags)
					item.Tags = tagJSON
				}
			}
		}
		if item.Title != "" {
			items = append(items, item)
		}
	}

	if len(items) == 0 {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "no valid items found in CSV"))
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
	format := c.DefaultQuery("format", "json")

	switch format {
	case "csv":
		h.exportCSV(c)
	default:
		h.exportJSON(c)
	}
}

func (h *Handler) exportJSON(c *gin.Context) {
	items, err := h.knowledgeService.ExportKnowledge(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items, "total": len(items)})
}

func (h *Handler) exportCSV(c *gin.Context) {
	items, err := h.knowledgeService.ExportKnowledge(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=knowledge_export_%d.csv", len(items)))

	writer := csv.NewWriter(c.Writer)
	writer.Write([]string{"title", "content", "category", "tags"})

	for _, item := range items {
		var tags []string
		if item.Tags != nil {
			json.Unmarshal(item.Tags, &tags)
		}
		writer.Write([]string{
			item.Title,
			item.Content,
			item.Category,
			strings.Join(tags, ";"),
		})
	}
	writer.Flush()
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

func (h *Handler) ListKnowledgeCategories(c *gin.Context) {
	categories, err := h.knowledgeService.ListCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": categories})
}

func (h *Handler) KnowledgeCategoryTree(c *gin.Context) {
	tree, err := h.knowledgeService.CategoryTree(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tree})
}

func (h *Handler) Ask(c *gin.Context) {
	var req struct {
		Question string `json:"question" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "question is required"))
		return
	}

	resp, err := h.knowledgeService.Ask(c.Request.Context(), req.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) RecommendKnowledge(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	results, err := h.knowledgeService.RecommendKnowledge(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": results})
}

func (h *Handler) RecordKnowledgeView(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "id is required"))
		return
	}

	if err := h.knowledgeService.RecordView(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
