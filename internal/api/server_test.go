package api

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"bytes"
	"encoding/json"

	"github.com/Cxiyuan/NTA/internal/asset"
	"github.com/Cxiyuan/NTA/internal/audit"
	"github.com/Cxiyuan/NTA/internal/license"
	"github.com/Cxiyuan/NTA/internal/probe"
	"github.com/Cxiyuan/NTA/internal/threatintel"
	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=test_user password=test_pass dbname=nta_test port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available: %v", err)
	}

	db.AutoMigrate(
		&models.Alert{},
		&models.Asset{},
		&models.ThreatIntel{},
		&models.Probe{},
		&models.APTIndicator{},
		&models.AuditLog{},
	)

	return db
}

func setupTestServer(t *testing.T, db *gorm.DB) *Server {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	assetScanner := asset.NewScanner(db, logger)
	threatIntel := threatintel.NewService(db, nil, logger, []threatintel.Source{})
	probeManager := probe.NewManager(db, nil, logger)
	auditService := audit.NewService(db, logger)
	licenseService := &license.Service{}

	return NewServer(
		db,
		logger,
		assetScanner,
		threatIntel,
		probeManager,
		licenseService,
		auditService,
		"test-secret-key",
	)
}

func TestHealthCheck(t *testing.T) {
	db := setupTestDB(t)
	server := setupTestServer(t, db)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "ok", response["status"])
}

func TestListAlertsWithPagination(t *testing.T) {
	db := setupTestDB(t)
	server := setupTestServer(t, db)

	for i := 0; i < 10; i++ {
		db.Create(&models.Alert{
			Severity:    "high",
			Type:        "test",
			SrcIP:       "192.168.1.1",
			Status:      "new",
		})
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/alerts?page=1&page_size=5", nil)
	
	token, _ := server.authMiddleware.GenerateToken("test-user", "testuser", []string{"admin"})
	req.Header.Set("Authorization", "Bearer "+token)
	
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	data := response["data"].([]interface{})
	assert.Equal(t, 5, len(data))
	assert.Equal(t, float64(10), response["total"])
}

func TestUpdateAlertWithValidation(t *testing.T) {
	db := setupTestDB(t)
	server := setupTestServer(t, db)

	alert := &models.Alert{
		Severity: "high",
		Type:     "test",
		SrcIP:    "192.168.1.1",
		Status:   "new",
	}
	db.Create(alert)

	updateData := map[string]string{"status": "resolved"}
	jsonData, _ := json.Marshal(updateData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/alerts/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	token, _ := server.authMiddleware.GenerateToken("test-user", "testuser", []string{"admin"})
	req.Header.Set("Authorization", "Bearer "+token)
	
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedAlert models.Alert
	db.First(&updatedAlert, 1)
	assert.Equal(t, "resolved", updatedAlert.Status)
}

func TestUpdateAlertWithInvalidStatus(t *testing.T) {
	db := setupTestDB(t)
	server := setupTestServer(t, db)

	alert := &models.Alert{
		Severity: "high",
		Type:     "test",
		SrcIP:    "192.168.1.1",
		Status:   "new",
	}
	db.Create(alert)

	updateData := map[string]string{"status": "invalid_status"}
	jsonData, _ := json.Marshal(updateData)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/alerts/1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	
	token, _ := server.authMiddleware.GenerateToken("test-user", "testuser", []string{"admin"})
	req.Header.Set("Authorization", "Bearer "+token)
	
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthenticationRequired(t *testing.T) {
	db := setupTestDB(t)
	server := setupTestServer(t, db)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/alerts", nil)
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRoleBasedAccess(t *testing.T) {
	db := setupTestDB(t)
	server := setupTestServer(t, db)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/license", nil)
	
	token, _ := server.authMiddleware.GenerateToken("test-user", "testuser", []string{"viewer"})
	req.Header.Set("Authorization", "Bearer "+token)
	
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}