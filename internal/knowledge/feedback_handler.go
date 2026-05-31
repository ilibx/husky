package knowledge

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

func (h *Handler) RecordFeedback(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "id is required")
		return
	}

	var req struct {
		Helpful bool   `json:"helpful"`
		Comment string `json:"comment,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		httputil.Error(c, http.StatusUnauthorized, errors.ErrUnauthorized, "user not authenticated")
		return
	}

	if err := h.knowledgeService.RecordFeedback(c.Request.Context(), id, uid, req.Helpful, req.Comment); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	httputil.Created(c, gin.H{"message": "feedback recorded"})
}

func (h *Handler) GetFeedbackStats(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "id is required")
		return
	}

	stats, err := h.knowledgeService.GetFeedbackStats(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, stats)
}

func parseUintParam(c *gin.Context, name string) (uint, error) {
	v, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}
