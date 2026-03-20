package port

import (
	"context"

	"github.com/MiltonJ23/Orus/internal/domain"
)

type MetadataExtractor interface {
	ExtractInfo(ctx context.Context, file_path string) (*domain.BookMetadata, error)
}
