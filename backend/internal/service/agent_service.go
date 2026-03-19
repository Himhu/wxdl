package service

import (
	"context"
	"strings"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/utils"

	"github.com/shopspring/decimal"
)

const (
	defaultAgentPage     = 1
	defaultAgentPageSize = 20
	maxAgentPageSize     = 100
)

type CreateAgentInput struct {
	ParentAgentID uint
	Username      string
	Password      string
	RealName      string
	Phone         string
}

type AgentListInput struct {
	ParentAgentID uint
	Page          int
	PageSize      int
	Status        *int
	Keyword       string
}

type UpdateAgentInput struct {
	Username string
	Password string
	RealName string
	Phone    string
}

type UpdateAgentStatusInput struct {
	Status int
}

type AgentListResult struct {
	List     []model.AgentResponse `json:"list"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"pageSize"`
}

type AgentService interface {
	Create(ctx context.Context, input CreateAgentInput) (*model.AgentResponse, error)
	List(ctx context.Context, input AgentListInput) (*AgentListResult, error)
	Detail(ctx context.Context, parentAgentID, agentID uint) (*model.AgentResponse, error)
	Update(ctx context.Context, parentAgentID, agentID uint, input UpdateAgentInput) (*model.AgentResponse, error)
	UpdateStatus(ctx context.Context, parentAgentID, agentID uint, input UpdateAgentStatusInput) (*model.AgentResponse, error)
	EnsureAgentActive(ctx context.Context, agentID uint) (*model.Agent, error)
}

type agentService struct {
	agentRepository repository.AgentRepository
}

func NewAgentService(agentRepository repository.AgentRepository) AgentService {
	return &agentService{agentRepository: agentRepository}
}

func (s *agentService) Create(ctx context.Context, input CreateAgentInput) (*model.AgentResponse, error) {
	parent, err := s.agentRepository.GetByID(ctx, input.ParentAgentID)
	if err != nil {
		return nil, utils.InternalError(err)
	}
	if parent == nil {
		return nil, utils.NotFoundError("parent agent not found")
	}
	if parent.Status != model.AgentStatusActive {
		return nil, utils.UnauthorizedError("account is disabled")
	}
	if parent.Level >= 3 {
		return nil, utils.BadRequestError("max agent level reached")
	}

	username := strings.TrimSpace(input.Username)
	if username == "" {
		return nil, utils.BadRequestError("username is required")
	}
	if existing, err := s.agentRepository.GetByUsername(ctx, username); err != nil {
		return nil, utils.InternalError(err)
	} else if existing != nil {
		return nil, utils.ConflictError("username already exists")
	}

	hashedPassword, err := utils.HashPassword(strings.TrimSpace(input.Password))
	if err != nil {
		return nil, utils.InternalError(err)
	}

	agent := &model.Agent{
		Username:  username,
		Password:  hashedPassword,
		RealName:  strings.TrimSpace(input.RealName),
		Phone:     strings.TrimSpace(input.Phone),
		Level:     parent.Level + 1,
		ParentID:  &parent.ID,
		Balance:   decimal.Zero,
		Status:    model.AgentStatusActive,
		StationID: parent.StationID,
	}
	if err := s.agentRepository.Create(ctx, agent); err != nil {
		return nil, utils.InternalError(err)
	}

	response := agent.ToResponse()
	return &response, nil
}

func (s *agentService) List(ctx context.Context, input AgentListInput) (*AgentListResult, error) {
	page := input.Page
	if page <= 0 {
		page = defaultAgentPage
	}
	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = defaultAgentPageSize
	}
	if pageSize > maxAgentPageSize {
		pageSize = maxAgentPageSize
	}
	if input.Status != nil && *input.Status != model.AgentStatusActive && *input.Status != model.AgentStatusDisabled {
		return nil, utils.BadRequestError("status must be 1 (active) or 2 (disabled)")
	}
	if err := s.ensureActiveParentAgent(ctx, input.ParentAgentID); err != nil {
		return nil, err
	}

	agents, total, err := s.agentRepository.ListDirectChildren(ctx, repository.AgentListQuery{
		ParentID: input.ParentAgentID,
		Page:     page,
		PageSize: pageSize,
		Status:   input.Status,
		Keyword:  strings.TrimSpace(input.Keyword),
	})
	if err != nil {
		return nil, utils.InternalError(err)
	}

	list := make([]model.AgentResponse, 0, len(agents))
	for index := range agents {
		list = append(list, agents[index].ToResponse())
	}

	return &AgentListResult{List: list, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *agentService) Detail(ctx context.Context, parentAgentID, agentID uint) (*model.AgentResponse, error) {
	if err := s.ensureActiveParentAgent(ctx, parentAgentID); err != nil {
		return nil, err
	}

	agent, err := s.agentRepository.GetDirectChildByID(ctx, parentAgentID, agentID)
	if err != nil {
		return nil, utils.InternalError(err)
	}
	if agent == nil {
		return nil, utils.NotFoundError("agent not found")
	}

	response := agent.ToResponse()
	return &response, nil
}

func (s *agentService) Update(ctx context.Context, parentAgentID, agentID uint, input UpdateAgentInput) (*model.AgentResponse, error) {
	if err := s.ensureActiveParentAgent(ctx, parentAgentID); err != nil {
		return nil, err
	}

	agent, err := s.agentRepository.GetDirectChildByID(ctx, parentAgentID, agentID)
	if err != nil {
		return nil, utils.InternalError(err)
	}
	if agent == nil {
		return nil, utils.NotFoundError("agent not found")
	}

	if username := strings.TrimSpace(input.Username); username != "" && username != agent.Username {
		existing, err := s.agentRepository.GetByUsername(ctx, username)
		if err != nil {
			return nil, utils.InternalError(err)
		}
		if existing != nil && existing.ID != agent.ID {
			return nil, utils.ConflictError("username already exists")
		}
		agent.Username = username
	}
	if password := strings.TrimSpace(input.Password); password != "" {
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			return nil, utils.InternalError(err)
		}
		agent.Password = hashedPassword
	}
	if input.RealName != "" {
		agent.RealName = strings.TrimSpace(input.RealName)
	}
	if input.Phone != "" {
		agent.Phone = strings.TrimSpace(input.Phone)
	}

	if err := s.agentRepository.Update(ctx, agent); err != nil {
		return nil, utils.InternalError(err)
	}

	response := agent.ToResponse()
	return &response, nil
}

func (s *agentService) UpdateStatus(ctx context.Context, parentAgentID, agentID uint, input UpdateAgentStatusInput) (*model.AgentResponse, error) {
	if input.Status != model.AgentStatusActive && input.Status != model.AgentStatusDisabled {
		return nil, utils.BadRequestError("status must be 1 (active) or 2 (disabled)")
	}

	if err := s.ensureActiveParentAgent(ctx, parentAgentID); err != nil {
		return nil, err
	}

	agent, err := s.agentRepository.GetDirectChildByID(ctx, parentAgentID, agentID)
	if err != nil {
		return nil, utils.InternalError(err)
	}
	if agent == nil {
		return nil, utils.NotFoundError("agent not found")
	}

	agent.Status = input.Status
	if err := s.agentRepository.UpdateStatus(ctx, agent.ID, input.Status); err != nil {
		return nil, utils.InternalError(err)
	}

	response := agent.ToResponse()
	return &response, nil
}


func (s *agentService) ensureActiveParentAgent(ctx context.Context, parentAgentID uint) error {
	_, err := s.EnsureAgentActive(ctx, parentAgentID)
	if err != nil {
		// 保持原有的错误消息以维持 API 稳定性
		if appErr, ok := err.(*utils.AppError); ok && appErr.Code == "NOT_FOUND" {
			return utils.NotFoundError("parent agent not found")
		}
	}
	return err
}

// EnsureAgentActive 验证代理是否存在且处于活跃状态
func (s *agentService) EnsureAgentActive(ctx context.Context, agentID uint) (*model.Agent, error) {
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
	return agent, nil
}
