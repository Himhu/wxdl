package handler

import (
	"strconv"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AdminAgentHandler struct {
	agentRepository repository.AgentRepository
}

func NewAdminAgentHandler(agentRepository repository.AgentRepository) *AdminAgentHandler {
	return &AdminAgentHandler{agentRepository: agentRepository}
}

func (h *AdminAgentHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var status *int
	if raw := c.Query("status"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			utils.Error(c, utils.BadRequestError("invalid status"))
			return
		}
		status = &value
	}

	agents, total, err := h.agentRepository.ListAll(c.Request.Context(), repository.AdminAgentListQuery{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
		Keyword:  c.Query("keyword"),
	})
	if err != nil {
		utils.Error(c, utils.InternalError(err))
		return
	}

	list := make([]model.AgentResponse, 0, len(agents))
	for i := range agents {
		list = append(list, agents[i].ToResponse())
	}

	utils.Success(c, "agents fetched", gin.H{
		"list":     list,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (h *AdminAgentHandler) UpdateStatus(c *gin.Context) {
	agentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.Error(c, utils.BadRequestError("invalid agent id"))
		return
	}

	var req struct {
		Status int `json:"status" binding:"required,oneof=1 2"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, utils.BadRequestError("参数错误"))
		return
	}

	agent, err := h.agentRepository.GetByID(c.Request.Context(), uint(agentID))
	if err != nil {
		utils.Error(c, utils.InternalError(err))
		return
	}
	if agent == nil {
		utils.Error(c, utils.NotFoundError("agent not found"))
		return
	}

	if err := h.agentRepository.UpdateStatus(c.Request.Context(), agent.ID, req.Status); err != nil {
		utils.Error(c, utils.InternalError(err))
		return
	}
	agent.Status = req.Status

	utils.Success(c, "agent status updated", gin.H{
		"agent": agent.ToResponse(),
	})
}
