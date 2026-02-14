package domain

import "time"

type AnnotationType string

const (
	AnnotationBookmark  AnnotationType = "bookmark"
	AnnotationHighlight AnnotationType = "highlight"
)

// Annotation represents a user interaction with a specific part of the book
type Annotation struct {
	ID        string         `json:"id"`
	BookID    string         `json:"book_id"`
	Type      AnnotationType `json:"type"`
	PageNo    int            `json:"page_no"`
	CreatedAt time.Time      `json:"created_at"`
}

// NewBookMark create a new bookmark for a specific page
func NewBookMark(bookID string, pageNo int) *Annotation {
	return &Annotation{
		ID:        "note-" + time.Now().Format("20060102150405"),
		BookID:    bookID,
		Type:      AnnotationBookmark,
		PageNo:    pageNo,
		CreatedAt: time.Now(),
	}
}
