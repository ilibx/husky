package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// KnowledgeBase 知识库条目
type KnowledgeBase struct {
	ID              string         `gorm:"type:uuid;primary_key" json:"id"`
	Title           string         `gorm:"type:varchar(255);not null" json:"title"`
	Content         string         `gorm:"type:text;not null" json:"content"`
	Language        string         `gorm:"type:varchar(20);default:'zh'" json:"language"`
	TitleEn         string         `gorm:"type:varchar(255)" json:"title_en,omitempty"`
	ContentEn       string         `gorm:"type:text" json:"content_en,omitempty"`
	TitleZh         string         `gorm:"type:varchar(255)" json:"title_zh,omitempty"`
	ContentZh       string         `gorm:"type:text" json:"content_zh,omitempty"`
	Category        string         `gorm:"type:varchar(100)" json:"category"`
	SourceType      string         `gorm:"type:varchar(50);default:'manual'" json:"source_type"` // manual, word, pdf, epub, markdown, text
	SourceURL       string         `gorm:"type:text" json:"source_url,omitempty"`                 // storage path or external URL
	Tags            datatypes.JSON `gorm:"type:jsonb" json:"tags"`
	Vector          datatypes.JSON `gorm:"type:vector(1536)" json:"-"`
	ViewCount       int            `gorm:"default:0" json:"view_count"`
	HotScore        float64        `gorm:"default:0" json:"hot_score"`
	Status          string         `gorm:"type:varchar(20);default:'active'" json:"status"`
	CreatedByUUID   *string        `gorm:"column:created_by;type:uuid" json:"created_by_uuid,omitempty"`
	CreatedByUserID uint           `gorm:"default:0" json:"created_by_user_id,omitempty"`
	Creator         User           `gorm:"foreignKey:CreatedByUserID" json:"creator,omitempty"`
	UpdatedAt       time.Time      `json:"updated_at"`
	CreatedAt       time.Time      `json:"created_at"`
}

// TableName 指定表名
func (KnowledgeBase) TableName() string {
	return "knowledge"
}

// BeforeCreate 钩子：生成 UUID
func (k *KnowledgeBase) BeforeCreate(tx *gorm.DB) error {
	if k.ID == "" {
		k.ID = uuid.New().String()
	}
	return nil
}

// KnowledgeQuery 知识查询请求
type KnowledgeQuery struct {
	Query    string   `json:"query" binding:"required"`
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
	Limit    int      `json:"limit" binding:"max=50"` // 默认返回最相关的 5 条
}

// KnowledgeResponse 知识响应
type KnowledgeResponse struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Language string   `json:"language,omitempty"`
	Score    float32  `json:"score"`    // 相似度得分
	Category string   `json:"category"`
	Tags     []string `json:"tags"`
}

// KnowledgeHotResponse 热门知识响应
type KnowledgeHotResponse struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Content   string  `json:"content"`
	Language  string  `json:"language,omitempty"`
	Category  string  `json:"category"`
	ViewCount int     `json:"view_count"`
	HotScore  float64 `json:"hot_score"`
}

// KnowledgeFeedback 知识库反馈
type KnowledgeFeedback struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	KnowledgeID string    `gorm:"type:uuid;not null;index" json:"knowledge_id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	Helpful     bool      `json:"helpful"`
	Comment     string    `gorm:"type:text" json:"comment,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (KnowledgeFeedback) TableName() string {
	return "knowledge_feedback"
}

// KnowledgeFeedbackStats 知识库反馈统计
type KnowledgeFeedbackStats struct {
	KnowledgeID  string  `json:"knowledge_id"`
	TotalCount   int     `json:"total_count"`
	HelpfulCount int     `json:"helpful_count"`
	HelpfulRate  float64 `json:"helpful_rate"`
}

// AnswerResponse RAG 问答响应
type AnswerResponse struct {
	Question string              `json:"question"`
	Answer   string              `json:"answer"`
	Sources  []KnowledgeResponse `json:"sources"`
	Model    string              `json:"model,omitempty"`
}
