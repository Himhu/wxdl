package service

import (
	"context"
	"time"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/utils"

	"github.com/shopspring/decimal"
)

const (
	defaultPointsPage     = 1
	defaultPointsPageSize = 20
	maxPointsPageSize     = 100
)

type RechargeApplyInput struct {
	AgentID       uint
	Amount        decimal.Decimal
	PaymentMethod string
	PaymentProof  string
	Remark        string
}

type RechargeApproveInput struct {
	ApproverID uint
	RequestID  uint
}

type RechargeRejectInput struct {
	ApproverID uint
	RequestID  uint
}

type RechargeHistoryInput struct {
	AgentID  uint
	Page     int
	PageSize int
	Status   *int
}

type PendingRechargeInput struct {
	AgentID  uint
	Page     int
	PageSize int
}

type PointsRecordsInput struct {
	AgentID  uint
	Page     int
	PageSize int
	Type     *int
}

type PointsRecordsResult struct {
	List     []model.PointsRecord `json:"list"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
}

type PointsBalanceResult struct {
	Balance decimal.Decimal `json:"balance"`
}

type RechargeApplyResult struct {
	RequestID uint            `json:"requestId"`
	Amount    decimal.Decimal `json:"amount"`
	Status    int             `json:"status"`
}

type RechargeActionResult struct {
	RequestID          uint            `json:"requestId"`
	Status             int             `json:"status"`
	ParentBalanceAfter decimal.Decimal `json:"parentBalanceAfter,omitempty"`
	ChildBalanceAfter  decimal.Decimal `json:"childBalanceAfter,omitempty"`
}

type RechargeListResult struct {
	List     []repository.RechargeRequestItem `json:"list"`
	Total    int64                            `json:"total"`
	Page     int                              `json:"page"`
	PageSize int                              `json:"pageSize"`
}

type PointsStatsResult = repository.PointsStats

type PointsService interface {
	Balance(ctx context.Context, agentID uint) (*PointsBalanceResult, error)
	ApplyRecharge(ctx context.Context, input RechargeApplyInput) (*RechargeApplyResult, error)
	PendingRechargeRequests(ctx context.Context, input PendingRechargeInput) (*RechargeListResult, error)
	ApproveRecharge(ctx context.Context, input RechargeApproveInput) (*RechargeActionResult, error)
	RejectRecharge(ctx context.Context, input RechargeRejectInput) (*RechargeActionResult, error)
	RechargeHistory(ctx context.Context, input RechargeHistoryInput) (*RechargeListResult, error)
	Records(ctx context.Context, input PointsRecordsInput) (*PointsRecordsResult, error)
	Stats(ctx context.Context, agentID uint) (*PointsStatsResult, error)
}

type pointsService struct {
	agentService     AgentService
	pointsRepository repository.PointsRepository
	txManager        repository.TxManager
}

func NewPointsService(agentService AgentService, pointsRepository repository.PointsRepository, txManager repository.TxManager) PointsService {
	return &pointsService{
		agentService:     agentService,
		pointsRepository: pointsRepository,
		txManager:        txManager,
	}
}

func (s *pointsService) Balance(ctx context.Context, agentID uint) (*PointsBalanceResult, error) {
	agent, err := s.agentService.EnsureAgentActive(ctx, agentID)
	if err != nil {
		return nil, err
	}
	return &PointsBalanceResult{Balance: agent.Balance}, nil
}

func (s *pointsService) ApplyRecharge(ctx context.Context, input RechargeApplyInput) (*RechargeApplyResult, error) {
	if input.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, utils.BadRequestError("amount must be greater than 0")
	}
	if _, err := s.agentService.EnsureAgentActive(ctx, input.AgentID); err != nil {
		return nil, err
	}

	var request model.RechargeRequest
	if err := s.txManager.WithinTx(ctx, func(repos repository.TxRepositories) error {
		agent, err := repos.Agent().GetByIDForUpdate(ctx, input.AgentID)
		if err != nil {
			return err
		}
		if agent == nil {
			return repository.ErrAgentNotFound
		}
		if agent.ParentID == nil {
			return repository.ErrNoDirectParent
		}

		parent, err := repos.Agent().GetByID(ctx, *agent.ParentID)
		if err != nil {
			return err
		}
		if parent == nil {
			return repository.ErrNoDirectParent
		}

		request = model.RechargeRequest{
			AgentID:       agent.ID,
			Amount:        input.Amount,
			Status:        model.RechargeRequestStatusPending,
			PaymentMethod: input.PaymentMethod,
			PaymentProof:  input.PaymentProof,
			Remark:        input.Remark,
		}
		return repos.RechargeRequest().Create(ctx, &request)
	}); err != nil {
		switch err {
		case repository.ErrAgentNotFound:
			return nil, utils.NotFoundError("agent not found")
		case repository.ErrNoDirectParent:
			return nil, utils.BadRequestError("only agents with a direct parent can apply for recharge")
		default:
			return nil, utils.InternalError(err)
		}
	}

	return &RechargeApplyResult{RequestID: request.ID, Amount: request.Amount, Status: request.Status}, nil
}

func (s *pointsService) PendingRechargeRequests(ctx context.Context, input PendingRechargeInput) (*RechargeListResult, error) {
	page, pageSize := normalizePointsPage(input.Page, input.PageSize)
	if _, err := s.agentService.EnsureAgentActive(ctx, input.AgentID); err != nil {
		return nil, err
	}
	list, total, err := s.pointsRepository.ListPendingRechargeRequests(ctx, repository.RechargeRequestQuery{
		AgentID:  input.AgentID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, utils.InternalError(err)
	}

	return &RechargeListResult{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *pointsService) ApproveRecharge(ctx context.Context, input RechargeApproveInput) (*RechargeActionResult, error) {
	if _, err := s.agentService.EnsureAgentActive(ctx, input.ApproverID); err != nil {
		return nil, err
	}

	var result RechargeActionResult
	if err := s.txManager.WithinTx(ctx, func(repos repository.TxRepositories) error {
		request, err := repos.RechargeRequest().GetByIDForUpdate(ctx, input.RequestID)
		if err != nil {
			return err
		}
		if request == nil {
			return repository.ErrRechargeRequestNotFound
		}
		if request.Status != model.RechargeRequestStatusPending {
			return repository.ErrRechargeRequestHandled
		}

		childID := request.AgentID
		parentID := input.ApproverID
		firstID, secondID := childID, parentID
		if parentID < childID {
			firstID, secondID = parentID, childID
		}

		first, err := repos.Agent().GetByIDForUpdate(ctx, firstID)
		if err != nil {
			return err
		}
		if first == nil {
			return repository.ErrAgentNotFound
		}

		second, err := repos.Agent().GetByIDForUpdate(ctx, secondID)
		if err != nil {
			return err
		}
		if second == nil {
			return repository.ErrAgentNotFound
		}

		var child, parent *model.Agent
		if first.ID == childID {
			child, parent = first, second
		} else {
			child, parent = second, first
		}

		if child.ParentID == nil || *child.ParentID != parentID {
			return repository.ErrRechargeApprovalDenied
		}
		if parent.Balance.LessThan(request.Amount) {
			return repository.ErrInsufficientBalance
		}

		parentBalanceBefore := parent.Balance
		parentBalanceAfter := parent.Balance.Sub(request.Amount)
		childBalanceBefore := child.Balance
		childBalanceAfter := child.Balance.Add(request.Amount)
		now := time.Now()

		if err := repos.Agent().UpdateBalance(ctx, parent.ID, parentBalanceAfter); err != nil {
			return err
		}
		if err := repos.Agent().UpdateBalance(ctx, child.ID, childBalanceAfter); err != nil {
			return err
		}
		if err := repos.RechargeRequest().UpdateStatus(ctx, request.ID, model.RechargeRequestStatusApproved, parentID, now); err != nil {
			return err
		}

		relatedID := request.ID
		if err := repos.PointsRecord().Create(ctx, &model.PointsRecord{
			AgentID:       parent.ID,
			Type:          model.PointsRecordTypeConsume,
			Amount:        request.Amount,
			BalanceBefore: parentBalanceBefore,
			BalanceAfter:  parentBalanceAfter,
			Description:   "审批下级充值申请扣除积分",
			RelatedID:     &relatedID,
		}); err != nil {
			return err
		}
		if err := repos.PointsRecord().Create(ctx, &model.PointsRecord{
			AgentID:       child.ID,
			Type:          model.PointsRecordTypeRecharge,
			Amount:        request.Amount,
			BalanceBefore: childBalanceBefore,
			BalanceAfter:  childBalanceAfter,
			Description:   "上级审批充值申请到账",
			RelatedID:     &relatedID,
		}); err != nil {
			return err
		}

		result = RechargeActionResult{
			RequestID:          request.ID,
			Status:             model.RechargeRequestStatusApproved,
			ParentBalanceAfter: parentBalanceAfter,
			ChildBalanceAfter:  childBalanceAfter,
		}
		return nil
	}); err != nil {
		switch err {
		case repository.ErrRechargeRequestNotFound:
			return nil, utils.NotFoundError("recharge request not found")
		case repository.ErrRechargeRequestHandled:
			return nil, utils.ConflictError("recharge request already handled")
		case repository.ErrRechargeApprovalDenied:
			return nil, utils.UnauthorizedError("cannot approve this recharge request")
		case repository.ErrAgentNotFound:
			return nil, utils.NotFoundError("agent not found")
		case repository.ErrInsufficientBalance:
			return nil, utils.ConflictError("insufficient balance")
		default:
			return nil, utils.InternalError(err)
		}
	}

	return &result, nil
}

func (s *pointsService) RejectRecharge(ctx context.Context, input RechargeRejectInput) (*RechargeActionResult, error) {
	if _, err := s.agentService.EnsureAgentActive(ctx, input.ApproverID); err != nil {
		return nil, err
	}

	var result RechargeActionResult
	if err := s.txManager.WithinTx(ctx, func(repos repository.TxRepositories) error {
		request, err := repos.RechargeRequest().GetByIDForUpdate(ctx, input.RequestID)
		if err != nil {
			return err
		}
		if request == nil {
			return repository.ErrRechargeRequestNotFound
		}
		if request.Status != model.RechargeRequestStatusPending {
			return repository.ErrRechargeRequestHandled
		}

		child, err := repos.Agent().GetByID(ctx, request.AgentID)
		if err != nil {
			return err
		}
		if child == nil {
			return repository.ErrAgentNotFound
		}
		if child.ParentID == nil || *child.ParentID != input.ApproverID {
			return repository.ErrRechargeApprovalDenied
		}

		now := time.Now()
		if err := repos.RechargeRequest().UpdateStatus(ctx, request.ID, model.RechargeRequestStatusRejected, input.ApproverID, now); err != nil {
			return err
		}

		result = RechargeActionResult{
			RequestID: request.ID,
			Status:    model.RechargeRequestStatusRejected,
		}
		return nil
	}); err != nil {
		switch err {
		case repository.ErrRechargeRequestNotFound:
			return nil, utils.NotFoundError("recharge request not found")
		case repository.ErrRechargeRequestHandled:
			return nil, utils.ConflictError("recharge request already handled")
		case repository.ErrRechargeApprovalDenied:
			return nil, utils.UnauthorizedError("cannot reject this recharge request")
		case repository.ErrAgentNotFound:
			return nil, utils.NotFoundError("agent not found")
		default:
			return nil, utils.InternalError(err)
		}
	}

	return &result, nil
}

func (s *pointsService) RechargeHistory(ctx context.Context, input RechargeHistoryInput) (*RechargeListResult, error) {
	page, pageSize := normalizePointsPage(input.Page, input.PageSize)
	if _, err := s.agentService.EnsureAgentActive(ctx, input.AgentID); err != nil {
		return nil, err
	}
	if input.Status != nil && (*input.Status < model.RechargeRequestStatusPending || *input.Status > model.RechargeRequestStatusRejected) {
		return nil, utils.BadRequestError("status must be 0, 1 or 2")
	}

	list, total, err := s.pointsRepository.ListRechargeHistory(ctx, repository.RechargeRequestQuery{
		AgentID:  input.AgentID,
		Page:     page,
		PageSize: pageSize,
		Status:   input.Status,
	})
	if err != nil {
		return nil, utils.InternalError(err)
	}

	return &RechargeListResult{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *pointsService) Records(ctx context.Context, input PointsRecordsInput) (*PointsRecordsResult, error) {
	page, pageSize := normalizePointsPage(input.Page, input.PageSize)
	if _, err := s.agentService.EnsureAgentActive(ctx, input.AgentID); err != nil {
		return nil, err
	}
	if input.Type != nil && (*input.Type < model.PointsRecordTypeRecharge || *input.Type > model.PointsRecordTypeRefund) {
		return nil, utils.BadRequestError("type must be 1, 2 or 3")
	}

	records, total, err := s.pointsRepository.ListRecords(ctx, repository.PointsRecordQuery{
		AgentID:  input.AgentID,
		Page:     page,
		PageSize: pageSize,
		Type:     input.Type,
	})
	if err != nil {
		return nil, utils.InternalError(err)
	}

	return &PointsRecordsResult{List: records, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *pointsService) Stats(ctx context.Context, agentID uint) (*PointsStatsResult, error) {
	if _, err := s.agentService.EnsureAgentActive(ctx, agentID); err != nil {
		return nil, err
	}
	stats, err := s.pointsRepository.GetStats(ctx, agentID)
	if err != nil {
		switch err {
		case repository.ErrAgentNotFound:
			return nil, utils.NotFoundError("agent not found")
		default:
			return nil, utils.InternalError(err)
		}
	}

	return stats, nil
}

func normalizePointsPage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = defaultPointsPage
	}
	if pageSize <= 0 {
		pageSize = defaultPointsPageSize
	}
	if pageSize > maxPointsPageSize {
		pageSize = maxPointsPageSize
	}
	return page, pageSize
}
