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

type Service interface {
	CreateKnowledge(ctx context.Context, req *model.KnowledgeBase) (*model.KnowledgeBase, error)
	SearchKnowledge(ctx context.Context, query model.KnowledgeQuery) ([]model.KnowledgeResponse, error)
	GetKnowledge(ctx context.Context, id string) (*model.KnowledgeBase, error)
	ListKnowledge(ctx context.Context, offset, limit int, category string) ([]model.KnowledgeBase, int64, error)
	UpdateKnowledge(ctx context.Context, kb *model.KnowledgeBase) error
	DeleteKnowledge(ctx context.Context, id string) error
	ImportKnowledge(ctx context.Context, items []model.KnowledgeBase) (int, error)
	ExportKnowledge(ctx context.Context) ([]model.KnowledgeBase, error)
	ListCategories(ctx context.Context) ([]model.Category, error)
	CategoryTree(ctx context.Context) ([]model.CategoryTreeNode, error)
	Ask(ctx context.Context, question string) (*model.AnswerResponse, error)
	RecordView(ctx context.Context, id string) error
	RecommendKnowledge(ctx context.Context, limit int) ([]model.KnowledgeHotResponse, error)
}

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
		vecJSON, _ := json.Marshal(vec)
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
		vecJSON, _ := json.Marshal(vec)
		kb.Vector = datatypes.JSON(vecJSON)
	} else if !contentChanged {
		kb.Vector = existing.Vector
	}

	if err := s.vectorRepo.UpdateKnowledge(ctx, kb); err != nil {
		return fmt.Errorf("failed to update knowledge: %w", err)
	}

	return nil
}

func (s *service) ImportKnowledge(ctx context.Context, items []model.KnowledgeBase) (int, error) {
	if len(items) == 0 {
		return 0, fmt.Errorf("no items to import")
	}
	count := 0
	for i := range items {
		if items[i].Title == "" {
			continue
		}
		if items[i].Status == "" {
			items[i].Status = "active"
		}

		if s.embedding != nil {
			text := items[i].Title + "\n" + items[i].Content
			vec, err := s.embedding.Embed(ctx, text)
			if err != nil {
				return count, fmt.Errorf("failed to generate embedding for item %d: %w", i, err)
			}
			vecJSON, _ := json.Marshal(vec)
			items[i].Vector = vecJSON
		}

		if err := s.vectorRepo.InsertKnowledge(ctx, &items[i]); err != nil {
			return count, fmt.Errorf("failed to import item %d: %w", i, err)
		}
		count++
	}
	return count, nil
}

func (s *service) ExportKnowledge(ctx context.Context) ([]model.KnowledgeBase, error) {
	list, _, err := s.vectorRepo.ListKnowledge(ctx, 0, 99999, "")
	if err != nil {
		return nil, fmt.Errorf("failed to export knowledge: %w", err)
	}
	return list, nil
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

func (s *service) ListCategories(ctx context.Context) ([]model.Category, error) {
	return s.categoryRepo.ListByType(ctx, "knowledge")
}

func (s *service) CategoryTree(ctx context.Context) ([]model.CategoryTreeNode, error) {
	categories, err := s.categoryRepo.ListByType(ctx, "knowledge")
	if err != nil {
		return nil, err
	}
	return model.BuildCategoryTree(categories), nil
}

func (s *service) Ask(ctx context.Context, question string) (*model.AnswerResponse, error) {
	if question == "" {
		return nil, fmt.Errorf("question is required")
	}

	query := model.KnowledgeQuery{
		Query: question,
		Limit: 3,
	}

	results, err := s.SearchKnowledge(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search knowledge: %w", err)
	}

	resp := &model.AnswerResponse{
		Question: question,
		Sources:  results,
	}

	if s.chatSvc == nil || len(results) == 0 {
		// No LLM or no results — return raw search results
		if len(results) > 0 {
			resp.Answer = results[0].Content
		} else {
			resp.Answer = "未找到相关知识"
		}
		return resp, nil
	}

	contextStr := ""
	for i, r := range results {
		contextStr += fmt.Sprintf("[来源 %d] %s\n%s\n\n", i+1, r.Title, r.Content)
	}

	prompt := fmt.Sprintf(`你是一个企业知识库问答助手。请根据以下知识库内容回答用户问题。

知识库内容：
%s

用户问题：%s

请用中文简洁准确地回答问题。如果知识库内容不足以回答问题，请如实告知。`, contextStr, question)

	chatResp, err := s.chatSvc.Chat(ctx, &llm.ChatRequest{
		Messages: []llm.ChatMessage{
			{Role: "system", Content: "你是一个专业的企业知识库问答助手，基于提供的知识库内容回答问题。"},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.3,
		MaxTokens:   1024,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate answer: %w", err)
	}

	resp.Answer = chatResp.Content
	resp.Model = chatResp.Model
	return resp, nil
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
