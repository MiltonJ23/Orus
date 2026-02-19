package extractor_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/MiltonJ23/Orus/internal/adapters/extractor"
	"github.com/MiltonJ23/Orus/internal/domain"
)

func TestLocalFileExtractor_ExtractInfo(t *testing.T) {
	ext := extractor.NewLocalFileExtractor()
	ctx := context.Background()

	// Ensure you have these files in your testdata directory!
	pdfPath := filepath.Join("testdata", "dummy.pdf")
	epubPath := filepath.Join("testdata", "dummy.epub")

	t.Run("PDF Extraction", func(t *testing.T) {
		meta, err := ext.ExtractInfo(ctx, pdfPath)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if meta.Format != domain.FormatPDF {
			t.Errorf("expected format PDF, got %v", meta.Format)
		}
		if meta.TotalPages <= 0 {
			t.Errorf("expected pages > 0, got %d", meta.TotalPages)
		}
	})

	t.Run("EPUB Extraction", func(t *testing.T) {
		meta, err := ext.ExtractInfo(ctx, epubPath)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if meta.Format != domain.FormatEPUB {
			t.Errorf("expected format EPUB, got %v", meta.Format)
		}
		if meta.TotalPages <= 0 {
			t.Errorf("expected pages > 0, got %d", meta.TotalPages)
		}
	})

	t.Run("Unsupported Format", func(t *testing.T) {
		_, err := ext.ExtractInfo(ctx, "testdata/dummy.txt")
		if err == nil {
			t.Fatal("expected unsupported format error, got nil")
		}
	})
}
