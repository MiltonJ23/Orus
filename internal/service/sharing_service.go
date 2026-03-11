package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/port"
)

type ShareFormat string

const (
	ShareFormatJSON     ShareFormat = "json"
	ShareFormatMarkdown ShareFormat = "md"
	ShareFormatText     ShareFormat = "txt"
)

type SharingService struct {
	bookRepo  port.BookRepository
	sheetRepo port.ReadingSheetRepository
}

func NewSharingService(bookRepo port.BookRepository, sheetRepo port.ReadingSheetRepository) *SharingService {
	return &SharingService{bookRepo: bookRepo, sheetRepo: sheetRepo}
}

// PickExportDirectory ouvre le sélecteur de dossier natif de l'OS.
func PickExportDirectory() string {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("osascript", "-e",
			`POSIX path of (choose folder with prompt "Choisir le dossier d'export Orus :")`)
	case "linux":
		if _, err := exec.LookPath("zenity"); err == nil {
			cmd = exec.Command("zenity", "--file-selection", "--directory", "--title=Dossier d'export Orus")
		} else if _, err := exec.LookPath("kdialog"); err == nil {
			cmd = exec.Command("kdialog", "--getexistingdirectory", ".", "--title", "Dossier d'export")
		}
	case "windows":
		cmd = exec.Command("powershell", "-Command",
			`(New-Object -ComObject Shell.Application).BrowseForFolder(0,"Dossier d'export",0).Self.Path`)
	}
	if cmd == nil {
		return "."
	}
	out, err := cmd.Output()
	if err != nil {
		return "."
	}
	dir := strings.TrimSpace(string(out))
	if dir == "" {
		return "."
	}
	return dir
}

func (s *SharingService) ExportLibrary(ctx context.Context, format ShareFormat, outputDir string) (string, error) {
	books, err := s.bookRepo.ListAll(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list books: %w", err)
	}
	var sb strings.Builder
	switch format {
	case ShareFormatMarkdown:
		sb.WriteString("# Ma bibliotheque Orus\n\n")
		sb.WriteString(fmt.Sprintf("*Exporte le %s — %d livre(s)*\n\n---\n\n", time.Now().Format("02 janvier 2006"), len(books)))
		for _, book := range books {
			sheet, _ := s.sheetRepo.GetSheetByBookID(ctx, book.ID)
			sb.WriteString(s.buildBookMarkdown(book, sheet))
			sb.WriteString("\n---\n\n")
		}
	case ShareFormatJSON:
		type entry struct {
			Book  *domain.Book         `json:"book"`
			Sheet *domain.ReadingSheet `json:"reading_sheet,omitempty"`
		}
		var entries []entry
		for _, book := range books {
			sheet, _ := s.sheetRepo.GetSheetByBookID(ctx, book.ID)
			entries = append(entries, entry{Book: book, Sheet: sheet})
		}
		data, _ := json.MarshalIndent(entries, "", "  ")
		sb.Write(data)
	default:
		sb.WriteString(fmt.Sprintf("=== BIBLIOTHEQUE ORUS — %s ===\n\n", time.Now().Format("02/01/2006")))
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
		return "", fmt.Errorf("failed to write: %w", err)
	}
	return filePath, nil
}

func (s *SharingService) ExportBookInfo(ctx context.Context, bookID string, format ShareFormat, outputDir string) (string, error) {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return "", fmt.Errorf("book not found: %w", err)
	}
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
		return "", err
	}
	filePath := filepath.Join(outputDir, fmt.Sprintf("orus_%s_%s.%s", sanitizeFileName(book.Title), time.Now().Format("20060102"), ext))
	return filePath, os.WriteFile(filePath, []byte(content), 0644)
}

func (s *SharingService) ExportReadingSheet(ctx context.Context, sheetID string, format ShareFormat, outputDir string) (string, error) {
	sheet, err := s.sheetRepo.GetSheetByID(ctx, sheetID)
	if err != nil {
		return "", fmt.Errorf("sheet not found: %w", err)
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
	filePath := filepath.Join(outputDir, fmt.Sprintf("fiche_%s_%s.%s", sanitizeFileName(sheet.BookTitle), time.Now().Format("20060102"), ext))
	return filePath, os.WriteFile(filePath, []byte(content), 0644)
}

func (s *SharingService) buildBookJSON(book *domain.Book, sheet *domain.ReadingSheet) (string, error) {
	data, err := json.MarshalIndent(map[string]any{"book": book, "reading_sheet": sheet, "exported_at": time.Now()}, "", "  ")
	return string(data), err
}

func (s *SharingService) buildBookMarkdown(book *domain.Book, sheet *domain.ReadingSheet) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %s\n\n**Auteur :** %s  \n**Format :** %s  \n**Pages :** %d  \n**Ajoute le :** %s  \n\n",
		book.Title, orUnknown(book.Author), book.Format, book.TotalPages, book.AddedAt.Format("02 Jan 2006")))
	if sheet != nil {
		sb.WriteString(s.buildSheetMarkdown(sheet))
	}
	return sb.String()
}

func (s *SharingService) buildSheetMarkdown(sheet *domain.ReadingSheet) string {
	var sb strings.Builder
	sb.WriteString("### Fiche de lecture\n\n")
	if sheet.Rating > 0 {
		sb.WriteString(fmt.Sprintf("**Note :** %s (%d/5)\n\n", sheet.StarString(), sheet.Rating))
	}
	if len(sheet.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("**Tags :** %s\n\n", strings.Join(sheet.Tags, " - ")))
	}
	if sheet.Summary != "" {
		sb.WriteString("**Resume :**\n\n" + sheet.Summary + "\n\n")
	}
	for _, q := range sheet.Quotes {
		sb.WriteString(fmt.Sprintf("> %s\n\n", q))
	}
	return sb.String()
}

func (s *SharingService) buildBookText(book *domain.Book, sheet *domain.ReadingSheet) string {
	out := fmt.Sprintf("TITRE: %s | AUTEUR: %s | FORMAT: %s | PAGES: %d\n", book.Title, orUnknown(book.Author), book.Format, book.TotalPages)
	if sheet != nil {
		out += s.buildSheetText(sheet)
	}
	return out
}

func (s *SharingService) buildSheetText(sheet *domain.ReadingSheet) string {
	var sb strings.Builder
	if sheet.Rating > 0 {
		sb.WriteString(fmt.Sprintf("Note: %s | ", sheet.StarString()))
	}
	if sheet.Summary != "" {
		sb.WriteString("Resume: " + sheet.Summary + "\n")
	}
	for i, q := range sheet.Quotes {
		sb.WriteString(fmt.Sprintf("Cit.%d: %s\n", i+1, q))
	}
	return sb.String()
}

func sanitizeFileName(name string) string {
	r := strings.NewReplacer(" ", "_", "/", "-", "\\", "-", ":", "", "?", "", "*", "")
	out := r.Replace(name)
	if len(out) > 50 {
		return out[:50]
	}
	return out
}

func orUnknown(s string) string {
	if s == "" {
		return "Inconnu"
	}
	return s
}
