package handler

import (
    "net/http"
    "strconv"
    "github.com/gin-gonic/gin"
    "github.com/husky/husky/internal/repository"
    "github.com/husky/husky/internal/model"
)

type MCPHandler struct {
    repo repository.MCPRepository
}

func NewMCPHandler(repo repository.MCPRepository) *MCPHandler {
    return &MCPHandler{repo: repo}
}

// List MCP services with pagination
func (h *MCPHandler) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    offset := (page - 1) * size
    toolType := c.Query("type")
    keyword := c.Query("keyword")
    list, total, err := h.repo.List(c, offset, size, toolType, keyword)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"list": list, "total": total})
}

func (h *MCPHandler) Get(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
    mcp, err := h.repo.GetByID(c, uint(id))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "MCP not found"})
        return
    }
    c.JSON(http.StatusOK, mcp)
}

func (h *MCPHandler) Create(c *gin.Context) {
	var m model.MCP
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	currentUserID, _ := c.Get("user_id")
	if uid, ok := currentUserID.(uint); ok {
		m.CreatedBy = uid
	}
	if err := h.repo.Create(c, &m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *MCPHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var m model.MCP
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m.ID = uint(id)
	currentUserID, _ := c.Get("user_id")
	if uid, ok := currentUserID.(uint); ok {
		m.CreatedBy = uid
	}
	if err := h.repo.Update(c, &m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *MCPHandler) Delete(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
    if err := h.repo.Delete(c, uint(id)); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
