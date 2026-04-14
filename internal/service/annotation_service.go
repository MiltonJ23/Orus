package service

import (
	"context"
	"fmt"
	"log"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/port"
)

// AnnotationService manages bookmarks and highlights for books.
type AnnotationService struct {
	annotRepo port.AnnotationRepository
	bookRepo  port.BookRepository
}

// NewAnnotationService creates a new AnnotationService with the given dependencies.
func NewAnnotationService(annotRepo port.AnnotationRepository, bookRepo port.BookRepository) *AnnotationService {
	return &AnnotationService{annotRepo: annotRepo, bookRepo: bookRepo}
}

// AddAnnotation creates a new annotation (bookmark or highlight) for a specific page.
func (a *AnnotationService) AddAnnotation(ctx context.Context, bookID string, annotationType domain.AnnotationType, pageNo int) (*domain.Annotation, error) {
	// Verify the book exists
	book, err := a.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("book not found: %w", err)
	}

	// Validate page is within the book's range
	if book.TotalPages > 0 && pageNo > book.TotalPages {
		return nil, fmt.Errorf("page %d exceeds book total pages (%d)", pageNo, book.TotalPages)
	}

	annot, err := domain.NewAnnotation(bookID, annotationType, pageNo)
	if err != nil {
		return nil, fmt.Errorf("invalid annotation: %w", err)
	}

	if err := a.annotRepo.SaveAnnotation(ctx, annot); err != nil {
		return nil, fmt.Errorf("save annotation: %w", err)
	}

	log.Printf("[Annotation] Ajoutée : type=%s page=%d livre=%s", annotationType, pageNo, bookID)
	return annot, nil
}

// ListAnnotationsForBook returns all annotations for a given book.
func (a *AnnotationService) ListAnnotationsForBook(ctx context.Context, bookID string) ([]*domain.Annotation, error) {
	annotations, err := a.annotRepo.ListAllAnnotationOfABook(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("list annotations: %w", err)
	}
	return annotations, nil
}

// GetAnnotationsByPage returns annotations for a specific page of a book.
func (a *AnnotationService) GetAnnotationsByPage(ctx context.Context, bookID string, pageNo int) ([]*domain.Annotation, error) {
	annotations, err := a.annotRepo.GetAnnotationByPage(ctx, pageNo, bookID)
	if err != nil {
		return nil, fmt.Errorf("get annotations by page: %w", err)
	}
	return annotations, nil
}

// GetAnnotationsByType returns all annotations of a specific type (bookmark or highlight).
func (a *AnnotationService) GetAnnotationsByType(ctx context.Context, annotationType domain.AnnotationType) ([]*domain.Annotation, error) {
	annotations, err := a.annotRepo.GetAnnotationByType(ctx, string(annotationType))
	if err != nil {
		return nil, fmt.Errorf("get annotations by type: %w", err)
	}
	return annotations, nil
}

// DeleteAnnotation removes an annotation by ID.
func (a *AnnotationService) DeleteAnnotation(ctx context.Context, annotationID string) error {
	if annotationID == "" {
		return fmt.Errorf("annotation ID cannot be empty")
	}
	if err := a.annotRepo.DeleteAnnotation(ctx, annotationID); err != nil {
		return fmt.Errorf("delete annotation: %w", err)
	}
	log.Printf("[Annotation] Supprimée : id=%s", annotationID)
	return nil
}

// CountAnnotationsForBook returns the number of annotations for a book.
func (a *AnnotationService) CountAnnotationsForBook(ctx context.Context, bookID string) (int, error) {
	annotations, err := a.annotRepo.ListAllAnnotationOfABook(ctx, bookID)
	if err != nil {
		return 0, fmt.Errorf("count annotations: %w", err)
	}
	return len(annotations), nil
}
