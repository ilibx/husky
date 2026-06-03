package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
)

type SkillMCPHandler struct {
	mcpRepo repository.MCPRepository
}

func NewSkillMCPHandler(mcpRepo repository.MCPRepository) *SkillMCPHandler {
	return &SkillMCPHandler{mcpRepo: mcpRepo}
}

func (h *SkillMCPHandler) List(c *gin.Context) {
	skillID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	list, err := h.mcpRepo.ListBySkill(c, uint(skillID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"list": list})
}

func (h *SkillMCPHandler) Create(c *gin.Context) {
	skillID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var m model.MCP
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m.SkillID = uintPtr(uint(skillID))
	currentUserID, _ := c.Get("user_id")
	if uid, ok := currentUserID.(uint); ok {
		m.CreatedBy = uid
	}
	if err := h.mcpRepo.Create(c, &m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *SkillMCPHandler) Update(c *gin.Context) {
	mcpID, _ := strconv.ParseUint(c.Param("mcpId"), 10, 64)
	var m model.MCP
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m.ID = uint(mcpID)
	currentUserID, _ := c.Get("user_id")
	if uid, ok := currentUserID.(uint); ok {
		m.CreatedBy = uid
	}
	if err := h.mcpRepo.Update(c, &m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *SkillMCPHandler) Delete(c *gin.Context) {
	mcpID, _ := strconv.ParseUint(c.Param("mcpId"), 10, 64)
	if err := h.mcpRepo.Delete(c, uint(mcpID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func uintPtr(v uint) *uint {
	return &v
}
