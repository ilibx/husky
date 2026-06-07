package knowledge

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

type Handler struct {
	knowledgeService Service
	storageBasePath  string
}

func NewHandler(knowledgeService Service, storageBasePath string) *Handler {
	if storageBasePath == "" {
		storageBasePath = "./data/knowledge"
	}
	return &Handler{
		knowledgeService: knowledgeService,
		storageBasePath:  storageBasePath,
	}
}

func (h *Handler) CreateKnowledge(c *gin.Context) {
	var req model.KnowledgeBase
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	if userID, ok := c.Get("user_id"); ok {
		if uid, ok := userID.(uint); ok {
			req.CreatedByUserID = uid
		}
	}

	kb, err := h.knowledgeService.CreateKnowledge(c.Request.Context(), &req)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Created(c, kb)
}

func (h *Handler) SearchKnowledge(c *gin.Context) {
	var query model.KnowledgeQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	results, err := h.knowledgeService.SearchKnowledge(c.Request.Context(), query)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"data":  results,
		"count": len(results),
		"query": query.Query,
	})
}

func (h *Handler) GetKnowledge(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "id is required")
		return
	}

	kb, err := h.knowledgeService.GetKnowledge(c.Request.Context(), id)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	if kb == nil {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, "knowledge not found")
		return
	}

	httputil.Success(c, kb)
}

func (h *Handler) UpdateKnowledge(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "id is required")
		return
	}

	var req model.KnowledgeBase
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	req.ID = id

	if err := h.knowledgeService.UpdateKnowledge(c.Request.Context(), &req); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, req)
}

func (h *Handler) ListKnowledge(c *gin.Context) {
	category := c.Query("category")
	offset, limit, ok := httputil.ParsePagination(c)
	if !ok {
		return
	}

	list, total, err := h.knowledgeService.ListKnowledge(c.Request.Context(), offset, limit, category)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"data":  list,
		"total": total,
	})
}

func (h *Handler) DeleteKnowledge(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "id is required")
		return
	}

	if err := h.knowledgeService.DeleteKnowledge(c.Request.Context(), id); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) UploadDoc(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "file is required")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	supported := map[string]string{
		".md":   "markdown",
		".txt":  "text",
		".pdf":  "pdf",
		".doc":  "word",
		".docx": "word",
		".epub": "epub",
	}
	sourceType, ok := supported[ext]
	if !ok {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "unsupported file format: "+ext)
		return
	}

	saveDir := h.storageBasePath
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, "failed to create storage dir")
		return
	}
	savePath := filepath.Join(saveDir, file.Filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, "failed to save file")
		return
	}

	title := strings.TrimSuffix(file.Filename, ext)
	content := fmt.Sprintf("[文档] %s\n\n文件路径: %s\n格式: %s\n上传时间: %s",
		file.Filename, savePath, sourceType, time.Now().Format("2006-01-02 15:04:05"))

	if sourceType == "markdown" || sourceType == "text" {
		data, readErr := os.ReadFile(savePath)
		if readErr == nil {
			if len(data) > 1*1024*1024 {
				data = data[:1*1024*1024]
			}
			content = string(data)
		}
	}

	kb := &model.KnowledgeBase{
		Title:      title,
		Content:    content,
		SourceType: sourceType,
		SourceURL:  savePath,
		Status:     "active",
	}
	if userID, ok := c.Get("user_id"); ok {
		if uid, ok := userID.(uint); ok {
			kb.CreatedByUserID = uid
		}
	}

	created, err := h.knowledgeService.CreateKnowledge(c.Request.Context(), kb)
	if err != nil {
		os.Remove(savePath)
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, "failed to create knowledge: "+err.Error())
		return
	}

	httputil.Created(c, gin.H{
		"message": "upload complete",
		"count":   1,
		"data":    created,
	})
}

func (h *Handler) ServeFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "filename is required")
		return
	}

	// Prevent directory traversal
	if strings.Contains(filename, "..") {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid filename")
		return
	}

	saveDir := h.storageBasePath
	filePath := filepath.Join(saveDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		httputil.Error(c, http.StatusNotFound, errors.ErrNotFound, "file not found")
		return
	}

	c.File(filePath)
}

func (h *Handler) QueryKnowledge(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "query parameter 'q' is required")
		return
	}

	category := c.Query("category")
	tagsStr := c.Query("tags")
	limitStr := c.Query("limit")

	var tags []string
	if tagsStr != "" {
		for _, tag := range strings.Split(tagsStr, ",") {
			if tag != "" {
				tags = append(tags, strings.TrimSpace(tag))
			}
		}
	}

	limit := 5
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	query := model.KnowledgeQuery{
		Query:    q,
		Category: category,
		Tags:     tags,
		Limit:    limit,
	}

	results, err := h.knowledgeService.SearchKnowledge(c.Request.Context(), query)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, err.Error())
		return
	}

	httputil.Success(c, gin.H{
		"data":  results,
		"count": len(results),
		"query": q,
	})
}
