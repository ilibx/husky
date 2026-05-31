package vectorstore

import "context"

type SearchResult struct {
	ID    string
	Score float32
}

type Store interface {
	SearchSimilar(ctx context.Context, vector []float32, limit int) ([]SearchResult, error)
	InsertVector(ctx context.Context, id string, vector []float32) error
	DeleteVector(ctx context.Context, id string) error
	UpdateVector(ctx context.Context, id string, vector []float32) error
	Close() error
}
