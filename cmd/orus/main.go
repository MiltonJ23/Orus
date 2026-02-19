package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/MiltonJ23/Orus/internal/adapters/extractor"
	"github.com/MiltonJ23/Orus/internal/adapters/storage/sqlite"
	"github.com/MiltonJ23/Orus/internal/service"
)

func main() {
	ctx := context.Background()

	// 1. Initialize Infrastructure (Storage)
	// We use a local SQLite file in the current directory for now
	dbPath := filepath.Join(".", "orus.db")
	store, err := sqlite.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("FATAL: failed to initialize storage: %v", err)
	}
	fmt.Println("[OK] Storage initialized")

	fileExtractor := extractor.NewLocalFileExtractor()
	fmt.Println("[OK] Extractors initialized")

	// 3. Initialize Application Services (Injecting Adapters)
	libService := service.NewLibraryService(store, fileExtractor)
	trackerService := service.NewTrackerService(store, store)
	fmt.Println("[OK] Services bound")

	// 4. Sanity Check: Run a dummy import if an argument is provided
	if len(os.Args) > 1 {
		filePath := os.Args[1]
		fmt.Printf("Attempting to import: %s\n", filePath)

		book, err := libService.ImportBook(ctx, filePath)
		if err != nil {
			log.Fatalf("ERR: Import failed: %v", err)
		}

		fmt.Printf("[SUCCESS] Imported '%s' (ID: %s) with %d pages.\n", book.Title, book.ID, book.TotalPages)

		// Test tracker
		session, err := trackerService.OpenBook(ctx, book.ID)
		if err != nil {
			log.Fatalf("ERR: Could not open book: %v", err)
		}
		fmt.Printf("[SUCCESS] Session started. Current Page: %d\n", session.CurrentPage)
	} else {
		fmt.Println("System ready. Run with a file path to test import: ./orus /path/to/book.pdf")
	}
}
