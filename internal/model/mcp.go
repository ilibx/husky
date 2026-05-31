package model

// MCP represents a managed MCP service configuration
type MCP struct {
    Base
    Name        string `gorm:"size:100;not null" json:"name"`
    Description string `gorm:"type:text" json:"description,omitempty"`
    Endpoint    string `gorm:"size:255;not null" json:"endpoint"` // URL of the MCP service
    Enabled     bool   `gorm:"default:true" json:"enabled"`
    CreatedBy   uint   `gorm:"not null" json:"created_by"`

    Creator User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (MCP) TableName() string { return "mcps" }
