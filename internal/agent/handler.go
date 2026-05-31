package agent

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
	"github.com/husky/husky/pkg/logger"
)

type CRUDHandler struct {
	agentService     Service
	sopService       SOPService
	workflowSvc      *WorkflowService
	log              *logger.Logger
}

func NewCRUDHandler(agentService Service, sopService SOPService, workflowSvc *WorkflowService, log *logger.Logger) *CRUDHandler {
	return &CRUDHandler{
		agentService: agentService,
		sopService:   sopService,
		workflowSvc:  workflowSvc,
		log:          log,
	}
}

// --- Agent CRUD ---

func (h *CRUDHandler) CreateAgent(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req model.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	agent, err := h.agentService.Create(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	httputil.Created(c, agent)
}

func (h *CRUDHandler) ListAgents(c *gin.Context) {
	offset, limit, ok := httputil.ParsePagination(c)
	if !ok {
		return
	}

	list, total, err := h.agentService.List(c.Request.Context(), offset, limit)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"data": list, "total": total})
}

func (h *CRUDHandler) GetAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid agent id")
		return
	}

	agent, err := h.agentService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, err.Error())
		return
	}

	httputil.Success(c, agent)
}

func (h *CRUDHandler) UpdateAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid agent id")
		return
	}

	var req model.UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	agent, err := h.agentService.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	httputil.Success(c, agent)
}

func (h *CRUDHandler) DeleteAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid agent id")
		return
	}

	if err := h.agentService.Delete(c.Request.Context(), uint(id)); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"message": "agent deleted"})
}

