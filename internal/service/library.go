package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/port"
)

var (
	ErrBookAlreadyExists = errors.New("book already exists")
)

type LibraryService struct {
	repo      port.BookRepository
	extractor port.MetadataExtractor
}

func NewLibraryService(repo port.BookRepository, extractor port.MetadataExtractor) *LibraryService {
	return &LibraryService{repo, extractor}
}

func (l *LibraryService) ImportBook(ctx context.Context, file_path string) (*domain.Book, error) {
	// Let's extract metadata
	metadata, extractionError := l.extractor.ExtractInfo(ctx, file_path)
	if extractionError != nil {
		return nil, fmt.Errorf("an error occured during book metadata extraction, %v", extractionError)
	}

	// we use the factory method to create a book
	book, bookCreationError := domain.NewBook(metadata.Title, metadata.Author, metadata.FilePath, metadata.Format, metadata.TotalPages)
	if bookCreationError != nil {
		return nil, fmt.Errorf("an error occured during book creation, %v", bookCreationError)
	}

	// we save the book to our db
	savingToDbError := l.repo.Save(ctx, book)
	if savingToDbError != nil {
		return nil, fmt.Errorf("an error occured during saving book to database, %v", savingToDbError)
	}
	return book, nil
}

func (l *LibraryService) GetLibrary(ctx context.Context) ([]*domain.Book, error) {
	return l.repo.ListAll(ctx)
}
