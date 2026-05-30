package model

import (
	"encoding/json"
	"time"
)

// User 用户模型
type User struct {
	Base
	Email       string     `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Phone       string     `gorm:"size:20" json:"phone"`
	Username    string     `gorm:"size:100;not null" json:"username"`
	Password    string     `gorm:"size:255;not null" json:"-"` // 不返回密码
	Avatar      string     `gorm:"size:500" json:"avatar"`
	Department  string     `gorm:"size:100" json:"department"`
	Title       string     `gorm:"size:100" json:"title"`
	Skills      string     `gorm:"type:text" json:"skills"`        // JSON array, e.g. ["network","hardware","account"]
	MaxLoad     int        `gorm:"default:10" json:"max_load"`     // 最大并发处理工单数
	Status      int        `gorm:"default:1" json:"status"`        // 1: 激活，0: 禁用
	Role        string     `gorm:"size:50;default:'user'" json:"role"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// UpdateUserRequest 用户资料更新请求
type UpdateUserRequest struct {
	Username   string `json:"username,omitempty"`
	Avatar     string `json:"avatar,omitempty"`
	Department string `json:"department,omitempty"`
	Title      string `json:"title,omitempty"`
	Phone      string `json:"phone,omitempty"`
}

// ChangeRoleRequest 角色变更请求
type ChangeRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// Role 角色
type Role struct {
	Base
	Name        string `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Permissions string `gorm:"type:text" json:"permissions"` // JSON array of permission keys, e.g. ["ticket:create","ticket:read"]
	Status      int    `gorm:"default:1" json:"status"`
}

// ScanPermissions 解析权限 JSON 到字符串切片
func (r *Role) ScanPermissions(out *[]string) error {
	if r.Permissions == "" {
		*out = nil
		return nil
	}
	return json.Unmarshal([]byte(r.Permissions), out)
}

// DefaultAdminPermissions 返回 admin 角色的默认权限列表
func DefaultAdminPermissions() []string {
	return []string{
		"ticket:manage", "user:manage", "knowledge:manage",
		"agent:manage", "sop:manage", "category:manage",
		"department:manage", "channel:manage", "role:manage",
		"stats:manage", "webhook:manage",
	}
}

// DefaultAgentPermissions 返回 agent 角色的默认权限列表
func DefaultAgentPermissions() []string {
	return []string{
		"ticket:create", "ticket:read", "ticket:update", "ticket:assign",
		"knowledge:read", "stats:read",
	}
}

func (User) TableName() string {
	return "users"
}

func (Role) TableName() string {
	return "roles"
}
