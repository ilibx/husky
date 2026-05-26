package service

import (
	"context"
	"fmt"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
)

// KnowledgeService 知识库服务接口
type KnowledgeService interface {
	// CreateKnowledge 创建知识条目
	CreateKnowledge(ctx context.Context, req *model.KnowledgeBase) (*model.KnowledgeBase, error)
	// SearchKnowledge 搜索知识
	SearchKnowledge(ctx context.Context, query model.KnowledgeQuery) ([]model.KnowledgeResponse, error)
	// GetKnowledge 获取单个知识条目
	GetKnowledge(ctx context.Context, id string) (*model.KnowledgeBase, error)
	// UpdateKnowledge 更新知识条目
	UpdateKnowledge(ctx context.Context, kb *model.KnowledgeBase) error
	// DeleteKnowledge 删除知识条目
	DeleteKnowledge(ctx context.Context, id string) error
}

type knowledgeService struct {
	vectorRepo repository.VectorStoreRepository
}

// NewKnowledgeService 创建知识库服务实例
func NewKnowledgeService(vectorRepo repository.VectorStoreRepository) KnowledgeService {
	return &knowledgeService{
		vectorRepo: vectorRepo,
	}
}

// CreateKnowledge 创建知识条目
func (s *knowledgeService) CreateKnowledge(ctx context.Context, req *model.KnowledgeBase) (*model.KnowledgeBase, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// 设置默认状态
	if req.Status == "" {
		req.Status = "active"
	}

	// 插入知识库（会自动生成向量）
	if err := s.vectorRepo.InsertKnowledge(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to insert knowledge: %w", err)
	}

	return req, nil
}

// SearchKnowledge 搜索知识
func (s *knowledgeService) SearchKnowledge(ctx context.Context, query model.KnowledgeQuery) ([]model.KnowledgeResponse, error) {
	if query.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// 设置默认限制
	if query.Limit <= 0 {
		query.Limit = 5
	}

	results, err := s.vectorRepo.SearchSimilar(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search knowledge: %w", err)
	}

	return results, nil
}

// GetKnowledge 获取单个知识条目
func (s *knowledgeService) GetKnowledge(ctx context.Context, id string) (*model.KnowledgeBase, error) {
	// 注意：这里需要直接从数据库获取，而不是通过向量仓库
	// 为了简化，我们暂时返回一个错误，实际实现需要从普通 repository 获取
	return nil, fmt.Errorf("not implemented yet")
}

// UpdateKnowledge 更新知识条目
func (s *knowledgeService) UpdateKnowledge(ctx context.Context, kb *model.KnowledgeBase) error {
	if kb.ID == "" {
		return fmt.Errorf("id is required")
	}

	if err := s.vectorRepo.UpdateKnowledge(ctx, kb); err != nil {
		return fmt.Errorf("failed to update knowledge: %w", err)
	}

	return nil
}

// DeleteKnowledge 删除知识条目
func (s *knowledgeService) DeleteKnowledge(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}

	if err := s.vectorRepo.DeleteKnowledge(ctx, id); err != nil {
		return fmt.Errorf("failed to delete knowledge: %w", err)
	}

	return nil
}
