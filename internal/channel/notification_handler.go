package channel

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

func (h *Handler) ListNotifications(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		httputil.Error(c, http.StatusUnauthorized, errors.ErrUnauthorized, "user not authenticated")
		return
	}

	offset, limit, ok := httputil.ParsePagination(c)
	if !ok {
		return
	}

	role, _ := c.Get("role")
	isAdmin := role == "admin"
	listUserID := uid
	if isAdmin && c.Query("scope") != "mine" {
		listUserID = 0
	}

	list, total, err := h.querySvc.ListNotifications(c.Request.Context(), listUserID, offset, limit, c.Query("keyword"))
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"data":  list,
		"total": total,
	})
}

func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		httputil.Error(c, http.StatusUnauthorized, errors.ErrUnauthorized, "user not authenticated")
		return
	}

	role, _ := c.Get("role")
	isAdmin := role == "admin"
	countUserID := uid
	if isAdmin && c.Query("scope") != "mine" {
		countUserID = 0
	}

	count, err := h.querySvc.GetUnreadCount(c.Request.Context(), countUserID)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"count": count})
}

func (h *Handler) MarkRead(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		httputil.Error(c, http.StatusUnauthorized, errors.ErrUnauthorized, "user not authenticated")
		return
	}

	idStr := c.Param("id")
	if idStr == "all" {
		if err := h.querySvc.MarkAllNotificationsRead(c.Request.Context(), uid); err != nil {
			httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
			return
		}
		httputil.Success(c, gin.H{"message": "all marked as read"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid notification id")
		return
	}

	if err := h.querySvc.MarkNotificationRead(c.Request.Context(), uint(id), uid); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"message": "marked as read"})
}
