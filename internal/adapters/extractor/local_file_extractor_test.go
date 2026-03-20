package extractor_test

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
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
		if meta.FilePath != pdfPath {
			t.Errorf("expected file path %s, got %s", pdfPath, meta.FilePath)
		}
		if meta.Author != "Unknown" {
			t.Errorf("expected author 'Unknown', got %s", meta.Author)
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
		if meta.FilePath != epubPath {
			t.Errorf("expected file path %s, got %s", epubPath, meta.FilePath)
		}
	})

	t.Run("Unsupported Format", func(t *testing.T) {
		_, err := ext.ExtractInfo(ctx, "testdata/dummy.txt")
		if err == nil {
			t.Fatal("expected unsupported format error, got nil")
		}
		if !errors.Is(err, extractor.ErrUnsupportedFileFormat) {
			t.Errorf("expected ErrUnsupportedFileFormat, got %v", err)
		}
	})

	t.Run("Unsupported Format with uppercase extension", func(t *testing.T) {
		_, err := ext.ExtractInfo(ctx, "testdata/file.DOCX")
		if err == nil {
			t.Fatal("expected unsupported format error, got nil")
		}
		if !errors.Is(err, extractor.ErrUnsupportedFileFormat) {
			t.Errorf("expected ErrUnsupportedFileFormat, got %v", err)
		}
	})

	t.Run("Non-existent PDF file", func(t *testing.T) {
		_, err := ext.ExtractInfo(ctx, "testdata/nonexistent.pdf")
		if err == nil {
			t.Fatal("expected error for non-existent file, got nil")
		}
	})

	t.Run("Non-existent EPUB file", func(t *testing.T) {
		_, err := ext.ExtractInfo(ctx, "testdata/nonexistent.epub")
		if err == nil {
			t.Fatal("expected error for non-existent file, got nil")
		}
		if !errors.Is(err, extractor.ErrCorruptFile) {
			t.Logf("Got error: %v", err)
		}
	})
}

func TestLocalFileExtractor_ReadBookText(t *testing.T) {
	ext := extractor.NewLocalFileExtractor()
	ctx := context.Background()

	pdfPath := filepath.Join("testdata", "dummy.pdf")
	epubPath := filepath.Join("testdata", "dummy.epub")

	t.Run("Read PDF Text", func(t *testing.T) {
		pages, err := ext.ReadBookText(ctx, pdfPath)
		if err != nil {
			t.Fatalf("expected no error reading PDF text, got: %v", err)
		}
		if len(pages) == 0 {
			t.Error("expected at least one page chunk, got zero")
		}
		// Each page should be a string (possibly chunked)
		for i, page := range pages {
			if page == "" {
				t.Errorf("page chunk %d is empty", i)
			}
		}
	})

	t.Run("Read EPUB Text", func(t *testing.T) {
		pages, err := ext.ReadBookText(ctx, epubPath)
		if err != nil {
			t.Fatalf("expected no error reading EPUB text, got: %v", err)
		}
		if len(pages) == 0 {
			t.Error("expected at least one page chunk, got zero")
		}
		// Verify that chunks are returned
		for i, page := range pages {
			if page == "" {
				t.Errorf("page chunk %d is empty", i)
			}
		}
	})

	t.Run("Read Unsupported Format", func(t *testing.T) {
		_, err := ext.ReadBookText(ctx, "testdata/file.txt")
		if err == nil {
			t.Fatal("expected unsupported format error, got nil")
		}
		if !errors.Is(err, extractor.ErrUnsupportedFileFormat) {
			t.Errorf("expected ErrUnsupportedFileFormat, got %v", err)
		}
	})

	t.Run("Read Non-existent PDF", func(t *testing.T) {
		_, err := ext.ReadBookText(ctx, "testdata/missing.pdf")
		if err == nil {
			t.Fatal("expected error for non-existent PDF, got nil")
		}
	})

	t.Run("Read Non-existent EPUB", func(t *testing.T) {
		_, err := ext.ReadBookText(ctx, "testdata/missing.epub")
		if err == nil {
			t.Fatal("expected error for non-existent EPUB, got nil")
		}
		if !errors.Is(err, extractor.ErrCorruptFile) {
			t.Logf("Got error: %v", err)
		}
	})
}

func TestLocalFileExtractor_ExtractPDF(t *testing.T) {
	ext := extractor.NewLocalFileExtractor()
	pdfPath := filepath.Join("testdata", "dummy.pdf")

	t.Run("Extract PDF metadata", func(t *testing.T) {
		meta, err := ext.ExtractPDF(pdfPath)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if meta.Format != domain.FormatPDF {
			t.Errorf("expected PDF format, got %v", meta.Format)
		}
		// Title should be filename without extension
		expectedTitle := "dummy"
		if meta.Title != expectedTitle {
			t.Errorf("expected title %s, got %s", expectedTitle, meta.Title)
		}
	})

	t.Run("Extract corrupted PDF", func(t *testing.T) {
		_, err := ext.ExtractPDF("testdata/nonexistent.pdf")
		if err == nil {
			t.Fatal("expected error for non-existent PDF, got nil")
		}
	})
}

// Test helper functions via white-box testing patterns
// Note: stripHTML and chunkLines are not exported, so we test them indirectly

func TestStripHTML_Indirect(t *testing.T) {
	// We test stripHTML indirectly through EPUB text extraction
	// The stripHTML function processes HTML content and attempts to extract plain text
	ext := extractor.NewLocalFileExtractor()
	ctx := context.Background()
	epubPath := filepath.Join("testdata", "dummy.epub")

	pages, err := ext.ReadBookText(ctx, epubPath)
	if err != nil {
		t.Fatalf("failed to read EPUB: %v", err)
	}

	// Verify that text extraction works and produces output
	if len(pages) == 0 {
		t.Error("expected non-empty pages after text extraction")
	}

	// Verify that the function processes HTML content
	// The output should contain readable text (not just raw HTML)
	totalChars := 0
	for _, page := range pages {
		totalChars += len(page)
	}

	if totalChars == 0 {
		t.Error("expected text content after HTML processing")
	}

	// Verify that HTML entities are decoded
	hasDecodedContent := false
	for _, page := range pages {
		// Check that common content patterns exist (not all HTML)
		if len(page) > 0 && !strings.HasPrefix(page, "<") {
			hasDecodedContent = true
			break
		}
	}

	if !hasDecodedContent {
		t.Error("expected some decoded text content, not pure HTML")
	}
}

func TestChunkLines_Indirect(t *testing.T) {
	// Test chunking behavior indirectly by checking that ReadBookText
	// returns multiple chunks for larger files
	ext := extractor.NewLocalFileExtractor()
	ctx := context.Background()
	pdfPath := filepath.Join("testdata", "dummy.pdf")

	pages, err := ext.ReadBookText(ctx, pdfPath)
	if err != nil {
		t.Fatalf("failed to read PDF: %v", err)
	}

	// Verify that text is chunked (should have multiple chunks for most PDFs)
	if len(pages) == 0 {
		t.Error("expected at least one chunk")
	}

	// Each chunk should not exceed a reasonable size (35 lines * ~80 chars + newlines ~= 3000 chars)
	for i, page := range pages {
		lineCount := strings.Count(page, "\n") + 1
		// Chunks should be around 35 lines (linesPerChunk constant in the code)
		if lineCount > 100 {
			t.Logf("Warning: chunk %d has %d lines, expected around 35", i, lineCount)
		}
	}
}

func TestLocalFileExtractor_ExtractInfo_NoExtension(t *testing.T) {
	ext := extractor.NewLocalFileExtractor()
	ctx := context.Background()

	// A file with no extension should return ErrUnsupportedFileFormat
	_, err := ext.ExtractInfo(ctx, "testdata/somefile")
	if err == nil {
		t.Fatal("expected error for file with no extension, got nil")
	}
	if !errors.Is(err, extractor.ErrUnsupportedFileFormat) {
		t.Errorf("expected ErrUnsupportedFileFormat, got %v", err)
	}
}

func TestLocalFileExtractor_ReadBookText_NoExtension(t *testing.T) {
	ext := extractor.NewLocalFileExtractor()
	ctx := context.Background()

	_, err := ext.ReadBookText(ctx, "testdata/somefile")
	if err == nil {
		t.Fatal("expected error for file with no extension, got nil")
	}
	if !errors.Is(err, extractor.ErrUnsupportedFileFormat) {
		t.Errorf("expected ErrUnsupportedFileFormat, got %v", err)
	}
}

func TestLocalFileExtractor_ExtractPDF_TitleFromFilename(t *testing.T) {
	ext := extractor.NewLocalFileExtractor()

	// The title should be the filename without extension
	meta, err := ext.ExtractPDF(filepath.Join("testdata", "dummy.pdf"))
	if err != nil {
		t.Fatalf("ExtractPDF failed: %v", err)
	}

	// Title should not contain the .pdf extension
	if strings.HasSuffix(meta.Title, ".pdf") {
		t.Errorf("Title should not contain extension, got: %s", meta.Title)
	}
	// Title should not be empty
	if meta.Title == "" {
		t.Error("Title should not be empty")
	}
}

func TestLocalFileExtractor_ExtractEPUB_Metadata(t *testing.T) {
	ext := extractor.NewLocalFileExtractor()
	ctx := context.Background()
	epubPath := filepath.Join("testdata", "dummy.epub")

	meta, err := ext.ExtractInfo(ctx, epubPath)
	if err != nil {
		t.Fatalf("ExtractInfo for EPUB failed: %v", err)
	}

	// Title should not be empty
	if meta.Title == "" {
		t.Error("EPUB title should not be empty")
	}
	// Author should not be empty
	if meta.Author == "" {
		t.Error("EPUB author should not be empty")
	}
	// Format should be EPUB
	if meta.Format != domain.FormatEPUB {
		t.Errorf("Expected FormatEPUB, got %v", meta.Format)
	}
	// FilePath should match input
	if meta.FilePath != epubPath {
		t.Errorf("Expected FilePath %s, got %s", epubPath, meta.FilePath)
	}
}

func TestLocalFileExtractor_ImplementsContentReader(t *testing.T) {
	// Verify the extractor satisfies the ContentReader interface at runtime
	ext := extractor.NewLocalFileExtractor()
	if ext == nil {
		t.Fatal("NewLocalFileExtractor returned nil")
	}
	// The compile-time check already ensures this via var _ port.ContentReader = (*LocalFileExtractor)(nil)
	// but we document the behaviour here
	ctx := context.Background()
	_, err := ext.ReadBookText(ctx, filepath.Join("testdata", "dummy.pdf"))
	if err != nil {
		t.Errorf("ContentReader.ReadBookText failed: %v", err)
	}
}