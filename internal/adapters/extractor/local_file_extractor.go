package extractor

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/port"
	"github.com/kapmahc/epub"
	"github.com/ledongthuc/pdf"
)

var (
	ErrUnsupportedFileFormat = errors.New("unsupported file format")
	ErrCorruptFile           = errors.New("file is corrupted or unreadable")
)

// linesPerChunk : nombre de lignes par "page" affichée dans le lecteur
const linesPerChunk = 35

var _ port.ContentReader = (*LocalFileExtractor)(nil)

// LocalFileExtractor extracts metadata and text content from PDF and EPUB files.
type LocalFileExtractor struct{}

// NewLocalFileExtractor creates a new LocalFileExtractor.
func NewLocalFileExtractor() *LocalFileExtractor {
	return &LocalFileExtractor{}
}

// ExtractInfo returns metadata (title, author, pages, format) for the given file.
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

// ReadBookText extrait le texte complet du fichier et le découpe en pages lisibles.
func (l *LocalFileExtractor) ReadBookText(ctx context.Context, filePath string) ([]string, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".pdf":
		return l.readPDFText(ctx, filePath)
	case ".epub":
		return l.readEPUBText(filePath)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedFileFormat, ext)
	}
}

func (l *LocalFileExtractor) readPDFText(ctx context.Context, filePath string) ([]string, error) {
	file, reader, err := pdf.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("impossible d'ouvrir le PDF : %w", err)
	}
	defer file.Close()

	var allLines []string

	for i := 1; i <= reader.NumPage(); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		// Découper le texte brut en lignes
		lines := strings.Split(text, "\n")
		for _, l := range lines {
			trimmed := strings.TrimSpace(l)
			if trimmed != "" {
				allLines = append(allLines, trimmed)
			}
		}
		// Séparateur de page
		allLines = append(allLines, fmt.Sprintf("── Page %d ──", i))
	}

	if len(allLines) == 0 {
		return []string{"Aucun texte extractible dans ce PDF.\n\nCe document est peut-être scanné ou protégé."}, nil
	}

	return chunkLines(allLines, linesPerChunk), nil
}

func (l *LocalFileExtractor) readEPUBText(filePath string) ([]string, error) {
	book, err := epub.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptFile, err)
	}
	defer book.Close()

	var allLines []string

	for i, item := range book.Opf.Spine.Items {
		var content string
		// Récupérer le contenu HTML de l'élément du spine
		for _, manifest := range book.Opf.Manifest {
			if manifest.ID == item.IDref {
				raw, err := book.Open(manifest.Href)
				if err != nil {
					continue
				}
				buf := new(strings.Builder)
				tmp := make([]byte, 4096)
				for {
					n, e := raw.Read(tmp)
					if n > 0 {
						buf.Write(tmp[:n])
					}
					if e != nil {
						break
					}
				}
				raw.Close()
				content = stripHTML(buf.String())
				break
			}
		}

		if content == "" {
			continue
		}

		lines := strings.Split(content, "\n")
		allLines = append(allLines, fmt.Sprintf("═══ Chapitre %d ═══", i+1))
		for _, ln := range lines {
			t := strings.TrimSpace(ln)
			if t != "" {
				allLines = append(allLines, t)
			}
		}
	}

	if len(allLines) == 0 {
		return []string{"Aucun texte trouvé dans ce fichier EPUB."}, nil
	}

	return chunkLines(allLines, linesPerChunk), nil
}

// stripHTML supprime les balises HTML basiques
func stripHTML(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
			result.WriteRune(' ')
		case !inTag:
			result.WriteRune(r)
		}
	}
	// Nettoyage des entités HTML communes
	out := result.String()
	out = strings.ReplaceAll(out, "&nbsp;", " ")
	out = strings.ReplaceAll(out, "&amp;", "&")
	out = strings.ReplaceAll(out, "&lt;", "<")
	out = strings.ReplaceAll(out, "&gt;", ">")
	out = strings.ReplaceAll(out, "&quot;", "\"")
	out = strings.ReplaceAll(out, "&#39;", "'")
	return out
}

// chunkLines découpe un slice de lignes en chunks de n lignes
func chunkLines(lines []string, n int) []string {
	var chunks []string
	for i := 0; i < len(lines); i += n {
		end := i + n
		if end > len(lines) {
			end = len(lines)
		}
		chunks = append(chunks, strings.Join(lines[i:end], "\n"))
	}
	return chunks
}

// --- Extraction de métadonnées (inchangé) ---

func (l *LocalFileExtractor) ExtractPDF(filePath string) (*domain.BookMetadata, error) {
	file, reader, pdfOpeningError := pdf.Open(filePath)
	if pdfOpeningError != nil {
		return nil, fmt.Errorf("an error occured trying to open the pdf file")
	}
	defer file.Close()

	totalPages := reader.NumPage()
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
	book, epubOpeningerr := epub.Open(filePath)
	if epubOpeningerr != nil {
		return nil, fmt.Errorf("an error occured trying to open file,%w: %v", ErrCorruptFile, epubOpeningerr)
	}
	defer book.Close()

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
