package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/config"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
	"github.com/husky/husky/pkg/logger"
)

type SystemConfigHandler struct {
	repo   *repository.SystemConfigRepository
	dynCfg *config.DynamicConfig
	log    *logger.Logger
}

func NewSystemConfigHandler(repo *repository.SystemConfigRepository, dynCfg *config.DynamicConfig, log *logger.Logger) *SystemConfigHandler {
	return &SystemConfigHandler{repo: repo, dynCfg: dynCfg, log: log}
}

func (h *SystemConfigHandler) ListSystemConfigs(c *gin.Context) {
	offset, limit, ok := httputil.ParsePagination(c)
	if !ok {
		return
	}

	configs, total, err := h.repo.ListAll(c.Request.Context(), offset, limit)
	if err != nil {
		h.log.Error("Failed to list system configs", "error", err)
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, "failed to list configs")
		return
	}

	var resp []model.SystemConfigResponse
	for _, cfg := range configs {
		resp = append(resp, toSysCfgResponse(cfg))
	}

	httputil.Success(c, gin.H{
		"data":  resp,
		"total": total,
	})
}

func (h *SystemConfigHandler) GetSystemConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}

	cfg, err := h.repo.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, "config not found")
		return
	}

	httputil.Success(c, toSysCfgResponse(*cfg))
}

func (h *SystemConfigHandler) GetSystemConfigByKey(c *gin.Context) {
	category := c.Query("category")
	key := c.Query("key")
	if category == "" || key == "" {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "category and key are required")
		return
	}

	cfg, err := h.repo.GetByKey(c.Request.Context(), category, key)
	if err != nil {
		httputil.Success(c, nil)
		return
	}

	httputil.Success(c, toSysCfgResponse(*cfg))
}

func (h *SystemConfigHandler) GetLLMConfig(c *gin.Context) {
	llmCfg, err := h.repo.GetLLMConfig(c.Request.Context())
	if err != nil {
		def := model.DefaultLLMConfig()
		llmCfg = &def
	}
	httputil.Success(c, llmCfg)
}

func (h *SystemConfigHandler) GetVectorConfig(c *gin.Context) {
	vCfg, err := h.repo.GetVectorConfig(c.Request.Context())
	if err != nil {
		def := model.DefaultVectorDBConfig()
		vCfg = &def
	}
	httputil.Success(c, vCfg)
}

func (h *SystemConfigHandler) CreateSystemConfig(c *gin.Context) {
	var req model.SystemConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	cfg := &model.SystemConfig{
		Category: req.Category,
		Key:      req.Key,
		Value:    req.Value,
		Enabled:  enabled,
	}

	if err := h.repo.Create(c.Request.Context(), cfg); err != nil {
		h.log.Error("Failed to create system config", "error", err)
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, "failed to create config")
		return
	}

	h.tryReloadLLM(c, cfg.Category)

	httputil.Created(c, toSysCfgResponse(*cfg))
}

func (h *SystemConfigHandler) UpdateSystemConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}

	var req model.SystemConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	cfg, err := h.repo.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, "config not found")
		return
	}

	cfg.Key = req.Key
	cfg.Value = req.Value
	if req.Enabled != nil {
		cfg.Enabled = *req.Enabled
	}

	if err := h.repo.Update(c.Request.Context(), cfg); err != nil {
		h.log.Error("Failed to update system config", "error", err)
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, "failed to update config")
		return
	}

	h.tryReloadLLM(c, cfg.Category)

	httputil.Success(c, toSysCfgResponse(*cfg))
}

func (h *SystemConfigHandler) UpsertSystemConfig(c *gin.Context) {
	var req model.SystemConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	cfg := &model.SystemConfig{
		Category: req.Category,
		Key:      req.Key,
		Value:    req.Value,
		Enabled:  enabled,
	}

	if err := h.repo.Upsert(c.Request.Context(), cfg); err != nil {
		h.log.Error("Failed to upsert system config", "error", err)
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, "failed to save config")
		return
	}

	h.tryReloadLLM(c, cfg.Category)

	httputil.Success(c, toSysCfgResponse(*cfg))
}

func (h *SystemConfigHandler) DeleteSystemConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}

	cfg, err := h.repo.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, "config not found")
		return
	}

	if err := h.repo.Delete(c.Request.Context(), uint(id)); err != nil {
		h.log.Error("Failed to delete system config", "error", err)
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, "failed to delete config")
		return
	}

	h.tryReloadLLM(c, cfg.Category)

	httputil.Success(c, gin.H{"message": "config deleted"})
}

func (h *SystemConfigHandler) tryReloadLLM(c *gin.Context, category string) {
	if category == model.SysCfgCategoryLLM && h.dynCfg != nil {
		if err := h.dynCfg.ReloadLLM(c.Request.Context()); err != nil {
			h.log.Warn("Failed to reload LLM config after update", "error", err)
		} else {
			h.log.Info("LLM config reloaded successfully")
		}
	}
	if category == model.SysCfgCategoryVector && h.dynCfg != nil {
		if err := h.dynCfg.ReloadVector(c.Request.Context()); err != nil {
			h.log.Warn("Failed to reload vector config after update", "error", err)
		} else {
			h.log.Info("Vector config reloaded successfully")
		}
	}
}

func toSysCfgResponse(cfg model.SystemConfig) model.SystemConfigResponse {
	return model.SystemConfigResponse{
		ID:        cfg.ID,
		Category:  cfg.Category,
		Key:       cfg.Key,
		Value:     cfg.Value,
		Enabled:   cfg.Enabled,
		CreatedAt: cfg.CreatedAt,
		UpdatedAt: cfg.UpdatedAt,
	}
}
