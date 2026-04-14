package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/service"
)

// --- MOCKS FOR READING SHEET ---

type mockSheetRepo struct {
	sheets     map[string]*domain.ReadingSheet
	byBookID   map[string]*domain.ReadingSheet
	failSave   bool
	failGet    bool
	failList   bool
	failUpdate bool
	failDelete bool
}

func newMockSheetRepo() *mockSheetRepo {
	return &mockSheetRepo{
		sheets:   make(map[string]*domain.ReadingSheet),
		byBookID: make(map[string]*domain.ReadingSheet),
	}
}

func (m *mockSheetRepo) SaveSheet(_ context.Context, s *domain.ReadingSheet) error {
	if m.failSave {
		return errors.New("save error")
	}
	m.sheets[s.ID] = s
	m.byBookID[s.BookID] = s
	return nil
}

func (m *mockSheetRepo) GetSheetByID(_ context.Context, id string) (*domain.ReadingSheet, error) {
	if m.failGet {
		return nil, errors.New("get error")
	}
	s, ok := m.sheets[id]
	if !ok {
		return nil, domain.ErrReadingSheetNotFound
	}
	return s, nil
}

func (m *mockSheetRepo) GetSheetByBookID(_ context.Context, bookID string) (*domain.ReadingSheet, error) {
	if m.failGet {
		return nil, errors.New("get error")
	}
	s, ok := m.byBookID[bookID]
	if !ok {
		return nil, domain.ErrReadingSheetNotFound
	}
	return s, nil
}

func (m *mockSheetRepo) ListAllSheets(_ context.Context) ([]*domain.ReadingSheet, error) {
	if m.failList {
		return nil, errors.New("list error")
	}
	var out []*domain.ReadingSheet
	for _, s := range m.sheets {
		out = append(out, s)
	}
	return out, nil
}

func (m *mockSheetRepo) UpdateSheet(_ context.Context, s *domain.ReadingSheet) error {
	if m.failUpdate {
		return errors.New("update error")
	}
	m.sheets[s.ID] = s
	m.byBookID[s.BookID] = s
	return nil
}

func (m *mockSheetRepo) DeleteSheet(_ context.Context, id string) error {
	if m.failDelete {
		return errors.New("delete error")
	}
	if s, ok := m.sheets[id]; ok {
		delete(m.byBookID, s.BookID)
	}
	delete(m.sheets, id)
	return nil
}

type mockSheetBookRepo struct {
	books   map[string]*domain.Book
	failGet bool
}

func newMockSheetBookRepo() *mockSheetBookRepo {
	return &mockSheetBookRepo{
		books: map[string]*domain.Book{
			"book-1": {ID: "book-1", Title: "Test Book"},
		},
	}
}

func (m *mockSheetBookRepo) Save(_ context.Context, b *domain.Book) error      { return nil }
func (m *mockSheetBookRepo) ListAll(_ context.Context) ([]*domain.Book, error) { return nil, nil }
func (m *mockSheetBookRepo) Delete(_ context.Context, id string) error         { return nil }

func (m *mockSheetBookRepo) GetByID(_ context.Context, id string) (*domain.Book, error) {
	if m.failGet {
		return nil, errors.New("book not found")
	}
	b, ok := m.books[id]
	if !ok {
		return nil, domain.ErrBookNotFound
	}
	return b, nil
}

// --- TESTS ---

func TestReadingSheetService_CreateSheet(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		sheetRepo := newMockSheetRepo()
		bookRepo := newMockSheetBookRepo()
		svc := service.NewReadingSheetService(sheetRepo, bookRepo)

		sheet, err := svc.CreateSheet(ctx, "book-1", "Great book", 4, []string{"quote1"}, []string{"fiction"})
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if sheet.BookTitle != "Test Book" {
			t.Errorf("expected book title 'Test Book', got %q", sheet.BookTitle)
		}
		if sheet.Rating != 4 {
			t.Errorf("expected rating 4, got %d", sheet.Rating)
		}
	})

	t.Run("BookNotFound", func(t *testing.T) {
		sheetRepo := newMockSheetRepo()
		bookRepo := newMockSheetBookRepo()
		svc := service.NewReadingSheetService(sheetRepo, bookRepo)

		_, err := svc.CreateSheet(ctx, "nonexistent", "summary", 3, nil, nil)
		if err == nil {
			t.Fatal("expected error for nonexistent book")
		}
	})

	t.Run("InvalidRating", func(t *testing.T) {
		sheetRepo := newMockSheetRepo()
		bookRepo := newMockSheetBookRepo()
		svc := service.NewReadingSheetService(sheetRepo, bookRepo)

		_, err := svc.CreateSheet(ctx, "book-1", "summary", 7, nil, nil)
		if err == nil {
			t.Fatal("expected error for invalid rating")
		}
	})

	t.Run("SaveError", func(t *testing.T) {
		sheetRepo := newMockSheetRepo()
		sheetRepo.failSave = true
		bookRepo := newMockSheetBookRepo()
		svc := service.NewReadingSheetService(sheetRepo, bookRepo)

		_, err := svc.CreateSheet(ctx, "book-1", "summary", 3, nil, nil)
		if err == nil {
			t.Fatal("expected save error, got nil")
		}
	})
}

func TestReadingSheetService_GetSheetForBook(t *testing.T) {
	ctx := context.Background()

	t.Run("Exists", func(t *testing.T) {
		sheetRepo := newMockSheetRepo()
		bookRepo := newMockSheetBookRepo()
		svc := service.NewReadingSheetService(sheetRepo, bookRepo)

		_, _ = svc.CreateSheet(ctx, "book-1", "summary", 3, nil, nil)

		sheet, err := svc.GetSheetForBook(ctx, "book-1")
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if sheet == nil {
			t.Fatal("expected sheet, got nil")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		sheetRepo := newMockSheetRepo()
		bookRepo := newMockSheetBookRepo()
		svc := service.NewReadingSheetService(sheetRepo, bookRepo)

		sheet, err := svc.GetSheetForBook(ctx, "book-1")
		if err != nil {
			t.Fatalf("expected nil error for missing sheet, got: %v", err)
		}
		if sheet != nil {
			t.Error("expected nil sheet for book without sheet")
		}
	})
}

func TestReadingSheetService_ListSheets(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		sheetRepo := newMockSheetRepo()
		bookRepo := newMockSheetBookRepo()
		svc := service.NewReadingSheetService(sheetRepo, bookRepo)

		_, _ = svc.CreateSheet(ctx, "book-1", "summary", 3, nil, nil)

		sheets, err := svc.ListSheets(ctx)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if len(sheets) != 1 {
			t.Errorf("expected 1 sheet, got %d", len(sheets))
		}
	})

	t.Run("Error", func(t *testing.T) {
		sheetRepo := newMockSheetRepo()
		sheetRepo.failList = true
		bookRepo := newMockSheetBookRepo()
		svc := service.NewReadingSheetService(sheetRepo, bookRepo)

		_, err := svc.ListSheets(ctx)
		if err == nil {
			t.Fatal("expected list error, got nil")
		}
	})
}

func TestReadingSheetService_UpdateSummary(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		sheetRepo := newMockSheetRepo()
		bookRepo := newMockSheetBookRepo()
		svc := service.NewReadingSheetService(sheetRepo, bookRepo)

		sheet, _ := svc.CreateSheet(ctx, "book-1", "old summary", 3, nil, nil)

		err := svc.UpdateSummary(ctx, sheet.ID, "new summary")
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		updated := sheetRepo.sheets[sheet.ID]
		if updated.Summary != "new summary" {
			t.Errorf("expected summary 'new summary', got %q", updated.Summary)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		sheetRepo := newMockSheetRepo()
		bookRepo := newMockSheetBookRepo()
		svc := service.NewReadingSheetService(sheetRepo, bookRepo)

		err := svc.UpdateSummary(ctx, "nonexistent", "new")
		if err == nil {
			t.Fatal("expected error for nonexistent sheet")
		}
	})
}

func TestReadingSheetService_SetRating(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		sheetRepo := newMockSheetRepo()
		bookRepo := newMockSheetBookRepo()
		svc := service.NewReadingSheetService(sheetRepo, bookRepo)

		sheet, _ := svc.CreateSheet(ctx, "book-1", "summary", 3, nil, nil)

		err := svc.SetRating(ctx, sheet.ID, 5)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		updated := sheetRepo.sheets[sheet.ID]
		if updated.Rating != 5 {
			t.Errorf("expected rating 5, got %d", updated.Rating)
		}
	})

	t.Run("InvalidRating", func(t *testing.T) {
		sheetRepo := newMockSheetRepo()
		bookRepo := newMockSheetBookRepo()
		svc := service.NewReadingSheetService(sheetRepo, bookRepo)

		sheet, _ := svc.CreateSheet(ctx, "book-1", "summary", 3, nil, nil)

		err := svc.SetRating(ctx, sheet.ID, 10)
		if err == nil {
			t.Fatal("expected error for invalid rating")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		sheetRepo := newMockSheetRepo()
		bookRepo := newMockSheetBookRepo()
		svc := service.NewReadingSheetService(sheetRepo, bookRepo)

		err := svc.SetRating(ctx, "nonexistent", 3)
		if err == nil {
			t.Fatal("expected error for nonexistent sheet")
		}
	})
}

func TestReadingSheetService_AddQuote(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		sheetRepo := newMockSheetRepo()
		bookRepo := newMockSheetBookRepo()
		svc := service.NewReadingSheetService(sheetRepo, bookRepo)

		sheet, _ := svc.CreateSheet(ctx, "book-1", "summary", 3, nil, nil)

		err := svc.AddQuote(ctx, sheet.ID, "To be or not to be")
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		updated := sheetRepo.sheets[sheet.ID]
		if len(updated.Quotes) != 1 {
			t.Errorf("expected 1 quote, got %d", len(updated.Quotes))
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		sheetRepo := newMockSheetRepo()
		bookRepo := newMockSheetBookRepo()
		svc := service.NewReadingSheetService(sheetRepo, bookRepo)

		err := svc.AddQuote(ctx, "nonexistent", "quote")
		if err == nil {
			t.Fatal("expected error for nonexistent sheet")
		}
	})
}

func TestReadingSheetService_DeleteSheet(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		sheetRepo := newMockSheetRepo()
		bookRepo := newMockSheetBookRepo()
		svc := service.NewReadingSheetService(sheetRepo, bookRepo)

		sheet, _ := svc.CreateSheet(ctx, "book-1", "summary", 3, nil, nil)

		err := svc.DeleteSheet(ctx, sheet.ID)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if _, ok := sheetRepo.sheets[sheet.ID]; ok {
			t.Error("expected sheet to be deleted")
		}
	})

	t.Run("Error", func(t *testing.T) {
		sheetRepo := newMockSheetRepo()
		sheetRepo.failDelete = true
		bookRepo := newMockSheetBookRepo()
		svc := service.NewReadingSheetService(sheetRepo, bookRepo)

		err := svc.DeleteSheet(ctx, "anything")
		if err == nil {
			t.Fatal("expected delete error, got nil")
		}
	})
}
