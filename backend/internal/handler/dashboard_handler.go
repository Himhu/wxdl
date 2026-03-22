package handler

import (
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

func (h *DashboardHandler) Overview(c *gin.Context) {
	type result struct {
		Count int64  `json:"count"`
		Sum   string `json:"sum"`
	}

	var userTotal int64
	h.db.Table("users").Count(&userTotal)

	var agentTotal int64
	h.db.Table("agents").Count(&agentTotal)

	var agentActiveTotal int64
	h.db.Table("agents").Where("status = 1").Count(&agentActiveTotal)

	var cardTotal int64
	h.db.Table("cards").Count(&cardTotal)

	var cardUnusedTotal int64
	h.db.Table("cards").Where("status = 1").Count(&cardUnusedTotal)

	var cardUsedTotal int64
	h.db.Table("cards").Where("status = 2").Count(&cardUsedTotal)

	type balanceResult struct {
		Total float64
	}
	var bal balanceResult
	h.db.Table("agents").Select("COALESCE(SUM(balance), 0) as total").Scan(&bal)

	var pendingApplicationTotal int64
	h.db.Table("agent_applications").Where("status = 'pending'").Count(&pendingApplicationTotal)

	utils.Success(c, "ok", gin.H{
		"userTotal":               userTotal,
		"agentTotal":              agentTotal,
		"agentActiveTotal":        agentActiveTotal,
		"cardTotal":               cardTotal,
		"cardUnusedTotal":         cardUnusedTotal,
		"cardUsedTotal":           cardUsedTotal,
		"agentBalanceTotal":       bal.Total,
		"pendingApplicationTotal": pendingApplicationTotal,
	})
}
