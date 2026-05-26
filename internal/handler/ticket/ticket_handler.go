package ticket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"husky/internal/middleware/auth"
	"husky/pkg/errors/herr"
	"husky/pkg/validator"
)

// Handler 工单 HTTP 处理器
type Handler interface {
	// CreateTicket 创建工单
	CreateTicket(c *gin.Context)
	// GetTicket 获取工单详情
	GetTicket(c *gin.Context)
	// ListTickets 查询工单列表
	ListTickets(c *gin.Context)
	// UpdateTicket 更新工单
	UpdateTicket(c *gin.Context)
	// DeleteTicket 删除工单
	DeleteTicket(c *gin.Context)
	// AssignTicket 分配工单
	AssignTicket(c *gin.Context)
	// UpdateStatus 更新工单状态
	UpdateStatus(c *gin.Context)
	// AddComment 添加工单评论
	AddComment(c *gin.Context)
	// ListComments 获取工单评论列表
	ListComments(c *gin.Context)
}

type handler struct {
	service Service
}

// NewHandler 创建工单处理器实例
func NewHandler(service Service) Handler {
	return &handler{service: service}
}

// CreateTicket godoc
// @Summary 创建工单
// @Description 创建一个新的工单
// @Tags tickets
// @Accept json
// @Produce json
// @Param request body CreateTicketRequest true "创建工单请求"
// @Success 201 {object} model.Ticket
// @Failure 400 {object} herr.Error
// @Failure 500 {object} herr.Error
// @Router /api/v1/tickets [post]
func (h *handler) CreateTicket(c *gin.Context) {
	var req CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		herr.ResponseError(c, http.StatusBadRequest, herr.ErrInvalidParameter, err.Error())
		return
	}

	if err := validator.Validate(&req); err != nil {
		herr.ResponseError(c, http.StatusBadRequest, herr.ErrInvalidParameter, err.Error())
		return
	}

	// 从认证上下文获取用户 ID
	userID := auth.GetUserID(c)
	if userID == "" {
		herr.ResponseError(c, http.StatusUnauthorized, herr.ErrUnauthorized, "user not authenticated")
		return
	}
	req.SubmitterID = userID

	ticket, err := h.service.CreateTicket(c.Request.Context(), &req)
	if err != nil {
		herr.ResponseError(c, http.StatusInternalServerError, herr.ErrInternal, "failed to create ticket")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": ticket})
}

// GetTicket godoc
// @Summary 获取工单详情
// @Description 根据 ID 获取工单详细信息
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "工单 ID"
// @Success 200 {object} model.Ticket
// @Failure 400 {object} herr.Error
// @Failure 404 {object} herr.Error
// @Router /api/v1/tickets/{id} [get]
func (h *handler) GetTicket(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		herr.ResponseError(c, http.StatusBadRequest, herr.ErrInvalidParameter, "ticket id is required")
		return
	}

	ticket, err := h.service.GetTicket(c.Request.Context(), id)
	if err != nil {
		if herr.IsErrNotFound(err) {
			herr.ResponseError(c, http.StatusNotFound, herr.ErrNotFound, "ticket not found")
			return
		}
		herr.ResponseError(c, http.StatusInternalServerError, herr.ErrInternal, "failed to get ticket")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ticket})
}

// ListTickets godoc
// @Summary 查询工单列表
// @Description 分页查询工单列表，支持多种筛选条件
// @Tags tickets
// @Accept json
// @Produce json
// @Param status query string false "状态 (逗号分隔)"
// @Param priority query string false "优先级 (逗号分隔)"
// @Param assignee_id query string false "处理人 ID"
// @Param submitter_id query string false "提交人 ID"
// @Param category_id query string false "分类 ID"
// @Param keyword query string false "关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param sort_by query string false "排序字段"
// @Param sort_order query string false "排序顺序 (asc/desc)" default(desc)
// @Success 200 {object} ListTicketsResponse
// @Failure 400 {object} herr.Error
// @Router /api/v1/tickets [get]
func (h *handler) ListTickets(c *gin.Context) {
	var req ListTicketsRequest

	// 解析查询参数
	statusStr := c.Query("status")
	if statusStr != "" {
		// TODO: 解析状态字符串
	}

	priorityStr := c.Query("priority")
	if priorityStr != "" {
		// TODO: 解析优先级字符串
	}

	req.AssigneeID = c.Query("assignee_id")
	req.SubmitterID = c.Query("submitter_id")
	req.CategoryID = c.Query("category_id")
	req.Keyword = c.Query("keyword")

	// 分页参数
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		// TODO: 解析页码
	}
	if ps := c.Query("page_size"); ps != "" {
		// TODO: 解析每页数量
	}
	req.Page = page
	req.PageSize = pageSize

	req.SortBy = c.Query("sort_by")
	req.SortOrder = c.Query("sort_order")
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	result, err := h.service.ListTickets(c.Request.Context(), &req)
	if err != nil {
		herr.ResponseError(c, http.StatusInternalServerError, herr.ErrInternal, "failed to list tickets")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// UpdateTicket godoc
// @Summary 更新工单
// @Description 更新工单信息
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "工单 ID"
// @Param request body UpdateTicketRequest true "更新工单请求"
// @Success 200 {object} model.Ticket
// @Failure 400 {object} herr.Error
// @Failure 404 {object} herr.Error
// @Router /api/v1/tickets/{id} [put]
func (h *handler) UpdateTicket(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		herr.ResponseError(c, http.StatusBadRequest, herr.ErrInvalidParameter, "ticket id is required")
		return
	}

	var req UpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		herr.ResponseError(c, http.StatusBadRequest, herr.ErrInvalidParameter, err.Error())
		return
	}

	if err := validator.Validate(&req); err != nil {
		herr.ResponseError(c, http.StatusBadRequest, herr.ErrInvalidParameter, err.Error())
		return
	}

	ticket, err := h.service.UpdateTicket(c.Request.Context(), id, &req)
	if err != nil {
		if herr.IsErrNotFound(err) {
			herr.ResponseError(c, http.StatusNotFound, herr.ErrNotFound, "ticket not found")
			return
		}
		herr.ResponseError(c, http.StatusInternalServerError, herr.ErrInternal, "failed to update ticket")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ticket})
}

// DeleteTicket godoc
// @Summary 删除工单
// @Description 软删除工单
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "工单 ID"
// @Success 204
// @Failure 400 {object} herr.Error
// @Failure 404 {object} herr.Error
// @Router /api/v1/tickets/{id} [delete]
func (h *handler) DeleteTicket(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		herr.ResponseError(c, http.StatusBadRequest, herr.ErrInvalidParameter, "ticket id is required")
		return
	}

	if err := h.service.DeleteTicket(c.Request.Context(), id); err != nil {
		if herr.IsErrNotFound(err) {
			herr.ResponseError(c, http.StatusNotFound, herr.ErrNotFound, "ticket not found")
			return
		}
		herr.ResponseError(c, http.StatusInternalServerError, herr.ErrInternal, "failed to delete ticket")
		return
	}

	c.Status(http.StatusNoContent)
}

// AssignTicket godoc
// @Summary 分配工单
// @Description 将工单分配给指定处理人
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "工单 ID"
// @Param request body AssignTicketRequest true "分配工单请求"
// @Success 200 {object} model.Ticket
// @Failure 400 {object} herr.Error
// @Failure 404 {object} herr.Error
// @Router /api/v1/tickets/{id}/assign [post]
func (h *handler) AssignTicket(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		herr.ResponseError(c, http.StatusBadRequest, herr.ErrInvalidParameter, "ticket id is required")
		return
	}

	var req struct {
		AssigneeID string `json:"assignee_id" validate:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		herr.ResponseError(c, http.StatusBadRequest, herr.ErrInvalidParameter, err.Error())
		return
	}

	if err := validator.Validate(&req); err != nil {
		herr.ResponseError(c, http.StatusBadRequest, herr.ErrInvalidParameter, err.Error())
		return
	}

	operatorID := auth.GetUserID(c)
	ticket, err := h.service.AssignTicket(c.Request.Context(), id, req.AssigneeID, operatorID)
	if err != nil {
		if herr.IsErrNotFound(err) {
			herr.ResponseError(c, http.StatusNotFound, herr.ErrNotFound, "ticket not found")
			return
		}
		herr.ResponseError(c, http.StatusInternalServerError, herr.ErrInternal, "failed to assign ticket")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ticket})
}

// UpdateStatus godoc
// @Summary 更新工单状态
// @Description 更新工单状态，可附带评论
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "工单 ID"
// @Param request body UpdateStatusRequest true "更新状态请求"
// @Success 200 {object} model.Ticket
// @Failure 400 {object} herr.Error
// @Failure 404 {object} herr.Error
// @Router /api/v1/tickets/{id}/status [put]
func (h *handler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		herr.ResponseError(c, http.StatusBadRequest, herr.ErrInvalidParameter, "ticket id is required")
		return
	}

	var req struct {
		Status  string `json:"status" validate:"required,oneof=open pending in_progress resolved closed"`
		Comment string `json:"comment,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		herr.ResponseError(c, http.StatusBadRequest, herr.ErrInvalidParameter, err.Error())
		return
	}

	if err := validator.Validate(&req); err != nil {
		herr.ResponseError(c, http.StatusBadRequest, herr.ErrInvalidParameter, err.Error())
		return
	}

	// TODO: 转换状态字符串为枚举
	operatorID := auth.GetUserID(c)
	ticket, err := h.service.UpdateStatus(c.Request.Context(), id, "open", operatorID, req.Comment)
	if err != nil {
		if herr.IsErrNotFound(err) {
			herr.ResponseError(c, http.StatusNotFound, herr.ErrNotFound, "ticket not found")
			return
		}
		herr.ResponseError(c, http.StatusInternalServerError, herr.ErrInternal, "failed to update status")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ticket})
}

// AddComment godoc
// @Summary 添加工单评论
// @Description 为工单添加评论
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "工单 ID"
// @Param request body AddCommentRequest true "添加评论请求"
// @Success 201 {object} model.Comment
// @Failure 400 {object} herr.Error
// @Failure 404 {object} herr.Error
// @Router /api/v1/tickets/{id}/comments [post]
func (h *handler) AddComment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		herr.ResponseError(c, http.StatusBadRequest, herr.ErrInvalidParameter, "ticket id is required")
		return
	}

	var req AddCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		herr.ResponseError(c, http.StatusBadRequest, herr.ErrInvalidParameter, err.Error())
		return
	}

	if err := validator.Validate(&req); err != nil {
		herr.ResponseError(c, http.StatusBadRequest, herr.ErrInvalidParameter, err.Error())
		return
	}

	req.AuthorID = auth.GetUserID(c)
	comment, err := h.service.AddComment(c.Request.Context(), id, &req)
	if err != nil {
		if herr.IsErrNotFound(err) {
			herr.ResponseError(c, http.StatusNotFound, herr.ErrNotFound, "ticket not found")
			return
		}
		herr.ResponseError(c, http.StatusInternalServerError, herr.ErrInternal, "failed to add comment")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": comment})
}

// ListComments godoc
// @Summary 获取工单评论列表
// @Description 获取工单的所有评论
// @Tags tickets
// @Accept json
// @Produce json
// @Param id path string true "工单 ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} ListCommentsResponse
// @Failure 400 {object} herr.Error
// @Router /api/v1/tickets/{id}/comments [get]
func (h *handler) ListComments(c *gin.Context) {
	// TODO: 实现评论列表
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"comments": []interface{}{}, "total": 0}})
}

// AssignTicketRequest 分配工单请求
type AssignTicketRequest struct {
	AssigneeID string `json:"assignee_id" validate:"required"`
}

// UpdateStatusRequest 更新状态请求
type UpdateStatusRequest struct {
	Status  string `json:"status" validate:"required"`
	Comment string `json:"comment,omitempty"`
}

// ListCommentsResponse 评论列表响应
type ListCommentsResponse struct {
	Comments []*interface{} `json:"comments"`
	Total    int64          `json:"total"`
}
