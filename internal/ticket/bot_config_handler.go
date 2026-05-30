package ticket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
)

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
