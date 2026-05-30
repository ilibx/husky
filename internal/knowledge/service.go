package knowledge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
	"github.com/husky/husky/pkg/llm"
	"gorm.io/datatypes"
)

type service struct {
	vectorRepo    repository.VectorStoreRepository
	embedding     *llm.EmbeddingService
	categoryRepo  *repository.CategoryRepository
	chatSvc       *llm.ChatService
}

func NewService(vectorRepo repository.VectorStoreRepository, embedding *llm.EmbeddingService, categoryRepo *repository.CategoryRepository, chatSvc *llm.ChatService) Service {
	return &service{
		vectorRepo:   vectorRepo,
		embedding:    embedding,
		categoryRepo: categoryRepo,
		chatSvc:      chatSvc,
	}
}

func (s *service) CreateKnowledge(ctx context.Context, req *model.KnowledgeBase) (*model.KnowledgeBase, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	if req.Status == "" {
		req.Status = "active"
	}

	if s.embedding != nil {
		text := req.Title + "\n" + req.Content
		vec, err := s.embedding.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding: %w", err)
		}
		vecJSON, err := json.Marshal(vec)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal embedding: %w", err)
		}
		req.Vector = datatypes.JSON(vecJSON)
	}

	if err := s.vectorRepo.InsertKnowledge(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to insert knowledge: %w", err)
	}

	return req, nil
}

func (s *service) SearchKnowledge(ctx context.Context, query model.KnowledgeQuery) ([]model.KnowledgeResponse, error) {
	if query.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	if query.Limit <= 0 {
		query.Limit = 5
	}

	var queryVector []float32
	if s.embedding != nil {
		vec, err := s.embedding.Embed(ctx, query.Query)
		if err == nil {
			queryVector = vec
		}
	}

	results, err := s.vectorRepo.SearchSimilar(ctx, query, queryVector)
	if err != nil {
		return nil, fmt.Errorf("failed to search knowledge: %w", err)
	}

	return results, nil
}

func (s *service) GetKnowledge(ctx context.Context, id string) (*model.KnowledgeBase, error) {
	return s.vectorRepo.GetKnowledge(ctx, id)
}

func (s *service) ListKnowledge(ctx context.Context, offset, limit int, category string) ([]model.KnowledgeBase, int64, error) {
	return s.vectorRepo.ListKnowledge(ctx, offset, limit, category)
}

func (s *service) UpdateKnowledge(ctx context.Context, kb *model.KnowledgeBase) error {
	if kb.ID == "" {
		return fmt.Errorf("id is required")
	}

	existing, err := s.vectorRepo.GetKnowledge(ctx, kb.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing knowledge: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("knowledge not found")
	}

	contentChanged := kb.Title != existing.Title || kb.Content != existing.Content
	if contentChanged && s.embedding != nil {
		text := kb.Title + "\n" + kb.Content
		vec, err := s.embedding.Embed(ctx, text)
		if err != nil {
			return fmt.Errorf("failed to regenerate embedding: %w", err)
		}
		vecJSON, err := json.Marshal(vec)
		if err != nil {
			return fmt.Errorf("failed to marshal embedding: %w", err)
		}
		kb.Vector = datatypes.JSON(vecJSON)
	} else if !contentChanged {
		kb.Vector = existing.Vector
	}

	if err := s.vectorRepo.UpdateKnowledge(ctx, kb); err != nil {
		return fmt.Errorf("failed to update knowledge: %w", err)
	}

	return nil
}

func (s *service) DeleteKnowledge(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}

	if err := s.vectorRepo.DeleteKnowledge(ctx, id); err != nil {
		return fmt.Errorf("failed to delete knowledge: %w", err)
	}

	return nil
}

func (s *service) RecordView(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	return s.vectorRepo.IncrementViewCount(ctx, id)
}

func (s *service) RecommendKnowledge(ctx context.Context, limit int) ([]model.KnowledgeHotResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.vectorRepo.ListHotKnowledge(ctx, limit)
}
