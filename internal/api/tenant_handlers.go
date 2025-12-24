package api

import (
	"net/http"
	"strconv"

	"github.com/Cxiyuan/NTA/internal/rbac"
	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/gin-gonic/gin"
)

func (s *Server) listTenants(c *gin.Context) {
	var tenants []models.Tenant
	if err := s.db.Find(&tenants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tenants"})
		return
	}
	c.JSON(http.StatusOK, tenants)
}

func (s *Server) createTenant(c *gin.Context) {
	var tenant models.Tenant
	if err := c.ShouldBindJSON(&tenant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rbacService := rbac.NewService(s.db)
	if err := rbacService.CreateTenant(&tenant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tenant"})
		return
	}

	username, _ := c.Get("username")
	s.auditService.Log(username.(string), "create_tenant", tenant.TenantID, map[string]interface{}{
		"tenant_id": tenant.ID,
	})

	c.JSON(http.StatusOK, tenant)
}

func (s *Server) updateTenant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant id"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.db.Model(&models.Tenant{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tenant"})
		return
	}

	username, _ := c.Get("username")
	s.auditService.Log(username.(string), "update_tenant", strconv.FormatUint(id, 10), updates)

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (s *Server) deleteTenant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant id"})
		return
	}

	if err := s.db.Delete(&models.Tenant{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete tenant"})
		return
	}

	username, _ := c.Get("username")
	s.auditService.Log(username.(string), "delete_tenant", strconv.FormatUint(id, 10), nil)

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func (s *Server) getTenantUsers(c *gin.Context) {
	tenantID := c.Param("id")

	rbacService := rbac.NewService(s.db)
	users, err := rbacService.GetTenantUsers(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get tenant users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (s *Server) uploadLicense(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}

	savePath := "/opt/nta-probe/config/license.key"
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save license file"})
		return
	}

	if err := s.licenseService.Verify(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid license file"})
		return
	}

	username, _ := c.Get("username")
	s.auditService.Log(username.(string), "upload_license", file.Filename, nil)

	c.JSON(http.StatusOK, gin.H{"status": "uploaded"})
}

func (s *Server) getSystemConfig(c *gin.Context) {
	config := map[string]interface{}{
		"detection": map[string]interface{}{
			"scan": map[string]interface{}{
				"threshold":     20,
				"time_window":   300,
				"min_fail_rate": 0.6,
			},
			"auth": map[string]interface{}{
				"fail_threshold": 5,
				"pth_window":     3600,
			},
			"ml": map[string]interface{}{
				"enabled":       true,
				"contamination": 0.01,
			},
		},
		"backup": map[string]interface{}{
			"enabled":        true,
			"backup_dir":     "/opt/nta-probe/backups",
			"interval_hours": 24,
			"retention_days": 7,
		},
	}
	c.JSON(http.StatusOK, config)
}

func (s *Server) updateDetectionConfig(c *gin.Context) {
	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, _ := c.Get("username")
	s.auditService.Log(username.(string), "update_detection_config", "", config)

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

func (s *Server) updateBackupConfig(c *gin.Context) {
	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, _ := c.Get("username")
	s.auditService.Log(username.(string), "update_backup_config", "", config)

	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}
