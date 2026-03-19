package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/utils"
)

// WeChatSettingView 微信配置视图
type WeChatSettingView struct {
	AppID        string     `json:"appId"`
	SecretMasked string     `json:"secretMasked"`
	Source       string     `json:"source"`
	Version      int        `json:"version"`
	UpdatedBy    *uint      `json:"updatedBy,omitempty"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	PublishedAt  *time.Time `json:"publishedAt,omitempty"`
}

// UpdateWeChatSettingsInput 更新微信配置输入
type UpdateWeChatSettingsInput struct {
	AppID      string
	AppSecret  string
	ChangeNote string
	ChangedBy  uint
	ChangedIP  string
}

// SystemSettingService 系统设置服务接口
type SystemSettingService interface {
	GetWeChatSettings(ctx context.Context) (*WeChatSettingView, error)
	UpdateWeChatSettings(ctx context.Context, input UpdateWeChatSettingsInput) (*WeChatSettingView, error)
}

type systemSettingService struct {
	repo      repository.SystemSettingRepository
	txManager repository.TxManager
	cipher    utils.SecretCipher
	runtime   RuntimeSettingsProvider
}

// NewSystemSettingService 创建系统设置服务
func NewSystemSettingService(
	repo repository.SystemSettingRepository,
	txManager repository.TxManager,
	cipher utils.SecretCipher,
	runtime RuntimeSettingsProvider,
) SystemSettingService {
	return &systemSettingService{
		repo:      repo,
		txManager: txManager,
		cipher:    cipher,
		runtime:   runtime,
	}
}

// GetWeChatSettings 获取微信配置
func (s *systemSettingService) GetWeChatSettings(ctx context.Context) (*WeChatSettingView, error) {
	settings, err := s.repo.GetWeChatSettings(ctx)
	if err != nil {
		return nil, err
	}

	view := &WeChatSettingView{
		Source: "database",
	}

	for _, setting := range settings {
		if setting.SettingKey == "app_id" && setting.ValuePlain != nil {
			view.AppID = *setting.ValuePlain
			view.Version = setting.Version
			view.UpdatedBy = setting.UpdatedBy
			view.UpdatedAt = setting.UpdatedAt
			view.PublishedAt = setting.PublishedAt
		}
		if setting.SettingKey == "app_secret" && setting.ValueMasked != nil {
			view.SecretMasked = *setting.ValueMasked
			if setting.Version > view.Version {
				view.Version = setting.Version
			}
		}
	}

	return view, nil
}

// UpdateWeChatSettings 更新微信配置
func (s *systemSettingService) UpdateWeChatSettings(ctx context.Context, input UpdateWeChatSettingsInput) (*WeChatSettingView, error) {
	if input.AppID == "" || input.AppSecret == "" {
		return nil, fmt.Errorf("appID and appSecret are required")
	}

	now := time.Now()

	// 加密 AppSecret
	ciphertext, masked, checksum, keyVersion, err := s.cipher.Encrypt(input.AppSecret, "wechat.app_secret")
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt app_secret: %w", err)
	}

	// 计算 AppID 的 checksum
	appIDChecksum := fmt.Sprintf("%x", sha256.Sum256([]byte(input.AppID)))

	// 更新 app_id
	appIDSetting := &model.SystemSetting{
		Category:    "wechat",
		SettingKey:  "app_id",
		DisplayName: "微信 AppID",
		ValueType:   "string",
		IsSecret:    false,
		ValuePlain:  &input.AppID,
		Checksum:    &appIDChecksum,
		Source:      "database",
		Status:      1,
		Description: "微信小程序 AppID",
		UpdatedBy:   &input.ChangedBy,
		PublishedAt: &now,
	}

	appIDRevision := &model.SystemSettingRevision{
		Category:       "wechat",
		SettingKey:     "app_id",
		Operation:      "update",
		NewValueMasked: &input.AppID,
		NewChecksum:    &appIDChecksum,
		ChangeNote:     input.ChangeNote,
		ChangedBy:      &input.ChangedBy,
		ChangedIP:      input.ChangedIP,
	}

	// 更新 app_secret
	appSecretSetting := &model.SystemSetting{
		Category:        "wechat",
		SettingKey:      "app_secret",
		DisplayName:     "微信 Secret",
		ValueType:       "secret",
		IsSecret:        true,
		ValueCiphertext: &ciphertext,
		ValueMasked:     &masked,
		Checksum:        &checksum,
		KeyVersion:      &keyVersion,
		Source:          "database",
		Status:          1,
		Description:     "微信小程序 AppSecret",
		UpdatedBy:       &input.ChangedBy,
		PublishedAt:     &now,
	}

	appSecretRevision := &model.SystemSettingRevision{
		Category:       "wechat",
		SettingKey:     "app_secret",
		Operation:      "update",
		NewValueMasked: &masked,
		NewChecksum:    &checksum,
		ChangeNote:     input.ChangeNote,
		ChangedBy:      &input.ChangedBy,
		ChangedIP:      input.ChangedIP,
	}

	if err := s.txManager.WithinTx(ctx, func(repos repository.TxRepositories) error {
		repo := repos.SystemSetting()

		if err := upsertSystemSettingWithRevision(ctx, repo, appIDSetting, appIDRevision); err != nil {
			return err
		}
		if err := upsertSystemSettingWithRevision(ctx, repo, appSecretSetting, appSecretRevision); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to update wechat settings: %w", err)
	}

	// 刷新运行时配置
	if err := s.runtime.Refresh(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh runtime config: %w", err)
	}

	return &WeChatSettingView{
		AppID:        input.AppID,
		SecretMasked: masked,
		Source:       "database",
		Version:      appSecretSetting.Version,
		UpdatedBy:    &input.ChangedBy,
		UpdatedAt:    now,
		PublishedAt:  &now,
	}, nil
}

// upsertSystemSettingWithRevision 在事务中更新或创建系统设置及其修订记录
func upsertSystemSettingWithRevision(
	ctx context.Context,
	repo repository.SystemSettingRepository,
	setting *model.SystemSetting,
	revision *model.SystemSettingRevision,
) error {
	existing, err := repo.GetByKeyForUpdate(ctx, setting.Category, setting.SettingKey)
	if err != nil {
		return err
	}

	if existing == nil {
		setting.Version = 1
		revision.Operation = "create"
		if err := repo.Create(ctx, setting); err != nil {
			return err
		}
		revision.SettingID = setting.ID
		revision.Version = setting.Version
		return repo.CreateRevision(ctx, revision)
	}

	revision.OldValueMasked = existing.ValueMasked
	revision.OldChecksum = existing.Checksum

	setting.ID = existing.ID
	setting.Version = existing.Version + 1

	if err := repo.Update(ctx, setting); err != nil {
		return err
	}

	revision.SettingID = existing.ID
	revision.Version = setting.Version

	return repo.CreateRevision(ctx, revision)
}

