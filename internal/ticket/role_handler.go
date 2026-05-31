package ticket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

// --- Roles ---

func (h *TicketHandler) ListRoles(c *gin.Context) {
	roles, err := h.ticketService.ListRoles(c.Request.Context())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, gin.H{"data": roles})
}

func (h *TicketHandler) GetRole(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid role id")
		return
	}

	role, err := h.ticketService.GetRole(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	if role == nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, "role not found")
		return
	}

	httputil.Success(c, role)
}

func (h *TicketHandler) CreateRole(c *gin.Context) {
	var req model.Role
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	if err := h.ticketService.CreateRole(c.Request.Context(), &req); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Created(c, req)
}

func (h *TicketHandler) UpdateRole(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid role id")
		return
	}

	var req model.Role
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	req.ID = id
	if err := h.ticketService.UpdateRole(c.Request.Context(), &req); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, req)
}

func (h *TicketHandler) DeleteRole(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid role id")
		return
	}

	if err := h.ticketService.DeleteRole(c.Request.Context(), id); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

