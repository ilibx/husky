package repository

import (
	"context"
	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// GetRolePermissions 获取角色的权限映射
func (r *TicketRepository) GetRolePermissions(ctx context.Context, roleName string) (map[string]bool, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).Where("name = ?", roleName).First(&role).Error; err != nil {
		if roleName == "admin" {
			return nil, nil
		}
		return make(map[string]bool), nil
	}
	if role.Permissions == "" {
		return make(map[string]bool), nil
	}

	var permList []string
	if err := role.ScanPermissions(&permList); err != nil {
		return make(map[string]bool), nil
	}

	perms := make(map[string]bool, len(permList))
	for _, p := range permList {
		perms[p] = true
	}
	return perms, nil
}

// ListRoles 获取所有角色
func (r *TicketRepository) ListRoles(ctx context.Context, keyword string) ([]model.Role, error) {
	var list []model.Role
	query := r.db.WithContext(ctx).Order("name ASC")
	if keyword != "" {
		query = query.Where("name ILIKE ?", "%"+keyword+"%")
	}
	if err := query.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetRole 获取单个角色
func (r *TicketRepository) GetRole(ctx context.Context, id uint) (*model.Role, error) {
	var role model.Role
	if err := r.db.WithContext(ctx).First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

// CreateRole 创建角色
func (r *TicketRepository) CreateRole(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// UpdateRole 更新角色
func (r *TicketRepository) UpdateRole(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// DeleteRole 删除角色
func (r *TicketRepository) DeleteRole(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Role{}, id).Error
}
