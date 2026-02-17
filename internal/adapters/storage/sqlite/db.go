package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(dbPath string) (*Storage, error) {
	if dbPath == "" {
		return nil, errors.New("dbPath cannot be empty")
	}

	db, dbReadingError := sql.Open("sqlite3", dbPath)
	if dbReadingError != nil {
		return nil, fmt.Errorf("failed to connect to sqlite database: %s", dbReadingError.Error())
	}

	// if the tables do not already exist, let's create them

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
		book_id TEXT PRIMARY KEY,
		current_page INTEGER,
		total_pages INTEGER,
		last_read_time DATETIME,
		FOREIGN KEY(book_id) REFERENCES books(id)
	);`

	_, queryExecutionError := db.Exec(query)
	return queryExecutionError
}
