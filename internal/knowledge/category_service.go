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

func (s *service) Ask(ctx context.Context, question string, modelName string, deepThinking bool, images []string, tool string) (*model.AnswerResponse, error) {
	if question == "" {
		return nil, fmt.Errorf("question is required")
	}

	resp := &model.AnswerResponse{
		Question: question,
	}

	// Use web search tool
	if tool == "web" && s.webSearcher != nil {
		webResults, err := s.webSearcher.Search(ctx, question, 5)
		if err != nil {
			return nil, fmt.Errorf("web search failed: %w", err)
		}
		resp.Sources = make([]model.KnowledgeResponse, 0, len(webResults))
		for _, r := range webResults {
			resp.Sources = append(resp.Sources, model.KnowledgeResponse{
				Title:   r.Title,
				Content: r.Content,
				Score:   1.0,
			})
		}
	}

	// Use knowledge base search (default or combined with web results)
	if tool != "web" || len(resp.Sources) == 0 {
		query := model.KnowledgeQuery{
			Query: question,
			Limit: 3,
		}
		results, err := s.SearchKnowledge(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to search knowledge: %w", err)
		}
		resp.Sources = append(resp.Sources, results...)
	}

	if s.chatSvc == nil {
		if len(resp.Sources) > 0 {
			resp.Answer = resp.Sources[0].Content
		} else {
			resp.Answer = "未找到相关知识"
		}
		return resp, nil
	}

	var chatReq *llm.ChatRequest

	if len(resp.Sources) > 0 {
		contextStr := ""
		for i, r := range resp.Sources {
			contextStr += fmt.Sprintf("[来源 %d] %s\n%s\n\n", i+1, r.Title, r.Content)
		}

		prompt := fmt.Sprintf(`你是一个企业知识库问答助手。请根据以下知识库内容回答用户问题。

知识库内容：
%s

用户问题：%s

请用中文简洁准确地回答问题。如果知识库内容不足以回答问题，请如实告知。`, contextStr, question)

		chatReq = &llm.ChatRequest{
			Messages: []llm.ChatMessage{
				{Role: "system", Content: "你是一个专业的企业知识库问答助手，基于提供的知识库内容回答问题。"},
				{Role: "user", Content: prompt, Images: images},
			},
			Temperature: 0.3,
			MaxTokens:   1024,
		}
	} else {
		chatReq = &llm.ChatRequest{
			Messages: []llm.ChatMessage{
				{Role: "system", Content: "你是一个智能助手，请用中文友好地回答用户的问题。"},
				{Role: "user", Content: question, Images: images},
			},
			Temperature: 0.7,
			MaxTokens:   2048,
		}
	}

	if modelName != "" {
		chatReq.Model = modelName
	}

	if deepThinking {
		chatReq.ReasoningEffort = "high"
		if len(chatReq.Messages) > 0 && chatReq.Messages[0].Role == "system" {
			chatReq.Messages[0].Content += " 请一步一步地推理和分析，在给出最终答案之前展示你的思考过程。"
		}
	}

	chatResp, err := s.chatSvc.Chat(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate answer: %w", err)
	}

	resp.Answer = chatResp.Content
	resp.Model = chatResp.Model
	return resp, nil
}

func (s *service) AskStream(ctx context.Context, question string, modelName string, deepThinking bool, images []string, tool string, onChunk func(string) error) (*model.AnswerResponse, error) {
	if question == "" {
		return nil, fmt.Errorf("question is required")
	}

	resp := &model.AnswerResponse{Question: question}

	if tool == "web" && s.webSearcher != nil {
		webResults, err := s.webSearcher.Search(ctx, question, 5)
		if err != nil {
			return nil, fmt.Errorf("web search failed: %w", err)
		}
		resp.Sources = make([]model.KnowledgeResponse, 0, len(webResults))
		for _, r := range webResults {
			resp.Sources = append(resp.Sources, model.KnowledgeResponse{
				Title:   r.Title,
				Content: r.Content,
				Score:   1.0,
			})
		}
	}

	if tool != "web" || len(resp.Sources) == 0 {
		query := model.KnowledgeQuery{Query: question, Limit: 3}
		results, err := s.SearchKnowledge(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("failed to search knowledge: %w", err)
		}
		resp.Sources = append(resp.Sources, results...)
	}

	if s.chatSvc == nil {
		if len(resp.Sources) > 0 {
			resp.Answer = resp.Sources[0].Content
		} else {
			resp.Answer = "未找到相关知识"
		}
		if err := onChunk(resp.Answer); err != nil {
			return nil, err
		}
		return resp, nil
	}

	var chatReq *llm.ChatRequest

	if len(resp.Sources) > 0 {
		contextStr := ""
		for i, r := range resp.Sources {
			contextStr += fmt.Sprintf("[来源 %d] %s\n%s\n\n", i+1, r.Title, r.Content)
		}
		prompt := fmt.Sprintf(`你是一个企业知识库问答助手。请根据以下知识库内容回答用户问题。

知识库内容：
%s

用户问题：%s

请用中文简洁准确地回答问题。如果知识库内容不足以回答问题，请如实告知。`, contextStr, question)
		chatReq = &llm.ChatRequest{
			Messages: []llm.ChatMessage{
				{Role: "system", Content: "你是一个专业的企业知识库问答助手，基于提供的知识库内容回答问题。"},
				{Role: "user", Content: prompt, Images: images},
			},
			Temperature: 0.3,
			MaxTokens:   1024,
		}
	} else {
		chatReq = &llm.ChatRequest{
			Messages: []llm.ChatMessage{
				{Role: "system", Content: "你是一个智能助手，请用中文友好地回答用户的问题。"},
				{Role: "user", Content: question, Images: images},
			},
			Temperature: 0.7,
			MaxTokens:   2048,
		}
	}

	if modelName != "" {
		chatReq.Model = modelName
	}
	if deepThinking {
		chatReq.ReasoningEffort = "high"
		if len(chatReq.Messages) > 0 && chatReq.Messages[0].Role == "system" {
			chatReq.Messages[0].Content += " 请一步一步地推理和分析，在给出最终答案之前展示你的思考过程。"
		}
	}

	chatResp, err := s.chatSvc.ChatStream(ctx, chatReq, func(chunk string) error {
		return onChunk(chunk)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate answer: %w", err)
	}

	resp.Answer = chatResp.Content
	resp.Model = chatResp.Model
	return resp, nil
}
