package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/service"
	"github.com/husky/husky/pkg/errors"
)

// DepartmentHandler 部门处理器
type DepartmentHandler struct {
	deptService service.DepartmentService
}

func NewDepartmentHandler(deptService service.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{deptService: deptService}
}

// CreateDepartment 创建部门
func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var req model.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	dept, err := h.deptService.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dept)
}

// ListDepartments 获取部门列表
func (h *DepartmentHandler) ListDepartments(c *gin.Context) {
	list, err := h.deptService.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": list})
}

// GetDepartment 获取部门详情
func (h *DepartmentHandler) GetDepartment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid department id"))
		return
	}

	dept, err := h.deptService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, errors.NewErrorResponse(errors.ErrNotFound, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dept)
}

// UpdateDepartment 更新部门
func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid department id"))
		return
	}

	var req model.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	dept, err := h.deptService.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dept)
}

// DeleteDepartment 删除部门
func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid department id"))
		return
	}

	if err := h.deptService.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "department deleted"})
}
