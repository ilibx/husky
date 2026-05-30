package repository

import (
	"context"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// WorkflowRepository 工作流 Repository
type WorkflowRepository struct {
	*BaseRepository
}

func NewWorkflowRepository(db *gorm.DB) *WorkflowRepository {
	return &WorkflowRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *WorkflowRepository) Create(ctx context.Context, wf *model.Workflow) error {
	return r.db.WithContext(ctx).Create(wf).Error
}

func (r *WorkflowRepository) GetByID(ctx context.Context, id uint) (*model.Workflow, error) {
	var wf model.Workflow
	err := r.db.WithContext(ctx).
		Preload("Ticket").
		Preload("SOP").
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_index ASC")
		}).
		First(&wf, id).Error
	if err != nil {
		return nil, err
	}
	return &wf, nil
}

func (r *WorkflowRepository) GetByTicket(ctx context.Context, ticketID uint) (*model.Workflow, error) {
	var wf model.Workflow
	err := r.db.WithContext(ctx).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_index ASC")
		}).
		Where("ticket_id = ?", ticketID).
		Order("created_at DESC").
		First(&wf).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &wf, nil
}

func (r *WorkflowRepository) List(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]model.Workflow, int64, error) {
	var list []model.Workflow
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Workflow{})
	for k, v := range filters {
		query = query.Where(k+" = ?", v)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Ticket").
		Preload("SOP").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&list).Error
	return list, total, err
}

func (r *WorkflowRepository) Update(ctx context.Context, wf *model.Workflow) error {
	return r.db.WithContext(ctx).Save(wf).Error
}

// CreateStep 创建步骤
func (r *WorkflowRepository) CreateStep(ctx context.Context, step *model.WorkflowStep) error {
	return r.db.WithContext(ctx).Create(step).Error
}

// GetStepByID 获取步骤
func (r *WorkflowRepository) GetStepByID(ctx context.Context, id uint) (*model.WorkflowStep, error) {
	var step model.WorkflowStep
	err := r.db.WithContext(ctx).Preload("Assignee").Preload("Agent").First(&step, id).Error
	if err != nil {
		return nil, err
	}
	return &step, nil
}

// GetPendingSteps 获取待执行步骤（按 index 排序）
func (r *WorkflowRepository) GetPendingSteps(ctx context.Context, workflowID uint) ([]model.WorkflowStep, error) {
	var steps []model.WorkflowStep
	err := r.db.WithContext(ctx).
		Where("workflow_id = ? AND status = 'pending'", workflowID).
		Order("step_index ASC").
		Find(&steps).Error
	return steps, err
}

// UpdateStep 更新步骤
func (r *WorkflowRepository) UpdateStep(ctx context.Context, step *model.WorkflowStep) error {
	return r.db.WithContext(ctx).Save(step).Error
}

// ListFeedbackExamples 获取最近的 HITL 反馈示例（已完成的 human 步骤，有 feedback/decision）
func (r *WorkflowRepository) ListFeedbackExamples(ctx context.Context, limit int) ([]model.WorkflowStep, error) {
	var steps []model.WorkflowStep
	err := r.db.WithContext(ctx).
		Where("type = 'human' AND status = 'completed' AND decision IN ?", []string{"approve", "reject", "revise"}).
		Order("created_at DESC").
		Limit(limit).
		Find(&steps).Error
	return steps, err
}

// ListActiveByAgent 获取某 Agent 的待处理步骤
func (r *WorkflowRepository) ListStepsByAssignee(ctx context.Context, userID uint, offset, limit int) ([]model.WorkflowStep, int64, error) {
	var steps []model.WorkflowStep
	var total int64

	query := r.db.WithContext(ctx).Model(&model.WorkflowStep{}).
		Where("assignee_id = ? AND status IN ?", userID, []string{"pending", "running"})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.
		Preload("Assignee").
		Preload("Agent").
		Preload("Workflow", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Ticket")
		}).
		Offset(offset).Limit(limit).
		Order("created_at ASC").
		Find(&steps).Error
	return steps, total, err
}
