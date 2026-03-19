package repository

import (
	"context"
	"errors"

	"backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SystemSettingRepository interface {
	GetByKey(ctx context.Context, category, key string) (*model.SystemSetting, error)
	GetByKeyForUpdate(ctx context.Context, category, key string) (*model.SystemSetting, error)
	ListByCategory(ctx context.Context, category string) ([]model.SystemSetting, error)
	GetWeChatSettings(ctx context.Context) ([]model.SystemSetting, error)

	Create(ctx context.Context, setting *model.SystemSetting) error
	Update(ctx context.Context, setting *model.SystemSetting) error
	CreateRevision(ctx context.Context, revision *model.SystemSettingRevision) error
}

type systemSettingRepository struct {
	db *gorm.DB
}

func NewSystemSettingRepository(db *gorm.DB) SystemSettingRepository {
	return &systemSettingRepository{db: db}
}

func (r *systemSettingRepository) GetByKey(ctx context.Context, category, key string) (*model.SystemSetting, error) {
	var setting model.SystemSetting
	err := r.db.WithContext(ctx).
		Where("category = ? AND setting_key = ? AND status = 1", category, key).
		First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *systemSettingRepository) ListByCategory(ctx context.Context, category string) ([]model.SystemSetting, error) {
	var settings []model.SystemSetting
	err := r.db.WithContext(ctx).
		Where("category = ? AND status = 1", category).
		Order("setting_key").
		Find(&settings).Error
	return settings, err
}

func (r *systemSettingRepository) GetWeChatSettings(ctx context.Context) ([]model.SystemSetting, error) {
	var settings []model.SystemSetting
	err := r.db.WithContext(ctx).
		Where("category = ? AND status = 1", "wechat").
		Find(&settings).Error
	return settings, err
}

func (r *systemSettingRepository) GetByKeyForUpdate(ctx context.Context, category, key string) (*model.SystemSetting, error) {
	var setting model.SystemSetting
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("category = ? AND setting_key = ?", category, key).
		First(&setting).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *systemSettingRepository) Create(ctx context.Context, setting *model.SystemSetting) error {
	return r.db.WithContext(ctx).Create(setting).Error
}

func (r *systemSettingRepository) Update(ctx context.Context, setting *model.SystemSetting) error {
	return r.db.WithContext(ctx).
		Model(&model.SystemSetting{}).
		Where("id = ?", setting.ID).
		Updates(map[string]interface{}{
			"display_name":     setting.DisplayName,
			"value_type":       setting.ValueType,
			"is_secret":        setting.IsSecret,
			"value_plain":      setting.ValuePlain,
			"value_ciphertext": setting.ValueCiphertext,
			"value_masked":     setting.ValueMasked,
			"checksum":         setting.Checksum,
			"key_version":      setting.KeyVersion,
			"source":           setting.Source,
			"status":           setting.Status,
			"description":      setting.Description,
			"version":          setting.Version,
			"updated_by":       setting.UpdatedBy,
			"published_at":     setting.PublishedAt,
		}).Error
}

func (r *systemSettingRepository) CreateRevision(ctx context.Context, revision *model.SystemSettingRevision) error {
	return r.db.WithContext(ctx).Create(revision).Error
}
