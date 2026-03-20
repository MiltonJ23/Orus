package domain

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
)

// ReadingSession tracks the user interaction with a book
type ReadingSession struct {
	SessionID       string
	BookID          string    `json:"book_id"`
	TotalPages      int       `json:"total_pages"`
	CurrentPage     int       `json:"current_page"`
	LastReadingTime time.Time `json:"last_reading_time"`
}

func NewSession(bookId string, totalPages, currentPages int, lastReadingTime time.Time) (*ReadingSession, error) {
	// we first of all check if the bookId is not empty
	if bookId == "" {
		return nil, ErrInvalidBookTitle
	}
	if currentPages < 1 {
		return nil, errors.New("invalid session page number")
	}
	var r ReadingSession
	r.SessionID = uuid.New().String()
	r.BookID = bookId
	r.TotalPages = totalPages
	r.CurrentPage = currentPages
	r.LastReadingTime = lastReadingTime

	return &r, nil
}

// CalculateCompletion calculate the rate of completion in percentage (0.0 % to 100 % )
func (r *ReadingSession) CalculateCompletion() float64 {
	if r.TotalPages == 0 {
		return 0
	}
	progress := (float64(r.CurrentPage) / float64(r.TotalPages)) * 100
	return math.Round(progress*100) / 100
}

// IsBookComplete assess if the book is complete
func (r *ReadingSession) IsBookComplete() bool {
	return r.CurrentPage >= r.TotalPages
}

// UpdatePosition moves the reader to a new page
func (r *ReadingSession) UpdatePosition(page int) {
	if page < 1 {
		r.CurrentPage = 1
		return
	}
	if page > r.TotalPages {
		r.CurrentPage = r.TotalPages
		return
	}

	r.CurrentPage = page
	r.LastReadingTime = time.Now()
}
