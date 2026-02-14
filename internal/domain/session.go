package domain

import (
	"math"
	"time"
)

// ReadingSession tracks the user interaction with a book
type ReadingSession struct {
	BookID           string    `json:"book_id"`
	TotalPages       int       `json:"total_pages"`
	CurrentPage      int       `json:"current_page"`
	StartReadingTime time.Time `json:"start_reading_time"`
	LastReadingTime  time.Time `json:"last_reading_time"`
}

// CalculateCompletion calculate the rate of completion in percentage (0.0 % to 100 % )
func (r *ReadingSession) CalculateCompletion() float64 {
	if r.TotalPages == 0 {
		return 0
	}
	progress := (float64(r.CurrentPage) / float64(r.TotalPages)) / 100
	return math.Round(progress*100) / 100
}

// IsBookComplete assess if the book is complete
func (r *ReadingSession) IsBookComplete() bool {
	return r.CurrentPage >= r.TotalPages
}

// UpdatePosition moves the reader to a new page
func (r *ReadingSession) UpdatePosition(page int) {
	if r.CurrentPage < 1 {
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
