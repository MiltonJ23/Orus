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
	return &TrackerService{
		repo:    repository,
		session: session,
	}
}

func (t *TrackerService) OpenBook(ctx context.Context, bookId string) (*domain.ReadingSession, error) {
	// 1. Retrieve the book first to ensure it exists and get its TotalPages
	book, retrievingBookError := t.repo.GetByID(ctx, bookId)
	if retrievingBookError != nil {
		return nil, fmt.Errorf("an error trying to retrieve the book for this session: %w", retrievingBookError)
	}

	// 2. Try to get the last reading session
	session, retrievingSessionError := t.session.GetLastReadingSession(ctx, bookId)

	// 3. Business Logic: Determine the starting page
	currentPage := 1 // Default to page 1 for new books

	if retrievingSessionError == nil && session != nil {
		// If a session exists, resume from where they left off
		currentPage = session.CurrentPage
	}

	// 4. Create the session object
	newSession, err := domain.NewSession(bookId, book.TotalPages, currentPage, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize session entity: %w", err)
	}

	// 5. CRITICAL: Save the session so it's tracked in the database immediately
	savingError := t.session.SaveSession(ctx, newSession)
	if savingError != nil {
		return nil, fmt.Errorf("failed to persist new reading session: %w", savingError)
	}

	return newSession, nil
}

func (t *TrackerService) UpdateProgress(ctx context.Context, page int, ses *domain.ReadingSession) error {
	ses.UpdatePosition(page)
	return nil
}
