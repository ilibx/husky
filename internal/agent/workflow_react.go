package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/pkg/llm"
)

type roundHist struct {
	action ReActAction
	result string
}

type ticketUpdate struct {
	Status   string `json:"status"`
	Priority string `json:"priority"`
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
		if err := json.Unmarshal([]byte(step.Config), &stepCfg); err != nil {
			log.Printf("reActLoop: failed to parse step config: %v", err)
		}
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
			result := s.executeSearchKB(ctx, action.ActionInput)
			history = append(history, roundHist{action: action, result: result})

		case "reply_user":
			s.executeReplyUser(ctx, ticket, agent, action.ActionInput)
			history = append(history, roundHist{action: action, result: "replied to user"})

		case "update_ticket":
			result := s.executeUpdateTicket(ctx, ticket, action.ActionInput)
			history = append(history, roundHist{action: action, result: result})

		case "complete":
			s.ticketRepo.AddComment(ctx, &model.Comment{
				TicketID:   ticket.ID,
				UserID:     s.agentUserID,
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

func (s *WorkflowService) executeSearchKB(ctx context.Context, query string) string {
	results, _ := s.kbRepo.SearchSimilar(ctx, model.KnowledgeQuery{Query: query, Limit: 3}, nil)
	var obs []string
	for _, r := range results {
		obs = append(obs, fmt.Sprintf("%s: %s", r.Title, r.Content))
	}
	if len(obs) > 0 {
		return "Observation:\n" + strings.Join(obs, "\n")
	}
	return "Observation: No relevant knowledge found."
}

func (s *WorkflowService) executeReplyUser(ctx context.Context, ticket *model.Ticket, agent *model.Agent, msg string) {
	s.ticketRepo.AddComment(ctx, &model.Comment{
		TicketID:   ticket.ID,
		UserID:     s.agentUserID,
		Content:    fmt.Sprintf("[Agent: %s]\n%s", agent.Name, msg),
		IsInternal: false,
	})
}

func (s *WorkflowService) executeUpdateTicket(ctx context.Context, ticket *model.Ticket, input string) string {
	var update ticketUpdate
	if err := json.Unmarshal([]byte(input), &update); err != nil {
		log.Printf("executeUpdateTicket: failed to parse input: %v", err)
		return "no changes"
	}
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
	if len(changes) > 0 {
		return "updated: " + strings.Join(changes, ", ")
	}
	return "no changes"
}

func (s *WorkflowService) selectAgent(ctx context.Context, step *model.WorkflowStep) *model.Agent {
	if step.AgentID != nil {
		agent, err := s.agentRepo.GetByID(ctx, *step.AgentID)
		if err == nil && agent.Enabled {
			return agent
		}
	}
	agents, _, err := s.agentRepo.List(ctx, 0, 100)
	if err != nil {
		log.Printf("findAvailableAgent: failed to list agents: %v", err)
		return nil
	}
	var best *model.Agent
	var bestScore int
	for i := range agents {
		a := &agents[i]
		if a.Enabled && (a.Type == "llm" || a.Type == "hybrid") {
			score := agentCapabilityScore(a.Model, a.MaxTokens)
			if best == nil || score > bestScore {
				best = a
				bestScore = score
			}
		}
	}
	return best
}

// agentCapabilityScore assigns a priority score to an agent model.
// Higher scores indicate more capable models.
func agentCapabilityScore(model string, maxTokens int) int {
	score := 0
	m := strings.ToLower(model)
	switch {
	case strings.Contains(m, "gpt-4") || strings.Contains(m, "gpt4"):
		score = 100
	case strings.Contains(m, "gpt-3.5") || strings.Contains(m, "gpt3.5"):
		score = 50
	case strings.Contains(m, "claude-3"):
		score = 90
	case strings.Contains(m, "claude"):
		score = 70
	case strings.Contains(m, "gemini-pro"):
		score = 80
	case strings.Contains(m, "gemini"):
		score = 60
	case strings.Contains(m, "qwen") || strings.Contains(m, "通义"):
		score = 50
	case strings.Contains(m, "baidu") || strings.Contains(m, "ernie"):
		score = 40
	default:
		score = 30
	}
	if maxTokens > 0 {
		score += maxTokens / 10000
	}
	return score
}

func (s *WorkflowService) reActLoopWithQueue(ctx context.Context, step *model.WorkflowStep, ticket *model.Ticket, wf *model.Workflow, msgCh chan string) {
	for {
		select {
		case msg, ok := <-msgCh:
			if !ok {
				return
			}
			step.Result += fmt.Sprintf("\n[User]: %s", msg)
			s.wfRepo.UpdateStep(ctx, step)
			s.reActLoop(ctx, step, ticket, wf)
		default:
			return
		}
	}
}
