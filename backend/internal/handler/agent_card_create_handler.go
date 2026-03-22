package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type AgentCardHandler struct {
	settingService  service.SystemSettingService
	agentRepository repository.AgentRepository
	cardRepository  repository.CardRepository
	txManager       repository.TxManager
}

func NewAgentCardHandler(
	settingService service.SystemSettingService,
	agentRepository repository.AgentRepository,
	cardRepository repository.CardRepository,
	txManager repository.TxManager,
) *AgentCardHandler {
	return &AgentCardHandler{
		settingService:  settingService,
		agentRepository: agentRepository,
		cardRepository:  cardRepository,
		txManager:       txManager,
	}
}

func (h *AgentCardHandler) Create(c *gin.Context) {
	agentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("未登录"))
		return
	}

	var req struct {
		Quota int `json:"quota" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.BadRequestError("请输入额度"))
		return
	}

	// 读取兑换码站点配置
	cfg, err := h.settingService.GetRedemptionConfig(c.Request.Context())
	if err != nil || cfg.BaseURL == "" || cfg.AdminAccessToken == "" || cfg.AdminUserID == "" {
		utils.Error(c, utils.BadRequestError("兑换码站点未配置"))
		return
	}

	// 查代理昵称
	agent, _ := h.agentRepository.GetByID(c.Request.Context(), agentID)
	agentName := fmt.Sprintf("agent_%d", agentID)
	if agent != nil && agent.RealName != "" {
		agentName = agent.RealName
	}

	// 面值转内部额度（1元 = 100内部额度）
	internalQuota := req.Quota * 500000

	// 调外部创建接口 → 写本地 cards
	var card *model.Card

	body, _ := json.Marshal(map[string]interface{}{
		"name":         agentName,
		"count":        1,
		"quota":        internalQuota,
		"expired_time": 0,
	})
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	externalReq, err := http.NewRequestWithContext(c.Request.Context(), "POST", baseURL+"/api/redemption/", bytes.NewReader(body))
	if err != nil {
		utils.Error(c, utils.BadRequestError("构造请求失败"))
		return
	}
	externalReq.Header.Set("Content-Type", "application/json")
	externalReq.Header.Set("Authorization", "Bearer "+cfg.AdminAccessToken)
	externalReq.Header.Set("New-Api-User", cfg.AdminUserID)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(externalReq)
	if err != nil {
		utils.Error(c, utils.BadRequestError(fmt.Sprintf("请求兑换码站点失败: %v", err)))
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var externalResp struct {
		Success bool     `json:"success"`
		Message string   `json:"message"`
		Data    []string `json:"data"`
	}
	if err := json.Unmarshal(respBody, &externalResp); err != nil {
		utils.Error(c, utils.BadRequestError("兑换码站点响应异常"))
		return
	}
	if !externalResp.Success || len(externalResp.Data) == 0 {
		msg := externalResp.Message
		if msg == "" {
			msg = "创建失败"
		}
		utils.Error(c, utils.BadRequestError(fmt.Sprintf("兑换码站点: %s", msg)))
		return
	}

	// 外部创建成功，事务化写入本地（card + 扣积分 + 写流水）
	if err := h.txManager.WithinTx(c.Request.Context(), func(repos repository.TxRepositories) error {
		// 锁代理余额
		lockedAgent, err := h.agentRepository.GetByIDForUpdate(c.Request.Context(), agentID)
		if err != nil || lockedAgent == nil {
			return fmt.Errorf("代理不存在")
		}

		// 计算成本（面值即成本，1:1）
		cost := decimal.NewFromInt(int64(req.Quota))

		// 校验余额
		if lockedAgent.Balance.LessThan(cost) {
			return fmt.Errorf("积分余额不足，需要 %s，当前 %s", cost.StringFixed(2), lockedAgent.Balance.StringFixed(2))
		}

		// 扣积分
		newBalance := lockedAgent.Balance.Sub(cost)
		if err := h.agentRepository.UpdateBalance(c.Request.Context(), agentID, newBalance); err != nil {
			return fmt.Errorf("扣减积分失败")
		}

		// 写积分流水（balance_logs）
		pointsRecord := &model.PointsRecord{
			AgentID:       agentID,
			Type:          model.PointsRecordTypeConsume,
			Amount:        cost.Neg(),
			BalanceBefore: lockedAgent.Balance,
			BalanceAfter:  newBalance,
			Description:   fmt.Sprintf("创建兑换码 面值%d元", req.Quota),
		}
		if err := h.cardRepository.CreatePointsRecord(c.Request.Context(), pointsRecord); err != nil {
			return fmt.Errorf("写积分流水失败")
		}

		// 写本地 cards
		card = &model.Card{
			CardKey: externalResp.Data[0],
			AgentID: agentID,
			Quota:   req.Quota,
			Cost:    cost,
			Status:  model.CardStatusUnused,
		}
		return h.cardRepository.Create(c.Request.Context(), card)
	}); err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	utils.Success(c, "创建成功", gin.H{
		"card": card,
	})
}
