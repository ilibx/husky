package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
	"github.com/husky/husky/pkg/errors"
	"github.com/husky/husky/pkg/httputil"
)

type MenuHandler struct {
	repo *repository.MenuRepository
}

func NewMenuHandler(repo *repository.MenuRepository) *MenuHandler {
	return &MenuHandler{repo: repo}
}

func (h *MenuHandler) GetMenus(c *gin.Context) {
	role, _ := c.Get("role")
	roleStr, _ := role.(string)

	menus, err := h.repo.GetByRole(c.Request.Context(), roleStr)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, "failed to get menus")
		return
	}

	tree := repository.BuildMenuTree(menus)
	httputil.Success(c, gin.H{"data": tree})
}

func (h *MenuHandler) ListAllMenus(c *gin.Context) {
	menus, err := h.repo.ListAll(c.Request.Context())
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, "failed to list menus")
		return
	}

	tree := repository.BuildMenuTree(menus)
	httputil.Success(c, gin.H{"data": tree, "list": menus})
}

func (h *MenuHandler) CreateMenu(c *gin.Context) {
	var req model.MenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	menu, err := h.repo.Create(c.Request.Context(), &req)
	if err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, "failed to create menu")
		return
	}

	httputil.Created(c, menu)
}

func (h *MenuHandler) UpdateMenu(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}

	var req model.MenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, err.Error())
		return
	}

	if err := h.repo.Update(c.Request.Context(), uint(id), &req); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, "failed to update menu")
		return
	}

	httputil.Success(c, gin.H{"message": "updated"})
}

func (h *MenuHandler) DeleteMenu(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		httputil.Error(c, http.StatusBadRequest, errors.ErrInvalidParams, "invalid id")
		return
	}

	if err := h.repo.Delete(c.Request.Context(), uint(id)); err != nil {
		httputil.Error(c, http.StatusInternalServerError, errors.ErrInternal, "failed to delete menu")
		return
	}

	httputil.Success(c, gin.H{"message": "deleted"})
}
