package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/port"
)

var _ port.ReadingSheetRepository = (*Storage)(nil)

// SaveSheet insère une fiche de lecture en base
func (s *Storage) SaveSheet(ctx context.Context, sheet *domain.ReadingSheet) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	quotesStr := strings.Join(sheet.Quotes, "||")
	tagsStr := strings.Join(sheet.Tags, ",")

	query := `INSERT INTO reading_sheets (id, book_id, book_title, summary, quotes, rating, tags, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query,
		sheet.ID, sheet.BookID, sheet.BookTitle, sheet.Summary,
		quotesStr, sheet.Rating, tagsStr, sheet.CreatedAt, sheet.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save reading sheet: %w", err)
	}
	return nil
}

// GetSheetByID récupère une fiche par son ID
func (s *Storage) GetSheetByID(ctx context.Context, id string) (*domain.ReadingSheet, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, book_id, book_title, summary, quotes, rating, tags, created_at, updated_at FROM reading_sheets WHERE id = ?`
	row := s.db.QueryRowContext(ctx, query, id)
	sheet, err := scanSheet(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReadingSheetNotFound
		}
		return nil, fmt.Errorf("failed to get sheet by id: %w", err)
	}
	return sheet, nil
}

// GetSheetByBookID récupère la fiche associée à un livre
func (s *Storage) GetSheetByBookID(ctx context.Context, bookID string) (*domain.ReadingSheet, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, book_id, book_title, summary, quotes, rating, tags, created_at, updated_at FROM reading_sheets WHERE book_id = ? LIMIT 1`
	row := s.db.QueryRowContext(ctx, query, bookID)
	sheet, err := scanSheet(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReadingSheetNotFound
		}
		return nil, fmt.Errorf("failed to get sheet by book id: %w", err)
	}
	return sheet, nil
}

// ListAllSheets retourne toutes les fiches de lecture
func (s *Storage) ListAllSheets(ctx context.Context) ([]*domain.ReadingSheet, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, book_id, book_title, summary, quotes, rating, tags, created_at, updated_at FROM reading_sheets ORDER BY updated_at DESC`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list reading sheets: %w", err)
	}
	defer rows.Close()

	var sheets []*domain.ReadingSheet
	for rows.Next() {
		sheet, err := scanSheetRows(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reading sheet: %w", err)
		}
		sheets = append(sheets, sheet)
	}
	return sheets, rows.Err()
}

// UpdateSheet met à jour une fiche existante
func (s *Storage) UpdateSheet(ctx context.Context, sheet *domain.ReadingSheet) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sheet.UpdatedAt = time.Now()
	quotesStr := strings.Join(sheet.Quotes, "||")
	tagsStr := strings.Join(sheet.Tags, ",")

	query := `UPDATE reading_sheets SET summary=?, quotes=?, rating=?, tags=?, updated_at=? WHERE id=?`
	_, err := s.db.ExecContext(ctx, query, sheet.Summary, quotesStr, sheet.Rating, tagsStr, sheet.UpdatedAt, sheet.ID)
	if err != nil {
		return fmt.Errorf("failed to update reading sheet: %w", err)
	}
	return nil
}

// DeleteSheet supprime une fiche de lecture
func (s *Storage) DeleteSheet(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, `DELETE FROM reading_sheets WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete reading sheet: %w", err)
	}
	return nil
}

// --- helpers de scan ---

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSheet(row rowScanner) (*domain.ReadingSheet, error) {
	var sheet domain.ReadingSheet
	var quotesStr, tagsStr string
	err := row.Scan(&sheet.ID, &sheet.BookID, &sheet.BookTitle, &sheet.Summary, &quotesStr, &sheet.Rating, &tagsStr, &sheet.CreatedAt, &sheet.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if quotesStr != "" {
		sheet.Quotes = strings.Split(quotesStr, "||")
	}
	if tagsStr != "" {
		sheet.Tags = strings.Split(tagsStr, ",")
	}
	return &sheet, nil
}

func scanSheetRows(rows *sql.Rows) (*domain.ReadingSheet, error) {
	var sheet domain.ReadingSheet
	var quotesStr, tagsStr string
	err := rows.Scan(&sheet.ID, &sheet.BookID, &sheet.BookTitle, &sheet.Summary, &quotesStr, &sheet.Rating, &tagsStr, &sheet.CreatedAt, &sheet.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if quotesStr != "" {
		sheet.Quotes = strings.Split(quotesStr, "||")
	}
	if tagsStr != "" {
		sheet.Tags = strings.Split(tagsStr, ",")
	}
	return &sheet, nil
}
