package notifier

import (
	"fmt"
	"log"
	"time"

	"github.com/MiltonJ23/Orus/internal/port"
)

// LogNotifier est l'implémentation par défaut : log en console + stdout.
// Remplacer par un adaptateur OS natif (beeep, dbus, etc.) selon la plateforme.
var _ port.Notifier = (*LogNotifier)(nil)

type LogNotifier struct{}

func NewLogNotifier() *LogNotifier {
	return &LogNotifier{}
}

func (n *LogNotifier) Notify(title, message string) error {
	log.Printf("[NOTIFICATION] %s — %s (à %s)", title, message, time.Now().Format("15:04"))
	fmt.Printf("\n╔══════════════════════════════╗\n║  %s\n║  %s\n╚══════════════════════════════╝\n\n", title, message)
	return nil
}
