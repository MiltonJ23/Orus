package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/service"
)

// --- MOCKS FOR ANNOTATION SERVICE ---

type mockAnnotationRepo struct {
	annotations []*domain.Annotation
	failSave    bool
	failList    bool
	failByPage  bool
	failByType  bool
	failDelete  bool
}

func (m *mockAnnotationRepo) SaveAnnotation(_ context.Context, annot *domain.Annotation) error {
	if m.failSave {
		return errors.New("db save error")
	}
	m.annotations = append(m.annotations, annot)
	return nil
}

func (m *mockAnnotationRepo) GetAnnotationByPage(_ context.Context, pageNo int, bookID string) ([]*domain.Annotation, error) {
	if m.failByPage {
		return nil, errors.New("db query error")
	}
	var result []*domain.Annotation
	for _, a := range m.annotations {
		if a.PageNo == pageNo && a.BookID == bookID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockAnnotationRepo) GetAnnotationByType(_ context.Context, annotationType string) ([]*domain.Annotation, error) {
	if m.failByType {
		return nil, errors.New("db query error")
	}
	var result []*domain.Annotation
	for _, a := range m.annotations {
		if string(a.AnnotationType) == annotationType {
			result = append(result, a)
		}
	}
	return result, nil
}

func (m *mockAnnotationRepo) DeleteAnnotation(_ context.Context, id string) error {
	if m.failDelete {
		return errors.New("db delete error")
	}
	for i, a := range m.annotations {
		if a.ID == id {
			m.annotations = append(m.annotations[:i], m.annotations[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockAnnotationRepo) ListAllAnnotationOfABook(_ context.Context, bookID string) ([]*domain.Annotation, error) {
	if m.failList {
		return nil, errors.New("db list error")
	}
	var result []*domain.Annotation
	for _, a := range m.annotations {
		if a.BookID == bookID {
			result = append(result, a)
		}
	}
	return result, nil
}

type mockAnnotBookRepo struct {
	books    map[string]*domain.Book
	failGet  bool
	failList bool
}

func (m *mockAnnotBookRepo) Save(_ context.Context, book *domain.Book) error {
	if m.books == nil {
		m.books = make(map[string]*domain.Book)
	}
	m.books[book.ID] = book
	return nil
}

func (m *mockAnnotBookRepo) GetByID(_ context.Context, id string) (*domain.Book, error) {
	if m.failGet {
		return nil, errors.New("book not found")
	}
	if m.books != nil {
		if b, ok := m.books[id]; ok {
			return b, nil
		}
	}
	return nil, errors.New("book not found")
}

func (m *mockAnnotBookRepo) ListAll(_ context.Context) ([]*domain.Book, error) {
	if m.failList {
		return nil, errors.New("db list error")
	}
	var result []*domain.Book
	for _, b := range m.books {
		result = append(result, b)
	}
	return result, nil
}

func (m *mockAnnotBookRepo) Delete(_ context.Context, id string) error {
	delete(m.books, id)
	return nil
}

// --- TESTS ---

func TestAnnotationService_AddAnnotation(t *testing.T) {
	ctx := context.Background()

	makeBookRepo := func() *mockAnnotBookRepo {
		return &mockAnnotBookRepo{
			books: map[string]*domain.Book{
				"book-1": {ID: "book-1", Title: "Test Book", TotalPages: 200},
			},
		}
	}

	t.Run("Success - Bookmark", func(t *testing.T) {
		bookRepo := makeBookRepo()
		annotRepo := &mockAnnotationRepo{}
		svc := service.NewAnnotationService(annotRepo, bookRepo)

		annot, err := svc.AddAnnotation(ctx, "book-1", domain.AnnotationBookmark, 42)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if annot.BookID != "book-1" {
			t.Errorf("expected book ID 'book-1', got: %s", annot.BookID)
		}
		if annot.AnnotationType != domain.AnnotationBookmark {
			t.Errorf("expected bookmark type, got: %s", annot.AnnotationType)
		}
		if annot.PageNo != 42 {
			t.Errorf("expected page 42, got: %d", annot.PageNo)
		}
		if len(annotRepo.annotations) != 1 {
			t.Errorf("expected 1 stored annotation, got: %d", len(annotRepo.annotations))
		}
	})

	t.Run("Success - Highlight", func(t *testing.T) {
		bookRepo := makeBookRepo()
		annotRepo := &mockAnnotationRepo{}
		svc := service.NewAnnotationService(annotRepo, bookRepo)

		annot, err := svc.AddAnnotation(ctx, "book-1", domain.AnnotationHighlight, 10)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if annot.AnnotationType != domain.AnnotationHighlight {
			t.Errorf("expected highlight type, got: %s", annot.AnnotationType)
		}
	})

	t.Run("Error - Book Not Found", func(t *testing.T) {
		bookRepo := makeBookRepo()
		annotRepo := &mockAnnotationRepo{}
		svc := service.NewAnnotationService(annotRepo, bookRepo)

		_, err := svc.AddAnnotation(ctx, "nonexistent", domain.AnnotationBookmark, 1)
		if err == nil {
			t.Fatal("expected error for nonexistent book, got nil")
		}
	})

	t.Run("Error - Page Exceeds Total", func(t *testing.T) {
		bookRepo := makeBookRepo()
		annotRepo := &mockAnnotationRepo{}
		svc := service.NewAnnotationService(annotRepo, bookRepo)

		_, err := svc.AddAnnotation(ctx, "book-1", domain.AnnotationBookmark, 999)
		if err == nil {
			t.Fatal("expected error for page exceeding total, got nil")
		}
	})

	t.Run("Error - Invalid Page Number", func(t *testing.T) {
		bookRepo := makeBookRepo()
		annotRepo := &mockAnnotationRepo{}
		svc := service.NewAnnotationService(annotRepo, bookRepo)

		_, err := svc.AddAnnotation(ctx, "book-1", domain.AnnotationBookmark, 0)
		if err == nil {
			t.Fatal("expected error for invalid page, got nil")
		}
	})

	t.Run("Error - Save Failure", func(t *testing.T) {
		bookRepo := makeBookRepo()
		annotRepo := &mockAnnotationRepo{failSave: true}
		svc := service.NewAnnotationService(annotRepo, bookRepo)

		_, err := svc.AddAnnotation(ctx, "book-1", domain.AnnotationBookmark, 1)
		if err == nil {
			t.Fatal("expected save error, got nil")
		}
	})

	t.Run("Error - Book Repo Failure", func(t *testing.T) {
		bookRepo := &mockAnnotBookRepo{failGet: true}
		annotRepo := &mockAnnotationRepo{}
		svc := service.NewAnnotationService(annotRepo, bookRepo)

		_, err := svc.AddAnnotation(ctx, "book-1", domain.AnnotationBookmark, 1)
		if err == nil {
			t.Fatal("expected book repo error, got nil")
		}
	})
}

func TestAnnotationService_ListAnnotationsForBook(t *testing.T) {
	ctx := context.Background()

	t.Run("Success - With Annotations", func(t *testing.T) {
		bookRepo := &mockAnnotBookRepo{
			books: map[string]*domain.Book{
				"book-1": {ID: "book-1", Title: "Test Book", TotalPages: 200},
			},
		}
		annotRepo := &mockAnnotationRepo{}
		svc := service.NewAnnotationService(annotRepo, bookRepo)

		// Add multiple annotations
		_, _ = svc.AddAnnotation(ctx, "book-1", domain.AnnotationBookmark, 10)
		_, _ = svc.AddAnnotation(ctx, "book-1", domain.AnnotationHighlight, 20)
		_, _ = svc.AddAnnotation(ctx, "book-1", domain.AnnotationBookmark, 30)

		annots, err := svc.ListAnnotationsForBook(ctx, "book-1")
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if len(annots) != 3 {
			t.Errorf("expected 3 annotations, got: %d", len(annots))
		}
	})

	t.Run("Success - No Annotations", func(t *testing.T) {
		annotRepo := &mockAnnotationRepo{}
		svc := service.NewAnnotationService(annotRepo, nil)

		annots, err := svc.ListAnnotationsForBook(ctx, "book-1")
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if len(annots) != 0 {
			t.Errorf("expected 0 annotations, got: %d", len(annots))
		}
	})

	t.Run("Error - Repo Failure", func(t *testing.T) {
		annotRepo := &mockAnnotationRepo{failList: true}
		svc := service.NewAnnotationService(annotRepo, nil)

		_, err := svc.ListAnnotationsForBook(ctx, "book-1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestAnnotationService_GetAnnotationsByPage(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		bookRepo := &mockAnnotBookRepo{
			books: map[string]*domain.Book{
				"book-1": {ID: "book-1", Title: "Test Book", TotalPages: 200},
			},
		}
		annotRepo := &mockAnnotationRepo{}
		svc := service.NewAnnotationService(annotRepo, bookRepo)

		_, _ = svc.AddAnnotation(ctx, "book-1", domain.AnnotationBookmark, 42)
		_, _ = svc.AddAnnotation(ctx, "book-1", domain.AnnotationHighlight, 42)
		_, _ = svc.AddAnnotation(ctx, "book-1", domain.AnnotationBookmark, 10)

		annots, err := svc.GetAnnotationsByPage(ctx, "book-1", 42)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if len(annots) != 2 {
			t.Errorf("expected 2 annotations on page 42, got: %d", len(annots))
		}
	})

	t.Run("Error - Repo Failure", func(t *testing.T) {
		annotRepo := &mockAnnotationRepo{failByPage: true}
		svc := service.NewAnnotationService(annotRepo, nil)

		_, err := svc.GetAnnotationsByPage(ctx, "book-1", 42)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestAnnotationService_GetAnnotationsByType(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		bookRepo := &mockAnnotBookRepo{
			books: map[string]*domain.Book{
				"book-1": {ID: "book-1", Title: "Test Book", TotalPages: 200},
			},
		}
		annotRepo := &mockAnnotationRepo{}
		svc := service.NewAnnotationService(annotRepo, bookRepo)

		_, _ = svc.AddAnnotation(ctx, "book-1", domain.AnnotationBookmark, 10)
		_, _ = svc.AddAnnotation(ctx, "book-1", domain.AnnotationBookmark, 20)
		_, _ = svc.AddAnnotation(ctx, "book-1", domain.AnnotationHighlight, 30)

		bookmarks, err := svc.GetAnnotationsByType(ctx, domain.AnnotationBookmark)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if len(bookmarks) != 2 {
			t.Errorf("expected 2 bookmarks, got: %d", len(bookmarks))
		}

		highlights, err := svc.GetAnnotationsByType(ctx, domain.AnnotationHighlight)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if len(highlights) != 1 {
			t.Errorf("expected 1 highlight, got: %d", len(highlights))
		}
	})

	t.Run("Error - Repo Failure", func(t *testing.T) {
		annotRepo := &mockAnnotationRepo{failByType: true}
		svc := service.NewAnnotationService(annotRepo, nil)

		_, err := svc.GetAnnotationsByType(ctx, domain.AnnotationBookmark)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestAnnotationService_DeleteAnnotation(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		bookRepo := &mockAnnotBookRepo{
			books: map[string]*domain.Book{
				"book-1": {ID: "book-1", Title: "Test Book", TotalPages: 200},
			},
		}
		annotRepo := &mockAnnotationRepo{}
		svc := service.NewAnnotationService(annotRepo, bookRepo)

		annot, _ := svc.AddAnnotation(ctx, "book-1", domain.AnnotationBookmark, 10)

		err := svc.DeleteAnnotation(ctx, annot.ID)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if len(annotRepo.annotations) != 0 {
			t.Errorf("expected 0 annotations after delete, got: %d", len(annotRepo.annotations))
		}
	})

	t.Run("Error - Empty ID", func(t *testing.T) {
		svc := service.NewAnnotationService(&mockAnnotationRepo{}, nil)

		err := svc.DeleteAnnotation(ctx, "")
		if err == nil {
			t.Fatal("expected error for empty ID, got nil")
		}
	})

	t.Run("Error - Repo Failure", func(t *testing.T) {
		annotRepo := &mockAnnotationRepo{failDelete: true}
		svc := service.NewAnnotationService(annotRepo, nil)

		err := svc.DeleteAnnotation(ctx, "some-id")
		if err == nil {
			t.Fatal("expected delete error, got nil")
		}
	})
}

func TestAnnotationService_CountAnnotationsForBook(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		bookRepo := &mockAnnotBookRepo{
			books: map[string]*domain.Book{
				"book-1": {ID: "book-1", Title: "Test Book", TotalPages: 200},
			},
		}
		annotRepo := &mockAnnotationRepo{}
		svc := service.NewAnnotationService(annotRepo, bookRepo)

		_, _ = svc.AddAnnotation(ctx, "book-1", domain.AnnotationBookmark, 10)
		_, _ = svc.AddAnnotation(ctx, "book-1", domain.AnnotationHighlight, 20)

		count, err := svc.CountAnnotationsForBook(ctx, "book-1")
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if count != 2 {
			t.Errorf("expected 2, got: %d", count)
		}
	})

	t.Run("Error - Repo Failure", func(t *testing.T) {
		annotRepo := &mockAnnotationRepo{failList: true}
		svc := service.NewAnnotationService(annotRepo, nil)

		_, err := svc.CountAnnotationsForBook(ctx, "book-1")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
