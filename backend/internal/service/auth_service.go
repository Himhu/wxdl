package service

import (
	"context"
	"errors"
	"strings"

	"backend/internal/config"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/utils"
)

const (
	loginTypePassword = 1
	loginTypeWechat   = 2
	loginStatusFailed = 0
	loginStatusOK     = 1
)

type LoginInput struct {
	Username  string
	Password  string
	ClientIP  string
	UserAgent string
}

type WechatLoginInput struct {
	Code      string
	ClientIP  string
	UserAgent string
}

type BindWechatInput struct {
	Code string
}

type AuthResult struct {
	Token string              `json:"token"`
	User  model.AgentResponse `json:"user"`
}

type AuthService interface {
	Login(ctx context.Context, input LoginInput) (*AuthResult, error)
	WechatLogin(ctx context.Context, input WechatLoginInput) (*AuthResult, error)
	BindWechat(ctx context.Context, agentID uint, input BindWechatInput) (*model.AgentResponse, error)
	Me(ctx context.Context, agentID uint) (*model.AgentResponse, error)
}

type WechatProvider interface {
	GetSession(ctx context.Context, code string) (*utils.WeChatSession, error)
}

type authService struct {
	agentRepository repository.AgentRepository
	jwtConfig       config.JWTConfig
	wechatClient    WechatProvider
}

func NewAuthService(agentRepository repository.AgentRepository, jwtConfig config.JWTConfig, wechatClient WechatProvider) AuthService {
	return &authService{
		agentRepository: agentRepository,
		jwtConfig:       jwtConfig,
		wechatClient:    wechatClient,
	}
}

func (s *authService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	username := strings.TrimSpace(input.Username)
	password := strings.TrimSpace(input.Password)

	agent, err := s.agentRepository.GetByUsername(ctx, username)
	if err != nil {
		return nil, utils.InternalError(err)
	}
	if agent == nil {
		s.recordLogin(ctx, nil, username, loginTypePassword, loginStatusFailed, input.ClientIP, input.UserAgent, "账号或密码错误")
		return nil, utils.UnauthorizedError("invalid username or password")
	}

	if agent.Status != model.AgentStatusActive {
		s.recordLogin(ctx, agent, username, loginTypePassword, loginStatusFailed, input.ClientIP, input.UserAgent, "账号已禁用")
		return nil, utils.UnauthorizedError("account is disabled")
	}

	if err := utils.CheckPassword(agent.Password, password); err != nil {
		s.recordLogin(ctx, agent, username, loginTypePassword, loginStatusFailed, input.ClientIP, input.UserAgent, "账号或密码错误")
		return nil, utils.UnauthorizedError("invalid username or password")
	}

	token, err := utils.GenerateTokenWithRole(s.jwtConfig, agent.ID, agent.Username, utils.RoleAgent)
	if err != nil {
		return nil, utils.InternalError(err)
	}

	s.recordLogin(ctx, agent, username, loginTypePassword, loginStatusOK, input.ClientIP, input.UserAgent, "")

	return &AuthResult{
		Token: token,
		User:  agent.ToResponse(),
	}, nil
}

func (s *authService) WechatLogin(ctx context.Context, input WechatLoginInput) (*AuthResult, error) {
	session, err := s.wechatClient.GetSession(ctx, strings.TrimSpace(input.Code))
	if err != nil {
		var wechatErr *utils.WeChatError
		if errors.As(err, &wechatErr) {
			s.recordLogin(ctx, nil, "wechat", loginTypeWechat, loginStatusFailed, input.ClientIP, input.UserAgent, wechatErr.Message)
			return nil, utils.BadRequestError("invalid wechat login code")
		}
		return nil, utils.InternalError(err)
	}
	if session.OpenID == "" {
		return nil, utils.BadRequestError("wechat openid is empty")
	}

	agent, err := s.agentRepository.GetByWechatOpenID(ctx, session.OpenID)
	if err != nil {
		return nil, utils.InternalError(err)
	}
	if agent == nil {
		s.recordLogin(ctx, nil, "wechat:"+session.OpenID, loginTypeWechat, loginStatusFailed, input.ClientIP, input.UserAgent, "微信账号未绑定")
		return nil, utils.UnauthorizedError("wechat account is not bound")
	}
	if agent.Status != model.AgentStatusActive {
		s.recordLogin(ctx, agent, agent.Username, loginTypeWechat, loginStatusFailed, input.ClientIP, input.UserAgent, "账号已禁用")
		return nil, utils.UnauthorizedError("account is disabled")
	}

	token, err := utils.GenerateTokenWithRole(s.jwtConfig, agent.ID, agent.Username, utils.RoleAgent)
	if err != nil {
		return nil, utils.InternalError(err)
	}

	s.recordLogin(ctx, agent, agent.Username, loginTypeWechat, loginStatusOK, input.ClientIP, input.UserAgent, "")

	return &AuthResult{
		Token: token,
		User:  agent.ToResponse(),
	}, nil
}

func (s *authService) BindWechat(ctx context.Context, agentID uint, input BindWechatInput) (*model.AgentResponse, error) {
	agent, err := s.agentRepository.GetByID(ctx, agentID)
	if err != nil {
		return nil, utils.InternalError(err)
	}
	if agent == nil {
		return nil, utils.NotFoundError("agent not found")
	}
	if agent.Status != model.AgentStatusActive {
		return nil, utils.UnauthorizedError("account is disabled")
	}

	session, err := s.wechatClient.GetSession(ctx, strings.TrimSpace(input.Code))
	if err != nil {
		var wechatErr *utils.WeChatError
		if errors.As(err, &wechatErr) {
			return nil, utils.BadRequestError("invalid wechat login code")
		}
		return nil, utils.InternalError(err)
	}
	if session.OpenID == "" {
		return nil, utils.BadRequestError("wechat openid is empty")
	}

	existingAgent, err := s.agentRepository.GetByWechatOpenID(ctx, session.OpenID)
	if err != nil {
		return nil, utils.InternalError(err)
	}
	if existingAgent != nil && existingAgent.ID != agent.ID {
		return nil, utils.ConflictError("wechat account already bound")
	}

	if err := s.agentRepository.BindWechat(ctx, agent.ID, session.OpenID, session.UnionID); err != nil {
		if utils.IsDuplicateKeyError(err) {
			return nil, utils.ConflictError("wechat account already bound")
		}
		return nil, utils.InternalError(err)
	}

	agent.WechatOpenID = session.OpenID
	agent.WechatUnionID = session.UnionID
	response := agent.ToResponse()
	return &response, nil
}

func (s *authService) Me(ctx context.Context, agentID uint) (*model.AgentResponse, error) {
	agent, err := s.agentRepository.GetByID(ctx, agentID)
	if err != nil {
		return nil, utils.InternalError(err)
	}
	if agent == nil {
		return nil, utils.NotFoundError("agent not found")
	}

	response := agent.ToResponse()
	return &response, nil
}

func (s *authService) recordLogin(ctx context.Context, agent *model.Agent, username string, loginType, status int, clientIP, userAgent, failReason string) {
	log := &model.LoginLog{
		Username:   username,
		LoginType:  loginType,
		Status:     status,
		IP:         clientIP,
		UserAgent:  userAgent,
		FailReason: failReason,
	}
	if agent != nil {
		agentID := agent.ID
		log.AgentID = &agentID
		if log.Username == "" {
			log.Username = agent.Username
		}
	}

	_ = s.agentRepository.CreateLoginLog(ctx, log)
}
