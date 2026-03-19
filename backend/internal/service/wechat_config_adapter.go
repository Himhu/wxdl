package service

import (
	"context"
)

// weChatConfigAdapter 适配器，将 RuntimeSettingsProvider 转换为 WeChatConfigProvider
type weChatConfigAdapter struct {
	provider RuntimeSettingsProvider
}

// NewWeChatConfigAdapter 创建微信配置适配器
func NewWeChatConfigAdapter(provider RuntimeSettingsProvider) *weChatConfigAdapter {
	return &weChatConfigAdapter{provider: provider}
}

// GetWeChatConfig 实现 WeChatConfigProvider 接口
func (a *weChatConfigAdapter) GetWeChatConfig(ctx context.Context) (appID, appSecret, apiBase string, err error) {
	cfg, err := a.provider.GetWeChatConfig(ctx)
	if err != nil {
		return "", "", "", err
	}
	return cfg.AppID, cfg.AppSecret, cfg.APIBase, nil
}
