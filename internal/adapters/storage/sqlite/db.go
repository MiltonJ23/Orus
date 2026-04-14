package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

// Storage provides SQLite-backed persistence for all repository interfaces.
type Storage struct {
	db *sql.DB
}

// NewStorage opens a SQLite database at dbPath, enables foreign keys, and
// creates all required tables. The caller must call Close() when finished.
func NewStorage(dbPath string) (*Storage, error) {
	if dbPath == "" {
		return nil, errors.New("dbPath cannot be empty")
	}

	db, dbReadingError := sql.Open("sqlite", "file:"+dbPath+"?_pragma=foreign_keys(1)")
	if dbReadingError != nil {
		return nil, fmt.Errorf("failed to connect to sqlite database: %s", dbReadingError.Error())
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	creatingTablesError := createTables(db)
	if creatingTablesError != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables in sqlite database: %s", creatingTablesError.Error())
	}

	return &Storage{db: db}, nil
}

func createTables(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS books (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		author TEXT,
		file_path TEXT NOT NULL,
		format TEXT,
		total_pages INTEGER,
		added_at DATETIME
	);

	CREATE TABLE IF NOT EXISTS sessions (
		session_id TEXT PRIMARY KEY,
		book_id TEXT NOT NULL,
		current_page INTEGER DEFAULT 0,
		last_read_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(book_id) REFERENCES books(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS annotations (
		id TEXT PRIMARY KEY,
		book_id TEXT NOT NULL,
		annotation_type TEXT NOT NULL,
		page_number INTEGER DEFAULT 0,
		created_at DATETIME,
		FOREIGN KEY(book_id) REFERENCES books(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS reading_sheets (
		id TEXT PRIMARY KEY,
		book_id TEXT NOT NULL,
		book_title TEXT NOT NULL,
		summary TEXT DEFAULT '',
		quotes TEXT DEFAULT '',   -- citations séparées par "||"
		rating INTEGER DEFAULT 0, -- 0 à 5
		tags TEXT DEFAULT '',     -- tags séparés par ","
		created_at DATETIME,
		updated_at DATETIME,
		FOREIGN KEY(book_id) REFERENCES books(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS reminders (
		id TEXT PRIMARY KEY,
		book_id TEXT DEFAULT '',     -- vide = rappel global
		book_title TEXT DEFAULT '',
		label TEXT NOT NULL,
		hour INTEGER NOT NULL,       -- 0-23
		minute INTEGER NOT NULL,     -- 0-59
		frequency TEXT NOT NULL,     -- daily | weekly | weekdays | once
		enabled INTEGER DEFAULT 1,   -- 0 ou 1 (booléen SQLite)
		next_ring DATETIME,
		created_at DATETIME
	);
	`

	_, queryExecutionError := db.Exec(query)
	return queryExecutionError
}

// Close closes the underlying database connection.
func (s *Storage) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}
