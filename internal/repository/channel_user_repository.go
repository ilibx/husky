package repository

import (
	"context"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

type ChannelUserRepository struct {
	*BaseRepository
}

func NewChannelUserRepository(db *gorm.DB) *ChannelUserRepository {
	return &ChannelUserRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *ChannelUserRepository) Create(ctx context.Context, user *model.ChannelUser) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *ChannelUserRepository) GetByID(ctx context.Context, id uint) (*model.ChannelUser, error) {
	var user model.ChannelUser
	err := r.db.WithContext(ctx).Preload("Tags").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *ChannelUserRepository) List(ctx context.Context, channelType string, offset, limit int) ([]model.ChannelUser, int64, error) {
	var list []model.ChannelUser
	var total int64
	q := r.db.WithContext(ctx).Model(&model.ChannelUser{})
	if channelType != "" {
		q = q.Where("channel_type = ?", channelType)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Tags").Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *ChannelUserRepository) Update(ctx context.Context, user *model.ChannelUser) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *ChannelUserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.ChannelUser{}, id).Error
}

func (r *ChannelUserRepository) UpdateTags(ctx context.Context, id uint, tagIDs []uint) error {
	user, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}
	var tags []model.Tag
	if len(tagIDs) > 0 {
		if err := r.db.WithContext(ctx).Find(&tags, tagIDs).Error; err != nil {
			return err
		}
	}
	return r.db.WithContext(ctx).Model(user).Association("Tags").Replace(tags)
}
