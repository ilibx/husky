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

// RuleAgentService defines the service methods used by rule-type agents.
// This ensures rule agents go through the proper service layer (audit, notification, validation).
type RuleAgentService interface {
	AssignTicket(ctx context.Context, id, assigneeID uint) error
	AutoAssignTicket(ctx context.Context, id uint) (*uint, error)
	UpdateStatus(ctx context.Context, id uint, status string) error
}

type Engine struct {
	agentRepo   *repository.AgentRepository
	ticketRepo  *repository.TicketRepository
	kbRepo      repository.VectorStoreRepository
	chatSvc     *llm.ChatService
	ruleSvc     RuleAgentService
	agentUserID uint
}

func NewEngine(
	agentRepo *repository.AgentRepository,
	ticketRepo *repository.TicketRepository,
	kbRepo repository.VectorStoreRepository,
	chatSvc *llm.ChatService,
	ruleSvc RuleAgentService,
) *Engine {
	return &Engine{
		agentRepo:   agentRepo,
		ticketRepo:  ticketRepo,
		kbRepo:      kbRepo,
		chatSvc:     chatSvc,
		ruleSvc:     ruleSvc,
		agentUserID: 1,
	}
}

func (e *Engine) OnTicketCreated(ctx context.Context, ticket *model.Ticket) {
	e.executeAgents(ctx, "ticket_created", ticket)
}

func (e *Engine) OnTicketUpdated(ctx context.Context, ticket *model.Ticket) {
	e.executeAgents(ctx, "ticket_updated", ticket)
}

func (e *Engine) executeAgents(ctx context.Context, triggerType string, ticket *model.Ticket) {
	agents, err := e.agentRepo.ListByTrigger(ctx, triggerType)
	if err != nil {
		log.Printf("failed to list agents: %v", err)
		return
	}

	for _, agent := range agents {
		e.runAgent(ctx, agent, ticket)
	}
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

	if e.ruleSvc == nil {
		log.Printf("rule agent service not configured, skipping")
		return
	}

	switch config.Action {
	case "auto_assign":
		agentID, err := e.ruleSvc.AutoAssignTicket(ctx, ticket.ID)
		if err == nil && agentID != nil {
			ticket.AssigneeID = agentID
		}
	case "set_status":
		if config.SetStatus != "" && model.IsValidTransition(model.TicketStatus(ticket.Status), model.TicketStatus(config.SetStatus)) {
			if err := e.ruleSvc.UpdateStatus(ctx, ticket.ID, config.SetStatus); err != nil {
				log.Printf("rule agent set_status failed: %v", err)
				return
			}
			ticket.Status = config.SetStatus
		}
	case "set_priority":
		if model.IsValidPriority(config.SetPriority) {
			if err := e.ticketRepo.UpdateFields(ctx, ticket.ID, map[string]interface{}{
				"priority": config.SetPriority,
			}); err != nil {
				log.Printf("rule agent set_priority failed: %v", err)
				return
			}
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
		UserID:     e.agentUserID,
		Content:    fmt.Sprintf("[AI Agent: %s]\n%s", agent.Name, resp.Content),
		IsInternal: true,
	})
	}
}
