package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/vectorstore"
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
	// RecordFeedback 记录知识库反馈
	RecordFeedback(ctx context.Context, fb *model.KnowledgeFeedback) error
	// GetFeedbackStats 获取知识库反馈统计
	GetFeedbackStats(ctx context.Context, knowledgeID string) (*model.KnowledgeFeedbackStats, error)
	// HasUserFeedbacked 检查用户是否已反馈
	HasUserFeedbacked(ctx context.Context, knowledgeID string, userID uint) (bool, error)
}

type vectorStoreRepository struct {
	db    *gorm.DB
	store vectorstore.Store
}

// NewVectorStoreRepository 创建向量存储仓库实例
func NewVectorStoreRepository(db *gorm.DB) VectorStoreRepository {
	return &vectorStoreRepository{db: db}
}

// NewVectorStoreRepositoryWithStore 创建向量存储仓库实例并附加外部向量存储后端
func NewVectorStoreRepositoryWithStore(db *gorm.DB, store vectorstore.Store) VectorStoreRepository {
	return &vectorStoreRepository{db: db, store: store}
}

func (r *vectorStoreRepository) InsertKnowledge(ctx context.Context, kb *model.KnowledgeBase) error {
	if len(kb.Vector) == 0 || string(kb.Vector) == "null" {
		vectorData := make([]float32, 1536)
		vectorJSON, err := json.Marshal(vectorData)
		if err != nil {
			return fmt.Errorf("marshal default vector: %w", err)
		}
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

	// When an external vector store is configured (e.g., Neo4j), delegate
	// vector search to the store, then fetch metadata from the main DB.
	if r.store != nil && len(queryVector) > 0 {
		storeResults, err := r.store.SearchSimilar(ctx, queryVector, limit)
		if err != nil {
			return nil, fmt.Errorf("external vector store search: %w", err)
		}
		// Build score map
		scoreMap := make(map[string]float32, len(storeResults))
		ids := make([]string, 0, len(storeResults))
		for _, sr := range storeResults {
			scoreMap[sr.ID] = sr.Score
			ids = append(ids, sr.ID)
		}
		// Fetch full knowledge entries from the main DB
		var kbs []model.KnowledgeBase
		if len(ids) > 0 {
			dbQuery := r.db.WithContext(ctx).Model(&model.KnowledgeBase{}).Where("id IN ?", ids)
			if query.Category != "" {
				dbQuery = dbQuery.Where("category = ?", query.Category)
			}
			if len(query.Tags) > 0 {
				tagsJSON, _ := json.Marshal(query.Tags)
				dbQuery = dbQuery.Where("tags @> ?", string(tagsJSON))
			}
			if err := dbQuery.Find(&kbs).Error; err != nil {
				return nil, fmt.Errorf("fetch knowledge by ids: %w", err)
			}
		}
		response := make([]model.KnowledgeResponse, 0, len(kbs))
		for _, kb := range kbs {
			var tags []string
			if kb.Tags != nil {
				json.Unmarshal(kb.Tags, &tags)
			}
			response = append(response, model.KnowledgeResponse{
				ID:       kb.ID,
				Title:    kb.Title,
				Content:  kb.Content,
				Score:    scoreMap[kb.ID],
				Category: kb.Category,
				Tags:     tags,
			})
		}
		return response, nil
	}

	// pgvector search (default)
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

	// Fallback: text ILIKE search
	var results []model.KnowledgeBase
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
		score := computeTextScore(query.Query, kb.Title, kb.Content)
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

// computeTextScore 计算文本匹配的相似度得分（0.0 ~ 1.0）
func computeTextScore(query, title, content string) float32 {
	if query == "" {
		return 0.5
	}
	query = strings.ToLower(strings.TrimSpace(query))
	titleLower := strings.ToLower(title)
	contentLower := strings.ToLower(content)

	// 完全匹配
	if strings.EqualFold(query, title) {
		return 1.0
	}
	if strings.EqualFold(query, content) {
		return 1.0
	}

	terms := strings.Fields(query)
	if len(terms) == 0 {
		return 0.5
	}

	matchedCount := 0
	for _, t := range terms {
		if len(t) < 2 {
			matchedCount++
			continue
		}
		if strings.Contains(titleLower, t) || strings.Contains(contentLower, t) {
			matchedCount++
		}
	}

	ratio := float64(matchedCount) / float64(len(terms))
	return float32(0.5 + 0.5*math.Min(ratio, 1.0))
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

func (r *vectorStoreRepository) RecordFeedback(ctx context.Context, fb *model.KnowledgeFeedback) error {
	return r.db.WithContext(ctx).Create(fb).Error
}

func (r *vectorStoreRepository) GetFeedbackStats(ctx context.Context, knowledgeID string) (*model.KnowledgeFeedbackStats, error) {
	var stats model.KnowledgeFeedbackStats
	stats.KnowledgeID = knowledgeID

	if err := r.db.WithContext(ctx).Model(&model.KnowledgeFeedback{}).
		Where("knowledge_id = ?", knowledgeID).
		Select("COUNT(*) as total_count, SUM(CASE WHEN helpful THEN 1 ELSE 0 END) as helpful_count").
		Scan(&stats).Error; err != nil {
		return nil, err
	}
	if stats.TotalCount > 0 {
		stats.HelpfulRate = float64(stats.HelpfulCount) / float64(stats.TotalCount) * 100
	}
	return &stats, nil
}

func (r *vectorStoreRepository) HasUserFeedbacked(ctx context.Context, knowledgeID string, userID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.KnowledgeFeedback{}).
		Where("knowledge_id = ? AND user_id = ?", knowledgeID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
