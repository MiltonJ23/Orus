package domain_test

import (
	"errors"
	"testing"

	"github.com/MiltonJ23/Orus/internal/domain"
)

func TestNewAnnotation(t *testing.T) {
	_, err := domain.NewAnnotation("3ubiosdfo24", domain.AnnotationBookmark, 30)
	if err != nil {
		t.Fatalf("expected no errors, got %v", err)
	}

	_, errBookID := domain.NewAnnotation("", domain.AnnotationBookmark, 30)
	if !errors.Is(errBookID, domain.ErrInvalidBookId) {
		t.Errorf("expected error to be ErrInvalidBookId, instead got %v", errBookID)
	}

	_, errPage := domain.NewAnnotation("3ubiosdfo24", domain.AnnotationBookmark, -2)
	if !errors.Is(errPage, domain.ErrInvalidPageNumber) {
		t.Errorf("expected error to be ErrInvalidPageNumber, instead got %v", errPage)
	}
}

func TestAnnotationType(t *testing.T) {
	bookmark, err := domain.NewAnnotation("kln934nalj4fs", domain.AnnotationBookmark, 30)
	if err != nil {
		t.Fatalf("expected no errors, got %v", err)
	}
	if bookmark.AnnotationType != "bookmark" {
		t.Fatalf("expected type 'bookmark', got %v", bookmark.AnnotationType)
	}

	highlight, err := domain.NewAnnotation("kln934nalj4fs", domain.AnnotationHighlight, 30)
	if err != nil {
		t.Fatalf("expected no errors, got %v", err)
	}
	if highlight.AnnotationType != "highlight" {
		t.Fatalf("expected type 'highlight', got %v", highlight.AnnotationType)
	}
}

func TestAnnotationIDUniqueness(t *testing.T) {
	a1, _ := domain.NewAnnotation("book1", domain.AnnotationBookmark, 1)
	a2, _ := domain.NewAnnotation("book1", domain.AnnotationBookmark, 1)
	if a1.ID == a2.ID {
		t.Fatal("two annotations created with the same ID")
	}
}
