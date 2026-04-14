# Storage Layer

The storage layer (`internal/adapters/storage/sqlite/`) implements all repository interfaces using SQLite.

## Database Initialization

`NewStorage(dbPath)` opens a SQLite connection with foreign keys enabled and creates all tables if they don't exist.

```go
store, err := sqlite.NewStorage("./orus.db")
if err != nil {
    log.Fatal(err)
}
defer store.Close()
```

The `Storage` struct implements all five repository interfaces:
- `port.BookRepository`
- `port.SessionRepository`
- `port.AnnotationRepository`
- `port.ReadingSheetRepository`
- `port.ReminderRepository`

## Schema

### books

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | TEXT | PRIMARY KEY |
| `title` | TEXT | NOT NULL |
| `author` | TEXT | |
| `file_path` | TEXT | NOT NULL |
| `format` | TEXT | |
| `total_pages` | INTEGER | |
| `added_at` | DATETIME | |

### sessions

| Column | Type | Constraints |
|--------|------|-------------|
| `session_id` | TEXT | PRIMARY KEY |
| `book_id` | TEXT | NOT NULL, FK → books(id) CASCADE |
| `current_page` | INTEGER | DEFAULT 0 |
| `last_read_time` | DATETIME | DEFAULT CURRENT_TIMESTAMP |

### annotations

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | TEXT | PRIMARY KEY |
| `book_id` | TEXT | NOT NULL, FK → books(id) CASCADE |
| `annotation_type` | TEXT | NOT NULL |
| `page_number` | INTEGER | DEFAULT 0 |
| `created_at` | DATETIME | |

### reading_sheets

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | TEXT | PRIMARY KEY |
| `book_id` | TEXT | NOT NULL, FK → books(id) CASCADE |
| `book_title` | TEXT | NOT NULL |
| `summary` | TEXT | DEFAULT '' |
| `quotes` | TEXT | DEFAULT '' (separated by `\|\|`) |
| `rating` | INTEGER | DEFAULT 0 (0–5) |
| `tags` | TEXT | DEFAULT '' (separated by `,`) |
| `created_at` | DATETIME | |
| `updated_at` | DATETIME | |

### reminders

| Column | Type | Constraints |
|--------|------|-------------|
| `id` | TEXT | PRIMARY KEY |
| `book_id` | TEXT | DEFAULT '' |
| `book_title` | TEXT | DEFAULT '' |
| `label` | TEXT | NOT NULL |
| `hour` | INTEGER | NOT NULL (0–23) |
| `minute` | INTEGER | NOT NULL (0–59) |
| `frequency` | TEXT | NOT NULL |
| `enabled` | INTEGER | DEFAULT 1 |
| `next_ring` | DATETIME | |
| `created_at` | DATETIME | |

## Design Decisions

- **Context timeouts:** All repository methods enforce a 5-second context timeout.
- **UPSERT pattern:** `Save()` for books uses `INSERT OR REPLACE` for idempotent writes.
- **Serialization:** Quotes are serialized as `||`-delimited strings, tags as `,`-delimited.
- **Foreign keys:** Enabled via pragma; all child tables use `ON DELETE CASCADE`.
