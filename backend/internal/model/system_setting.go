package model

import "time"

type SystemSetting struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Category        string     `gorm:"size:50;not null;uniqueIndex:uk_system_setting" json:"category"`
	SettingKey      string     `gorm:"size:100;not null;uniqueIndex:uk_system_setting" json:"settingKey"`
	DisplayName     string     `gorm:"size:100;not null" json:"displayName"`
	ValueType       string     `gorm:"size:20;not null" json:"valueType"`
	IsSecret        bool       `gorm:"not null;default:false" json:"isSecret"`
	ValuePlain      *string    `gorm:"type:text" json:"-"`
	ValueCiphertext *string    `gorm:"type:text" json:"-"`
	ValueMasked     *string    `gorm:"size:255" json:"valueMasked"`
	Checksum        *string    `gorm:"size:64" json:"-"`
	KeyVersion      *string    `gorm:"size:32" json:"keyVersion,omitempty"`
	Source          string     `gorm:"size:20;not null;default:'database'" json:"source"`
	Status          int        `gorm:"not null;default:1" json:"status"`
	Description     string     `gorm:"size:255" json:"description"`
	Version         int        `gorm:"not null;default:1" json:"version"`
	UpdatedBy       *uint      `json:"updatedBy"`
	PublishedAt     *time.Time `json:"publishedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func (SystemSetting) TableName() string {
	return "system_settings"
}

type SystemSettingRevision struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	SettingID      uint      `gorm:"not null" json:"settingId"`
	Category       string    `gorm:"size:50;not null" json:"category"`
	SettingKey     string    `gorm:"size:100;not null" json:"settingKey"`
	Version        int       `gorm:"not null" json:"version"`
	Operation      string    `gorm:"size:20;not null" json:"operation"`
	OldValueMasked *string   `gorm:"size:255" json:"oldValueMasked,omitempty"`
	NewValueMasked *string   `gorm:"size:255" json:"newValueMasked,omitempty"`
	OldChecksum    *string   `gorm:"size:64" json:"oldChecksum,omitempty"`
	NewChecksum    *string   `gorm:"size:64" json:"newChecksum,omitempty"`
	ChangeNote     string    `gorm:"size:255" json:"changeNote"`
	ChangedBy      *uint     `json:"changedBy"`
	ChangedIP      string    `gorm:"size:64" json:"changedIp"`
	CreatedAt      time.Time `json:"createdAt"`
}

func (SystemSettingRevision) TableName() string {
	return "system_setting_revisions"
}
