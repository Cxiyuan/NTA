package api

import (
	"net/http"
	"strconv"

	"github.com/Cxiyuan/NTA/internal/rbac"
	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) listUsers(c *gin.Context) {
	var users []models.User
	if err := s.db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (s *Server) createUser(c *gin.Context) {
	var req struct {
		Username string   `json:"username" binding:"required"`
		Email    string   `json:"email" binding:"required,email"`
		Password string   `json:"password" binding:"required,min=8"`
		TenantID string   `json:"tenant_id" binding:"required"`
		Status   string   `json:"status" binding:"required"`
		RoleIDs  []uint   `json:"role_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		TenantID:     req.TenantID,
		Status:       req.Status,
	}

	if err := s.db.Create(user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	rbacService := rbac.NewService(s.db)
	for _, roleID := range req.RoleIDs {
		rbacService.AssignRole(user.ID, roleID, req.TenantID)
	}

	username, _ := c.Get("username")
	s.auditService.Log(username.(string), "create_user", user.Username, map[string]interface{}{
		"user_id": user.ID,
	})

	c.JSON(http.StatusOK, user)
}

func (s *Server) updateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req struct {
		Email   string `json:"email"`
		Status  string `json:"status"`
		RoleIDs []uint `json:"role_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if err := s.db.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	username, _ := c.Get("username")
	s.auditService.Log(username.(string), "update_user", strconv.FormatUint(id, 10), updates)

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (s *Server) deleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := s.db.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	username, _ := c.Get("username")
	s.auditService.Log(username.(string), "delete_user", strconv.FormatUint(id, 10), nil)

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (s *Server) resetUserPassword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	newPassword := generateRandomPassword(12)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := s.db.Model(&models.User{}).Where("id = ?", id).Update("password_hash", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset password"})
		return
	}

	username, _ := c.Get("username")
	s.auditService.Log(username.(string), "reset_password", strconv.FormatUint(id, 10), nil)

	c.JSON(http.StatusOK, gin.H{"new_password": newPassword})
}

func (s *Server) listRoles(c *gin.Context) {
	var roles []models.Role
	if err := s.db.Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list roles"})
		return
	}
	c.JSON(http.StatusOK, roles)
}

func (s *Server) createRole(c *gin.Context) {
	var role models.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rbacService := rbac.NewService(s.db)
	if err := rbacService.CreateRole(&role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create role"})
		return
	}

	username, _ := c.Get("username")
	s.auditService.Log(username.(string), "create_role", role.Name, map[string]interface{}{
		"role_id": role.ID,
	})

	c.JSON(http.StatusOK, role)
}

func (s *Server) updateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rbacService := rbac.NewService(s.db)
	if err := rbacService.UpdateRole(uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}

	username, _ := c.Get("username")
	s.auditService.Log(username.(string), "update_role", strconv.FormatUint(id, 10), updates)

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (s *Server) deleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	rbacService := rbac.NewService(s.db)
	if err := rbacService.DeleteRole(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete role"})
		return
	}

	username, _ := c.Get("username")
	s.auditService.Log(username.(string), "delete_role", strconv.FormatUint(id, 10), nil)

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (s *Server) updateRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role id"})
		return
	}

	var req struct {
		Permissions string `json:"permissions" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.db.Model(&models.Role{}).Where("id = ?", id).Update("permissions", req.Permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update permissions"})
		return
	}

	username, _ := c.Get("username")
	s.auditService.Log(username.(string), "update_role_permissions", strconv.FormatUint(id, 10), map[string]interface{}{
		"permissions": req.Permissions,
	})

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%"
	password := make([]byte, length)
	for i := range password {
		password[i] = charset[i%len(charset)]
	}
	return string(password)
}
