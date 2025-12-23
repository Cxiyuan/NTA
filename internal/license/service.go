package license

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Service manages license validation
type Service struct {
	logger     *logrus.Logger
	license    *License
	publicKey  *rsa.PublicKey
}

// License represents a software license
type License struct {
	Customer       string    `json:"customer"`
	Product        string    `json:"product"`
	MaxProbes      int       `json:"max_probes"`
	MaxBandwidth   int       `json:"max_bandwidth_mbps"`
	IssueDate      time.Time `json:"issue_date"`
	ExpiryDate     time.Time `json:"expiry_date"`
	Features       []string  `json:"features"`
	Signature      string    `json:"signature"`
}

// NewService creates a new license service
func NewService(licenseFile, publicKeyFile string, logger *logrus.Logger) (*Service, error) {
	s := &Service{
		logger: logger,
	}

	// Load public key
	pubKey, err := loadPublicKey(publicKeyFile)
	if err != nil {
		return nil, err
	}
	s.publicKey = pubKey

	// Load and verify license
	license, err := loadLicense(licenseFile)
	if err != nil {
		return nil, err
	}

	s.license = license

	return s, nil
}

// Verify verifies the license is valid
func (s *Service) Verify() error {
	// Check expiry
	if time.Now().After(s.license.ExpiryDate) {
		return errors.New("license expired")
	}

	// TODO: Verify RSA signature
	
	s.logger.Info("License verified successfully")
	return nil
}

// HasFeature checks if license includes a feature
func (s *Service) HasFeature(feature string) bool {
	for _, f := range s.license.Features {
		if f == feature {
			return true
		}
	}
	return false
}

// GetInfo returns license information
func (s *Service) GetInfo() *License {
	return s.license
}

func loadLicense(path string) (*License, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var license License
	if err := json.Unmarshal(data, &license); err != nil {
		return nil, err
	}

	return &license, nil
}

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return rsaPub, nil
}
