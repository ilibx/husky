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

type ChannelGroupHandler struct {
	groupRepo *repository.ChannelGroupRepository
}

func NewChannelGroupHandler(groupRepo *repository.ChannelGroupRepository) *ChannelGroupHandler {
	return &ChannelGroupHandler{groupRepo: groupRepo}
}

func (h *ChannelGroupHandler) List(c *gin.Context) {
	channelType := c.Query("channel_type")
	offset, limit, ok := httputil.ParsePagination(c)
	if !ok {
		return
	}
	list, total, err := h.groupRepo.List(c.Request.Context(), channelType, offset, limit)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, gin.H{"data": list, "total": total})
}

func (h *ChannelGroupHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}
	group, err := h.groupRepo.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, err.Error())
		return
	}
	httputil.Success(c, group)
}

func (h *ChannelGroupHandler) Create(c *gin.Context) {
	var req struct {
		ChannelType string `json:"channel_type" binding:"required"`
		ExternalID  string `json:"external_id" binding:"required"`
		Name        string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}
	group := &model.ChannelGroup{
		ChannelType: req.ChannelType,
		ExternalID:  req.ExternalID,
		Name:        req.Name,
	}
	if err := h.groupRepo.Create(c.Request.Context(), group); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}
	httputil.Created(c, group)
}

func (h *ChannelGroupHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}
	var req struct {
		Name *string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}
	group, err := h.groupRepo.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, err.Error())
		return
	}
	if req.Name != nil {
		group.Name = *req.Name
	}
	if err := h.groupRepo.Update(c.Request.Context(), group); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, group)
}

func (h *ChannelGroupHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}
	if err := h.groupRepo.Delete(c.Request.Context(), uint(id)); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, gin.H{"message": "deleted"})
}

func (h *ChannelGroupHandler) UpdateTags(c *gin.Context) {
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
	if err := h.groupRepo.UpdateTags(c.Request.Context(), uint(id), req.TagIDs); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}
	httputil.Success(c, gin.H{"message": "tags updated"})
}
