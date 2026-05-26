package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/husky/husky/internal/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// VectorStoreRepository 向量存储仓库接口
type VectorStoreRepository interface {
	// InsertKnowledge 插入知识库条目并生成向量
	InsertKnowledge(ctx context.Context, kb *model.KnowledgeBase) error
	// SearchSimilar 搜索相似知识
	SearchSimilar(ctx context.Context, query model.KnowledgeQuery) ([]model.KnowledgeResponse, error)
	// DeleteKnowledge 删除知识库条目
	DeleteKnowledge(ctx context.Context, id string) error
	// UpdateKnowledge 更新知识库条目
	UpdateKnowledge(ctx context.Context, kb *model.KnowledgeBase) error
}

type vectorStoreRepository struct {
	db *gorm.DB
}

// NewVectorStoreRepository 创建向量存储仓库实例
func NewVectorStoreRepository(db *gorm.DB) VectorStoreRepository {
	return &vectorStoreRepository{db: db}
}

// InsertKnowledge 插入知识库条目
// 注意：实际生产中需要调用 LLM API 生成向量，这里仅做模拟
func (r *vectorStoreRepository) InsertKnowledge(ctx context.Context, kb *model.KnowledgeBase) error {
	// TODO: 调用 LLM API 生成向量 (例如 OpenAI text-embedding-ada-002)
	// 这里使用占位符向量，实际应替换为真实的 768 维向量
	vectorData := make([]float32, 768)
	for i := range vectorData {
		vectorData[i] = 0.0 // 占位符，实际应由 embedding 模型生成
	}

	vectorJSON, err := json.Marshal(vectorData)
	if err != nil {
		return fmt.Errorf("failed to marshal vector: %w", err)
	}

	kb.Vector = datatypes.JSON(vectorJSON)

	return r.db.WithContext(ctx).Create(kb).Error
}

// SearchSimilar 使用 pgvector 进行相似度搜索
func (r *vectorStoreRepository) SearchSimilar(ctx context.Context, query model.KnowledgeQuery) ([]model.KnowledgeResponse, error) {
	// TODO: 调用 LLM API 生成查询向量
	// 这里使用占位符向量
	queryVector := make([]float32, 768)
	for i := range queryVector {
		queryVector[i] = 0.0
	}

	vectorJSON, err := json.Marshal(queryVector)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query vector: %w", err)
	}

	// 构建基础查询
	var results []model.KnowledgeBase
	dbQuery := r.db.WithContext(ctx).Model(&model.KnowledgeBase{}).Where("status = ?", "active")

	// 添加分类过滤
	if query.Category != "" {
		dbQuery = dbQuery.Where("category = ?", query.Category)
	}

	// 添加标签过滤 (PostgreSQL JSONB 操作)
	if len(query.Tags) > 0 {
		tagsJSON, _ := json.Marshal(query.Tags)
		dbQuery = dbQuery.Where("tags @> ?", string(tagsJSON))
	}

	// 限制返回数量
	limit := query.Limit
	if limit == 0 {
		limit = 5
	}
	dbQuery = dbQuery.Limit(limit)

	// 使用 pgvector 进行相似度搜索
	// 注意：这里需要使用原生 SQL 来计算余弦相似度
	// SELECT *, vector <-> '[...]' AS score FROM knowledge_base ORDER BY score LIMIT 5
	rawSQL := `SELECT *, vector <-> ?::vector AS score 
			   FROM knowledge_base 
			   WHERE status = 'active'`
	
	args := []interface{}{string(vectorJSON)}
	
	if query.Category != "" {
		rawSQL += " AND category = ?"
		args = append(args, query.Category)
	}

	if len(query.Tags) > 0 {
		tagsJSON, _ := json.Marshal(query.Tags)
		rawSQL += " AND tags @> ?::jsonb"
		args = append(args, string(tagsJSON))
	}

	rawSQL += " ORDER BY score LIMIT ?"
	args = append(args, limit)

	if err := dbQuery.Raw(rawSQL, args...).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to search knowledge: %w", err)
	}

	// 转换为响应格式
	response := make([]model.KnowledgeResponse, 0, len(results))
	for _, kb := range results {
		var tags []string
		if kb.Tags != nil {
			json.Unmarshal(kb.Tags, &tags)
		}

		// 提取分数 (需要从原始 SQL 结果中获取，这里简化处理)
		score := float32(0.0) 

		response = append(response, model.KnowledgeResponse{
			ID:       kb.ID,
			Title:    kb.Title,
			Content:  kb.Content,
			Score:    score,
			Category: kb.Category,
			Tags:     tags,
		})
	}

	return response, nil
}

// DeleteKnowledge 删除知识库条目
func (r *vectorStoreRepository) DeleteKnowledge(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.KnowledgeBase{}, id).Error
}

// UpdateKnowledge 更新知识库条目
func (r *vectorStoreRepository) UpdateKnowledge(ctx context.Context, kb *model.KnowledgeBase) error {
	// 如果需要重新生成向量，可以在这里调用 LLM API
	return r.db.WithContext(ctx).Save(kb).Error
}
