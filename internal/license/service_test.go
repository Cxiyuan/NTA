package license

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func generateTestKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	return privateKey, &privateKey.PublicKey
}

func signLicense(t *testing.T, privateKey *rsa.PrivateKey, license *License) string {
	licenseData := struct {
		Customer     string    `json:"customer"`
		Product      string    `json:"product"`
		MaxProbes    int       `json:"max_probes"`
		MaxBandwidth int       `json:"max_bandwidth_mbps"`
		IssueDate    time.Time `json:"issue_date"`
		ExpiryDate   time.Time `json:"expiry_date"`
		Features     []string  `json:"features"`
	}{
		Customer:     license.Customer,
		Product:      license.Product,
		MaxProbes:    license.MaxProbes,
		MaxBandwidth: license.MaxBandwidth,
		IssueDate:    license.IssueDate,
		ExpiryDate:   license.ExpiryDate,
		Features:     license.Features,
	}

	dataBytes, _ := json.Marshal(licenseData)
	hashed := sha256.Sum256(dataBytes)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		t.Fatalf("Failed to sign license: %v", err)
	}

	return base64.StdEncoding.EncodeToString(signature)
}

func TestLicenseVerification(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)

	license := &License{
		Customer:     "Test Company",
		Product:      "NTA",
		MaxProbes:    10,
		MaxBandwidth: 1000,
		IssueDate:    time.Now().Add(-24 * time.Hour),
		ExpiryDate:   time.Now().Add(365 * 24 * time.Hour),
		Features:     []string{"threat_intel", "apt_detection"},
	}

	license.Signature = signLicense(t, privateKey, license)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := &Service{
		logger:    logger,
		license:   license,
		publicKey: publicKey,
	}

	err := service.Verify()
	assert.NoError(t, err)
}

func TestExpiredLicense(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)

	license := &License{
		Customer:     "Test Company",
		Product:      "NTA",
		MaxProbes:    10,
		MaxBandwidth: 1000,
		IssueDate:    time.Now().Add(-48 * time.Hour),
		ExpiryDate:   time.Now().Add(-24 * time.Hour),
		Features:     []string{"threat_intel"},
	}

	license.Signature = signLicense(t, privateKey, license)

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := &Service{
		logger:    logger,
		license:   license,
		publicKey: publicKey,
	}

	err := service.Verify()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestInvalidSignature(t *testing.T) {
	_, publicKey := generateTestKeyPair(t)

	license := &License{
		Customer:     "Test Company",
		Product:      "NTA",
		MaxProbes:    10,
		MaxBandwidth: 1000,
		IssueDate:    time.Now().Add(-24 * time.Hour),
		ExpiryDate:   time.Now().Add(365 * 24 * time.Hour),
		Features:     []string{"threat_intel"},
		Signature:    "invalid_signature",
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	service := &Service{
		logger:    logger,
		license:   license,
		publicKey: publicKey,
	}

	err := service.Verify()
	assert.Error(t, err)
}

func TestHasFeature(t *testing.T) {
	license := &License{
		Features: []string{"threat_intel", "apt_detection", "encryption_analysis"},
	}

	service := &Service{
		license: license,
	}

	assert.True(t, service.HasFeature("threat_intel"))
	assert.True(t, service.HasFeature("apt_detection"))
	assert.False(t, service.HasFeature("non_existent_feature"))
}

func TestLoadPublicKey(t *testing.T) {
	privateKey, _ := generateTestKeyPair(t)
	publicKey := &privateKey.PublicKey

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	assert.NoError(t, err)

	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	}

	tmpFile, err := os.CreateTemp("", "test-public-*.pem")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	err = pem.Encode(tmpFile, pemBlock)
	assert.NoError(t, err)
	tmpFile.Close()

	loadedKey, err := loadPublicKey(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, loadedKey)
	assert.Equal(t, publicKey.N, loadedKey.N)
}
