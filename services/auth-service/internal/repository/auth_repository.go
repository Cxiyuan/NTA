package repository

import (
	"context"

	"github.com/Cxiyuan/NTA/pkg/models"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *AuthRepository) GetUserByID(ctx context.Context, userID uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, userID).Error
	return &user, err
}

func (r *AuthRepository) GetUserRoles(ctx context.Context, userID uint) ([]string, error) {
	var userRoles []models.UserRole
	var roles []string

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return nil, err
	}

	for _, ur := range userRoles {
		var role models.Role
		if err := r.db.WithContext(ctx).First(&role, ur.RoleID).Error; err == nil {
			roles = append(roles, role.Name)
		}
	}

	return roles, nil
}

func (r *AuthRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *AuthRepository) UpdateUser(ctx context.Context, userID uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (r *AuthRepository) DeleteUser(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, userID).Error
}

func (r *AuthRepository) ListUsers(ctx context.Context, limit, offset int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error
	return users, total, err
}

func (r *AuthRepository) AssignRole(ctx context.Context, userID, roleID uint, tenantID string) error {
	userRole := &models.UserRole{
		UserID:   userID,
		RoleID:   roleID,
		TenantID: tenantID,
	}
	return r.db.WithContext(ctx).Create(userRole).Error
}

func (r *AuthRepository) RemoveRole(ctx context.Context, userID, roleID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&models.UserRole{}).Error
}

func (r *AuthRepository) GetRoleByID(ctx context.Context, roleID uint) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).First(&role, roleID).Error
	return &role, err
}

func (r *AuthRepository) ListRoles(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	err := r.db.WithContext(ctx).Find(&roles).Error
	return roles, err
}

func (r *AuthRepository) CreateRole(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *AuthRepository) UpdateRole(ctx context.Context, roleID uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&models.Role{}).Where("id = ?", roleID).Updates(updates).Error
}

func (r *AuthRepository) DeleteRole(ctx context.Context, roleID uint) error {
	return r.db.WithContext(ctx).Delete(&models.Role{}, roleID).Error
}
