package model

// CreateDepartmentRequest 创建部门请求
type CreateDepartmentRequest struct {
	Name      string `json:"name" binding:"required"`
	Code      string `json:"code" binding:"required"`
	ParentID  *uint  `json:"parent_id,omitempty"`
	ManagerID *uint  `json:"manager_id,omitempty"`
}

// UpdateDepartmentRequest 更新部门请求
type UpdateDepartmentRequest struct {
	Name      *string `json:"name,omitempty"`
	Code      *string `json:"code,omitempty"`
	ParentID  *uint   `json:"parent_id,omitempty"`
	ManagerID *uint   `json:"manager_id,omitempty"`
	Status    *int    `json:"status,omitempty"`
}

// Department 部门
type Department struct {
	Base
	Name      string `gorm:"size:100;not null" json:"name"`
	Code      string `gorm:"size:50;uniqueIndex;not null" json:"code"`
	ParentID  *uint  `json:"parent_id,omitempty"`
	Path      string `gorm:"size:500" json:"path"`
	ManagerID *uint  `json:"manager_id,omitempty"`
	Status    int    `gorm:"default:1" json:"status"`

	Parent  *Department `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Manager *User       `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
}

func (Department) TableName() string {
	return "departments"
}
