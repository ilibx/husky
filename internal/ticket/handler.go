package ticket

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
	"github.com/husky/husky/pkg/logger"
)

type TicketHandler struct {
	ticketService  Service
	slaConfigSvc   *SLAConfigService
	log            *logger.Logger
}

func NewTicketHandler(ticketService Service, slaConfigSvc *SLAConfigService, log *logger.Logger) *TicketHandler {
	return &TicketHandler{
		ticketService: ticketService,
		slaConfigSvc:  slaConfigSvc,
		log:           log,
	}
}

func (h *TicketHandler) CreateTicket(c *gin.Context) {
	var req model.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
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
			httputil.Error(c, http.StatusForbidden, errors.ErrForbidden, "cannot create ticket for another user")
			return
		}
		req.RequesterID = fmt.Sprintf("%d", uid)
	}

	resp, err := h.ticketService.CreateTicket(c.Request.Context(), &req)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Created(c, resp)
}

func (h *TicketHandler) ListTickets(c *gin.Context) {
	offset, limit, ok := httputil.ParsePagination(c)
	if !ok {
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

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

	// scope=my → only assigned to current user
	// scope=my_team → current user + subordinates by role level
	// scope=all or empty → all tickets (admin default)
	if scope := c.Query("scope"); scope != "" {
		filters["scope"] = scope
		if uid, ok := userID.(uint); ok {
			filters["scope_user_id"] = uid
		}
		if r, ok := role.(string); ok {
			filters["scope_role"] = r
		}
	}

	tickets, total, err := h.ticketService.ListTickets(c.Request.Context(), offset, limit, filters)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"data":  tickets,
		"total": total,
		"page":  offset/limit + 1,
	})
}

func (h *TicketHandler) GetTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	ticket, err := h.ticketService.GetTicket(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, "ticket not found")
		return
	}

	httputil.Success(c, ticket)
}

func (h *TicketHandler) UpdateTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	var req struct {
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
		Status      string `json:"status,omitempty"`
		Priority    string `json:"priority,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	ticket := &model.Ticket{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
	}
	ticket.ID = id
	if err := h.ticketService.UpdateTicket(c.Request.Context(), ticket); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, ticket)
}

func (h *TicketHandler) DeleteTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	if err := h.ticketService.DeleteTicket(c.Request.Context(), id); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TicketHandler) WatchTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}
	userID, _ := c.Get("user_id")
	if err := h.ticketService.WatchTicket(c.Request.Context(), id, userID.(uint)); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, gin.H{"message": "watching ticket"})
}

func (h *TicketHandler) UnwatchTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}
	userID, _ := c.Get("user_id")
	if err := h.ticketService.UnwatchTicket(c.Request.Context(), id, userID.(uint)); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, gin.H{"message": "stopped watching ticket"})
}

func (h *TicketHandler) AutoAssignTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	agentID, err := h.ticketService.AutoAssignTicket(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"message": "ticket auto-assigned", "assignee_id": agentID})
}

func (h *TicketHandler) AssignTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	var req struct {
		AssigneeID uint `json:"assignee_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	if err := h.ticketService.AssignTicket(c.Request.Context(), id, req.AssigneeID); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"message": "ticket assigned"})
}

func (h *TicketHandler) ClaimTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		httputil.Error(c, http.StatusUnauthorized, errors.ErrUnauthorized, "user not authenticated")
		return
	}

	if err := h.ticketService.ClaimTicket(c.Request.Context(), id, uid); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"message": "ticket claimed"})
}

func (h *TicketHandler) ListAuditLogs(c *gin.Context) {
	ticketID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	resourceType := c.DefaultQuery("resource_type", "")
	offset, limit, ok := httputil.ParsePagination(c)
	if !ok {
		return
	}

	logs, total, err := h.ticketService.ListAuditLogs(c.Request.Context(), resourceType, uint(ticketID), offset, limit)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"data":  logs,
		"total": total,
	})
}

func parseUintParam(c *gin.Context, name string) (uint, error) {
	v, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

