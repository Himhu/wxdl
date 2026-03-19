package model

import "time"

// AdminRole 平台管理员角色模型
type AdminRole struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Code        string    `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`
	Status      int       `gorm:"not null;default:1" json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (AdminRole) TableName() string {
	return "admin_roles"
}

// AdminUser 平台管理员账号模型
type AdminUser struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Username    string     `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password    string     `gorm:"size:255;not null" json:"-"`
	RealName    string     `gorm:"size:50" json:"realName"`
	Status      int        `gorm:"not null;default:1" json:"status"`
	RoleID      uint       `gorm:"index;not null" json:"roleId"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func (AdminUser) TableName() string {
	return "admin_users"
}
