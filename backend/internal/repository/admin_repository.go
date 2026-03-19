package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"backend/internal/model"

	"gorm.io/gorm"
)

type AdminRepository interface {
	GetByUsername(ctx context.Context, username string) (*model.AdminUser, error)
	GetByID(ctx context.Context, id uint) (*model.AdminUser, error)
	UpdateLastLogin(ctx context.Context, id uint) error
	GetPermissions(ctx context.Context, adminID uint) ([]string, error)
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) GetByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	var admin model.AdminUser
	err := r.db.WithContext(ctx).
		Where("username = ?", strings.TrimSpace(username)).
		Take(&admin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &admin, nil
}

func (r *adminRepository) GetByID(ctx context.Context, id uint) (*model.AdminUser, error) {
	var admin model.AdminUser
	err := r.db.WithContext(ctx).Take(&admin, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &admin, nil
}

func (r *adminRepository) UpdateLastLogin(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&model.AdminUser{}).
		Where("id = ?", id).
		Update("last_login_at", time.Now()).Error
}

func (r *adminRepository) GetPermissions(ctx context.Context, adminID uint) ([]string, error) {
	var permissions []string
	err := r.db.WithContext(ctx).
		Table("admin_users AS au").
		Select("arp.permission_code").
		Joins("JOIN admin_roles AS ar ON ar.id = au.role_id").
		Joins("JOIN admin_role_permissions AS arp ON arp.role_id = ar.id").
		Where("au.id = ?", adminID).
		Where("au.status = ?", model.AdminStatusActive).
		Where("ar.status = ?", model.AdminStatusActive).
		Order("arp.permission_code ASC").
		Pluck("arp.permission_code", &permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}
