package service

import (
	"context"
	"errors"
	"time"

	"backend/internal/model"
	"backend/internal/repository"

	"gorm.io/gorm"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// WechatLoginOrRegister 微信登录（自动注册）
func (s *UserService) WechatLoginOrRegister(ctx context.Context, openID, unionID, nickname, avatar string) (*model.User, error) {
	user, err := s.userRepo.FindByOpenID(ctx, openID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	now := time.Now()

	if user != nil {
		// 已注册 — 更新微信信息和登录时间
		if nickname != "" {
			user.Nickname = nickname
		}
		if avatar != "" {
			user.Avatar = avatar
		}
		user.LastLoginAt = &now
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, err
		}
		return user, nil
	}

	// 新用户 — 自动注册为普通用户
	if nickname == "" {
		nickname = "微信用户"
	}
	newUser := &model.User{
		OpenID:      openID,
		UnionID:     unionID,
		Nickname:    nickname,
		Avatar:      avatar,
		Role:        model.UserRoleUser,
		Status:      model.UserStatusActive,
		LastLoginAt: &now,
	}
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}
	return newUser, nil
}

// GetProfile 获取用户信息
func (s *UserService) GetProfile(ctx context.Context, userID uint64) (*model.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

// ListUsers 分页查询用户列表（管理端）
func (s *UserService) ListUsers(ctx context.Context, page, pageSize int, role, keyword string) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.userRepo.List(ctx, repository.UserListQuery{
		Page:     page,
		PageSize: pageSize,
		Role:     role,
		Keyword:  keyword,
	})
}

// UpdateUserRole 管理员修改用户角色
func (s *UserService) UpdateUserRole(ctx context.Context, userID uint64, newRole string, agentLevel *int, parentUserID *uint64, adminID uint64, remark string) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if newRole != model.UserRoleUser && newRole != model.UserRoleAgent {
		return nil, errors.New("无效的角色类型")
	}

	// 记录变更日志
	changeLog := &model.UserRoleChangeLog{
		UserID:           userID,
		OldRole:          user.Role,
		NewRole:          newRole,
		OldAgentLevel:    user.AgentLevel,
		NewAgentLevel:    agentLevel,
		ChangedByAdminID: &adminID,
		Remark:           remark,
	}
	if err := s.userRepo.CreateRoleChangeLog(ctx, changeLog); err != nil {
		return nil, err
	}

	// 更新用户角色
	user.Role = newRole
	if newRole == model.UserRoleAgent {
		user.AgentLevel = agentLevel
		user.ParentUserID = parentUserID
	} else {
		user.AgentLevel = nil
		user.ParentUserID = nil
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
