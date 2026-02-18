package port

import (
	"context"

	"github.com/MiltonJ23/Orus/internal/domain"
)

// BookRepository defines the contract to implement for persistence operations related to Books.
type BookRepository interface {
	Save(ctx context.Context, book *domain.Book) error
	GetByID(ctx context.Context, id string) (*domain.Book, error)
	ListAll(ctx context.Context) ([]*domain.Book, error)
	Delete(ctx context.Context, bookId string) error
}

// SessionRepository defines the contract to implement for persistence operations related to Reading Sessions.
type SessionRepository interface {
	SaveSession(ctx context.Context, session *domain.ReadingSession) error
	GetSessionByID(ctx context.Context, bookID string) (*domain.ReadingSession, error)
}

type AnnotationRepository interface {
	SaveAnnotation(ctx context.Context, annotation *domain.Annotation) error
	GetAnnotationByPage(ctx context.Context, pageNo int, book_id string) ([]*domain.Annotation, error)
	GetAnnotationByType(ctx context.Context, annotationType string) ([]*domain.Annotation, error)
	DeleteAnnotation(ctx context.Context, id string) error
	ListAllAnnotationOfABook(ctx context.Context, book_id string) ([]*domain.Annotation, error)
}
