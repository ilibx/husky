package model

import (
	"testing"
)

func TestIsValidTransition(t *testing.T) {
	tests := []struct {
		from TicketStatus
		to   TicketStatus
		want bool
	}{
		{TicketStatusOpen, TicketStatusInProgress, true},
		{TicketStatusOpen, TicketStatusResolved, false},
		{TicketStatusOpen, TicketStatusClosed, true},
		{TicketStatusOpen, TicketStatusPending, false},
		{TicketStatusInProgress, TicketStatusPending, true},
		{TicketStatusInProgress, TicketStatusResolved, true},
		{TicketStatusInProgress, TicketStatusOpen, true},
		{TicketStatusPending, TicketStatusInProgress, true},
		{TicketStatusPending, TicketStatusResolved, true},
		{TicketStatusPending, TicketStatusOpen, false},
		{TicketStatusResolved, TicketStatusClosed, true},
		{TicketStatusResolved, TicketStatusOpen, true},
		{TicketStatusClosed, TicketStatusOpen, false},
		{TicketStatusClosed, TicketStatusInProgress, false},
	}

	for _, tt := range tests {
		got := IsValidTransition(tt.from, tt.to)
		if got != tt.want {
			t.Errorf("IsValidTransition(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}

func TestIsValidTransitionInvalidFrom(t *testing.T) {
	if IsValidTransition("nonexistent", TicketStatusOpen) {
		t.Error("expected false for invalid from status")
	}
}

func TestIsValidPriority(t *testing.T) {
	valid := []string{"low", "medium", "high", "urgent"}
	invalid := []string{"", "critical", "LOW", "Medium"}

	for _, p := range valid {
		if !IsValidPriority(p) {
			t.Errorf("IsValidPriority(%s) = false, want true", p)
		}
	}
	for _, p := range invalid {
		if IsValidPriority(p) {
			t.Errorf("IsValidPriority(%s) = true, want false", p)
		}
	}
}
