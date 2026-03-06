package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/port"
)

var _ port.ReminderRepository = (*Storage)(nil)

func (s *Storage) SaveReminder(ctx context.Context, r *domain.Reminder) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `INSERT INTO reminders (id, book_id, book_title, label, hour, minute, frequency, enabled, next_ring, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query,
		r.ID, r.BookID, r.BookTitle, r.Label,
		r.Hour, r.Minute, string(r.Frequency),
		r.Enabled, r.NextRing, r.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save reminder: %w", err)
	}
	return nil
}

func (s *Storage) GetReminderByID(ctx context.Context, id string) (*domain.Reminder, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, book_id, book_title, label, hour, minute, frequency, enabled, next_ring, created_at FROM reminders WHERE id=?`
	row := s.db.QueryRowContext(ctx, query, id)
	r, err := scanReminder(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReminderNotFound
		}
		return nil, fmt.Errorf("failed to get reminder: %w", err)
	}
	return r, nil
}

func (s *Storage) ListAllReminders(ctx context.Context) ([]*domain.Reminder, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.queryReminders(ctx, `SELECT id, book_id, book_title, label, hour, minute, frequency, enabled, next_ring, created_at FROM reminders ORDER BY hour, minute`)
}

func (s *Storage) ListEnabledReminders(ctx context.Context) ([]*domain.Reminder, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return s.queryReminders(ctx, `SELECT id, book_id, book_title, label, hour, minute, frequency, enabled, next_ring, created_at FROM reminders WHERE enabled=1 ORDER BY next_ring ASC`)
}

func (s *Storage) UpdateReminder(ctx context.Context, r *domain.Reminder) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `UPDATE reminders SET label=?, hour=?, minute=?, frequency=?, enabled=?, next_ring=? WHERE id=?`
	_, err := s.db.ExecContext(ctx, query, r.Label, r.Hour, r.Minute, string(r.Frequency), r.Enabled, r.NextRing, r.ID)
	if err != nil {
		return fmt.Errorf("failed to update reminder: %w", err)
	}
	return nil
}

func (s *Storage) DeleteReminder(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.db.ExecContext(ctx, `DELETE FROM reminders WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete reminder: %w", err)
	}
	return nil
}

// --- helpers ---

func (s *Storage) queryReminders(ctx context.Context, query string) ([]*domain.Reminder, error) {
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query reminders: %w", err)
	}
	defer rows.Close()

	var reminders []*domain.Reminder
	for rows.Next() {
		var r domain.Reminder
		var freqStr string
		err := rows.Scan(&r.ID, &r.BookID, &r.BookTitle, &r.Label, &r.Hour, &r.Minute, &freqStr, &r.Enabled, &r.NextRing, &r.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reminder: %w", err)
		}
		r.Frequency = domain.ReminderFrequency(freqStr)
		reminders = append(reminders, &r)
	}
	return reminders, rows.Err()
}

func scanReminder(row rowScanner) (*domain.Reminder, error) {
	var r domain.Reminder
	var freqStr string
	err := row.Scan(&r.ID, &r.BookID, &r.BookTitle, &r.Label, &r.Hour, &r.Minute, &freqStr, &r.Enabled, &r.NextRing, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	r.Frequency = domain.ReminderFrequency(freqStr)
	return &r, nil
}
