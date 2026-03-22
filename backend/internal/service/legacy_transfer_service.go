package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/utils"

	"github.com/shopspring/decimal"
)

type LegacyTransferService struct {
	legacySiteService *LegacySiteService
	userRepo          repository.UserRepository
	agentRepo         repository.AgentRepository
	txManager         repository.TxManager
	legacyAdminUser   string
	legacyAdminPass   string
}

type LegacyTransferConfirmInput struct {
	UserID         uint64
	LegacyUsername string
	LegacyPassword string
}

func NewLegacyTransferService(legacySiteService *LegacySiteService, userRepo repository.UserRepository, agentRepo repository.AgentRepository, txManager repository.TxManager, legacyAdminUser, legacyAdminPass string) *LegacyTransferService {
	return &LegacyTransferService{
		legacySiteService: legacySiteService,
		userRepo:          userRepo,
		agentRepo:         agentRepo,
		txManager:         txManager,
		legacyAdminUser:   legacyAdminUser,
		legacyAdminPass:   legacyAdminPass,
	}
}

func (s *LegacyTransferService) Confirm(ctx context.Context, input LegacyTransferConfirmInput) (*LegacyConfirmResult, error) {
	if strings.TrimSpace(input.LegacyUsername) == "" || strings.TrimSpace(input.LegacyPassword) == "" {
		return nil, fmt.Errorf("请输入旧站账号和密码")
	}
	if input.LegacyUsername == "rn19970217" {
		return nil, fmt.Errorf("该旧站账号禁止执行自动禁用，请走人工处理")
	}

	legacyInfo, err := s.legacySiteService.FetchBalance(ctx, input.LegacyUsername, input.LegacyPassword)
	if err != nil {
		return nil, err
	}
	amount, err := decimal.NewFromString(legacyInfo.Balance)
	if err != nil {
		return nil, fmt.Errorf("解析旧站余额失败")
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("旧站余额必须大于 0")
	}

	user, err := s.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("当前用户不存在")
	}
	if user.Role != model.UserRoleAgent {
		return nil, fmt.Errorf("当前账号尚未成为代理，无法转移余额")
	}

	agent, err := s.agentRepo.GetByWechatOpenID(ctx, user.OpenID)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		username := fmt.Sprintf("wx_%d", user.ID)
		if user.Mobile != "" {
			username = user.Mobile
		}
		if agentByUsername, err := s.agentRepo.GetByUsername(ctx, username); err != nil {
			return nil, err
		} else if agentByUsername != nil {
			username = fmt.Sprintf("wx_%d", user.ID)
		}

		hashedPassword, err := utils.HashPassword(fmt.Sprintf("legacy-%d", user.ID))
		if err != nil {
			return nil, err
		}
		createAgent := &model.Agent{
			Username:      username,
			Password:      hashedPassword,
			RealName:      user.Nickname,
			Phone:         user.Mobile,
			Level:         1,
			Status:        model.AgentStatusActive,
			Balance:       decimal.Zero,
			WechatOpenID:  user.OpenID,
			WechatUnionID: user.UnionID,
		}
		if err := s.agentRepo.Create(ctx, createAgent); err != nil {
			return nil, err
		}
		agent = createAgent
	}

	adminSession, err := s.legacySiteService.AdminLogin(ctx, s.legacyAdminUser, s.legacyAdminPass)
	if err != nil {
		return nil, err
	}
	legacyAgent, err := s.legacySiteService.SearchAgentByUsername(ctx, adminSession, input.LegacyUsername)
	if err != nil {
		return nil, err
	}
	if legacyAgent.Username != input.LegacyUsername {
		return nil, fmt.Errorf("旧站代理身份校验失败")
	}

	if err := s.legacySiteService.DisableAgent(ctx, adminSession, legacyAgent.ID); err != nil {
		return nil, err
	}

	var newBalance decimal.Decimal
	if err := s.txManager.WithinTx(ctx, func(repos repository.TxRepositories) error {
		lockedAgent, err := repos.Agent().GetByIDForUpdate(ctx, agent.ID)
		if err != nil {
			return err
		}
		if lockedAgent == nil {
			return fmt.Errorf("代理账户不存在")
		}

		recordKey := fmt.Sprintf("legacy-transfer:%d:%s", legacyInfo.LegacyUserID, legacyInfo.LegacyUsername)
		exists, err := repos.PointsRecord().ExistsByDescription(ctx, lockedAgent.ID, fmt.Sprintf("旧站余额转移入账 %s", recordKey))
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("该旧站账号已完成余额转移，请勿重复提交")
		}

		balanceBefore := lockedAgent.Balance
		newBalance = balanceBefore.Add(amount)
		if err := repos.Agent().UpdateBalance(ctx, lockedAgent.ID, newBalance); err != nil {
			return err
		}

		record := &model.PointsRecord{
			AgentID:       lockedAgent.ID,
			Type:          model.PointsRecordTypeRecharge,
			Amount:        amount,
			BalanceBefore: balanceBefore,
			BalanceAfter:  newBalance,
			Description:   fmt.Sprintf("旧站余额转移入账 %s", recordKey),
			RelatedID:     nil,
			CreatedAt:     time.Now(),
		}
		return repos.PointsRecord().Create(ctx, record)
	}); err != nil {
		return nil, err
	}

	return &LegacyConfirmResult{
		LegacyUserID:      legacyInfo.LegacyUserID,
		LegacyUsername:    legacyInfo.LegacyUsername,
		TransferredAmount: legacyInfo.Balance,
		LegacyDisabled:    true,
		NewBalance:        newBalance.StringFixed(2),
		User:              user,
		Agent:             agent,
	}, nil
}
