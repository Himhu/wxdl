package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AdminSystemSettingHandler struct {
	service service.SystemSettingService
}

func NewAdminSystemSettingHandler(service service.SystemSettingService) *AdminSystemSettingHandler {
	return &AdminSystemSettingHandler{service: service}
}

type updateWeChatSettingsRequest struct {
	AppID      string `json:"appId" binding:"required"`
	AppSecret  string `json:"appSecret" binding:"required"`
	ChangeNote string `json:"changeNote" binding:"max=255"`
}

func (h *AdminSystemSettingHandler) GetWeChatSettings(c *gin.Context) {
	result, err := h.service.GetWeChatSettings(c.Request.Context())
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, "wechat settings fetched", result)
}

func (h *AdminSystemSettingHandler) UpdateWeChatSettings(c *gin.Context) {
	var req updateWeChatSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	adminID, ok := middleware.GetAdminID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("unauthorized"))
		return
	}
	result, err := h.service.UpdateWeChatSettings(c.Request.Context(), service.UpdateWeChatSettingsInput{
		AppID:      req.AppID,
		AppSecret:  req.AppSecret,
		ChangeNote: req.ChangeNote,
		ChangedBy:  adminID,
		ChangedIP:  c.ClientIP(),
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "wechat settings updated", result)
}

type updateObjectStorageSettingsRequest struct {
	Enabled      bool   `json:"enabled"`
	Provider     string `json:"provider"`
	Endpoint     string `json:"endpoint"`
	Bucket       string `json:"bucket"`
	AccessKeyID  string `json:"accessKeyId"`
	SecretKey    string `json:"secretKey"`
	Region       string `json:"region"`
	CustomDomain string `json:"customDomain"`
	PathPrefix   string `json:"pathPrefix"`
	ChangeNote   string `json:"changeNote" binding:"max=255"`
}

func (h *AdminSystemSettingHandler) GetObjectStorageSettings(c *gin.Context) {
	result, err := h.service.GetObjectStorageSettings(c.Request.Context())
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, "object storage settings fetched", result)
}

func (h *AdminSystemSettingHandler) UpdateObjectStorageSettings(c *gin.Context) {
	var req updateObjectStorageSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	adminID, ok := middleware.GetAdminID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("unauthorized"))
		return
	}

	result, err := h.service.UpdateObjectStorageSettings(c.Request.Context(), service.UpdateObjectStorageSettingsInput{
		Enabled:      req.Enabled,
		Provider:     req.Provider,
		Endpoint:     req.Endpoint,
		Bucket:       req.Bucket,
		AccessKeyID:  req.AccessKeyID,
		SecretKey:    req.SecretKey,
		Region:       req.Region,
		CustomDomain: req.CustomDomain,
		PathPrefix:   req.PathPrefix,
		ChangeNote:   req.ChangeNote,
		ChangedBy:    adminID,
		ChangedIP:    c.ClientIP(),
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "object storage settings updated", result)
}

type updateRedemptionSettingsRequest struct {
	BaseURL          string `json:"baseUrl"`
	AdminAccessToken string `json:"adminAccessToken"`
	AdminUserID      string `json:"adminUserId"`
	PriceRules       string `json:"priceRules"`
	ChangeNote       string `json:"changeNote" binding:"max=255"`
}

func (h *AdminSystemSettingHandler) GetRedemptionSettings(c *gin.Context) {
	result, err := h.service.GetRedemptionSettings(c.Request.Context())
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, "redemption settings fetched", result)
}

func (h *AdminSystemSettingHandler) UpdateRedemptionSettings(c *gin.Context) {
	var req updateRedemptionSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	adminID, ok := middleware.GetAdminID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("unauthorized"))
		return
	}

	result, err := h.service.UpdateRedemptionSettings(c.Request.Context(), service.UpdateRedemptionSettingsInput{
		BaseURL:          req.BaseURL,
		AdminAccessToken: req.AdminAccessToken,
		AdminUserID:      req.AdminUserID,
		PriceRules:       req.PriceRules,
		ChangeNote:       req.ChangeNote,
		ChangedBy:        adminID,
		ChangedIP:        c.ClientIP(),
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "redemption settings updated", result)
}

type updateAgentPricingSettingsRequest struct {
	Level1Price string `json:"level1Price" binding:"required"`
	Level2Price string `json:"level2Price" binding:"required"`
	ChangeNote  string `json:"changeNote" binding:"max=255"`
}

func (h *AdminSystemSettingHandler) GetAgentPricingSettings(c *gin.Context) {
	result, err := h.service.GetAgentPricingSettings(c.Request.Context())
	if err != nil {
		utils.Error(c, err)
		return
	}
	utils.Success(c, "agent pricing settings fetched", result)
}

func (h *AdminSystemSettingHandler) UpdateAgentPricingSettings(c *gin.Context) {
	var req updateAgentPricingSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	adminID, ok := middleware.GetAdminID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("unauthorized"))
		return
	}

	result, err := h.service.UpdateAgentPricingSettings(c.Request.Context(), service.UpdateAgentPricingSettingsInput{
		Level1Price: req.Level1Price,
		Level2Price: req.Level2Price,
		ChangeNote:  req.ChangeNote,
		ChangedBy:   adminID,
		ChangedIP:   c.ClientIP(),
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "agent pricing settings updated", result)
}
