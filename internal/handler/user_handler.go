package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/ldap"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/service"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
	"github.com/husky/husky/pkg/logger"
)

type UserHandler struct {
	userService service.UserService
	ldapSvc     *ldap.Service
	log         *logger.Logger
}

func NewUserHandler(userService service.UserService, ldapSvc *ldap.Service, log *logger.Logger) *UserHandler {
	return &UserHandler{userService: userService, ldapSvc: ldapSvc, log: log}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	user, err := h.userService.Create(c.Request.Context(), &req)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	httputil.Created(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
	})
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	offset, limit, ok := httputil.ParsePagination(c)
	if !ok {
		return
	}

	keyword := c.Query("keyword")
	users, total, err := h.userService.List(c.Request.Context(), offset, limit, keyword)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

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

	httputil.Success(c, gin.H{"data": list, "total": total})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid user id")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, err.Error())
		return
	}

	httputil.Success(c, gin.H{
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

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid user id")
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	user, err := h.userService.Update(c.Request.Context(), uint(id), &req)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"avatar":     user.Avatar,
		"department": user.Department,
		"title":      user.Title,
	})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	currentUserID, _ := c.Get("user_id")
	uid, _ := currentUserID.(uint)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid user id")
		return
	}

	if uint(id) == uid {
		httputil.Error(c, http.StatusForbidden, errors.ErrForbidden, "cannot delete yourself")
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, err.Error())
		return
	}
	if user.Role == "admin" || user.Username == "admin" {
		httputil.Error(c, http.StatusForbidden, errors.ErrForbidden, "admin user cannot be deleted")
		return
	}

	if err := h.userService.Delete(c.Request.Context(), uint(id)); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"message": "user deleted"})
}

func (h *UserHandler) ChangeRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid user id")
		return
	}

	var req model.ChangeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	if err := h.userService.ChangeRole(c.Request.Context(), uint(id), req.Role); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	httputil.Success(c, gin.H{"message": "role updated"})
}

func (h *UserHandler) SyncLDAPUsers(c *gin.Context) {
	if h.ldapSvc == nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "LDAP not configured")
		return
	}

	result, err := h.ldapSvc.SyncUsers(c.Request.Context())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, result)
}
