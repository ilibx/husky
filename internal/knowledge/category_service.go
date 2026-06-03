package knowledge

import (
	"context"
	"fmt"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/llm"
)

func (s *service) ListCategories(ctx context.Context) ([]model.Category, error) {
	return s.categoryRepo.List(ctx, "")
}

func (s *service) CategoryTree(ctx context.Context) ([]model.CategoryTreeNode, error) {
	categories, err := s.categoryRepo.List(ctx, "")
	if err != nil {
		return nil, err
	}
	return model.BuildCategoryTree(categories), nil
}

func (s *service) Ask(ctx context.Context, question string, modelName string, deepThinking bool, images []string) (*model.AnswerResponse, error) {
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

	chatReq := &llm.ChatRequest{
		Messages: []llm.ChatMessage{
			{Role: "system", Content: "你是一个专业的企业知识库问答助手，基于提供的知识库内容回答问题。"},
			{Role: "user", Content: prompt, Images: images},
		},
		Temperature: 0.3,
		MaxTokens:   1024,
	}

	if modelName != "" {
		chatReq.Model = modelName
	}

	if deepThinking {
		chatReq.ReasoningEffort = "high"
	}

	chatResp, err := s.chatSvc.Chat(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate answer: %w", err)
	}

	resp.Answer = chatResp.Content
	resp.Model = chatResp.Model
	return resp, nil
}
