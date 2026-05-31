package channel

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/gateway"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
	"github.com/husky/husky/pkg/logger"
)

type QueryService interface {
	SaveWebhookMessage(ctx context.Context, msg *model.WebhookMessageRecord) error
	ListWebhookMessages(ctx context.Context, offset, limit int, channel string) ([]model.WebhookMessageRecord, int64, error)
	ListNotifications(ctx context.Context, userID uint, offset, limit int) ([]model.Notification, int64, error)
	GetUnreadCount(ctx context.Context, userID uint) (int64, error)
	MarkNotificationRead(ctx context.Context, id, userID uint) error
	MarkAllNotificationsRead(ctx context.Context, userID uint) error
}

type Handler struct {
	gw    *gateway.Gateway
	cfgSvc ConfigService
	querySvc QueryService
	log    *logger.Logger
}

func NewHandler(gw *gateway.Gateway, cfgSvc ConfigService, querySvc QueryService, log *logger.Logger) *Handler {
	return &Handler{gw: gw, cfgSvc: cfgSvc, querySvc: querySvc, log: log}
}

func (h *Handler) JoinGroup(c *gin.Context) {
	ticketIDStr := c.Query("ticket_id")
	userID := c.Query("user_id")

	if ticketIDStr == "" || userID == "" {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "missing ticket_id or user_id")
		return
	}

	h.log.Info("Join group request", "ticket_id", ticketIDStr, "user_id", userID)
	httputil.Success(c, gin.H{
		"success": true,
		"message": "joined ticket group",
	})
}

func (h *Handler) ListWebhookMessages(c *gin.Context) {
	channel := c.Query("channel")
	offset, limit, ok := httputil.ParsePagination(c)
	if !ok {
		return
	}

	list, total, err := h.querySvc.ListWebhookMessages(c.Request.Context(), offset, limit, channel)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"data": list, "total": total})
}

func (h *Handler) saveWebhookRecord(c *gin.Context, channel string, msg *gateway.IncomingMessage, ticketID *uint) {
	rawBytes, _ := json.Marshal(msg.Raw)
	record := &model.WebhookMessageRecord{
		Channel:    channel,
		MessageID:  msg.MessageID,
		UserID:     msg.UserID,
		Content:    msg.Content,
		TicketID:   ticketID,
		RawPayload: string(rawBytes),
		Status:     "received",
	}
	if ticketID != nil {
		record.Status = "processed"
	}
	if err := h.querySvc.SaveWebhookMessage(c.Request.Context(), record); err != nil {
		h.log.Warn("Failed to save webhook message record", "error", err)
	}
}
