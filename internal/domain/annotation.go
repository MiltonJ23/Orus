package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// AnnotationType distinguishes between bookmark and highlight annotations.
type AnnotationType string

const (
	AnnotationBookmark  AnnotationType = "bookmark"
	AnnotationHighlight AnnotationType = "highlight"
)

// ErrInvalidPageNumber indicates an invalid page number (must be >= 1).
var (
	ErrInvalidPageNumber = errors.New("invalid annotation page number")
	// ErrInvalidBookId indicates an empty book ID was provided.
	ErrInvalidBookId = errors.New("invalid book id")
)

// Annotation represents a user interaction with a specific part of the book
type Annotation struct {
	ID             string         `json:"id"`
	BookID         string         `json:"book_id"`
	AnnotationType AnnotationType `json:"type"`
	PageNo         int            `json:"page_no"`
	CreatedAt      time.Time      `json:"created_at"`
}

// NewAnnotation creates a new annotation (bookmark or highlight) for a specific page.
func NewAnnotation(bookID string, annotationType AnnotationType, pageNo int) (*Annotation, error) {
	if pageNo < 1 {
		return nil, ErrInvalidPageNumber
	}
	if bookID == "" {
		return nil, ErrInvalidBookId
	}

	return &Annotation{
		ID:             uuid.New().String(),
		BookID:         bookID,
		AnnotationType: annotationType,
		PageNo:         pageNo,
		CreatedAt:      time.Now(),
	}, nil
}
