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
	// let's manage the context lifecycle
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if session.SessionID == "" {
		return fmt.Errorf("session id is required for saving the session")
	}
	// now let's build the query supposed to store a session
	query := `INSERT INTO sessions (session_id,book_id,current_page,last_read_time) VALUES (?,?,?,?)`

	_, queryExecError := s.db.ExecContext(ctx, query, session.SessionID, session.BookID, session.CurrentPage, session.LastReadingTime)
	if queryExecError != nil {
		return fmt.Errorf("unable to save session, an error occured during saving, %v", queryExecError)
	}
	return nil
}

func (s *Storage) GetSessionByID(ctx context.Context, bookID string) ([]*domain.ReadingSession, error) {
	// let's manage the context lifecycle
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// we build the fetching query
	query := `SELECT * FROM  sessions WHERE book_id = ?`

	rows, fetchingError := s.db.QueryContext(ctx, query, bookID)
	if fetchingError != nil {
		return nil, fmt.Errorf("unable to retrieve session, an error occured : %v", fetchingError)
	}
	defer rows.Close()

	// now let's exploit the stream
	var sessions []*domain.ReadingSession
	for rows.Next() {
		ses := &domain.ReadingSession{}
		scanningError := rows.Scan(&ses.SessionID, &ses.BookID, &ses.CurrentPage, &ses.LastReadingTime)
		if scanningError != nil {
			return nil, fmt.Errorf("unable to unload session from rows pointer, an error occured : %v", scanningError)
		}
		sessions = append(sessions, ses)
	}

	streamIterationError := rows.Err()
	if streamIterationError != nil {
		return nil, fmt.Errorf("an error occured while iterating annotations row: %v", streamIterationError)
	}

	return sessions, nil
}

func (s *Storage) GetLastReadingSession(ctx context.Context, bookId string) (*domain.ReadingSession, error) {
	// we manage the context lifecycle
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// we build the query
	query := `SELECT * FROM sessions WHERE book_id = ? ORDER BY last_read_time DESC LIMIT 1`

	rows, fetchingError := s.db.QueryContext(ctx, query, bookId)
	if fetchingError != nil {
		return nil, fmt.Errorf("unable to fetch last reading, an error occured : %v", fetchingError)
	}
	var session domain.ReadingSession
	if rows.Next() {
		scanningRowError := rows.Scan(&session.SessionID, &session.BookID, &session.CurrentPage, &session.LastReadingTime)
		if scanningRowError != nil {
			return nil, fmt.Errorf("an error occured when scanning the session rows pointer,%v", scanningRowError)
		}
	}
	return &session, nil
}
