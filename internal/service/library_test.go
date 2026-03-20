package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/service"
)

// --- MOCKS FOR LIBRARY ---

type mockLibBookRepo struct {
	failSave bool
	failList bool
}

func (m *mockLibBookRepo) Save(ctx context.Context, book *domain.Book) error {
	if m.failSave {
		return errors.New("db save error")
	}
	return nil
}

func (m *mockLibBookRepo) GetByID(ctx context.Context, id string) (*domain.Book, error) {
	return nil, nil
}

func (m *mockLibBookRepo) ListAll(ctx context.Context) ([]*domain.Book, error) {
	if m.failList {
		return nil, errors.New("db list error")
	}
	return []*domain.Book{{ID: "123", Title: "Mock Book"}}, nil
}

func (m *mockLibBookRepo) Delete(ctx context.Context, id string) error {
	return nil
}

type mockExtractor struct {
	failExtract              bool
	triggerBookCreationError bool
}

func (m *mockExtractor) ExtractInfo(ctx context.Context, path string) (*domain.BookMetadata, error) {
	if m.failExtract {
		return nil, errors.New("extraction failed")
	}

	title := "Mock Title"
	if m.triggerBookCreationError {
		title = "" // Triggers domain.ErrInvalidBookTitle in NewBook
	}

	return &domain.BookMetadata{
		Title:      title,
		Author:     "Mock Author",
		FilePath:   path,
		Format:     domain.FormatPDF,
		TotalPages: 100,
	}, nil
}

// --- TESTS ---

func TestLibraryService_ImportBook(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		repo := &mockLibBookRepo{}
		extractor := &mockExtractor{}
		svc := service.NewLibraryService(repo, extractor)

		book, err := svc.ImportBook(ctx, "/path/to/book.pdf")
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if book.Title != "Mock Title" {
			t.Errorf("expected book title 'Mock Title', got: %s", book.Title)
		}
	})

	t.Run("Extraction Error", func(t *testing.T) {
		repo := &mockLibBookRepo{}
		extractor := &mockExtractor{failExtract: true}
		svc := service.NewLibraryService(repo, extractor)

		_, err := svc.ImportBook(ctx, "/path/to/book.pdf")
		if err == nil {
			t.Fatal("expected extraction error, got nil")
		}
	})

	t.Run("Book Creation Error", func(t *testing.T) {
		repo := &mockLibBookRepo{}
		extractor := &mockExtractor{triggerBookCreationError: true}
		svc := service.NewLibraryService(repo, extractor)

		_, err := svc.ImportBook(ctx, "/path/to/book.pdf")
		if err == nil {
			t.Fatal("expected book creation error, got nil")
		}
	})

	t.Run("Database Save Error", func(t *testing.T) {
		repo := &mockLibBookRepo{failSave: true}
		extractor := &mockExtractor{}
		svc := service.NewLibraryService(repo, extractor)

		_, err := svc.ImportBook(ctx, "/path/to/book.pdf")
		if err == nil {
			t.Fatal("expected db save error, got nil")
		}
	})
}

func TestLibraryService_GetLibrary(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		repo := &mockLibBookRepo{}
		svc := service.NewLibraryService(repo, nil)

		books, err := svc.GetLibrary(ctx)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if len(books) != 1 {
			t.Errorf("expected 1 book, got %d", len(books))
		}
	})
}
