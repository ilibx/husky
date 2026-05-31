package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/llm"
)

func (s *WorkflowService) executeHumanStep(ctx context.Context, step *model.WorkflowStep, ticket *model.Ticket) {
	// 1. 生成 AI 建议（从知识库检索 + LLM 生成）
	suggestion := s.generateSuggestion(ctx, ticket, step)
	step.Suggestion = suggestion
	step.Status = "running"
	now := time.Now()
	step.StartedAt = &now
	s.wfRepo.UpdateStep(ctx, step)

	// 2. 在工单中添加 AI 建议备注
	commentContent := fmt.Sprintf("[HITL] 需要人工审核: %s\n---\nAI 建议:\n%s", step.Name, suggestion)
	s.ticketRepo.AddComment(ctx, &model.Comment{
		TicketID:   ticket.ID,
		UserID:     s.agentUserID,
		Content:    commentContent,
		IsInternal: true,
	})

	// 3. 通知负责人
	notifyUserID := uint(0)
	if step.AssigneeID != nil {
		notifyUserID = *step.AssigneeID
	} else if ticket.AssigneeID != nil {
		notifyUserID = *ticket.AssigneeID
	}
	if notifyUserID > 0 {
		notificationContent := fmt.Sprintf("工单 #%s 需要您审核\n步骤: %s\n\nAI 建议:\n%s", ticket.TicketNo, step.Name, suggestion)
		s.createNotification(ctx, notifyUserID, "workflow",
			fmt.Sprintf("需要审核: %s", step.Name),
			notificationContent,
			ticket.ID, "ticket")
	}

	// 4. 超时处理
	var cfg struct {
		TimeoutMinutes int `json:"timeout_minutes"`
	}
	if step.Config != "" {
		if err := json.Unmarshal([]byte(step.Config), &cfg); err != nil {
			log.Printf("executeHumanStep: failed to parse config: %v", err)
		}
	}
	if cfg.TimeoutMinutes <= 0 {
		cfg.TimeoutMinutes = 1440
	}

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, time.Duration(cfg.TimeoutMinutes)*time.Minute)
	s.humanStepCancels.Store(step.ID, timeoutCancel)

	go func(stepID uint, cancel context.CancelFunc) {
		defer cancel()
		<-timeoutCtx.Done()
		if timeoutCtx.Err() != context.DeadlineExceeded {
			return
		}
		saved, err := s.wfRepo.GetStepByID(context.WithoutCancel(ctx), stepID)
		if err != nil || saved.Status != "pending" {
			return
		}
		s.failStep(context.WithoutCancel(ctx), saved, "human step timed out")
	}(step.ID, timeoutCancel)
}

// generateSuggestion 基于知识库和工单上下文生成 AI 建议
func (s *WorkflowService) generateSuggestion(ctx context.Context, ticket *model.Ticket, step *model.WorkflowStep) string {
	if s.kbRepo == nil {
		return "请人工处理此步骤。"
	}

	query := model.KnowledgeQuery{
		Query: fmt.Sprintf("%s %s %s", ticket.Title, ticket.Description, step.Name),
		Limit: 3,
	}

	results, err := s.kbRepo.SearchSimilar(ctx, query, nil)
	if err != nil || len(results) == 0 {
		return "未找到相关知识库建议，请根据实际情况处理。"
	}

	if s.chatSvc == nil {
		// 无 LLM，直接返回最相关知识
		return fmt.Sprintf("相关参考:\n%s\n---\n%s", results[0].Title, results[0].Content)
	}

	contextStr := ""
	for i, r := range results {
		contextStr += fmt.Sprintf("[%d] %s\n%s\n\n", i+1, r.Title, r.Content)
	}

	// 注入历史反馈示例（HITL 闭环）
	feedbackExamples, _ := s.wfRepo.ListFeedbackExamples(ctx, 5)
	fbStr := ""
	for _, ex := range feedbackExamples {
		if ex.Feedback != "" {
			fbStr += fmt.Sprintf("- 决策: %s | 反馈: %s", ex.Decision, ex.Feedback)
			if ex.RejectionReason != "" {
				fbStr += fmt.Sprintf(" | 拒绝原因: %s", ex.RejectionReason)
			}
			fbStr += "\n"
		}
	}
	if fbStr != "" {
		fbStr = "\n历史类似步骤的审核反馈参考:\n" + fbStr
	}

	prompt := fmt.Sprintf(`你是一个工单处理助手。工单 #%s 当前需要执行步骤 "%s"。

工单标题: %s
工单描述: %s

知识库参考:
%s%s

请基于以上信息，给出具体的处理建议和操作步骤。`, ticket.TicketNo, step.Name, ticket.Title, ticket.Description, contextStr, fbStr)

	chatResp, err := s.chatSvc.Chat(ctx, &llm.ChatRequest{
		Messages: []llm.ChatMessage{
			{Role: "system", Content: "你是一个专业的工单处理助手，基于知识库给出具体的处理建议。"},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.3,
		MaxTokens:   1024,
	})
	if err != nil {
		return fmt.Sprintf("相关参考:\n%s\n---\n%s", results[0].Title, results[0].Content)
	}

	return chatResp.Content
}

func (s *WorkflowService) executeConditionStep(ctx context.Context, step *model.WorkflowStep, ticket *model.Ticket) {
	var cfg struct {
		Field    string      `json:"field"`
		Operator string      `json:"operator"`
		Value    interface{} `json:"value"`
	}
	if step.Config != "" {
		if err := json.Unmarshal([]byte(step.Config), &cfg); err != nil {
			log.Printf("executeConditionStep: failed to parse config: %v", err)
		}
	}

	matched := false
	switch cfg.Field {
	case "status":
		if cfg.Operator == "eq" && ticket.Status == cfg.Value {
			matched = true
		}
	case "priority":
		if cfg.Operator == "eq" && ticket.Priority == cfg.Value {
			matched = true
		}
	case "assignee_id":
		if cfg.Operator == "ne" && cfg.Value == nil {
			matched = ticket.AssigneeID != nil
		}
	}

	if matched {
		s.completeStep(ctx, step, "condition matched")
	} else {
		step.Status = "skipped"
		s.wfRepo.UpdateStep(ctx, step)
		s.executeNextStep(ctx, step.WorkflowID)
	}
}

func (s *WorkflowService) executeNotificationStep(ctx context.Context, step *model.WorkflowStep, ticket *model.Ticket) {
	var cfg struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if step.Config != "" {
		if err := json.Unmarshal([]byte(step.Config), &cfg); err != nil {
			log.Printf("executeNotificationStep: failed to parse config: %v", err)
		}
	}
	title := cfg.Title
	if title == "" {
		title = "Workflow notification: " + ticket.TicketNo
	}
	content := cfg.Content
	if content == "" {
		content = fmt.Sprintf("Ticket #%s workflow step: %s", ticket.TicketNo, step.Name)
	}
	if step.AssigneeID != nil {
		s.ticketRepo.CreateNotification(ctx, &model.Notification{
			UserID:        *step.AssigneeID,
			Type:          "workflow",
			Title:         title,
			Content:       content,
			ReferenceID:   ticket.ID,
			ReferenceType: "ticket",
			Status:        "unread",
		})
	}
	s.completeStep(ctx, step, "notification sent")
}
