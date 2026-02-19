package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// First of all, let's create the custom errors that might be raised here

var (
	ErrBookNotFound     = errors.New("book not found")
	ErrInvalidBookTitle = errors.New("invalid book title")
)

type BookFormat string

const (
	FormatPDF  BookFormat = "PDF"
	FormatEPUB BookFormat = "EPUB"
	FormatMOBI BookFormat = "MOBI" // TODO: Add more formats as needed, this one shall be developed in the future
)

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

// Now let's create a factory method to create a new book, this will help us to ensure that the book is created with valid data

// NewBook is a factory method to create a new book, it validates the input data and returns an error if the data is invalid
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

// BookMetadata represent the metadata to be extracted from a file
type BookMetadata struct {
	Title      string
	Author     string
	TotalPages int
	filePath   string
	format     BookFormat
}
