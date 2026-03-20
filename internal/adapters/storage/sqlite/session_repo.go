package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/port"
)

var _ port.SessionRepository = (*Storage)(nil)

func (s *Storage) SaveSession(ctx context.Context, session *domain.ReadingSession) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if session.SessionID == "" {
		return fmt.Errorf("session id is required")
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO sessions (session_id, book_id, current_page, last_read_time)
		 VALUES (?, ?, ?, ?)`,
		session.SessionID, session.BookID, session.CurrentPage, session.LastReadingTime)
	return err
}

// GetSessionByID returns all sessions for a book, joining TotalPages from books.
func (s *Storage) GetSessionByID(ctx context.Context, bookID string) ([]*domain.ReadingSession, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.session_id, s.book_id, s.current_page, s.last_read_time,
		       COALESCE(b.total_pages, 0)
		FROM sessions s
		LEFT JOIN books b ON b.id = s.book_id
		WHERE s.book_id = ?`, bookID)
	if err != nil {
		return nil, fmt.Errorf("GetSessionByID: %w", err)
	}
	defer rows.Close()
	var out []*domain.ReadingSession
	for rows.Next() {
		ses := &domain.ReadingSession{}
		if err := rows.Scan(&ses.SessionID, &ses.BookID, &ses.CurrentPage,
			&ses.LastReadingTime, &ses.TotalPages); err != nil {
			return nil, err
		}
		out = append(out, ses)
	}
	return out, rows.Err()
}

// GetLastReadingSession returns the most recent session, with TotalPages from books.
func (s *Storage) GetLastReadingSession(ctx context.Context, bookId string) (*domain.ReadingSession, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.session_id, s.book_id, s.current_page, s.last_read_time,
		       COALESCE(b.total_pages, 0)
		FROM sessions s
		LEFT JOIN books b ON b.id = s.book_id
		WHERE s.book_id = ?
		ORDER BY s.last_read_time DESC LIMIT 1`, bookId)
	if err != nil {
		return nil, fmt.Errorf("GetLastReadingSession: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		var ses domain.ReadingSession
		if err := rows.Scan(&ses.SessionID, &ses.BookID, &ses.CurrentPage,
			&ses.LastReadingTime, &ses.TotalPages); err != nil {
			return nil, err
		}
		return &ses, nil
	}
	return nil, nil
}
