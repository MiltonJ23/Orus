package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/service"
)

// --- MOCKS FOR REMINDER ---

type mockReminderRepo struct {
	reminders  map[string]*domain.Reminder
	failSave   bool
	failGet    bool
	failList   bool
	failUpdate bool
	failDelete bool
}

func newMockReminderRepo() *mockReminderRepo {
	return &mockReminderRepo{reminders: make(map[string]*domain.Reminder)}
}

func (m *mockReminderRepo) SaveReminder(_ context.Context, r *domain.Reminder) error {
	if m.failSave {
		return errors.New("save error")
	}
	m.reminders[r.ID] = r
	return nil
}

func (m *mockReminderRepo) GetReminderByID(_ context.Context, id string) (*domain.Reminder, error) {
	if m.failGet {
		return nil, errors.New("get error")
	}
	r, ok := m.reminders[id]
	if !ok {
		return nil, domain.ErrReminderNotFound
	}
	return r, nil
}

func (m *mockReminderRepo) ListAllReminders(_ context.Context) ([]*domain.Reminder, error) {
	if m.failList {
		return nil, errors.New("list error")
	}
	var out []*domain.Reminder
	for _, r := range m.reminders {
		out = append(out, r)
	}
	return out, nil
}

func (m *mockReminderRepo) UpdateReminder(_ context.Context, r *domain.Reminder) error {
	if m.failUpdate {
		return errors.New("update error")
	}
	m.reminders[r.ID] = r
	return nil
}

func (m *mockReminderRepo) DeleteReminder(_ context.Context, id string) error {
	if m.failDelete {
		return errors.New("delete error")
	}
	delete(m.reminders, id)
	return nil
}

func (m *mockReminderRepo) ListEnabledReminders(_ context.Context) ([]*domain.Reminder, error) {
	if m.failList {
		return nil, errors.New("list error")
	}
	var out []*domain.Reminder
	for _, r := range m.reminders {
		if r.Enabled {
			out = append(out, r)
		}
	}
	return out, nil
}

type mockNotifier struct {
	lastTitle   string
	lastMessage string
	failNotify  bool
}

func (m *mockNotifier) Notify(title, message string) error {
	if m.failNotify {
		return errors.New("notify error")
	}
	m.lastTitle = title
	m.lastMessage = message
	return nil
}

// --- TESTS ---

func TestReminderService_AddReminder(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		repo := newMockReminderRepo()
		notifier := &mockNotifier{}
		svc := service.NewReminderService(repo, notifier)

		r, err := svc.AddReminder(ctx, "book-1", "Test Book", "Read 30 min", 18, 30, domain.FrequencyDaily)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if r.Label != "Read 30 min" {
			t.Errorf("expected label 'Read 30 min', got %q", r.Label)
		}
		if !r.Enabled {
			t.Error("expected reminder to be enabled")
		}
		if _, ok := repo.reminders[r.ID]; !ok {
			t.Error("expected reminder to be saved in repo")
		}
	})

	t.Run("InvalidTime", func(t *testing.T) {
		repo := newMockReminderRepo()
		svc := service.NewReminderService(repo, nil)

		_, err := svc.AddReminder(ctx, "book-1", "Test", "Read", 25, 0, domain.FrequencyDaily)
		if err == nil {
			t.Fatal("expected error for invalid hour, got nil")
		}
	})

	t.Run("SaveError", func(t *testing.T) {
		repo := newMockReminderRepo()
		repo.failSave = true
		svc := service.NewReminderService(repo, nil)

		_, err := svc.AddReminder(ctx, "book-1", "Test", "Read", 18, 30, domain.FrequencyDaily)
		if err == nil {
			t.Fatal("expected save error, got nil")
		}
	})
}

func TestReminderService_ListReminders(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		repo := newMockReminderRepo()
		svc := service.NewReminderService(repo, nil)

		_, _ = svc.AddReminder(ctx, "b1", "Book1", "Read", 8, 0, domain.FrequencyDaily)
		_, _ = svc.AddReminder(ctx, "b2", "Book2", "Study", 9, 0, domain.FrequencyWeekly)

		list, err := svc.ListReminders(ctx)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 reminders, got %d", len(list))
		}
	})

	t.Run("Error", func(t *testing.T) {
		repo := newMockReminderRepo()
		repo.failList = true
		svc := service.NewReminderService(repo, nil)

		_, err := svc.ListReminders(ctx)
		if err == nil {
			t.Fatal("expected list error, got nil")
		}
	})
}

func TestReminderService_ToggleReminder(t *testing.T) {
	ctx := context.Background()

	t.Run("Disable", func(t *testing.T) {
		repo := newMockReminderRepo()
		svc := service.NewReminderService(repo, nil)

		r, _ := svc.AddReminder(ctx, "b1", "Book", "Read", 10, 0, domain.FrequencyDaily)
		if !r.Enabled {
			t.Fatal("reminder should be enabled initially")
		}

		err := svc.ToggleReminder(ctx, r.ID)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		updated := repo.reminders[r.ID]
		if updated.Enabled {
			t.Error("expected reminder to be disabled after toggle")
		}
	})

	t.Run("Enable", func(t *testing.T) {
		repo := newMockReminderRepo()
		svc := service.NewReminderService(repo, nil)

		r, _ := svc.AddReminder(ctx, "b1", "Book", "Read", 10, 0, domain.FrequencyDaily)
		_ = svc.ToggleReminder(ctx, r.ID)    // disable
		err := svc.ToggleReminder(ctx, r.ID) // re-enable
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		updated := repo.reminders[r.ID]
		if !updated.Enabled {
			t.Error("expected reminder to be enabled after double toggle")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		repo := newMockReminderRepo()
		svc := service.NewReminderService(repo, nil)

		err := svc.ToggleReminder(ctx, "nonexistent")
		if err == nil {
			t.Fatal("expected error for nonexistent reminder")
		}
	})
}

func TestReminderService_DismissReminder(t *testing.T) {
	ctx := context.Background()

	t.Run("DismissOnce", func(t *testing.T) {
		repo := newMockReminderRepo()
		svc := service.NewReminderService(repo, nil)

		r, _ := svc.AddReminder(ctx, "b1", "Book", "Read", 10, 0, domain.FrequencyOnce)

		err := svc.DismissReminder(ctx, r.ID)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		updated := repo.reminders[r.ID]
		if updated.Enabled {
			t.Error("once reminder should be disabled after dismiss")
		}
	})

	t.Run("DismissDaily", func(t *testing.T) {
		repo := newMockReminderRepo()
		svc := service.NewReminderService(repo, nil)

		r, _ := svc.AddReminder(ctx, "b1", "Book", "Read", 10, 0, domain.FrequencyDaily)

		err := svc.DismissReminder(ctx, r.ID)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		updated := repo.reminders[r.ID]
		if !updated.Enabled {
			t.Error("daily reminder should still be enabled after dismiss")
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		repo := newMockReminderRepo()
		repo.failGet = true
		svc := service.NewReminderService(repo, nil)

		err := svc.DismissReminder(ctx, "nonexistent")
		if err != nil {
			t.Fatal("DismissReminder should return nil for not found")
		}
	})
}

func TestReminderService_DeleteReminder(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		repo := newMockReminderRepo()
		svc := service.NewReminderService(repo, nil)

		r, _ := svc.AddReminder(ctx, "b1", "Book", "Read", 10, 0, domain.FrequencyDaily)

		err := svc.DeleteReminder(ctx, r.ID)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if _, ok := repo.reminders[r.ID]; ok {
			t.Error("expected reminder to be deleted from repo")
		}
	})

	t.Run("Error", func(t *testing.T) {
		repo := newMockReminderRepo()
		repo.failDelete = true
		svc := service.NewReminderService(repo, nil)

		err := svc.DeleteReminder(ctx, "anything")
		if err == nil {
			t.Fatal("expected delete error, got nil")
		}
	})
}

func TestReminderService_StartSchedulerAndStop(t *testing.T) {
	repo := newMockReminderRepo()
	notifier := &mockNotifier{}
	svc := service.NewReminderService(repo, notifier)

	done := make(chan struct{})
	go func() {
		svc.StartScheduler()
		close(done)
	}()

	// Let the scheduler run briefly
	time.Sleep(100 * time.Millisecond)
	svc.Stop()

	select {
	case <-done:
		// Scheduler stopped gracefully
	case <-time.After(5 * time.Second):
		t.Fatal("scheduler did not stop within timeout")
	}
}

func TestReminderService_SetCallback(t *testing.T) {
	repo := newMockReminderRepo()
	svc := service.NewReminderService(repo, nil)

	called := false
	svc.SetCallback(func(r *domain.Reminder) {
		called = true
	})

	// SetCallback should not panic — just verify it doesn't
	if called {
		t.Error("callback should not have been called yet")
	}
}
