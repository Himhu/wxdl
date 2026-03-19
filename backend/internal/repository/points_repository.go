package repository

import (
	"context"
	"errors"
	"time"

	"backend/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type RechargeRequestQuery struct {
	AgentID  uint
	Page     int
	PageSize int
	Status   *int
}

type RechargeRequestItem struct {
	ID                uint            `json:"id"`
	AgentID           uint            `json:"agentId"`
	ApplicantUsername string          `json:"applicantUsername"`
	ApplicantRealName string          `json:"applicantRealName"`
	Amount            decimal.Decimal `json:"amount"`
	Status            int             `json:"status"`
	PaymentMethod     string          `json:"paymentMethod"`
	PaymentProof      string          `json:"paymentProof"`
	Remark            string          `json:"remark"`
	ReviewedBy        *uint           `json:"reviewedBy"`
	ReviewerUsername  string          `json:"reviewerUsername,omitempty"`
	ReviewedAt        *time.Time      `json:"reviewedAt"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

type PointsRecordQuery struct {
	AgentID  uint
	Page     int
	PageSize int
	Type     *int
}

type PointsStats struct {
	CurrentBalance decimal.Decimal `json:"currentBalance"`
	TotalRecords   int64           `json:"totalRecords"`
	RechargeCount  int64           `json:"rechargeCount"`
	ConsumeCount   int64           `json:"consumeCount"`
	RefundCount    int64           `json:"refundCount"`
	RechargeAmount decimal.Decimal `json:"rechargeAmount"`
	ConsumeAmount  decimal.Decimal `json:"consumeAmount"`
	RefundAmount   decimal.Decimal `json:"refundAmount"`
}

type PointsRepository interface {
	ListPendingRechargeRequests(ctx context.Context, query RechargeRequestQuery) ([]RechargeRequestItem, int64, error)
	ListRechargeHistory(ctx context.Context, query RechargeRequestQuery) ([]RechargeRequestItem, int64, error)
	ListRecords(ctx context.Context, query PointsRecordQuery) ([]model.PointsRecord, int64, error)
	GetStats(ctx context.Context, agentID uint) (*PointsStats, error)
}

type pointsRepository struct {
	db *gorm.DB
}

func NewPointsRepository(db *gorm.DB) PointsRepository {
	return &pointsRepository{db: db}
}

func (r *pointsRepository) ListPendingRechargeRequests(ctx context.Context, query RechargeRequestQuery) ([]RechargeRequestItem, int64, error) {
	base := r.rechargeListBase(ctx, query.AgentID).
		Where("applicant.parent_id = ?", query.AgentID).
		Where("rr.status = ?", model.RechargeRequestStatusPending)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []RechargeRequestItem
	if err := base.Select(rechargeRequestSelect()).
		Order("rr.id DESC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Scan(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *pointsRepository) ListRechargeHistory(ctx context.Context, query RechargeRequestQuery) ([]RechargeRequestItem, int64, error) {
	base := r.rechargeListBase(ctx, query.AgentID).
		Where("rr.agent_id = ? OR applicant.parent_id = ?", query.AgentID, query.AgentID)
	if query.Status != nil {
		base = base.Where("rr.status = ?", *query.Status)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []RechargeRequestItem
	if err := base.Select(rechargeRequestSelect()).
		Order("rr.id DESC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Scan(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *pointsRepository) ListRecords(ctx context.Context, query PointsRecordQuery) ([]model.PointsRecord, int64, error) {
	base := r.db.WithContext(ctx).Model(&model.PointsRecord{}).Where("agent_id = ?", query.AgentID)
	if query.Type != nil {
		base = base.Where("type = ?", *query.Type)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []model.PointsRecord
	if err := base.Order("id DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (r *pointsRepository) GetStats(ctx context.Context, agentID uint) (*PointsStats, error) {
	stats := &PointsStats{}

	var agent model.Agent
	if err := r.db.WithContext(ctx).Take(&agent, agentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAgentNotFound
		}
		return nil, err
	}
	stats.CurrentBalance = agent.Balance

	var rows []struct {
		Type        int
		TotalCount  int64
		TotalAmount decimal.Decimal
	}
	if err := r.db.WithContext(ctx).
		Model(&model.PointsRecord{}).
		Select("type, COUNT(*) AS total_count, CAST(COALESCE(SUM(amount), 0) AS DECIMAL(20,2)) AS total_amount").
		Where("agent_id = ?", agentID).
		Group("type").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		stats.TotalRecords += row.TotalCount
		switch row.Type {
		case model.PointsRecordTypeRecharge:
			stats.RechargeCount = row.TotalCount
			stats.RechargeAmount = row.TotalAmount
		case model.PointsRecordTypeConsume:
			stats.ConsumeCount = row.TotalCount
			stats.ConsumeAmount = row.TotalAmount
		case model.PointsRecordTypeRefund:
			stats.RefundCount = row.TotalCount
			stats.RefundAmount = row.TotalAmount
		}
	}

	return stats, nil
}

func (r *pointsRepository) rechargeListBase(ctx context.Context, agentID uint) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("recharge_requests AS rr").
		Joins("JOIN agents applicant ON applicant.id = rr.agent_id").
		Joins("LEFT JOIN agents reviewer ON reviewer.id = rr.reviewed_by")
}

func rechargeRequestSelect() string {
	return "rr.id, rr.agent_id, applicant.username AS applicant_username, applicant.real_name AS applicant_real_name, rr.amount, rr.status, rr.payment_method, rr.payment_proof, rr.remark, rr.reviewed_by, COALESCE(reviewer.username, '') AS reviewer_username, rr.reviewed_at, rr.created_at, rr.updated_at"
}
