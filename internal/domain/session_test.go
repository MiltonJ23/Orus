package domain_test

import (
	"testing"

	"github.com/MiltonJ23/Orus/internal/domain"
)

func TestCalculateProgress(t *testing.T) {
	tests := []struct {
		name        string
		current     int
		total       int
		want        float64
		shouldPanic bool
	}{
		{"Start", 0, 100, 0.0, false},
		{"Middle", 50, 100, 50.0, false},
		{"Almost Done", 99, 300, 33.0, false},
		{"Finished", 200, 200, 100.0, false},
		{"Empty Book", 0, 0, 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := domain.ReadingSession{
				CurrentPage: tt.current,
				TotalPages:  tt.total,
			}
			got := session.CalculateCompletion()
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdatePosition(t *testing.T) {
	s := domain.ReadingSession{TotalPages: 100, CurrentPage: 50}

	// Case 1: Normal update
	s.UpdatePosition(60)
	if s.CurrentPage != 60 {
		t.Errorf("Expected page 60, got %d", s.CurrentPage)
	}

	// Case 2: Underflow (User tries to go to page -1)
	s.UpdatePosition(-5)
	if s.CurrentPage != 1 {
		t.Errorf("Expected page 1 (clamped), got %d", s.CurrentPage)
	}

	// Case 3: Overflow (User tries to go past end)
	s.UpdatePosition(150)
	if s.CurrentPage != 100 {
		t.Errorf("Expected page 100 (clamped), got %d", s.CurrentPage)
	}
}
