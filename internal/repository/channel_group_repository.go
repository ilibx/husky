package repository

import (
	"context"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

type ChannelGroupRepository struct {
	*BaseRepository
}

func NewChannelGroupRepository(db *gorm.DB) *ChannelGroupRepository {
	return &ChannelGroupRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *ChannelGroupRepository) Create(ctx context.Context, group *model.ChannelGroup) error {
	return r.db.WithContext(ctx).Create(group).Error
}

func (r *ChannelGroupRepository) GetByID(ctx context.Context, id uint) (*model.ChannelGroup, error) {
	var group model.ChannelGroup
	err := r.db.WithContext(ctx).Preload("Tags").First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *ChannelGroupRepository) List(ctx context.Context, channelType string, offset, limit int) ([]model.ChannelGroup, int64, error) {
	var list []model.ChannelGroup
	var total int64
	q := r.db.WithContext(ctx).Model(&model.ChannelGroup{})
	if channelType != "" {
		q = q.Where("channel_type = ?", channelType)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Tags").Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *ChannelGroupRepository) Update(ctx context.Context, group *model.ChannelGroup) error {
	return r.db.WithContext(ctx).Save(group).Error
}

func (r *ChannelGroupRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.ChannelGroup{}, id).Error
}

func (r *ChannelGroupRepository) UpdateTags(ctx context.Context, id uint, tagIDs []uint) error {
	group, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	var tags []model.Tag
	if len(tagIDs) > 0 {
		if err := r.db.WithContext(ctx).Find(&tags, tagIDs).Error; err != nil {
			return err
		}
	}
	return r.db.WithContext(ctx).Model(group).Association("Tags").Replace(tags)
}
