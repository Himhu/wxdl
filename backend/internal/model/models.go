package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Agent 代理商模型
type Agent struct {
	ID            uint            `gorm:"primaryKey" json:"id"`
	Username      string          `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password      string          `gorm:"size:255;not null" json:"-"`
	RealName      string          `gorm:"size:50" json:"realName"`
	Phone         string          `gorm:"size:20" json:"phone"`
	Level         int             `gorm:"not null;default:1" json:"level"`
	ParentID      *uint           `gorm:"index" json:"parentId"`
	Balance       decimal.Decimal `gorm:"type:decimal(20,2);not null;default:0.00" json:"balance"`
	Status        int             `gorm:"not null;default:1" json:"status"`
	StationID     uint            `gorm:"index;default:0" json:"stationId"`
	WechatOpenID  string          `gorm:"uniqueIndex;size:100" json:"wechatOpenId"`
	WechatUnionID string          `gorm:"index;size:100" json:"wechatUnionId"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

func (Agent) TableName() string {
	return "agents"
}

// Card 卡密模型
type Card struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	CardKey     string          `gorm:"uniqueIndex;size:64;not null" json:"cardKey"`
	AgentID     uint            `gorm:"index;not null" json:"agentId"`
	StationID   uint            `gorm:"index;default:0" json:"stationId"`
	Quota       int             `gorm:"not null;default:0" json:"quota"`
	Cost        decimal.Decimal `gorm:"type:decimal(20,2);not null;default:0.00" json:"cost"`
	Status      int             `gorm:"not null;default:1" json:"status"`
	UsedAt      *time.Time      `json:"usedAt"`
	UsedBy      string          `gorm:"size:100" json:"usedBy"`
	DestroyedAt *time.Time      `json:"destroyedAt"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
}

func (Card) TableName() string {
	return "cards"
}

// PointsRecord 积分记录模型
type PointsRecord struct {
	ID            uint            `gorm:"primaryKey" json:"id"`
	AgentID       uint            `gorm:"index;not null" json:"agentId"`
	Type          int             `gorm:"not null" json:"type"`
	Amount        decimal.Decimal `gorm:"type:decimal(20,2);not null" json:"amount"`
	BalanceBefore decimal.Decimal `gorm:"type:decimal(20,2);not null" json:"balanceBefore"`
	BalanceAfter  decimal.Decimal `gorm:"type:decimal(20,2);not null" json:"balanceAfter"`
	Description   string          `gorm:"column:remark;size:500" json:"description"`
	RelatedID     *uint           `gorm:"index" json:"relatedId"`
	CreatedAt     time.Time       `json:"createdAt"`
}

func (PointsRecord) TableName() string {
	return "balance_logs"
}

// RechargeRequest 充值申请模型
type RechargeRequest struct {
	ID            uint            `gorm:"primaryKey" json:"id"`
	AgentID       uint            `gorm:"index;not null" json:"agentId"`
	Amount        decimal.Decimal `gorm:"type:decimal(20,2);not null" json:"amount"`
	Status        int             `gorm:"not null;default:0" json:"status"`
	PaymentMethod string          `gorm:"size:50" json:"paymentMethod"`
	PaymentProof  string          `gorm:"size:500" json:"paymentProof"`
	Remark        string          `gorm:"size:500" json:"remark"`
	ReviewedBy    *uint           `json:"reviewedBy"`
	ReviewedAt    *time.Time      `json:"reviewedAt"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
}

func (RechargeRequest) TableName() string {
	return "recharge_requests"
}

// LoginLog 登录日志模型
type LoginLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	AgentID    *uint     `gorm:"index" json:"agentId"`
	Username   string    `gorm:"size:50;not null" json:"username"`
	LoginType  int       `gorm:"not null" json:"loginType"`
	Status     int       `gorm:"not null" json:"status"`
	IP         string    `gorm:"size:50" json:"ip"`
	UserAgent  string    `gorm:"size:255" json:"userAgent"`
	FailReason string    `gorm:"size:255" json:"failReason"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (LoginLog) TableName() string {
	return "login_logs"
}

// Station 站点模型
type Station struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Code      string    `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Status    int       `gorm:"not null;default:1" json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Station) TableName() string {
	return "stations"
}
