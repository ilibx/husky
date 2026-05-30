package knowledge

import (
	"context"
	"fmt"

	"github.com/husky/husky/internal/model"
)

func (s *service) RecordFeedback(ctx context.Context, knowledgeID string, userID uint, helpful bool, comment string) error {
	kb, err := s.vectorRepo.GetKnowledge(ctx, knowledgeID)
	if err != nil {
		return fmt.Errorf("failed to get knowledge: %w", err)
	}
	if kb == nil {
		return fmt.Errorf("knowledge not found")
	}

	exists, err := s.vectorRepo.HasUserFeedbacked(ctx, knowledgeID, userID)
	if err != nil {
		return fmt.Errorf("failed to check existing feedback: %w", err)
	}
	if exists {
		return fmt.Errorf("user already provided feedback for this article")
	}

	fb := &model.KnowledgeFeedback{
		KnowledgeID: knowledgeID,
		UserID:      userID,
		Helpful:     helpful,
		Comment:     comment,
	}
	return s.vectorRepo.RecordFeedback(ctx, fb)
}

func (s *service) GetFeedbackStats(ctx context.Context, knowledgeID string) (*model.KnowledgeFeedbackStats, error) {
	return s.vectorRepo.GetFeedbackStats(ctx, knowledgeID)
}
