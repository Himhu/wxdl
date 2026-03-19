package repository

import (
	"context"

	"backend/internal/model"
	"gorm.io/gorm"
)

type PointsRecordRepository interface {
	Create(ctx context.Context, record *model.PointsRecord) error
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
