package ticket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/pkg/errors"
)

func (h *TicketHandler) UploadAttachment(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "file is required"))
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			h.log.Error("failed to close uploaded file", "error", err)
		}
	}()

	userID, _ := c.Get("user_id")
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, errors.NewErrorResponse(errors.ErrUnauthorized, "user not authenticated"))
		return
	}

	att, err := h.ticketService.UploadAttachment(c.Request.Context(), id, uid, header.Filename, file)
	if err != nil {
		h.log.Error("Failed to upload attachment", "error", err)
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, att)
}

func (h *TicketHandler) ListAttachments(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid ticket id"))
		return
	}

	attachments, err := h.ticketService.ListAttachments(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": attachments})
}

func (h *TicketHandler) DeleteAttachment(c *gin.Context) {
	attachmentID, err := parseUintParam(c, "attachmentId")
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewErrorResponse(errors.ErrInvalidParams, "invalid attachment id"))
		return
	}

	if err := h.ticketService.DeleteAttachment(c.Request.Context(), attachmentID); err != nil {
		c.JSON(http.StatusInternalServerError, errors.NewErrorResponse(errors.ErrInternal, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}
