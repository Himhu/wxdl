package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrAgentNotFound        = errors.New("agent not found")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrCardNotFound         = errors.New("card not found")
	ErrCardAlreadyUsed      = errors.New("card already used")
	ErrCardAlreadyDestroyed = errors.New("card already destroyed")
)

type CardQuery struct {
	AgentID  uint
	Page     int
	PageSize int
	Status   *int
	Keyword  string
}

type CardStats struct {
	Total     int64 `json:"total"`
	Unused    int64 `json:"unused"`
	Used      int64 `json:"used"`
	Destroyed int64 `json:"destroyed"`
}

type AdminCardQuery struct {
	AgentID  *uint
	Page     int
	PageSize int
	Status   *int
	Keyword  string
}

type AdminCardListItem struct {
	model.Card
	AgentUsername string `json:"agentUsername"`
	AgentRealName string `json:"agentRealName"`
}

type CardRepository interface {
	ListByAgent(ctx context.Context, query CardQuery) ([]model.Card, int64, error)
	ListAll(ctx context.Context, query AdminCardQuery) ([]AdminCardListItem, int64, error)
	GetByID(ctx context.Context, agentID, cardID uint) (*model.Card, error)
	GetByIDForUpdate(ctx context.Context, agentID, cardID uint) (*model.Card, error)
	MarkDestroyed(ctx context.Context, cardID uint, destroyedAt time.Time) error
	SyncStatusByCardKey(ctx context.Context, cardKey string, status int, usedAt *time.Time) (bool, error)
	ListUnusedCardKeys(ctx context.Context) ([]string, error)
	GetStats(ctx context.Context, agentID uint) (*CardStats, error)
	Create(ctx context.Context, card *model.Card) error
	CreatePointsRecord(ctx context.Context, record *model.PointsRecord) error
}

type cardRepository struct {
	db *gorm.DB
}

func NewCardRepository(db *gorm.DB) CardRepository {
	return &cardRepository{db: db}
}

func (r *cardRepository) ListByAgent(ctx context.Context, query CardQuery) ([]model.Card, int64, error) {
	base := r.db.WithContext(ctx).Model(&model.Card{}).Where("agent_id = ?", query.AgentID)
	if query.Status != nil {
		base = base.Where("status = ?", *query.Status)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		base = base.Where("card_key LIKE ? OR used_by LIKE ?", like, like)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var cards []model.Card
	if err := base.Order("id DESC").Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&cards).Error; err != nil {
		return nil, 0, err
	}

	return cards, total, nil
}

func (r *cardRepository) ListAll(ctx context.Context, query AdminCardQuery) ([]AdminCardListItem, int64, error) {
	base := r.db.WithContext(ctx).
		Table("cards").
		Joins("LEFT JOIN agents ON agents.id = cards.agent_id")

	if query.AgentID != nil {
		base = base.Where("cards.agent_id = ?", *query.AgentID)
	}
	if query.Status != nil {
		base = base.Where("cards.status = ?", *query.Status)
	}
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		base = base.Where("cards.card_key LIKE ?", "%"+keyword+"%")
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []AdminCardListItem
	if err := base.
		Select("cards.*, agents.username AS agent_username, agents.real_name AS agent_real_name").
		Order("cards.id DESC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Scan(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *cardRepository) GetByID(ctx context.Context, agentID, cardID uint) (*model.Card, error) {
	var card model.Card
	err := r.db.WithContext(ctx).
		Where("id = ? AND agent_id = ?", cardID, agentID).
		Take(&card).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &card, nil
}

func (r *cardRepository) GetByIDForUpdate(ctx context.Context, agentID, cardID uint) (*model.Card, error) {
	var card model.Card
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND agent_id = ?", cardID, agentID).
		Take(&card).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &card, nil
}

func (r *cardRepository) MarkDestroyed(ctx context.Context, cardID uint, destroyedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&model.Card{}).
		Where("id = ?", cardID).
		Updates(map[string]any{
			"status":       model.CardStatusDestroyed,
			"destroyed_at": &destroyedAt,
		}).Error
}

func (r *cardRepository) SyncStatusByCardKey(ctx context.Context, cardKey string, status int, usedAt *time.Time) (bool, error) {
	updates := map[string]any{"status": status}
	if usedAt != nil {
		updates["used_at"] = usedAt
	}

	db := r.db.WithContext(ctx).Model(&model.Card{}).
		Where("card_key = ? AND status = ?", cardKey, model.CardStatusUnused)

	if status != model.CardStatusUsed && status != model.CardStatusDestroyed {
		return false, nil
	}

	result := db.Updates(updates)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *cardRepository) ListUnusedCardKeys(ctx context.Context) ([]string, error) {
	var keys []string
	err := r.db.WithContext(ctx).
		Model(&model.Card{}).
		Where("status = ?", model.CardStatusUnused).
		Pluck("card_key", &keys).Error
	return keys, err
}

func (r *cardRepository) GetStats(ctx context.Context, agentID uint) (*CardStats, error) {
	stats := &CardStats{}

	var rows []struct {
		Status int
		Total  int64
	}
	if err := r.db.WithContext(ctx).
		Model(&model.Card{}).
		Select("status, COUNT(*) AS total").
		Where("agent_id = ?", agentID).
		Group("status").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		stats.Total += row.Total
		switch row.Status {
		case model.CardStatusUnused:
			stats.Unused = row.Total
		case model.CardStatusUsed:
			stats.Used = row.Total
		case model.CardStatusDestroyed:
			stats.Destroyed = row.Total
		}
	}

	return stats, nil
}

func (r *cardRepository) Create(ctx context.Context, card *model.Card) error {
	return r.db.WithContext(ctx).Create(card).Error
}

func (r *cardRepository) CreatePointsRecord(ctx context.Context, record *model.PointsRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}
