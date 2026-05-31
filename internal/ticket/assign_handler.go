package ticket

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

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
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, cfg)
}

func (h *TicketHandler) SetAssignConfig(c *gin.Context) {
	var req model.AssignConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	cfg := &model.AssignConfig{
		Strategy:   req.Strategy,
		CategoryID: req.CategoryID,
	}

	if err := h.ticketService.SetAssignConfig(c.Request.Context(), cfg); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, cfg)
}
