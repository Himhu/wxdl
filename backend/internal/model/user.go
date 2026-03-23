package model

import "time"

// User 统一用户模型（微信登录自动注册）
type User struct {
	ID           uint64     `json:"id" gorm:"primaryKey;autoIncrement"`
	OpenID       string     `json:"openId" gorm:"column:open_id;type:varchar(128);uniqueIndex;not null"`
	UnionID      string     `json:"unionId" gorm:"column:union_id;type:varchar(128);default:''"`
	Nickname     string     `json:"nickname" gorm:"column:nickname;type:varchar(100);default:'微信用户'"`
	Avatar       string     `json:"avatar" gorm:"column:avatar;type:varchar(500);default:''"`
	Mobile       string     `json:"mobile" gorm:"column:mobile;type:varchar(20);default:''"`
	Role         string     `json:"role" gorm:"column:role;type:varchar(20);not null;default:'user'"`
	InviterUserID *uint64   `json:"inviterUserId" gorm:"column:inviter_user_id;default:null"`
	Status       int        `json:"status" gorm:"column:status;type:tinyint;not null;default:1"`
	LastLoginAt  *time.Time `json:"lastLoginAt" gorm:"column:last_login_at"`
	CreatedAt    time.Time  `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time  `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

func (User) TableName() string {
	return "users"
}

// UserRoleChangeLog 角色变更日志
type UserRoleChangeLog struct {
	ID               uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID           uint64    `json:"userId" gorm:"column:user_id;not null;index"`
	OldRole          string    `json:"oldRole" gorm:"column:old_role;type:varchar(20);not null"`
	NewRole          string    `json:"newRole" gorm:"column:new_role;type:varchar(20);not null"`
	ChangedByAdminID *uint64   `json:"changedByAdminId" gorm:"column:changed_by_admin_id"`
	Remark           string    `json:"remark" gorm:"column:remark;type:varchar(500);default:''"`
	CreatedAt        time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
}

func (UserRoleChangeLog) TableName() string {
	return "user_role_change_logs"
}

// UserInvite 邀请记录
type UserInvite struct {
	ID            uint64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Code          string     `json:"code" gorm:"column:code;type:varchar(32);uniqueIndex;not null"`
	InviterUserID uint64     `json:"inviterUserId" gorm:"column:inviter_user_id;not null;index"`
	Status        string     `json:"status" gorm:"column:status;type:varchar(20);not null;default:'active'"`
	ExpiresAt     *time.Time `json:"expiresAt" gorm:"column:expires_at"`
	CreatedAt     time.Time  `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time  `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

func (UserInvite) TableName() string {
	return "user_invites"
}

// AgentApplication 代理申请
 type AgentApplication struct {
	ID               uint64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID           uint64     `json:"userId" gorm:"column:user_id;not null;index"`
	InviterUserID    *uint64    `json:"inviterUserId" gorm:"column:inviter_user_id;index"`
	InviteCode       string     `json:"inviteCode" gorm:"column:invite_code;type:varchar(32);default:''"`
	Status           string     `json:"status" gorm:"column:status;type:varchar(20);not null;default:'pending'"`
	RejectReason     string     `json:"rejectReason" gorm:"column:reject_reason;type:varchar(500);default:''"`
	ReviewedByAdminID *uint64   `json:"reviewedByAdminId" gorm:"column:reviewed_by_admin_id"`
	ReviewedAt       *time.Time `json:"reviewedAt" gorm:"column:reviewed_at"`
	CreatedAt        time.Time  `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time  `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

func (AgentApplication) TableName() string {
	return "agent_applications"
}


type AgentApplicationListItem struct {
	ID          uint64            `json:"id"`
	InviteCode  string            `json:"inviteCode"`
	Status      string            `json:"status"`
	RejectReason string           `json:"rejectReason"`
	CreatedAt   string            `json:"createdAt"`
	ReviewedAt  string            `json:"reviewedAt"`
	Applicant   UserInfoResponse  `json:"applicant"`
	Inviter     *UserInfoResponse `json:"inviter,omitempty"`
}

func NewAgentApplicationListItem(application *AgentApplication, applicant *User, inviter *User) AgentApplicationListItem {
	reviewedAt := ""
	if application.ReviewedAt != nil {
		reviewedAt = application.ReviewedAt.Format("2006-01-02T15:04:05Z")
	}
	var inviterResp *UserInfoResponse
	if inviter != nil {
		resp := inviter.ToResponse()
		inviterResp = &resp
	}
	return AgentApplicationListItem{
		ID:           application.ID,
		InviteCode:   application.InviteCode,
		Status:       application.Status,
		RejectReason: application.RejectReason,
		CreatedAt:    application.CreatedAt.Format("2006-01-02T15:04:05Z"),
		ReviewedAt:   reviewedAt,
		Applicant:    applicant.ToResponse(),
		Inviter:      inviterResp,
	}
}

// 角色常量
const (
	UserRoleUser  = "user"
	UserRoleAgent = "agent"
)

// 状态常量
const (
	UserStatusActive   = 1
	UserStatusDisabled = 0
)

const (
	InviteStatusActive = "active"
)

const (
	ApplicationStatusPending  = "pending"
	ApplicationStatusApproved = "approved"
	ApplicationStatusRejected = "rejected"
)

// UserInfoResponse 返回给前端的用户信息
 type UserInfoResponse struct {
	ID            uint64  `json:"id"`
	OpenID        string  `json:"openId"`
	Nickname      string  `json:"nickname"`
	Avatar        string  `json:"avatar"`
	Mobile        string  `json:"mobile"`
	Role          string  `json:"role"`
	InviterUserID *uint64 `json:"inviterUserId"`
	LastLoginAt   string  `json:"lastLoginAt"`
	AgentBalance  string  `json:"agentBalance,omitempty"`
	AgentLevel    int     `json:"agentLevel,omitempty"`
}

// ToResponse 转换为前端响应
func (u *User) ToResponse() UserInfoResponse {
	lastLogin := ""
	if u.LastLoginAt != nil {
		lastLogin = u.LastLoginAt.Format("2006-01-02T15:04:05Z")
	}
	return UserInfoResponse{
		ID:            u.ID,
		OpenID:        u.OpenID,
		Nickname:      u.Nickname,
		Avatar:        u.Avatar,
		Mobile:        u.Mobile,
		Role:          u.Role,
		InviterUserID: u.InviterUserID,
		LastLoginAt:   lastLogin,
	}
}
