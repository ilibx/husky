package repository

import (
	"context"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

// AddWatcher 添加工单关注者
func (r *TicketRepository) AddWatcher(ctx context.Context, ticketID, userID uint) error {
	return r.db.WithContext(ctx).FirstOrCreate(&model.TicketWatcher{}, map[string]interface{}{
		"ticket_id": ticketID,
		"user_id":   userID,
	}).Error
}

// RemoveWatcher 移除工单关注者
func (r *TicketRepository) RemoveWatcher(ctx context.Context, ticketID, userID uint) error {
	return r.db.WithContext(ctx).Where("ticket_id = ? AND user_id = ?", ticketID, userID).
		Delete(&model.TicketWatcher{}).Error
}

// GetWatchers 获取工单关注者列表
func (r *TicketRepository) GetWatchers(ctx context.Context, ticketID uint) ([]model.TicketWatcher, error) {
	var watchers []model.TicketWatcher
	err := r.db.WithContext(ctx).Preload("User").
		Where("ticket_id = ?", ticketID).Find(&watchers).Error
	return watchers, err
}

// FindLeastBusyAgent 查找当前工单最少的客服
func (r *TicketRepository) FindLeastBusyAgent(ctx context.Context) (*uint, error) {
	type agentLoad struct {
		AssigneeID uint
		Count      int64
	}
	var result agentLoad
	err := r.db.WithContext(ctx).Model(&model.Ticket{}).
		Select("assignee_id, COUNT(*) as count").
		Where("assignee_id IS NOT NULL AND status NOT IN ('resolved', 'closed')").
		Group("assignee_id").
		Order("count ASC").
		Limit(1).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	if result.AssigneeID == 0 {
		return nil, nil
	}
	return &result.AssigneeID, nil
}

// ListAgentIDs 获取所有活跃客服/管理员ID
func (r *TicketRepository) ListAgentIDs(ctx context.Context) ([]uint, error) {
	var ids []uint
	err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("role IN ? AND status = 1", []string{"admin", "agent"}).
		Pluck("id", &ids).Error
	return ids, err
}

// CountAgentLoad 统计客服当前进行中工单数
func (r *TicketRepository) CountAgentLoad(ctx context.Context, agentID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Ticket{}).
		Where("assignee_id = ? AND status NOT IN ('resolved', 'closed')", agentID).
		Count(&count).Error
	return count, err
}

// FindAgentsBySkill 根据技能匹配客服（分类名模糊匹配）
func (r *TicketRepository) FindAgentsBySkill(ctx context.Context, categoryID uint) ([]uint, error) {
	var cat model.Category
	if err := r.db.WithContext(ctx).First(&cat, categoryID).Error; err != nil {
		return nil, err
	}
	var ids []uint
	err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("role IN ? AND status = 1 AND skills ILIKE ?", []string{"admin", "agent"}, "%"+cat.Name+"%").
		Pluck("id", &ids).Error
	return ids, err
}

// IncrementRoundRobin 递增轮询并返回新索引
func (r *TicketRepository) IncrementRoundRobin(ctx context.Context, cfgID uint, maxIndex int) (int, error) {
	err := r.db.WithContext(ctx).Model(&model.AssignConfig{}).
		Where("id = ?", cfgID).
		UpdateColumn("round_robin_index", gorm.Expr("(round_robin_index + 1) % ?", maxIndex)).
		Error
	if err != nil {
		return 0, err
	}
	var cfg model.AssignConfig
	r.db.WithContext(ctx).First(&cfg, cfgID)
	return cfg.RoundRobinIndex, nil
}
