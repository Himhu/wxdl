package repository

import (
	"context"
	"errors"

	"backend/internal/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	FindByOpenID(ctx context.Context, openID string) (*model.User, error)
	FindByID(ctx context.Context, id uint64) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	List(ctx context.Context, query UserListQuery) ([]model.User, int64, error)
	CreateRoleChangeLog(ctx context.Context, log *model.UserRoleChangeLog) error
	FindInviteByCode(ctx context.Context, code string) (*model.UserInvite, error)
	FindActiveInviteByInviterID(ctx context.Context, inviterUserID uint64) (*model.UserInvite, error)
	CreateInvite(ctx context.Context, invite *model.UserInvite) error
	FindPendingApplicationByUserID(ctx context.Context, userID uint64) (*model.AgentApplication, error)
	CreateApplication(ctx context.Context, application *model.AgentApplication) error
	ListApplications(ctx context.Context, status string) ([]model.AgentApplication, error)
	ListApplicationsByInviter(ctx context.Context, inviterUserID uint64) ([]model.AgentApplication, error)
	FindApplicationByID(ctx context.Context, id uint64) (*model.AgentApplication, error)
	UpdateApplication(ctx context.Context, application *model.AgentApplication) error
}

type UserListQuery struct {
	Page     int
	PageSize int
	Role     string
	Keyword  string
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByOpenID(ctx context.Context, openID string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("open_id = ?", openID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) List(ctx context.Context, query UserListQuery) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	q := r.db.WithContext(ctx).Model(&model.User{})

	if query.Role != "" {
		q = q.Where("role = ?", query.Role)
	}
	if query.Keyword != "" {
		kw := "%" + query.Keyword + "%"
		q = q.Where("nickname LIKE ? OR open_id LIKE ? OR mobile LIKE ?", kw, kw, kw)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := q.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) CreateRoleChangeLog(ctx context.Context, log *model.UserRoleChangeLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *userRepository) FindInviteByCode(ctx context.Context, code string) (*model.UserInvite, error) {
	var invite model.UserInvite
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&invite).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *userRepository) FindActiveInviteByInviterID(ctx context.Context, inviterUserID uint64) (*model.UserInvite, error) {
	var invite model.UserInvite
	err := r.db.WithContext(ctx).
		Where("inviter_user_id = ? AND status = ?", inviterUserID, model.InviteStatusActive).
		Order("id DESC").
		First(&invite).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *userRepository) CreateInvite(ctx context.Context, invite *model.UserInvite) error {
	return r.db.WithContext(ctx).Create(invite).Error
}

func (r *userRepository) FindPendingApplicationByUserID(ctx context.Context, userID uint64) (*model.AgentApplication, error) {
	var application model.AgentApplication
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, model.ApplicationStatusPending).
		Order("id DESC").
		First(&application).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &application, nil
}

func (r *userRepository) CreateApplication(ctx context.Context, application *model.AgentApplication) error {
	return r.db.WithContext(ctx).Create(application).Error
}

func (r *userRepository) ListApplications(ctx context.Context, status string) ([]model.AgentApplication, error) {
	var applications []model.AgentApplication
	q := r.db.WithContext(ctx).Model(&model.AgentApplication{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if err := q.Order("created_at DESC").Find(&applications).Error; err != nil {
		return nil, err
	}
	return applications, nil
}

func (r *userRepository) ListApplicationsByInviter(ctx context.Context, inviterUserID uint64) ([]model.AgentApplication, error) {
	var applications []model.AgentApplication
	if err := r.db.WithContext(ctx).
		Where("inviter_user_id = ?", inviterUserID).
		Order("created_at DESC").
		Find(&applications).Error; err != nil {
		return nil, err
	}
	return applications, nil
}

func (r *userRepository) FindApplicationByID(ctx context.Context, id uint64) (*model.AgentApplication, error) {
	var application model.AgentApplication
	err := r.db.WithContext(ctx).First(&application, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &application, nil
}

func (r *userRepository) UpdateApplication(ctx context.Context, application *model.AgentApplication) error {
	return r.db.WithContext(ctx).Save(application).Error
}
