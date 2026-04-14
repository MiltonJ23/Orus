# Domain Model

The domain layer (`internal/domain/`) contains the core business entities. Each entity is self-validating and constructed through factory functions.

## Entities

### Book

Represents an imported book in the library.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `string` | UUID, generated on creation |
| `Title` | `string` | Book title (required) |
| `Author` | `string` | Author name |
| `FilePath` | `string` | Absolute path to the file (required) |
| `Format` | `BookFormat` | `PDF`, `EPUB`, or `MOBI` |
| `TotalPages` | `int` | Total pages or spine items |
| `CoverImage` | `[]byte` | Optional cover image data |
| `AddedAt` | `time.Time` | Import timestamp |
| `UpdatedAt` | `time.Time` | Last update timestamp |

**Factory:** `NewBook(title, author, filePath, format, totalPages) (*Book, error)`

**Errors:**
- `ErrInvalidBookTitle` — empty title
- `ErrBookNotFound` — empty file path

---

### ReadingSession

Tracks the user's reading position within a book.

| Field | Type | Description |
|-------|------|-------------|
| `SessionID` | `string` | UUID |
| `BookID` | `string` | Reference to the book |
| `TotalPages` | `int` | Book's total pages |
| `CurrentPage` | `int` | Current reading position |
| `LastReadingTime` | `time.Time` | Last activity timestamp |

**Methods:**
- `CalculateCompletion() float64` — returns progress as 0.0–100.0%
- `IsBookComplete() bool` — true when `CurrentPage >= TotalPages`
- `UpdatePosition(page int)` — moves to a new page (clamped to bounds)

---

### Annotation

Represents a bookmark or highlight on a specific page.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `string` | UUID |
| `BookID` | `string` | Reference to the book |
| `AnnotationType` | `AnnotationType` | `bookmark` or `highlight` |
| `PageNo` | `int` | Target page number |
| `CreatedAt` | `time.Time` | Creation timestamp |

**Factory:** `NewAnnotation(bookID, annotationType, pageNo) (*Annotation, error)`

---

### ReadingSheet

Personal reading notes for a book.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `string` | UUID |
| `BookID` | `string` | Reference to the book |
| `BookTitle` | `string` | Denormalized book title |
| `Summary` | `string` | Personal summary |
| `Quotes` | `[]string` | Favorite quotes |
| `Rating` | `int` | Rating 0–5 |
| `Tags` | `[]string` | Tags (e.g., "roman", "histoire") |
| `CreatedAt` | `time.Time` | Creation timestamp |
| `UpdatedAt` | `time.Time` | Last update timestamp |

**Methods:**
- `AddQuote(quote string)` — appends a trimmed quote
- `RemoveQuote(index int)` — removes by index
- `UpdateSummary(summary string)` — updates summary
- `UpdateRating(rating int) error` — validates and updates rating
- `StarString() string` — returns `"★★★☆☆"` representation

---

### Reminder

A scheduled reading reminder.

| Field | Type | Description |
|-------|------|-------------|
| `ID` | `string` | UUID |
| `BookID` | `string` | Optional book reference |
| `BookTitle` | `string` | Denormalized title |
| `Label` | `string` | Reminder message |
| `Hour` | `int` | Hour (0–23) |
| `Minute` | `int` | Minute (0–59) |
| `Frequency` | `ReminderFrequency` | `daily`, `weekly`, `weekdays`, `once` |
| `Enabled` | `bool` | Active state |
| `NextRing` | `time.Time` | Next scheduled occurrence |
| `CreatedAt` | `time.Time` | Creation timestamp |

**Methods:**
- `ComputeNextRing(from time.Time) time.Time` — calculates next occurrence
- `IsDue(now time.Time) bool` — true if due within 1-minute tolerance
- `Advance(from time.Time)` — advances `NextRing` after firing
- `FrequencyLabel() string` — human-readable frequency string
