package handler

import (
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	agentRepository repository.AgentRepository
	userRepository  repository.UserRepository
}

func NewAuditHandler(agentRepository repository.AgentRepository, userRepository repository.UserRepository) *AuditHandler {
	return &AuditHandler{agentRepository: agentRepository, userRepository: userRepository}
}

func (h *AuditHandler) ListLogs(c *gin.Context) {
	var agentID uint
	if id, ok := middleware.GetAgentID(c); ok {
		agentID = id
	} else {
		userID, ok := middleware.GetUserID(c)
		if !ok {
			utils.Error(c, utils.ForbiddenError("insufficient permissions"))
			return
		}
		user, err := h.userRepository.FindByID(c.Request.Context(), uint64(userID))
		if err != nil {
			utils.Error(c, utils.ForbiddenError("insufficient permissions"))
			return
		}
		agent, err := h.agentRepository.GetByWechatOpenID(c.Request.Context(), user.OpenID)
		if err != nil || agent == nil {
			utils.Error(c, utils.ForbiddenError("insufficient permissions"))
			return
		}
		agentID = agent.ID
	}

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

	logs, total, err := h.agentRepository.ListLoginLogs(c.Request.Context(), agentID, page, pageSize)
	if err != nil {
		utils.Error(c, utils.InternalError(err))
		return
	}

	list := make([]gin.H, 0, len(logs))
	for _, log := range logs {
		action := "系统登录"
		if log.LoginType == 2 {
			action = "微信登录"
		}
		detail := "登录成功"
		if log.Status == 0 {
			if log.FailReason != "" {
				detail = log.FailReason
			} else {
				detail = "登录失败"
			}
		}
		list = append(list, gin.H{
			"id":         log.ID,
			"action":     action,
			"detail":     detail,
			"operator":   log.Username,
			"createTime": log.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	hasMore := int64(page*pageSize) < total
	utils.Success(c, "ok", gin.H{
		"list":       list,
		"hasMore":    hasMore,
		"serverTime": time.Now().Format(time.RFC3339),
	})
}
