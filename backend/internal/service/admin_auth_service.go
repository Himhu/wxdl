package service

import (
	"context"
	"strings"
	"time"

	"backend/internal/config"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/utils"
)

const adminStatusActive = 1

type AdminUserResponse struct {
	ID          uint       `json:"id"`
	Username    string     `json:"username"`
	RealName    string     `json:"realName"`
	Status      int        `json:"status"`
	RoleID      uint       `json:"roleId"`
	Permissions []string   `json:"permissions,omitempty"`
	LastLoginAt *time.Time `json:"lastLoginAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type AdminAuthResult struct {
	Token string            `json:"token"`
	User  AdminUserResponse `json:"user"`
}

type AdminAuthService interface {
	Login(ctx context.Context, username, password, clientIP string) (*AdminAuthResult, error)
	Me(ctx context.Context, adminID uint) (*AdminUserResponse, error)
}

type adminAuthService struct {
	adminRepository repository.AdminRepository
	jwtConfig       config.JWTConfig
}

func NewAdminAuthService(adminRepository repository.AdminRepository, jwtConfig config.JWTConfig) AdminAuthService {
	return &adminAuthService{
		adminRepository: adminRepository,
		jwtConfig:       jwtConfig,
	}
}

func (s *adminAuthService) Login(ctx context.Context, username, password, clientIP string) (*AdminAuthResult, error) {
	_ = clientIP

	normalizedUsername := strings.TrimSpace(username)
	normalizedPassword := strings.TrimSpace(password)

	admin, err := s.adminRepository.GetByUsername(ctx, normalizedUsername)
	if err != nil {
		return nil, utils.InternalError(err)
	}
	if admin == nil {
		return nil, utils.UnauthorizedError("invalid username or password")
	}
	if admin.Status != adminStatusActive {
		return nil, utils.UnauthorizedError("account is disabled")
	}
	if err := utils.CheckPassword(admin.Password, normalizedPassword); err != nil {
		return nil, utils.UnauthorizedError("invalid username or password")
	}

	token, err := utils.GenerateTokenWithRole(s.jwtConfig, admin.ID, admin.Username, utils.RoleAdmin)
	if err != nil {
		return nil, utils.InternalError(err)
	}

	if err := s.adminRepository.UpdateLastLogin(ctx, admin.ID); err != nil {
		return nil, utils.InternalError(err)
	}

	permissions, err := s.adminRepository.GetPermissions(ctx, admin.ID)
	if err != nil {
		return nil, utils.InternalError(err)
	}

	now := time.Now()
	admin.LastLoginAt = &now

	return &AdminAuthResult{
		Token: token,
		User:  toAdminUserResponse(admin, permissions),
	}, nil
}

func (s *adminAuthService) Me(ctx context.Context, adminID uint) (*AdminUserResponse, error) {
	admin, err := s.adminRepository.GetByID(ctx, adminID)
	if err != nil {
		return nil, utils.InternalError(err)
	}
	if admin == nil {
		return nil, utils.NotFoundError("admin user not found")
	}

	permissions, err := s.adminRepository.GetPermissions(ctx, adminID)
	if err != nil {
		return nil, utils.InternalError(err)
	}

	response := toAdminUserResponse(admin, permissions)
	return &response, nil
}

func toAdminUserResponse(admin *model.AdminUser, permissions []string) AdminUserResponse {
	return AdminUserResponse{
		ID:          admin.ID,
		Username:    admin.Username,
		RealName:    admin.RealName,
		Status:      admin.Status,
		RoleID:      admin.RoleID,
		Permissions: permissions,
		LastLoginAt: admin.LastLoginAt,
		CreatedAt:   admin.CreatedAt,
		UpdatedAt:   admin.UpdatedAt,
	}
}
