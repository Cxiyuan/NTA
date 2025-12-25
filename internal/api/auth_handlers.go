package api

import (
	"net/http"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  UserInfo    `json:"user"`
}

type UserInfo struct {
	ID       uint     `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	TenantID string   `json:"tenant_id"`
	Roles    []string `json:"roles"`
}

func (s *Server) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
		return
	}

	var user models.User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		s.logger.Warnf("Login attempt for non-existent user: %s", req.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials", "message": "用户名或密码错误"})
		return
	}

	if user.Status != models.StatusActive {
		s.logger.Warnf("Login attempt for inactive user: %s (status: %s)", req.Username, user.Status)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user inactive", "message": "用户账号已被禁用"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.logger.Warnf("Failed login attempt for user: %s (hash: %s, error: %v)", req.Username, user.PasswordHash, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials", "message": "用户名或密码错误"})
		return
	}

	var userRoles []models.UserRole
	var roles []string
	if err := s.db.Where("user_id = ?", user.ID).Find(&userRoles).Error; err == nil {
		for _, ur := range userRoles {
			var role models.Role
			if err := s.db.First(&role, ur.RoleID).Error; err == nil {
				roles = append(roles, role.Name)
			}
		}
	}

	if len(roles) == 0 {
		roles = []string{models.RoleViewer}
	}

	token, err := s.authMiddleware.GenerateToken(
		string(user.ID),
		user.Username,
		roles,
	)
	if err != nil {
		s.logger.Errorf("Failed to generate token for user %s: %v", user.Username, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed", "message": "生成令牌失败"})
		return
	}

	s.auditService.Log(user.Username, "login", "", map[string]interface{}{
		"ip": c.ClientIP(),
	})

	s.logger.Infof("User %s logged in successfully", user.Username)

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			TenantID: user.TenantID,
			Roles:    roles,
		},
	})
}

func (s *Server) logout(c *gin.Context) {
	username, exists := c.Get("username")
	if exists {
		s.auditService.Log(username.(string), "logout", "", map[string]interface{}{
			"ip": c.ClientIP(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func (s *Server) getCurrentUser(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	roles, _ := c.Get("roles")

	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	rolesList, _ := roles.([]string)

	c.JSON(http.StatusOK, UserInfo{
		ID:       user.ID,
		Username: username.(string),
		Email:    user.Email,
		TenantID: user.TenantID,
		Roles:    rolesList,
	})
}