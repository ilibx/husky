package agent

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

// --- SOP CRUD ---

func (h *CRUDHandler) CreateSOP(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req model.CreateSOPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	sop, err := h.sopService.Create(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	httputil.Created(c, sop)
}

func (h *CRUDHandler) ListSOPs(c *gin.Context) {
	offset, limit, ok := httputil.ParsePagination(c)
	if !ok {
		return
	}

	list, total, err := h.sopService.List(c.Request.Context(), offset, limit)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"data": list, "total": total})
}

func (h *CRUDHandler) GetSOP(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid sop id")
		return
	}

	sop, err := h.sopService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, err.Error())
		return
	}

	httputil.Success(c, sop)
}

func (h *CRUDHandler) UpdateSOP(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid sop id")
		return
	}

	var req model.UpdateSOPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	sop, err := h.sopService.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	httputil.Success(c, sop)
}

func (h *CRUDHandler) DeleteSOP(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid sop id")
		return
	}

	if err := h.sopService.Delete(c.Request.Context(), uint(id)); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"message": "sop deleted"})
}
