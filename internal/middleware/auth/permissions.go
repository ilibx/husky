package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/husky/husky/pkg/errors"
)

// Resource constants
const (
	ResourceTicket     = "ticket"
	ResourceUser       = "user"
	ResourceKnowledge  = "knowledge"
	ResourceAgent      = "agent"
	ResourceSOP        = "sop"
	ResourceCategory   = "category"
	ResourceDepartment = "department"
	ResourceChannel    = "channel"
	ResourceRole       = "role"
	ResourceStats      = "stats"
	ResourceWebhook    = "webhook"
)

// Action constants
const (
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionAssign = "assign"
	ActionManage = "manage" // 完全控制
)

// PermissionKey 格式 "resource:action"
func PermissionKey(resource, action string) string {
	return resource + ":" + action
}

// PermissionMiddleware 精确权限检查中间件
// 用法: auth.PermissionMiddleware("ticket", "delete")
func PermissionMiddleware(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, errors.NewErrorResponse(errors.ErrForbidden, "permission denied"))
			c.Abort()
			return
		}

		// admin 角色拥有所有权限
		if role.(string) == "admin" {
			c.Next()
			return
		}

		// 从上下文获取已缓存的权限列表
		permsVal, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusForbidden, errors.NewErrorResponse(errors.ErrForbidden, "permission denied"))
			c.Abort()
			return
		}

		perms, ok := permsVal.(map[string]bool)
		if !ok {
			c.JSON(http.StatusForbidden, errors.NewErrorResponse(errors.ErrForbidden, "permission denied"))
			c.Abort()
			return
		}

		key := PermissionKey(resource, action)
		if !perms[key] && !perms[PermissionKey(resource, ActionManage)] {
			c.JSON(http.StatusForbidden, errors.NewErrorResponse(errors.ErrForbidden, "permission denied"))
			c.Abort()
			return
		}

		c.Next()
	}
}
