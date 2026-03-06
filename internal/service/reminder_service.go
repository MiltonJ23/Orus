package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/port"
)

// ReminderCallback est appelé quand un rappel doit sonner (hook UI)
type ReminderCallback func(reminder *domain.Reminder)

type ReminderService struct {
	repo     port.ReminderRepository
	notifier port.Notifier
	onRing   ReminderCallback
	stop     chan struct{}
}

func NewReminderService(repo port.ReminderRepository, notifier port.Notifier) *ReminderService {
	return &ReminderService{
		repo:     repo,
		notifier: notifier,
		stop:     make(chan struct{}),
	}
}

// SetCallback enregistre la fonction UI à appeler quand un rappel sonne
func (s *ReminderService) SetCallback(cb ReminderCallback) {
	s.onRing = cb
}

// AddReminder crée et persiste un nouveau rappel
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

// ListReminders retourne tous les rappels
func (s *ReminderService) ListReminders(ctx context.Context) ([]*domain.Reminder, error) {
	return s.repo.ListAllReminders(ctx)
}

// ToggleReminder active ou désactive un rappel
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

// DeleteReminder supprime un rappel
func (s *ReminderService) DeleteReminder(ctx context.Context, id string) error {
	return s.repo.DeleteReminder(ctx, id)
}

// StartScheduler lance la boucle de vérification des rappels en arrière-plan.
// À appeler dans une goroutine : go reminderSvc.StartScheduler()
func (s *ReminderService) StartScheduler() {
	log.Println("[ReminderService] Planificateur de rappels démarré")
	ticker := time.NewTicker(30 * time.Second) // vérification toutes les 30s
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

// Stop arrête le planificateur proprement
func (s *ReminderService) Stop() {
	close(s.stop)
}

func (s *ReminderService) checkDueReminders(now time.Time) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reminders, err := s.repo.ListEnabledReminders(ctx)
	if err != nil {
		log.Printf("[ReminderService] Erreur lecture rappels : %v", err)
		return
	}

	for _, r := range reminders {
		if !r.IsDue(now) {
			continue
		}

		// Notification système
		if s.notifier != nil {
			msg := r.Label
			if r.BookTitle != "" {
				msg = fmt.Sprintf("%s — %s", r.Label, r.BookTitle)
			}
			if err := s.notifier.Notify("📖 Orus — Rappel de lecture", msg); err != nil {
				log.Printf("[ReminderService] Notification échouée : %v", err)
			}
		}

		// Hook UI (ex: afficher une bannière dans l'app)
		if s.onRing != nil {
			s.onRing(r)
		}

		// Avancer la prochaine occurrence et persister
		r.Advance(now)
		if err := s.repo.UpdateReminder(ctx, r); err != nil {
			log.Printf("[ReminderService] Mise à jour rappel échouée : %v", err)
		}

		log.Printf("[ReminderService] Rappel déclenché : %s", r.Label)
	}
}
