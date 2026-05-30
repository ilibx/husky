package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/service"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/logger"
)

// StatsHandler 统计处理器
type StatsHandler struct {
	statsService service.StatsService
	log          *logger.Logger
}

// NewStatsHandler 创建统计处理器实例
func NewStatsHandler(statsService service.StatsService, log *logger.Logger) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
		log:          log,
	}
}

// Overview 获取统计概览
func (h *StatsHandler) Overview(c *gin.Context) {
	stats, err := h.statsService.GetOverview(c.Request.Context())
	if err != nil {
		h.log.Error("Failed to get stats overview", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, stats)
}

// Tickets 获取工单统计
func (h *StatsHandler) Tickets(c *gin.Context) {
	stats, err := h.statsService.GetTicketStats(c.Request.Context())
	if err != nil {
		h.log.Error("Failed to get ticket stats", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, stats)
}

// Performance 获取绩效统计
func (h *StatsHandler) Performance(c *gin.Context) {
	satisfaction, err := h.statsService.GetSatisfactionStats(c.Request.Context())
	if err != nil {
		h.log.Error("Failed to get satisfaction stats", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, satisfaction)
}

// AgentPerformance 获取客服绩效统计
func (h *StatsHandler) AgentPerformance(c *gin.Context) {
	stats, err := h.statsService.GetAgentPerformance(c.Request.Context())
	if err != nil {
		h.log.Error("Failed to get agent performance", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}
