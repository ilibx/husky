package model

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description,omitempty"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	SortOrder   int    `json:"sort_order,omitempty"`
}

// UpdateCategoryRequest 更新分类请求
type UpdateCategoryRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	SortOrder   int    `json:"sort_order,omitempty"`
}

// Category 工单分类
type Category struct {
	Base
	Name        string `gorm:"size:100;not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	Path        string `gorm:"size:500" json:"path"`
	SortOrder   int    `gorm:"default:0" json:"sort_order"`
	Status      int    `gorm:"default:1" json:"status"` // 1: 激活，0: 禁用
	Type        string `gorm:"size:50;default:'ticket'" json:"type"` // ticket, knowledge, both

	Parent   *Category   `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category  `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// CategoryTreeNode 分类树节点
type CategoryTreeNode struct {
	ID        uint               `json:"id"`
	Name      string             `json:"name"`
	ParentID  *uint              `json:"parent_id,omitempty"`
	SortOrder int                `json:"sort_order"`
	Children  []CategoryTreeNode `json:"children,omitempty"`
}

// BuildCategoryTree 将扁平分类列表构建为树
func BuildCategoryTree(categories []Category) []CategoryTreeNode {
	byParent := make(map[uint][]Category)
	var roots []Category
	for _, c := range categories {
		if c.ParentID == nil {
			roots = append(roots, c)
		} else {
			byParent[*c.ParentID] = append(byParent[*c.ParentID], c)
		}
	}

	var build func(parentID uint) []CategoryTreeNode
	build = func(parentID uint) []CategoryTreeNode {
		var nodes []CategoryTreeNode
		for _, c := range byParent[parentID] {
			node := CategoryTreeNode{
				ID:        c.ID,
				Name:      c.Name,
				ParentID:  c.ParentID,
				SortOrder: c.SortOrder,
				Children:  build(c.ID),
			}
			nodes = append(nodes, node)
		}
		return nodes
	}

	var tree []CategoryTreeNode
	for _, r := range roots {
		node := CategoryTreeNode{
			ID:        r.ID,
			Name:      r.Name,
			ParentID:  r.ParentID,
			SortOrder: r.SortOrder,
			Children:  build(r.ID),
		}
		tree = append(tree, node)
	}
	return tree
}

func (Category) TableName() string {
	return "categories"
}
