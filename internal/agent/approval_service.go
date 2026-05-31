package agent

import (
	"context"
	"fmt"
	"time"
)

// ApproveStep 审核通过，记录 feedback 并继续工作流
func (s *WorkflowService) ApproveStep(ctx context.Context, stepID, userID uint, feedback, result string) error {
	step, err := s.wfRepo.GetStepByID(ctx, stepID)
	if err != nil {
		return err
	}
	if step.Type != "human" {
		return fmt.Errorf("step %d is not a human step", stepID)
	}
	if step.Status != "running" {
		return fmt.Errorf("step %d is not in running state", stepID)
	}

	step.Decision = "approve"
	step.Feedback = feedback
	step.ApprovedBy = &userID

	finalResult := result
	if finalResult == "" {
		finalResult = feedback
		if finalResult == "" {
			finalResult = step.Suggestion
		}
	}

	s.completeStep(ctx, step, finalResult)
	return nil
}

// RejectStep 拒绝，记录原因
func (s *WorkflowService) RejectStep(ctx context.Context, stepID, userID uint, reason string) error {
	step, err := s.wfRepo.GetStepByID(ctx, stepID)
	if err != nil {
		return err
	}
	if step.Type != "human" {
		return fmt.Errorf("step %d is not a human step", stepID)
	}
	if step.Status != "running" {
		return fmt.Errorf("step %d is not in running state", stepID)
	}

	step.Decision = "reject"
	step.RejectionReason = reason
	step.ApprovedBy = &userID
	now := time.Now()
	step.Status = "failed"
	step.Result = fmt.Sprintf("rejected: %s", reason)
	step.CompletedAt = &now
	s.wfRepo.UpdateStep(ctx, step)

	wf, _ := s.wfRepo.GetByID(ctx, step.WorkflowID)
	if wf != nil {
		wf.Status = "failed"
		s.wfRepo.Update(ctx, wf)

		if userID > 0 {
			s.createNotification(ctx, userID, "workflow",
				fmt.Sprintf("步骤被拒绝: %s", step.Name),
				fmt.Sprintf("拒绝原因: %s", reason),
				wf.TicketID, "ticket")
		}
	}

	return nil
}

// ReviseStep 打回修改，重置步骤状态为 pending 等待 AI 重新建议
func (s *WorkflowService) ReviseStep(ctx context.Context, stepID, userID uint, feedback string) error {
	step, err := s.wfRepo.GetStepByID(ctx, stepID)
	if err != nil {
		return err
	}
	if step.Type != "human" {
		return fmt.Errorf("step %d is not a human step", stepID)
	}
	if step.Status != "running" {
		return fmt.Errorf("step %d is not in running state", stepID)
	}

	step.Decision = "revise"
	step.Feedback = feedback
	step.ApprovedBy = &userID
	step.Status = "pending"
	step.Suggestion = ""
	step.StartedAt = nil
	s.wfRepo.UpdateStep(ctx, step)

	if userID > 0 {
		s.createNotification(ctx, userID, "workflow",
			fmt.Sprintf("步骤需重新处理: %s", step.Name),
			fmt.Sprintf("修改意见: %s", feedback),
			step.WorkflowID, "workflow")
	}

	wf, _ := s.wfRepo.GetByID(ctx, step.WorkflowID)
	if wf != nil {
		go s.executeStep(context.WithoutCancel(ctx), step, wf)
	}
	return nil
}
