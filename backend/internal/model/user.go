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
	AgentLevel   *int       `json:"agentLevel" gorm:"column:agent_level;type:tinyint;default:null"`
	ParentUserID *uint64    `json:"parentUserId" gorm:"column:parent_user_id;default:null"`
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
	OldAgentLevel    *int      `json:"oldAgentLevel" gorm:"column:old_agent_level;type:tinyint"`
	NewAgentLevel    *int      `json:"newAgentLevel" gorm:"column:new_agent_level;type:tinyint"`
	ChangedByAdminID *uint64   `json:"changedByAdminId" gorm:"column:changed_by_admin_id"`
	Remark           string    `json:"remark" gorm:"column:remark;type:varchar(500);default:''"`
	CreatedAt        time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
}

func (UserRoleChangeLog) TableName() string {
	return "user_role_change_logs"
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

// AgentLevelName 根据等级返回名称
func AgentLevelName(level *int) string {
	if level == nil {
		return ""
	}
	switch *level {
	case 0:
		return "总代理"
	case 1:
		return "一级代理"
	case 2:
		return "二级代理"
	case 3:
		return "三级代理"
	default:
		return "代理商"
	}
}

// UserInfoResponse 返回给前端的用户信息
type UserInfoResponse struct {
	ID             uint64  `json:"id"`
	OpenID         string  `json:"openId"`
	Nickname       string  `json:"nickname"`
	Avatar         string  `json:"avatar"`
	Mobile         string  `json:"mobile"`
	Role           string  `json:"role"`
	AgentLevel     *int    `json:"agentLevel"`
	AgentLevelName string  `json:"agentLevelName"`
	ParentUserID   *uint64 `json:"parentUserId"`
	LastLoginAt    string  `json:"lastLoginAt"`
}

// ToResponse 转换为前端响应
func (u *User) ToResponse() UserInfoResponse {
	lastLogin := ""
	if u.LastLoginAt != nil {
		lastLogin = u.LastLoginAt.Format("2006-01-02T15:04:05Z")
	}
	return UserInfoResponse{
		ID:             u.ID,
		OpenID:         u.OpenID,
		Nickname:       u.Nickname,
		Avatar:         u.Avatar,
		Mobile:         u.Mobile,
		Role:           u.Role,
		AgentLevel:     u.AgentLevel,
		AgentLevelName: AgentLevelName(u.AgentLevel),
		ParentUserID:   u.ParentUserID,
		LastLoginAt:    lastLogin,
	}
}
