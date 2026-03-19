package repository

import (
	"context"
	"errors"
	"time"

	"backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrNoDirectParent          = errors.New("no direct parent")
	ErrRechargeRequestNotFound = errors.New("recharge request not found")
	ErrRechargeRequestHandled  = errors.New("recharge request already handled")
	ErrRechargeApprovalDenied  = errors.New("recharge approval denied")
)

type RechargeRequestRepository interface {
	Create(ctx context.Context, request *model.RechargeRequest) error
	GetByID(ctx context.Context, requestID uint) (*model.RechargeRequest, error)
	GetByIDForUpdate(ctx context.Context, requestID uint) (*model.RechargeRequest, error)
	UpdateStatus(ctx context.Context, requestID uint, status int, reviewedBy uint, reviewedAt time.Time) error
}

type rechargeRequestRepository struct {
	db *gorm.DB
}

func NewRechargeRequestRepository(db *gorm.DB) RechargeRequestRepository {
	return &rechargeRequestRepository{db: db}
}

func (r *rechargeRequestRepository) Create(ctx context.Context, request *model.RechargeRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}

func (r *rechargeRequestRepository) GetByID(ctx context.Context, requestID uint) (*model.RechargeRequest, error) {
	var request model.RechargeRequest
	err := r.db.WithContext(ctx).Take(&request, requestID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *rechargeRequestRepository) GetByIDForUpdate(ctx context.Context, requestID uint) (*model.RechargeRequest, error) {
	var request model.RechargeRequest
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Take(&request, requestID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *rechargeRequestRepository) UpdateStatus(
	ctx context.Context,
	requestID uint,
	status int,
	reviewedBy uint,
	reviewedAt time.Time,
) error {
	return r.db.WithContext(ctx).
		Model(&model.RechargeRequest{}).
		Where("id = ?", requestID).
		Updates(map[string]any{
			"status":      status,
			"reviewed_by": reviewedBy,
			"reviewed_at": &reviewedAt,
		}).Error
}
