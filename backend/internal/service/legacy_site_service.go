package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"backend/internal/model"
)

type LegacyBalanceResult struct {
	LegacyUserID      int    `json:"legacyUserId"`
	LegacyUsername    string `json:"legacyUsername"`
	Balance           string `json:"balance"`
	Role              string `json:"role"`
	AgentLevelID      int    `json:"agentLevelId"`
	SVIPRemainingDays int    `json:"svipRemainingDays"`
	AgentType         string `json:"agentType"`
}

type LegacyAdminSession struct {
	client    *http.Client
	csrfToken string
}

type LegacyAgentSearchItem struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Status   string `json:"status"`
}

type LegacyConfirmResult struct {
	LegacyUserID      int         `json:"legacyUserId"`
	LegacyUsername    string      `json:"legacyUsername"`
	TransferredAmount string      `json:"transferredAmount"`
	LegacyDisabled    bool        `json:"legacyDisabled"`
	NewBalance        string      `json:"newBalance"`
	User              *model.User `json:"-"`
	Agent             *model.Agent `json:"-"`
}

type LegacySiteService struct {
	baseURL string
	timeout time.Duration
}

func NewLegacySiteService(baseURL string, timeoutSeconds int) *LegacySiteService {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 10
	}
	return &LegacySiteService{
		baseURL: strings.TrimRight(baseURL, "/"),
		timeout: time.Duration(timeoutSeconds) * time.Second,
	}
}

func (s *LegacySiteService) FetchBalance(ctx context.Context, username, password string) (*LegacyBalanceResult, error) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: s.timeout}

	csrfToken, err := s.fetchCSRFToken(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("获取旧站CSRF失败: %w", err)
	}

	result, err := s.login(ctx, client, csrfToken, username, password)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *LegacySiteService) AdminLogin(ctx context.Context, username, password string) (*LegacyAdminSession, error) {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: s.timeout}

	csrfToken, err := s.fetchCSRFToken(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("获取旧站管理员CSRF失败: %w", err)
	}
	if _, err := s.login(ctx, client, csrfToken, username, password); err != nil {
		return nil, fmt.Errorf("旧站管理员登录失败: %w", err)
	}
	return &LegacyAdminSession{client: client, csrfToken: csrfToken}, nil
}

func (s *LegacySiteService) SearchAgentByUsername(ctx context.Context, session *LegacyAdminSession, username string) (*LegacyAgentSearchItem, error) {
	if session == nil {
		return nil, fmt.Errorf("旧站管理员会话不存在")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/api/admin/agents?page=1&limit=10&search="+username, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-csrf-token", session.csrfToken)
	req.Header.Set("Origin", s.baseURL)
	req.Header.Set("Referer", s.baseURL+"/admin?tab=agents")

	resp, err := session.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("搜索旧站代理失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("旧站搜索代理失败: %d", resp.StatusCode)
	}

	var result struct {
		Agents []LegacyAgentSearchItem `json:"agents"`
		Data   []LegacyAgentSearchItem `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析旧站搜索结果失败: %w", err)
	}

	items := result.Agents
	if len(items) == 0 {
		items = result.Data
	}
	for _, item := range items {
		if item.Username == username {
			return &item, nil
		}
	}
	return nil, fmt.Errorf("未找到对应旧站代理")
}

func (s *LegacySiteService) DisableAgent(ctx context.Context, session *LegacyAdminSession, agentID int) error {
	if session == nil {
		return fmt.Errorf("旧站管理员会话不存在")
	}

	payload := `{"status":"disabled"}`
	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/api/admin/agents/%d/status", s.baseURL, agentID), strings.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-csrf-token", session.csrfToken)
	req.Header.Set("Origin", s.baseURL)
	req.Header.Set("Referer", s.baseURL+"/admin?tab=agents")

	resp, err := session.client.Do(req)
	if err != nil {
		return fmt.Errorf("禁用旧站代理失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("旧站禁用代理失败: %d %s", resp.StatusCode, string(body))
	}
	return nil
}

func (s *LegacySiteService) fetchCSRFToken(ctx context.Context, client *http.Client) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/api/csrf-token", nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var body struct {
		CSRFToken string `json:"csrfToken"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("解析CSRF响应失败: %w", err)
	}
	if body.CSRFToken == "" {
		return "", fmt.Errorf("旧站未返回CSRF token")
	}
	return body.CSRFToken, nil
}

func (s *LegacySiteService) login(ctx context.Context, client *http.Client, csrfToken, username, password string) (*LegacyBalanceResult, error) {
	payload := fmt.Sprintf(`{"username":%q,"password":%q}`, username, password)
	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/api/auth/login", strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-csrf-token", csrfToken)
	req.Header.Set("Origin", s.baseURL)
	req.Header.Set("Referer", s.baseURL+"/login")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("旧站登录请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取旧站响应失败: %w", err)
	}

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("旧站账号或密码错误")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("旧站返回异常状态: %d", resp.StatusCode)
	}

	var loginResp struct {
		User struct {
			ID                int    `json:"id"`
			Username          string `json:"username"`
			Balance           string `json:"balance"`
			Role              string `json:"role"`
			AgentLevelID      int    `json:"agent_level_id"`
			SVIPRemainingDays int    `json:"svip_remaining_days"`
		} `json:"user"`
	}
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return nil, fmt.Errorf("解析旧站登录响应失败: %w", err)
	}
	if loginResp.User.ID == 0 {
		return nil, fmt.Errorf("旧站账号或密码错误")
	}

	agentType := "普通代理"
	if loginResp.User.AgentLevelID == 2 {
		agentType = "VIP代理"
	}

	return &LegacyBalanceResult{
		LegacyUserID:      loginResp.User.ID,
		LegacyUsername:    loginResp.User.Username,
		Balance:           loginResp.User.Balance,
		Role:              loginResp.User.Role,
		AgentLevelID:      loginResp.User.AgentLevelID,
		SVIPRemainingDays: loginResp.User.SVIPRemainingDays,
		AgentType:         agentType,
	}, nil
}
