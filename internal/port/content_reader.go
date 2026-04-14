package port

import "context"

// ContentReader defines the contract for extracting readable text from book files.
type ContentReader interface {
	// ReadBookText extracts all text from a book and splits it into
	// reasonably sized chunks. Each chunk represents a "page" in the reader.
	ReadBookText(ctx context.Context, filePath string) ([]string, error)
}
