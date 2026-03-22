package repository

import (
	"context"

	"backend/internal/model"
	"gorm.io/gorm"
)

type PointsRecordRepository interface {
	Create(ctx context.Context, record *model.PointsRecord) error
	ExistsByDescription(ctx context.Context, agentID uint, description string) (bool, error)
}

type pointsRecordRepository struct {
	db *gorm.DB
}

func NewPointsRecordRepository(db *gorm.DB) PointsRecordRepository {
	return &pointsRecordRepository{db: db}
}

func (r *pointsRecordRepository) Create(ctx context.Context, record *model.PointsRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *pointsRecordRepository) ExistsByDescription(ctx context.Context, agentID uint, description string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.PointsRecord{}).
		Where("agent_id = ? AND remark = ?", agentID, description).
		Count(&count).Error
	return count > 0, err
}
