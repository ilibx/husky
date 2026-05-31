package ticket

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
)

// SLAEventHandler 外部依赖
type SLAEventHandler interface {
	SendToChannel(ctx context.Context, channel, targetID, content string) error
}

// SLAEventAdapterFunc is a functional adapter for SLAEventHandler.
type SLAEventAdapterFunc func(ctx context.Context, channel, targetID, content string) error

func (f SLAEventAdapterFunc) SendToChannel(ctx context.Context, channel, targetID, content string) error {
	return f(ctx, channel, targetID, content)
}

type SLAConfigService struct {
	repo       *repository.SLAConfigRepository
	webhookRepo *repository.WebhookConfigRepository
	eventHandler SLAEventHandler
	ticketRepo   *repository.TicketRepository
}

func NewSLAConfigService(repo *repository.SLAConfigRepository, webhookRepo *repository.WebhookConfigRepository, eventHandler SLAEventHandler, ticketRepo *repository.TicketRepository) *SLAConfigService {
	return &SLAConfigService{repo: repo, webhookRepo: webhookRepo, eventHandler: eventHandler, ticketRepo: ticketRepo}
}

func (s *SLAConfigService) List(ctx context.Context) ([]model.SLAConfig, error) {
	return s.repo.List(ctx)
}

func (s *SLAConfigService) GetByID(ctx context.Context, id uint) (*model.SLAConfig, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SLAConfigService) Create(ctx context.Context, c *model.SLAConfig) error {
	if c.WarningThreshold <= 0 || c.WarningThreshold > 1 {
		c.WarningThreshold = 0.8
	}
	if c.ResponseMinutes <= 0 {
		c.ResponseMinutes = 60
	}
	if c.ResolutionMinutes <= 0 {
		c.ResolutionMinutes = 480
	}
	return s.repo.Create(ctx, c)
}

func (s *SLAConfigService) Update(ctx context.Context, c *model.SLAConfig) error {
	return s.repo.Update(ctx, c)
}

func (s *SLAConfigService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

// ComputeDueAt 根据 SLA 规则计算工单 DueAt
func (s *SLAConfigService) ComputeDueAt(ctx context.Context, ticket *model.Ticket) *time.Time {
	cfg, err := s.repo.FindMatch(ctx, ticket.Priority, &ticket.CategoryID)
	if err != nil || cfg == nil {
		return nil
	}
	now := time.Now()
	due := now.Add(time.Duration(cfg.ResolutionMinutes) * time.Minute)
	return &due
}

// SLAEvent SLA 事件
type SLAEvent struct {
	Type       string `json:"type"`        // sla_warning, sla_breach, sla_restored
	TicketNo   string `json:"ticket_no"`
	TicketID   uint   `json:"ticket_id"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	Priority   string `json:"priority"`
	AssigneeID *uint  `json:"assignee_id,omitempty"`
	DueAt      string `json:"due_at,omitempty"`
	Timestamp  string `json:"timestamp"`
}

// FireSLAEvent 触发 SLA 事件（创建通知 + Webhook + 渠道消息）
func (s *SLAConfigService) FireSLAEvent(ctx context.Context, ticket *model.Ticket, eventType string) {
	event := SLAEvent{
		Type:       eventType,
		TicketNo:   ticket.TicketNo,
		TicketID:   ticket.ID,
		Title:      ticket.Title,
		Status:     ticket.Status,
		Priority:   ticket.Priority,
		AssigneeID: ticket.AssigneeID,
		Timestamp:  time.Now().Format(time.RFC3339),
	}
	if ticket.DueAt != nil {
		event.DueAt = ticket.DueAt.Format(time.RFC3339)
	}

	// 1. 内部通知
	title := fmt.Sprintf("SLA %s: %s", eventType, ticket.TicketNo)
	content := fmt.Sprintf("工单 #%s（%s）SLA %s", ticket.TicketNo, ticket.Title, eventType)
	if eventType == "sla_breach" {
		content = fmt.Sprintf("工单 #%s（%s）SLA 已超期，请立即处理。", ticket.TicketNo, ticket.Title)
	}

	if ticket.AssigneeID != nil {
		s.createNotification(ctx, *ticket.AssigneeID, title, content, ticket.ID)
	}

	// 2. Webhook 推送
	s.fireWebhooks(ctx, event)

	// 3. 渠道消息（处理人）- use the ticket source channel if available
	if s.eventHandler != nil && ticket.AssigneeID != nil {
		channel := ticket.Source
		if channel == "" || channel == "web" || channel == "api" {
			channel = "lark"
		}
		msg := fmt.Sprintf("⚠️ SLA 告警\n工单：%s\n标题：%s\n状态：%s\n事件：%s", ticket.TicketNo, ticket.Title, ticket.Status, eventType)
		s.eventHandler.SendToChannel(ctx, channel, fmt.Sprintf("%d", *ticket.AssigneeID), msg)
	}
}

func (s *SLAConfigService) fireWebhooks(ctx context.Context, event SLAEvent) {
	webhooks, err := s.webhookRepo.FindByEvent(ctx, event.Type)
	if err != nil {
		log.Printf("SLA webhook: failed to find webhooks for event %q: %v", event.Type, err)
		return
	}
	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("SLA webhook: failed to marshal event: %v", err)
		return
	}
	for _, wh := range webhooks {
		go s.sendWebhook(wh, body)
	}
}

func (s *SLAConfigService) sendWebhook(wh model.WebhookConfig, body []byte) {
	req, err := http.NewRequest(http.MethodPost, wh.URL, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Event-Type", "sla")

	if wh.Secret != "" {
		mac := hmac.New(sha256.New, []byte(wh.Secret))
		mac.Write(body)
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Signature", sig)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("SLA webhook: failed to send to %s: %v", wh.URL, err)
		return
	}
	if err := resp.Body.Close(); err != nil {
		log.Printf("SLA webhook: failed to close response body: %v", err)
	}
}

func (s *SLAConfigService) createNotification(ctx context.Context, userID uint, title, content string, ticketID uint) {
	if err := s.ticketRepo.CreateNotification(ctx, &model.Notification{
		UserID:        userID,
		Type:          "sla",
		Title:         title,
		Content:       content,
		ReferenceID:   ticketID,
		ReferenceType: "ticket",
		Status:        "unread",
	}); err != nil {
		log.Printf("SLA: failed to create notification: %v", err)
	}
}
