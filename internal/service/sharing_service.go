package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/port"
)

// ShareFormat définit le format d'export
type ShareFormat string

const (
	ShareFormatJSON     ShareFormat = "json"
	ShareFormatMarkdown ShareFormat = "md"
	ShareFormatText     ShareFormat = "txt"
)

// SharingService gère l'export et le partage de livres et fiches
type SharingService struct {
	bookRepo  port.BookRepository
	sheetRepo port.ReadingSheetRepository
}

func NewSharingService(bookRepo port.BookRepository, sheetRepo port.ReadingSheetRepository) *SharingService {
	return &SharingService{bookRepo: bookRepo, sheetRepo: sheetRepo}
}

// ExportBookInfo exporte les informations d'un livre dans un fichier
func (s *SharingService) ExportBookInfo(ctx context.Context, bookID string, format ShareFormat, outputDir string) (string, error) {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return "", fmt.Errorf("book not found: %w", err)
	}

	// Récupérer la fiche si elle existe
	sheet, _ := s.sheetRepo.GetSheetByBookID(ctx, bookID)

	var content string
	var ext string

	switch format {
	case ShareFormatJSON:
		content, err = s.buildBookJSON(book, sheet)
		ext = "json"
	case ShareFormatMarkdown:
		content = s.buildBookMarkdown(book, sheet)
		ext = "md"
	default:
		content = s.buildBookText(book, sheet)
		ext = "txt"
	}

	if err != nil {
		return "", fmt.Errorf("failed to build export content: %w", err)
	}

	fileName := fmt.Sprintf("orus_%s_%s.%s",
		sanitizeFileName(book.Title),
		time.Now().Format("20060102"),
		ext,
	)
	filePath := filepath.Join(outputDir, fileName)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write export file: %w", err)
	}
	return filePath, nil
}

// ExportLibrary exporte toute la bibliothèque
func (s *SharingService) ExportLibrary(ctx context.Context, format ShareFormat, outputDir string) (string, error) {
	books, err := s.bookRepo.ListAll(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list books: %w", err)
	}

	var sb strings.Builder

	switch format {
	case ShareFormatMarkdown:
		sb.WriteString("# 📚 Ma bibliothèque Orus\n\n")
		sb.WriteString(fmt.Sprintf("*Exporté le %s*\n\n", time.Now().Format("02 janvier 2006")))
		sb.WriteString("---\n\n")
		for _, book := range books {
			sheet, _ := s.sheetRepo.GetSheetByBookID(ctx, book.ID)
			sb.WriteString(s.buildBookMarkdown(book, sheet))
			sb.WriteString("\n---\n\n")
		}
	case ShareFormatJSON:
		type exportEntry struct {
			Book  *domain.Book         `json:"book"`
			Sheet *domain.ReadingSheet `json:"reading_sheet,omitempty"`
		}
		var entries []exportEntry
		for _, book := range books {
			sheet, _ := s.sheetRepo.GetSheetByBookID(ctx, book.ID)
			entries = append(entries, exportEntry{Book: book, Sheet: sheet})
		}
		data, _ := json.MarshalIndent(entries, "", "  ")
		sb.Write(data)
	default:
		sb.WriteString(fmt.Sprintf("=== MA BIBLIOTHÈQUE ORUS — %s ===\n\n", time.Now().Format("02/01/2006")))
		for i, book := range books {
			sheet, _ := s.sheetRepo.GetSheetByBookID(ctx, book.ID)
			sb.WriteString(fmt.Sprintf("[%d] ", i+1))
			sb.WriteString(s.buildBookText(book, sheet))
			sb.WriteString("\n\n")
		}
	}

	ext := string(format)
	if format == ShareFormatText {
		ext = "txt"
	}

	fileName := fmt.Sprintf("orus_bibliotheque_%s.%s", time.Now().Format("20060102"), ext)
	filePath := filepath.Join(outputDir, fileName)

	if err := os.WriteFile(filePath, []byte(sb.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write library export: %w", err)
	}
	return filePath, nil
}

// ExportReadingSheet exporte uniquement la fiche de lecture d'un livre
func (s *SharingService) ExportReadingSheet(ctx context.Context, sheetID string, format ShareFormat, outputDir string) (string, error) {
	sheet, err := s.sheetRepo.GetSheetByID(ctx, sheetID)
	if err != nil {
		return "", fmt.Errorf("reading sheet not found: %w", err)
	}

	var content string
	var ext string

	switch format {
	case ShareFormatJSON:
		data, _ := json.MarshalIndent(sheet, "", "  ")
		content = string(data)
		ext = "json"
	case ShareFormatMarkdown:
		content = s.buildSheetMarkdown(sheet)
		ext = "md"
	default:
		content = s.buildSheetText(sheet)
		ext = "txt"
	}

	fileName := fmt.Sprintf("fiche_%s_%s.%s",
		sanitizeFileName(sheet.BookTitle),
		time.Now().Format("20060102"),
		ext,
	)
	filePath := filepath.Join(outputDir, fileName)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write sheet export: %w", err)
	}
	return filePath, nil
}

// --- Builders ---

func (s *SharingService) buildBookJSON(book *domain.Book, sheet *domain.ReadingSheet) (string, error) {
	payload := map[string]any{
		"book":          book,
		"reading_sheet": sheet,
		"exported_at":   time.Now(),
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	return string(data), err
}

func (s *SharingService) buildBookMarkdown(book *domain.Book, sheet *domain.ReadingSheet) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 📖 %s\n\n", book.Title))
	sb.WriteString(fmt.Sprintf("**Auteur :** %s  \n", orUnknown(book.Author)))
	sb.WriteString(fmt.Sprintf("**Format :** %s  \n", book.Format))
	sb.WriteString(fmt.Sprintf("**Pages :** %d  \n", book.TotalPages))
	sb.WriteString(fmt.Sprintf("**Ajouté le :** %s  \n\n", book.AddedAt.Format("02 Jan 2006")))

	if sheet != nil {
		sb.WriteString(s.buildSheetMarkdown(sheet))
	}
	return sb.String()
}

func (s *SharingService) buildSheetMarkdown(sheet *domain.ReadingSheet) string {
	var sb strings.Builder
	sb.WriteString("### 🗒️ Fiche de lecture\n\n")

	if sheet.Rating > 0 {
		sb.WriteString(fmt.Sprintf("**Note :** %s (%d/5)  \n\n", sheet.StarString(), sheet.Rating))
	}
	if len(sheet.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("**Tags :** %s  \n\n", strings.Join(sheet.Tags, " • ")))
	}
	if sheet.Summary != "" {
		sb.WriteString("**Résumé personnel :**\n\n")
		sb.WriteString(sheet.Summary + "\n\n")
	}
	if len(sheet.Quotes) > 0 {
		sb.WriteString("**Citations favorites :**\n\n")
		for _, q := range sheet.Quotes {
			sb.WriteString(fmt.Sprintf("> %s\n\n", q))
		}
	}
	return sb.String()
}

func (s *SharingService) buildBookText(book *domain.Book, sheet *domain.ReadingSheet) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("TITRE   : %s\n", book.Title))
	sb.WriteString(fmt.Sprintf("AUTEUR  : %s\n", orUnknown(book.Author)))
	sb.WriteString(fmt.Sprintf("FORMAT  : %s | PAGES : %d\n", book.Format, book.TotalPages))
	if sheet != nil {
		sb.WriteString(s.buildSheetText(sheet))
	}
	return sb.String()
}

func (s *SharingService) buildSheetText(sheet *domain.ReadingSheet) string {
	var sb strings.Builder
	sb.WriteString("--- FICHE DE LECTURE ---\n")
	if sheet.Rating > 0 {
		sb.WriteString(fmt.Sprintf("Note    : %s\n", sheet.StarString()))
	}
	if sheet.Summary != "" {
		sb.WriteString(fmt.Sprintf("Résumé  : %s\n", sheet.Summary))
	}
	for i, q := range sheet.Quotes {
		sb.WriteString(fmt.Sprintf("Citation %d : « %s »\n", i+1, q))
	}
	return sb.String()
}

func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer(" ", "_", "/", "-", "\\", "-", ":", "", "?", "", "*", "")
	result := replacer.Replace(name)
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}

func orUnknown(s string) string {
	if s == "" {
		return "Inconnu"
	}
	return s
}
