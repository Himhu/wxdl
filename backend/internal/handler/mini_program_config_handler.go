package handler

import (
	"backend/internal/service"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type MiniProgramConfigHandler struct {
	service *service.MiniProgramConfigService
}

func NewMiniProgramConfigHandler(service *service.MiniProgramConfigService) *MiniProgramConfigHandler {
	return &MiniProgramConfigHandler{service: service}
}

// GetBootstrapConfig 获取小程序启动配置
func (h *MiniProgramConfigHandler) GetBootstrapConfig(c *gin.Context) {
	scopeType := c.DefaultQuery("scopeType", "global")
	scopeCode := c.DefaultQuery("scopeCode", "default")

	config, err := h.service.GetBootstrapConfig(scopeType, scopeCode)
	if err != nil {
		utils.Error(c, utils.InternalError(err))
		return
	}

	utils.Success(c, "bootstrap config fetched", gin.H{
		"version": 1,
		"config":  config,
	})
}
