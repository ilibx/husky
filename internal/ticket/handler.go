package ticket

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/logger"
)

type TicketGroupHandler interface {
	CreateGroupForTicket(ctx context.Context, ticket *model.Ticket) (*model.TicketGroup, error)
	OnGroupMessage(ctx context.Context, chatID, userID, content string) error
}

type AgentEngineHandler interface {
	OnTicketCreated(ctx context.Context, ticket *model.Ticket)
	OnTicketUpdated(ctx context.Context, ticket *model.Ticket)
}

type TicketHandler struct {
	ticketService   Service
	ticketGroupSvc  TicketGroupHandler
	agentEngine     AgentEngineHandler
	slaConfigSvc    *SLAConfigService
	log             *logger.Logger
}

func NewTicketHandler(ticketService Service, ticketGroupSvc TicketGroupHandler, agentEngine AgentEngineHandler, slaConfigSvc *SLAConfigService, log *logger.Logger) *TicketHandler {
	return &TicketHandler{
		ticketService:  ticketService,
		ticketGroupSvc: ticketGroupSvc,
		agentEngine:    agentEngine,
		slaConfigSvc:   slaConfigSvc,
		log:            log,
	}
}

func (h *TicketHandler) CreateTicket(c *gin.Context) {
	var req model.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	uid, _ := userID.(uint)

	// Admin/agent can create ticket on behalf of others
	if role == "admin" || role == "agent" {
		if req.RequesterID == "" {
			req.RequesterID = fmt.Sprintf("%d", uid)
		}
	} else {
		if req.RequesterID != "" && req.RequesterID != fmt.Sprintf("%d", uid) {
			c.JSON(http.StatusForbidden, errors.NewErrorResponse(errors.ErrForbidden, "cannot create ticket for another user"))
			return
		}
		req.RequesterID = fmt.Sprintf("%d", uid)
	}

	resp, err := h.ticketService.CreateTicket(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	go func(ticketID uint, requesterID uint) {
		ctx := context.Background()
		t, err := h.ticketService.GetTicket(ctx, ticketID)
		if err != nil {
			return
		}

		// Compute SLA due date
		if h.slaConfigSvc != nil {
			if dueAt := h.slaConfigSvc.ComputeDueAt(ctx, t); dueAt != nil {
				h.ticketService.SetDueAt(ctx, ticketID, *dueAt)
			}
		}

		if h.ticketGroupSvc != nil {
			if _, err := h.ticketGroupSvc.CreateGroupForTicket(ctx, t); err != nil {
				h.log.Error("Failed to create group", "error", err, "ticket_id", ticketID)
			}
		}

		if h.agentEngine != nil {
			h.agentEngine.OnTicketCreated(ctx, t)
		}
	}(resp.ID, resp.RequesterID)

	c.JSON(http.StatusCreated, resp)
}

func (h *TicketHandler) ListTickets(c *gin.Context) {
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "20")
	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)

	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if priority := c.Query("priority"); priority != "" {
		filters["priority"] = priority
	}
	if source := c.Query("source"); source != "" {
		filters["source"] = source
	}
	if assigneeID := c.Query("assignee_id"); assigneeID != "" {
		if id, err := strconv.ParseUint(assigneeID, 10, 64); err == nil {
			filters["assignee_id"] = id
		}
	}
	if keyword := c.Query("q"); keyword != "" {
		filters["keyword"] = keyword
	}

	tickets, total, err := h.ticketService.ListTickets(c.Request.Context(), offset, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  tickets,
		"total": total,
		"page":  offset/limit + 1,
	})
}

func (h *TicketHandler) GetTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	ticket, err := h.ticketService.GetTicket(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, errors.NewErrorResponse(errors.ErrNotFound, "ticket not found"))
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func (h *TicketHandler) UpdateTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	var req model.Ticket
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}
	req.ID = id

	if err := h.ticketService.UpdateTicket(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, req)
}

func (h *TicketHandler) DeleteTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	if err := h.ticketService.DeleteTicket(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TicketHandler) WatchTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}
	userID, _ := c.Get("user_id")
	if err := h.ticketService.WatchTicket(c.Request.Context(), id, userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "watching ticket"})
}

func (h *TicketHandler) UnwatchTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}
	userID, _ := c.Get("user_id")
	if err := h.ticketService.UnwatchTicket(c.Request.Context(), id, userID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "stopped watching ticket"})
}

func (h *TicketHandler) AutoAssignTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	agentID, err := h.ticketService.AutoAssignTicket(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ticket auto-assigned", "assignee_id": agentID})
}

func (h *TicketHandler) AssignTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	var req struct {
		AssigneeID uint `json:"assignee_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	if err := h.ticketService.AssignTicket(c.Request.Context(), id, req.AssigneeID); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ticket assigned"})
}

func (h *TicketHandler) ClaimTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, errors.NewErrorResponse(errors.ErrUnauthorized, "user not authenticated"))
		return
	}

	if err := h.ticketService.ClaimTicket(c.Request.Context(), id, uid); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ticket claimed"})
}

func (h *TicketHandler) UpdateStatus(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	if err := h.ticketService.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		h.log.Error("Failed to update ticket status", "error", err)
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidTransition, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func (h *TicketHandler) AddComment(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	var req struct {
		Content    string `json:"content"`
		IsInternal bool   `json:"is_internal"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, errors.NewErrorResponse(errors.ErrUnauthorized, "user not authenticated"))
		return
	}

	comment, err := h.ticketService.AddComment(c.Request.Context(), id, uid, req.Content, req.IsInternal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, comment)
}

func (h *TicketHandler) GetComments(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "20")
	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)

	comments, total, err := h.ticketService.ListComments(c.Request.Context(), id, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  comments,
		"total": total,
	})
}

func (h *TicketHandler) ListAuditLogs(c *gin.Context) {
	ticketID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	resourceType := c.DefaultQuery("resource_type", "")
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "20")
	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)

	logs, total, err := h.ticketService.ListAuditLogs(c.Request.Context(), resourceType, uint(ticketID), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
	})
}

func (h *TicketHandler) RateTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	var req struct {
		Score   int    `json:"score"`
		Comment string `json:"comment,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, errors.NewErrorResponse(errors.ErrUnauthorized, "user not authenticated"))
		return
	}

	sat, err := h.ticketService.RateTicket(c.Request.Context(), id, uid, req.Score, req.Comment)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, sat)
}

func (h *TicketHandler) GetSatisfaction(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	sat, err := h.ticketService.GetSatisfaction(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	if sat == nil {
		c.JSON(http.StatusNotFound, errors.NewErrorResponse(errors.ErrNotFound, "not rated yet"))
		return
	}

	c.JSON(http.StatusOK, sat)
}

func (h *TicketHandler) SetDueAt(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	var req struct {
		DueAt time.Time `json:"due_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	if err := h.ticketService.SetDueAt(c.Request.Context(), id, req.DueAt); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "due date set"})
}

func (h *TicketHandler) ListOverdue(c *gin.Context) {
	tickets, err := h.ticketService.ListOverdue(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tickets})
}

func (h *TicketHandler) UploadAttachment(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "file is required"))
		return
	}
	defer file.Close()

	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, errors.NewErrorResponse(errors.ErrUnauthorized, "user not authenticated"))
		return
	}

	att, err := h.ticketService.UploadAttachment(c.Request.Context(), id, uid, header.Filename, file)
	if err != nil {
		h.log.Error("Failed to upload attachment", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, att)
}

func (h *TicketHandler) ListAttachments(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	attachments, err := h.ticketService.ListAttachments(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": attachments})
}

func (h *TicketHandler) DeleteAttachment(c *gin.Context) {
	attachmentID, err := parseUintParam(c, "attachmentId")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid attachment id"))
		return
	}

	if err := h.ticketService.DeleteAttachment(c.Request.Context(), attachmentID); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}

func parseUintParam(c *gin.Context, name string) (uint, error) {
	v, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

// --- Tags ---

func (h *TicketHandler) CreateTag(c *gin.Context) {
	var req struct {
		Name  string `json:"name" binding:"required"`
		Color string `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	tag, err := h.ticketService.CreateTag(c.Request.Context(), req.Name, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, tag)
}

func (h *TicketHandler) ListTags(c *gin.Context) {
	tags, err := h.ticketService.ListTags(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tags})
}

func (h *TicketHandler) GetTag(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid tag id"))
		return
	}

	tag, err := h.ticketService.GetTag(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	if tag == nil {
		c.JSON(http.StatusNotFound, errors.NewErrorResponse(errors.ErrNotFound, "tag not found"))
		return
	}

	c.JSON(http.StatusOK, tag)
}

func (h *TicketHandler) UpdateTag(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid tag id"))
		return
	}

	var req struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	tag := &model.Tag{Name: req.Name, Color: req.Color}
	tag.ID = id

	if err := h.ticketService.UpdateTag(c.Request.Context(), tag); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, tag)
}

func (h *TicketHandler) DeleteTag(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid tag id"))
		return
	}

	if err := h.ticketService.DeleteTag(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TicketHandler) GetTicketTags(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	tags, err := h.ticketService.GetTicketTags(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tags})
}

func (h *TicketHandler) UpdateTicketTags(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	var req struct {
		TagIDs []uint `json:"tag_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	if err := h.ticketService.UpdateTicketTags(c.Request.Context(), id, req.TagIDs); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *TicketHandler) AddTicketTags(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	var req struct {
		TagIDs []uint `json:"tag_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	if err := h.ticketService.AddTagsToTicket(c.Request.Context(), id, req.TagIDs); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *TicketHandler) RemoveTicketTag(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	tagID, err := parseUintParam(c, "tagId")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid tag id"))
		return
	}

	if err := h.ticketService.RemoveTagFromTicket(c.Request.Context(), id, tagID); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}

// --- Assign Config ---

func (h *TicketHandler) GetAssignConfig(c *gin.Context) {
	categoryIDStr := c.Query("category_id")
	var categoryID *uint
	if categoryIDStr != "" {
		if id, err := strconv.ParseUint(categoryIDStr, 10, 64); err == nil {
			cid := uint(id)
			categoryID = &cid
		}
	}

	cfg, err := h.ticketService.GetAssignConfig(c.Request.Context(), categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, cfg)
}

func (h *TicketHandler) SetAssignConfig(c *gin.Context) {
	var req model.AssignConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	cfg := &model.AssignConfig{
		Strategy:   req.Strategy,
		CategoryID: req.CategoryID,
	}

	if err := h.ticketService.SetAssignConfig(c.Request.Context(), cfg); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, cfg)
}

// --- Ticket Relations ---

func (h *TicketHandler) CreateTicketRelation(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	var req struct {
		RelatedID    uint   `json:"related_id" binding:"required"`
		RelationType string `json:"relation_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	rel, err := h.ticketService.CreateTicketRelation(c.Request.Context(), id, req.RelatedID, req.RelationType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, rel)
}

func (h *TicketHandler) ListTicketRelations(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	rels, err := h.ticketService.ListTicketRelations(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rels})
}

func (h *TicketHandler) DeleteTicketRelation(c *gin.Context) {
	relID, err := parseUintParam(c, "relationId")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid relation id"))
		return
	}

	if err := h.ticketService.DeleteTicketRelation(c.Request.Context(), relID); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}

// --- Roles ---

func (h *TicketHandler) ListRoles(c *gin.Context) {
	roles, err := h.ticketService.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": roles})
}

func (h *TicketHandler) GetRole(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid role id"))
		return
	}

	role, err := h.ticketService.GetRole(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	if role == nil {
		c.JSON(http.StatusNotFound, errors.NewErrorResponse(errors.ErrNotFound, "role not found"))
		return
	}

	c.JSON(http.StatusOK, role)
}

func (h *TicketHandler) CreateRole(c *gin.Context) {
	var req model.Role
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	if err := h.ticketService.CreateRole(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, req)
}

func (h *TicketHandler) UpdateRole(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid role id"))
		return
	}

	var req model.Role
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	req.ID = id
	if err := h.ticketService.UpdateRole(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, req)
}

func (h *TicketHandler) DeleteRole(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid role id"))
		return
	}

	if err := h.ticketService.DeleteRole(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}

// --- Ticket Fields ---

func (h *TicketHandler) ListTicketFields(c *gin.Context) {
	fields, err := h.ticketService.ListTicketFields(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": fields})
}

func (h *TicketHandler) CreateTicketField(c *gin.Context) {
	var req model.CreateTicketFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	field, err := h.ticketService.CreateTicketField(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, field)
}

func (h *TicketHandler) UpdateTicketField(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid field id"))
		return
	}

	var req model.UpdateTicketFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	field, err := h.ticketService.UpdateTicketField(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, field)
}

func (h *TicketHandler) DeleteTicketField(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid field id"))
		return
	}

	if err := h.ticketService.DeleteTicketField(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TicketHandler) UpdateTicketFieldValues(c *gin.Context) {
	ticketID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	var req []model.TicketFieldValue
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	if err := h.ticketService.UpdateTicketFieldValues(c.Request.Context(), ticketID, req); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *TicketHandler) GetTicketFieldValues(c *gin.Context) {
	ticketID, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	values, err := h.ticketService.GetTicketFieldValues(c.Request.Context(), ticketID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": values})
}

// --- Bot Config ---

func (h *TicketHandler) GetBotConfig(c *gin.Context) {
	channel := c.Param("channel")
	cfg, err := h.ticketService.GetBotConfig(c.Request.Context(), channel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *TicketHandler) SetBotConfig(c *gin.Context) {
	var cfg model.BotConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}
	if err := h.ticketService.SetBotConfig(c.Request.Context(), &cfg); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// --- SLA Config ---

func (h *TicketHandler) ListSLAConfigs(c *gin.Context) {
	list, err := h.slaConfigSvc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *TicketHandler) GetSLAConfig(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid id"))
		return
	}
	cfg, err := h.slaConfigSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, errors.NewErrorResponse(errors.ErrNotFound, "SLA config not found"))
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *TicketHandler) CreateSLAConfig(c *gin.Context) {
	var req model.SLAConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}
	if err := h.slaConfigSvc.Create(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusCreated, req)
}

func (h *TicketHandler) UpdateSLAConfig(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid id"))
		return
	}
	var req model.SLAConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}
	req.ID = id
	if err := h.slaConfigSvc.Update(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, req)
}

func (h *TicketHandler) DeleteSLAConfig(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid id"))
		return
	}
	if err := h.slaConfigSvc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Webhook Config ---

func (h *TicketHandler) ListWebhookConfigs(c *gin.Context) {
	list, err := h.slaConfigSvc.webhookRepo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *TicketHandler) GetWebhookConfig(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid id"))
		return
	}
	cfg, err := h.slaConfigSvc.webhookRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, errors.NewErrorResponse(errors.ErrNotFound, "webhook config not found"))
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *TicketHandler) CreateWebhookConfig(c *gin.Context) {
	var req model.WebhookConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}
	if err := h.slaConfigSvc.webhookRepo.Create(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusCreated, req)
}

func (h *TicketHandler) UpdateWebhookConfig(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid id"))
		return
	}
	var req model.WebhookConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}
	req.ID = id
	if err := h.slaConfigSvc.webhookRepo.Update(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, req)
}

func (h *TicketHandler) DeleteWebhookConfig(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid id"))
		return
	}
	if err := h.slaConfigSvc.webhookRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}
