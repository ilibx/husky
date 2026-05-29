package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/service"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/logger"
)

// UserHandler 用户管理处理器
type UserHandler struct {
	userService service.UserService
	log         *logger.Logger
}

// NewUserHandler 创建用户管理处理器
func NewUserHandler(userService service.UserService, log *logger.Logger) *UserHandler {
	return &UserHandler{userService: userService, log: log}
}

// ListUsers 获取用户列表（仅管理员）
func (h *UserHandler) ListUsers(c *gin.Context) {
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "20")
	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)

	users, total, err := h.userService.List(c.Request.Context(), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	// 不返回密码
	type userVO struct {
		ID         uint   `json:"id"`
		Email      string `json:"email"`
		Username   string `json:"username"`
		Avatar     string `json:"avatar"`
		Department string `json:"department"`
		Title      string `json:"title"`
		Role       string `json:"role"`
		Status     int    `json:"status"`
		Phone      string `json:"phone"`
	}
	list := make([]userVO, 0, len(users))
	for _, u := range users {
		list = append(list, userVO{
			ID:         u.ID,
			Email:      u.Email,
			Username:   u.Username,
			Avatar:     u.Avatar,
			Department: u.Department,
			Title:      u.Title,
			Role:       u.Role,
			Status:     u.Status,
			Phone:      u.Phone,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": list, "total": total})
}

// GetUser 获取用户详情
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid user id"))
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, errors.NewErrorResponse(errors.ErrNotFound, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"avatar":     user.Avatar,
		"department": user.Department,
		"title":      user.Title,
		"role":       user.Role,
		"status":     user.Status,
		"phone":      user.Phone,
		"created_at": user.CreatedAt,
		"last_login": user.LastLoginAt,
	})
}

// UpdateUser 更新用户资料
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid user id"))
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	user, err := h.userService.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"avatar":     user.Avatar,
		"department": user.Department,
		"title":      user.Title,
	})
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	currentUserID, _ := c.Get("user_id")
	uid, _ := currentUserID.(uint)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid user id"))
		return
	}

	if uint(id) == uid {
		c.JSON(http.StatusForbidden, errors.NewErrorResponse(errors.ErrForbidden, "cannot delete yourself"))
		return
	}

	if err := h.userService.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

// ChangeRole 变更用户角色（仅管理员）
func (h *UserHandler) ChangeRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid user id"))
		return
	}

	var req model.ChangeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	if err := h.userService.ChangeRole(c.Request.Context(), uint(id), req.Role); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}
