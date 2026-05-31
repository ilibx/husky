package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/service"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
	"github.com/husky/husky/pkg/logger"
)

type AuthHandler struct {
	authService service.AuthService
	log         *logger.Logger
}

func NewAuthHandler(authService service.AuthService, log *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		log:         log,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	loginName := req.Email
	if loginName == "" {
		loginName = req.Username
	}

	token, user, err := h.authService.Login(c.Request.Context(), loginName, req.Password)
	if err != nil {
		httputil.Error(c, http.StatusUnauthorized, errors.ErrUnauthorized, err.Error())
		return
	}

	httputil.Success(c, gin.H{
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

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	httputil.Created(c, gin.H{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	newToken, err := h.authService.RefreshToken(c.Request.Context(), req.Token)
	if err != nil {
		httputil.Error(c, http.StatusUnauthorized, errors.ErrUnauthorized, err.Error())
		return
	}

	httputil.Success(c, gin.H{"token": newToken})
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		httputil.Error(c, http.StatusUnauthorized, errors.ErrUnauthorized, "user not authenticated")
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	user, err := h.authService.UpdateProfile(c.Request.Context(), uid, &req)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"role":       user.Role,
		"avatar":     user.Avatar,
		"department": user.Department,
		"title":      user.Title,
		"phone":      user.Phone,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		httputil.Error(c, http.StatusUnauthorized, errors.ErrUnauthorized, "user not authenticated")
		return
	}

	user, err := h.authService.GetCurrentUser(c.Request.Context(), uid)
	if err != nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, "user not found")
		return
	}

	httputil.Success(c, gin.H{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"nickname":   user.Username,
		"role":       user.Role,
		"avatar":     user.Avatar,
		"department": user.Department,
		"title":      user.Title,
	})
}
