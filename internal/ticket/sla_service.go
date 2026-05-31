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
	"net/smtp"
	"strings"
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

// PushMethod 推送方式（从 system_config 读取）
type PushMethod struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"` // webhook, email
	Level    string            `json:"level"`
	Template string            `json:"template"`
	Enabled  bool              `json:"enabled"`
	Params   map[string]string `json:"params"`
}

type SLAConfigService struct {
	repo         *repository.SLAConfigRepository
	webhookRepo  *repository.WebhookConfigRepository
	sysCfgRepo   *repository.SystemConfigRepository
	eventHandler SLAEventHandler
	ticketRepo   *repository.TicketRepository
}

func NewSLAConfigService(repo *repository.SLAConfigRepository, webhookRepo *repository.WebhookConfigRepository, sysCfgRepo *repository.SystemConfigRepository, eventHandler SLAEventHandler, ticketRepo *repository.TicketRepository) *SLAConfigService {
	return &SLAConfigService{repo: repo, webhookRepo: webhookRepo, sysCfgRepo: sysCfgRepo, eventHandler: eventHandler, ticketRepo: ticketRepo}
}

func (s *SLAConfigService) List(ctx context.Context, offset, limit int, keyword string) ([]model.SLAConfig, int64, error) {
	return s.repo.List(ctx, offset, limit, keyword)
}

func (s *SLAConfigService) ListAll(ctx context.Context) ([]model.SLAConfig, error) {
	return s.repo.ListAll(ctx)
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
func (s *SLAConfigService) FindMatch(ctx context.Context, priority string, categoryID *uint) (*model.SLAConfig, error) {
	return s.repo.FindMatch(ctx, priority, categoryID)
}

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

// templateVars 返回模板变量映射
func templateVars(ticket *model.Ticket, eventType string) map[string]string {
	vars := map[string]string{
		"ticket_id":   fmt.Sprintf("%d", ticket.ID),
		"ticket_no":   ticket.TicketNo,
		"title":       ticket.Title,
		"status":      ticket.Status,
		"priority":    ticket.Priority,
		"event_type":  eventType,
		"assignee_id": "",
	}
	if ticket.AssigneeID != nil {
		vars["assignee_id"] = fmt.Sprintf("%d", *ticket.AssigneeID)
	}
	if ticket.DueAt != nil {
		vars["due_at"] = ticket.DueAt.Format(time.RFC3339)
	}
	return vars
}

// renderTemplate 渲染模板变量
func renderTemplate(tmpl string, vars map[string]string) string {
	result := tmpl
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{"+k+"}", v)
	}
	return result
}

// FireSLAEvent 触发 SLA 事件（创建通知 + 推送方式 + 渠道消息）
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

	vars := templateVars(ticket, eventType)

	// 1. 内部通知
	title := fmt.Sprintf("SLA %s: %s", eventType, ticket.TicketNo)
	content := fmt.Sprintf("工单 #%s（%s）SLA %s", ticket.TicketNo, ticket.Title, eventType)
	if eventType == "sla_breach" {
		content = fmt.Sprintf("工单 #%s（%s）SLA 已超期，请立即处理。", ticket.TicketNo, ticket.Title)
	}

	if ticket.AssigneeID != nil {
		s.createNotification(ctx, *ticket.AssigneeID, title, content, ticket.ID)
	}

	// 2. 推送管理中的推送方式（按等级匹配）
	hasPushMethods := s.firePushMethods(ctx, ticket, eventType, vars)

	// 3. 旧的 Webhook 配置（兼容 —— 仅在未使用推送管理时触发）
	if !hasPushMethods {
		s.fireWebhooks(ctx, event)
	}

	// 4. 渠道消息（处理人）
	if s.eventHandler != nil && ticket.AssigneeID != nil {
		channel := ticket.Source
		if channel == "" || channel == "web" || channel == "api" {
			channel = "lark"
		}
		msg := fmt.Sprintf("⚠️ SLA 告警\n工单：%s\n标题：%s\n状态：%s\n事件：%s", ticket.TicketNo, ticket.Title, ticket.Status, eventType)
		s.eventHandler.SendToChannel(ctx, channel, fmt.Sprintf("%d", *ticket.AssigneeID), msg)
	}
}

// firePushMethods 从推送管理读取匹配的推送方式并执行，返回是否有推送方式被触发
func (s *SLAConfigService) firePushMethods(ctx context.Context, ticket *model.Ticket, eventType string, vars map[string]string) bool {
	configs, err := s.sysCfgRepo.ListByCategory(ctx, "push")
	if err != nil {
		log.Printf("SLA push: failed to list push configs: %v", err)
		return false
	}

	// 获取 SLA 配置中的预警等级
	slaCfg, _ := s.repo.FindMatch(ctx, ticket.Priority, &ticket.CategoryID)
	warningLevel := ""
	if slaCfg != nil {
		warningLevel = slaCfg.WarningLevel
	}

	fired := false
	for _, cfg := range configs {
		if !strings.HasPrefix(cfg.Key, "method:") {
			continue
		}
		var method PushMethod
		if err := json.Unmarshal([]byte(cfg.Value), &method); err != nil {
			continue
		}
		if !method.Enabled {
			continue
		}
		// 按等级匹配：推送方式的 level 需匹配 SLA 的 warning_level
		if method.Level != "" && warningLevel != "" && method.Level != warningLevel {
			continue
		}

		rendered := renderTemplate(method.Template, vars)
		switch method.Type {
		case "webhook":
			go s.sendPushWebhook(method, rendered)
		case "email":
			go s.sendPushEmail(method, rendered)
		default:
			log.Printf("SLA push: unknown push type %q for method %q", method.Type, method.Name)
		}
		fired = true
	}
	return fired
}

// sendPushWebhook 执行 Webhook 推送
func (s *SLAConfigService) sendPushWebhook(method PushMethod, body string) {
	url := method.Params["url"]
	if url == "" {
		log.Printf("SLA push: webhook %q has no URL", method.Name)
		return
	}
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Event-Type", "sla")

	if secret := method.Params["secret"]; secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(body))
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Signature", sig)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("SLA push: webhook %q failed: %v", method.Name, err)
		return
	}
	resp.Body.Close()
}

// sendPushEmail 执行邮件推送
func (s *SLAConfigService) sendPushEmail(method PushMethod, body string) {
	host := method.Params["host"]
	port := method.Params["port"]
	user := method.Params["user"]
	password := method.Params["password"]
	from := method.Params["from"]
	if host == "" || user == "" || from == "" {
		log.Printf("SLA push: email %q has incomplete SMTP config", method.Name)
		return
	}
	if port == "" {
		port = "587"
	}

	addr := host + ":" + port
	auth := smtp.PlainAuth("", user, password, host)

	to := method.Params["to"]
	if to == "" {
		to = from // fallback: send to self
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: SLA 告警\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", from, to, body)

	if err := smtp.SendMail(addr, auth, from, strings.Split(to, ","), []byte(msg)); err != nil {
		log.Printf("SLA push: email %q failed: %v", method.Name, err)
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
	resp.Body.Close()
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
