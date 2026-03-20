package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/service"
)

// --- MOCKS FOR TRACKER ---

type mockTrackerBookRepo struct {
	failGet bool
}

func (m *mockTrackerBookRepo) Save(ctx context.Context, book *domain.Book) error   { return nil }
func (m *mockTrackerBookRepo) ListAll(ctx context.Context) ([]*domain.Book, error) { return nil, nil }
func (m *mockTrackerBookRepo) Delete(ctx context.Context, id string) error         { return nil }

func (m *mockTrackerBookRepo) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	if m.failGet {
		return nil, errors.New("db get error")
	}
	return &domain.Book{ID: id, TotalPages: 200}, nil
}

type mockSessionRepo struct {
	failGetLast bool
}

func (m *mockSessionRepo) SaveSession(ctx context.Context, session *domain.ReadingSession) error {
	return nil
}
func (m *mockSessionRepo) GetSessionByID(ctx context.Context, bookID string) ([]*domain.ReadingSession, error) {
	return nil, nil
}

func (m *mockSessionRepo) GetLastReadingSession(ctx context.Context, bookId string) (*domain.ReadingSession, error) {
	if m.failGetLast {
		return nil, errors.New("session retrieve error")
	}
	return &domain.ReadingSession{CurrentPage: 42}, nil
}

// --- TESTS ---

func TestTrackerService_OpenBook(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		bRepo := &mockTrackerBookRepo{}
		sRepo := &mockSessionRepo{}
		svc := service.NewTrackerService(bRepo, sRepo)

		session, err := svc.OpenBook(ctx, "book-123")
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if session == nil {
			t.Fatal("expected session, got nil")
		}
	})

	t.Run("Session Retrieve Error", func(t *testing.T) {
		bRepo := &mockTrackerBookRepo{}
		sRepo := &mockSessionRepo{failGetLast: true}
		svc := service.NewTrackerService(bRepo, sRepo)

		_, err := svc.OpenBook(ctx, "book-123")
		if err == nil {
			t.Fatal("expected session retrieval error, got nil")
		}
	})

	t.Run("Book Retrieve Error", func(t *testing.T) {
		bRepo := &mockTrackerBookRepo{failGet: true}
		sRepo := &mockSessionRepo{}
		svc := service.NewTrackerService(bRepo, sRepo)

		_, err := svc.OpenBook(ctx, "book-123")
		if err == nil {
			t.Fatal("expected book retrieval error, got nil")
		}
	})
}

func TestTrackerService_UpdateProgress(t *testing.T) {
	svc := service.NewTrackerService(nil, nil)
	session := &domain.ReadingSession{TotalPages: 100, CurrentPage: 10}

	err := svc.UpdateProgress(context.Background(), 20, session)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if session.CurrentPage != 20 {
		t.Errorf("expected current page 20, got %d", session.CurrentPage)
	}
}
