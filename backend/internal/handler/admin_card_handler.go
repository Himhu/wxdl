package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AdminCardHandler struct {
	settingService service.SystemSettingService
	cardRepo       repository.CardRepository
}

func NewAdminCardHandler(settingService service.SystemSettingService, cardRepo repository.CardRepository) *AdminCardHandler {
	return &AdminCardHandler{settingService: settingService, cardRepo: cardRepo}
}

func (h *AdminCardHandler) List(c *gin.Context) {
	_, ok := middleware.GetAdminID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("unauthorized"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	var statusPtr *int
	if s := strings.TrimSpace(c.Query("status")); s != "" {
		v, _ := strconv.Atoi(s)
		if v >= 1 && v <= 3 {
			statusPtr = &v
		}
	}

	var agentIDPtr *uint
	if s := strings.TrimSpace(c.Query("agentId")); s != "" {
		v, _ := strconv.ParseUint(s, 10, 64)
		if v > 0 {
			aid := uint(v)
			agentIDPtr = &aid
		}
	}

	items, total, err := h.cardRepo.ListAll(c.Request.Context(), repository.AdminCardQuery{
		AgentID:  agentIDPtr,
		Page:     page,
		PageSize: pageSize,
		Status:   statusPtr,
		Keyword:  c.Query("keyword"),
	})
	if err != nil {
		utils.Error(c, utils.InternalError(err))
		return
	}

	list := make([]gin.H, 0, len(items))
	for _, item := range items {
		agentName := item.AgentUsername
		if strings.TrimSpace(item.AgentRealName) != "" {
			agentName = item.AgentRealName
		}
		list = append(list, gin.H{
			"id":        item.ID,
			"cardKey":   item.CardKey,
			"quota":     item.Quota,
			"cost":      item.Cost,
			"status":    item.Status,
			"agentId":   item.AgentID,
			"agentName": agentName,
			"createdAt": item.CreatedAt,
		})
	}

	utils.Success(c, "ok", gin.H{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

type batchCreateCardRequest struct {
	Name        string `json:"name" binding:"required"`
	Count       int    `json:"count" binding:"required,min=1,max=100"`
	Quota       int    `json:"quota" binding:"required,min=1"`
	ExpiredTime int64  `json:"expired_time"`
}

type externalRedemptionListResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Items []struct {
			Key          string `json:"key"`
			Status       int    `json:"status"`
			RedeemedTime int64  `json:"redeemed_time"`
		} `json:"items"`
	} `json:"data"`
}

func (h *AdminCardHandler) SyncStatuses(c *gin.Context) {
	_, ok := middleware.GetAdminID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("unauthorized"))
		return
	}

	cfg, err := h.settingService.GetRedemptionConfig(c.Request.Context())
	if err != nil || cfg.BaseURL == "" || cfg.AdminAccessToken == "" || cfg.AdminUserID == "" {
		utils.Error(c, utils.BadRequestError("兑换码站点未配置"))
		return
	}

	// 先获取本地未使用的 card_key 集合
	unusedKeys, err := h.cardRepo.ListUnusedCardKeys(c.Request.Context())
	if err != nil {
		utils.Error(c, utils.InternalError(err))
		return
	}
	if len(unusedKeys) == 0 {
		utils.Success(c, "无需同步，本地没有未使用的卡密", gin.H{
			"localUnused":  0,
			"updatedCount": 0,
		})
		return
	}

	// 构建 key→true 的查找集合
	keySet := make(map[string]bool, len(unusedKeys))
	for _, k := range unusedKeys {
		keySet[k] = true
	}

	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	client := &http.Client{Timeout: 20 * time.Second}

	var (
		page        = 1
		pageSize    = 100
		updatedUsed int
		matched     int
	)

	for {
		req, err := http.NewRequestWithContext(
			c.Request.Context(), "GET",
			fmt.Sprintf("%s/api/redemption/?p=%d&page_size=%d", baseURL, page, pageSize),
			nil,
		)
		if err != nil {
			utils.Error(c, utils.InternalError(fmt.Errorf("构造同步请求失败: %w", err)))
			return
		}
		req.Header.Set("Authorization", "Bearer "+cfg.AdminAccessToken)
		req.Header.Set("New-Api-User", cfg.AdminUserID)

		resp, err := client.Do(req)
		if err != nil {
			utils.Error(c, utils.BadRequestError(fmt.Sprintf("请求兑换码站点失败: %v", err)))
			return
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var externalResp externalRedemptionListResponse
		if err := json.Unmarshal(respBody, &externalResp); err != nil {
			utils.Error(c, utils.BadRequestError("兑换码站点返回格式异常"))
			return
		}
		if !externalResp.Success {
			utils.Error(c, utils.BadRequestError(externalResp.Message))
			return
		}

		items := externalResp.Data.Items
		if len(items) == 0 {
			break
		}

		for _, item := range items {
			if !keySet[item.Key] {
				continue
			}
			matched++
			if item.Status == 3 {
				var usedAt *time.Time
				if item.RedeemedTime > 0 {
					t := time.Unix(item.RedeemedTime, 0)
					usedAt = &t
				}
				updated, _ := h.cardRepo.SyncStatusByCardKey(c.Request.Context(), item.Key, 2, usedAt)
				if updated {
					updatedUsed++
				}
			}
		}

		if len(items) < pageSize {
			break
		}
		page++
	}

	utils.Success(c, "同步完成", gin.H{
		"localUnused":  len(unusedKeys),
		"matched":      matched,
		"updatedCount": updatedUsed,
	})
}

func (h *AdminCardHandler) BatchCreate(c *gin.Context) {
	_, ok := middleware.GetAdminID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("unauthorized"))
		return
	}

	var req batchCreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.BadRequestError("参数错误：请填写名称、数量和额度"))
		return
	}

	cfg, err := h.settingService.GetRedemptionConfig(c.Request.Context())
	if err != nil || cfg.BaseURL == "" || cfg.AdminAccessToken == "" || cfg.AdminUserID == "" {
		utils.Error(c, utils.BadRequestError("兑换码站点未配置，请先在系统设置中配置"))
		return
	}

	// 构造外部请求
	body, _ := json.Marshal(map[string]interface{}{
		"name":         req.Name,
		"count":        req.Count,
		"quota":        req.Quota,
		"expired_time": req.ExpiredTime,
	})

	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	externalReq, err := http.NewRequestWithContext(c.Request.Context(), "POST", baseURL+"/api/redemption/", bytes.NewReader(body))
	if err != nil {
		utils.Error(c, utils.InternalError(fmt.Errorf("构造请求失败: %w", err)))
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
		utils.Error(c, utils.BadRequestError("兑换码站点返回格式异常"))
		return
	}

	if !externalResp.Success {
		msg := externalResp.Message
		if msg == "" {
			msg = "创建失败"
		}
		// 部分成功：仍然返回已生成的 key
		if len(externalResp.Data) > 0 {
			utils.Success(c, fmt.Sprintf("部分成功: %s", msg), gin.H{"keys": externalResp.Data})
			return
		}
		utils.Error(c, utils.BadRequestError(msg))
		return
	}

	utils.Success(c, "创建成功", gin.H{"keys": externalResp.Data})
}
