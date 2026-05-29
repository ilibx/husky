package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// KnowledgeBase 知识库条目
type KnowledgeBase struct {
	ID        string         `gorm:"type:uuid;primary_key" json:"id"`
	Title     string         `gorm:"type:varchar(255);not null" json:"title"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Language  string         `gorm:"type:varchar(20);default:'zh'" json:"language"` // zh, en, ja, etc.
	TitleEn   string         `gorm:"type:varchar(255)" json:"title_en,omitempty"`
	ContentEn string         `gorm:"type:text" json:"content_en,omitempty"`
	TitleZh   string         `gorm:"type:varchar(255)" json:"title_zh,omitempty"`
	ContentZh string         `gorm:"type:text" json:"content_zh,omitempty"`
	Category  string         `gorm:"type:varchar(100)" json:"category"`
	Tags      datatypes.JSON `gorm:"type:jsonb" json:"tags"`
	Vector    datatypes.JSON `gorm:"type:vector(1536)" json:"-"` // 存储向量数据，1536 维 (text-embedding-ada-002 默认维度)
	ViewCount int            `gorm:"default:0" json:"view_count"`
	HotScore  float64        `gorm:"default:0" json:"hot_score"`
	Status    string         `gorm:"type:varchar(20);default:'active'" json:"status"` // active, archived
	CreatedBy string         `gorm:"type:uuid" json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedAt time.Time      `json:"created_at"`
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

// AnswerResponse RAG 问答响应
type AnswerResponse struct {
	Question string            `json:"question"`
	Answer   string            `json:"answer"`
	Sources  []KnowledgeResponse `json:"sources"`
	Model    string            `json:"model,omitempty"`
}
