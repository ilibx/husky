package ticket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

// --- Ticket Fields ---

func (h *TicketHandler) ListTicketFields(c *gin.Context) {
	fields, err := h.ticketService.ListTicketFields(c.Request.Context())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, gin.H{"data": fields})
}

func (h *TicketHandler) CreateTicketField(c *gin.Context) {
	var req model.CreateTicketFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	field, err := h.ticketService.CreateTicketField(c.Request.Context(), &req)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Created(c, field)
}

func (h *TicketHandler) UpdateTicketField(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid field id")
		return
	}

	var req model.UpdateTicketFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	field, err := h.ticketService.UpdateTicketField(c.Request.Context(), id, &req)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, field)
}

func (h *TicketHandler) DeleteTicketField(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid field id")
		return
	}

	if err := h.ticketService.DeleteTicketField(c.Request.Context(), id); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TicketHandler) UpdateTicketFieldValues(c *gin.Context) {
	ticketID, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	var req []model.TicketFieldValue
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	if err := h.ticketService.UpdateTicketFieldValues(c.Request.Context(), ticketID, req); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"success": true})
}

func (h *TicketHandler) GetTicketFieldValues(c *gin.Context) {
	ticketID, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	values, err := h.ticketService.GetTicketFieldValues(c.Request.Context(), ticketID)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"data": values})
}

