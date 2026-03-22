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
	"backend/internal/service"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AdminCardHandler struct {
	settingService service.SystemSettingService
}

func NewAdminCardHandler(settingService service.SystemSettingService) *AdminCardHandler {
	return &AdminCardHandler{settingService: settingService}
}

type batchCreateCardRequest struct {
	Name        string `json:"name" binding:"required"`
	Count       int    `json:"count" binding:"required,min=1,max=100"`
	Quota       int    `json:"quota" binding:"required,min=1"`
	ExpiredTime int64  `json:"expired_time"`
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
