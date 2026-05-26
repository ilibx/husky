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
	Category  string         `gorm:"type:varchar(100)" json:"category"`
	Tags      datatypes.JSON `gorm:"type:jsonb" json:"tags"`
	Vector    datatypes.JSON `gorm:"type:vector(768)" json:"-"` // 存储向量数据，假设使用 768 维向量 (如 text-embedding-ada-002)
	Status    string         `gorm:"type:varchar(20);default:'active'" json:"status"` // active, archived
	CreatedBy string         `gorm:"type:uuid" json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	CreatedAt time.Time      `json:"created_at"`
}

// TableName 指定表名
func (KnowledgeBase) TableName() string {
	return "knowledge_base"
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
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Content  string  `json:"content"`
	Score    float32 `json:"score"`    // 相似度得分
	Category string  `json:"category"`
	Tags     []string `json:"tags"`
}
