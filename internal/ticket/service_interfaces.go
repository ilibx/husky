package ticket

import (
	"context"
	"io"
	"time"

	"github.com/husky/husky/internal/model"
)

type TicketCreator interface {
	CreateTicket(ctx context.Context, req *model.CreateTicketRequest) (*model.TicketResponse, error)
	OnTicketCreated(h TicketCreatedHandler)
}

type TicketProvider interface {
	GetTicket(ctx context.Context, id uint) (*model.Ticket, error)
	ListTickets(ctx context.Context, offset, limit int, filters map[string]interface{}) ([]model.Ticket, int64, error)
}

type TicketUpdater interface {
	UpdateTicket(ctx context.Context, ticket *model.Ticket) error
}

type TicketDeleter interface {
	DeleteTicket(ctx context.Context, id uint) error
}

type TicketOperator interface {
	AssignTicket(ctx context.Context, id, assigneeID uint) error
	AutoAssignTicket(ctx context.Context, id uint) (*uint, error)
	ClaimTicket(ctx context.Context, id, userID uint) error
	UpdateStatus(ctx context.Context, id uint, status string) error
	SetPriority(ctx context.Context, id uint, priority string) error
	SetDueAt(ctx context.Context, id uint, dueAt time.Time) error
	ListOverdue(ctx context.Context) ([]model.Ticket, error)
}

type WatcherService interface {
	WatchTicket(ctx context.Context, ticketID, userID uint) error
	UnwatchTicket(ctx context.Context, ticketID, userID uint) error
	GetWatchers(ctx context.Context, ticketID uint) ([]model.TicketWatcher, error)
}

type CommentService interface {
	AddComment(ctx context.Context, ticketID, userID uint, content string, isInternal bool) (*model.Comment, error)
	ListComments(ctx context.Context, ticketID uint, offset, limit int) ([]model.Comment, int64, error)
}

type AttachmentService interface {
	UploadAttachment(ctx context.Context, ticketID, userID uint, fileName string, file io.Reader) (*model.Attachment, error)
	ListAttachments(ctx context.Context, ticketID uint) ([]model.Attachment, error)
	DeleteAttachment(ctx context.Context, id uint) error
}

type SatisfactionService interface {
	RateTicket(ctx context.Context, ticketID, userID uint, score int, comment string) (*model.Satisfaction, error)
	GetSatisfaction(ctx context.Context, ticketID uint) (*model.Satisfaction, error)
	GetSatisfactionStats(ctx context.Context) (map[string]interface{}, error)
}

type AuditLogService interface {
	ListAuditLogs(ctx context.Context, resourceType string, resourceID uint, offset, limit int) ([]model.AuditLog, int64, error)
}

type NotificationService interface {
	CreateNotification(ctx context.Context, userID uint, nType, title, content string, refID uint, refType string) error
	ListNotifications(ctx context.Context, userID uint, offset, limit int) ([]model.Notification, int64, error)
	GetUnreadCount(ctx context.Context, userID uint) (int64, error)
	MarkNotificationRead(ctx context.Context, id, userID uint) error
	MarkAllNotificationsRead(ctx context.Context, userID uint) error
}

type WebhookService interface {
	SaveWebhookMessage(ctx context.Context, msg *model.WebhookMessageRecord) error
	ListWebhookMessages(ctx context.Context, offset, limit int, channel string) ([]model.WebhookMessageRecord, int64, error)
}

type TagService interface {
	CreateTag(ctx context.Context, name, color string) (*model.Tag, error)
	ListTags(ctx context.Context) ([]model.Tag, error)
	GetTag(ctx context.Context, id uint) (*model.Tag, error)
	UpdateTag(ctx context.Context, tag *model.Tag) error
	DeleteTag(ctx context.Context, id uint) error
	AddTagsToTicket(ctx context.Context, ticketID uint, tagIDs []uint) error
	RemoveTagFromTicket(ctx context.Context, ticketID, tagID uint) error
	GetTicketTags(ctx context.Context, ticketID uint) ([]model.Tag, error)
	UpdateTicketTags(ctx context.Context, ticketID uint, tagIDs []uint) error
}

type AssignConfigService interface {
	GetAssignConfig(ctx context.Context, categoryID *uint) (*model.AssignConfig, error)
	SetAssignConfig(ctx context.Context, cfg *model.AssignConfig) error
}

type RelationService interface {
	CreateTicketRelation(ctx context.Context, ticketID, relatedID uint, relationType string) (*model.TicketRelation, error)
	DeleteTicketRelation(ctx context.Context, id uint) error
	ListTicketRelations(ctx context.Context, ticketID uint) ([]model.TicketRelation, error)
}

type RoleService interface {
	ListRoles(ctx context.Context) ([]model.Role, error)
	GetRole(ctx context.Context, id uint) (*model.Role, error)
	CreateRole(ctx context.Context, role *model.Role) error
	UpdateRole(ctx context.Context, role *model.Role) error
	DeleteRole(ctx context.Context, id uint) error
}

type FieldService interface {
	ListTicketFields(ctx context.Context) ([]model.TicketField, error)
	CreateTicketField(ctx context.Context, req *model.CreateTicketFieldRequest) (*model.TicketField, error)
	UpdateTicketField(ctx context.Context, id uint, req *model.UpdateTicketFieldRequest) (*model.TicketField, error)
	DeleteTicketField(ctx context.Context, id uint) error
	UpdateTicketFieldValues(ctx context.Context, ticketID uint, values []model.TicketFieldValue) error
	GetTicketFieldValues(ctx context.Context, ticketID uint) ([]model.TicketFieldValue, error)
}

type BotConfigService interface {
	GetBotConfig(ctx context.Context, channel string) (*model.BotConfig, error)
	SetBotConfig(ctx context.Context, cfg *model.BotConfig) error
}

type Service interface {
	TicketCreator
	TicketProvider
	TicketUpdater
	TicketDeleter
	TicketOperator
	WatcherService
	CommentService
	AttachmentService
	SatisfactionService
	AuditLogService
	NotificationService
	WebhookService
	TagService
	AssignConfigService
	RelationService
	RoleService
	FieldService
	BotConfigService
}
