package domain_test

import (
	"errors"
	"testing"

	"github.com/MiltonJ23/Orus/internal/domain"
)

func TestNewBook_Validation(t *testing.T) {
	// Test Valid Book
	book, err := domain.NewBook("Dune", "Frank Herbert", "/path/dune.pdf", domain.FormatPDF, 500)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if book.ID == "" {
		t.Error("Expected UUID to be generated")
	}

	// Test Empty Title
	_, err = domain.NewBook("", "Author", "path", domain.FormatPDF, 100)
	if !errors.Is(err, domain.ErrInvalidBookTitle) {
		t.Errorf("Expected ErrInvalidBookTitle, got %v", err)
	}
}
