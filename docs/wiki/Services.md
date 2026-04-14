# Services

The service layer (`internal/service/`) contains the application-level orchestration logic. Services compose domain entities with port interfaces.

## LibraryService

Manages book import and library operations.

| Method | Description |
|--------|-------------|
| `ImportBook(ctx, filePath) (*Book, error)` | Extracts metadata, creates a domain book, and persists it |
| `ImportBooks(ctx, filePaths) ([]*Book, []error)` | Batch import; returns successes and per-file errors |
| `GetLibrary(ctx) ([]*Book, error)` | Lists all books |
| `DeleteBook(ctx, bookID) error` | Permanently removes a book |

**Dependencies:** `BookRepository`, `MetadataExtractor`

---

## TrackerService

Tracks reading progress across sessions.

| Method | Description |
|--------|-------------|
| `OpenBook(ctx, bookID) (*ReadingSession, error)` | Creates or resumes a reading session |
| `UpdateProgress(ctx, page, session) error` | Updates position and persists |
| `GetMostRecentBook(ctx) (*Book, *ReadingSession, error)` | Returns the most recently read book |
| `GetRecentSessions(ctx) ([]*ReadingSession, error)` | Returns latest session for each book |
| `BookCompletionStatus(ctx) (map[string]string, error)` | Returns status map: `"unread"`, `"reading"`, or `"done"` |

**Dependencies:** `BookRepository`, `SessionRepository`

---

## ReadingSheetService

CRUD operations for reading sheets (personal reading notes).

| Method | Description |
|--------|-------------|
| `CreateSheet(ctx, bookID, summary, rating, quotes, tags) (*ReadingSheet, error)` | Creates a reading sheet for a book |
| `GetSheetForBook(ctx, bookID) (*ReadingSheet, error)` | Returns the sheet for a book (nil if none) |
| `ListSheets(ctx) ([]*ReadingSheet, error)` | Lists all sheets |
| `UpdateSummary(ctx, sheetID, newSummary) error` | Updates summary text |
| `SetRating(ctx, sheetID, rating) error` | Updates rating (0–5) |
| `AddQuote(ctx, sheetID, quote) error` | Appends a quote |
| `DeleteSheet(ctx, sheetID) error` | Removes a sheet |

**Dependencies:** `ReadingSheetRepository`, `BookRepository`

---

## ReminderService

Manages reading reminders with a background scheduler.

| Method | Description |
|--------|-------------|
| `AddReminder(ctx, ...) (*Reminder, error)` | Creates and persists a reminder |
| `ListReminders(ctx) ([]*Reminder, error)` | Lists all reminders |
| `ToggleReminder(ctx, id) error` | Enables/disables a reminder |
| `DismissReminder(ctx, id) error` | Acknowledges and advances a reminder |
| `DeleteReminder(ctx, id) error` | Removes a reminder |
| `StartScheduler()` | Runs a 30-second polling loop for due reminders |
| `Stop()` | Stops the scheduler |

The scheduler polls `ListEnabledReminders` every 30 seconds and fires notifications for reminders where `IsDue()` returns true.

**Dependencies:** `ReminderRepository`, `Notifier`

---

## SharingService

Exports library data to files.

| Method | Description |
|--------|-------------|
| `ExportLibrary(ctx, format, outputDir) (string, error)` | Exports entire library |
| `ExportBookInfo(ctx, bookID, format, outputDir) (string, error)` | Exports a single book |
| `ExportReadingSheet(ctx, sheetID, format, outputDir) (string, error)` | Exports a reading sheet |

**Supported formats:** JSON, Markdown, Plain Text

Uses `PickExportDirectory()` to invoke OS-native folder picker dialogs.

**Dependencies:** `BookRepository`, `ReadingSheetRepository`
