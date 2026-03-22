package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo    repository.UserRepository
	agentRepo   repository.AgentRepository
	settingRepo repository.SystemSettingRepository
	cipher      utils.SecretCipher
}

func NewUserService(userRepo repository.UserRepository, agentRepo repository.AgentRepository, settingRepo repository.SystemSettingRepository, cipher utils.SecretCipher) *UserService {
	return &UserService{userRepo: userRepo, agentRepo: agentRepo, settingRepo: settingRepo, cipher: cipher}
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

func (s *UserService) GetAgentByID(ctx context.Context, agentID uint64) (*model.Agent, error) {
	return s.agentRepo.GetByID(ctx, uint(agentID))
}

func (s *UserService) GetAgentByOpenID(ctx context.Context, openID string) (*model.Agent, error) {
	return s.agentRepo.GetByWechatOpenID(ctx, openID)
}

func (s *UserService) GetPendingApplication(ctx context.Context, userID uint64) (*model.AgentApplication, error) {
	return s.userRepo.FindPendingApplicationByUserID(ctx, userID)
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
func (s *UserService) UpdateUserRole(ctx context.Context, userID uint64, newRole string, adminID uint64, remark string) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if newRole != model.UserRoleUser && newRole != model.UserRoleAgent {
		return nil, errors.New("无效的角色类型")
	}

	changeLog := &model.UserRoleChangeLog{
		UserID:           userID,
		OldRole:          user.Role,
		NewRole:          newRole,
		ChangedByAdminID: &adminID,
		Remark:           remark,
	}
	if err := s.userRepo.CreateRoleChangeLog(ctx, changeLog); err != nil {
		return nil, err
	}

	user.Role = newRole
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetOrCreateInvite(ctx context.Context, inviterUserID uint64) (*model.UserInvite, error) {
	invite, err := s.userRepo.FindActiveInviteByInviterID(ctx, inviterUserID)
	if err != nil {
		return nil, err
	}
	if invite != nil {
		return invite, nil
	}

	invite = &model.UserInvite{
		Code:          generateInviteCode(),
		InviterUserID: inviterUserID,
		Status:        model.InviteStatusActive,
	}
	if err := s.userRepo.CreateInvite(ctx, invite); err != nil {
		return nil, err
	}
	return invite, nil
}

func (s *UserService) ApplyAgent(ctx context.Context, userID uint64, inviteCode string) (*model.AgentApplication, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Role == model.UserRoleAgent {
		return nil, errors.New("当前用户已是代理")
	}

	pending, err := s.userRepo.FindPendingApplicationByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if pending != nil {
		return pending, nil
	}

	var inviterUserID *uint64
	if inviteCode != "" {
		invite, err := s.userRepo.FindInviteByCode(ctx, inviteCode)
		if err != nil {
			return nil, err
		}
		if invite == nil || invite.Status != model.InviteStatusActive {
			return nil, errors.New("邀请码无效")
		}
		inviterUserID = &invite.InviterUserID
		user.InviterUserID = inviterUserID
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, err
		}
	}

	application := &model.AgentApplication{
		UserID:        userID,
		InviterUserID: inviterUserID,
		InviteCode:    inviteCode,
		Status:        model.ApplicationStatusPending,
	}
	if err := s.userRepo.CreateApplication(ctx, application); err != nil {
		return nil, err
	}
	return application, nil
}

func (s *UserService) ListApplications(ctx context.Context, status string) ([]model.AgentApplicationListItem, error) {
	applications, err := s.userRepo.ListApplications(ctx, status)
	if err != nil {
		return nil, err
	}
	items := make([]model.AgentApplicationListItem, 0, len(applications))
	for _, application := range applications {
		applicant, err := s.userRepo.FindByID(ctx, application.UserID)
		if err != nil {
			return nil, err
		}
		var inviter *model.User
		if application.InviterUserID != nil {
			inviter, err = s.userRepo.FindByID(ctx, *application.InviterUserID)
			if err != nil {
				return nil, err
			}
		}
		items = append(items, model.NewAgentApplicationListItem(&application, applicant, inviter))
	}
	return items, nil
}

func (s *UserService) ListMyInvitedApplications(ctx context.Context, inviterUserID uint64) ([]model.AgentApplicationListItem, error) {
	applications, err := s.userRepo.ListApplicationsByInviter(ctx, inviterUserID)
	if err != nil {
		return nil, err
	}
	items := make([]model.AgentApplicationListItem, 0, len(applications))
	for _, application := range applications {
		applicant, err := s.userRepo.FindByID(ctx, application.UserID)
		if err != nil {
			return nil, err
		}
		inviter, err := s.userRepo.FindByID(ctx, inviterUserID)
		if err != nil {
			return nil, err
		}
		items = append(items, model.NewAgentApplicationListItem(&application, applicant, inviter))
	}
	return items, nil
}

func (s *UserService) ReviewApplication(ctx context.Context, applicationID uint64, approved bool, rejectReason string, adminID uint64) (*model.AgentApplication, error) {
	application, err := s.userRepo.FindApplicationByID(ctx, applicationID)
	if err != nil {
		return nil, err
	}
	if application == nil {
		return nil, errors.New("申请不存在")
	}
	if application.Status != model.ApplicationStatusPending {
		return nil, errors.New("该申请已处理")
	}

	now := time.Now()
	application.ReviewedByAdminID = &adminID
	application.ReviewedAt = &now
	if approved {
		application.Status = model.ApplicationStatusApproved
		user, err := s.userRepo.FindByID(ctx, application.UserID)
		if err != nil {
			return nil, err
		}
		user.Role = model.UserRoleAgent
		if application.InviterUserID != nil {
			user.InviterUserID = application.InviterUserID
		}
		if _, err := s.UpdateUserRole(ctx, user.ID, model.UserRoleAgent, adminID, "审批通过代理申请"); err != nil {
			return nil, err
		}
		if err := s.EnsureAgentAccount(ctx, user); err != nil {
			return nil, err
		}
	} else {
		application.Status = model.ApplicationStatusRejected
		application.RejectReason = rejectReason
	}

	if err := s.userRepo.UpdateApplication(ctx, application); err != nil {
		return nil, err
	}
	return application, nil
}

func (s *UserService) EnsureAgentAccount(ctx context.Context, user *model.User) error {
	if user == nil {
		return errors.New("用户不存在")
	}
	if user.OpenID == "" {
		return errors.New("用户缺少 openid")
	}

	existingAgent, err := s.agentRepo.GetByWechatOpenID(ctx, user.OpenID)
	if err != nil {
		return err
	}
	if existingAgent != nil {
		return nil
	}

	username := fmt.Sprintf("wx_%d", user.ID)
	if user.Mobile != "" {
		username = user.Mobile
	}
	if agentByUsername, err := s.agentRepo.GetByUsername(ctx, username); err != nil {
		return err
	} else if agentByUsername != nil {
		username = fmt.Sprintf("wx_%d", user.ID)
	}

	hashedPassword, err := utils.HashPassword(generateInviteCode())
	if err != nil {
		return err
	}

	agent := &model.Agent{
		Username:      username,
		Password:      hashedPassword,
		RealName:      user.Nickname,
		Phone:         user.Mobile,
		Level:         1,
		Status:        model.AgentStatusActive,
		Balance:       decimal.Zero,
		WechatOpenID:  user.OpenID,
		WechatUnionID: user.UnionID,
	}
	return s.agentRepo.Create(ctx, agent)
}

func generateInviteCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	buf := make([]byte, 8)
	for i := range buf {
		buf[i] = chars[r.Intn(len(chars))]
	}
	return fmt.Sprintf("%s", buf)
}

func (s *UserService) GetObjectStorageConfig(ctx context.Context) (*utils.ObjectStorageConfig, error) {
	settings, err := s.settingRepo.ListByCategory(ctx, "object_storage")
	if err != nil {
		return nil, err
	}

	cfg := &utils.ObjectStorageConfig{}
	for _, st := range settings {
		val := ""
		if st.ValuePlain != nil {
			val = *st.ValuePlain
		}
		switch st.SettingKey {
		case "enabled":
			cfg.Enabled = val == "true"
		case "provider":
			cfg.Provider = val
		case "endpoint":
			cfg.Endpoint = val
		case "bucket":
			cfg.Bucket = val
		case "access_key_id":
			cfg.AccessKeyID = val
		case "secret_key":
			if st.ValueCiphertext != nil && *st.ValueCiphertext != "" {
				decrypted, err := s.cipher.Decrypt(*st.ValueCiphertext, "oss.secret_key")
				if err == nil {
					cfg.SecretKey = decrypted
				}
			}
		case "region":
			cfg.Region = val
		case "custom_domain":
			cfg.CustomDomain = val
		case "path_prefix":
			cfg.PathPrefix = val
		}
	}
	return cfg, nil
}

func (s *UserService) UpdateAvatar(ctx context.Context, userID uint64, avatarURL string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	user.Avatar = avatarURL
	return s.userRepo.Update(ctx, user)
}
