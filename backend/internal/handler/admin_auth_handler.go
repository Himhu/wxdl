package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AdminAuthHandler struct {
	adminAuthService service.AdminAuthService
}

type adminLoginRequest struct {
	Username string `json:"username" binding:"required,min=2,max=50"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

func NewAdminAuthHandler(adminAuthService service.AdminAuthService) *AdminAuthHandler {
	return &AdminAuthHandler{adminAuthService: adminAuthService}
}

func (h *AdminAuthHandler) Login(c *gin.Context) {
	var request adminLoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	result, err := h.adminAuthService.Login(
		c.Request.Context(),
		request.Username,
		request.Password,
		c.ClientIP(),
	)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "admin login success", result)
}

func (h *AdminAuthHandler) Me(c *gin.Context) {
	adminID, ok := middleware.GetAdminID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	profile, err := h.adminAuthService.Me(c.Request.Context(), adminID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "current admin fetched", profile)
}
