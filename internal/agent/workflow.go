package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
	"github.com/husky/husky/pkg/llm"
)

type WorkflowService struct {
	wfRepo      *repository.WorkflowRepository
	ticketRepo  *repository.TicketRepository
	agentRepo   *repository.AgentRepository
	kbRepo      repository.VectorStoreRepository
	chatSvc     *llm.ChatService
	agentUserID uint

	workflowMu          sync.Map
	humanStepCancels    sync.Map
}

func NewWorkflowService(
	wfRepo *repository.WorkflowRepository,
	ticketRepo *repository.TicketRepository,
	agentRepo *repository.AgentRepository,
	kbRepo repository.VectorStoreRepository,
	chatSvc *llm.ChatService,
) *WorkflowService {
	return &WorkflowService{
		wfRepo:      wfRepo,
		ticketRepo:  ticketRepo,
		agentRepo:   agentRepo,
		kbRepo:      kbRepo,
		chatSvc:     chatSvc,
		agentUserID: 1,
	}
}

// SetAgentUserID sets the system user ID for agent-generated actions.
func (s *WorkflowService) SetAgentUserID(id uint) {
	s.agentUserID = id
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
		go s.executeNextStep(context.WithoutCancel(ctx), wf.ID)
	}
	return wf, nil
}

func (s *WorkflowService) buildStep(workflowID uint, index int, step SOPStep, ticket *model.Ticket) *model.WorkflowStep {
	var cfgStr string
	cfgBytes, err := json.Marshal(step.Config)
	if err != nil {
		log.Printf("buildStep: failed to marshal config for step %q at index %d: %v", step.Name, index, err)
	} else {
		cfgStr = string(cfgBytes)
	}
	ws := &model.WorkflowStep{
		WorkflowID: workflowID,
		StepIndex:  index,
		Name:       step.Name,
		Type:       step.Type,
		AgentID:    step.AgentID,
		Config:     cfgStr,
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

func (s *WorkflowService) OnUserMessage(ctx context.Context, ticketID uint, message string) {
	wf, err := s.wfRepo.GetByTicket(ctx, ticketID)
	if err != nil || wf == nil || wf.Status != "running" {
		return
	}

	for _, step := range wf.Steps {
		if step.Status == "running" && step.Type == "react" {
			key := fmt.Sprintf("react-%d", step.ID)
			ch, loaded := s.workflowMu.LoadOrStore(key, make(chan string, 16))
			msgCh := ch.(chan string)
			if loaded {
				select {
				case msgCh <- message:
				default:
					log.Printf("ReAct message queue full for step %d, dropping message", step.ID)
				}
				return
			}
			defer func() {
				s.workflowMu.Delete(key)
				close(msgCh)
			}()

			ticket, _ := s.ticketRepo.GetByID(ctx, ticketID)
			if ticket != nil {
				stepCopy := step
				go s.reActLoopWithQueue(context.WithoutCancel(ctx), &stepCopy, ticket, wf, msgCh)
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
