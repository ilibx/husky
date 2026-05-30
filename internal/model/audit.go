package model

// AuditLog 审计日志
type AuditLog struct {
	Base
	UserID     uint   `gorm:"index" json:"user_id"`
	Action     string `gorm:"size:100;not null" json:"action"`
	Resource   string `gorm:"size:100;not null" json:"resource"`
	ResourceID uint   `json:"resource_id"`
	OldValue   string `gorm:"type:text" json:"old_value,omitempty"` // JSON 格式
	NewValue   string `gorm:"type:text" json:"new_value,omitempty"` // JSON 格式
	IP         string `gorm:"size:50" json:"ip"`
	UserAgent  string `gorm:"size:500" json:"user_agent"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
