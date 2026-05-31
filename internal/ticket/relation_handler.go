package ticket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

// --- Ticket Relations ---

func (h *TicketHandler) CreateTicketRelation(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	var req struct {
		RelatedID    uint   `json:"related_id" binding:"required"`
		RelationType string `json:"relation_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	rel, err := h.ticketService.CreateTicketRelation(c.Request.Context(), id, req.RelatedID, req.RelationType)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Created(c, rel)
}

func (h *TicketHandler) ListTicketRelations(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	rels, err := h.ticketService.ListTicketRelations(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"data": rels})
}

func (h *TicketHandler) DeleteTicketRelation(c *gin.Context) {
	relID, err := parseUintParam(c, "relationId")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid relation id")
		return
	}

	if err := h.ticketService.DeleteTicketRelation(c.Request.Context(), relID); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
