package model

// Comment 工单评论
type Comment struct {
	Base
	TicketID   uint   `gorm:"not null;index" json:"ticket_id"`
	UserID     uint   `gorm:"not null" json:"user_id"`
	Content    string `gorm:"type:text;not null" json:"content"`
	IsInternal bool   `gorm:"default:false" json:"is_internal"` // 是否内部评论
	Visibility string `gorm:"size:20;default:'public'" json:"visibility"` // public, internal, private

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Attachment 附件
type Attachment struct {
	Base
	TicketID   uint   `gorm:"not null;index" json:"ticket_id"`
	FileName   string `gorm:"size:255;not null" json:"file_name"`
	FileSize   int64  `gorm:"not null" json:"file_size"`
	FileType   string `gorm:"size:100" json:"file_type"`
	FileURL    string `gorm:"size:500;not null" json:"file_url"`
	UploadedBy uint   `gorm:"not null" json:"uploaded_by"`

	Uploader User `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
}

func (Comment) TableName() string {
	return "comments"
}

func (Attachment) TableName() string {
	return "attachments"
}
