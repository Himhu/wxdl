package service

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"backend/internal/config"
	"backend/internal/repository"
	"backend/internal/utils"
)

// RuntimeWeChatConfig 运行时微信配置
type RuntimeWeChatConfig struct {
	AppID          string
	AppSecret      string
	APIBase        string
	TimeoutSeconds int
	Source         string
	Version        int
	LoadedAt       time.Time
}

// RuntimeSettingsSnapshot 运行时配置快照
type RuntimeSettingsSnapshot struct {
	WeChat RuntimeWeChatConfig
}

// RuntimeSettingsProvider 运行时配置提供者接口
type RuntimeSettingsProvider interface {
	GetWeChatConfig(ctx context.Context) (*RuntimeWeChatConfig, error)
	Refresh(ctx context.Context) error
}

type runtimeSettingsProvider struct {
	repo     repository.SystemSettingRepository
	cipher   utils.SecretCipher
	fallback config.WeChatConfig
	snapshot atomic.Value
}

// NewRuntimeSettingsProvider 创建运行时配置提供者
func NewRuntimeSettingsProvider(
	repo repository.SystemSettingRepository,
	cipher utils.SecretCipher,
	fallback config.WeChatConfig,
) RuntimeSettingsProvider {
	return &runtimeSettingsProvider{
		repo:     repo,
		cipher:   cipher,
		fallback: fallback,
	}
}

// GetWeChatConfig 获取微信配置
func (p *runtimeSettingsProvider) GetWeChatConfig(ctx context.Context) (*RuntimeWeChatConfig, error) {
	if current, ok := p.snapshot.Load().(RuntimeSettingsSnapshot); ok && current.WeChat.AppID != "" {
		cfg := current.WeChat
		return &cfg, nil
	}

	if err := p.Refresh(ctx); err != nil {
		return nil, err
	}

	current := p.snapshot.Load().(RuntimeSettingsSnapshot)
	cfg := current.WeChat
	return &cfg, nil
}

// Refresh 刷新配置
func (p *runtimeSettingsProvider) Refresh(ctx context.Context) error {
	rows, err := p.repo.GetWeChatSettings(ctx)
	if err != nil || len(rows) == 0 {
		return p.refreshFromFallback(err)
	}

	var appID, appSecret string
	var version int

	for _, row := range rows {
		if row.SettingKey == "app_id" && row.ValuePlain != nil {
			appID = *row.ValuePlain
			version = row.Version
		}
		if row.SettingKey == "app_secret" && row.ValueCiphertext != nil {
			decrypted, err := p.cipher.Decrypt(*row.ValueCiphertext, "wechat.app_secret")
			if err != nil {
				return fmt.Errorf("failed to decrypt app_secret: %w", err)
			}
			appSecret = decrypted
			if row.Version > version {
				version = row.Version
			}
		}
	}

	if appID == "" || appSecret == "" {
		return p.refreshFromFallback(fmt.Errorf("incomplete wechat settings"))
	}

	snap := RuntimeSettingsSnapshot{
		WeChat: RuntimeWeChatConfig{
			AppID:          appID,
			AppSecret:      appSecret,
			APIBase:        p.fallback.APIBase,
			TimeoutSeconds: p.fallback.TimeoutSeconds,
			Source:         "database",
			Version:        version,
			LoadedAt:       time.Now(),
		},
	}

	p.snapshot.Store(snap)
	return nil
}

func (p *runtimeSettingsProvider) refreshFromFallback(err error) error {
	if p.fallback.AppID == "" || p.fallback.AppSecret == "" {
		return fmt.Errorf("no valid wechat config available: %w", err)
	}

	snap := RuntimeSettingsSnapshot{
		WeChat: RuntimeWeChatConfig{
			AppID:          p.fallback.AppID,
			AppSecret:      p.fallback.AppSecret,
			APIBase:        p.fallback.APIBase,
			TimeoutSeconds: p.fallback.TimeoutSeconds,
			Source:         "env",
			Version:        0,
			LoadedAt:       time.Now(),
		},
	}

	p.snapshot.Store(snap)
	return nil
}
