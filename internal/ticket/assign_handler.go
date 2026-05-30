package ticket

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
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
