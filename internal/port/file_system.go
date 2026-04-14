package port

import (
	"context"

	"github.com/MiltonJ23/Orus/internal/domain"
)

// MetadataExtractor defines the contract for extracting metadata from book files.
type MetadataExtractor interface {
	ExtractInfo(ctx context.Context, filePath string) (*domain.BookMetadata, error)
}
