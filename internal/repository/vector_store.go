package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/husky/husky/internal/model"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// kbScanResult captures vector search results including the computed score.
type kbScanResult struct {
	model.KnowledgeBase
	Score float32 `gorm:"-" json:"score"`
}

// VectorStoreRepository 向量存储仓库接口
type VectorStoreRepository interface {
	// InsertKnowledge 插入知识库条目并生成向量
	InsertKnowledge(ctx context.Context, kb *model.KnowledgeBase) error
	// GetKnowledge 获取单个知识条目
	GetKnowledge(ctx context.Context, id string) (*model.KnowledgeBase, error)
	// ListKnowledge 获取知识条目列表
	ListKnowledge(ctx context.Context, offset, limit int, category string) ([]model.KnowledgeBase, int64, error)
	// SearchSimilar 搜索相似知识
	SearchSimilar(ctx context.Context, query model.KnowledgeQuery, queryVector []float32) ([]model.KnowledgeResponse, error)
	// DeleteKnowledge 删除知识库条目
	DeleteKnowledge(ctx context.Context, id string) error
	// UpdateKnowledge 更新知识库条目
	UpdateKnowledge(ctx context.Context, kb *model.KnowledgeBase) error
	// IncrementViewCount 增加浏览次数
	IncrementViewCount(ctx context.Context, id string) error
	// ListHotKnowledge 获取热门知识
	ListHotKnowledge(ctx context.Context, limit int) ([]model.KnowledgeHotResponse, error)
}

type vectorStoreRepository struct {
	db *gorm.DB
}

// NewVectorStoreRepository 创建向量存储仓库实例
func NewVectorStoreRepository(db *gorm.DB) VectorStoreRepository {
	return &vectorStoreRepository{db: db}
}

func (r *vectorStoreRepository) InsertKnowledge(ctx context.Context, kb *model.KnowledgeBase) error {
	if len(kb.Vector) == 0 || string(kb.Vector) == "null" {
		vectorData := make([]float32, 1536)
		vectorJSON, _ := json.Marshal(vectorData)
		kb.Vector = datatypes.JSON(vectorJSON)
	}
	return r.db.WithContext(ctx).Create(kb).Error
}

// SearchSimilar 搜索相关知识（向量搜索优先，文本搜索降级）
func (r *vectorStoreRepository) SearchSimilar(ctx context.Context, query model.KnowledgeQuery, queryVector []float32) ([]model.KnowledgeResponse, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = 5
	}

	var results []model.KnowledgeBase

	if len(queryVector) > 0 {
		vectorJSON, _ := json.Marshal(queryVector)
		rawSQL := `SELECT id, title, content, category, tags, vector, status, created_by,
				   created_at, updated_at, vector <-> ?::vector AS score
				   FROM knowledge
				   WHERE status = 'active'`

		var args []interface{}
		args = append(args, string(vectorJSON))

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

		var scanResults []kbScanResult
		if err := r.db.WithContext(ctx).Raw(rawSQL, args...).Scan(&scanResults).Error; err != nil {
			return nil, fmt.Errorf("failed to search knowledge: %w", err)
		}

		response := make([]model.KnowledgeResponse, 0, len(scanResults))
		for _, sr := range scanResults {
			var tags []string
			if sr.Tags != nil {
				json.Unmarshal(sr.Tags, &tags)
			}
			response = append(response, model.KnowledgeResponse{
				ID:       sr.ID,
				Title:    sr.Title,
				Content:  sr.Content,
				Score:    sr.Score,
				Category: sr.Category,
				Tags:     tags,
			})
		}
		return response, nil
	}

	dbQuery := r.db.WithContext(ctx).Model(&model.KnowledgeBase{}).Where("status = ?", "active")
	if query.Query != "" {
		keyword := "%" + query.Query + "%"
		dbQuery = dbQuery.Where("title ILIKE ? OR content ILIKE ?", keyword, keyword)
	}
	if query.Category != "" {
		dbQuery = dbQuery.Where("category = ?", query.Category)
	}
	if len(query.Tags) > 0 {
		tagsJSON, _ := json.Marshal(query.Tags)
		dbQuery = dbQuery.Where("tags @> ?", string(tagsJSON))
	}
	if err := dbQuery.Order("created_at DESC").Limit(limit).Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to search knowledge: %w", err)
	}

	response := make([]model.KnowledgeResponse, 0, len(results))
	for _, kb := range results {
		var tags []string
		if kb.Tags != nil {
			json.Unmarshal(kb.Tags, &tags)
		}
		response = append(response, model.KnowledgeResponse{
			ID:       kb.ID,
			Title:    kb.Title,
			Content:  kb.Content,
			Score:    0.5,
			Category: kb.Category,
			Tags:     tags,
		})
	}

	return response, nil
}

// GetKnowledge 获取单个知识条目
func (r *vectorStoreRepository) GetKnowledge(ctx context.Context, id string) (*model.KnowledgeBase, error) {
	var kb model.KnowledgeBase
	if err := r.db.WithContext(ctx).First(&kb, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &kb, nil
}

// ListKnowledge 获取知识条目列表
func (r *vectorStoreRepository) ListKnowledge(ctx context.Context, offset, limit int, category string) ([]model.KnowledgeBase, int64, error) {
	var list []model.KnowledgeBase
	var total int64

	query := r.db.WithContext(ctx).Model(&model.KnowledgeBase{})
	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// DeleteKnowledge 删除知识库条目
func (r *vectorStoreRepository) DeleteKnowledge(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.KnowledgeBase{}, id).Error
}

// UpdateKnowledge 更新知识库条目
func (r *vectorStoreRepository) UpdateKnowledge(ctx context.Context, kb *model.KnowledgeBase) error {
	return r.db.WithContext(ctx).Save(kb).Error
}

func (r *vectorStoreRepository) IncrementViewCount(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&model.KnowledgeBase{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).
		UpdateColumn("hot_score", gorm.Expr("hot_score + 1.0 / (EXTRACT(EPOCH FROM NOW() - created_at) / 3600 + 1)")).
		Error
}

func (r *vectorStoreRepository) ListHotKnowledge(ctx context.Context, limit int) ([]model.KnowledgeHotResponse, error) {
	var results []model.KnowledgeHotResponse
	if err := r.db.WithContext(ctx).Model(&model.KnowledgeBase{}).
		Where("status = ?", "active").
		Order("hot_score DESC").
		Limit(limit).
		Select("id, title, content, language, category, view_count, hot_score").
		Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}
