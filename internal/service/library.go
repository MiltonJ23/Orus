package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/port"
)

var ErrBookAlreadyExists = errors.New("book already exists")

type LibraryService struct {
	repo      port.BookRepository
	extractor port.MetadataExtractor
}

func NewLibraryService(repo port.BookRepository, extractor port.MetadataExtractor) *LibraryService {
	return &LibraryService{repo, extractor}
}

// ImportBook imports a single book by file path.
func (l *LibraryService) ImportBook(ctx context.Context, filePath string) (*domain.Book, error) {
	log.Printf("[Import] Tentative : %s", filePath)

	metadata, err := l.extractor.ExtractInfo(ctx, filePath)
	if err != nil {
		log.Printf("[Import] Echec extraction metadonnees : %v", err)
		return nil, fmt.Errorf("extraction metadonnees : %w", err)
	}
	log.Printf("[Import] Metadonnees OK — titre=%q auteur=%q pages=%d", metadata.Title, metadata.Author, metadata.TotalPages)

	book, err := domain.NewBook(metadata.Title, metadata.Author, metadata.FilePath, metadata.Format, metadata.TotalPages)
	if err != nil {
		log.Printf("[Import] Echec creation domaine : %v", err)
		return nil, fmt.Errorf("creation livre : %w", err)
	}

	if err := l.repo.Save(ctx, book); err != nil {
		log.Printf("[Import] Echec sauvegarde BDD : %v", err)
		return nil, fmt.Errorf("sauvegarde BDD : %w", err)
	}

	log.Printf("[Import] Succes : %q (id=%s)", book.Title, book.ID)
	return book, nil
}

// ImportBooks imports multiple books; returns successes and per-file errors.
func (l *LibraryService) ImportBooks(ctx context.Context, filePaths []string) ([]*domain.Book, []error) {
	log.Printf("[Import] %d fichier(s) recu(s)", len(filePaths))
	var books []*domain.Book
	var errs []error
	for _, fp := range filePaths {
		b, err := l.ImportBook(ctx, fp)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s : %w", fp, err))
		} else {
			books = append(books, b)
		}
	}
	return books, errs
}

// GetLibrary returns all books.
func (l *LibraryService) GetLibrary(ctx context.Context) ([]*domain.Book, error) {
	return l.repo.ListAll(ctx)
}

// DeleteBook permanently removes a book.
func (l *LibraryService) DeleteBook(ctx context.Context, bookID string) error {
	return l.repo.Delete(ctx, bookID)
}
