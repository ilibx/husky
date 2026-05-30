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

// SOPMatcher matches SOPs to tickets and creates workflows.
// This is separated from the Agent Engine to keep orthogonal concerns independent.
type SOPMatcher struct {
	sopRepo         *repository.SOPRepository
	workflowService *WorkflowService
	chatSvc         *llm.ChatService
}

func NewSOPMatcher(sopRepo *repository.SOPRepository, workflowService *WorkflowService, chatSvc *llm.ChatService) *SOPMatcher {
	return &SOPMatcher{
		sopRepo:         sopRepo,
		workflowService: workflowService,
		chatSvc:         chatSvc,
	}
}

// MatchAndCreateWorkflow picks the best SOP for a ticket and creates a workflow.
func (m *SOPMatcher) MatchAndCreateWorkflow(ctx context.Context, ticket *model.Ticket) {
	if m.workflowService == nil {
		return
	}

	sop := m.selectSOP(ctx, ticket)
	if sop == nil {
		log.Printf("no matching SOP found for ticket %d", ticket.ID)
		return
	}

	log.Printf("selected SOP %q for ticket %d", sop.Name, ticket.ID)
	if _, err := m.workflowService.CreateFromSOP(ctx, ticket, sop); err != nil {
		log.Printf("failed to create workflow from SOP: %v", err)
	}
}

// selectSOP chooses the best SOP for a ticket.
// Uses LLM when multiple active SOPs exist, falls back to trigger config matching.
func (m *SOPMatcher) selectSOP(ctx context.Context, ticket *model.Ticket) *model.SOP {
	active, err := m.sopRepo.ListActive(ctx)
	if err != nil || len(active) == 0 {
		return nil
	}
	if len(active) == 1 {
		return &active[0]
	}

	if m.chatSvc != nil {
		if sop := m.selectSOPWithLLM(ctx, ticket, active); sop != nil {
			return sop
		}
	}

	return m.selectSOPByTrigger(ctx, ticket, active)
}

// selectSOPWithLLM uses the LLM to analyze the ticket and pick the best SOP.
func (m *SOPMatcher) selectSOPWithLLM(ctx context.Context, ticket *model.Ticket, candidates []model.SOP) *model.SOP {
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

	resp, err := m.chatSvc.Chat(ctx, &llm.ChatRequest{
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
func (m *SOPMatcher) selectSOPByTrigger(ctx context.Context, ticket *model.Ticket, candidates []model.SOP) *model.SOP {
	matched := make([]*model.SOP, 0)

	for i := range candidates {
		sop := &candidates[i]
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
	return matched[0]
}
