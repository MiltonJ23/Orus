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
	dbPath := filepath.Join(".", "orus.db")
	store, err := sqlite.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("FATAL: failed to initialize storage: %v", err)
	}
	defer store.Close()

	fileExtractor := extractor.NewLocalFileExtractor()
	logNotifier := notifier.NewLogNotifier()

	libService := service.NewLibraryService(store, fileExtractor)
	trackerService := service.NewTrackerService(store, store)
	sheetService := service.NewReadingSheetService(store, store)
	reminderService := service.NewReminderService(store, logNotifier)
	sharingService := service.NewSharingService(store, store)

	go reminderService.StartScheduler()
	defer reminderService.Stop()

	// fileExtractor implémente port.ContentReader (ReadBookText)
	windowManager := views.NewWindowManager(
		libService,
		trackerService,
		sheetService,
		reminderService,
		sharingService,
		fileExtractor,
	)

	go func() {
		if err := windowManager.Run(); err != nil {
			log.Fatalf("UI Engine crashed: %v", err)
		}
		os.Exit(0)
	}()

	app.Main()
}
