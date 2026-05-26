package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/service"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/logger"
)

// WebhookHandler Webhook 处理器
type WebhookHandler struct {
	ticketService service.TicketService
	log           *logger.Logger
}

// NewWebhookHandler 创建 Webhook 处理器实例
func NewWebhookHandler(ticketService service.TicketService, log *logger.Logger) *WebhookHandler {
	return &WebhookHandler{
		ticketService: ticketService,
		log:           log,
	}
}

// LarkWebhook 飞书 Webhook 回调
// @Summary 飞书 Webhook
// @Description 接收飞书机器人消息，自动创建工单
// @Tags webhook
// @Accept json
// @Produce json
// @Param event body model.LarkWebhookEvent true "飞书事件"
// @Success 200 {object} model.LarkCallbackResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /webhook/lark [post]
func (h *WebhookHandler) LarkWebhook(c *gin.Context) {
	var event model.LarkWebhookEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		h.log.Error("Failed to parse Lark webhook", "error", err)
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid request body"))
		return
	}

	// URL 验证挑战
	if event.Type == "url_verification" && event.Challenge != "" {
		h.log.Info("Lark URL verification challenge")
		c.JSON(http.StatusOK, model.LarkCallbackResponse{
			Challenge: event.Challenge,
		})
		return
	}

	// 处理事件回调
	if event.Header == nil || event.Header.EventType == "" {
		h.log.Warn("Received Lark webhook without header or event type")
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "missing event header"))
		return
	}

	h.log.Info("Received Lark event", "type", event.Header.EventType, "event_id", event.Header.EventID)

	// 目前只处理 IM 消息事件
	if event.Header.EventType != "im.message.receive_v1" {
		h.log.Debug("Ignoring non-message event", "type", event.Header.EventType)
		c.JSON(http.StatusOK, model.LarkCallbackResponse{Success: true})
		return
	}

	// 解析消息内容
	if event.Event == nil {
		h.log.Warn("Lark event has no data")
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "missing event data"))
		return
	}

	// 提取消息内容
	content := event.Event.Content
	if content == "" && event.Event.Text.Content != "" {
		content = event.Event.Text.Content
	}

	if content == "" {
		h.log.Warn("Lark message has no content")
		c.JSON(http.StatusOK, model.LarkCallbackResponse{Success: true})
		return
	}

	// 清理富文本格式（飞书消息可能包含标签）
	cleanContent := cleanLarkContent(content)

	// 创建工单请求
	ticketReq := &model.CreateTicketRequest{
		Title:       "飞书消息工单 - " + truncateString(cleanContent, 50),
		Description: cleanContent,
		Priority:    "medium", // 默认中优先级
		RequesterID: event.Event.UserID,
		Channel:     string(model.ChannelLark),
		Metadata: map[string]interface{}{
			"lark_message_id": event.Event.MessageID,
			"lark_tenant_key": event.Event.TenantKey,
			"raw_content":     content,
		},
	}

	// 创建工单
	ticket, err := h.ticketService.CreateTicket(c.Request.Context(), ticketReq)
	if err != nil {
		h.log.Error("Failed to create ticket from Lark message", "error", err, "message_id", event.Event.MessageID)
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, "failed to create ticket"))
		return
	}

	h.log.Info("Successfully created ticket from Lark message", 
		"ticket_id", ticket.ID, 
		"message_id", event.Event.MessageID)

	// 返回成功响应
	c.JSON(http.StatusOK, model.LarkCallbackResponse{
		Success: true,
		Message: "工单已创建：" + ticket.ID,
	})
}

// DingTalkWebhook 钉钉 Webhook 回调
// @Summary 钉钉 Webhook
// @Description 接收钉钉机器人消息，自动创建工单
// @Tags webhook
// @Accept json
// @Produce json
// @Param event body model.DingTalkWebhookEvent true "钉钉事件"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /webhook/dingtalk [post]
func (h *WebhookHandler) DingTalkWebhook(c *gin.Context) {
	var event model.DingTalkWebhookEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		h.log.Error("Failed to parse DingTalk webhook", "error", err)
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid request body"))
		return
	}

	h.log.Info("Received DingTalk webhook", "msg_type", event.MsgType, "sender", event.SenderNick)

	// 只处理文本消息
	if event.MsgType != "text" || event.Text == nil {
		h.log.Debug("Ignoring non-text DingTalk message")
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	content := strings.TrimSpace(event.Text.Content)
	if content == "" {
		h.log.Warn("DingTalk message has no content")
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	// 创建工单请求
	ticketReq := &model.CreateTicketRequest{
		Title:       "钉钉消息工单 - " + truncateString(content, 50),
		Description: content,
		Priority:    "medium",
		RequesterID: event.SenderID,
		Channel:     string(model.ChannelDingTalk),
		Metadata: map[string]interface{}{
			"dingtalk_conversation_id": event.ConversationID,
			"dingtalk_robot_code":      event.RobotCode,
			"raw_content":              content,
		},
	}

	// 创建工单
	ticket, err := h.ticketService.CreateTicket(c.Request.Context(), ticketReq)
	if err != nil {
		h.log.Error("Failed to create ticket from DingTalk message", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, "failed to create ticket"))
		return
	}

	h.log.Info("Successfully created ticket from DingTalk message", 
		"ticket_id", ticket.ID, 
		"sender", event.SenderNick)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"ticket_id": ticket.ID,
		"message":   "工单已创建",
	})
}

// WeComWebhook 企业微信 Webhook 回调
// @Summary 企业微信 Webhook
// @Description 接收企业微信应用消息，自动创建工单
// @Tags webhook
// @Accept json
// @Produce json
// @Param event body model.WeComWebhookEvent true "企业微信事件"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} errors.ErrorResponse
// @Failure 500 {object} errors.ErrorResponse
// @Router /webhook/wecom [post]
func (h *WebhookHandler) WeComWebhook(c *gin.Context) {
	var event model.WeComWebhookEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		h.log.Error("Failed to parse WeCom webhook", "error", err)
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid request body"))
		return
	}

	h.log.Info("Received WeCom webhook", "msg_type", event.MsgType, "from", event.FromUserName)

	// 只处理文本消息
	if event.MsgType != "text" || event.Text == nil {
		h.log.Debug("Ignoring non-text WeCom message")
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	content := strings.TrimSpace(event.Text.Content)
	if content == "" {
		h.log.Warn("WeCom message has no content")
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	// 创建工单请求
	ticketReq := &model.CreateTicketRequest{
		Title:       "企微消息工单 - " + truncateString(content, 50),
		Description: content,
		Priority:    "medium",
		RequesterID: event.FromUserName,
		Channel:     string(model.ChannelWeCom),
		Metadata: map[string]interface{}{
			"wecom_msg_id":      event.MsgID,
			"wecom_agent_id":    event.AgentID,
			"enterprise_corp_id": event.EnterpriseCorpID,
			"raw_content":       content,
		},
	}

	// 创建工单
	ticket, err := h.ticketService.CreateTicket(c.Request.Context(), ticketReq)
	if err != nil {
		h.log.Error("Failed to create ticket from WeCom message", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, "failed to create ticket"))
		return
	}

	h.log.Info("Successfully created ticket from WeCom message", 
		"ticket_id", ticket.ID, 
		"from", event.FromUserName)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"ticket_id": ticket.ID,
		"message":   "工单已创建",
	})
}

// cleanLarkContent 清理飞书富文本内容
func cleanLarkContent(content string) string {
	// 简单的清理逻辑，移除常见的富文本标签
	// 实际生产环境可能需要更复杂的解析
	
	// 移除 <at> 标签
	result := content
	result = strings.ReplaceAll(result, "<at>", "")
	result = strings.ReplaceAll(result, "</at>", "")
	
	// 移除 <a> 标签但保留文本
	// 这里使用简单替换，实际应该用正则或 HTML 解析器
	result = strings.ReplaceAll(result, "<a>", "")
	result = strings.ReplaceAll(result, "</a>", "")
	
	// 移除其他常见标签
	result = strings.ReplaceAll(result, "<b>", "")
	result = strings.ReplaceAll(result, "</b>", "")
	result = strings.ReplaceAll(result, "<i>", "")
	result = strings.ReplaceAll(result, "</i>", "")
	result = strings.ReplaceAll(result, "<br/>", "\n")
	result = strings.ReplaceAll(result, "<br>", "\n")
	
	return strings.TrimSpace(result)
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return s
}

// ParseWebhookMessage 通用 Webhook 消息解析器
func ParseWebhookMessage(channel model.ChannelType, rawJSON []byte) (*model.WebhookMessage, error) {
	msg := &model.WebhookMessage{
		Channel: channel,
	}

	switch channel {
	case model.ChannelLark:
		var event model.LarkWebhookEvent
		if err := json.Unmarshal(rawJSON, &event); err != nil {
			return nil, err
		}
		if event.Event != nil {
			msg.UserID = event.Event.UserID
			msg.MessageID = event.Event.MessageID
			content := event.Event.Content
			if content == "" && event.Event.Text.Content != "" {
				content = event.Event.Text.Content
			}
			msg.Content = cleanLarkContent(content)
		}
		msg.RawPayload = event

	case model.ChannelDingTalk:
		var event model.DingTalkWebhookEvent
		if err := json.Unmarshal(rawJSON, &event); err != nil {
			return nil, err
		}
		msg.UserID = event.SenderID
		msg.MessageID = event.ConversationID // 钉钉没有单独的消息 ID，使用会话 ID
		if event.Text != nil {
			msg.Content = strings.TrimSpace(event.Text.Content)
		}
		msg.RawPayload = event

	case model.ChannelWeCom:
		var event model.WeComWebhookEvent
		if err := json.Unmarshal(rawJSON, &event); err != nil {
			return nil, err
		}
		msg.UserID = event.FromUserName
		msg.MessageID = event.MsgID
		if event.Text != nil {
			msg.Content = strings.TrimSpace(event.Text.Content)
		}
		msg.RawPayload = event

	default:
		return nil, errors.New(errors.ErrInvalidParams, "unsupported channel type")
	}

	return msg, nil
}
