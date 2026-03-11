package service

import (
	"context"
	"fmt"
	"time"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/port"
)

type TrackerService struct {
	repo    port.BookRepository
	session port.SessionRepository
}

func NewTrackerService(repository port.BookRepository, session port.SessionRepository) *TrackerService {
	return &TrackerService{repo: repository, session: session}
}

func (t *TrackerService) OpenBook(ctx context.Context, bookId string) (*domain.ReadingSession, error) {
	book, err := t.repo.GetByID(ctx, bookId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving book: %w", err)
	}
	session, sessErr := t.session.GetLastReadingSession(ctx, bookId)
	currentPage := 1
	if sessErr == nil && session != nil {
		currentPage = session.CurrentPage
	}
	newSession, err := domain.NewSession(bookId, book.TotalPages, currentPage, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to init session: %w", err)
	}
	if err := t.session.SaveSession(ctx, newSession); err != nil {
		return nil, fmt.Errorf("failed to persist session: %w", err)
	}
	return newSession, nil
}

func (t *TrackerService) UpdateProgress(ctx context.Context, page int, ses *domain.ReadingSession) error {
	ses.UpdatePosition(page)
	return nil
}

// GetLastSession returns the most recent reading session for a book.
func (t *TrackerService) GetLastSession(ctx context.Context, bookID string) (*domain.ReadingSession, error) {
	return t.session.GetLastReadingSession(ctx, bookID)
}

// GetRecentSessions aggregates all sessions across all books.
func (t *TrackerService) GetRecentSessions(ctx context.Context) ([]*domain.ReadingSession, error) {
	books, err := t.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list books: %w", err)
	}
	var result []*domain.ReadingSession
	for _, b := range books {
		sessions, err := t.session.GetSessionByID(ctx, b.ID)
		if err != nil {
			continue
		}
		result = append(result, sessions...)
	}
	return result, nil
}

// GetMostRecentBook returns the book with the most recent reading activity.
func (t *TrackerService) GetMostRecentBook(ctx context.Context) (*domain.Book, *domain.ReadingSession, error) {
	books, err := t.repo.ListAll(ctx)
	if err != nil {
		return nil, nil, err
	}
	var latestBook *domain.Book
	var latestSession *domain.ReadingSession
	for _, book := range books {
		s, err := t.session.GetLastReadingSession(ctx, book.ID)
		if err != nil || s == nil {
			continue
		}
		if latestSession == nil || s.LastReadingTime.After(latestSession.LastReadingTime) {
			b := book
			latestBook = b
			latestSession = s
		}
	}
	return latestBook, latestSession, nil
}
