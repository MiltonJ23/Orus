package service_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/service"
)

// --- MOCKS FOR SHARING ---

type mockSharingBookRepo struct {
	books []*domain.Book
}

func (m *mockSharingBookRepo) Save(_ context.Context, _ *domain.Book) error { return nil }
func (m *mockSharingBookRepo) Delete(_ context.Context, _ string) error     { return nil }

func (m *mockSharingBookRepo) GetByID(_ context.Context, id string) (*domain.Book, error) {
	for _, b := range m.books {
		if b.ID == id {
			return b, nil
		}
	}
	return nil, domain.ErrBookNotFound
}

func (m *mockSharingBookRepo) ListAll(_ context.Context) ([]*domain.Book, error) {
	return m.books, nil
}

type mockSharingSheetRepo struct {
	sheets map[string]*domain.ReadingSheet
}

func newMockSharingSheetRepo() *mockSharingSheetRepo {
	return &mockSharingSheetRepo{sheets: make(map[string]*domain.ReadingSheet)}
}

func (m *mockSharingSheetRepo) SaveSheet(_ context.Context, s *domain.ReadingSheet) error {
	m.sheets[s.ID] = s
	return nil
}

func (m *mockSharingSheetRepo) GetSheetByID(_ context.Context, id string) (*domain.ReadingSheet, error) {
	s, ok := m.sheets[id]
	if !ok {
		return nil, domain.ErrReadingSheetNotFound
	}
	return s, nil
}

func (m *mockSharingSheetRepo) GetSheetByBookID(_ context.Context, bookID string) (*domain.ReadingSheet, error) {
	for _, s := range m.sheets {
		if s.BookID == bookID {
			return s, nil
		}
	}
	return nil, domain.ErrReadingSheetNotFound
}

func (m *mockSharingSheetRepo) ListAllSheets(_ context.Context) ([]*domain.ReadingSheet, error) {
	var out []*domain.ReadingSheet
	for _, s := range m.sheets {
		out = append(out, s)
	}
	return out, nil
}

func (m *mockSharingSheetRepo) UpdateSheet(_ context.Context, s *domain.ReadingSheet) error {
	m.sheets[s.ID] = s
	return nil
}

func (m *mockSharingSheetRepo) DeleteSheet(_ context.Context, id string) error {
	delete(m.sheets, id)
	return nil
}

// --- TESTS ---

func TestSharingService_ExportLibrary(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	bookRepo := &mockSharingBookRepo{
		books: []*domain.Book{
			{ID: "b1", Title: "Go Programming", Author: "Author A", Format: domain.FormatPDF, TotalPages: 300},
			{ID: "b2", Title: "Clean Architecture", Author: "Author B", Format: domain.FormatEPUB, TotalPages: 200},
		},
	}

	sheetRepo := newMockSharingSheetRepo()
	sheet, _ := domain.NewReadingSheet("b1", "Go Programming", "A great book about Go", 5, []string{"Go is simple"}, []string{"programming"})
	sheetRepo.sheets[sheet.ID] = sheet

	svc := service.NewSharingService(bookRepo, sheetRepo)

	t.Run("Markdown", func(t *testing.T) {
		path, err := svc.ExportLibrary(ctx, service.ShareFormatMarkdown, tmpDir)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !strings.HasSuffix(path, ".md") {
			t.Errorf("expected .md extension, got %q", path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read exported file: %v", err)
		}
		content := string(data)
		if !strings.Contains(content, "Go Programming") {
			t.Error("expected markdown to contain book title")
		}
		if !strings.Contains(content, "Clean Architecture") {
			t.Error("expected markdown to contain second book title")
		}
	})

	t.Run("JSON", func(t *testing.T) {
		path, err := svc.ExportLibrary(ctx, service.ShareFormatJSON, tmpDir)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !strings.HasSuffix(path, ".json") {
			t.Errorf("expected .json extension, got %q", path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read exported file: %v", err)
		}
		content := string(data)
		if !strings.Contains(content, "Go Programming") {
			t.Error("expected JSON to contain book title")
		}
	})

	t.Run("Text", func(t *testing.T) {
		path, err := svc.ExportLibrary(ctx, service.ShareFormatText, tmpDir)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !strings.HasSuffix(path, ".txt") {
			t.Errorf("expected .txt extension, got %q", path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read exported file: %v", err)
		}
		content := string(data)
		if !strings.Contains(content, "Go Programming") {
			t.Error("expected text to contain book title")
		}
	})

	t.Run("InvalidDir", func(t *testing.T) {
		_, err := svc.ExportLibrary(ctx, service.ShareFormatText, "/nonexistent/dir")
		if err == nil {
			t.Fatal("expected error for invalid directory")
		}
	})
}

func TestSharingService_ExportBookInfo(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	bookRepo := &mockSharingBookRepo{
		books: []*domain.Book{
			{ID: "b1", Title: "Go Programming", Author: "Author A", Format: domain.FormatPDF, TotalPages: 300},
		},
	}
	sheetRepo := newMockSharingSheetRepo()
	svc := service.NewSharingService(bookRepo, sheetRepo)

	t.Run("Markdown", func(t *testing.T) {
		path, err := svc.ExportBookInfo(ctx, "b1", service.ShareFormatMarkdown, tmpDir)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !strings.HasSuffix(path, ".md") {
			t.Errorf("expected .md extension, got %q", path)
		}
	})

	t.Run("JSON", func(t *testing.T) {
		path, err := svc.ExportBookInfo(ctx, "b1", service.ShareFormatJSON, tmpDir)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !strings.HasSuffix(path, ".json") {
			t.Errorf("expected .json extension, got %q", path)
		}
	})

	t.Run("Text", func(t *testing.T) {
		path, err := svc.ExportBookInfo(ctx, "b1", service.ShareFormatText, tmpDir)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !strings.HasSuffix(path, ".txt") {
			t.Errorf("expected .txt extension, got %q", path)
		}
	})

	t.Run("BookNotFound", func(t *testing.T) {
		_, err := svc.ExportBookInfo(ctx, "nonexistent", service.ShareFormatJSON, tmpDir)
		if err == nil {
			t.Fatal("expected error for nonexistent book")
		}
	})
}

func TestSharingService_ExportReadingSheet(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	bookRepo := &mockSharingBookRepo{}
	sheetRepo := newMockSharingSheetRepo()

	sheet, _ := domain.NewReadingSheet("b1", "Go Programming", "A great book", 5, []string{"Simplicity is key"}, []string{"go", "programming"})
	sheetRepo.sheets[sheet.ID] = sheet

	svc := service.NewSharingService(bookRepo, sheetRepo)

	t.Run("Markdown", func(t *testing.T) {
		path, err := svc.ExportReadingSheet(ctx, sheet.ID, service.ShareFormatMarkdown, tmpDir)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		data, _ := os.ReadFile(path)
		content := string(data)
		if !strings.Contains(content, "Fiche de lecture") {
			t.Error("expected markdown heading in exported sheet")
		}
		if !strings.Contains(content, "Simplicity is key") {
			t.Error("expected quote in exported sheet")
		}
	})

	t.Run("JSON", func(t *testing.T) {
		path, err := svc.ExportReadingSheet(ctx, sheet.ID, service.ShareFormatJSON, tmpDir)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !strings.HasSuffix(path, ".json") {
			t.Errorf("expected .json extension, got %q", path)
		}
	})

	t.Run("Text", func(t *testing.T) {
		path, err := svc.ExportReadingSheet(ctx, sheet.ID, service.ShareFormatText, tmpDir)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !strings.HasSuffix(path, ".txt") {
			t.Errorf("expected .txt extension, got %q", path)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := svc.ExportReadingSheet(ctx, "nonexistent", service.ShareFormatJSON, tmpDir)
		if err == nil {
			t.Fatal("expected error for nonexistent sheet")
		}
	})
}

func TestSharingService_ExportLibraryWithSheet(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	bookRepo := &mockSharingBookRepo{
		books: []*domain.Book{
			{ID: "b1", Title: "Test Book", Author: "", Format: domain.FormatPDF, TotalPages: 100},
		},
	}
	sheetRepo := newMockSharingSheetRepo()
	svc := service.NewSharingService(bookRepo, sheetRepo)

	// Export book with no author (tests orUnknown helper)
	path, err := svc.ExportLibrary(ctx, service.ShareFormatMarkdown, tmpDir)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "Inconnu") {
		t.Error("expected 'Inconnu' for missing author")
	}
}

func TestSharingService_ExportBookInfoWithSheet(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	bookRepo := &mockSharingBookRepo{
		books: []*domain.Book{
			{ID: "b1", Title: "Rated Book", Author: "Author", Format: domain.FormatEPUB, TotalPages: 150},
		},
	}
	sheetRepo := newMockSharingSheetRepo()
	sheet, _ := domain.NewReadingSheet("b1", "Rated Book", "Summary", 4, []string{"A quote"}, []string{"tag"})
	sheetRepo.sheets[sheet.ID] = sheet

	svc := service.NewSharingService(bookRepo, sheetRepo)

	// Export as text to test buildBookText and buildSheetText paths
	path, err := svc.ExportBookInfo(ctx, "b1", service.ShareFormatText, tmpDir)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "Rated Book") {
		t.Error("expected book title in text export")
	}

	// Clean up all generated files
	files, _ := filepath.Glob(filepath.Join(tmpDir, "*"))
	for _, f := range files {
		os.Remove(f)
	}
}
