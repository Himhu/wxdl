package service

import (
	"context"
	"strings"
	"time"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/utils"

	"github.com/shopspring/decimal"
)

const (
	defaultCardPage     = 1
	defaultCardPageSize = 20
	maxCardPageSize     = 100
)

type CardListInput struct {
	AgentID  uint
	Page     int
	PageSize int
	Status   *int
	Keyword  string
}

type CardListResult struct {
	List     []model.Card `json:"list"`
	Total    int64        `json:"total"`
	Page     int          `json:"page"`
	PageSize int          `json:"pageSize"`
}

type DestroyCardResult struct {
	Card           model.Card      `json:"card"`
	RefundedPoints decimal.Decimal `json:"refundedPoints"`
	BalanceAfter   decimal.Decimal `json:"balanceAfter"`
}

type CardStatsResult struct {
	Total     int64 `json:"total"`
	Unused    int64 `json:"unused"`
	Used      int64 `json:"used"`
	Destroyed int64 `json:"destroyed"`
}

type CardService interface {
	List(ctx context.Context, input CardListInput) (*CardListResult, error)
	Detail(ctx context.Context, agentID, cardID uint) (*model.Card, error)
	Destroy(ctx context.Context, agentID, cardID uint) (*DestroyCardResult, error)
	Stats(ctx context.Context, agentID uint) (*CardStatsResult, error)
}

type cardService struct {
	agentService   AgentService
	cardRepository repository.CardRepository
	txManager      repository.TxManager
}

func NewCardService(
	agentService AgentService,
	cardRepository repository.CardRepository,
	txManager repository.TxManager,
) CardService {
	return &cardService{
		agentService:   agentService,
		cardRepository: cardRepository,
		txManager:      txManager,
	}
}

func (s *cardService) List(ctx context.Context, input CardListInput) (*CardListResult, error) {
	page := input.Page
	if page <= 0 {
		page = defaultCardPage
	}
	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = defaultCardPageSize
	}
	if pageSize > maxCardPageSize {
		pageSize = maxCardPageSize
	}
	if input.Status != nil {
		if *input.Status < model.CardStatusUnused || *input.Status > model.CardStatusDestroyed {
			return nil, utils.BadRequestError("status must be 1, 2 or 3")
		}
	}
	if _, err := s.agentService.EnsureAgentActive(ctx, input.AgentID); err != nil {
		return nil, err
	}

	cards, total, err := s.cardRepository.ListByAgent(ctx, repository.CardQuery{
		AgentID:  input.AgentID,
		Page:     page,
		PageSize: pageSize,
		Status:   input.Status,
		Keyword:  strings.TrimSpace(input.Keyword),
	})
	if err != nil {
		return nil, utils.InternalError(err)
	}

	return &CardListResult{
		List:     cards,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *cardService) Detail(ctx context.Context, agentID, cardID uint) (*model.Card, error) {
	if _, err := s.agentService.EnsureAgentActive(ctx, agentID); err != nil {
		return nil, err
	}

	card, err := s.cardRepository.GetByID(ctx, agentID, cardID)
	if err != nil {
		return nil, utils.InternalError(err)
	}
	if card == nil {
		return nil, utils.NotFoundError("card not found")
	}

	return card, nil
}

func (s *cardService) Destroy(ctx context.Context, agentID, cardID uint) (*DestroyCardResult, error) {
	if _, err := s.agentService.EnsureAgentActive(ctx, agentID); err != nil {
		return nil, err
	}

	var result DestroyCardResult
	if err := s.txManager.WithinTx(ctx, func(repos repository.TxRepositories) error {
		card, err := repos.Card().GetByIDForUpdate(ctx, agentID, cardID)
		if err != nil {
			return err
		}
		if card == nil {
			return repository.ErrCardNotFound
		}

		switch card.Status {
		case model.CardStatusUsed:
			return repository.ErrCardAlreadyUsed
		case model.CardStatusDestroyed:
			return repository.ErrCardAlreadyDestroyed
		}

		agent, err := repos.Agent().GetByIDForUpdate(ctx, agentID)
		if err != nil {
			return err
		}
		if agent == nil {
			return repository.ErrAgentNotFound
		}

		now := time.Now()
		if err := repos.Card().MarkDestroyed(ctx, cardID, now); err != nil {
			return err
		}

		refundAmount := card.Cost
		if refundAmount.IsZero() {
			refundAmount = decimal.NewFromInt(int64(card.Quota))
		}
		balanceBefore := agent.Balance
		balanceAfter := balanceBefore.Add(refundAmount)
		if err := repos.Agent().UpdateBalance(ctx, agentID, balanceAfter); err != nil {
			return err
		}

		relatedID := cardID
		if err := repos.PointsRecord().Create(ctx, &model.PointsRecord{
			AgentID:       agentID,
			Type:          model.PointsRecordTypeRefund,
			Amount:        refundAmount,
			BalanceBefore: balanceBefore,
			BalanceAfter:  balanceAfter,
			Description:   "销毁卡密返还积分",
			RelatedID:     &relatedID,
		}); err != nil {
			return err
		}

		card.Status = model.CardStatusDestroyed
		card.DestroyedAt = &now
		result = DestroyCardResult{
			Card:           *card,
			RefundedPoints: refundAmount,
			BalanceAfter:   balanceAfter,
		}
		return nil
	}); err != nil {
		switch err {
		case repository.ErrCardNotFound:
			return nil, utils.NotFoundError("card not found")
		case repository.ErrCardAlreadyUsed:
			return nil, utils.ConflictError("used card cannot be destroyed")
		case repository.ErrCardAlreadyDestroyed:
			return nil, utils.ConflictError("card already destroyed")
		case repository.ErrAgentNotFound:
			return nil, utils.NotFoundError("agent not found")
		default:
			return nil, utils.InternalError(err)
		}
	}

	return &result, nil
}

func (s *cardService) Stats(ctx context.Context, agentID uint) (*CardStatsResult, error) {
	if _, err := s.agentService.EnsureAgentActive(ctx, agentID); err != nil {
		return nil, err
	}

	stats, err := s.cardRepository.GetStats(ctx, agentID)
	if err != nil {
		return nil, utils.InternalError(err)
	}

	return &CardStatsResult{
		Total:     stats.Total,
		Unused:    stats.Unused,
		Used:      stats.Used,
		Destroyed: stats.Destroyed,
	}, nil
}
