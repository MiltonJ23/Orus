package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrReminderNotFound    = errors.New("reminder not found")
	ErrInvalidReminderTime = errors.New("reminder time must be set in the future")
)

// ReminderFrequency définit la récurrence d'un rappel
type ReminderFrequency string

const (
	FrequencyDaily    ReminderFrequency = "daily"
	FrequencyWeekly   ReminderFrequency = "weekly"
	FrequencyWeekdays ReminderFrequency = "weekdays" // lundi–vendredi
	FrequencyOnce     ReminderFrequency = "once"
)

// Reminder représente un rappel de lecture planifié
type Reminder struct {
	ID        string            `json:"id"`
	BookID    string            `json:"book_id"`    // vide = rappel global de lecture
	BookTitle string            `json:"book_title"` // dénormalisé
	Label     string            `json:"label"`      // ex: "Lire 30 minutes"
	Hour      int               `json:"hour"`       // heure du rappel (0-23)
	Minute    int               `json:"minute"`     // minute (0-59)
	Frequency ReminderFrequency `json:"frequency"`
	Enabled   bool              `json:"enabled"`
	NextRing  time.Time         `json:"next_ring"` // prochaine occurrence calculée
	CreatedAt time.Time         `json:"created_at"`
}

// NewReminder crée un rappel valide
func NewReminder(bookID, bookTitle, label string, hour, minute int, freq ReminderFrequency) (*Reminder, error) {
	if hour < 0 || hour > 23 {
		return nil, ErrInvalidReminderTime
	}
	if minute < 0 || minute > 59 {
		return nil, ErrInvalidReminderTime
	}
	if label == "" {
		label = "📖 C'est l'heure de lire !"
	}

	r := &Reminder{
		ID:        uuid.New().String(),
		BookID:    bookID,
		BookTitle: bookTitle,
		Label:     label,
		Hour:      hour,
		Minute:    minute,
		Frequency: freq,
		Enabled:   true,
		CreatedAt: time.Now(),
	}
	r.NextRing = r.ComputeNextRing(time.Now())
	return r, nil
}

// ComputeNextRing calcule la prochaine occurrence du rappel à partir de 'from'
func (r *Reminder) ComputeNextRing(from time.Time) time.Time {
	// On construit l'heure du rappel pour aujourd'hui
	candidate := time.Date(from.Year(), from.Month(), from.Day(), r.Hour, r.Minute, 0, 0, from.Location())

	// Si l'heure est déjà passée aujourd'hui, on avance au lendemain
	if !candidate.After(from) {
		candidate = candidate.Add(24 * time.Hour)
	}

	// Pour FrequencyWeekdays, on saute le week-end
	if r.Frequency == FrequencyWeekdays {
		for candidate.Weekday() == time.Saturday || candidate.Weekday() == time.Sunday {
			candidate = candidate.Add(24 * time.Hour)
		}
	}

	return candidate
}

// IsDue retourne true si le rappel doit sonner maintenant (avec une tolérance d'1 minute)
func (r *Reminder) IsDue(now time.Time) bool {
	if !r.Enabled {
		return false
	}
	diff := now.Sub(r.NextRing)
	return diff >= 0 && diff < time.Minute
}

// Advance met à jour NextRing après que le rappel a sonné
func (r *Reminder) Advance(from time.Time) {
	switch r.Frequency {
	case FrequencyOnce:
		r.Enabled = false
	case FrequencyDaily, FrequencyWeekdays:
		r.NextRing = r.ComputeNextRing(from)
	case FrequencyWeekly:
		r.NextRing = r.NextRing.Add(7 * 24 * time.Hour)
	}
}

// FrequencyLabel retourne le label lisible de la fréquence
func (r *Reminder) FrequencyLabel() string {
	switch r.Frequency {
	case FrequencyDaily:
		return "Tous les jours"
	case FrequencyWeekly:
		return "Chaque semaine"
	case FrequencyWeekdays:
		return "Jours ouvrés (Lun-Ven)"
	case FrequencyOnce:
		return "Une seule fois"
	default:
		return string(r.Frequency)
	}
}
