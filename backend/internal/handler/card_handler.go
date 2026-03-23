package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
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

type CardHandler struct {
	cardService    service.CardService
	cardRepo       repository.CardRepository
	settingService service.SystemSettingService
}

func NewCardHandler(
	cardService service.CardService,
	cardRepo repository.CardRepository,
	settingService service.SystemSettingService,
) *CardHandler {
	return &CardHandler{
		cardService:    cardService,
		cardRepo:       cardRepo,
		settingService: settingService,
	}
}

func (h *CardHandler) List(c *gin.Context) {
	agentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	page, err := parsePositiveInt(c.Query("page"), 1)
	if err != nil {
		utils.Error(c, utils.BadRequestError("invalid page"))
		return
	}
	pageSize, err := parsePositiveInt(c.Query("pageSize"), 20)
	if err != nil {
		utils.Error(c, utils.BadRequestError("invalid pageSize"))
		return
	}

	var status *int
	if rawStatus := c.Query("status"); rawStatus != "" {
		value, err := strconv.Atoi(rawStatus)
		if err != nil {
			utils.Error(c, utils.BadRequestError("invalid status"))
			return
		}
		status = &value
	}

	result, err := h.cardService.List(c.Request.Context(), service.CardListInput{
		AgentID:  agentID,
		Page:     page,
		PageSize: pageSize,
		Status:   status,
		Keyword:  c.Query("keyword"),
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "cards fetched", result)
}

func (h *CardHandler) Detail(c *gin.Context) {
	agentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	cardID, err := parseUintParam(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.BadRequestError("invalid card id"))
		return
	}

	card, err := h.cardService.Detail(c.Request.Context(), agentID, cardID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	h.enrichCardExternalInfo(c, card)

	utils.Success(c, "card fetched", card)
}

func (h *CardHandler) Destroy(c *gin.Context) {
	agentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	cardID, err := parseUintParam(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.BadRequestError("invalid card id"))
		return
	}

	card, err := h.cardService.Detail(c.Request.Context(), agentID, cardID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	externalID, refundAmount, err := h.calculateRefund(c, card)
	if err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	if err := h.disableExternalByID(c, externalID); err != nil {
		log.Printf("warn: external disable failed for card %d: %v", cardID, err)
		utils.Error(c, utils.BadRequestError("外部站点禁用失败，请稍后重试"))
		return
	}

	result, err := h.cardService.Destroy(c.Request.Context(), agentID, cardID, refundAmount)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "card destroyed", result)
}

func (h *CardHandler) Stats(c *gin.Context) {
	agentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	stats, err := h.cardService.Stats(c.Request.Context(), agentID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "card stats fetched", stats)
}

func (h *CardHandler) enrichCardExternalInfo(c *gin.Context, card *model.Card) {
	card.EstimatedRefund = card.Cost

	if card.Quota <= 0 {
		return
	}
	originalUnits := card.Quota * 500000

	cfg, err := h.settingService.GetRedemptionConfig(c.Request.Context())
	if err != nil || cfg == nil || cfg.BaseURL == "" {
		return
	}

	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	client := &http.Client{Timeout: 10 * time.Second}

	externalID, err := h.findExternalIDByKey(c, client, baseURL, cfg.AdminAccessToken, cfg.AdminUserID, card.CardKey)
	if err != nil || externalID == 0 {
		return
	}

	detail, err := h.getExternalDetail(c, client, baseURL, cfg.AdminAccessToken, cfg.AdminUserID, externalID)
	if err != nil {
		return
	}

	log.Printf("enrichCardExternalInfo: card=%s, externalStatus=%d, usedUserID=%d, quota=%d",
		card.CardKey, detail.Status, detail.UsedUserID, detail.Quota)

	externalUsed := detail.Status == 3 || detail.UsedUserID > 0
	if externalUsed && card.Status == model.CardStatusUnused {
		card.Status = model.CardStatusUsed
		updated, syncErr := h.cardRepo.SyncStatusByCardKey(c.Request.Context(), card.CardKey, model.CardStatusUsed, nil)
		if syncErr != nil {
			log.Printf("warn: sync used status failed for card %s: %v", card.CardKey, syncErr)
		} else if updated {
			log.Printf("enrichCardExternalInfo: synced local status to used for card=%s", card.CardKey)
		}
	}

	if detail.UsedUserID == 0 {
		card.ExternalRemaining = originalUnits
		card.ExternalUsed = 0
		return
	}

	user, err := h.getExternalUser(c, client, baseURL, cfg.AdminAccessToken, cfg.AdminUserID, detail.UsedUserID)
	if err != nil {
		return
	}

	if user.Username != "" {
		card.UsedBy = user.Username
	}

	remaining := user.Quota
	if remaining < 0 {
		remaining = 0
	}
	if remaining > originalUnits {
		remaining = originalUnits
	}

	card.ExternalRemaining = remaining
	card.ExternalUsed = originalUnits - remaining
	card.EstimatedRefund = decimal.NewFromInt(int64(remaining)).
		Div(decimal.NewFromInt(int64(originalUnits))).
		Mul(card.Cost).
		Round(2)

	log.Printf("enrichCardExternalInfo: userQuota=%d, remaining=%d, used=%d, estimatedRefund=%s",
		user.Quota, remaining, card.ExternalUsed, card.EstimatedRefund.String())
}

func (h *CardHandler) calculateRefund(c *gin.Context, card *model.Card) (int, decimal.Decimal, error) {
	cfg, err := h.settingService.GetRedemptionConfig(c.Request.Context())
	if err != nil || cfg == nil || cfg.BaseURL == "" || cfg.AdminAccessToken == "" || cfg.AdminUserID == "" {
		return 0, decimal.Zero, fmt.Errorf("兑换码站点未配置")
	}

	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	client := &http.Client{Timeout: 20 * time.Second}

	externalID, err := h.findExternalIDByKey(c, client, baseURL, cfg.AdminAccessToken, cfg.AdminUserID, card.CardKey)
	if err != nil {
		return 0, decimal.Zero, err
	}
	if externalID == 0 {
		return 0, decimal.Zero, fmt.Errorf("外站未找到该兑换码")
	}

	detail, err := h.getExternalDetail(c, client, baseURL, cfg.AdminAccessToken, cfg.AdminUserID, externalID)
	if err != nil {
		return 0, decimal.Zero, err
	}

	log.Printf("External detail for card %s: status=%d, usedUserID=%d, quota=%d", card.CardKey, detail.Status, detail.UsedUserID, detail.Quota)

	if detail.UsedUserID == 0 {
		log.Printf("UsedUserID is 0, returning full refund: %s", card.Cost.String())
		return externalID, card.Cost, nil
	}

	user, err := h.getExternalUser(c, client, baseURL, cfg.AdminAccessToken, cfg.AdminUserID, detail.UsedUserID)
	if err != nil {
		return 0, decimal.Zero, err
	}

	if card.Quota <= 0 {
		return 0, decimal.Zero, fmt.Errorf("卡密额度异常")
	}

	originalUnits := card.Quota * 500000
	userRemaining := user.Quota
	if userRemaining < 0 {
		userRemaining = 0
	}

	deductUnits := originalUnits
	if userRemaining < originalUnits {
		deductUnits = userRemaining
	}

	refundAmount := decimal.Zero
	if !card.Cost.IsZero() && deductUnits > 0 {
		refundAmount = decimal.NewFromInt(int64(deductUnits)).
			Div(decimal.NewFromInt(int64(originalUnits))).
			Mul(card.Cost).
			Round(2)
	}

	if deductUnits > 0 {
		newQuota := userRemaining - deductUnits
		if err := h.updateExternalUserQuota(c, client, baseURL, cfg.AdminAccessToken, cfg.AdminUserID, user, newQuota); err != nil {
			return 0, decimal.Zero, err
		}
	}

	return externalID, refundAmount, nil
}

type externalDetail struct {
	ID          int `json:"id"`
	Status      int `json:"status"`
	Quota       int `json:"quota"`
	UsedUserID  int `json:"used_user_id"`
}

type externalUser struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Quota       int    `json:"quota"`
	UsedQuota   int    `json:"used_quota"`
	Group       string `json:"group"`
	GitHubID    string `json:"github_id"`
	OidcID      string `json:"oidc_id"`
	DiscordID   string `json:"discord_id"`
	WechatID    string `json:"wechat_id"`
	TelegramID  string `json:"telegram_id"`
	LinuxDoID   string `json:"linux_do_id"`
}

func (h *CardHandler) getExternalDetail(c *gin.Context, client *http.Client, baseURL, token, userID string, id int) (*externalDetail, error) {
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, fmt.Sprintf("%s/api/redemption/%d", baseURL, id), nil)
	if err != nil {
		return nil, fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("New-Api-User", userID)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求外站详情失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Success bool           `json:"success"`
		Message string         `json:"message"`
		Data    externalDetail `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil || !result.Success {
		return nil, fmt.Errorf("外站详情查询失败")
	}
	return &result.Data, nil
}

func (h *CardHandler) getExternalUser(c *gin.Context, client *http.Client, baseURL, token, userID string, id int) (*externalUser, error) {
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, fmt.Sprintf("%s/api/user/%d", baseURL, id), nil)
	if err != nil {
		return nil, fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("New-Api-User", userID)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求外站用户失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Success bool         `json:"success"`
		Message string       `json:"message"`
		Data    externalUser `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil || !result.Success {
		return nil, fmt.Errorf("外站用户查询失败")
	}
	return &result.Data, nil
}

func (h *CardHandler) updateExternalUserQuota(c *gin.Context, client *http.Client, baseURL, token, userID string, user *externalUser, newQuota int) error {
	user.Quota = newQuota
	body, _ := json.Marshal(user)
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPut, baseURL+"/api/user/", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("New-Api-User", userID)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求外站失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil || !result.Success {
		return fmt.Errorf("更新用户额度失败")
	}
	return nil
}

func (h *CardHandler) disableExternalByID(c *gin.Context, externalID int) error {
	cfg, err := h.settingService.GetRedemptionConfig(c.Request.Context())
	if err != nil || cfg == nil {
		return fmt.Errorf("兑换码站点未配置")
	}

	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	client := &http.Client{Timeout: 20 * time.Second}

	body, _ := json.Marshal(map[string]interface{}{"id": externalID, "status": 2})
	req, _ := http.NewRequestWithContext(c.Request.Context(), http.MethodPut, baseURL+"/api/redemption/?status_only=1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.AdminAccessToken)
	req.Header.Set("New-Api-User", cfg.AdminUserID)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求外站失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil || !result.Success {
		return fmt.Errorf("外站禁用失败")
	}
	return nil
}

// disableExternalRedemption 在外站查找并禁用兑换码（status=2）
func (h *CardHandler) disableExternalRedemption(c *gin.Context, cardKey string) error {
	cfg, err := h.settingService.GetRedemptionConfig(c.Request.Context())
	if err != nil || cfg == nil || cfg.BaseURL == "" || cfg.AdminAccessToken == "" || cfg.AdminUserID == "" {
		return fmt.Errorf("兑换码站点未配置")
	}

	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	client := &http.Client{Timeout: 20 * time.Second}

	externalID, err := h.findExternalIDByKey(c, client, baseURL, cfg.AdminAccessToken, cfg.AdminUserID, cardKey)
	if err != nil {
		return err
	}
	if externalID == 0 {
		return fmt.Errorf("外站未找到该兑换码")
	}

	// PUT /api/redemption/?status_only=1  {"id": X, "status": 2}
	body, _ := json.Marshal(map[string]interface{}{
		"id":     externalID,
		"status": 2,
	})
	req, err := http.NewRequestWithContext(
		c.Request.Context(), http.MethodPut,
		baseURL+"/api/redemption/?status_only=1",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("构造禁用请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.AdminAccessToken)
	req.Header.Set("New-Api-User", cfg.AdminUserID)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求外站失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("外站响应异常")
	}
	if !result.Success {
		return fmt.Errorf("外站禁用失败: %s", result.Message)
	}

	return nil
}

// findExternalIDByKey 分页遍历外站兑换码，按 key 匹配找到外站 numeric ID
func (h *CardHandler) findExternalIDByKey(
	c *gin.Context,
	client *http.Client,
	baseURL, accessToken, adminUserID, cardKey string,
) (int, error) {
	type externalItem struct {
		ID  int    `json:"id"`
		Key string `json:"key"`
	}
	type externalResp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			Items []externalItem `json:"items"`
		} `json:"data"`
	}

	page, pageSize := 1, 100
	for {
		req, err := http.NewRequestWithContext(
			c.Request.Context(), http.MethodGet,
			fmt.Sprintf("%s/api/redemption/?p=%d&page_size=%d", baseURL, page, pageSize),
			nil,
		)
		if err != nil {
			return 0, fmt.Errorf("构造查询请求失败: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("New-Api-User", adminUserID)

		resp, err := client.Do(req)
		if err != nil {
			return 0, fmt.Errorf("请求外站列表失败: %w", err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var listResp externalResp
		if err := json.Unmarshal(body, &listResp); err != nil {
			return 0, fmt.Errorf("外站列表响应异常")
		}
		if !listResp.Success {
			return 0, fmt.Errorf("外站列表请求失败: %s", listResp.Message)
		}

		for _, item := range listResp.Data.Items {
			if item.Key == cardKey {
				return item.ID, nil
			}
		}

		if len(listResp.Data.Items) < pageSize {
			return 0, nil
		}
		page++
	}
}

func parsePositiveInt(value string, defaultValue int) (int, error) {
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if parsed <= 0 {
		return 0, strconv.ErrSyntax
	}
	return parsed, nil
}

func parseUintParam(value string) (uint, error) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}
