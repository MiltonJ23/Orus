package main

import (
	_ "fmt"
	"log"
	"os"
	"path/filepath"

	"gioui.org/app"
	"github.com/MiltonJ23/Orus/internal/adapters/extractor"
	notifier "github.com/MiltonJ23/Orus/internal/adapters/notifier"
	"github.com/MiltonJ23/Orus/internal/adapters/storage/sqlite"
	"github.com/MiltonJ23/Orus/internal/adapters/ui/views"
	"github.com/MiltonJ23/Orus/internal/service"
)

func main() {
	// 1. Infrastructure
	dbPath := filepath.Join(".", "orus.db")
	store, err := sqlite.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("FATAL: failed to initialize storage: %v", err)
	}
	defer store.Close()

	fileExtractor := extractor.NewLocalFileExtractor()
	logNotifier := notifier.NewLogNotifier()

	// 2. Services
	libService := service.NewLibraryService(store, fileExtractor)
	trackerService := service.NewTrackerService(store, store)
	sheetService := service.NewReadingSheetService(store, store)
	reminderService := service.NewReminderService(store, logNotifier)
	sharingService := service.NewSharingService(store, store)

	// 3. Planificateur de rappels (goroutine de fond)
	go reminderService.StartScheduler()
	defer reminderService.Stop()

	// 4. UI
	windowManager := views.NewWindowManager(
		libService,
		trackerService,
		sheetService,
		reminderService,
		sharingService,
	)

	// 5. Boucle UI dans une goroutine
	go func() {
		if err := windowManager.Run(); err != nil {
			log.Fatalf("UI Engine crashed: %v", err)
		}
		os.Exit(0)
	}()

	// 6. Thread principal OS (requis par Gio/macOS)
	app.Main()
}
