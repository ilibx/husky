package ticket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

// --- Tags ---

func (h *TicketHandler) CreateTag(c *gin.Context) {
	var req struct {
		Name  string `json:"name" binding:"required"`
		Color string `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	tag, err := h.ticketService.CreateTag(c.Request.Context(), req.Name, req.Color)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Created(c, tag)
}

func (h *TicketHandler) ListTags(c *gin.Context) {
	tags, err := h.ticketService.ListTags(c.Request.Context())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"data": tags})
}

func (h *TicketHandler) GetTag(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid tag id")
		return
	}

	tag, err := h.ticketService.GetTag(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	if tag == nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, "tag not found")
		return
	}

	httputil.Success(c, tag)
}

func (h *TicketHandler) UpdateTag(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid tag id")
		return
	}

	var req struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	tag := &model.Tag{Name: req.Name, Color: req.Color}
	tag.ID = id

	if err := h.ticketService.UpdateTag(c.Request.Context(), tag); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, tag)
}

func (h *TicketHandler) DeleteTag(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid tag id")
		return
	}

	if err := h.ticketService.DeleteTag(c.Request.Context(), id); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TicketHandler) GetTicketTags(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	tags, err := h.ticketService.GetTicketTags(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"data": tags})
}

func (h *TicketHandler) UpdateTicketTags(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	var req struct {
		TagIDs []uint `json:"tag_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	if err := h.ticketService.UpdateTicketTags(c.Request.Context(), id, req.TagIDs); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"success": true})
}

func (h *TicketHandler) AddTicketTags(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	var req struct {
		TagIDs []uint `json:"tag_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	if err := h.ticketService.AddTagsToTicket(c.Request.Context(), id, req.TagIDs); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"success": true})
}

func (h *TicketHandler) RemoveTicketTag(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	tagID, err := parseUintParam(c, "tagId")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid tag id")
		return
	}

	if err := h.ticketService.RemoveTagFromTicket(c.Request.Context(), id, tagID); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

