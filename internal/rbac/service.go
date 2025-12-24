package rbac

import (
	"encoding/json"
	"errors"

	"github.com/Cxiyuan/NTA/pkg/models"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateTenant(tenant *models.Tenant) error {
	return s.db.Create(tenant).Error
}

func (s *Service) GetTenant(tenantID string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := s.db.Where("tenant_id = ?", tenantID).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (s *Service) CreateUser(user *models.User) error {
	return s.db.Create(user).Error
}

func (s *Service) GetUser(username string) (*models.User, error) {
	var user models.User
	err := s.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) AssignRole(userID uint, roleID uint, tenantID string) error {
	userRole := &models.UserRole{
		UserID:   userID,
		RoleID:   roleID,
		TenantID: tenantID,
	}
	return s.db.Create(userRole).Error
}

func (s *Service) GetUserRoles(userID uint, tenantID string) ([]models.Role, error) {
	var roles []models.Role
	
	err := s.db.Table("roles").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND user_roles.tenant_id = ?", userID, tenantID).
		Find(&roles).Error
		
	if err != nil {
		return nil, err
	}
	
	return roles, nil
}

func (s *Service) HasPermission(userID uint, tenantID, resource, action string) (bool, error) {
	roles, err := s.GetUserRoles(userID, tenantID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		var permissions []models.Permission
		if err := json.Unmarshal([]byte(role.Permissions), &permissions); err != nil {
			continue
		}

		for _, perm := range permissions {
			if (perm.Resource == resource || perm.Resource == "*") &&
				(perm.Action == action || perm.Action == "*") {
				return true, nil
			}
		}
	}

	return false, nil
}

func (s *Service) CreateRole(role *models.Role) error {
	return s.db.Create(role).Error
}

func (s *Service) UpdateRole(roleID uint, updates map[string]interface{}) error {
	return s.db.Model(&models.Role{}).Where("id = ?", roleID).Updates(updates).Error
}

func (s *Service) DeleteRole(roleID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&models.UserRole{}).Error; err != nil {
			return err
		}
		
		if err := tx.Delete(&models.Role{}, roleID).Error; err != nil {
			return err
		}
		
		return nil
	})
}

func (s *Service) GetTenantUsers(tenantID string) ([]models.User, error) {
	var users []models.User
	err := s.db.Where("tenant_id = ?", tenantID).Find(&users).Error
	return users, err
}

func (s *Service) ValidateTenantAccess(userID uint, tenantID string) error {
	var user models.User
	err := s.db.Where("id = ? AND tenant_id = ?", userID, tenantID).First(&user).Error
	if err != nil {
		return errors.New("access denied: user not in tenant")
	}
	
	if user.Status != models.StatusActive {
		return errors.New("access denied: user account not active")
	}
	
	return nil
}

func (s *Service) InitializeDefaultRoles() error {
	roles := []models.Role{
		{
			Name:        models.RoleAdmin,
			Description: "Full system access",
			Permissions: `[{"resource":"*","action":"*"}]`,
		},
		{
			Name:        models.RoleAnalyst,
			Description: "Security analyst access",
			Permissions: `[
				{"resource":"alerts","action":"read"},
				{"resource":"alerts","action":"update"},
				{"resource":"assets","action":"read"},
				{"resource":"threat_intel","action":"read"},
				{"resource":"probes","action":"read"}
			]`,
		},
		{
			Name:        models.RoleViewer,
			Description: "Read-only access",
			Permissions: `[
				{"resource":"alerts","action":"read"},
				{"resource":"assets","action":"read"},
				{"resource":"probes","action":"read"}
			]`,
		},
	}

	for _, role := range roles {
		var existing models.Role
		err := s.db.Where("name = ?", role.Name).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := s.db.Create(&role).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
