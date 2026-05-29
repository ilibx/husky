package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
	"github.com/husky/husky/pkg/llm"
)

var workflowMu sync.Map

type WorkflowService struct {
	wfRepo     *repository.WorkflowRepository
	ticketRepo *repository.TicketRepository
	agentRepo  *repository.AgentRepository
	kbRepo     repository.VectorStoreRepository
	chatSvc    *llm.ChatService
}

func NewWorkflowService(
	wfRepo *repository.WorkflowRepository,
	ticketRepo *repository.TicketRepository,
	agentRepo *repository.AgentRepository,
	kbRepo repository.VectorStoreRepository,
	chatSvc *llm.ChatService,
) *WorkflowService {
	return &WorkflowService{
		wfRepo:     wfRepo,
		ticketRepo: ticketRepo,
		agentRepo:  agentRepo,
		kbRepo:     kbRepo,
		chatSvc:    chatSvc,
	}
}

type SOPStep struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	AgentID  *uint                  `json:"agent_id,omitempty"`
	Tool     string                 `json:"tool,omitempty"`
	Goal     string                 `json:"goal,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty"`
	AssignTo string                 `json:"assign_to,omitempty"`
}

type ReActAction struct {
	Thought     string `json:"thought"`
	Action      string `json:"action"`
	ActionInput string `json:"action_input"`
}

func (s *WorkflowService) CreateFromSOP(ctx context.Context, ticket *model.Ticket, sop *model.SOP) (*model.Workflow, error) {
	existing, _ := s.wfRepo.GetByTicket(ctx, ticket.ID)
	if existing != nil {
		return existing, nil
	}

	var steps []SOPStep
	if err := json.Unmarshal([]byte(sop.Steps), &steps); err != nil {
		return nil, fmt.Errorf("failed to parse SOP steps: %w", err)
	}

	wf := &model.Workflow{
		TicketID:    ticket.ID,
		SOPID:       sop.ID,
		Status:      "pending",
		TotalSteps:  len(steps),
		CurrentStep: 0,
	}
	if err := s.wfRepo.Create(ctx, wf); err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	for i, step := range steps {
		ws := s.buildStep(wf.ID, i, step, ticket)
		if err := s.wfRepo.CreateStep(ctx, ws); err != nil {
			return nil, fmt.Errorf("failed to create step %d: %w", i, err)
		}
	}

	wf.Status = "running"
	s.wfRepo.Update(ctx, wf)

	if len(steps) > 0 {
		go s.executeNextStep(context.Background(), wf.ID)
	}
	return wf, nil
}

func (s *WorkflowService) buildStep(workflowID uint, index int, step SOPStep, ticket *model.Ticket) *model.WorkflowStep {
	cfgBytes, _ := json.Marshal(step.Config)
	ws := &model.WorkflowStep{
		WorkflowID: workflowID,
		StepIndex:  index,
		Name:       step.Name,
		Type:       step.Type,
		AgentID:    step.AgentID,
		Config:     string(cfgBytes),
		Status:     "pending",
	}
	if step.Type == "human" {
		switch step.AssignTo {
		case "requester":
			ws.AssigneeID = &ticket.RequesterID
		default:
			if ticket.AssigneeID != nil {
				ws.AssigneeID = ticket.AssigneeID
			}
		}
	}
	return ws
}

func (s *WorkflowService) executeNextStep(ctx context.Context, workflowID uint) {
	wf, err := s.wfRepo.GetByID(ctx, workflowID)
	if err != nil || wf == nil {
		return
	}
	pending, err := s.wfRepo.GetPendingSteps(ctx, workflowID)
	if err != nil {
		return
	}
	if len(pending) == 0 {
		wf.Status = "completed"
		s.wfRepo.Update(ctx, wf)
		return
	}

	step := pending[0]
	step.Status = "running"
	now := time.Now()
	step.StartedAt = &now
	s.wfRepo.UpdateStep(ctx, &step)

	wf.CurrentStep = step.StepIndex
	s.wfRepo.Update(ctx, wf)

	s.executeStep(ctx, &step, wf)
}

func (s *WorkflowService) executeStep(ctx context.Context, step *model.WorkflowStep, wf *model.Workflow) {
	ticket, err := s.ticketRepo.GetByID(ctx, wf.TicketID)
	if err != nil {
		return
	}

	switch step.Type {
	case "react":
		s.reActLoop(ctx, step, ticket, wf)
	case "human":
		s.executeHumanStep(ctx, step, ticket)
	case "condition":
		s.executeConditionStep(ctx, step, ticket)
	case "notification":
		s.executeNotificationStep(ctx, step, ticket)
	default:
		s.completeStep(ctx, step, "unknown step type, skipped")
	}
}

type roundHist struct {
	action ReActAction
	result string
}

func (s *WorkflowService) reActLoop(ctx context.Context, step *model.WorkflowStep, ticket *model.Ticket, wf *model.Workflow) {
	if s.chatSvc == nil {
		s.failStep(ctx, step, "LLM not configured")
		return
	}

	agent := s.selectAgent(ctx, step)
	if agent == nil {
		s.failStep(ctx, step, "no available agent")
		return
	}

	var stepCfg struct {
		Goal string `json:"goal"`
	}
	if step.Config != "" {
		json.Unmarshal([]byte(step.Config), &stepCfg)
	}
	goal := stepCfg.Goal
	if goal == "" {
		goal = step.Name
	}

	kbResults, _ := s.kbRepo.SearchSimilar(ctx, model.KnowledgeQuery{
		Query: fmt.Sprintf("%s %s %s", ticket.Title, ticket.Description, goal),
		Limit: 4,
	}, nil)
	var kbContext string
	for _, r := range kbResults {
		kbContext += fmt.Sprintf("- %s: %s\n", r.Title, r.Content)
	}

	systemPrompt := fmt.Sprintf(`You are an AI agent executing SOP steps. You MUST respond with valid JSON only:
{"thought":"...","action":"tool_name","action_input":"..."}

Tools:
- search_kb: Search knowledge base. Input: search query string
- reply_user: Reply to user. Input: your response text
- update_ticket: Update ticket status or priority. Input: JSON like {"status":"resolved"}
- complete: Mark this step done. Input: result summary
- fail: Mark this step failed. Input: reason

Step: %s
Goal: %s
Ticket: %s | Status: %s | Priority: %s
Description: %s

Relevant knowledge:
%s

Stay focused. When goal achieved, use "complete".`, step.Name, goal, ticket.Title, ticket.Status, ticket.Priority, ticket.Description, kbContext)

	var history []roundHist
	maxRounds := 10

	for roundIdx := 0; roundIdx < maxRounds; roundIdx++ {
		var historyText string
		start := 0
		if len(history) > 5 {
			start = len(history) - 5
		}
		for i := start; i < len(history); i++ {
			h := history[i]
			historyText += fmt.Sprintf("Round %d: thought=%s action=%s input=%s\n  -> %s\n",
				i+1, h.action.Thought, h.action.Action, h.action.ActionInput, h.result)
		}

		messages := []llm.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: fmt.Sprintf("Current state: step=%s, ticket_status=%s\nHistory:\n%s\n\nWhat should I do next?", step.Name, ticket.Status, historyText)},
		}

		resp, err := s.chatSvc.Chat(ctx, &llm.ChatRequest{
			Model:       agent.Model,
			Messages:    messages,
			Temperature: agent.Temperature,
			MaxTokens:   agent.MaxTokens,
		})
		if err != nil {
			s.failStep(ctx, step, fmt.Sprintf("LLM error: %v", err))
			return
		}

		var action ReActAction
		if err := json.Unmarshal([]byte(resp.Content), &action); err != nil {
			history = append(history, roundHist{
				action: ReActAction{Thought: "parse_error", Action: "invalid_json", ActionInput: resp.Content},
				result: "Your response was not valid JSON. Respond only with the JSON format.",
			})
			continue
		}

		log.Printf("ReAct [%s] round %d: %s -> %s(%s)", step.Name, roundIdx+1, action.Thought, action.Action, action.ActionInput)

		switch action.Action {
		case "search_kb":
			results, _ := s.kbRepo.SearchSimilar(ctx, model.KnowledgeQuery{Query: action.ActionInput, Limit: 3}, nil)
			var obs []string
			for _, r := range results {
				obs = append(obs, fmt.Sprintf("%s: %s", r.Title, r.Content))
			}
			result := "Observation: No relevant knowledge found."
			if len(obs) > 0 {
				result = "Observation:\n" + strings.Join(obs, "\n")
			}
			history = append(history, roundHist{action: action, result: result})

		case "reply_user":
			s.ticketRepo.AddComment(ctx, &model.Comment{
				TicketID:   ticket.ID,
				UserID:     AgentUserID,
				Content:    fmt.Sprintf("[Agent: %s]\n%s", agent.Name, action.ActionInput),
				IsInternal: false,
			})
			history = append(history, roundHist{action: action, result: "replied to user"})

		case "update_ticket":
			var update struct {
				Status   string `json:"status"`
				Priority string `json:"priority"`
			}
			json.Unmarshal([]byte(action.ActionInput), &update)
			var changes []string
			if update.Status != "" && model.IsValidTransition(model.TicketStatus(ticket.Status), model.TicketStatus(update.Status)) {
				ticket.Status = update.Status
				s.ticketRepo.Update(ctx, ticket)
				changes = append(changes, "status="+update.Status)
			}
			if update.Priority != "" && model.IsValidPriority(update.Priority) {
				ticket.Priority = update.Priority
				s.ticketRepo.Update(ctx, ticket)
				changes = append(changes, "priority="+update.Priority)
			}
			result := "no changes"
			if len(changes) > 0 {
				result = "updated: " + strings.Join(changes, ", ")
			}
			history = append(history, roundHist{action: action, result: result})

		case "complete":
			s.ticketRepo.AddComment(ctx, &model.Comment{
				TicketID:   ticket.ID,
				UserID:     AgentUserID,
				Content:    fmt.Sprintf("[Agent: %s] Step completed: %s\nResult: %s", agent.Name, step.Name, action.ActionInput),
				IsInternal: true,
			})
			s.completeStep(ctx, step, action.ActionInput)
			return

		case "fail":
			s.failStep(ctx, step, action.ActionInput)
			return

		default:
			history = append(history, roundHist{action: action, result: fmt.Sprintf("unknown tool '%s', use one of: search_kb, reply_user, update_ticket, complete, fail", action.Action)})
		}
	}

	s.completeStep(ctx, step, "reached max rounds, auto-completed")
}

func (s *WorkflowService) selectAgent(ctx context.Context, step *model.WorkflowStep) *model.Agent {
	if step.AgentID != nil {
		agent, err := s.agentRepo.GetByID(ctx, *step.AgentID)
		if err == nil && agent.Enabled {
			return agent
		}
	}
	agents, _, _ := s.agentRepo.List(ctx, 0, 100)
	var best *model.Agent
	for i := range agents {
		a := &agents[i]
		if a.Enabled && (a.Type == "llm" || a.Type == "hybrid") {
			if best == nil || a.ID < best.ID {
				best = a
			}
		}
	}
	return best
}

func (s *WorkflowService) executeHumanStep(ctx context.Context, step *model.WorkflowStep, ticket *model.Ticket) {
	s.ticketRepo.AddComment(ctx, &model.Comment{
		TicketID:   ticket.ID,
		UserID:     AgentUserID,
		Content:    fmt.Sprintf("Requires human processing: %s", step.Name),
		IsInternal: true,
	})

	notifyUserID := uint(0)
	if step.AssigneeID != nil {
		notifyUserID = *step.AssigneeID
	} else if ticket.AssigneeID != nil {
		notifyUserID = *ticket.AssigneeID
	}
	if notifyUserID > 0 {
		s.createNotification(ctx, notifyUserID, "workflow",
			fmt.Sprintf("需要进行人工处理: %s", step.Name),
			fmt.Sprintf("工单 #%s 需要您进行人工处理，步骤: %s", ticket.TicketNo, step.Name),
			ticket.ID, "ticket")
	}

	var cfg struct {
		TimeoutMinutes int `json:"timeout_minutes"`
	}
	if step.Config != "" {
		json.Unmarshal([]byte(step.Config), &cfg)
	}
	if cfg.TimeoutMinutes <= 0 {
		cfg.TimeoutMinutes = 1440
	}

	go func(stepID uint, timeout time.Duration) {
		select {
		case <-time.After(timeout):
			ctx := context.Background()
			saved, err := s.wfRepo.GetStepByID(ctx, stepID)
			if err != nil || saved.Status != "pending" {
				return
			}
			s.failStep(ctx, saved, "human step timed out")
		case <-ctx.Done():
		}
	}(step.ID, time.Duration(cfg.TimeoutMinutes)*time.Minute)
}

func (s *WorkflowService) executeConditionStep(ctx context.Context, step *model.WorkflowStep, ticket *model.Ticket) {
	var cfg struct {
		Field    string      `json:"field"`
		Operator string      `json:"operator"`
		Value    interface{} `json:"value"`
	}
	if step.Config != "" {
		json.Unmarshal([]byte(step.Config), &cfg)
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
		json.Unmarshal([]byte(step.Config), &cfg)
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

func (s *WorkflowService) completeStep(ctx context.Context, step *model.WorkflowStep, result string) {
	now := time.Now()
	step.Status = "completed"
	step.Result = result
	step.CompletedAt = &now
	s.wfRepo.UpdateStep(ctx, step)

	wf, _ := s.wfRepo.GetByID(ctx, step.WorkflowID)
	if wf != nil {
		isLast := step.StepIndex >= wf.TotalSteps-1
		if isLast {
			wf.Status = "completed"
			s.wfRepo.Update(ctx, wf)

			ticket, _ := s.ticketRepo.GetByID(ctx, wf.TicketID)
			if ticket != nil && ticket.RequesterID > 0 {
				s.createNotification(ctx, ticket.RequesterID, "workflow",
					fmt.Sprintf("工单流程已完成: %s", step.Name),
					fmt.Sprintf("工单 #%s 工作流执行完成，最终步骤: %s\n结果: %s", ticket.TicketNo, step.Name, result),
					ticket.ID, "ticket")
			}
			return
		}
	}

	s.executeNextStep(ctx, step.WorkflowID)
}

func (s *WorkflowService) failStep(ctx context.Context, step *model.WorkflowStep, reason string) {
	now := time.Now()
	step.Status = "failed"
	step.Result = reason
	step.CompletedAt = &now
	s.wfRepo.UpdateStep(ctx, step)

	wf, _ := s.wfRepo.GetByID(ctx, step.WorkflowID)
	if wf != nil {
		wf.Status = "failed"
		s.wfRepo.Update(ctx, wf)

		ticket, _ := s.ticketRepo.GetByID(ctx, wf.TicketID)
		if ticket != nil {
			notifyID := uint(0)
			if ticket.AssigneeID != nil {
				notifyID = *ticket.AssigneeID
			}
			if notifyID > 0 {
				s.createNotification(ctx, notifyID, "workflow",
					fmt.Sprintf("工作流步骤失败: %s", step.Name),
					fmt.Sprintf("工单 #%s 的工作流步骤 %s 执行失败:\n%s", ticket.TicketNo, step.Name, reason),
					ticket.ID, "ticket")
			}
		}
	}
}

func (s *WorkflowService) OnUserMessage(ctx context.Context, ticketID uint, message string) {
	wf, err := s.wfRepo.GetByTicket(ctx, ticketID)
	if err != nil || wf == nil || wf.Status != "running" {
		return
	}

	for _, step := range wf.Steps {
		if step.Status == "running" && step.Type == "react" {
			key := fmt.Sprintf("react-%d", step.ID)
			if _, loaded := workflowMu.LoadOrStore(key, true); loaded {
				log.Printf("ReAct loop already running for step %d, deferring message", step.ID)
				return
			}
			defer workflowMu.Delete(key)

			step.Result += fmt.Sprintf("\n[User]: %s", message)
			s.wfRepo.UpdateStep(ctx, &step)

			ticket, _ := s.ticketRepo.GetByID(ctx, ticketID)
			if ticket != nil {
				stepCopy := step
				s.reActLoop(context.Background(), &stepCopy, ticket, wf)
			}
			return
		}
	}
}

func (s *WorkflowService) CompleteHumanStep(ctx context.Context, stepID uint, result string) error {
	step, err := s.wfRepo.GetStepByID(ctx, stepID)
	if err != nil {
		return err
	}
	if step.Type != "human" {
		return fmt.Errorf("step %d is not a human step", stepID)
	}
	s.completeStep(ctx, step, result)
	return nil
}

func (s *WorkflowService) GetWorkflowByID(ctx context.Context, id uint) (*model.Workflow, error) {
	return s.wfRepo.GetByID(ctx, id)
}

func (s *WorkflowService) GetWorkflowByTicket(ctx context.Context, ticketID uint) (*model.Workflow, error) {
	return s.wfRepo.GetByTicket(ctx, ticketID)
}

func (s *WorkflowService) ListWorkflows(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]model.Workflow, int64, error) {
	return s.wfRepo.List(ctx, offset, limit, filters)
}

func (s *WorkflowService) ListPendingSteps(ctx context.Context, userID uint, offset, limit int) ([]model.WorkflowStep, int64, error) {
	return s.wfRepo.ListStepsByAssignee(ctx, userID, offset, limit)
}

func (s *WorkflowService) createNotification(ctx context.Context, userID uint, nType, title, content string, refID uint, refType string) {
	if userID == 0 {
		return
	}
	s.ticketRepo.CreateNotification(ctx, &model.Notification{
		UserID:        userID,
		Type:          nType,
		Title:         title,
		Content:       content,
		ReferenceID:   refID,
		ReferenceType: refType,
		Status:        "unread",
	})
}
