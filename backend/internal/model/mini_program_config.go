package model

import "time"

// MiniProgramConfigItem 小程序配置项
type MiniProgramConfigItem struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Namespace      string    `gorm:"size:50;not null" json:"namespace"`
	ConfigKey      string    `gorm:"size:100;not null" json:"configKey"`
	ScopeType      string    `gorm:"size:20;not null;default:'global'" json:"scopeType"`
	ScopeCode      string    `gorm:"size:64;not null;default:'default'" json:"scopeCode"`
	PublishedValue *string   `gorm:"type:text" json:"publishedValue"`
	Visibility     string    `gorm:"size:20;not null;default:'public'" json:"visibility"`
	Description    string    `gorm:"size:255" json:"description"`
	Status         int       `gorm:"not null;default:1" json:"status"`
	UpdatedBy      *uint     `json:"updatedBy"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (MiniProgramConfigItem) TableName() string {
	return "mini_program_config_items"
}
