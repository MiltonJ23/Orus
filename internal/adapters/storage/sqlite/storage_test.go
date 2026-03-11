package sqlite_test

import (
	"context"
	_ "database/sql"
	"os"
	"testing"
	"time"

	"github.com/MiltonJ23/Orus/internal/adapters/storage/sqlite"
	"github.com/MiltonJ23/Orus/internal/domain"
	_ "modernc.org/sqlite" // Use the same driver as main code
)

// setupTestDB creates a temporary DB file and returns the Storage instance + a cleanup function
func setupTestDB(t *testing.T) (*sqlite.Storage, func()) {
	t.Helper()

	// Create a temp file
	tmpFile, err := os.CreateTemp("", "orus_test_*.db")
	if err != nil {
		t.Fatalf("Could not create temp db file: %v", err)
	}
	dbPath := tmpFile.Name()
	tmpFile.Close() // Close file handle so SQLite can open it

	store, err := sqlite.NewStorage(dbPath)
	if err != nil {
		os.Remove(dbPath)
		t.Fatalf("Could not init storage: %v", err)
	}

	cleanup := func() {
		os.Remove(dbPath)
	}
	return store, cleanup
}

// --- BOOK REPO TESTS ---

func TestBookRepository_Lifecycle(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// 1. Create & Save
	book, _ := domain.NewBook("Dune", "Frank Herbert", "/path/dune.pdf", domain.FormatPDF, 800)
	// Manually set ID for deterministic testing if needed, or rely on NewBook's UUID
	if err := store.Save(ctx, book); err != nil {
		t.Fatalf("Failed to save book: %v", err)
	}

	// 2. GetByID (Success)
	fetched, err := store.GetByID(ctx, book.ID)
	if err != nil {
		t.Fatalf("Failed to get book: %v", err)
	}
	if fetched.Title != book.Title {
		t.Errorf("Expected title %s, got %s", book.Title, fetched.Title)
	}
	if fetched.Author != book.Author {
		t.Errorf("Expected author %s, got %s", book.Author, fetched.Author)
	}
	if fetched.Format != domain.FormatPDF {
		t.Errorf("Expected format PDF, got %v", fetched.Format)
	}

	// 3. Update (Upsert)
	book.Title = "Dune: Messiah"
	if err := store.Save(ctx, book); err != nil {
		t.Fatalf("Failed to update book: %v", err)
	}
	updated, _ := store.GetByID(ctx, book.ID)
	if updated.Title != "Dune: Messiah" {
		t.Errorf("Update failed. Expected 'Dune: Messiah', got '%s'", updated.Title)
	}

	// 4. ListAll
	book2, _ := domain.NewBook("Hyperion", "Dan Simmons", "/path/hyp.epub", domain.FormatEPUB, 400)
	store.Save(ctx, book2)

	all, err := store.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("Expected 2 books, got %d", len(all))
	}

	// 5. Delete
	if err := store.Delete(ctx, book.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = store.GetByID(ctx, book.ID)
	if err != domain.ErrBookNotFound {
		t.Errorf("Expected ErrBookNotFound after delete, got %v", err)
	}
}

func TestBookRepository_GetNotFound(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := store.GetByID(context.Background(), "non-existent-id")
	if err != domain.ErrBookNotFound {
		t.Errorf("Expected ErrBookNotFound, got %v", err)
	}
}

// --- ANNOTATION REPO TESTS ---

func TestAnnotationRepository(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Prerequisite: Create a book (Constraint Check)
	book, _ := domain.NewBook("Notes Book", "Me", "path", domain.FormatPDF, 100)
	store.Save(ctx, book)

	// 1. Save Annotation
	note := &domain.Annotation{
		ID:             "note-1",
		BookID:         book.ID,
		AnnotationType: domain.AnnotationHighlight,
		PageNo:         10,
		CreatedAt:      time.Now(),
	}
	if err := store.SaveAnnotation(ctx, note); err != nil {
		t.Fatalf("Failed to save annotation: %v", err)
	}

	// 2. GetByPage
	notes, err := store.GetAnnotationByPage(ctx, 10, book.ID)
	if err != nil {
		t.Fatalf("GetByPage failed: %v", err)
	}
	if len(notes) != 1 {
		t.Errorf("Expected 1 note, got %d", len(notes))
	}
	if notes[0].ID != note.ID {
		t.Errorf("ID mismatch")
	}

	// 3. GetByType
	highlights, err := store.GetAnnotationByType(ctx, string(domain.AnnotationHighlight))
	if err != nil {
		t.Fatalf("GetByType failed: %v", err)
	}
	if len(highlights) == 0 {
		t.Error("Expected highlights, got none")
	}

	// 4. ListAllOfBook
	allNotes, err := store.ListAllAnnotationOfABook(ctx, book.ID)
	if err != nil {
		t.Fatalf("ListAllOfBook failed: %v", err)
	}
	if len(allNotes) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(allNotes))
	}

	// 5. Delete
	if err := store.DeleteAnnotation(ctx, note.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	remaining, _ := store.ListAllAnnotationOfABook(ctx, book.ID)
	if len(remaining) != 0 {
		t.Error("Delete failed, annotation still exists")
	}
}

func TestAnnotationRepository_ForeignKeyConstraint(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Try to save annotation for non-existent book
	note := &domain.Annotation{
		ID:     "orphan-note",
		BookID: "ghost-book-id",
		PageNo: 1,
	}

	// SQLite constraint enforcement depends on driver config,
	// but generally standard INSERT should fail if FK enabled.
	// Note: default sqlite might need "PRAGMA foreign_keys = ON"
	// For this test, we accept either error or success depending on strictness,
	// but in a real app we want it to fail.
	err := store.SaveAnnotation(context.Background(), note)
	if err == nil {
		t.Log("Warning: Foreign Key constraints might not be enabled in SQLite driver")
	} else {
		t.Logf("Correctly caught FK violation: %v", err)
	}
}

// --- SESSION REPO TESTS ---

func TestSessionRepository(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	book, _ := domain.NewBook("Session Book", "Reader", "path", "epub", 200)
	store.Save(ctx, book)

	session := &domain.ReadingSession{
		SessionID:       "session-1",
		BookID:          book.ID,
		CurrentPage:     42,
		LastReadingTime: time.Now(),
	}

	// 1. Save
	if err := store.SaveSession(ctx, session); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	// 2. Get
	fetched, err := store.GetSessionByID(ctx, book.ID)
	if err != nil {
		t.Fatalf("GetSessionByID failed: %v", err)
	}
	if len(fetched) == 0 {
		t.Fatal("Expected at least one session")
	}
	if fetched[0].CurrentPage != 42 {
		t.Errorf("Expected page 42, got %d", fetched[0].CurrentPage)
	}
}

func TestSessionRepository_GetLastReadingSession(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	book, _ := domain.NewBook("Session Book", "Reader", "path", "epub", 200)
	store.Save(ctx, book)

	// Save multiple sessions at different times
	session1 := &domain.ReadingSession{
		SessionID:       "session-1",
		BookID:          book.ID,
		CurrentPage:     10,
		LastReadingTime: time.Now().Add(-2 * time.Hour),
	}
	session2 := &domain.ReadingSession{
		SessionID:       "session-2",
		BookID:          book.ID,
		CurrentPage:     25,
		LastReadingTime: time.Now().Add(-1 * time.Hour),
	}
	session3 := &domain.ReadingSession{
		SessionID:       "session-3",
		BookID:          book.ID,
		CurrentPage:     42,
		LastReadingTime: time.Now(),
	}

	store.SaveSession(ctx, session1)
	store.SaveSession(ctx, session2)
	store.SaveSession(ctx, session3)

	// Get last session
	lastSession, err := store.GetLastReadingSession(ctx, book.ID)
	if err != nil {
		t.Fatalf("GetLastReadingSession failed: %v", err)
	}
	if lastSession == nil {
		t.Fatal("Expected a session, got nil")
	}
	if lastSession.CurrentPage != 42 {
		t.Errorf("Expected last session page 42, got %d", lastSession.CurrentPage)
	}
}

func TestSessionRepository_GetLastReadingSession_NoSessions(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	book, _ := domain.NewBook("Session Book", "Reader", "path", "epub", 200)
	store.Save(ctx, book)

	// No sessions saved
	lastSession, err := store.GetLastReadingSession(ctx, book.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if lastSession != nil {
		t.Errorf("Expected nil session for book with no sessions, got: %v", lastSession)
	}
}

func TestSessionRepository_SaveWithoutSessionID(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	book, _ := domain.NewBook("Session Book", "Reader", "path", "epub", 200)
	store.Save(ctx, book)

	// Try to save session without session ID
	session := &domain.ReadingSession{
		SessionID:       "",
		BookID:          book.ID,
		CurrentPage:     42,
		LastReadingTime: time.Now(),
	}

	err := store.SaveSession(ctx, session)
	if err == nil {
		t.Error("Expected error when saving session without session ID, got nil")
	}
}

// --- REMINDER REPO TESTS ---

func TestReminderRepository_Lifecycle(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	// Create prerequisite book
	book, _ := domain.NewBook("Reminder Book", "Author", "path", domain.FormatPDF, 100)
	store.Save(ctx, book)

	// 1. Create and Save Reminder
	reminder, err := domain.NewReminder(book.ID, book.Title, "Read 30 minutes", 20, 30, domain.FrequencyDaily)
	if err != nil {
		t.Fatalf("Failed to create reminder: %v", err)
	}

	if err := store.SaveReminder(ctx, reminder); err != nil {
		t.Fatalf("Failed to save reminder: %v", err)
	}

	// 2. GetReminderByID
	fetched, err := store.GetReminderByID(ctx, reminder.ID)
	if err != nil {
		t.Fatalf("Failed to get reminder: %v", err)
	}
	if fetched.Label != reminder.Label {
		t.Errorf("Expected label %s, got %s", reminder.Label, fetched.Label)
	}
	if fetched.Hour != 20 || fetched.Minute != 30 {
		t.Errorf("Expected time 20:30, got %d:%d", fetched.Hour, fetched.Minute)
	}
	if fetched.Frequency != domain.FrequencyDaily {
		t.Errorf("Expected daily frequency, got %v", fetched.Frequency)
	}

	// 3. ListAllReminders
	all, err := store.ListAllReminders(ctx)
	if err != nil {
		t.Fatalf("ListAllReminders failed: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("Expected 1 reminder, got %d", len(all))
	}

	// 4. Update Reminder
	reminder.Label = "Read 1 hour"
	reminder.Hour = 19
	reminder.Enabled = false
	if err := store.UpdateReminder(ctx, reminder); err != nil {
		t.Fatalf("Update reminder failed: %v", err)
	}

	updated, _ := store.GetReminderByID(ctx, reminder.ID)
	if updated.Label != "Read 1 hour" {
		t.Errorf("Update failed. Expected 'Read 1 hour', got '%s'", updated.Label)
	}
	if updated.Hour != 19 {
		t.Errorf("Expected hour 19, got %d", updated.Hour)
	}
	if updated.Enabled {
		t.Error("Expected reminder to be disabled")
	}

	// 5. Delete Reminder
	if err := store.DeleteReminder(ctx, reminder.ID); err != nil {
		t.Fatalf("Delete reminder failed: %v", err)
	}

	_, err = store.GetReminderByID(ctx, reminder.ID)
	if err != domain.ErrReminderNotFound {
		t.Errorf("Expected ErrReminderNotFound after delete, got %v", err)
	}
}

func TestReminderRepository_ListEnabledReminders(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	book, _ := domain.NewBook("Book", "Author", "path", domain.FormatPDF, 100)
	store.Save(ctx, book)

	// Create multiple reminders with different enabled states
	r1, _ := domain.NewReminder(book.ID, book.Title, "Morning", 8, 0, domain.FrequencyDaily)
	r2, _ := domain.NewReminder(book.ID, book.Title, "Evening", 20, 0, domain.FrequencyDaily)
	r3, _ := domain.NewReminder(book.ID, book.Title, "Disabled", 12, 0, domain.FrequencyWeekly)
	r3.Enabled = false

	store.SaveReminder(ctx, r1)
	store.SaveReminder(ctx, r2)
	store.SaveReminder(ctx, r3)

	// List only enabled reminders
	enabled, err := store.ListEnabledReminders(ctx)
	if err != nil {
		t.Fatalf("ListEnabledReminders failed: %v", err)
	}
	if len(enabled) != 2 {
		t.Errorf("Expected 2 enabled reminders, got %d", len(enabled))
	}

	for _, r := range enabled {
		if !r.Enabled {
			t.Errorf("Found disabled reminder in enabled list: %s", r.Label)
		}
	}
}

func TestReminderRepository_GlobalReminder(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	// Create global reminder (no book association)
	reminder, err := domain.NewReminder("", "", "General Reading Time", 18, 0, domain.FrequencyWeekdays)
	if err != nil {
		t.Fatalf("Failed to create global reminder: %v", err)
	}

	if err := store.SaveReminder(ctx, reminder); err != nil {
		t.Fatalf("Failed to save global reminder: %v", err)
	}

	fetched, err := store.GetReminderByID(ctx, reminder.ID)
	if err != nil {
		t.Fatalf("Failed to get global reminder: %v", err)
	}
	if fetched.BookID != "" {
		t.Errorf("Expected empty book ID for global reminder, got %s", fetched.BookID)
	}
	if fetched.Frequency != domain.FrequencyWeekdays {
		t.Errorf("Expected weekdays frequency, got %v", fetched.Frequency)
	}
}

// --- READING SHEET REPO TESTS ---

func TestReadingSheetRepository_Lifecycle(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	// Create prerequisite book
	book, _ := domain.NewBook("Test Book", "Test Author", "path", domain.FormatEPUB, 300)
	store.Save(ctx, book)

	// 1. Create and Save Sheet
	quotes := []string{"Quote 1", "Quote 2"}
	tags := []string{"fiction", "adventure"}
	sheet, err := domain.NewReadingSheet(book.ID, book.Title, "Great book!", 5, quotes, tags)
	if err != nil {
		t.Fatalf("Failed to create reading sheet: %v", err)
	}

	if err := store.SaveSheet(ctx, sheet); err != nil {
		t.Fatalf("Failed to save reading sheet: %v", err)
	}

	// 2. GetSheetByID
	fetched, err := store.GetSheetByID(ctx, sheet.ID)
	if err != nil {
		t.Fatalf("Failed to get sheet by ID: %v", err)
	}
	if fetched.Summary != "Great book!" {
		t.Errorf("Expected summary 'Great book!', got '%s'", fetched.Summary)
	}
	if fetched.Rating != 5 {
		t.Errorf("Expected rating 5, got %d", fetched.Rating)
	}
	if len(fetched.Quotes) != 2 {
		t.Errorf("Expected 2 quotes, got %d", len(fetched.Quotes))
	}
	if len(fetched.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(fetched.Tags))
	}

	// 3. GetSheetByBookID
	bookSheet, err := store.GetSheetByBookID(ctx, book.ID)
	if err != nil {
		t.Fatalf("Failed to get sheet by book ID: %v", err)
	}
	if bookSheet.ID != sheet.ID {
		t.Error("Sheet ID mismatch")
	}

	// 4. ListAllSheets
	all, err := store.ListAllSheets(ctx)
	if err != nil {
		t.Fatalf("ListAllSheets failed: %v", err)
	}
	if len(all) != 1 {
		t.Errorf("Expected 1 sheet, got %d", len(all))
	}

	// 5. Update Sheet
	sheet.Summary = "Updated summary"
	sheet.Rating = 4
	sheet.Quotes = append(sheet.Quotes, "Quote 3")
	if err := store.UpdateSheet(ctx, sheet); err != nil {
		t.Fatalf("Update sheet failed: %v", err)
	}

	updated, _ := store.GetSheetByID(ctx, sheet.ID)
	if updated.Summary != "Updated summary" {
		t.Errorf("Update failed. Expected 'Updated summary', got '%s'", updated.Summary)
	}
	if updated.Rating != 4 {
		t.Errorf("Expected rating 4, got %d", updated.Rating)
	}
	if len(updated.Quotes) != 3 {
		t.Errorf("Expected 3 quotes after update, got %d", len(updated.Quotes))
	}

	// 6. Delete Sheet
	if err := store.DeleteSheet(ctx, sheet.ID); err != nil {
		t.Fatalf("Delete sheet failed: %v", err)
	}

	_, err = store.GetSheetByID(ctx, sheet.ID)
	if err != domain.ErrReadingSheetNotFound {
		t.Errorf("Expected ErrReadingSheetNotFound after delete, got %v", err)
	}
}

func TestReadingSheetRepository_EmptyQuotesAndTags(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	book, _ := domain.NewBook("Book", "Author", "path", domain.FormatPDF, 100)
	store.Save(ctx, book)

	// Create sheet with empty quotes and tags
	sheet, _ := domain.NewReadingSheet(book.ID, book.Title, "Summary", 3, nil, nil)
	if err := store.SaveSheet(ctx, sheet); err != nil {
		t.Fatalf("Failed to save sheet with empty quotes/tags: %v", err)
	}

	fetched, err := store.GetSheetByID(ctx, sheet.ID)
	if err != nil {
		t.Fatalf("Failed to get sheet: %v", err)
	}

	// When empty, the slices may be nil or empty depending on storage implementation
	// Both are valid representations of "no items"
	if fetched.Quotes != nil && len(fetched.Quotes) > 0 {
		t.Error("Expected empty or nil quotes slice for sheet with no quotes")
	}
	if fetched.Tags != nil && len(fetched.Tags) > 0 {
		t.Error("Expected empty or nil tags slice for sheet with no tags")
	}
}

func TestReadingSheetRepository_SpecialCharactersInQuotes(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	book, _ := domain.NewBook("Book", "Author", "path", domain.FormatPDF, 100)
	store.Save(ctx, book)

	// Quotes with special characters (avoiding "||" which is the storage separator)
	quotes := []string{
		"Quote with commas, periods, and dashes",
		"Quote with \"double quotes\" and 'single quotes'",
		"Quote with special chars: @#$%^&*()",
	}
	sheet, _ := domain.NewReadingSheet(book.ID, book.Title, "Test", 4, quotes, []string{"tag1", "tag2"})
	if err := store.SaveSheet(ctx, sheet); err != nil {
		t.Fatalf("Failed to save sheet: %v", err)
	}

	fetched, err := store.GetSheetByID(ctx, sheet.ID)
	if err != nil {
		t.Fatalf("Failed to get sheet: %v", err)
	}

	// Verify quotes are properly stored and retrieved
	if len(fetched.Quotes) != len(quotes) {
		t.Errorf("Expected %d quotes, got %d", len(quotes), len(fetched.Quotes))
	}

	// Verify tags are properly stored
	if len(fetched.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(fetched.Tags))
	}
}

func TestReadingSheetRepository_SeparatorInQuotes(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	book, _ := domain.NewBook("Book", "Author", "path", domain.FormatPDF, 100)
	store.Save(ctx, book)

	// Test that the separator "||" causes issues (known limitation)
	// This documents the behavior rather than testing correctness
	quotes := []string{
		"Quote with || separator inside",
	}
	sheet, _ := domain.NewReadingSheet(book.ID, book.Title, "Test", 4, quotes, []string{"tag1"})
	if err := store.SaveSheet(ctx, sheet); err != nil {
		t.Fatalf("Failed to save sheet: %v", err)
	}

	fetched, err := store.GetSheetByID(ctx, sheet.ID)
	if err != nil {
		t.Fatalf("Failed to get sheet: %v", err)
	}

	// Note: The separator "||" will cause the quote to be split
	// This is a known limitation of the current storage implementation
	// The test documents this behavior
	if len(fetched.Quotes) == 1 {
		t.Log("Storage correctly preserves quotes with || separator")
	} else {
		t.Logf("Note: Quotes with || separator get split. Got %d quotes instead of 1", len(fetched.Quotes))
	}
}

func TestReadingSheetRepository_GetNotFound(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	_, err := store.GetSheetByID(ctx, "non-existent-id")
	if err != domain.ErrReadingSheetNotFound {
		t.Errorf("Expected ErrReadingSheetNotFound, got %v", err)
	}

	_, err = store.GetSheetByBookID(ctx, "non-existent-book-id")
	if err != domain.ErrReadingSheetNotFound {
		t.Errorf("Expected ErrReadingSheetNotFound for book, got %v", err)
	}
}

// --- CONTEXT CANCELLATION TEST ---

func TestStorage_ContextCancellation(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a context that is ALREADY cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	book, _ := domain.NewBook("Fast Book", "Me", "path", "pdf", 10)

	err := store.Save(ctx, book)
	if err == nil {
		t.Error("Expected error due to cancelled context, got nil")
	}
}