package ticket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
	"github.com/husky/husky/pkg/errors"
)

type Service interface {
	CreateTicket(ctx context.Context, req *model.CreateTicketRequest) (*model.TicketResponse, error)
	GetTicket(ctx context.Context, id uint) (*model.Ticket, error)
	ListTickets(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]model.Ticket, int64, error)
	UpdateTicket(ctx context.Context, ticket *model.Ticket) error
	DeleteTicket(ctx context.Context, id uint) error
	AssignTicket(ctx context.Context, id, assigneeID uint) error
	AutoAssignTicket(ctx context.Context, id uint) (*uint, error)
	ClaimTicket(ctx context.Context, id, userID uint) error
	UpdateStatus(ctx context.Context, id uint, status string) error
	AddComment(ctx context.Context, ticketID, userID uint, content string, isInternal bool) (*model.Comment, error)
	ListComments(ctx context.Context, ticketID uint, offset, limit int) ([]model.Comment, int64, error)
	UploadAttachment(ctx context.Context, ticketID, userID uint, fileName string, file io.Reader) (*model.Attachment, error)
	ListAttachments(ctx context.Context, ticketID uint) ([]model.Attachment, error)
	DeleteAttachment(ctx context.Context, id uint) error
	RateTicket(ctx context.Context, ticketID, userID uint, score int, comment string) (*model.Satisfaction, error)
	GetSatisfaction(ctx context.Context, ticketID uint) (*model.Satisfaction, error)
	GetSatisfactionStats(ctx context.Context) (map[string]interface{}, error)
	SetDueAt(ctx context.Context, id uint, dueAt time.Time) error
	ListOverdue(ctx context.Context) ([]model.Ticket, error)
	CreateNotification(ctx context.Context, userID uint, nType, title, content string, refID uint, refType string) error
	ListNotifications(ctx context.Context, userID uint, offset, limit int) ([]model.Notification, int64, error)
	GetUnreadCount(ctx context.Context, userID uint) (int64, error)
	MarkNotificationRead(ctx context.Context, id, userID uint) error
	MarkAllNotificationsRead(ctx context.Context, userID uint) error
	SaveWebhookMessage(ctx context.Context, msg *model.WebhookMessageRecord) error
	ListWebhookMessages(ctx context.Context, offset, limit int, channel string) ([]model.WebhookMessageRecord, int64, error)
	WatchTicket(ctx context.Context, ticketID, userID uint) error
	UnwatchTicket(ctx context.Context, ticketID, userID uint) error
	GetWatchers(ctx context.Context, ticketID uint) ([]model.TicketWatcher, error)
	ListAuditLogs(ctx context.Context, resourceType string, resourceID uint, offset, limit int) ([]model.AuditLog, int64, error)
}

type service struct {
	ticketRepo *repository.TicketRepository
	auditSvc   *AuditService
}

func NewService(ticketRepo *repository.TicketRepository) Service {
	return &service{
		ticketRepo: ticketRepo,
		auditSvc:   NewAuditService(ticketRepo),
	}
}

func (s *service) CreateTicket(ctx context.Context, req *model.CreateTicketRequest) (*model.TicketResponse, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.Description == "" {
		return nil, fmt.Errorf("description is required")
	}
	if req.Priority != "" && !model.IsValidPriority(req.Priority) {
		return nil, fmt.Errorf("invalid priority: %s", req.Priority)
	}

	ticketNo := fmt.Sprintf("TK%s%s", time.Now().Format("20060102"), strings.ToUpper(uuid.New().String()[:6]))

	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}

	ticket := &model.Ticket{
		TicketNo:    ticketNo,
		Title:       req.Title,
		Description: req.Description,
		Priority:    priority,
		Status:      string(model.TicketStatusOpen),
		Source:      req.Channel,
	}

	if req.RequesterID != "" {
		if id, err := strconv.ParseUint(req.RequesterID, 10, 64); err == nil {
			ticket.RequesterID = uint(id)
		}
	}

	if err := s.ticketRepo.Create(ctx, ticket); err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	s.logAudit(ctx, ticket.RequesterID, "create", "ticket", ticket.ID, nil, ticket)

	return &model.TicketResponse{
		ID:          ticket.ID,
		TicketNo:    ticket.TicketNo,
		Title:       ticket.Title,
		Status:      ticket.Status,
		Priority:    ticket.Priority,
		Source:      ticket.Source,
		CreatedAt:   ticket.CreatedAt,
		UpdatedAt:   ticket.UpdatedAt,
		RequesterID: ticket.RequesterID,
		AssigneeID:  ticket.AssigneeID,
	}, nil
}

func (s *service) GetTicket(ctx context.Context, id uint) (*model.Ticket, error) {
	return s.ticketRepo.GetByID(ctx, id)
}

func (s *service) ListTickets(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]model.Ticket, int64, error) {
	return s.ticketRepo.List(ctx, offset, limit, filters)
}

func (s *service) UpdateTicket(ctx context.Context, ticket *model.Ticket) error {
	existing, err := s.ticketRepo.GetByID(ctx, ticket.ID)
	if err != nil {
		return err
	}

	if ticket.Status != "" && ticket.Status != existing.Status {
		if !model.IsValidTransition(model.TicketStatus(existing.Status), model.TicketStatus(ticket.Status)) {
			return errors.New(errors.ErrInvalidTransition, fmt.Sprintf("cannot transition from %s to %s", existing.Status, ticket.Status))
		}
	}
	if ticket.Priority != "" && !model.IsValidPriority(ticket.Priority) {
		return fmt.Errorf("invalid priority: %s", ticket.Priority)
	}

	return s.ticketRepo.Update(ctx, ticket)
}

func (s *service) DeleteTicket(ctx context.Context, id uint) error {
	return s.ticketRepo.Delete(ctx, id)
}

func (s *service) AutoAssignTicket(ctx context.Context, id uint) (*uint, error) {
	agentID, err := s.ticketRepo.FindLeastBusyAgent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find available agent: %w", err)
	}
	if agentID == nil {
		return nil, fmt.Errorf("no available agents found")
	}
	if err := s.AssignTicket(ctx, id, *agentID); err != nil {
		return nil, err
	}
	return agentID, nil
}

func (s *service) AssignTicket(ctx context.Context, id, assigneeID uint) error {
	ticket, err := s.ticketRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	previousAssignee := ticket.AssigneeID
	ticket.AssigneeID = &assigneeID
	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return err
	}

	s.logAudit(ctx, assigneeID, "assign", "ticket", ticket.ID,
		map[string]interface{}{"assignee_id": previousAssignee},
		map[string]interface{}{"assignee_id": assigneeID})

	s.CreateNotification(ctx, assigneeID, "ticket_assigned",
		"New ticket assigned: "+ticket.Title,
		"#"+ticket.TicketNo+" has been assigned to you.\nPriority: "+ticket.Priority,
		ticket.ID, "ticket")
	return nil
}

func (s *service) ClaimTicket(ctx context.Context, id, userID uint) error {
	ticket, err := s.ticketRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if ticket.AssigneeID != nil && *ticket.AssigneeID != userID {
		return errors.New(errors.ErrForbidden, "ticket already assigned to another user")
	}
	previousAssignee := ticket.AssigneeID
	ticket.AssigneeID = &userID
	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return err
	}

	if previousAssignee != nil && *previousAssignee != userID {
		s.CreateNotification(ctx, *previousAssignee, "ticket_assigned",
			"Ticket re-assigned: "+ticket.Title,
			"#"+ticket.TicketNo+" has been re-assigned to another user.",
			ticket.ID, "ticket")
	}
	s.CreateNotification(ctx, userID, "ticket_assigned",
		"Ticket claimed: "+ticket.Title,
		"#"+ticket.TicketNo+" has been claimed by you.\nPriority: "+ticket.Priority,
		ticket.ID, "ticket")
	return nil
}

func (s *service) UpdateStatus(ctx context.Context, id uint, status string) error {
	ticket, err := s.ticketRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if !model.IsValidTransition(model.TicketStatus(ticket.Status), model.TicketStatus(status)) {
		return errors.New(errors.ErrInvalidTransition, fmt.Sprintf("cannot transition from %s to %s", ticket.Status, status))
	}

	oldStatus := ticket.Status
	ticket.Status = status
	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return err
	}

	s.logAudit(ctx, 0, "update_status", "ticket", ticket.ID,
		map[string]string{"status": oldStatus},
		map[string]string{"status": status})

	if ticket.AssigneeID != nil {
		s.CreateNotification(ctx, *ticket.AssigneeID, "ticket_status",
			"Status updated: "+ticket.Title,
			"#"+ticket.TicketNo+" status changed from "+oldStatus+" to "+status,
			ticket.ID, "ticket")
	}
	return nil
}

func (s *service) AddComment(ctx context.Context, ticketID, userID uint, content string, isInternal bool) (*model.Comment, error) {
	if content == "" {
		return nil, fmt.Errorf("comment content is required")
	}

	comment := &model.Comment{
		TicketID:   ticketID,
		UserID:     userID,
		Content:    content,
		IsInternal: isInternal,
	}

	if err := s.ticketRepo.AddComment(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}

	ticket, _ := s.ticketRepo.GetByID(ctx, ticketID)
	if ticket != nil {
		if ticket.AssigneeID != nil && *ticket.AssigneeID != userID {
			s.CreateNotification(ctx, *ticket.AssigneeID, "ticket_comment",
				"New comment on: "+ticket.Title,
				content,
				ticket.ID, "ticket")
		}
		watchers, _ := s.ticketRepo.GetWatchers(ctx, ticketID)
		for _, w := range watchers {
			if w.UserID != userID && (ticket.AssigneeID == nil || w.UserID != *ticket.AssigneeID) {
				s.CreateNotification(ctx, w.UserID, "ticket_comment",
					"New comment on: "+ticket.Title,
					content,
					ticket.ID, "ticket")
			}
		}
	}

	return comment, nil
}

func (s *service) ListComments(ctx context.Context, ticketID uint, offset, limit int) ([]model.Comment, int64, error) {
	return s.ticketRepo.ListComments(ctx, ticketID, offset, limit)
}

func (s *service) UploadAttachment(ctx context.Context, ticketID, userID uint, fileName string, file io.Reader) (*model.Attachment, error) {
	if _, err := s.ticketRepo.GetByID(ctx, ticketID); err != nil {
		return nil, fmt.Errorf("ticket not found: %w", err)
	}

	ext := filepath.Ext(fileName)
	storageName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload dir: %w", err)
	}
	filePath := filepath.Join(uploadDir, storageName)

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	contentType := detectContentType(fileName)

	att := &model.Attachment{
		TicketID:   ticketID,
		FileName:   fileName,
		FileSize:   written,
		FileType:   contentType,
		FileURL:    filePath,
		UploadedBy: userID,
	}

	if err := s.ticketRepo.CreateAttachment(ctx, att); err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save attachment record: %w", err)
	}

	return att, nil
}

func (s *service) ListAttachments(ctx context.Context, ticketID uint) ([]model.Attachment, error) {
	return s.ticketRepo.ListAttachments(ctx, ticketID)
}

func (s *service) DeleteAttachment(ctx context.Context, id uint) error {
	att, err := s.ticketRepo.GetAttachment(ctx, id)
	if err != nil {
		return err
	}
	// Delete DB record first, then file — avoids orphaned records on file-delete-but-DB-fail
	if err := s.ticketRepo.DeleteAttachment(ctx, id); err != nil {
		return err
	}
	os.Remove(att.FileURL)
	return nil
}

func (s *service) RateTicket(ctx context.Context, ticketID, userID uint, score int, comment string) (*model.Satisfaction, error) {
	if score < 1 || score > 5 {
		return nil, fmt.Errorf("score must be between 1 and 5")
	}

	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket.Status != string(model.TicketStatusResolved) && ticket.Status != string(model.TicketStatusClosed) {
		return nil, fmt.Errorf("can only rate resolved or closed tickets")
	}

	existing, _ := s.ticketRepo.GetSatisfactionByTicket(ctx, ticketID)
	if existing != nil {
		return nil, fmt.Errorf("ticket already rated")
	}

	sat := &model.Satisfaction{
		TicketID:  ticketID,
		Score:     score,
		Comment:   comment,
		CreatedBy: userID,
	}
	if err := s.ticketRepo.CreateSatisfaction(ctx, sat); err != nil {
		return nil, err
	}
	return sat, nil
}

func (s *service) GetSatisfaction(ctx context.Context, ticketID uint) (*model.Satisfaction, error) {
	return s.ticketRepo.GetSatisfactionByTicket(ctx, ticketID)
}

func (s *service) GetSatisfactionStats(ctx context.Context) (map[string]interface{}, error) {
	return s.ticketRepo.GetSatisfactionStats(ctx)
}

func (s *service) SetDueAt(ctx context.Context, id uint, dueAt time.Time) error {
	if err := s.ticketRepo.SetTicketDueAt(ctx, id, dueAt); err != nil {
		return err
	}

	ticket, _ := s.ticketRepo.GetByID(ctx, id)
	if ticket != nil && ticket.AssigneeID != nil {
		s.CreateNotification(ctx, *ticket.AssigneeID, "ticket_status",
			"SLA deadline set: "+ticket.Title,
			"#"+ticket.TicketNo+" must be resolved by "+dueAt.Format("2006-01-02 15:04"),
			ticket.ID, "ticket")
	}
	return nil
}

func (s *service) ListOverdue(ctx context.Context) ([]model.Ticket, error) {
	var tickets []model.Ticket
	err := s.ticketRepo.GetDB().WithContext(ctx).
		Where("due_at IS NOT NULL AND due_at < NOW() AND status NOT IN ('resolved', 'closed')").
		Order("due_at ASC").
		Find(&tickets).Error
	return tickets, err
}

func (s *service) logAudit(ctx context.Context, userID uint, action, resource string, resourceID uint, oldVal, newVal interface{}) {
	oldJSON, _ := json.Marshal(oldVal)
	newJSON, _ := json.Marshal(newVal)
	audit := &model.AuditLog{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		OldValue:   string(oldJSON),
		NewValue:   string(newJSON),
	}
	s.ticketRepo.CreateAuditLog(ctx, audit)
}

func (s *service) WatchTicket(ctx context.Context, ticketID, userID uint) error {
	return s.ticketRepo.AddWatcher(ctx, ticketID, userID)
}

func (s *service) UnwatchTicket(ctx context.Context, ticketID, userID uint) error {
	return s.ticketRepo.RemoveWatcher(ctx, ticketID, userID)
}

func (s *service) GetWatchers(ctx context.Context, ticketID uint) ([]model.TicketWatcher, error) {
	return s.ticketRepo.GetWatchers(ctx, ticketID)
}

func (s *service) SaveWebhookMessage(ctx context.Context, msg *model.WebhookMessageRecord) error {
	return s.ticketRepo.CreateWebhookMessage(ctx, msg)
}

func (s *service) ListAuditLogs(ctx context.Context, resourceType string, resourceID uint, offset, limit int) ([]model.AuditLog, int64, error) {
	return s.ticketRepo.ListAuditLogs(ctx, resourceType, resourceID, offset, limit)
}

func (s *service) ListWebhookMessages(ctx context.Context, offset, limit int, channel string) ([]model.WebhookMessageRecord, int64, error) {
	return s.ticketRepo.ListWebhookMessages(ctx, offset, limit, channel)
}

func (s *service) CreateNotification(ctx context.Context, userID uint, nType, title, content string, refID uint, refType string) error {
	n := &model.Notification{
		UserID:        userID,
		Type:          nType,
		Title:         title,
		Content:       content,
		ReferenceID:   refID,
		ReferenceType: refType,
		Status:        "unread",
	}
	return s.ticketRepo.CreateNotification(ctx, n)
}

func (s *service) ListNotifications(ctx context.Context, userID uint, offset, limit int) ([]model.Notification, int64, error) {
	return s.ticketRepo.ListNotifications(ctx, userID, offset, limit)
}

func (s *service) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	return s.ticketRepo.GetUnreadNotificationCount(ctx, userID)
}

func (s *service) MarkNotificationRead(ctx context.Context, id, userID uint) error {
	return s.ticketRepo.MarkNotificationRead(ctx, id, userID)
}

func (s *service) MarkAllNotificationsRead(ctx context.Context, userID uint) error {
	return s.ticketRepo.MarkAllNotificationsRead(ctx, userID)
}

func detectContentType(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".pdf":
		return "application/pdf"
	case ".doc", ".docx":
		return "application/msword"
	case ".xls", ".xlsx":
		return "application/vnd.ms-excel"
	case ".zip":
		return "application/zip"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
