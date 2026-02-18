package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/MiltonJ23/Orus/internal/domain"
	"github.com/MiltonJ23/Orus/internal/port"
)

var _ port.AnnotationRepository = (*Storage)(nil)

func (s *Storage) SaveAnnotation(ctx context.Context, annotation *domain.Annotation) error {
	// let's manage the context lifecycle
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// now let's build the query
	query := `INSERT INTO annotations (id,book_id,annotation_type,page_number,created_at) VALUES (?,?,?,?,?);`

	_, queryExecutionError := s.db.ExecContext(ctx, query, annotation.ID, annotation.BookID, annotation.AnnotationType, annotation.PageNo, annotation.CreatedAt)
	if queryExecutionError != nil {
		return fmt.Errorf("an error occured while inserting annotation into database: %v", queryExecutionError)
	}
	return nil
}

// GetAnnotationByPage will retrieve the annotations for the given page of a book
func (s *Storage) GetAnnotationByPage(ctx context.Context, pageNo int, book_id string) ([]*domain.Annotation, error) {
	// we manage the context lifecycle
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// let's build the query
	query := `SELECT * FROM annotations WHERE page_number = ? AND book_id=? ORDER BY page_number ASC;`

	rows, fetchingError := s.db.QueryContext(ctx, query, pageNo, book_id)
	if fetchingError != nil {
		return nil, fmt.Errorf("an error occured while querying annotations table: %v", fetchingError)
	}
	defer rows.Close()

	var annotations []*domain.Annotation

	for rows.Next() {
		var annot domain.Annotation
		var formatStr string

		scanningError := rows.Scan(&annot.ID, &annot.BookID, formatStr, &annot.PageNo, &annot.CreatedAt)
		if scanningError != nil {
			return nil, fmt.Errorf("an error occured while scanning annotations table: %v", scanningError)
		}
		annot.AnnotationType = domain.AnnotationType(formatStr)
		annotations = append(annotations, &annot)
	}
	// meaning all went well

	return annotations, nil
}

// GetAnnotationByType will retrieve the annotation based on their types
func (s *Storage) GetAnnotationByType(ctx context.Context, annotationType string) ([]*domain.Annotation, error) {
	// let's manage the context lifecycle
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// let's build the query
	query := `SELECT * FROM annotations WHERE annotation_type = ?;`

	rows, fetchingError := s.db.QueryContext(ctx, query, annotationType)
	if fetchingError != nil {
		return nil, fmt.Errorf("an error occured while querying annotations table: %v", fetchingError)
	}
	defer rows.Close()

	var annotations []*domain.Annotation

	for rows.Next() {
		var annot domain.Annotation
		var formatStr string

		scanningError := rows.Scan(&annot.ID, &annot.BookID, formatStr, &annot.PageNo, &annot.CreatedAt)
		if scanningError != nil {
			return nil, fmt.Errorf("an error occured while scanning annotations table: %v", scanningError)
		}
		annot.AnnotationType = domain.AnnotationType(formatStr)
		annotations = append(annotations, &annot)
	}
	return annotations, nil
}

// DeleteAnnotation will delete a specified annotation
func (s *Storage) DeleteAnnotation(ctx context.Context, id string) error {
	// let's manage the context lifecycle
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// we build the query
	query := `DELETE  FROM annotations WHERE id=?`

	_, queryExecutionError := s.db.ExecContext(ctx, query, id)
	if queryExecutionError != nil {
		return fmt.Errorf("an error occured while deleting annotation: %v", queryExecutionError)
	}
	return nil
}

// ListAllAnnotationOfABook will retrieve all of the annotations for a given book
func (s *Storage) ListAllAnnotationOfABook(ctx context.Context, book_id string) ([]*domain.Annotation, error) {
	// let's manage the context lifecycle
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT * FROM annotations WHERE book_id=?;`

	rows, fetchingError := s.db.QueryContext(ctx, query, book_id)
	if fetchingError != nil {
		return nil, fmt.Errorf("an error occured while querying annotations table: %v", fetchingError)
	}
	defer rows.Close()

	var annotations []*domain.Annotation

	for rows.Next() {
		var annot domain.Annotation
		var formatStr string

		scanningError := rows.Scan(&annot.ID, &annot.BookID, formatStr, &annot.PageNo, &annot.CreatedAt)
		if scanningError != nil {
			return nil, fmt.Errorf("an error occured while scanning annotations rows pointer: %v", scanningError)
		}
		annot.AnnotationType = domain.AnnotationType(formatStr)
		annotations = append(annotations, &annot)
	}
	return annotations, nil

}
