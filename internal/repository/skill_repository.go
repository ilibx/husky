package repository

import (
	"context"

	"github.com/husky/husky/internal/model"
	"gorm.io/gorm"
)

type SkillRepository struct {
	*BaseRepository
}

func NewSkillRepository(db *gorm.DB) *SkillRepository {
	return &SkillRepository{BaseRepository: NewBaseRepository(db)}
}

func (r *SkillRepository) Create(ctx context.Context, skill *model.Skill) error {
	return r.db.WithContext(ctx).Create(skill).Error
}

func (r *SkillRepository) GetByID(ctx context.Context, id uint) (*model.Skill, error) {
	var s model.Skill
	err := r.db.WithContext(ctx).Preload("Creator").First(&s, id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SkillRepository) List(ctx context.Context, offset, limit int) ([]model.Skill, int64, error) {
	var list []model.Skill
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Skill{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *SkillRepository) Update(ctx context.Context, skill *model.Skill) error {
	return r.db.WithContext(ctx).Save(skill).Error
}

func (r *SkillRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Skill{}, id).Error
}
