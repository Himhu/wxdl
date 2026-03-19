package repository

import (
	"backend/internal/model"
	"gorm.io/gorm"
)

type MiniProgramConfigRepository struct {
	db *gorm.DB
}

func NewMiniProgramConfigRepository(db *gorm.DB) *MiniProgramConfigRepository {
	return &MiniProgramConfigRepository{db: db}
}

// List 查询配置列表
func (r *MiniProgramConfigRepository) List(namespace, scopeType, scopeCode string) ([]model.MiniProgramConfigItem, error) {
	var items []model.MiniProgramConfigItem
	query := r.db.Where("status = ?", 1)

	if namespace != "" {
		query = query.Where("namespace = ?", namespace)
	}
	if scopeType != "" {
		query = query.Where("scope_type = ?", scopeType)
	}
	if scopeCode != "" {
		query = query.Where("scope_code = ?", scopeCode)
	}

	err := query.Order("namespace, config_key").Find(&items).Error
	return items, err
}

// GetByID 根据ID查询
func (r *MiniProgramConfigRepository) GetByID(id uint) (*model.MiniProgramConfigItem, error) {
	var item model.MiniProgramConfigItem
	err := r.db.First(&item, id).Error
	return &item, err
}

// Create 创建配置
func (r *MiniProgramConfigRepository) Create(item *model.MiniProgramConfigItem) error {
	return r.db.Create(item).Error
}

// Update 更新配置
func (r *MiniProgramConfigRepository) Update(item *model.MiniProgramConfigItem) error {
	return r.db.Save(item).Error
}

// GetPublishedConfigs 获取已发布的配置（供小程序使用）
func (r *MiniProgramConfigRepository) GetPublishedConfigs(scopeType, scopeCode string) ([]model.MiniProgramConfigItem, error) {
	var items []model.MiniProgramConfigItem
	err := r.db.Where("status = ? AND scope_type = ? AND scope_code = ? AND visibility = ?",
		1, scopeType, scopeCode, "public").
		Find(&items).Error
	return items, err
}
