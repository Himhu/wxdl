package handler

import (
	"strconv"

	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/service"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AgentHandler struct {
	agentService service.AgentService
}

type createAgentRequest struct {
	Username string `json:"username" binding:"required,min=2,max=50"`
	Password string `json:"password" binding:"required,min=6,max=64"`
	RealName string `json:"realName" binding:"max=50"`
	Phone    string `json:"phone" binding:"max=20"`
}

type updateAgentRequest struct {
	Username string `json:"username" binding:"omitempty,min=2,max=50"`
	Password string `json:"password" binding:"omitempty,min=6,max=64"`
	RealName string `json:"realName" binding:"omitempty,max=50"`
	Phone    string `json:"phone" binding:"omitempty,max=20"`
}

type updateAgentStatusRequest struct {
	Status int `json:"status" binding:"required,oneof=1 2"` // 1=active, 2=disabled
}

func NewAgentHandler(agentService service.AgentService) *AgentHandler {
	return &AgentHandler{agentService: agentService}
}

func (h *AgentHandler) Create(c *gin.Context) {
	var request createAgentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	parentAgentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	result, err := h.agentService.Create(c.Request.Context(), service.CreateAgentInput{
		ParentAgentID: parentAgentID,
		Username:      request.Username,
		Password:      request.Password,
		RealName:      request.RealName,
		Phone:         request.Phone,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Created(c, "agent created", result)
}

func (h *AgentHandler) List(c *gin.Context) {
	parentAgentID, ok := middleware.GetAgentID(c)
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
		if value != model.AgentStatusActive && value != model.AgentStatusDisabled {
			utils.Error(c, utils.BadRequestError("status must be 1 (active) or 2 (disabled)"))
			return
		}
		status = &value
	}

	result, err := h.agentService.List(c.Request.Context(), service.AgentListInput{
		ParentAgentID: parentAgentID,
		Page:          page,
		PageSize:      pageSize,
		Status:        status,
		Keyword:       c.Query("keyword"),
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "agents fetched", result)
}

func (h *AgentHandler) Detail(c *gin.Context) {
	parentAgentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	agentID, err := parseUintParam(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.BadRequestError("invalid agent id"))
		return
	}

	result, err := h.agentService.Detail(c.Request.Context(), parentAgentID, agentID)
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "agent fetched", result)
}

func (h *AgentHandler) Update(c *gin.Context) {
	parentAgentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	agentID, err := parseUintParam(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.BadRequestError("invalid agent id"))
		return
	}

	var request updateAgentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	result, err := h.agentService.Update(c.Request.Context(), parentAgentID, agentID, service.UpdateAgentInput{
		Username: request.Username,
		Password: request.Password,
		RealName: request.RealName,
		Phone:    request.Phone,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "agent updated", result)
}

func (h *AgentHandler) UpdateStatus(c *gin.Context) {
	parentAgentID, ok := middleware.GetAgentID(c)
	if !ok {
		utils.Error(c, utils.UnauthorizedError("failed to resolve current user"))
		return
	}

	agentID, err := parseUintParam(c.Param("id"))
	if err != nil {
		utils.Error(c, utils.BadRequestError("invalid agent id"))
		return
	}

	var request updateAgentStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.Error(c, utils.BadRequestError(err.Error()))
		return
	}

	result, err := h.agentService.UpdateStatus(c.Request.Context(), parentAgentID, agentID, service.UpdateAgentStatusInput{
		Status: request.Status,
	})
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "agent status updated", result)
}
