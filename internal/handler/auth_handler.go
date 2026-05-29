package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/service"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/logger"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService service.AuthService
	log         *logger.Logger
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(authService service.AuthService, log *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		log:         log,
	}
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	token, user, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errors.NewErrorResponse(errors.ErrUnauthorized, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"email":    user.Email,
			"username": user.Username,
			"role":     user.Role,
			"avatar":   user.Avatar,
		},
	})
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
	})
}

// Refresh 刷新 Token
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, err.Error()))
		return
	}

	newToken, err := h.authService.RefreshToken(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, errors.NewErrorResponse(errors.ErrUnauthorized, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": newToken})
}

// Me 获取当前用户信息
func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, errors.NewErrorResponse(errors.ErrUnauthorized, "user not authenticated"))
		return
	}

	user, err := h.authService.GetCurrentUser(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, errors.NewErrorResponse(errors.ErrNotFound, "user not found"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"role":       user.Role,
		"avatar":     user.Avatar,
		"department": user.Department,
		"title":      user.Title,
	})
}
