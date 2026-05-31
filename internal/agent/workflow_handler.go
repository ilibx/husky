package agent

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

// --- Workflow CRUD ---

func (h *CRUDHandler) ListWorkflows(c *gin.Context) {
	offset, limit, ok := httputil.ParsePagination(c)
	if !ok {
		return
	}

	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	list, total, err := h.workflowSvc.ListWorkflows(c.Request.Context(), offset, limit, filters)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, gin.H{"data": list, "total": total})
}

func (h *CRUDHandler) GetWorkflow(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}

	wf, err := h.workflowSvc.GetWorkflowByID(c.Request.Context(), uint(id))
	if err != nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, err.Error())
		return
	}
	httputil.Success(c, wf)
}

func (h *CRUDHandler) CompleteStep(c *gin.Context) {
	stepIDStr := c.Param("stepId")
	stepID, err := strconv.ParseUint(stepIDStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid step id")
		return
	}

	var req struct {
		Result string `json:"result"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	if err := h.workflowSvc.CompleteHumanStep(c.Request.Context(), uint(stepID), req.Result); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	httputil.Success(c, gin.H{"message": "step completed"})
}

func (h *CRUDHandler) PendingSteps(c *gin.Context) {
	userID, _ := c.Get("user_id")
	offset, limit, ok := httputil.ParsePagination(c)
	if !ok {
		return
	}

	steps, total, err := h.workflowSvc.ListPendingSteps(c.Request.Context(), userID.(uint), offset, limit)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, gin.H{"data": steps, "total": total})
}

// ApproveStep 审核通过人工步骤
func (h *CRUDHandler) ApproveStep(c *gin.Context) {
	stepIDStr := c.Param("stepId")
	stepID, err := strconv.ParseUint(stepIDStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid step id")
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := userID.(uint)

	var req struct {
		Feedback string `json:"feedback,omitempty"`
		Result   string `json:"result,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid request body")
		return
	}

	if err := h.workflowSvc.ApproveStep(c.Request.Context(), uint(stepID), uid, req.Feedback, req.Result); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	httputil.Success(c, gin.H{"message": "step approved"})
}

// RejectStep 拒绝人工步骤
func (h *CRUDHandler) RejectStep(c *gin.Context) {
	stepIDStr := c.Param("stepId")
	stepID, err := strconv.ParseUint(stepIDStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid step id")
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := userID.(uint)

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "reason is required")
		return
	}

	if err := h.workflowSvc.RejectStep(c.Request.Context(), uint(stepID), uid, req.Reason); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	httputil.Success(c, gin.H{"message": "step rejected", "reason": req.Reason})
}

// ReviseStep 要求修改（打回重做）
func (h *CRUDHandler) ReviseStep(c *gin.Context) {
	stepIDStr := c.Param("stepId")
	stepID, err := strconv.ParseUint(stepIDStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid step id")
		return
	}

	userID, _ := c.Get("user_id")
	uid, _ := userID.(uint)

	var req struct {
		Feedback string `json:"feedback" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "feedback is required")
		return
	}

	if err := h.workflowSvc.ReviseStep(c.Request.Context(), uint(stepID), uid, req.Feedback); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	httputil.Success(c, gin.H{"message": "step sent back for revision"})
}
