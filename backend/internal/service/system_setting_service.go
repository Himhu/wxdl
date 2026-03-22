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
	GetObjectStorageSettings(ctx context.Context) (*ObjectStorageSettingView, error)
	UpdateObjectStorageSettings(ctx context.Context, input UpdateObjectStorageSettingsInput) (*ObjectStorageSettingView, error)
	GetRedemptionSettings(ctx context.Context) (*RedemptionSettingView, error)
	UpdateRedemptionSettings(ctx context.Context, input UpdateRedemptionSettingsInput) (*RedemptionSettingView, error)
	GetRedemptionConfig(ctx context.Context) (*RedemptionConfig, error)
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

// ObjectStorageSettingView 对象存储配置视图
type ObjectStorageSettingView struct {
	Enabled         bool      `json:"enabled"`
	Provider        string    `json:"provider"`
	Endpoint        string    `json:"endpoint"`
	Bucket          string    `json:"bucket"`
	AccessKeyID     string    `json:"accessKeyId"`
	SecretKeyMasked string    `json:"secretKeyMasked"`
	Region          string    `json:"region"`
	CustomDomain    string    `json:"customDomain"`
	PathPrefix      string    `json:"pathPrefix"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// UpdateObjectStorageSettingsInput 更新对象存储配置输入
type UpdateObjectStorageSettingsInput struct {
	Enabled      bool
	Provider     string
	Endpoint     string
	Bucket       string
	AccessKeyID  string
	SecretKey    string
	Region       string
	CustomDomain string
	PathPrefix   string
	ChangeNote   string
	ChangedBy    uint
	ChangedIP    string
}

func (s *systemSettingService) GetObjectStorageSettings(ctx context.Context) (*ObjectStorageSettingView, error) {
	settings, err := s.repo.ListByCategory(ctx, "object_storage")
	if err != nil {
		return nil, err
	}

	view := &ObjectStorageSettingView{UpdatedAt: time.Now()}
	for _, st := range settings {
		val := ""
		if st.ValuePlain != nil {
			val = *st.ValuePlain
		}
		switch st.SettingKey {
		case "enabled":
			view.Enabled = val == "true"
		case "provider":
			view.Provider = val
		case "endpoint":
			view.Endpoint = val
		case "bucket":
			view.Bucket = val
		case "access_key_id":
			view.AccessKeyID = val
		case "secret_key":
			if st.ValueMasked != nil {
				view.SecretKeyMasked = *st.ValueMasked
			}
		case "region":
			view.Region = val
		case "custom_domain":
			view.CustomDomain = val
		case "path_prefix":
			view.PathPrefix = val
		}
		if st.UpdatedAt.After(view.UpdatedAt) {
			view.UpdatedAt = st.UpdatedAt
		}
	}
	return view, nil
}

func (s *systemSettingService) UpdateObjectStorageSettings(ctx context.Context, input UpdateObjectStorageSettingsInput) (*ObjectStorageSettingView, error) {
	now := time.Now()
	enabledStr := "false"
	if input.Enabled {
		enabledStr = "true"
	}

	plainFields := map[string]struct{ value, display, desc string }{
		"enabled":       {enabledStr, "是否开启", "对象存储开关"},
		"provider":      {input.Provider, "存储类型", "如 aliyun-oss"},
		"endpoint":      {input.Endpoint, "Endpoint", "对象存储节点地址"},
		"bucket":        {input.Bucket, "Bucket", "存储桶名称"},
		"access_key_id": {input.AccessKeyID, "AccessKeyID", "访问密钥 ID"},
		"region":        {input.Region, "Region", "地域"},
		"custom_domain": {input.CustomDomain, "自定义域名", "CDN 或自定义访问域名"},
		"path_prefix":   {input.PathPrefix, "目录前缀", "上传目录前缀"},
	}

	if err := s.txManager.WithinTx(ctx, func(repos repository.TxRepositories) error {
		repo := repos.SystemSetting()
		for key, info := range plainFields {
			v := info.value
			setting := &model.SystemSetting{
				Category: "object_storage", SettingKey: key,
				DisplayName: info.display, ValueType: "string",
				IsSecret: false, ValuePlain: &v,
				Source: "database", Status: 1, Description: info.desc,
				UpdatedBy: &input.ChangedBy, PublishedAt: &now,
			}
			revision := &model.SystemSettingRevision{
				Category: "object_storage", SettingKey: key, Operation: "update",
				NewValueMasked: &v, ChangeNote: input.ChangeNote,
				ChangedBy: &input.ChangedBy, ChangedIP: input.ChangedIP,
			}
			if err := upsertSystemSettingWithRevision(ctx, repo, setting, revision); err != nil {
				return err
			}
		}

		if input.SecretKey != "" {
			ct, masked, checksum, keyVer, err := s.cipher.Encrypt(input.SecretKey, "oss.secret_key")
			if err != nil {
				return fmt.Errorf("encrypt secret_key: %w", err)
			}
			setting := &model.SystemSetting{
				Category: "object_storage", SettingKey: "secret_key",
				DisplayName: "SecretKey", ValueType: "secret",
				IsSecret: true, ValueCiphertext: &ct, ValueMasked: &masked,
				Checksum: &checksum, KeyVersion: &keyVer,
				Source: "database", Status: 1, Description: "访问密钥 Secret",
				UpdatedBy: &input.ChangedBy, PublishedAt: &now,
			}
			revision := &model.SystemSettingRevision{
				Category: "object_storage", SettingKey: "secret_key", Operation: "update",
				NewValueMasked: &masked, NewChecksum: &checksum,
				ChangeNote: input.ChangeNote, ChangedBy: &input.ChangedBy, ChangedIP: input.ChangedIP,
			}
			if err := upsertSystemSettingWithRevision(ctx, repo, setting, revision); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("update object storage settings: %w", err)
	}

	return s.GetObjectStorageSettings(ctx)
}

// RedemptionSettingView 兑换码站点配置视图
type RedemptionSettingView struct {
	BaseURL                string    `json:"baseUrl"`
	AdminAccessTokenMasked string    `json:"adminAccessTokenMasked"`
	AdminUserID            string    `json:"adminUserId"`
	PriceRules             string    `json:"priceRules"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

// UpdateRedemptionSettingsInput 更新兑换码站点配置输入
type UpdateRedemptionSettingsInput struct {
	BaseURL          string
	AdminAccessToken string
	AdminUserID      string
	PriceRules       string
	ChangeNote       string
	ChangedBy        uint
	ChangedIP        string
}

// RedemptionConfig 兑换码站点运行时配置（含解密后的 token）
type RedemptionConfig struct {
	BaseURL          string
	AdminAccessToken string
	AdminUserID      string
	PriceRules       string
}

func (s *systemSettingService) GetRedemptionSettings(ctx context.Context) (*RedemptionSettingView, error) {
	settings, err := s.repo.ListByCategory(ctx, "redemption")
	if err != nil {
		return nil, err
	}

	view := &RedemptionSettingView{UpdatedAt: time.Now()}
	for _, setting := range settings {
		switch setting.SettingKey {
		case "base_url":
			if setting.ValuePlain != nil {
				view.BaseURL = *setting.ValuePlain
			}
		case "admin_user_id":
			if setting.ValuePlain != nil {
				view.AdminUserID = *setting.ValuePlain
			}
		case "admin_access_token":
			if setting.ValueMasked != nil {
				view.AdminAccessTokenMasked = *setting.ValueMasked
			}
		case "price_rules":
			if setting.ValuePlain != nil {
				view.PriceRules = *setting.ValuePlain
			}
		}
		if setting.UpdatedAt.After(view.UpdatedAt) || view.UpdatedAt.After(time.Now()) {
			view.UpdatedAt = setting.UpdatedAt
		}
	}
	return view, nil
}

func (s *systemSettingService) UpdateRedemptionSettings(ctx context.Context, input UpdateRedemptionSettingsInput) (*RedemptionSettingView, error) {
	now := time.Now()
	plainFields := map[string]struct{ value, display, desc string }{
		"base_url":      {input.BaseURL, "站点地址", "兑换码站点 BaseURL"},
		"admin_user_id": {input.AdminUserID, "管理员用户ID", "外部站点管理员 UserID"},
		"price_rules":   {input.PriceRules, "面值规则", "兑换码面值和成本价规则(JSON)"},
	}

	if err := s.txManager.WithinTx(ctx, func(repos repository.TxRepositories) error {
		repo := repos.SystemSetting()
		for key, info := range plainFields {
			v := info.value
			setting := &model.SystemSetting{
				Category: "redemption", SettingKey: key,
				DisplayName: info.display, ValueType: "string",
				IsSecret: false, ValuePlain: &v,
				Source: "database", Status: 1, Description: info.desc,
				UpdatedBy: &input.ChangedBy, PublishedAt: &now,
			}
			revision := &model.SystemSettingRevision{
				Category: "redemption", SettingKey: key, Operation: "update",
				NewValueMasked: &v, ChangeNote: input.ChangeNote,
				ChangedBy: &input.ChangedBy, ChangedIP: input.ChangedIP,
			}
			if err := upsertSystemSettingWithRevision(ctx, repo, setting, revision); err != nil {
				return err
			}
		}

		if input.AdminAccessToken != "" {
			ct, masked, checksum, keyVer, err := s.cipher.Encrypt(input.AdminAccessToken, "redemption.admin_access_token")
			if err != nil {
				return fmt.Errorf("encrypt admin_access_token: %w", err)
			}
			setting := &model.SystemSetting{
				Category: "redemption", SettingKey: "admin_access_token",
				DisplayName: "AccessToken", ValueType: "secret",
				IsSecret: true, ValueCiphertext: &ct, ValueMasked: &masked,
				Checksum: &checksum, KeyVersion: &keyVer,
				Source: "database", Status: 1, Description: "外部站点管理员 AccessToken",
				UpdatedBy: &input.ChangedBy, PublishedAt: &now,
			}
			revision := &model.SystemSettingRevision{
				Category: "redemption", SettingKey: "admin_access_token", Operation: "update",
				NewValueMasked: &masked, NewChecksum: &checksum,
				ChangeNote: input.ChangeNote, ChangedBy: &input.ChangedBy, ChangedIP: input.ChangedIP,
			}
			if err := upsertSystemSettingWithRevision(ctx, repo, setting, revision); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("update redemption settings: %w", err)
	}

	return s.GetRedemptionSettings(ctx)
}

func (s *systemSettingService) GetRedemptionConfig(ctx context.Context) (*RedemptionConfig, error) {
	settings, err := s.repo.ListByCategory(ctx, "redemption")
	if err != nil {
		return nil, err
	}

	cfg := &RedemptionConfig{}
	for _, setting := range settings {
		switch setting.SettingKey {
		case "base_url":
			if setting.ValuePlain != nil {
				cfg.BaseURL = *setting.ValuePlain
			}
		case "admin_user_id":
			if setting.ValuePlain != nil {
				cfg.AdminUserID = *setting.ValuePlain
			}
		case "admin_access_token":
			if setting.ValueCiphertext != nil && *setting.ValueCiphertext != "" {
				plaintext, err := s.cipher.Decrypt(*setting.ValueCiphertext, "redemption.admin_access_token")
				if err != nil {
					return nil, fmt.Errorf("decrypt admin_access_token: %w", err)
				}
				cfg.AdminAccessToken = plaintext
			}
		case "price_rules":
			if setting.ValuePlain != nil {
				cfg.PriceRules = *setting.ValuePlain
			}
		}
	}
	return cfg, nil
}

