package model

import "time"

// Notification 通知记录
type Notification struct {
	Base
	UserID        uint       `gorm:"not null;index" json:"user_id"`
	Type          string     `gorm:"size:50;not null" json:"type"` // ticket_assigned, ticket_status, ticket_comment, system
	Title         string     `gorm:"size:255;not null" json:"title"`
	Content       string     `gorm:"type:text;not null" json:"content"`
	ReferenceID   uint       `json:"reference_id,omitempty"`     // 关联工单/资源 ID
	ReferenceType string     `gorm:"size:50" json:"reference_type,omitempty"` // ticket, knowledge, etc.
	IsRead        bool       `gorm:"default:false;index" json:"is_read"`
	Status        string     `gorm:"size:20;default:'unread'" json:"status"` // unread, read, sent, failed
	SentAt        *time.Time `json:"sent_at,omitempty"`
	Error         string     `gorm:"type:text" json:"error,omitempty"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Satisfaction 满意度评价
type Satisfaction struct {
	Base
	TicketID  uint   `gorm:"uniqueIndex;not null" json:"ticket_id"`
	Score     int    `gorm:"not null" json:"score"` // 1-5
	Comment   string `gorm:"type:text" json:"comment,omitempty"`
	CreatedBy uint   `gorm:"not null" json:"created_by"`

	Ticket Ticket `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
	User   User   `gorm:"foreignKey:CreatedBy" json:"user,omitempty"`
}

func (Notification) TableName() string {
	return "notifications"
}

func (Satisfaction) TableName() string {
	return "satisfactions"
}
