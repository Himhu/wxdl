package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/internal/utils"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminMiniProgramConfigHandler struct {
	service *service.MiniProgramConfigService
}

func NewAdminMiniProgramConfigHandler(service *service.MiniProgramConfigService) *AdminMiniProgramConfigHandler {
	return &AdminMiniProgramConfigHandler{service: service}
}

// ListConfigs 查询配置列表
func (h *AdminMiniProgramConfigHandler) ListConfigs(c *gin.Context) {
	namespace := c.Query("namespace")
	scopeType := c.DefaultQuery("scopeType", "global")
	scopeCode := c.DefaultQuery("scopeCode", "default")

	items, err := h.service.ListConfigs(namespace, scopeType, scopeCode)
	if err != nil {
		utils.Error(c, utils.InternalError(err))
		return
	}

	utils.Success(c, "configs fetched", items)
}

// GetConfig 获取单个配置
func (h *AdminMiniProgramConfigHandler) GetConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, utils.BadRequestError("invalid id"))
		return
	}

	item, err := h.service.GetConfig(uint(id))
	if err != nil {
		utils.Error(c, utils.NotFoundError("config not found"))
		return
	}

	utils.Success(c, "config fetched", item)
}

// UpdateConfig 更新配置
func (h *AdminMiniProgramConfigHandler) UpdateConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.Error(c, utils.BadRequestError("invalid id"))
		return
	}

	var req struct {
		DraftValue     string `json:"draftValue"`
		PublishedValue string `json:"publishedValue"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	// 兼容旧字段，优先使用publishedValue
	value := req.PublishedValue
	if value == "" {
		value = req.DraftValue
	}
	if value == "" {
		utils.Error(c, utils.BadRequestError("value is required"))
		return
	}

	adminID, ok := middleware.GetAdminID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("unauthorized"))
		return
	}
	if err := h.service.UpdateConfig(uint(id), value, adminID); err != nil {
		if errors.Is(err, service.ErrInvalidJSON) {
			utils.Error(c, utils.BadRequestError("invalid JSON format"))
			return
		}
		utils.Error(c, utils.InternalError(err))
		return
	}

	utils.Success(c, "success", nil)
}
