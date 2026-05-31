package ticket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

// --- SLA Config ---

func (h *TicketHandler) ListSLAConfigs(c *gin.Context) {
	offset, limit, ok := httputil.ParsePagination(c)
	if !ok {
		return
	}
	keyword := c.Query("keyword")
	list, total, err := h.slaConfigSvc.List(c.Request.Context(), offset, limit, keyword)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, gin.H{"data": list, "total": total})
}

func (h *TicketHandler) GetSLAConfig(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}
	cfg, err := h.slaConfigSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, "SLA config not found")
		return
	}
	httputil.Success(c, cfg)
}

func (h *TicketHandler) CreateSLAConfig(c *gin.Context) {
	var req model.SLAConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}
	if err := h.slaConfigSvc.Create(c.Request.Context(), &req); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Created(c, req)
}

func (h *TicketHandler) UpdateSLAConfig(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}
	var req model.SLAConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}
	req.ID = id
	if err := h.slaConfigSvc.Update(c.Request.Context(), &req); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, req)
}

func (h *TicketHandler) DeleteSLAConfig(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}
	if err := h.slaConfigSvc.Delete(c.Request.Context(), id); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Webhook Config ---

func (h *TicketHandler) ListWebhookConfigs(c *gin.Context) {
	list, err := h.slaConfigSvc.webhookRepo.List(c.Request.Context())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, gin.H{"data": list})
}

func (h *TicketHandler) GetWebhookConfig(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}
	cfg, err := h.slaConfigSvc.webhookRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, "webhook config not found")
		return
	}
	httputil.Success(c, cfg)
}

func (h *TicketHandler) CreateWebhookConfig(c *gin.Context) {
	var req model.WebhookConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}
	if err := h.slaConfigSvc.webhookRepo.Create(c.Request.Context(), &req); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Created(c, req)
}

func (h *TicketHandler) UpdateWebhookConfig(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}
	var req model.WebhookConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}
	req.ID = id
	if err := h.slaConfigSvc.webhookRepo.Update(c.Request.Context(), &req); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, req)
}

func (h *TicketHandler) DeleteWebhookConfig(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}
	if err := h.slaConfigSvc.webhookRepo.Delete(c.Request.Context(), id); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

