package extractor

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/kapmahc/epub"
	"github.com/ledongthuc/pdf"
)

var (
	ErrUnsupportedFileFormat = errors.New("unsupported file format")
	ErrCorruptFile           = errors.New("file is corrupted or unreadable")
)

type LocalFileExtractor struct{}

func NewLocalFileExtractor() *LocalFileExtractor {
	return &LocalFileExtractor{}
}

func (l *LocalFileExtractor) ExtractInfo(ctx context.Context, filePath string) (*domain.BookMetadata, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		return l.ExtractPDF(filePath)
	case ".epub":
		return l.extractEPUB(filePath)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedFileFormat, ext)
	}
}

func (l *LocalFileExtractor) ExtractPDF(filePath string) (*domain.BookMetadata, error) {
	// let's open the file using the pure-Go library
	file, reader, pdfOpeningError := pdf.Open(filePath)
	if pdfOpeningError != nil {
		return nil, fmt.Errorf("an error occured trying to open the pdf file")
	}
	defer file.Close()

	// let's start to extract basic data
	totalPages := reader.NumPage()
	// fallback title to filename if pdf metadata is empty
	filename := filepath.Base(filePath)
	title := strings.TrimSuffix(filename, filepath.Ext(filename))

	return &domain.BookMetadata{
		Title:      title,
		Author:     "Unknown",
		TotalPages: totalPages,
		FilePath:   filePath,
		Format:     domain.FormatPDF,
	}, nil
}

func (e *LocalFileExtractor) extractEPUB(filePath string) (*domain.BookMetadata, error) {
	// Open EPUB
	book, epubOpeningerr := epub.Open(filePath)
	if epubOpeningerr != nil {
		return nil, fmt.Errorf("an error occured trying to open file,%w: %v", ErrCorruptFile, epubOpeningerr)
	}
	defer book.Close()

	// EPUBs have explicit metadata in their OPF file
	title := "Unknown Title"
	if len(book.Opf.Metadata.Title) > 0 {
		title = book.Opf.Metadata.Title[0]
	}

	author := "Unknown Author"
	if len(book.Opf.Metadata.Creator) > 0 {
		author = book.Opf.Metadata.Creator[0].Data
	}

	spineCount := len(book.Opf.Spine.Items)
	if spineCount == 0 {
		spineCount = 1
	}

	return &domain.BookMetadata{
		Title:      title,
		Author:     author,
		FilePath:   filePath,
		Format:     domain.FormatEPUB,
		TotalPages: spineCount,
	}, nil
}
