package port

import (
	"context"

	"github.com/MiltonJ23/Orus/internal/domain"
)

// ReadingSheetRepository defines the contract for reading sheet persistence.
type ReadingSheetRepository interface {
	SaveSheet(ctx context.Context, sheet *domain.ReadingSheet) error
	GetSheetByID(ctx context.Context, id string) (*domain.ReadingSheet, error)
	GetSheetByBookID(ctx context.Context, bookID string) (*domain.ReadingSheet, error)
	ListAllSheets(ctx context.Context) ([]*domain.ReadingSheet, error)
	UpdateSheet(ctx context.Context, sheet *domain.ReadingSheet) error
	DeleteSheet(ctx context.Context, id string) error
}

// ReminderRepository defines the contract for reminder persistence.
type ReminderRepository interface {
	SaveReminder(ctx context.Context, reminder *domain.Reminder) error
	GetReminderByID(ctx context.Context, id string) (*domain.Reminder, error)
	ListAllReminders(ctx context.Context) ([]*domain.Reminder, error)
	UpdateReminder(ctx context.Context, reminder *domain.Reminder) error
	DeleteReminder(ctx context.Context, id string) error
	ListEnabledReminders(ctx context.Context) ([]*domain.Reminder, error)
}

// Notifier defines the contract for sending system notifications.
type Notifier interface {
	Notify(title, message string) error
}
