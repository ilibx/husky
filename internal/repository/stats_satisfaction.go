package repository

import (
	"context"
	"time"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// GetStatsOverview 获取工单统计概览
func (r *TicketRepository) GetStatsOverview(ctx context.Context) (map[string]interface{}, error) {
	var total int64
	r.db.WithContext(ctx).Model(&model.Ticket{}).Count(&total)

	var openCount int64
	r.db.WithContext(ctx).Model(&model.Ticket{}).Where("status = ?", "open").Count(&openCount)

	var inProgressCount int64
	r.db.WithContext(ctx).Model(&model.Ticket{}).Where("status = ?", "in_progress").Count(&inProgressCount)

	var resolvedCount int64
	r.db.WithContext(ctx).Model(&model.Ticket{}).Where("status = ?", "resolved").Count(&resolvedCount)

	var closedCount int64
	r.db.WithContext(ctx).Model(&model.Ticket{}).Where("status = ?", "closed").Count(&closedCount)

	var overdueCount int64
	r.db.WithContext(ctx).Model(&model.Ticket{}).Where("due_at IS NOT NULL AND due_at < NOW() AND status NOT IN ('resolved', 'closed')").Count(&overdueCount)

	return map[string]interface{}{
		"total":       total,
		"open":        openCount,
		"in_progress": inProgressCount,
		"resolved":    resolvedCount,
		"closed":      closedCount,
		"overdue":     overdueCount,
	}, nil
}

// GetTicketStats 获取工单按渠道/优先级分布统计
func (r *TicketRepository) GetTicketStats(ctx context.Context) (map[string]interface{}, error) {
	type SourceCount struct {
		Source string
		Count  int64
	}
	var bySource []SourceCount
	r.db.WithContext(ctx).Model(&model.Ticket{}).Select("source, count(*) as count").Group("source").Find(&bySource)

	type PriorityCount struct {
		Priority string
		Count    int64
	}
	var byPriority []PriorityCount
	r.db.WithContext(ctx).Model(&model.Ticket{}).Select("priority, count(*) as count").Group("priority").Find(&byPriority)

	sourceMap := make(map[string]int64)
	for _, s := range bySource {
		sourceMap[s.Source] = s.Count
	}
	priorityMap := make(map[string]int64)
	for _, p := range byPriority {
		priorityMap[p.Priority] = p.Count
	}

	return map[string]interface{}{
		"by_source":   sourceMap,
		"by_priority": priorityMap,
	}, nil
}

// CreateSatisfaction 创建满意度评价
func (r *TicketRepository) CreateSatisfaction(ctx context.Context, s *model.Satisfaction) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// GetSatisfactionByTicket 获取工单满意度评价
func (r *TicketRepository) GetSatisfactionByTicket(ctx context.Context, ticketID uint) (*model.Satisfaction, error) {
	var s model.Satisfaction
	err := r.db.WithContext(ctx).Preload("User").Where("ticket_id = ?", ticketID).First(&s).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// GetSatisfactionStats 获取满意度统计
func (r *TicketRepository) GetSatisfactionStats(ctx context.Context) (map[string]interface{}, error) {
	var total int64
	r.db.WithContext(ctx).Model(&model.Satisfaction{}).Count(&total)

	var avg float64
	r.db.WithContext(ctx).Model(&model.Satisfaction{}).Select("COALESCE(AVG(score), 0)").Scan(&avg)

	var distribution []struct {
		Score int
		Count int64
	}
	r.db.WithContext(ctx).Model(&model.Satisfaction{}).Select("score, count(*) as count").Group("score").Order("score").Find(&distribution)

	distMap := make(map[int]int64)
	for _, d := range distribution {
		distMap[d.Score] = d.Count
	}

	return map[string]interface{}{
		"total":        total,
		"average":      avg,
		"distribution": distMap,
	}, nil
}

// AgentPerformance 客服绩效统计
type AgentPerformance struct {
	UserID           uint    `json:"user_id"`
	Username         string  `json:"username"`
	DisplayName      string  `json:"display_name"`
	ResolvedCount    int64   `json:"resolved_count"`
	AvgResolutionHrs float64 `json:"avg_resolution_hrs"`
	AvgSatisfaction  float64 `json:"avg_satisfaction"`
	AssignedCount    int64   `json:"assigned_count"`
}

// GetAgentPerformance 获取客服绩效统计
func (r *TicketRepository) GetAgentPerformance(ctx context.Context) ([]AgentPerformance, error) {
	var results []AgentPerformance
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			u.id AS user_id,
			u.username,
			u.display_name,
			COALESCE(resolved.resolved_count, 0) AS resolved_count,
			COALESCE(resolved.avg_hrs, 0) AS avg_resolution_hrs,
			COALESCE(sat.avg_score, 0) AS avg_satisfaction,
			COALESCE(assigned.assigned_count, 0) AS assigned_count
		FROM users u
		LEFT JOIN (
			SELECT assignee_id,
				COUNT(*) AS resolved_count,
				AVG(EXTRACT(EPOCH FROM (COALESCE(resolved_at, closed_at, updated_at) - created_at)) / 3600) AS avg_hrs
			FROM tickets
			WHERE status IN ('resolved', 'closed')
			GROUP BY assignee_id
		) resolved ON resolved.assignee_id = u.id
		LEFT JOIN (
			SELECT s.created_by,
				AVG(s.score) AS avg_score
			FROM satisfactions s
			GROUP BY s.created_by
		) sat ON sat.created_by = u.id
		LEFT JOIN (
			SELECT assignee_id,
				COUNT(*) AS assigned_count
			FROM tickets
			WHERE assignee_id IS NOT NULL AND status NOT IN ('resolved', 'closed')
			GROUP BY assignee_id
		) assigned ON assigned.assignee_id = u.id
		WHERE u.role IN ('agent', 'admin')
		ORDER BY resolved_count DESC
	`).Scan(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// SetTicketDueAt 设置工单 SLA 截止时间
func (r *TicketRepository) SetTicketDueAt(ctx context.Context, id uint, dueAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Ticket{}).Where("id = ?", id).Update("due_at", dueAt).Error
}
