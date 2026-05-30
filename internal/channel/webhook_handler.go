package channel

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/gateway"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
)

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

	if ticket == nil {
		c.JSON(http.StatusOK, model.LarkCallbackResponse{Success: true})
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

	if ticket == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "bot_reply": true})
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

	if ticket == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "bot_reply": true})
		return
	}

	h.saveWebhookRecord(c, "wecom", msg, &ticket.ID)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"ticket_id": ticket.ID,
		"message":   "Ticket created",
	})
}
