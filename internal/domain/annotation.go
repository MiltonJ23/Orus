package domain

import (
	"errors"
	"time"
)

type AnnotationType string

const (
	AnnotationBookmark  AnnotationType = "bookmark"
	AnnotationHighlight AnnotationType = "highlight"
)

var (
	ErrInvalidPageNumber = errors.New("invalid annotation page number")
	ErrInvalidBookId     = errors.New("invalid book id")
)

// Annotation represents a user interaction with a specific part of the book
type Annotation struct {
	ID             string         `json:"id"`
	BookID         string         `json:"book_id"`
	AnnotationType AnnotationType `json:"type"`
	PageNo         int            `json:"page_no"`
	CreatedAt      time.Time      `json:"created_at"`
}

// NewBookMark create a new bookmark for a specific page
func NewBookMark(bookID string, format AnnotationType, pageNo int) (*Annotation, error) {
	if pageNo < 1 { // we check if the method receives an invalid page number , we reject the creation of the bookmark
		return nil, ErrInvalidPageNumber
	}
	// now let's check if the bookID is not empty
	if bookID == "" {
		return nil, ErrInvalidBookId
	}

	return &Annotation{
		ID:             "note-" + time.Now().Format("20060102150405"),
		BookID:         bookID,
		AnnotationType: format,
		PageNo:         pageNo,
		CreatedAt:      time.Now(),
	}, nil
}
