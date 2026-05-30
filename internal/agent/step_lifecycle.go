package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/husky/husky/internal/model"
)

func (s *WorkflowService) completeStep(ctx context.Context, step *model.WorkflowStep, result string) {
	if cancel, ok := s.humanStepCancels.LoadAndDelete(step.ID); ok {
		cancel.(context.CancelFunc)()
	}
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
	if cancel, ok := s.humanStepCancels.LoadAndDelete(step.ID); ok {
		cancel.(context.CancelFunc)()
	}
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
