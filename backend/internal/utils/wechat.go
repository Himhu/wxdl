package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultWeChatAPIBase = "https://api.weixin.qq.com"

type WeChatSession struct {
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
	SessionKey string `json:"session_key"`
}

type WeChatError struct {
	Code    int    `json:"errcode"`
	Message string `json:"errmsg"`
}

func (e *WeChatError) Error() string {
	if e == nil {
		return "wechat api error"
	}
	return fmt.Sprintf("wechat api error %d: %s", e.Code, e.Message)
}

// WeChatConfigProvider 微信配置提供者接口
type WeChatConfigProvider interface {
	GetWeChatConfig(ctx context.Context) (appID, appSecret, apiBase string, err error)
}

type WeChatClient struct {
	provider   WeChatConfigProvider
	httpClient *http.Client
}

func NewWeChatClient(provider WeChatConfigProvider, timeoutSeconds int) *WeChatClient {
	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &WeChatClient{
		provider:   provider,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *WeChatClient) GetSession(ctx context.Context, code string) (*WeChatSession, error) {
	appID, appSecret, apiBase, err := c.provider.GetWeChatConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get wechat config: %w", err)
	}

	if appID == "" || appSecret == "" {
		return nil, fmt.Errorf("wechat app_id or app_secret is not configured")
	}

	if apiBase == "" {
		apiBase = defaultWeChatAPIBase
	}

	query := url.Values{}
	query.Set("appid", appID)
	query.Set("secret", appSecret)
	query.Set("js_code", strings.TrimSpace(code))
	query.Set("grant_type", "authorization_code")

	requestURL := strings.TrimRight(apiBase, "/") + "/sns/jscode2session?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload struct {
		WeChatSession
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest || payload.ErrCode != 0 {
		return nil, &WeChatError{Code: payload.ErrCode, Message: payload.ErrMsg}
	}

	return &payload.WeChatSession, nil
}
