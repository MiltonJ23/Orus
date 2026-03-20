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

// OpenBook creates or resumes a reading session for the given book.
func (t *TrackerService) OpenBook(ctx context.Context, bookId string) (*domain.ReadingSession, error) {
	book, err := t.repo.GetByID(ctx, bookId)
	if err != nil {
		return nil, fmt.Errorf("OpenBook: retrieve book: %w", err)
	}

	session, _ := t.session.GetLastReadingSession(ctx, bookId)

	currentPage := 1
	if session != nil && session.CurrentPage > 0 {
		currentPage = session.CurrentPage
	}

	newSession, err := domain.NewSession(bookId, book.TotalPages, currentPage, time.Now())
	if err != nil {
		return nil, fmt.Errorf("OpenBook: init session: %w", err)
	}

	if err := t.session.SaveSession(ctx, newSession); err != nil {
		return nil, fmt.Errorf("OpenBook: save session: %w", err)
	}
	return newSession, nil
}

// UpdateProgress updates the current page and persists the session.
func (t *TrackerService) UpdateProgress(ctx context.Context, page int, ses *domain.ReadingSession) error {
	ses.UpdatePosition(page)
	ses.LastReadingTime = time.Now()
	if err := t.session.SaveSession(ctx, ses); err != nil {
		return fmt.Errorf("UpdateProgress: %w", err)
	}
	return nil
}

// GetMostRecentBook returns the book and its latest reading session.
// It iterates all books and picks the one whose last session is most recent.
func (t *TrackerService) GetMostRecentBook(ctx context.Context) (*domain.Book, *domain.ReadingSession, error) {
	books, err := t.repo.ListAll(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("GetMostRecentBook: list books: %w", err)
	}

	var bestBook *domain.Book
	var bestSession *domain.ReadingSession
	for _, b := range books {
		ses, err := t.session.GetLastReadingSession(ctx, b.ID)
		if err != nil || ses == nil {
			continue
		}
		if bestSession == nil || ses.LastReadingTime.After(bestSession.LastReadingTime) {
			bestBook = b
			bestSession = ses
		}
	}
	if bestBook == nil {
		return nil, nil, fmt.Errorf("no reading session found")
	}
	return bestBook, bestSession, nil
}

// GetRecentSessions returns the most recent session for every book that has one.
func (t *TrackerService) GetRecentSessions(ctx context.Context) ([]*domain.ReadingSession, error) {
	books, err := t.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetRecentSessions: list books: %w", err)
	}

	var out []*domain.ReadingSession
	for _, b := range books {
		ses, err := t.session.GetLastReadingSession(ctx, b.ID)
		if err != nil || ses == nil {
			continue
		}
		out = append(out, ses)
	}
	return out, nil
}

// BookCompletionStatus returns a map[bookID]→"done"|"reading"|"unread" for all books.
func (t *TrackerService) BookCompletionStatus(ctx context.Context) (map[string]string, error) {
	books, err := t.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("BookCompletionStatus: list books: %w", err)
	}

	result := make(map[string]string, len(books))
	for _, b := range books {
		ses, err := t.session.GetLastReadingSession(ctx, b.ID)
		if err != nil || ses == nil {
			result[b.ID] = "unread"
			continue
		}
		if ses.IsBookComplete() {
			result[b.ID] = "done"
		} else {
			result[b.ID] = "reading"
		}
	}
	return result, nil
}
