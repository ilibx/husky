package model

import "time"

type Menu struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Path      string    `gorm:"size:500" json:"path"`
	Icon      string    `gorm:"size:50" json:"icon"`
	ParentID  *uint     `gorm:"index" json:"parent_id"`
	Sort      int       `gorm:"default:0" json:"sort"`
	Roles     string    `gorm:"size:500" json:"roles"`
	Hidden    bool      `gorm:"default:false" json:"hidden"`
	External  bool      `gorm:"default:false" json:"external"`
	Iframe    bool      `gorm:"default:false" json:"iframe"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Menu) TableName() string { return "menus" }

type MenuRequest struct {
	Name     string `json:"name" binding:"required"`
	Path     string `json:"path"`
	Icon     string `json:"icon"`
	ParentID *uint  `json:"parent_id"`
	Sort     int    `json:"sort"`
	Roles    string `json:"roles"`
	Hidden   bool   `json:"hidden"`
	External bool   `json:"external"`
	Iframe   bool   `json:"iframe"`
}

type MenuResponse struct {
	ID        uint           `json:"id"`
	Name      string         `json:"name"`
	Path      string         `json:"path"`
	Icon      string         `json:"icon"`
	ParentID  *uint          `json:"parent_id"`
	Sort      int            `json:"sort"`
	Roles     string         `json:"roles"`
	Hidden    bool           `json:"hidden"`
	External  bool           `json:"external"`
	Iframe    bool           `json:"iframe"`
	Children  []MenuResponse `json:"children,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
