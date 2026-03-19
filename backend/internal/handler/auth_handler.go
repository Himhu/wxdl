package handler

import (
	"backend/internal/middleware"
	"backend/internal/service"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

type loginRequest struct {
	Username string `json:"username" binding:"required,min=2,max=50"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

type wechatLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

type bindWechatRequest struct {
	Code string `json:"code" binding:"required"`
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request loginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	result, err := h.authService.Login(c.Request.Context(), service.LoginInput{
		Username:  request.Username,
		Password:  request.Password,
		ClientIP:  c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "login success", result)
}

func (h *AuthHandler) WechatLogin(c *gin.Context) {
	var request wechatLoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	result, err := h.authService.WechatLogin(c.Request.Context(), service.WechatLoginInput{
		Code:      request.Code,
		ClientIP:  c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "wechat login success", result)
}

func (h *AuthHandler) BindWechat(c *gin.Context) {
	var request bindWechatRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	agentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	result, err := h.authService.BindWechat(c.Request.Context(), agentID, service.BindWechatInput{
		Code: request.Code,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "wechat bind success", result)
}

func (h *AuthHandler) Me(c *gin.Context) {
	agentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	profile, err := h.authService.Me(c.Request.Context(), agentID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "current user fetched", profile)
}
