package port

import (
	"context"

	"github.com/MiltonJ23/Orus/internal/domain"
)

// ReadingSheetRepository définit le contrat de persistance pour les fiches de lecture
type ReadingSheetRepository interface {
	SaveSheet(ctx context.Context, sheet *domain.ReadingSheet) error
	GetSheetByID(ctx context.Context, id string) (*domain.ReadingSheet, error)
	GetSheetByBookID(ctx context.Context, bookID string) (*domain.ReadingSheet, error)
	ListAllSheets(ctx context.Context) ([]*domain.ReadingSheet, error)
	UpdateSheet(ctx context.Context, sheet *domain.ReadingSheet) error
	DeleteSheet(ctx context.Context, id string) error
}

// ReminderRepository définit le contrat de persistance pour les rappels
type ReminderRepository interface {
	SaveReminder(ctx context.Context, reminder *domain.Reminder) error
	GetReminderByID(ctx context.Context, id string) (*domain.Reminder, error)
	ListAllReminders(ctx context.Context) ([]*domain.Reminder, error)
	UpdateReminder(ctx context.Context, reminder *domain.Reminder) error
	DeleteReminder(ctx context.Context, id string) error
	ListEnabledReminders(ctx context.Context) ([]*domain.Reminder, error)
}

// Notifier est le port de sortie pour envoyer des notifications système
type Notifier interface {
	Notify(title, message string) error
}
