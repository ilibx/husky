package ticket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/pkg/errors"
)

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
