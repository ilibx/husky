package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
	"github.com/husky/husky/pkg/llm"
)

// AgentUserID is the system user ID used for agent-generated comments and actions.
// Defaults to 1 (admin). Change via config if needed.
var AgentUserID uint = 1

type Engine struct {
	agentRepo       *repository.AgentRepository
	ticketRepo      *repository.TicketRepository
	kbRepo          repository.VectorStoreRepository
	chatSvc         *llm.ChatService
	sopRepo         *repository.SOPRepository
	workflowService *WorkflowService
}

func NewEngine(
	agentRepo *repository.AgentRepository,
	ticketRepo *repository.TicketRepository,
	kbRepo repository.VectorStoreRepository,
	chatSvc *llm.ChatService,
	sopRepo *repository.SOPRepository,
	workflowService *WorkflowService,
) *Engine {
	return &Engine{
		agentRepo:       agentRepo,
		ticketRepo:      ticketRepo,
		kbRepo:          kbRepo,
		chatSvc:         chatSvc,
		sopRepo:         sopRepo,
		workflowService: workflowService,
	}
}

func (e *Engine) OnTicketCreated(ctx context.Context, ticket *model.Ticket) {
	e.executeAgents(ctx, "ticket_created", ticket)
	e.createWorkflowForTicket(ctx, ticket)
}

func (e *Engine) OnTicketUpdated(ctx context.Context, ticket *model.Ticket) {
	e.executeAgents(ctx, "ticket_updated", ticket)
}

func (e *Engine) executeAgents(ctx context.Context, triggerType string, ticket *model.Ticket) {
	agents, _, err := e.agentRepo.List(ctx, 0, 100)
	if err != nil {
		log.Printf("failed to list agents: %v", err)
		return
	}

	for _, agent := range agents {
		if !agent.Enabled {
			continue
		}
		if !e.matchesTrigger(agent, triggerType) {
			continue
		}
		e.runAgent(ctx, agent, ticket)
	}
}

// createWorkflowForTicket picks the best SOP and creates a workflow.
// Selection priority:
//  1. LLM-based intelligent selection (analyzes ticket content vs SOP descriptions)
//  2. Fallback to trigger config matching (source/priority)
func (e *Engine) createWorkflowForTicket(ctx context.Context, ticket *model.Ticket) {
	if e.workflowService == nil {
		return
	}

	sop := e.selectSOP(ctx, ticket)
	if sop == nil {
		log.Printf("no matching SOP found for ticket %d", ticket.ID)
		return
	}

	log.Printf("selected SOP %q for ticket %d", sop.Name, ticket.ID)
	if _, err := e.workflowService.CreateFromSOP(ctx, ticket, sop); err != nil {
		log.Printf("failed to create workflow from SOP: %v", err)
	}
}

// selectSOP chooses the best SOP for a ticket.
// Uses LLM when multiple active SOPs exist, falls back to trigger config matching.
func (e *Engine) selectSOP(ctx context.Context, ticket *model.Ticket) *model.SOP {
	allSOPs, _, err := e.sopRepo.List(ctx, 0, 100)
	if err != nil || len(allSOPs) == 0 {
		return nil
	}

	active := make([]model.SOP, 0, len(allSOPs))
	for _, s := range allSOPs {
		if s.Status == 1 {
			active = append(active, s)
		}
	}
	if len(active) == 0 {
		return nil
	}
	if len(active) == 1 {
		return &active[0]
	}

	// Multiple active SOPs — try LLM selection first
	if e.chatSvc != nil {
		if sop := e.selectSOPWithLLM(ctx, ticket, active); sop != nil {
			return sop
		}
	}

	// Fallback to trigger config matching
	return e.selectSOPByTrigger(ctx, ticket, active)
}

// selectSOPWithLLM uses the LLM to analyze the ticket and pick the best SOP.
func (e *Engine) selectSOPWithLLM(ctx context.Context, ticket *model.Ticket, candidates []model.SOP) *model.SOP {
	var sopList []string
	for _, s := range candidates {
		desc := s.Description
		if len(desc) > 120 {
			desc = desc[:120] + "..."
		}
		sopList = append(sopList, fmt.Sprintf("- [ID:%d] %s: %s", s.ID, s.Name, desc))
	}

	prompt := fmt.Sprintf(`You are a ticket classification system. Given a ticket and a list of available SOPs, select the single best SOP to handle this ticket.

Ticket title: %s
Ticket description: %s
Priority: %s

Available SOPs:
%s

Respond with ONLY the SOP ID number (e.g. "3"). If no SOP fits, respond with "0".`,
		ticket.Title, ticket.Description, ticket.Priority,
		strings.Join(sopList, "\n"))

	resp, err := e.chatSvc.Chat(ctx, &llm.ChatRequest{
		Messages: []llm.ChatMessage{
			{Role: "system", Content: "You are a ticket-SOP matcher. Respond with only a number."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.1,
		MaxTokens:   10,
	})
	if err != nil {
		log.Printf("LLM SOP selection error: %v", err)
		return nil
	}

	var sopID uint
	if _, err := fmt.Sscanf(strings.TrimSpace(resp.Content), "%d", &sopID); err != nil || sopID == 0 {
		log.Printf("LLM SOP selection returned invalid ID: %s", resp.Content)
		return nil
	}

	for i := range candidates {
		if candidates[i].ID == sopID {
			log.Printf("LLM selected SOP %q (ID:%d) for ticket %d", candidates[i].Name, sopID, ticket.ID)
			return &candidates[i]
		}
	}
	return nil
}

// selectSOPByTrigger matches SOPs by trigger config (source/priority).
func (e *Engine) selectSOPByTrigger(ctx context.Context, ticket *model.Ticket, candidates []model.SOP) *model.SOP {
	matched := make([]*model.SOP, 0)

	for i := range candidates {
		sop := &candidates[i]
		// Only match SOPs that are explicitly configured for this trigger type
		if sop.TriggerType != "ticket_created" && sop.TriggerType != "event" {
			continue
		}
		if sop.TriggerConfig == "" {
			matched = append(matched, sop)
			continue
		}

		var tc struct {
			Source   string `json:"source"`
			Priority string `json:"priority"`
		}
		if err := json.Unmarshal([]byte(sop.TriggerConfig), &tc); err != nil {
			matched = append(matched, sop)
			continue
		}

		ok := true
		if tc.Source != "" && ticket.Source != tc.Source {
			ok = false
		}
		if tc.Priority != "" && ticket.Priority != tc.Priority {
			ok = false
		}
		if ok {
			matched = append(matched, sop)
		}
	}

	if len(matched) == 0 {
		return nil
	}
	// Return first match (limit to one SOP per ticket)
	return matched[0]
}

func (e *Engine) SetAgentUserID(id uint) {
	AgentUserID = id
}

func (e *Engine) matchesTrigger(agent model.Agent, triggerType string) bool {
	var cfg struct {
		TriggerOn string `json:"trigger_on"`
	}
	if agent.Config != "" {
		json.Unmarshal([]byte(agent.Config), &cfg)
	}
	return cfg.TriggerOn == triggerType || cfg.TriggerOn == "any"
}

func (e *Engine) runAgent(ctx context.Context, agent model.Agent, ticket *model.Ticket) {
	switch strings.ToLower(agent.Type) {
	case "rule":
		e.runRuleAgent(ctx, agent, ticket)
	case "llm":
		e.runLLMAgent(ctx, agent, ticket)
	case "hybrid":
		e.runRuleAgent(ctx, agent, ticket)
		e.runLLMAgent(ctx, agent, ticket)
	}
}

func (e *Engine) runRuleAgent(ctx context.Context, agent model.Agent, ticket *model.Ticket) {
	var config struct {
		Action      string `json:"action"`
		AutoAssign  bool   `json:"auto_assign"`
		SetStatus   string `json:"set_status"`
		SetPriority string `json:"set_priority"`
	}
	if err := json.Unmarshal([]byte(agent.Config), &config); err != nil {
		log.Printf("failed to parse rule agent config: %v", err)
		return
	}

	switch config.Action {
	case "auto_assign":
		agentID, err := e.ticketRepo.FindLeastBusyAgent(ctx)
		if err == nil && agentID != nil {
			e.ticketRepo.UpdateFields(ctx, ticket.ID, map[string]interface{}{
				"assignee_id": *agentID,
			})
			ticket.AssigneeID = agentID
		}
	case "set_status":
		if config.SetStatus != "" && model.IsValidTransition(model.TicketStatus(ticket.Status), model.TicketStatus(config.SetStatus)) {
			e.ticketRepo.UpdateFields(ctx, ticket.ID, map[string]interface{}{
				"status": config.SetStatus,
			})
			ticket.Status = config.SetStatus
		}
	case "set_priority":
		if model.IsValidPriority(config.SetPriority) {
			e.ticketRepo.UpdateFields(ctx, ticket.ID, map[string]interface{}{
				"priority": config.SetPriority,
			})
			ticket.Priority = config.SetPriority
		}
	}
}

func (e *Engine) runLLMAgent(ctx context.Context, agent model.Agent, ticket *model.Ticket) {
	if e.chatSvc == nil {
		return
	}

	results, err := e.kbRepo.SearchSimilar(ctx, model.KnowledgeQuery{
		Query: ticket.Title + " " + ticket.Description,
		Limit: 3,
	}, nil)
	if err != nil {
		return
	}

	var knowledgeParts []string
	for _, r := range results {
		knowledgeParts = append(knowledgeParts, fmt.Sprintf("- %s:\n  %s", r.Title, r.Content))
	}
	knowledgeContext := strings.Join(knowledgeParts, "\n")

	prompt := fmt.Sprintf(`You are an intelligent support agent named "%s".
Ticket: %s
Description: %s
Relevant knowledge:
%s

Based on the above, provide a helpful response. If the knowledge is relevant, use it to answer.`, agent.Name, ticket.Title, ticket.Description, knowledgeContext)

	resp, err := e.chatSvc.Chat(ctx, &llm.ChatRequest{
		Model: agent.Model,
		Messages: []llm.ChatMessage{
			{Role: "system", Content: "You are a helpful support agent."},
			{Role: "user", Content: prompt},
		},
		Temperature: agent.Temperature,
		MaxTokens:   agent.MaxTokens,
	})
	if err != nil {
		log.Printf("LLM agent error: %v", err)
		return
	}

	if resp.Content != "" {
		e.ticketRepo.AddComment(ctx, &model.Comment{
			TicketID:   ticket.ID,
			UserID:     AgentUserID,
			Content:    fmt.Sprintf("[AI Agent: %s]\n%s", agent.Name, resp.Content),
			IsInternal: true,
		})
	}
}
