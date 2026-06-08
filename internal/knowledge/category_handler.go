package knowledge

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
		Question     string   `json:"question" binding:"required"`
		Model        string   `json:"model,omitempty"`
		DeepThinking bool     `json:"deep_thinking,omitempty"`
		Images       []string `json:"images,omitempty"`
		Tool         string   `json:"tool,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "question is required")
		return
	}

	resp, err := h.knowledgeService.Ask(c.Request.Context(), req.Question, req.Model, req.DeepThinking, req.Images, req.Tool)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, resp)
}

func (h *Handler) AskStream(c *gin.Context) {
	var req struct {
		Question     string   `json:"question" binding:"required"`
		Model        string   `json:"model,omitempty"`
		DeepThinking bool     `json:"deep_thinking,omitempty"`
		Images       []string `json:"images,omitempty"`
		Tool         string   `json:"tool,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "question is required")
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, "streaming not supported")
		return
	}

	resp, err := h.knowledgeService.AskStream(c.Request.Context(), req.Question, req.Model, req.DeepThinking, req.Images, req.Tool, func(chunk string) error {
		for _, part := range strings.Split(chunk, "\n") {
			_, err := fmt.Fprintf(c.Writer, "data: %s\n", part)
			if err != nil {
				return err
			}
		}
		_, err := fmt.Fprintf(c.Writer, "\n")
		if err != nil {
			return err
		}
		flusher.Flush()
		return nil
	})
	if err != nil {
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		return
	}

	// Send sources as final event
	if len(resp.Sources) > 0 {
		srcJson, _ := json.Marshal(resp.Sources)
		fmt.Fprintf(c.Writer, "event: sources\ndata: %s\n\n", string(srcJson))
		flusher.Flush()
	}

	fmt.Fprintf(c.Writer, "event: done\ndata: [DONE]\n\n")
	flusher.Flush()
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
