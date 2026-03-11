package notifier_test

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/MiltonJ23/Orus/internal/adapters/notifier"
)

func TestLogNotifier_Notify(t *testing.T) {
	t.Run("Basic Notification", func(t *testing.T) {
		// Capture log output
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr) // Restore default output

		notif := notifier.NewLogNotifier()
		err := notif.Notify("Test Title", "Test Message")

		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}

		// Verify log output contains the notification
		output := buf.String()
		if !strings.Contains(output, "Test Title") {
			t.Errorf("expected log to contain 'Test Title', got: %s", output)
		}
		if !strings.Contains(output, "Test Message") {
			t.Errorf("expected log to contain 'Test Message', got: %s", output)
		}
		if !strings.Contains(output, "[NOTIFICATION]") {
			t.Errorf("expected log to contain '[NOTIFICATION]', got: %s", output)
		}
	})

	t.Run("Empty Title and Message", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		notif := notifier.NewLogNotifier()
		err := notif.Notify("", "")

		if err != nil {
			t.Errorf("expected no error even with empty strings, got: %v", err)
		}

		// Should still log something
		output := buf.String()
		if !strings.Contains(output, "[NOTIFICATION]") {
			t.Errorf("expected log to contain '[NOTIFICATION]', got: %s", output)
		}
	})

	t.Run("Special Characters in Notification", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		notif := notifier.NewLogNotifier()
		title := "Special: Test & Demo"
		message := "Message with \"quotes\" and 'apostrophes' and \nnewlines"

		err := notif.Notify(title, message)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Special: Test & Demo") {
			t.Errorf("expected log to contain title with special chars, got: %s", output)
		}
	})

	t.Run("Long Title and Message", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		notif := notifier.NewLogNotifier()
		longTitle := strings.Repeat("A very long title ", 20)
		longMessage := strings.Repeat("A very long message with lots of text ", 50)

		err := notif.Notify(longTitle, longMessage)
		if err != nil {
			t.Errorf("expected no error with long strings, got: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "[NOTIFICATION]") {
			t.Errorf("expected log to contain notification marker")
		}
	})

	t.Run("Multiple Sequential Notifications", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		notif := notifier.NewLogNotifier()

		for i := 1; i <= 5; i++ {
			err := notif.Notify("Title", "Message")
			if err != nil {
				t.Errorf("notification %d failed: %v", i, err)
			}
		}

		output := buf.String()
		count := strings.Count(output, "[NOTIFICATION]")
		if count != 5 {
			t.Errorf("expected 5 notifications in log, got %d", count)
		}
	})

	t.Run("Unicode Characters in Notification", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		notif := notifier.NewLogNotifier()
		title := "📖 Time to Read!"
		message := "Bonjour! It's time to read your favorite book 🎉"

		err := notif.Notify(title, message)
		if err != nil {
			t.Errorf("expected no error with unicode, got: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "📖") {
			t.Errorf("expected log to contain emoji, got: %s", output)
		}
	})
}

func TestNewLogNotifier(t *testing.T) {
	t.Run("Create New Notifier", func(t *testing.T) {
		notif := notifier.NewLogNotifier()
		if notif == nil {
			t.Error("expected non-nil notifier")
		}
	})

	t.Run("Multiple Instances", func(t *testing.T) {
		notif1 := notifier.NewLogNotifier()
		notif2 := notifier.NewLogNotifier()

		if notif1 == nil || notif2 == nil {
			t.Error("expected both notifiers to be non-nil")
		}

		// Both should work independently
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stderr)

		notif1.Notify("From 1", "Message 1")
		notif2.Notify("From 2", "Message 2")

		output := buf.String()
		if !strings.Contains(output, "From 1") || !strings.Contains(output, "From 2") {
			t.Error("expected both notifications in output")
		}
	})
}