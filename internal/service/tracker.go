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
	// we retrieve the last reading session
	session, retrievingSessionError := t.session.GetLastReadingSession(ctx, bookId)
	if retrievingSessionError != nil {
		return nil, fmt.Errorf("an error occured trying to retrieve  the last reading session,  %v", retrievingSessionError)
	}
	// which means that we have the data about the last reading session
	// we retrieve the book using that bookId
	book, retrievingBookError := t.repo.GetByID(ctx, bookId)
	if retrievingBookError != nil {
		return nil, fmt.Errorf("an error trying to retrieve the book for this session, %v", retrievingBookError)
	}
	newsession, _ := domain.NewSession(bookId, book.TotalPages, session.CurrentPage, time.Now())
	return newsession, nil
}

func (t *TrackerService) UpdateProgress(ctx context.Context, page int, ses *domain.ReadingSession) error {
	ses.UpdatePosition(page)
	return nil
}
