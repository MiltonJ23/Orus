package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/port"
)

var _ port.BookRepository = (*Storage)(nil) // an interface assertion to check at compile time that Storage implements the BookRepository interface

func (s *Storage) Save(ctx context.Context, book *domain.Book) error {
	// let's handle the context so that the operation doesn't exceed my defined time limit
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// first, let's build the query || the query is a kind of UPSERT
	query := `INSERT INTO books (id, title, author, file_path, format, total_pages, added_at) VALUES (?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET title=excluded.title, file_path=excluded.file_path`

	// then, let's execute the query
	_, queryExecutionerr := s.db.ExecContext(ctx, query, book.ID, book.Title, book.Author, book.FilePath, book.Format, book.TotalPages, book.AddedAt)
	return queryExecutionerr

}

func (s *Storage) GetByID(ctx context.Context, id string) (*domain.Book, error) {

	// let's handle the context so that the operation doesn't exceed my defined time limit
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// let's build the query
	query := `SELECT * FROM books WHERE id=?`

	row := s.db.QueryRowContext(ctx, query, id)

	var (
		b         domain.Book
		formatStr string
	)

	copyingDataFromRowError := row.Scan(&b.ID, &b.Title, &b.Author, &b.FilePath, &formatStr, &b.TotalPages, &b.AddedAt)
	if copyingDataFromRowError != nil {
		// maybe because the row was empty
		if errors.Is(copyingDataFromRowError, sql.ErrNoRows) {
			return nil, domain.ErrBookNotFound
		}
		return nil, fmt.Errorf("unable to scan the data from database row, %v", copyingDataFromRowError)
	}
	b.Format = domain.BookFormat(formatStr)
	return &b, nil
}

func (s *Storage) ListAll(ctx context.Context) ([]*domain.Book, error) {
	// first of all, we manage the context lifecycle
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var (
		bookList []*domain.Book
	)
	// now let's build the query to fetch the entries
	query := `SELECT * FROM books`

	rows, fetchingError := s.db.QueryContext(ctx, query)
	if fetchingError != nil {
		return nil, fmt.Errorf("unable to fetch all books, %v", fetchingError)
	}
	defer rows.Close() // close the connection back to the pool when the function finishes

	for rows.Next() {
		// we create a new book instance to store the data of the current row
		var b domain.Book
		var formatStr string
		scanningRowError := rows.Scan(&b.ID, &b.Title, &b.Author, &b.FilePath, &formatStr, &b.TotalPages, &b.AddedAt)
		if scanningRowError != nil {
			return nil, fmt.Errorf("unable to scan the data from database row, %v", scanningRowError)
		}
		b.Format = domain.BookFormat(formatStr)
		bookList = append(bookList, &b)
	}

	return bookList, nil
}

func (s *Storage) Delete(ctx context.Context, bookId string) error {
	// first of all, we manage the context lifecycle
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// now let's build the query to delete the entry
	query := `DELETE FROM books WHERE id=?`

	_, queryExecutionError := s.db.ExecContext(ctx, query, bookId)
	if queryExecutionError != nil {
		return fmt.Errorf("unable to delete book, %v", queryExecutionError)
	}

	return nil
}
