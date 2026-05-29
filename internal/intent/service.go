package intent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/husky/husky/pkg/llm"
)

// IntentType 用户意图类型
type IntentType string

const (
	IntentCreateTicket IntentType = "create_ticket" // 请求人工帮助 / 报告问题
	IntentAskKnowledge IntentType = "ask_knowledge" // 咨询知识库问题
	IntentGreeting     IntentType = "greeting"      // 打招呼 / 问候
	IntentUnknown      IntentType = "unknown"       // 无法识别
)

// IntentResult 意图识别结果
type IntentResult struct {
	Intent   IntentType `json:"intent"`
	Title    string     `json:"title,omitempty"`    // 工单标题（create_ticket 时）
	Priority string     `json:"priority,omitempty"` // 优先级（create_ticket 时）
	Category string     `json:"category,omitempty"` // 分类（create_ticket 时）
	Summary  string     `json:"summary,omitempty"`  // 问题摘要
}

type Service struct {
	chatSvc *llm.ChatService
}

func NewService(chatSvc *llm.ChatService) *Service {
	return &Service{chatSvc: chatSvc}
}

// Classify 对用户消息进行意图分类
func (s *Service) Classify(ctx context.Context, message string) (*IntentResult, error) {
	// 快速规则匹配
	msg := strings.TrimSpace(message)
	lower := strings.ToLower(msg)

	// 简单问候语直接匹配
	greetings := []string{"hi", "hello", "hey", "你好", "您好", "hi there", "good morning", "下午好", "早上好", "晚上好", "在吗", "在不在", "help"}
	for _, g := range greetings {
		if lower == g || strings.HasPrefix(lower, g+" ") {
			return &IntentResult{Intent: IntentGreeting}, nil
		}
	}

	if s.chatSvc == nil {
		return &IntentResult{Intent: IntentUnknown}, nil
	}

	prompt := fmt.Sprintf(`你是一个工单系统意图识别助手。请分析用户消息的意图，返回 JSON。

意图类型:
- create_ticket: 用户请求帮助、报告问题、需要人工处理、投诉、申请资源等需要创建工单的场景
- ask_knowledge: 用户询问信息、查询知识、咨询问题（不要求人工介入）
- greeting: 打招呼、问候、测试消息
- unknown: 无法识别

用户消息: %s

请返回 JSON 格式（不要 markdown）:
{
  "intent": "create_ticket|ask_knowledge|greeting|unknown",
  "title": "工单标题（仅 create_ticket 时）",
  "priority": "low|medium|high|urgent（仅 create_ticket 时）",
  "category": "分类名（仅 create_ticket 时，可选）",
  "summary": "问题摘要描述"
}`, message)

	chatResp, err := s.chatSvc.Chat(ctx, &llm.ChatRequest{
		Messages: []llm.ChatMessage{
			{Role: "system", Content: "你是工单系统的意图识别助手，只返回 JSON。"},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.1,
		MaxTokens:   512,
	})
	if err != nil {
		return &IntentResult{Intent: IntentUnknown}, nil
	}

	content := strings.TrimSpace(chatResp.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var result IntentResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return &IntentResult{Intent: IntentUnknown}, nil
	}
	if result.Intent == "" {
		result.Intent = IntentUnknown
	}
	return &result, nil
}
