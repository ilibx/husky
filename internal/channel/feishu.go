package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/husky/husky/internal/channel/feishu"
	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
)

type TicketGroupService struct {
	ticketRepo *repository.TicketRepository
	feishuCli  *feishu.Client
	kbRepo     repository.VectorStoreRepository
}

func NewTicketGroupService(ticketRepo *repository.TicketRepository, feishuCli *feishu.Client, kbRepo repository.VectorStoreRepository) *TicketGroupService {
	return &TicketGroupService{
		ticketRepo: ticketRepo,
		feishuCli:  feishuCli,
		kbRepo:     kbRepo,
	}
}

func (s *TicketGroupService) CreateGroupForTicket(ctx context.Context, ticket *model.Ticket) (*model.TicketGroup, error) {
	if s.feishuCli == nil {
		return nil, nil
	}

	existing, _ := s.ticketRepo.GetTicketGroupByTicket(ctx, ticket.ID)
	if existing != nil {
		return existing, nil
	}

	groupName := fmt.Sprintf("Ticket %s: %s", ticket.TicketNo, truncateString(ticket.Title, 40))
	groupDesc := fmt.Sprintf("Ticket #%s\nPriority: %s\nStatus: %s\nCreated: %s",
		ticket.TicketNo, ticket.Priority, ticket.Status, ticket.CreatedAt.Format("2006-01-02 15:04"))

	chatID, name, err := s.feishuCli.CreateChat(ctx, groupName, groupDesc)
	if err != nil {
		return nil, fmt.Errorf("failed to create feishu chat: %w", err)
	}

	joinLink := fmt.Sprintf("https://applink.feishu.cn/client/chat/open?chat_id=%s", chatID)

	tg := &model.TicketGroup{
		TicketID:  ticket.ID,
		GroupID:   chatID,
		GroupName: name,
		JoinLink:  joinLink,
		Status:    "active",
	}
	if err := s.ticketRepo.CreateTicketGroup(ctx, tg); err != nil {
		return nil, fmt.Errorf("failed to save ticket group: %w", err)
	}

	s.sendWelcomeMessage(ctx, chatID, ticket, joinLink)
	return tg, nil
}

func (s *TicketGroupService) AddUserToGroup(ctx context.Context, ticketID uint, feishuOpenID string) error {
	tg, err := s.ticketRepo.GetTicketGroupByTicket(ctx, ticketID)
	if err != nil {
		return err
	}
	if tg == nil {
		return fmt.Errorf("no group found for ticket %d", ticketID)
	}
	return s.feishuCli.AddChatMembers(ctx, tg.GroupID, []string{feishuOpenID})
}

func (s *TicketGroupService) OnGroupMessage(ctx context.Context, chatID, userID, content string) error {
	tg, err := s.ticketRepo.GetTicketGroupByGroupID(ctx, chatID)
	if err != nil {
		return err
	}
	if tg == nil {
		return nil
	}

	ticket, err := s.ticketRepo.GetByID(ctx, tg.TicketID)
	if err != nil {
		return err
	}

	needHuman := s.detectEscalation(content)
	if needHuman {
		return s.escalateToHuman(ctx, ticket, tg, content)
	}

	return s.autoReply(ctx, ticket, tg, content)
}

func (s *TicketGroupService) autoReply(ctx context.Context, ticket *model.Ticket, tg *model.TicketGroup, content string) error {
	results, err := s.kbRepo.SearchSimilar(ctx, model.KnowledgeQuery{
		Query: content,
		Limit: 3,
	}, nil)
	if err != nil || len(results) == 0 {
		reply := fmt.Sprintf("Hello, ticket #%s has been received. We are processing it. For human assistance, reply 'agent'.\nOnline: %s",
			ticket.TicketNo, tg.JoinLink)

		msg := feishuCardMessage("Auto Reply", reply, "blue")
		return s.feishuCli.SendCardMessage(ctx, tg.GroupID, msg)
	}

	var parts []string
	for _, r := range results {
		rTitle := r.Title
		rContent := r.Content
		if len([]rune(rContent)) > 200 {
			rContent = string([]rune(rContent)[:200]) + "..."
		}
		parts = append(parts, fmt.Sprintf("%s\n%s", rTitle, rContent))
	}

	replyText := fmt.Sprintf("Found relevant knowledge:\n\n%s\n\nFor further help, reply 'agent'.", joinStrings(parts, "\n\n"))
	msg := feishuCardMessage("Knowledge Match Results", replyText, "green")
	return s.feishuCli.SendCardMessage(ctx, tg.GroupID, msg)
}

func (s *TicketGroupService) detectEscalation(content string) bool {
	keywords := []string{"human", "agent", "help", "人工", "客服"}
	for _, kw := range keywords {
		if containsString(content, kw) {
			return true
		}
	}
	return false
}

func (s *TicketGroupService) escalateToHuman(ctx context.Context, ticket *model.Ticket, tg *model.TicketGroup, content string) error {
	var openIDs []string

	if ticket.AssigneeID != nil {
		openIDs = append(openIDs, fmt.Sprintf("%d", *ticket.AssigneeID))
	}

	if ticket.RequesterID != 0 {
		openIDs = append(openIDs, fmt.Sprintf("%d", ticket.RequesterID))
	}

	for _, openID := range openIDs {
		notifyMsg := fmt.Sprintf("Ticket #%s needs human processing\nUser: %s\nPlease handle.", ticket.TicketNo, truncateString(content, 100))
		s.feishuCli.SendMessage(ctx, openID, "text", jsonEscape(notifyMsg))
	}

	reply := fmt.Sprintf("Forwarding to human agent, please wait. Ticket: #%s", ticket.TicketNo)
	card := feishuCardMessage("Transfer to Human", reply, "red")
	return s.feishuCli.SendCardMessage(ctx, tg.GroupID, card)
}

func (s *TicketGroupService) sendWelcomeMessage(ctx context.Context, chatID string, ticket *model.Ticket, joinLink string) {
	msg := fmt.Sprintf("Ticket #%s created\n\nTitle: %s\nDescription: %s\n%s\n\nDescribe your issue here. The bot will auto-reply.\nFor human help, reply 'agent'.",
		ticket.TicketNo, ticket.Title, ticket.Description, joinLink)
	card := feishuCardMessage("New Ticket Created", msg, "blue")
	if err := s.feishuCli.SendCardMessage(ctx, chatID, card); err != nil {
		log.Printf("failed to send welcome message: %v", err)
	}
}

func feishuCardMessage(title, content, color string) string {
	card := map[string]interface{}{
		"config": map[string]interface{}{
			"wide_screen_mode": true,
		},
		"header": map[string]interface{}{
			"title": map[string]interface{}{
				"tag":     "plain_text",
				"content": title,
			},
			"template": color,
		},
		"elements": []map[string]interface{}{
			{
				"tag":     "markdown",
				"content": content,
			},
		},
	}
	b, _ := json.Marshal(card)
	return string(b)
}

func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func joinStrings(parts []string, sep string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && contains(s, substr)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return s
}
