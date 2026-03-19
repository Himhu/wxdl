package repository

import (
	"context"

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
