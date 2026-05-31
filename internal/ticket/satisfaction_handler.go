package ticket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

func (h *TicketHandler) RateTicket(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	var req struct {
		Score   int    `json:"score"`
		Comment string `json:"comment,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		httputil.Error(c, http.StatusUnauthorized, errors.ErrUnauthorized, "user not authenticated")
		return
	}

	sat, err := h.ticketService.RateTicket(c.Request.Context(), id, uid, req.Score, req.Comment)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	httputil.Created(c, sat)
}

func (h *TicketHandler) GetSatisfaction(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	sat, err := h.ticketService.GetSatisfaction(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	if sat == nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, "not rated yet")
		return
	}

	httputil.Success(c, sat)
}
