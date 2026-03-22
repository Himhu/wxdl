package handler

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/utils"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuditHandler struct {
	agentRepository repository.AgentRepository
	userRepository  repository.UserRepository
	db              *gorm.DB
}

func NewAuditHandler(agentRepository repository.AgentRepository, userRepository repository.UserRepository, db *gorm.DB) *AuditHandler {
	return &AuditHandler{agentRepository: agentRepository, userRepository: userRepository, db: db}
}

type auditLogItem struct {
	ID         uint   `json:"id"`
	Action     string `json:"action"`
	Detail     string `json:"detail"`
	Operator   string `json:"operator"`
	CreateTime string `json:"createTime"`
	SortTime   time.Time
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

	// 查登录日志
	loginLogs, _, _ := h.agentRepository.ListLoginLogs(c.Request.Context(), agentID, 1, 200)

	// 查余额变动日志
	var balanceLogs []model.PointsRecord
	h.db.WithContext(c.Request.Context()).
		Where("agent_id = ?", agentID).
		Order("id DESC").
		Limit(200).
		Find(&balanceLogs)

	// 合并
	var items []auditLogItem
	for _, log := range loginLogs {
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
		items = append(items, auditLogItem{
			ID:         log.ID,
			Action:     action,
			Detail:     detail,
			Operator:   log.Username,
			CreateTime: log.CreatedAt.Format("2006-01-02 15:04:05"),
			SortTime:   log.CreatedAt,
		})
	}
	for _, log := range balanceLogs {
		action := "余额变动"
		detail := fmt.Sprintf("%s %s", log.Amount.StringFixed(2), log.Description)
		items = append(items, auditLogItem{
			ID:         log.ID + 1000000, // 避免和登录日志 ID 冲突
			Action:     action,
			Detail:     detail,
			Operator:   "系统",
			CreateTime: log.CreatedAt.Format("2006-01-02 15:04:05"),
			SortTime:   log.CreatedAt,
		})
	}

	// 按时间倒序
	sort.Slice(items, func(i, j int) bool {
		return items[i].SortTime.After(items[j].SortTime)
	})

	// 分页
	total := int64(len(items))
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(items) {
		start = len(items)
	}
	if end > len(items) {
		end = len(items)
	}
	pageItems := items[start:end]

	list := make([]gin.H, 0, len(pageItems))
	for _, item := range pageItems {
		list = append(list, gin.H{
			"id":         item.ID,
			"action":     item.Action,
			"detail":     item.Detail,
			"operator":   item.Operator,
			"createTime": item.CreateTime,
		})
	}

	hasMore := int64(end) < total
	utils.Success(c, "ok", gin.H{
		"list":       list,
		"hasMore":    hasMore,
		"serverTime": time.Now().Format(time.RFC3339),
	})
}
