package port

import (
	"context"

	"github.com/MiltonJ23/Orus/internal/domain"
)

// BookRepository defines the contract to implement for book persistence operations.
type BookRepository interface {
	Save(ctx context.Context, book *domain.Book) error
	GetByID(ctx context.Context, id string) (*domain.Book, error)
	ListAll(ctx context.Context) ([]*domain.Book, error)
	Delete(ctx context.Context, bookId string) error
}

// SessionRepository defines the contract for reading session persistence.
type SessionRepository interface {
	SaveSession(ctx context.Context, session *domain.ReadingSession) error
	GetSessionByID(ctx context.Context, bookID string) ([]*domain.ReadingSession, error)
	GetLastReadingSession(ctx context.Context, bookID string) (*domain.ReadingSession, error)
}

// AnnotationRepository defines the contract for annotation persistence.
type AnnotationRepository interface {
	SaveAnnotation(ctx context.Context, annotation *domain.Annotation) error
	GetAnnotationByPage(ctx context.Context, pageNo int, bookID string) ([]*domain.Annotation, error)
	GetAnnotationByType(ctx context.Context, annotationType string) ([]*domain.Annotation, error)
	DeleteAnnotation(ctx context.Context, id string) error
	ListAllAnnotationOfABook(ctx context.Context, bookID string) ([]*domain.Annotation, error)
}
