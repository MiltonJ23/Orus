package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrBookNotFound indicates a book was not found in the repository.
var (
	ErrBookNotFound     = errors.New("book not found")
	ErrInvalidBookTitle = errors.New("invalid book title")
)

// BookFormat represents the file format of a book.
type BookFormat string

const (
	// FormatPDF represents a PDF file.
	FormatPDF BookFormat = "PDF"
	// FormatEPUB represents an EPUB file.
	FormatEPUB BookFormat = "EPUB"
	// FormatMOBI represents a MOBI file (not yet supported).
	FormatMOBI BookFormat = "MOBI"
)

// Book represents an imported book in the user's library.
type Book struct {
	ID         string
	Title      string
	Author     string
	FilePath   string
	Format     BookFormat
	TotalPages int
	CoverImage []byte // Optional: Store cover image as bytes, can be nil if not available
	AddedAt    time.Time
	UpdatedAt  time.Time
}

// NewBook creates a new Book with validated fields. Returns an error if title or
// filePath is empty.
func NewBook(title, author, filePath string, format BookFormat, totalPages int) (*Book, error) {
	if title == "" {
		return nil, ErrInvalidBookTitle
	}
	if filePath == "" {
		return nil, ErrBookNotFound
	}
	return &Book{
		ID:         uuid.New().String(),
		Title:      title,
		Author:     author,
		FilePath:   filePath,
		Format:     format,
		TotalPages: totalPages,
		CoverImage: []byte{}, // This can be set later using a method to update the cover image
		AddedAt:    time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

// BookMetadata holds the metadata extracted from a book file.
type BookMetadata struct {
	Title      string
	Author     string
	TotalPages int
	FilePath   string
	Format     BookFormat
}
