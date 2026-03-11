package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/port"
)

type ReminderCallback func(reminder *domain.Reminder)

type ReminderService struct {
	repo     port.ReminderRepository
	notifier port.Notifier
	onRing   ReminderCallback
	stop     chan struct{}
}

func NewReminderService(repo port.ReminderRepository, notifier port.Notifier) *ReminderService {
	return &ReminderService{repo: repo, notifier: notifier, stop: make(chan struct{})}
}

func (s *ReminderService) SetCallback(cb ReminderCallback) { s.onRing = cb }

func (s *ReminderService) AddReminder(ctx context.Context, bookID, bookTitle, label string, hour, minute int, freq domain.ReminderFrequency) (*domain.Reminder, error) {
	r, err := domain.NewReminder(bookID, bookTitle, label, hour, minute, freq)
	if err != nil {
		return nil, fmt.Errorf("invalid reminder: %w", err)
	}
	if err := s.repo.SaveReminder(ctx, r); err != nil {
		return nil, fmt.Errorf("failed to save reminder: %w", err)
	}
	return r, nil
}

func (s *ReminderService) ListReminders(ctx context.Context) ([]*domain.Reminder, error) {
	return s.repo.ListAllReminders(ctx)
}

func (s *ReminderService) ToggleReminder(ctx context.Context, id string) error {
	r, err := s.repo.GetReminderByID(ctx, id)
	if err != nil {
		return err
	}
	r.Enabled = !r.Enabled
	if r.Enabled {
		r.NextRing = r.ComputeNextRing(time.Now())
	}
	return s.repo.UpdateReminder(ctx, r)
}

// DismissReminder acquitte un rappel depuis la bannière UI.
// "once" → désactivé définitivement. Autres → NextRing avancé.
func (s *ReminderService) DismissReminder(ctx context.Context, id string) error {
	r, err := s.repo.GetReminderByID(ctx, id)
	if err != nil {
		log.Printf("[ReminderService] DismissReminder: %s introuvable : %v", id, err)
		return nil
	}
	r.Advance(time.Now())
	if err := s.repo.UpdateReminder(ctx, r); err != nil {
		return fmt.Errorf("failed to persist dismiss: %w", err)
	}
	log.Printf("[ReminderService] Acquitté %q — enabled=%v, next=%s",
		r.Label, r.Enabled, r.NextRing.Format("02/01 15:04"))
	return nil
}

func (s *ReminderService) DeleteReminder(ctx context.Context, id string) error {
	return s.repo.DeleteReminder(ctx, id)
}

func (s *ReminderService) StartScheduler() {
	log.Println("[ReminderService] Planificateur démarré")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			log.Println("[ReminderService] Planificateur arrêté")
			return
		case now := <-ticker.C:
			s.checkDueReminders(now)
		}
	}
}

func (s *ReminderService) Stop() { close(s.stop) }

func (s *ReminderService) checkDueReminders(now time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reminders, err := s.repo.ListEnabledReminders(ctx)
	if err != nil {
		log.Printf("[ReminderService] Erreur : %v", err)
		return
	}
	for _, r := range reminders {
		if !r.IsDue(now) {
			continue
		}
		if s.notifier != nil {
			msg := r.Label
			if r.BookTitle != "" {
				msg = fmt.Sprintf("%s — %s", r.Label, r.BookTitle)
			}
			_ = s.notifier.Notify("📖 Orus — Rappel de lecture", msg)
		}
		if s.onRing != nil {
			s.onRing(r)
		}
		r.Advance(now)
		if err := s.repo.UpdateReminder(ctx, r); err != nil {
			log.Printf("[ReminderService] Update échoué : %v", err)
		}
	}
}
