package knowledge

import (
	"context"

	"github.com/husky/husky/internal/model"
)

type Service interface {
	CreateKnowledge(ctx context.Context, req *model.KnowledgeBase) (*model.KnowledgeBase, error)
	SearchKnowledge(ctx context.Context, query model.KnowledgeQuery) ([]model.KnowledgeResponse, error)
	GetKnowledge(ctx context.Context, id string) (*model.KnowledgeBase, error)
	ListKnowledge(ctx context.Context, offset, limit int, category string) ([]model.KnowledgeBase, int64, error)
	UpdateKnowledge(ctx context.Context, kb *model.KnowledgeBase) error
	DeleteKnowledge(ctx context.Context, id string) error
	ImportKnowledge(ctx context.Context, items []model.KnowledgeBase) (int, error)
	ExportKnowledge(ctx context.Context) ([]model.KnowledgeBase, error)
	ListCategories(ctx context.Context) ([]model.Category, error)
	CategoryTree(ctx context.Context) ([]model.CategoryTreeNode, error)
	Ask(ctx context.Context, question string, model string, deepThinking bool, images []string) (*model.AnswerResponse, error)
	RecordView(ctx context.Context, id string) error
	RecommendKnowledge(ctx context.Context, limit int) ([]model.KnowledgeHotResponse, error)
	RecordFeedback(ctx context.Context, knowledgeID string, userID uint, helpful bool, comment string) error
	GetFeedbackStats(ctx context.Context, knowledgeID string) (*model.KnowledgeFeedbackStats, error)
}
