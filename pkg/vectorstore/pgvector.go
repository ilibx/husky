package vectorstore

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type pgVectorStore struct {
	db *gorm.DB
}

func NewPGVectorStore(db *gorm.DB) Store {
	return &pgVectorStore{db: db}
}

func (s *pgVectorStore) SearchSimilar(ctx context.Context, queryVector []float32, limit int) ([]SearchResult, error) {
	if len(queryVector) == 0 {
		return nil, nil
	}

	vectorJSON, _ := json.Marshal(queryVector)
	rawSQL := `SELECT id, vector <-> ?::vector AS score
			   FROM knowledge
			   WHERE status = 'active'
			   ORDER BY score LIMIT ?`

	type scanResult struct {
		ID    string  `gorm:"column:id"`
		Score float32 `gorm:"column:score"`
	}

	var results []scanResult
	if err := s.db.WithContext(ctx).Raw(rawSQL, string(vectorJSON), limit).Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("pgvector search: %w", err)
	}

	out := make([]SearchResult, len(results))
	for i, r := range results {
		out[i] = SearchResult{ID: r.ID, Score: r.Score}
	}
	return out, nil
}

func (s *pgVectorStore) InsertVector(ctx context.Context, id string, vector []float32) error {
	vectorJSON, err := json.Marshal(vector)
	if err != nil {
		return fmt.Errorf("marshal vector: %w", err)
	}
	return s.db.WithContext(ctx).Model(&struct{}{}).
		Table("knowledge").
		Where("id = ?", id).
		UpdateColumn("vector", string(vectorJSON)).Error
}

func (s *pgVectorStore) DeleteVector(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Exec("UPDATE knowledge SET vector = NULL WHERE id = ?", id).Error
}

func (s *pgVectorStore) UpdateVector(ctx context.Context, id string, vector []float32) error {
	return s.InsertVector(ctx, id, vector)
}

func (s *pgVectorStore) Close() error { return nil }
