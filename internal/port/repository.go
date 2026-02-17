package port

import (
	"context"

	"github.com/MiltonJ23/Orus/internal/domain"
)

// BookRepository defines the contract to implement for persistence operations related to Books.
type BookRepository interface {
	Save(ctx context.Context, book *domain.Book) error
	GetByID(ctx context.Context, id int) (*domain.Book, error)
	ListAll(ctx context.Context) ([]*domain.Book, error)
	Delete(ctx context.Context, id string) error
}

// SessionRepository defines the contract to implement for persistence operations related to Reading Sessions.
type SessionRepository interface {
	Save(ctx context.Context, session *domain.ReadingSession) error
	GetByID(ctx context.Context, bookID string) (*domain.ReadingSession, error)
}
