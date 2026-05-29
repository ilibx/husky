package ticket

import (
	"context"
	"strings"
	"testing"

	"github.com/husky/husky/internal/model"
	"github.com/husky/husky/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.Ticket{},
		&model.Comment{},
		&model.Attachment{},
		&model.Notification{},
		&model.AuditLog{},
		&model.KnowledgeBase{},
		&model.Category{},
		&model.Department{},
		&model.SOP{},
		&model.Satisfaction{},
		&model.WebhookMessageRecord{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestCreateTicket_Validation(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTicketRepository(db)
	svc := NewService(repo)

	_, err := svc.CreateTicket(context.Background(), &model.CreateTicketRequest{
		Title:       "",
		Description: "desc",
	})
	if err == nil {
		t.Error("expected error for empty title")
	}

	_, err = svc.CreateTicket(context.Background(), &model.CreateTicketRequest{
		Title:       "title",
		Description: "",
	})
	if err == nil {
		t.Error("expected error for empty description")
	}
}

func TestCreateTicket_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTicketRepository(db)
	svc := NewService(repo)

	resp, err := svc.CreateTicket(context.Background(), &model.CreateTicketRequest{
		Title:       "Test Ticket",
		Description: "Test Description",
		Priority:    "high",
		Channel:     "web",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Title != "Test Ticket" {
		t.Errorf("got title %s, want Test Ticket", resp.Title)
	}
	if resp.Status != "open" {
		t.Errorf("got status %s, want open", resp.Status)
	}
	if resp.Priority != "high" {
		t.Errorf("got priority %s, want high", resp.Priority)
	}
	if resp.Source != "web" {
		t.Errorf("got source %s, want web", resp.Source)
	}
}

func TestCreateTicket_InvalidPriority(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTicketRepository(db)
	svc := NewService(repo)

	_, err := svc.CreateTicket(context.Background(), &model.CreateTicketRequest{
		Title:       "Test",
		Description: "Test",
		Priority:    "urgent",
	})
	if err != nil {
		t.Errorf("expected no error for valid priority 'urgent': %v", err)
	}

	_, err = svc.CreateTicket(context.Background(), &model.CreateTicketRequest{
		Title:       "Test",
		Description: "Test",
		Priority:    "invalid",
	})
	if err == nil {
		t.Error("expected error for invalid priority")
	}
}

func TestAssignTicket(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTicketRepository(db)
	svc := NewService(repo)

	resp, _ := svc.CreateTicket(context.Background(), &model.CreateTicketRequest{
		Title:       "Assign Test",
		Description: "Test",
	})

	if err := svc.AssignTicket(context.Background(), resp.ID, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ticket, _ := svc.GetTicket(context.Background(), resp.ID)
	if ticket.AssigneeID == nil || *ticket.AssigneeID != 1 {
		t.Errorf("expected assignee 1, got %v", ticket.AssigneeID)
	}
}

func TestUpdateStatus_Valid(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTicketRepository(db)
	svc := NewService(repo)

	resp, _ := svc.CreateTicket(context.Background(), &model.CreateTicketRequest{
		Title:       "Status Test",
		Description: "Test",
	})

	if err := svc.UpdateStatus(context.Background(), resp.ID, "in_progress"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ticket, _ := svc.GetTicket(context.Background(), resp.ID)
	if ticket.Status != "in_progress" {
		t.Errorf("got status %s, want in_progress", ticket.Status)
	}
}

func TestUpdateStatus_Invalid(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTicketRepository(db)
	svc := NewService(repo)

	resp, _ := svc.CreateTicket(context.Background(), &model.CreateTicketRequest{
		Title:       "Status Test",
		Description: "Test",
	})

	err := svc.UpdateStatus(context.Background(), resp.ID, "resolved")
	if err == nil {
		t.Fatal("expected error for invalid transition open -> resolved")
	}
	if err == nil || !strings.Contains(err.Error(), "cannot transition") {
		t.Errorf("expected transition error, got %v", err)
	}
}

func TestAddComment_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTicketRepository(db)
	svc := NewService(repo)

	_, err := svc.AddComment(context.Background(), 1, 1, "", false)
	if err == nil {
		t.Error("expected error for empty comment")
	}
}

func TestDeleteTicket(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTicketRepository(db)
	svc := NewService(repo)

	resp, _ := svc.CreateTicket(context.Background(), &model.CreateTicketRequest{
		Title:       "Delete Test",
		Description: "Test",
	})

	if err := svc.DeleteTicket(context.Background(), resp.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.GetTicket(context.Background(), resp.ID)
	if err == nil {
		t.Error("expected error for deleted ticket")
	}
}

func TestRateTicket_Validation(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewTicketRepository(db)
	svc := NewService(repo)

	resp, _ := svc.CreateTicket(context.Background(), &model.CreateTicketRequest{
		Title:       "Rate Test",
		Description: "Test",
	})

	// Cannot rate open ticket
	_, err := svc.RateTicket(context.Background(), resp.ID, 1, 5, "")
	if err == nil {
		t.Error("expected error for rating open ticket")
	}

	// Invalid score
	_, err = svc.RateTicket(context.Background(), resp.ID, 1, 0, "")
	if err == nil {
		t.Error("expected error for score 0")
	}
	_, err = svc.RateTicket(context.Background(), resp.ID, 1, 6, "")
	if err == nil {
		t.Error("expected error for score 6")
	}
}
