package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

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

	// if the tables do not already exist, let's create them
	creatingTablesError := createTables(db)
	if creatingTablesError != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables in sqlite database: %s", creatingTablesError.Error())
	}

	return &Storage{db: db}, nil
}

func createTables(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS books (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		author TEXT,
		file_path TEXT NOT NULL,
		format TEXT,
		total_pages INTEGER,
		added_at DATETIME
	);
	
	CREATE TABLE IF NOT EXISTS sessions (
    session_id TEXT PRIMARY KEY, -- Unique ID for every entry
    book_id TEXT NOT NULL,                         -- Links to the book
    current_page INTEGER DEFAULT 0,
    last_read_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(book_id) REFERENCES books(id) ON DELETE CASCADE
);

	CREATE TABLE IF NOT EXISTS annotations (
	    id TEXT PRIMARY KEY,
	    book_id TEXT NOT NULL,
	    annotation_type TEXT NOT NULl,
	    page_number INTEGER DEFAULT 0,
	    created_at DATETIME,
	    FOREIGN KEY(book_id) REFERENCES books(id) ON DELETE CASCADE
	)
`

	_, queryExecutionError := db.Exec(query)
	return queryExecutionError
}

func (s *Storage) Close() error {

	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}
