package channel

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/gateway"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
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
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "missing ticket_id or user_id"))
		return
	}

	h.log.Info("Join group request", "ticket_id", ticketIDStr, "user_id", userID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "joined ticket group",
	})
}

func (h *Handler) LarkWebhook(c *gin.Context) {
	rawBody, _ := io.ReadAll(c.Request.Body)

	var event model.LarkWebhookEvent
	if err := json.Unmarshal(rawBody, &event); err != nil {
		h.log.Error("Failed to parse Lark webhook", "error", err)
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid request body"))
		return
	}

	if event.Type == "url_verification" && event.Challenge != "" {
		h.log.Info("Lark URL verification challenge")
		c.JSON(http.StatusOK, model.LarkCallbackResponse{Challenge: event.Challenge})
		return
	}

	if event.Header == nil || event.Header.EventType == "" {
		h.log.Warn("Received Lark webhook without header")
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "missing event header"))
		return
	}

	if event.Header.EventType != "im.message.receive_v1" {
		h.log.Debug("Ignoring non-message event", "type", event.Header.EventType)
		c.JSON(http.StatusOK, model.LarkCallbackResponse{Success: true})
		return
	}

	if event.Event == nil {
		h.log.Warn("Lark event has no data")
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "missing event data"))
		return
	}

	content := event.Event.Content
	if content == "" && event.Event.Text.Content != "" {
		content = event.Event.Text.Content
	}
	cleanContent := cleanLarkContent(content)

	userCtx := &gateway.UserContext{
		ChannelUserID: event.Event.UserID,
		OpenID:        event.Event.OpenID,
		UnionID:       event.Event.UnionID,
	}

	msg := &gateway.IncomingMessage{
		Channel:   gateway.ChannelLark,
		MessageID: event.Header.EventID,
		UserID:    event.Event.UserID,
		Content:   cleanContent,
		ChatID:    event.Event.ChatID,
		Raw:       event,
		User:      userCtx,
	}

	if event.Event.ChatID != "" {
		if err := h.gw.HandleGroupMessage(c.Request.Context(), msg); err != nil {
			h.log.Error("Failed to handle group message", "error", err)
		}
		h.saveWebhookRecord(c, "lark", msg, nil)
		c.JSON(http.StatusOK, model.LarkCallbackResponse{Success: true})
		return
	}

	ticket, err := h.gw.HandleIncoming(c.Request.Context(), msg)
	if err != nil {
		h.log.Error("Failed to create ticket from Lark message", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, "failed to create ticket"))
		return
	}

	h.saveWebhookRecord(c, "lark", msg, &ticket.ID)

	c.JSON(http.StatusOK, model.LarkCallbackResponse{
		Success: true,
		Message: "Ticket created: " + strconv.FormatUint(uint64(ticket.ID), 10),
	})
}

func (h *Handler) DingTalkWebhook(c *gin.Context) {
	rawBody, _ := io.ReadAll(c.Request.Body)

	var event model.DingTalkWebhookEvent
	if err := json.Unmarshal(rawBody, &event); err != nil {
		h.log.Error("Failed to parse DingTalk webhook", "error", err)
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid request body"))
		return
	}

	h.log.Info("Received DingTalk webhook", "msg_type", event.MsgType, "sender", event.SenderNick)

	if event.MsgType != "text" || event.Text == nil || strings.TrimSpace(event.Text.Content) == "" {
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	content := strings.TrimSpace(event.Text.Content)
	msg := &gateway.IncomingMessage{
		Channel:   gateway.ChannelDingTalk,
		MessageID: event.ConversationID,
		UserID:    event.SenderID,
		Content:   content,
		Raw:       event,
		User: &gateway.UserContext{
			ChannelUserID: event.SenderID,
			DisplayName:   event.SenderNick,
		},
	}

	ticket, err := h.gw.HandleIncoming(c.Request.Context(), msg)
	if err != nil {
		h.log.Error("Failed to create ticket from DingTalk message", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, "failed to create ticket"))
		return
	}

	h.saveWebhookRecord(c, "dingtalk", msg, &ticket.ID)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"ticket_id": ticket.ID,
		"message":   "Ticket created",
	})
}

func (h *Handler) WeComWebhook(c *gin.Context) {
	rawBody, _ := io.ReadAll(c.Request.Body)

	var event model.WeComWebhookEvent
	if err := json.Unmarshal(rawBody, &event); err != nil {
		h.log.Error("Failed to parse WeCom webhook", "error", err)
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid request body"))
		return
	}

	h.log.Info("Received WeCom webhook", "msg_type", event.MsgType, "from", event.FromUserName)

	if event.MsgType != "text" || event.Text == nil || strings.TrimSpace(event.Text.Content) == "" {
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	content := strings.TrimSpace(event.Text.Content)
	msg := &gateway.IncomingMessage{
		Channel:   gateway.ChannelWeCom,
		MessageID: event.MsgID,
		UserID:    event.FromUserName,
		Content:   content,
		Raw:       event,
		User: &gateway.UserContext{
			ChannelUserID: event.FromUserName,
		},
	}

	ticket, err := h.gw.HandleIncoming(c.Request.Context(), msg)
	if err != nil {
		h.log.Error("Failed to create ticket from WeCom message", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, "failed to create ticket"))
		return
	}

	h.saveWebhookRecord(c, "wecom", msg, &ticket.ID)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"ticket_id": ticket.ID,
		"message":   "Ticket created",
	})
}

func (h *Handler) ListWebhookMessages(c *gin.Context) {
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "20")
	channel := c.Query("channel")
	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)

	list, total, err := h.querySvc.ListWebhookMessages(c.Request.Context(), offset, limit, channel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

func (h *Handler) CreateChannelConfig(c *gin.Context) {
	var req model.CreateChannelConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	cfg, err := h.cfgSvc.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, cfg)
}

func (h *Handler) ListChannelConfigs(c *gin.Context) {
	list, err := h.cfgSvc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *Handler) GetChannelConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid channel config id"))
		return
	}

	cfg, err := h.cfgSvc.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, errors.NewErrorResponse(errors.ErrNotFound, err.Error()))
		return
	}

	c.JSON(http.StatusOK, cfg)
}

func (h *Handler) UpdateChannelConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid channel config id"))
		return
	}

	var req model.UpdateChannelConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	cfg, err := h.cfgSvc.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	c.JSON(http.StatusOK, cfg)
}

func (h *Handler) DeleteChannelConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid channel config id"))
		return
	}

	if err := h.cfgSvc.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "channel config deleted"})
}

func (h *Handler) ListNotifications(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, errors.NewErrorResponse(errors.ErrUnauthorized, "user not authenticated"))
		return
	}

	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "20")
	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)

	list, total, err := h.querySvc.ListNotifications(c.Request.Context(), uid, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  list,
		"total": total,
	})
}

func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, errors.NewErrorResponse(errors.ErrUnauthorized, "user not authenticated"))
		return
	}

	count, err := h.querySvc.GetUnreadCount(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
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

func (h *Handler) MarkRead(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, errors.NewErrorResponse(errors.ErrUnauthorized, "user not authenticated"))
		return
	}

	idStr := c.Param("id")
	if idStr == "all" {
		if err := h.querySvc.MarkAllNotificationsRead(c.Request.Context(), uid); err != nil {
			c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "all marked as read"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid notification id"))
		return
	}

	if err := h.querySvc.MarkNotificationRead(c.Request.Context(), uint(id), uid); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}
