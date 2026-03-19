package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type AgentResponse struct {
	ID          uint            `json:"id"`
	Username    string          `json:"username"`
	RealName    string          `json:"realName"`
	Phone       string          `json:"phone"`
	Level       int             `json:"level"`
	ParentID    *uint           `json:"parentId"`
	Balance     decimal.Decimal `json:"balance"`
	Status      int             `json:"status"`
	StationID   uint            `json:"stationId"`
	WechatBound bool            `json:"wechatBound"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

func (a *Agent) ToResponse() AgentResponse {
	return AgentResponse{
		ID:          a.ID,
		Username:    a.Username,
		RealName:    a.RealName,
		Phone:       a.Phone,
		Level:       a.Level,
		ParentID:    a.ParentID,
		Balance:     a.Balance,
		Status:      a.Status,
		StationID:   a.StationID,
		WechatBound: a.WechatOpenID != "",
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}
