package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

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

type TestConnectionRequest struct {
	Type    string `json:"type"`
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

func modelsURL(baseURL, typ string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	hasV1 := strings.HasSuffix(baseURL, "/v1")
	switch typ {
	case "azure":
		return baseURL + "/openai/deployments?api-version=2024-08-01-preview"
	case "gemini":
		return baseURL + "/v1/models"
	case "claude":
		if hasV1 {
			return baseURL + "/models"
		}
		return baseURL + "/v1/models"
	default:
		if hasV1 {
			return baseURL + "/models"
		}
		return baseURL + "/v1/models"
	}
}

func (h *SystemConfigHandler) TestConnection(c *gin.Context) {
	var req TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}
	if req.BaseURL == "" || req.APIKey == "" {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "base_url and api_key are required")
		return
	}

	client := http.Client{Timeout: 15 * time.Second}
	var httpReq *http.Request
	var err error

	url := modelsURL(req.BaseURL, req.Type)
	if req.Type == "gemini" {
		url += "?key=" + req.APIKey
		httpReq, err = http.NewRequest("GET", url, nil)
	} else if req.Type == "azure" {
		httpReq, err = http.NewRequest("GET", url, nil)
		if err == nil {
			httpReq.Header.Set("api-key", req.APIKey)
		}
	} else {
		httpReq, err = http.NewRequest("GET", url, nil)
		if err == nil {
			httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)
			if req.Type == "claude" {
				httpReq.Header.Set("anthropic-version", "2023-06-01")
			}
		}
	}

	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, fmt.Sprintf("invalid request: %v", err))
		return
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		httputil.Error(c, http.StatusBadGateway, errors.ErrInternal, fmt.Sprintf("连接失败: %v", err))
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		httputil.Success(c, gin.H{"message": "连接成功"})
	} else {
		body, _ := io.ReadAll(resp.Body)
		httputil.Error(c, http.StatusBadGateway, errors.ErrInternal, fmt.Sprintf("连接成功但认证失败 (%d): %s", resp.StatusCode, string(body)))
	}
}

func (h *SystemConfigHandler) FetchProviderModels(c *gin.Context) {
	var req TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}
	if req.BaseURL == "" || req.APIKey == "" {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "base_url and api_key are required")
		return
	}

	client := http.Client{Timeout: 15 * time.Second}
	var httpReq *http.Request
	var err error

	url := modelsURL(req.BaseURL, req.Type)
	if req.Type == "gemini" {
		url += "?key=" + req.APIKey
		httpReq, err = http.NewRequest("GET", url, nil)
	} else if req.Type == "azure" {
		httpReq, err = http.NewRequest("GET", url, nil)
		if err == nil {
			httpReq.Header.Set("api-key", req.APIKey)
		}
	} else {
		httpReq, err = http.NewRequest("GET", url, nil)
		if err == nil {
			httpReq.Header.Set("Authorization", "Bearer "+req.APIKey)
			if req.Type == "claude" {
				httpReq.Header.Set("anthropic-version", "2023-06-01")
			}
		}
	}

	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, fmt.Sprintf("invalid request: %v", err))
		return
	}

	h.log.Info("fetching models", "url", url, "type", req.Type)

	resp, err := client.Do(httpReq)
	if err != nil {
		httputil.Error(c, http.StatusBadGateway, errors.ErrInternal, fmt.Sprintf("获取模型列表失败: %v", err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, "failed to read response")
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		httputil.Error(c, http.StatusBadGateway, errors.ErrInternal, fmt.Sprintf("API 返回错误 (%d): %s", resp.StatusCode, string(body)))
		return
	}

	h.log.Info("models API response", "body", string(body))

	type modelInfo struct {
		ID      string `json:"id"`
		OwnedBy string `json:"owned_by"`
	}

	var models []modelInfo

	switch req.Type {
	case "gemini":
		var result struct {
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}
		if err := json.Unmarshal(body, &result); err == nil {
			for _, m := range result.Models {
				models = append(models, modelInfo{ID: strings.TrimPrefix(m.Name, "models/"), OwnedBy: "gemini"})
			}
		} else {
			h.log.Error("failed to parse gemini response", "error", err)
		}
	case "azure":
		var result struct {
			Value []struct {
				ID string `json:"id"`
			} `json:"value"`
		}
		if err := json.Unmarshal(body, &result); err == nil {
			for _, d := range result.Value {
				models = append(models, modelInfo{ID: d.ID, OwnedBy: "azure"})
			}
		} else {
			h.log.Error("failed to parse azure response", "error", err)
		}
	default:
		var result struct {
			Data []struct {
				ID      string `json:"id"`
				OwnedBy string `json:"owned_by"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &result); err == nil {
			for _, d := range result.Data {
				models = append(models, modelInfo{ID: d.ID, OwnedBy: d.OwnedBy})
			}
		} else {
			h.log.Error("failed to parse response", "error", err, "body", string(body))
		}
	}

	if models == nil {
		models = []modelInfo{}
	}

	httputil.Success(c, gin.H{"models": models})
}

func (h *SystemConfigHandler) GetModelCatalog(c *gin.Context) {
	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get("https://models.dev/api.json")
	if err != nil {
		httputil.Error(c, http.StatusBadGateway, errors.ErrInternal, "failed to fetch model catalog")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		httputil.Error(c, http.StatusBadGateway, errors.ErrInternal, "failed to fetch model catalog")
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		httputil.Error(c, http.StatusBadGateway, errors.ErrInternal, "failed to read model catalog")
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
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

func (h *SystemConfigHandler) GetKnowledgeStores(c *gin.Context) {
	stores, err := h.repo.ListKnowledgeStoreConfigs(c.Request.Context())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	if stores == nil {
		stores = []model.KnowledgeStoreConfig{}
	}
	httputil.Success(c, gin.H{"data": stores})
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
