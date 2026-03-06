package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidRating        = errors.New("rating must be between 0 and 5")
	ErrReadingSheetNotFound = errors.New("reading sheet not found")
)

// ReadingSheet représente la fiche de lecture d'un livre (résumé personnel, citations, note)
type ReadingSheet struct {
	ID        string    `json:"id"`
	BookID    string    `json:"book_id"`
	BookTitle string    `json:"book_title"` // dénormalisé pour l'affichage
	Summary   string    `json:"summary"`    // résumé personnel du lecteur
	Quotes    []string  `json:"quotes"`     // citations favorites
	Rating    int       `json:"rating"`     // note sur 5 (0 = non noté)
	Tags      []string  `json:"tags"`       // ex: "roman", "histoire", "incontournable"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewReadingSheet crée une fiche de lecture valide
func NewReadingSheet(bookID, bookTitle, summary string, rating int, quotes, tags []string) (*ReadingSheet, error) {
	if bookID == "" {
		return nil, ErrInvalidBookId
	}
	if rating < 0 || rating > 5 {
		return nil, ErrInvalidRating
	}

	// Nettoyage des citations et tags vides
	cleanQuotes := filterEmpty(quotes)
	cleanTags := filterEmpty(tags)

	now := time.Now()
	return &ReadingSheet{
		ID:        uuid.New().String(),
		BookID:    bookID,
		BookTitle: bookTitle,
		Summary:   strings.TrimSpace(summary),
		Quotes:    cleanQuotes,
		Rating:    rating,
		Tags:      cleanTags,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// AddQuote ajoute une citation à la fiche
func (rs *ReadingSheet) AddQuote(quote string) {
	q := strings.TrimSpace(quote)
	if q != "" {
		rs.Quotes = append(rs.Quotes, q)
		rs.UpdatedAt = time.Now()
	}
}

// RemoveQuote supprime une citation par son index
func (rs *ReadingSheet) RemoveQuote(index int) {
	if index < 0 || index >= len(rs.Quotes) {
		return
	}
	rs.Quotes = append(rs.Quotes[:index], rs.Quotes[index+1:]...)
	rs.UpdatedAt = time.Now()
}

// UpdateSummary met à jour le résumé
func (rs *ReadingSheet) UpdateSummary(summary string) {
	rs.Summary = strings.TrimSpace(summary)
	rs.UpdatedAt = time.Now()
}

// UpdateRating met à jour la note (valide entre 0 et 5)
func (rs *ReadingSheet) UpdateRating(rating int) error {
	if rating < 0 || rating > 5 {
		return ErrInvalidRating
	}
	rs.Rating = rating
	rs.UpdatedAt = time.Now()
	return nil
}

// StarString retourne la représentation étoilée de la note (ex: "★★★☆☆")
func (rs *ReadingSheet) StarString() string {
	const filled = "★"
	const empty = "☆"
	result := ""
	for i := 1; i <= 5; i++ {
		if i <= rs.Rating {
			result += filled
		} else {
			result += empty
		}
	}
	return result
}

func filterEmpty(items []string) []string {
	var result []string
	for _, s := range items {
		if t := strings.TrimSpace(s); t != "" {
			result = append(result, t)
		}
	}
	return result
}
