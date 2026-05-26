package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/service"
	"github.com/husky/husky/pkg/errors"
)

// KnowledgeHandler 知识库处理器
type KnowledgeHandler struct {
	knowledgeService service.KnowledgeService
}

// NewKnowledgeHandler 创建知识库处理器实例
func NewKnowledgeHandler(knowledgeService service.KnowledgeService) *KnowledgeHandler {
	return &KnowledgeHandler{
		knowledgeService: knowledgeService,
	}
}

// CreateKnowledge 创建知识条目
// @Summary 创建知识条目
// @Description 创建新的知识库条目，自动生成向量嵌入
// @Tags knowledge
// @Accept json
// @Produce json
// @Param request body model.KnowledgeBase true "知识条目信息"
// @Success 201 {object} model.KnowledgeBase
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/knowledge [post]
func (h *KnowledgeHandler) CreateKnowledge(c *gin.Context) {
	var req model.KnowledgeBase
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	// 从上下文中获取用户 ID（如果已实现认证）
	// userID := c.GetString("user_id")
	// req.CreatedBy = userID

	kb, err := h.knowledgeService.CreateKnowledge(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, kb)
}

// SearchKnowledge 搜索知识
// @Summary 搜索知识
// @Description 基于向量相似度搜索相关知识条目
// @Tags knowledge
// @Accept json
// @Produce json
// @Param request body model.KnowledgeQuery true "搜索查询"
// @Success 200 {array} model.KnowledgeResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/knowledge/search [post]
func (h *KnowledgeHandler) SearchKnowledge(c *gin.Context) {
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
		"data":   results,
		"count":  len(results),
		"query":  query.Query,
	})
}

// GetKnowledge 获取单个知识条目
// @Summary 获取知识条目
// @Description 根据 ID 获取知识条目详情
// @Tags knowledge
// @Accept json
// @Produce json
// @Param id path string true "知识条目 ID"
// @Success 200 {object} model.KnowledgeBase
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/knowledge/:id [get]
func (h *KnowledgeHandler) GetKnowledge(c *gin.Context) {
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

// UpdateKnowledge 更新知识条目
// @Summary 更新知识条目
// @Description 更新现有知识条目
// @Tags knowledge
// @Accept json
// @Produce json
// @Param id path string true "知识条目 ID"
// @Param request body model.KnowledgeBase true "知识条目信息"
// @Success 200 {object} model.KnowledgeBase
// @Failure 400 {object} errors.ErrorResponse
// @Failure 404 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/knowledge/:id [put]
func (h *KnowledgeHandler) UpdateKnowledge(c *gin.Context) {
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

// DeleteKnowledge 删除知识条目
// @Summary 删除知识条目
// @Description 根据 ID 删除知识条目
// @Tags knowledge
// @Accept json
// @Produce json
// @Param id path string true "知识条目 ID"
// @Success 204
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/knowledge/:id [delete]
func (h *KnowledgeHandler) DeleteKnowledge(c *gin.Context) {
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

// QueryKnowledge 通过查询参数搜索知识（GET 方式）
// @Summary 查询知识
// @Description 通过查询参数搜索相关知识
// @Tags knowledge
// @Accept json
// @Produce json
// @Param q query string true "搜索关键词"
// @Param category query string false "分类"
// @Param tags query string false "标签 (逗号分隔)"
// @Param limit query int false "返回数量" default(5)
// @Success 200 {array} model.KnowledgeResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/v1/knowledge/query [get]
func (h *KnowledgeHandler) QueryKnowledge(c *gin.Context) {
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
		// 简单分割逗号分隔的标签
		// 实际生产环境可能需要更复杂的解析
		for _, tag := range splitString(tagsStr, ",") {
			if tag != "" {
				tags = append(tags, tag)
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
		"data":   results,
		"count":  len(results),
		"query":  q,
	})
}

// splitString 简单的字符串分割辅助函数
func splitString(s, sep string) []string {
	// 简单实现，实际可以使用 strings.Split
	result := []string{}
	current := ""
	for _, ch := range s {
		if string(ch) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
