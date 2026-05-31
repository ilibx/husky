package channel

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

type ChannelUserHandler struct {
	userRepo *repository.ChannelUserRepository
}

func NewChannelUserHandler(userRepo *repository.ChannelUserRepository) *ChannelUserHandler {
	return &ChannelUserHandler{userRepo: userRepo}
}

func (h *ChannelUserHandler) List(c *gin.Context) {
	channelType := c.Query("channel_type")
	offset, limit, ok := httputil.ParsePagination(c)
	if !ok {
		return
	}
	list, total, err := h.userRepo.List(c.Request.Context(), channelType, offset, limit)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, gin.H{"data": list, "total": total})
}

func (h *ChannelUserHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}
	user, err := h.userRepo.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, err.Error())
		return
	}
	httputil.Success(c, user)
}

func (h *ChannelUserHandler) Create(c *gin.Context) {
	var req struct {
		ChannelType string `json:"channel_type" binding:"required"`
		ExternalID  string `json:"external_id" binding:"required"`
		Name        string `json:"name"`
		Avatar      string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}
	user := &model.ChannelUser{
		ChannelType: req.ChannelType,
		ExternalID:  req.ExternalID,
		Name:        req.Name,
		Avatar:      req.Avatar,
	}
	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}
	httputil.Created(c, user)
}

func (h *ChannelUserHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}
	var req struct {
		Name   *string `json:"name"`
		Avatar *string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}
	user, err := h.userRepo.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, err.Error())
		return
	}
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, user)
}

func (h *ChannelUserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}
	if err := h.userRepo.Delete(c.Request.Context(), uint(id)); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, gin.H{"message": "deleted"})
}

func (h *ChannelUserHandler) UpdateTags(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}
	var req struct {
		TagIDs []uint `json:"tag_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}
	if err := h.userRepo.UpdateTags(c.Request.Context(), uint(id), req.TagIDs); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, gin.H{"message": "tags updated"})
}
