package ticket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
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
	CreateTag(ctx context.Context, name, color string) (*model.Tag, error)
	ListTags(ctx context.Context) ([]model.Tag, error)
	GetTag(ctx context.Context, id uint) (*model.Tag, error)
	UpdateTag(ctx context.Context, tag *model.Tag) error
	DeleteTag(ctx context.Context, id uint) error
	AddTagsToTicket(ctx context.Context, ticketID uint, tagIDs []uint) error
	RemoveTagFromTicket(ctx context.Context, ticketID, tagID uint) error
	GetTicketTags(ctx context.Context, ticketID uint) ([]model.Tag, error)
	UpdateTicketTags(ctx context.Context, ticketID uint, tagIDs []uint) error
	GetAssignConfig(ctx context.Context, categoryID *uint) (*model.AssignConfig, error)
	SetAssignConfig(ctx context.Context, cfg *model.AssignConfig) error
	CreateTicketRelation(ctx context.Context, ticketID, relatedID uint, relationType string) (*model.TicketRelation, error)
	DeleteTicketRelation(ctx context.Context, id uint) error
	ListTicketRelations(ctx context.Context, ticketID uint) ([]model.TicketRelation, error)
	ListRoles(ctx context.Context) ([]model.Role, error)
	GetRole(ctx context.Context, id uint) (*model.Role, error)
	CreateRole(ctx context.Context, role *model.Role) error
	UpdateRole(ctx context.Context, role *model.Role) error
	DeleteRole(ctx context.Context, id uint) error
	ListTicketFields(ctx context.Context) ([]model.TicketField, error)
	CreateTicketField(ctx context.Context, req *model.CreateTicketFieldRequest) (*model.TicketField, error)
	UpdateTicketField(ctx context.Context, id uint, req *model.UpdateTicketFieldRequest) (*model.TicketField, error)
	DeleteTicketField(ctx context.Context, id uint) error
	UpdateTicketFieldValues(ctx context.Context, ticketID uint, values []model.TicketFieldValue) error
	GetTicketFieldValues(ctx context.Context, ticketID uint) ([]model.TicketFieldValue, error)
	GetBotConfig(ctx context.Context, channel string) (*model.BotConfig, error)
	SetBotConfig(ctx context.Context, cfg *model.BotConfig) error
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
	ticket, err := s.ticketRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	cfg, err := s.ticketRepo.GetAssignConfig(ctx, &ticket.CategoryID)
	if err != nil || cfg == nil {
		cfg = &model.AssignConfig{Strategy: model.AssignStrategyLeastBusy}
	}

	var agentID *uint
	switch cfg.Strategy {
	case model.AssignStrategyRoundRobin:
		ids, err := s.ticketRepo.ListAgentIDs(ctx)
		if err != nil || len(ids) == 0 {
			return nil, fmt.Errorf("no available agents found")
		}
		idx, _ := s.ticketRepo.IncrementRoundRobin(ctx, cfg.ID, len(ids))
		agentID = &ids[idx]

	case model.AssignStrategySkillBased:
		if ticket.CategoryID > 0 {
			ids, err := s.ticketRepo.FindAgentsBySkill(ctx, ticket.CategoryID)
			if err == nil && len(ids) > 0 {
				agentID = &ids[0]
				break
			}
		}
		fallthrough

	case model.AssignStrategyRandom:
		ids, err := s.ticketRepo.ListAgentIDs(ctx)
		if err != nil || len(ids) == 0 {
			return nil, fmt.Errorf("no available agents found")
		}
		agentID = &ids[rand.Intn(len(ids))]

	default: // least_busy
		id, err := s.ticketRepo.FindLeastBusyAgent(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to find available agent: %w", err)
		}
		agentID = id
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
	now := time.Now()

	// Transition hooks — auto-set timestamps
	switch model.TicketStatus(status) {
	case model.TicketStatusResolved:
		ticket.ResolvedAt = &now
	case model.TicketStatusClosed:
		ticket.ClosedAt = &now
	}
	// Clear ResolvedAt when leaving resolved
	if model.TicketStatus(oldStatus) == model.TicketStatusResolved && status != string(model.TicketStatusResolved) {
		ticket.ResolvedAt = nil
	}

	ticket.Status = status
	if err := s.ticketRepo.Update(ctx, ticket); err != nil {
		return err
	}

	s.logAudit(ctx, 0, "update_status", "ticket", ticket.ID,
		map[string]string{"status": oldStatus},
		map[string]string{"status": status})

	// Notify assignee
	if ticket.AssigneeID != nil {
		s.CreateNotification(ctx, *ticket.AssigneeID, "ticket_status",
			"Status updated: "+ticket.Title,
			"#"+ticket.TicketNo+" status changed from "+oldStatus+" to "+status,
			ticket.ID, "ticket")
	}

	// Notify watchers
	watchers, _ := s.ticketRepo.GetWatchers(ctx, ticket.ID)
	for _, w := range watchers {
		if ticket.AssigneeID != nil && w.UserID == *ticket.AssigneeID {
			continue // already notified above
		}
		s.CreateNotification(ctx, w.UserID, "ticket_status",
			"Ticket status updated: "+ticket.Title,
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

func (s *service) CreateTag(ctx context.Context, name, color string) (*model.Tag, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	tag := &model.Tag{Name: name, Color: color}
	if err := s.ticketRepo.CreateTag(ctx, tag); err != nil {
		return nil, err
	}
	return tag, nil
}

func (s *service) ListTags(ctx context.Context) ([]model.Tag, error) {
	return s.ticketRepo.ListTags(ctx)
}

func (s *service) GetTag(ctx context.Context, id uint) (*model.Tag, error) {
	return s.ticketRepo.GetTag(ctx, id)
}

func (s *service) UpdateTag(ctx context.Context, tag *model.Tag) error {
	if tag.ID == 0 {
		return fmt.Errorf("id is required")
	}
	return s.ticketRepo.UpdateTag(ctx, tag)
}

func (s *service) DeleteTag(ctx context.Context, id uint) error {
	return s.ticketRepo.DeleteTag(ctx, id)
}

func (s *service) AddTagsToTicket(ctx context.Context, ticketID uint, tagIDs []uint) error {
	return s.ticketRepo.AddTagsToTicket(ctx, ticketID, tagIDs)
}

func (s *service) RemoveTagFromTicket(ctx context.Context, ticketID, tagID uint) error {
	return s.ticketRepo.RemoveTagFromTicket(ctx, ticketID, tagID)
}

func (s *service) GetTicketTags(ctx context.Context, ticketID uint) ([]model.Tag, error) {
	return s.ticketRepo.GetTicketTags(ctx, ticketID)
}

func (s *service) UpdateTicketTags(ctx context.Context, ticketID uint, tagIDs []uint) error {
	return s.ticketRepo.UpdateTicketTags(ctx, ticketID, tagIDs)
}

func (s *service) GetAssignConfig(ctx context.Context, categoryID *uint) (*model.AssignConfig, error) {
	return s.ticketRepo.GetAssignConfig(ctx, categoryID)
}

func (s *service) SetAssignConfig(ctx context.Context, cfg *model.AssignConfig) error {
	return s.ticketRepo.SetAssignConfig(ctx, cfg)
}

func (s *service) CreateTicketRelation(ctx context.Context, ticketID, relatedID uint, relationType string) (*model.TicketRelation, error) {
	if ticketID == relatedID {
		return nil, fmt.Errorf("cannot relate a ticket to itself")
	}
	rel := &model.TicketRelation{
		TicketID:     ticketID,
		RelatedID:    relatedID,
		RelationType: relationType,
	}
	if err := s.ticketRepo.CreateTicketRelation(ctx, rel); err != nil {
		return nil, err
	}
	return rel, nil
}

func (s *service) DeleteTicketRelation(ctx context.Context, id uint) error {
	return s.ticketRepo.DeleteTicketRelation(ctx, id)
}

func (s *service) ListTicketRelations(ctx context.Context, ticketID uint) ([]model.TicketRelation, error) {
	return s.ticketRepo.ListTicketRelations(ctx, ticketID)
}

func (s *service) ListRoles(ctx context.Context) ([]model.Role, error) {
	return s.ticketRepo.ListRoles(ctx)
}

func (s *service) GetRole(ctx context.Context, id uint) (*model.Role, error) {
	return s.ticketRepo.GetRole(ctx, id)
}

func (s *service) CreateRole(ctx context.Context, role *model.Role) error {
	return s.ticketRepo.CreateRole(ctx, role)
}

func (s *service) UpdateRole(ctx context.Context, role *model.Role) error {
	return s.ticketRepo.UpdateRole(ctx, role)
}

func (s *service) DeleteRole(ctx context.Context, id uint) error {
	return s.ticketRepo.DeleteRole(ctx, id)
}

func (s *service) ListTicketFields(ctx context.Context) ([]model.TicketField, error) {
	return s.ticketRepo.ListTicketFields(ctx)
}

func (s *service) CreateTicketField(ctx context.Context, req *model.CreateTicketFieldRequest) (*model.TicketField, error) {
	field := &model.TicketField{
		Name:        req.Name,
		FieldKey:    req.FieldKey,
		FieldType:   req.FieldType,
		Options:     req.Options,
		Required:    req.Required,
		SortOrder:   req.SortOrder,
		Placeholder: req.Placeholder,
	}
	if err := s.ticketRepo.CreateTicketField(ctx, field); err != nil {
		return nil, err
	}
	return field, nil
}

func (s *service) UpdateTicketField(ctx context.Context, id uint, req *model.UpdateTicketFieldRequest) (*model.TicketField, error) {
	existing, err := s.ticketRepo.GetTicketField(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("ticket field not found")
	}
	if req.Name != nil { existing.Name = *req.Name }
	if req.FieldType != nil { existing.FieldType = *req.FieldType }
	if req.Options != nil { existing.Options = *req.Options }
	if req.Required != nil { existing.Required = *req.Required }
	if req.SortOrder != nil { existing.SortOrder = *req.SortOrder }
	if req.Placeholder != nil { existing.Placeholder = *req.Placeholder }
	if req.Enabled != nil { existing.Enabled = *req.Enabled }
	if err := s.ticketRepo.UpdateTicketField(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *service) DeleteTicketField(ctx context.Context, id uint) error {
	return s.ticketRepo.DeleteTicketField(ctx, id)
}

func (s *service) UpdateTicketFieldValues(ctx context.Context, ticketID uint, values []model.TicketFieldValue) error {
	return s.ticketRepo.SetTicketFieldValues(ctx, ticketID, values)
}

func (s *service) GetTicketFieldValues(ctx context.Context, ticketID uint) ([]model.TicketFieldValue, error) {
	return s.ticketRepo.GetTicketFieldValues(ctx, ticketID)
}

func (s *service) GetBotConfig(ctx context.Context, channel string) (*model.BotConfig, error) {
	return s.ticketRepo.GetBotConfig(ctx, channel)
}

func (s *service) SetBotConfig(ctx context.Context, cfg *model.BotConfig) error {
	return s.ticketRepo.SetBotConfig(ctx, cfg)
}
