package ticket

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"husky/internal/model"
	"husky/internal/repository"
	"husky/pkg/errors/herr"
)

// Service 工单服务接口
type Service interface {
	// CreateTicket 创建工单
	CreateTicket(ctx context.Context, req *CreateTicketRequest) (*model.Ticket, error)
	// GetTicket 获取工单详情
	GetTicket(ctx context.Context, id string) (*model.Ticket, error)
	// ListTickets 查询工单列表
	ListTickets(ctx context.Context, req *ListTicketsRequest) (*ListTicketsResponse, error)
	// UpdateTicket 更新工单
	UpdateTicket(ctx context.Context, id string, req *UpdateTicketRequest) (*model.Ticket, error)
	// DeleteTicket 删除工单
	DeleteTicket(ctx context.Context, id string) error
	// AssignTicket 分配工单
	AssignTicket(ctx context.Context, id string, assigneeID string, operatorID string) (*model.Ticket, error)
	// UpdateStatus 更新工单状态
	UpdateStatus(ctx context.Context, id string, status model.TicketStatus, operatorID string, comment string) (*model.Ticket, error)
	// AddComment 添加工单评论
	AddComment(ctx context.Context, ticketID string, req *AddCommentRequest) (*model.Comment, error)
}

// CreateTicketRequest 创建工单请求
type CreateTicketRequest struct {
	Title       string            `json:"title" validate:"required,min=1,max=200"`
	Description string            `json:"description" validate:"required,min=1,max=10000"`
	Priority    model.Priority    `json:"priority" validate:"required,oneof=low medium high urgent"`
	CategoryID  string            `json:"category_id"`
	Tags        []string          `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	SubmitterID string            `json:"submitter_id"`
	Channel     string            `json:"channel" validate:"omitempty,oneof=web email lark dingtalk wecom"`
}

// ListTicketsRequest 查询工单列表请求
type ListTicketsRequest struct {
	Status     []model.TicketStatus `json:"status,omitempty"`
	Priority   []model.Priority     `json:"priority,omitempty"`
	AssigneeID string               `json:"assignee_id,omitempty"`
	SubmitterID string              `json:"submitter_id,omitempty"`
	CategoryID string               `json:"category_id,omitempty"`
	Keyword    string               `json:"keyword,omitempty"`
	Page       int                  `json:"page" validate:"min=1"`
	PageSize   int                  `json:"page_size" validate:"min=1,max=100"`
	SortBy     string               `json:"sort_by,omitempty"`
	SortOrder  string               `json:"sort_order,omitempty"`
}

// ListTicketsResponse 查询工单列表响应
type ListTicketsResponse struct {
	Tickets []*model.Ticket `json:"tickets"`
	Total   int64           `json:"total"`
	Page    int             `json:"page"`
	PageSize int            `json:"page_size"`
}

// UpdateTicketRequest 更新工单请求
type UpdateTicketRequest struct {
	Title        string                 `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description  string                 `json:"description,omitempty" validate:"omitempty,min=1,max=10000"`
	Priority     *model.Priority        `json:"priority,omitempty"`
	CategoryID   string                 `json:"category_id,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// AddCommentRequest 添加评论请求
type AddCommentRequest struct {
	Content   string `json:"content" validate:"required,min=1,max=10000"`
	AuthorID  string `json:"author_id" validate:"required"`
	IsInternal bool   `json:"is_internal,omitempty"`
}

type service struct {
	repo repository.TicketRepository
}

// NewService 创建工单服务实例
func NewService(repo repository.TicketRepository) Service {
	return &service{repo: repo}
}

func (s *service) CreateTicket(ctx context.Context, req *CreateTicketRequest) (*model.Ticket, error) {
	if req.SubmitterID == "" {
		return nil, herr.New(herr.ErrInvalidParameter, "submitter_id is required")
	}

	now := time.Now()
	ticket := &model.Ticket{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		Status:      model.TicketStatusOpen,
		CategoryID:  req.CategoryID,
		Tags:        req.Tags,
		CustomFields: req.CustomFields,
		SubmitterID: req.SubmitterID,
		Channel:     req.Channel,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, ticket); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (s *service) GetTicket(ctx context.Context, id string) (*model.Ticket, error) {
	if id == "" {
		return nil, herr.New(herr.ErrInvalidParameter, "ticket id is required")
	}

	ticket, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, herr.New(herr.ErrNotFound, "ticket not found")
		}
		return nil, err
	}

	return ticket, nil
}

func (s *service) ListTickets(ctx context.Context, req *ListTicketsRequest) (*ListTicketsResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	tickets, total, err := s.repo.List(ctx, &repository.ListTicketsOptions{
		Status:      req.Status,
		Priority:    req.Priority,
		AssigneeID:  req.AssigneeID,
		SubmitterID: req.SubmitterID,
		CategoryID:  req.CategoryID,
		Keyword:     req.Keyword,
		Offset:      (req.Page - 1) * req.PageSize,
		Limit:       req.PageSize,
		SortBy:      req.SortBy,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		return nil, err
	}

	return &ListTicketsResponse{
		Tickets:  tickets,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *service) UpdateTicket(ctx context.Context, id string, req *UpdateTicketRequest) (*model.Ticket, error) {
	ticket, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, herr.New(herr.ErrNotFound, "ticket not found")
		}
		return nil, err
	}

	// 更新字段
	if req.Title != "" {
		ticket.Title = req.Title
	}
	if req.Description != "" {
		ticket.Description = req.Description
	}
	if req.Priority != nil {
		ticket.Priority = *req.Priority
	}
	if req.CategoryID != "" {
		ticket.CategoryID = req.CategoryID
	}
	if req.Tags != nil {
		ticket.Tags = req.Tags
	}
	if req.CustomFields != nil {
		ticket.CustomFields = req.CustomFields
	}
	ticket.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, ticket); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (s *service) DeleteTicket(ctx context.Context, id string) error {
	if id == "" {
		return herr.New(herr.ErrInvalidParameter, "ticket id is required")
	}

	// 软删除
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return herr.New(herr.ErrNotFound, "ticket not found")
		}
		return err
	}

	return nil
}

func (s *service) AssignTicket(ctx context.Context, id string, assigneeID string, operatorID string) (*model.Ticket, error) {
	if id == "" {
		return nil, herr.New(herr.ErrInvalidParameter, "ticket id is required")
	}
	if assigneeID == "" {
		return nil, herr.New(herr.ErrInvalidParameter, "assignee_id is required")
	}

	ticket, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, herr.New(herr.ErrNotFound, "ticket not found")
		}
		return nil, err
	}

	// 检查状态是否允许分配
	if !ticket.Status.CanAssign() {
		return nil, herr.New(herr.ErrInvalidState, "ticket cannot be assigned in current status")
	}

	ticket.AssigneeID = assigneeID
	ticket.AssignedAt = time.Now()
	ticket.UpdatedAt = time.Now()

	// 记录操作日志
	ticket.OperationLogs = append(ticket.OperationLogs, model.OperationLog{
		OperatorID: operatorID,
		Action:     "assign",
		Details:    map[string]string{"assignee_id": assigneeID},
		Timestamp:  time.Now(),
	})

	if err := s.repo.Update(ctx, ticket); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (s *service) UpdateStatus(ctx context.Context, id string, status model.TicketStatus, operatorID string, comment string) (*model.Ticket, error) {
	if id == "" {
		return nil, herr.New(herr.ErrInvalidParameter, "ticket id is required")
	}

	ticket, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, herr.New(herr.ErrNotFound, "ticket not found")
		}
		return nil, err
	}

	// 检查状态转换是否合法
	if !ticket.Status.CanTransitionTo(status) {
		return nil, herr.New(herr.ErrInvalidState, "invalid status transition")
	}

	oldStatus := ticket.Status
	ticket.Status = status
	ticket.UpdatedAt = time.Now()

	// 状态变更时的特殊处理
	switch status {
	case model.TicketStatusResolved:
		ticket.ResolvedAt = time.Now()
	case model.TicketStatusClosed:
		ticket.ClosedAt = time.Now()
	}

	// 记录操作日志
	logEntry := model.OperationLog{
		OperatorID: operatorID,
		Action:     "status_change",
		Details: map[string]string{
			"old_status": string(oldStatus),
			"new_status": string(status),
		},
		Timestamp: time.Now(),
	}
	if comment != "" {
		logEntry.Details["comment"] = comment
	}
	ticket.OperationLogs = append(ticket.OperationLogs, logEntry)

	if err := s.repo.Update(ctx, ticket); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (s *service) AddComment(ctx context.Context, ticketID string, req *AddCommentRequest) (*model.Comment, error) {
	if ticketID == "" {
		return nil, herr.New(herr.ErrInvalidParameter, "ticket id is required")
	}

	// 验证工单存在
	_, err := s.repo.GetByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, herr.New(herr.ErrNotFound, "ticket not found")
		}
		return nil, err
	}

	now := time.Now()
	comment := &model.Comment{
		ID:         uuid.New().String(),
		TicketID:   ticketID,
		Content:    req.Content,
		AuthorID:   req.AuthorID,
		IsInternal: req.IsInternal,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.AddComment(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}
