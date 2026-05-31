package ticket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

func (h *TicketHandler) UploadAttachment(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "file is required")
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
		httputil.Error(c, http.StatusUnauthorized, errors.ErrUnauthorized, "user not authenticated")
		return
	}

	att, err := h.ticketService.UploadAttachment(c.Request.Context(), id, uid, header.Filename, file)
	if err != nil {
		h.log.Error("Failed to upload attachment", "error", err)
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Created(c, att)
}

func (h *TicketHandler) ListAttachments(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid ticket id")
		return
	}

	attachments, err := h.ticketService.ListAttachments(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{"data": attachments})
}

func (h *TicketHandler) DeleteAttachment(c *gin.Context) {
	attachmentID, err := parseUintParam(c, "attachmentId")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid attachment id")
		return
	}

	if err := h.ticketService.DeleteAttachment(c.Request.Context(), attachmentID); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}
