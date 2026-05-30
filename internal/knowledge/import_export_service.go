package knowledge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/husky/husky/internal/model"
)

func (s *service) ImportKnowledge(ctx context.Context, items []model.KnowledgeBase) (int, error) {
	if len(items) == 0 {
		return 0, fmt.Errorf("no items to import")
	}
	count := 0
	for i := range items {
		if items[i].Title == "" {
			continue
		}
		if items[i].Status == "" {
			items[i].Status = "active"
		}

		if s.embedding != nil {
			text := items[i].Title + "\n" + items[i].Content
			vec, err := s.embedding.Embed(ctx, text)
			if err != nil {
				return count, fmt.Errorf("failed to generate embedding for item %d: %w", i, err)
			}
			vecJSON, err := json.Marshal(vec)
			if err != nil {
				return count, fmt.Errorf("failed to marshal embedding for item %d: %w", i, err)
			}
			items[i].Vector = vecJSON
		}

		if err := s.vectorRepo.InsertKnowledge(ctx, &items[i]); err != nil {
			return count, fmt.Errorf("failed to import item %d: %w", i, err)
		}
		count++
	}
	return count, nil
}

func (s *service) ExportKnowledge(ctx context.Context) ([]model.KnowledgeBase, error) {
	var all []model.KnowledgeBase
	offset := 0
	pageSize := 1000
	for {
		list, _, err := s.vectorRepo.ListKnowledge(ctx, offset, pageSize, "")
		if err != nil {
			return nil, fmt.Errorf("failed to export knowledge: %w", err)
		}
		all = append(all, list...)
		if len(list) < pageSize {
			break
		}
		offset += pageSize
	}
	return all, nil
}
