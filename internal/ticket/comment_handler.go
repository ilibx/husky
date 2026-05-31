package ticket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

func (h *TicketHandler) AddComment(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	var req struct {
		Content    string `json:"content"`
		IsInternal bool   `json:"is_internal"`
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

	comment, err := h.ticketService.AddComment(c.Request.Context(), id, uid, req.Content, req.IsInternal)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Created(c, comment)
}

func (h *TicketHandler) GetComments(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	offset, limit, ok := httputil.ParsePagination(c)
	if !ok {
		return
	}

	comments, total, err := h.ticketService.ListComments(c.Request.Context(), id, offset, limit)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"data":  comments,
		"total": total,
	})
}
