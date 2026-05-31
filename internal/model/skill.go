package model

// Skill represents an AI skill configuration
type Skill struct {
    Base
    Name        string `gorm:"size:100;not null" json:"name"`
    Description string `gorm:"type:text" json:"description,omitempty"`
    Category    string `gorm:"size:50" json:"category"`
    Enabled     bool   `gorm:"default:true" json:"enabled"`
    CreatedBy   uint   `gorm:"not null" json:"created_by"`

    Creator User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

func (Skill) TableName() string { return "skills" }
