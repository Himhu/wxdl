package repository

import (
	"context"
	"errors"
	"strings"

	"backend/internal/model"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AgentListQuery struct {
	ParentID uint
	Page     int
	PageSize int
	Status   *int
	Keyword  string
}

type AdminAgentListQuery struct {
	Page     int
	PageSize int
	Status   *int
	Keyword  string
}

type AgentRepository interface {
	Create(ctx context.Context, agent *model.Agent) error
	GetByID(ctx context.Context, id uint) (*model.Agent, error)
	GetByIDForUpdate(ctx context.Context, id uint) (*model.Agent, error)
	GetByUsername(ctx context.Context, username string) (*model.Agent, error)
	GetByWechatOpenID(ctx context.Context, openID string) (*model.Agent, error)
	GetDirectChildByID(ctx context.Context, parentID, agentID uint) (*model.Agent, error)
	ListDirectChildren(ctx context.Context, query AgentListQuery) ([]model.Agent, int64, error)
	ListAll(ctx context.Context, query AdminAgentListQuery) ([]model.Agent, int64, error)
	Update(ctx context.Context, agent *model.Agent) error
	UpdateStatus(ctx context.Context, agentID uint, status int) error
	UpdateBalance(ctx context.Context, agentID uint, balance decimal.Decimal) error
	UpdateLevel(ctx context.Context, agentID uint, level int) error
	BindWechat(ctx context.Context, agentID uint, openID, unionID string) error
	CreateLoginLog(ctx context.Context, log *model.LoginLog) error
	ListLoginLogs(ctx context.Context, agentID uint, page, pageSize int) ([]model.LoginLog, int64, error)
}

type agentRepository struct {
	db *gorm.DB
}

func NewAgentRepository(db *gorm.DB) AgentRepository {
	return &agentRepository{db: db}
}

func (r *agentRepository) Create(ctx context.Context, agent *model.Agent) error {
	return r.db.WithContext(ctx).Create(agent).Error
}

func (r *agentRepository) GetByID(ctx context.Context, id uint) (*model.Agent, error) {
	var agent model.Agent
	err := r.db.WithContext(ctx).Take(&agent, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &agent, nil
}

func (r *agentRepository) GetByIDForUpdate(ctx context.Context, id uint) (*model.Agent, error) {
	var agent model.Agent
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Take(&agent, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *agentRepository) GetByUsername(ctx context.Context, username string) (*model.Agent, error) {
	var agent model.Agent
	err := r.db.WithContext(ctx).Where("username = ?", username).Take(&agent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &agent, nil
}

func (r *agentRepository) GetByWechatOpenID(ctx context.Context, openID string) (*model.Agent, error) {
	var agent model.Agent
	err := r.db.WithContext(ctx).Where("wechat_open_id = ?", openID).Take(&agent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &agent, nil
}

func (r *agentRepository) GetDirectChildByID(ctx context.Context, parentID, agentID uint) (*model.Agent, error) {
	var agent model.Agent
	err := r.db.WithContext(ctx).
		Where("id = ? AND parent_id = ?", agentID, parentID).
		Take(&agent).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &agent, nil
}

func (r *agentRepository) ListDirectChildren(ctx context.Context, query AgentListQuery) ([]model.Agent, int64, error) {
	base := r.db.WithContext(ctx).Model(&model.Agent{}).Where("parent_id = ?", query.ParentID)
	if query.Status != nil {
		base = base.Where("status = ?", *query.Status)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		base = base.Where("username LIKE ? OR real_name LIKE ? OR phone LIKE ?", like, like, like)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var agents []model.Agent
	if err := base.Order("id DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&agents).Error; err != nil {
		return nil, 0, err
	}

	return agents, total, nil
}

func (r *agentRepository) ListAll(ctx context.Context, query AdminAgentListQuery) ([]model.Agent, int64, error) {
	base := r.db.WithContext(ctx).Model(&model.Agent{})
	if query.Status != nil {
		base = base.Where("status = ?", *query.Status)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		base = base.Where("username LIKE ? OR real_name LIKE ? OR phone LIKE ? OR wechat_open_id LIKE ?", like, like, like, like)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var agents []model.Agent
	if err := base.Order("id DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&agents).Error; err != nil {
		return nil, 0, err
	}

	return agents, total, nil
}

func (r *agentRepository) Update(ctx context.Context, agent *model.Agent) error {
	return r.db.WithContext(ctx).Save(agent).Error
}

func (r *agentRepository) UpdateStatus(ctx context.Context, agentID uint, status int) error {
	return r.db.WithContext(ctx).
		Model(&model.Agent{}).
		Where("id = ?", agentID).
		Update("status", status).Error
}

func (r *agentRepository) UpdateBalance(ctx context.Context, agentID uint, balance decimal.Decimal) error {
	return r.db.WithContext(ctx).
		Model(&model.Agent{}).
		Where("id = ?", agentID).
		Update("balance", balance).Error
}

func (r *agentRepository) UpdateLevel(ctx context.Context, agentID uint, level int) error {
	return r.db.WithContext(ctx).
		Model(&model.Agent{}).
		Where("id = ?", agentID).
		Update("level", level).Error
}

func (r *agentRepository) BindWechat(ctx context.Context, agentID uint, openID, unionID string) error {
	return r.db.WithContext(ctx).
		Model(&model.Agent{}).
		Where("id = ?", agentID).
		Updates(map[string]any{
			"wechat_open_id":  openID,
			"wechat_union_id": unionID,
		}).Error
}

func (r *agentRepository) CreateLoginLog(ctx context.Context, log *model.LoginLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *agentRepository) ListLoginLogs(ctx context.Context, agentID uint, page, pageSize int) ([]model.LoginLog, int64, error) {
	base := r.db.WithContext(ctx).Model(&model.LoginLog{}).Where("agent_id = ?", agentID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []model.LoginLog
	if err := base.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
