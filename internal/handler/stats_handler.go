package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/service"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
	"github.com/husky/husky/pkg/logger"
)

type StatsHandler struct {
	statsService service.StatsService
	log          *logger.Logger
}

func NewStatsHandler(statsService service.StatsService, log *logger.Logger) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
		log:          log,
	}
}

func (h *StatsHandler) Overview(c *gin.Context) {
	stats, err := h.statsService.GetOverview(c.Request.Context())
	if err != nil {
		h.log.Error("Failed to get stats overview", "error", err)
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, stats)
}

func (h *StatsHandler) Tickets(c *gin.Context) {
	stats, err := h.statsService.GetTicketStats(c.Request.Context())
	if err != nil {
		h.log.Error("Failed to get ticket stats", "error", err)
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, stats)
}

func (h *StatsHandler) Performance(c *gin.Context) {
	satisfaction, err := h.statsService.GetSatisfactionStats(c.Request.Context())
	if err != nil {
		h.log.Error("Failed to get satisfaction stats", "error", err)
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, satisfaction)
}

func (h *StatsHandler) AgentPerformance(c *gin.Context) {
	stats, err := h.statsService.GetAgentPerformance(c.Request.Context())
	if err != nil {
		h.log.Error("Failed to get agent performance", "error", err)
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, gin.H{"data": stats})
}
