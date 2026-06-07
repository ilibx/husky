package repository

import (
	"context"
	"strings"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

type MenuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) ListAll(ctx context.Context) ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.WithContext(ctx).Order("sort asc, id asc").Find(&menus).Error
	return menus, err
}

func (r *MenuRepository) GetByRole(ctx context.Context, role string) ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.WithContext(ctx).Where("roles = '' OR roles IS NULL OR roles = ? OR roles LIKE ? OR roles LIKE ? OR roles LIKE ?",
		role, role+",%", "%,"+role, "%,"+role+",%",
	).Where("hidden = ?", false).Order("sort asc, id asc").Find(&menus).Error
	return menus, err
}

func (r *MenuRepository) GetByID(ctx context.Context, id uint) (*model.Menu, error) {
	var menu model.Menu
	err := r.db.WithContext(ctx).First(&menu, id).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

func (r *MenuRepository) Create(ctx context.Context, req *model.MenuRequest) (*model.Menu, error) {
	menu := &model.Menu{
		Name:     req.Name,
		Path:     req.Path,
		Icon:     req.Icon,
		ParentID: req.ParentID,
		Sort:     req.Sort,
		Roles:    req.Roles,
		Hidden:   req.Hidden,
		External: req.External,
		Iframe:   req.Iframe,
	}
	err := r.db.WithContext(ctx).Create(menu).Error
	return menu, err
}

func (r *MenuRepository) Update(ctx context.Context, id uint, req *model.MenuRequest) error {
	return r.db.WithContext(ctx).Model(&model.Menu{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":      req.Name,
		"path":      req.Path,
		"icon":      req.Icon,
		"parent_id": req.ParentID,
		"sort":      req.Sort,
		"roles":     req.Roles,
		"hidden":    req.Hidden,
		"external":  req.External,
		"iframe":    req.Iframe,
	}).Error
}

func (r *MenuRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Menu{}, id).Error
}

func (r *MenuRepository) UpdateRoleMenus(ctx context.Context, roleName string, menuIDs []uint) error {
	var roles []model.Role
	if err := r.db.WithContext(ctx).Find(&roles).Error; err != nil {
		return err
	}
	roleNames := make([]string, 0, len(roles))
	roleExists := false
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
		if role.Name == roleName {
			roleExists = true
		}
	}
	if !roleExists {
		return gorm.ErrRecordNotFound
	}

	checked := make(map[uint]bool, len(menuIDs))
	for _, id := range menuIDs {
		checked[id] = true
	}

	var menus []model.Menu
	if err := r.db.WithContext(ctx).Find(&menus).Error; err != nil {
		return err
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, menu := range menus {
			roleSet := map[string]bool{}
			if menu.Roles == "" {
				for _, name := range roleNames {
					roleSet[name] = true
				}
			} else {
				for _, name := range strings.Split(menu.Roles, ",") {
					name = strings.TrimSpace(name)
					if name != "" {
						roleSet[name] = true
					}
				}
			}
			if checked[menu.ID] {
				roleSet[roleName] = true
			} else {
				delete(roleSet, roleName)
			}

			nextRoles := make([]string, 0, len(roleNames))
			for _, name := range roleNames {
				if roleSet[name] {
					nextRoles = append(nextRoles, name)
				}
			}
			rolesValue := strings.Join(nextRoles, ",")
			if len(nextRoles) == len(roleNames) {
				rolesValue = ""
			}
			if err := tx.Model(&model.Menu{}).Where("id = ?", menu.ID).Update("roles", rolesValue).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *MenuRepository) HasRole(roles string, role string) bool {
	if roles == "" {
		return true
	}
	parts := strings.Split(roles, ",")
	for _, p := range parts {
		if strings.TrimSpace(p) == role {
			return true
		}
	}
	return false
}

func BuildMenuTree(menus []model.Menu) []model.MenuResponse {
	menuMap := make(map[uint]*model.MenuResponse)
	var roots []model.MenuResponse

	for _, m := range menus {
		node := model.MenuResponse{
			ID:        m.ID,
			Name:      m.Name,
			Path:      m.Path,
			Icon:      m.Icon,
			ParentID:  m.ParentID,
			Sort:      m.Sort,
			Roles:     m.Roles,
			Hidden:    m.Hidden,
			External:  m.External,
			Iframe:    m.Iframe,
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
		}
		menuMap[m.ID] = &node
	}

	for _, m := range menus {
		node := menuMap[m.ID]
		if m.ParentID != nil {
			if parent, ok := menuMap[*m.ParentID]; ok {
				parent.Children = append(parent.Children, *node)
			}
		} else {
			roots = append(roots, *node)
		}
	}

	return roots
}
