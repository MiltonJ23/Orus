package service

import (
	"context"
	"fmt"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/port"
)

// ReadingSheetService handles CRUD operations for reading sheets.
type ReadingSheetService struct {
	sheetRepo port.ReadingSheetRepository
	bookRepo  port.BookRepository
}

// NewReadingSheetService creates a new ReadingSheetService with the given dependencies.
func NewReadingSheetService(sheetRepo port.ReadingSheetRepository, bookRepo port.BookRepository) *ReadingSheetService {
	return &ReadingSheetService{sheetRepo: sheetRepo, bookRepo: bookRepo}
}

// CreateSheet crée et persiste une nouvelle fiche de lecture
func (s *ReadingSheetService) CreateSheet(ctx context.Context, bookID, summary string, rating int, quotes, tags []string) (*domain.ReadingSheet, error) {
	// On récupère le titre du livre pour le dénormaliser
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("book not found: %w", err)
	}

	sheet, err := domain.NewReadingSheet(bookID, book.Title, summary, rating, quotes, tags)
	if err != nil {
		return nil, fmt.Errorf("invalid reading sheet data: %w", err)
	}

	if err := s.sheetRepo.SaveSheet(ctx, sheet); err != nil {
		return nil, fmt.Errorf("failed to persist reading sheet: %w", err)
	}
	return sheet, nil
}

// GetSheetForBook retourne la fiche d'un livre (nil si elle n'existe pas encore)
func (s *ReadingSheetService) GetSheetForBook(ctx context.Context, bookID string) (*domain.ReadingSheet, error) {
	sheet, err := s.sheetRepo.GetSheetByBookID(ctx, bookID)
	if err == domain.ErrReadingSheetNotFound {
		return nil, nil // pas d'erreur : la fiche n'existe pas encore
	}
	return sheet, err
}

// ListSheets retourne toutes les fiches de lecture
func (s *ReadingSheetService) ListSheets(ctx context.Context) ([]*domain.ReadingSheet, error) {
	return s.sheetRepo.ListAllSheets(ctx)
}

// UpdateSummary met à jour le résumé d'une fiche
func (s *ReadingSheetService) UpdateSummary(ctx context.Context, sheetID, newSummary string) error {
	sheet, err := s.sheetRepo.GetSheetByID(ctx, sheetID)
	if err != nil {
		return err
	}
	sheet.UpdateSummary(newSummary)
	return s.sheetRepo.UpdateSheet(ctx, sheet)
}

// SetRating met à jour la note d'une fiche
func (s *ReadingSheetService) SetRating(ctx context.Context, sheetID string, rating int) error {
	sheet, err := s.sheetRepo.GetSheetByID(ctx, sheetID)
	if err != nil {
		return err
	}
	if err := sheet.UpdateRating(rating); err != nil {
		return err
	}
	return s.sheetRepo.UpdateSheet(ctx, sheet)
}

// AddQuote ajoute une citation à une fiche existante
func (s *ReadingSheetService) AddQuote(ctx context.Context, sheetID, quote string) error {
	sheet, err := s.sheetRepo.GetSheetByID(ctx, sheetID)
	if err != nil {
		return err
	}
	sheet.AddQuote(quote)
	return s.sheetRepo.UpdateSheet(ctx, sheet)
}

// DeleteSheet supprime une fiche de lecture
func (s *ReadingSheetService) DeleteSheet(ctx context.Context, sheetID string) error {
	return s.sheetRepo.DeleteSheet(ctx, sheetID)
}
