package ticket

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/pkg/errors"
)

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
