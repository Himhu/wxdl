package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
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

type WeChatAccessToken struct {
	Token     string `json:"access_token"`
	ExpiresIn int    `json:"expires_in"`
}

type MiniProgramCodeResponse struct {
	ContentType string
	Data        []byte
}

// WeChatConfigProvider 微信配置提供者接口
type WeChatConfigProvider interface {
	GetWeChatConfig(ctx context.Context) (appID, appSecret, apiBase string, err error)
}

type WeChatClient struct {
	provider   WeChatConfigProvider
	httpClient *http.Client

	tokenMu           sync.RWMutex
	accessToken       string
	accessTokenExpire time.Time
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

func (c *WeChatClient) GetAccessToken(ctx context.Context) (string, error) {
	c.tokenMu.RLock()
	if c.accessToken != "" && time.Now().Before(c.accessTokenExpire) {
		token := c.accessToken
		c.tokenMu.RUnlock()
		return token, nil
	}
	c.tokenMu.RUnlock()

	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.accessTokenExpire) {
		return c.accessToken, nil
	}

	appID, appSecret, apiBase, err := c.provider.GetWeChatConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get wechat config: %w", err)
	}
	if appID == "" || appSecret == "" {
		return "", fmt.Errorf("wechat app_id or app_secret is not configured")
	}
	if apiBase == "" {
		apiBase = defaultWeChatAPIBase
	}

	query := url.Values{}
	query.Set("appid", appID)
	query.Set("secret", appSecret)
	query.Set("grant_type", "client_credential")

	requestURL := strings.TrimRight(apiBase, "/") + "/cgi-bin/token?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var payload struct {
		WeChatAccessToken
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if resp.StatusCode >= http.StatusBadRequest || payload.ErrCode != 0 || payload.Token == "" {
		return "", &WeChatError{Code: payload.ErrCode, Message: payload.ErrMsg}
	}

	expireAt := time.Now().Add(time.Duration(payload.ExpiresIn-300) * time.Second)
	if payload.ExpiresIn <= 300 {
		expireAt = time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	}

	c.accessToken = payload.Token
	c.accessTokenExpire = expireAt
	return payload.Token, nil
}

func (c *WeChatClient) GetUnlimitedMiniProgramCode(ctx context.Context, scene string) (*MiniProgramCodeResponse, error) {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get wechat access token: %w", err)
	}

	_, _, apiBase, err := c.provider.GetWeChatConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get wechat config: %w", err)
	}
	if apiBase == "" {
		apiBase = defaultWeChatAPIBase
	}

	body, err := json.Marshal(map[string]any{
		"scene":      strings.TrimSpace(scene),
		"page":       "pages/common/login/index",
		"check_path": false,
		"width":      430,
	})
	if err != nil {
		return nil, err
	}

	requestURL := strings.TrimRight(apiBase, "/") + "/wxa/getwxacodeunlimit?access_token=" + url.QueryEscape(token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	contentType := strings.ToLower(strings.TrimSpace(resp.Header.Get("Content-Type")))
	if strings.Contains(contentType, "application/json") || (len(respBody) > 0 && respBody[0] == '{') {
		var apiErr WeChatError
		if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Code != 0 {
			return nil, fmt.Errorf("failed to generate mini program code: %w", &apiErr)
		}
		return nil, fmt.Errorf("failed to generate mini program code: unexpected wechat response")
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("failed to generate mini program code: wechat http status %d", resp.StatusCode)
	}
	if contentType == "" {
		contentType = "image/png"
	}

	return &MiniProgramCodeResponse{
		ContentType: contentType,
		Data:        respBody,
	}, nil
}
