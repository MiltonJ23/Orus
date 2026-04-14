# Architecture Overview

Orus follows the **Clean Architecture** pattern (see [ADR-0001](../adr/0001-record-architecture-pattern.md)). The codebase is organized into four concentric layers, each with a strict dependency direction: outer layers depend on inner layers, never the reverse.

## Layers

### 1. Domain Layer (`internal/domain/`)

The innermost layer containing business entities and validation rules. This layer has **zero external dependencies** — it depends only on the Go standard library and `github.com/google/uuid`.

**Entities:**
- `Book` — represents an imported book with metadata
- `ReadingSession` — tracks a user's reading position in a book
- `Annotation` — bookmarks and highlights on specific pages
- `ReadingSheet` — personal reading notes (summary, quotes, rating, tags)
- `Reminder` — scheduled reading reminders with frequency management

Each entity has a factory function (e.g., `NewBook()`) that validates input and returns a fully constructed instance or an error.

### 2. Port Layer (`internal/port/`)

Defines interfaces that decouple the application from infrastructure:

| Interface | Purpose |
|-----------|---------|
| `BookRepository` | CRUD operations for books |
| `SessionRepository` | Reading session persistence |
| `AnnotationRepository` | Bookmark/highlight persistence |
| `ReadingSheetRepository` | Reading sheet persistence |
| `ReminderRepository` | Reminder persistence |
| `ContentReader` | Text extraction from files |
| `MetadataExtractor` | Metadata extraction from files |
| `Notifier` | System notification delivery |

### 3. Service Layer (`internal/service/`)

Application-level orchestration. Each service composes one or more ports:

| Service | Dependencies | Responsibility |
|---------|-------------|----------------|
| `LibraryService` | `BookRepository`, `MetadataExtractor` | Book import and library management |
| `TrackerService` | `BookRepository`, `SessionRepository` | Reading session tracking |
| `ReadingSheetService` | `ReadingSheetRepository`, `BookRepository` | Reading sheet CRUD |
| `ReminderService` | `ReminderRepository`, `Notifier` | Reminder scheduling and notification |
| `SharingService` | `BookRepository`, `ReadingSheetRepository` | Library export (JSON/Markdown/Text) |

### 4. Adapter Layer (`internal/adapters/`)

Infrastructure implementations:

| Adapter | Implements | Technology |
|---------|-----------|------------|
| `sqlite.Storage` | All repository interfaces | SQLite via `modernc.org/sqlite` |
| `extractor.LocalFileExtractor` | `ContentReader`, `MetadataExtractor` | `ledongthuc/pdf`, `kapmahc/epub` |
| `notifier.LogNotifier` | `Notifier` | Console logging |
| `views.WindowManager` | UI controller | Gio UI framework |

## Dependency Graph

```
main.go
  ├─→ service.LibraryService
  ├─→ service.TrackerService
  ├─→ service.ReadingSheetService
  ├─→ service.ReminderService
  ├─→ service.SharingService
  │
  ├─→ sqlite.Storage          (implements all port.Repository interfaces)
  ├─→ extractor.LocalFileExtractor (implements port.ContentReader, port.MetadataExtractor)
  ├─→ notifier.LogNotifier    (implements port.Notifier)
  └─→ views.WindowManager     (UI entry point)
```

## Compile-Time Interface Assertions

Each adapter includes a compile-time assertion to guarantee interface compliance:

```go
var _ port.BookRepository = (*Storage)(nil)
var _ port.ContentReader  = (*LocalFileExtractor)(nil)
var _ port.Notifier       = (*LogNotifier)(nil)
```
