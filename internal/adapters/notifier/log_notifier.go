package notifier

import (
	"fmt"
	"log"
	"time"

	"github.com/MiltonJ23/Orus/internal/port"
)

// LogNotifier implements port.Notifier by logging to the console.
// Replace with a platform-specific adapter (beeep, dbus, etc.) for
// native desktop notifications.
var _ port.Notifier = (*LogNotifier)(nil)

// LogNotifier sends notifications via standard logging output.
type LogNotifier struct{}

// NewLogNotifier creates a new LogNotifier.
func NewLogNotifier() *LogNotifier {
	return &LogNotifier{}
}

// Notify prints a notification to the console with box-drawing characters.
func (n *LogNotifier) Notify(title, message string) error {
	log.Printf("[NOTIFICATION] %s — %s (à %s)", title, message, time.Now().Format("15:04"))
	fmt.Printf("\n╔══════════════════════════════╗\n║  %s\n║  %s\n╚══════════════════════════════╝\n\n", title, message)
	return nil
}
