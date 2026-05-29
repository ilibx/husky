package agent

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
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
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	agent, err := h.agentService.Create(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, agent)
}

func (h *CRUDHandler) ListAgents(c *gin.Context) {
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "20")
	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)

	list, total, err := h.agentService.List(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

func (h *CRUDHandler) GetAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid agent id"))
		return
	}

	agent, err := h.agentService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, errors.NewErrorResponse(errors.ErrNotFound, err.Error()))
		return
	}

	c.JSON(http.StatusOK, agent)
}

func (h *CRUDHandler) UpdateAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid agent id"))
		return
	}

	var req model.UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	agent, err := h.agentService.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	c.JSON(http.StatusOK, agent)
}

func (h *CRUDHandler) DeleteAgent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid agent id"))
		return
	}

	if err := h.agentService.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "agent deleted"})
}

// --- SOP CRUD ---

func (h *CRUDHandler) CreateSOP(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req model.CreateSOPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	sop, err := h.sopService.Create(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, sop)
}

func (h *CRUDHandler) ListSOPs(c *gin.Context) {
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "20")
	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)

	list, total, err := h.sopService.List(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

func (h *CRUDHandler) GetSOP(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid sop id"))
		return
	}

	sop, err := h.sopService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, errors.NewErrorResponse(errors.ErrNotFound, err.Error()))
		return
	}

	c.JSON(http.StatusOK, sop)
}

func (h *CRUDHandler) UpdateSOP(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid sop id"))
		return
	}

	var req model.UpdateSOPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	sop, err := h.sopService.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	c.JSON(http.StatusOK, sop)
}

func (h *CRUDHandler) DeleteSOP(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid sop id"))
		return
	}

	if err := h.sopService.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "sop deleted"})
}

// --- Workflow CRUD ---

func (h *CRUDHandler) ListWorkflows(c *gin.Context) {
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "20")
	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)

	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	list, total, err := h.workflowSvc.ListWorkflows(c.Request.Context(), offset, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

func (h *CRUDHandler) GetWorkflow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid id"))
		return
	}

	wf, err := h.workflowSvc.GetWorkflowByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, errors.NewErrorResponse(errors.ErrNotFound, err.Error()))
		return
	}
	c.JSON(http.StatusOK, wf)
}

func (h *CRUDHandler) CompleteStep(c *gin.Context) {
	stepIDStr := c.Param("stepId")
	stepID, err := strconv.ParseUint(stepIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid step id"))
		return
	}

	var req struct {
		Result string `json:"result"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	if err := h.workflowSvc.CompleteHumanStep(c.Request.Context(), uint(stepID), req.Result); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "step completed"})
}

func (h *CRUDHandler) PendingSteps(c *gin.Context) {
	userID, _ := c.Get("user_id")
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "20")
	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)

	steps, total, err := h.workflowSvc.ListPendingSteps(c.Request.Context(), userID.(uint), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": steps, "total": total})
}
