package threatintel

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Service provides threat intelligence lookup
type Service struct {
	db         *gorm.DB
	redis      *redis.Client
	logger     *logrus.Logger
	sources    []Source
	cache      map[string]*CacheEntry
	cacheMu    sync.RWMutex
	cacheTTL   time.Duration
}

type Source struct {
	Name    string
	URL     string
	APIKey  string
	Enabled bool
}

type CacheEntry struct {
	Result    *models.ThreatIntel
	ExpiresAt time.Time
}

// NewService creates a new threat intelligence service
func NewService(db *gorm.DB, rdb *redis.Client, logger *logrus.Logger, sources []Source) *Service {
	return &Service{
		db:       db,
		redis:    rdb,
		logger:   logger,
		sources:  sources,
		cache:    make(map[string]*CacheEntry),
		cacheTTL: 1 * time.Hour,
	}
}

// CheckIP checks if an IP is malicious
func (s *Service) CheckIP(ctx context.Context, ip string) (*models.ThreatIntel, error) {
	// Check cache first
	if cached := s.getFromCache("ip:" + ip); cached != nil {
		return cached, nil
	}

	// Check database
	var intel models.ThreatIntel
	err := s.db.Where("type = ? AND value = ?", "ip", ip).First(&intel).Error
	if err == nil {
		s.putToCache("ip:"+ip, &intel)
		return &intel, nil
	}

	// Check external sources
	for _, source := range s.sources {
		if !source.Enabled {
			continue
		}

		result, err := s.querySource(ctx, source, "ip", ip)
		if err != nil {
			s.logger.Warnf("Failed to query source %s: %v", source.Name, err)
			continue
		}

		if result != nil {
			// Save to database
			s.db.Create(result)
			s.putToCache("ip:"+ip, result)
			return result, nil
		}
	}

	// Not found - create benign entry
	benign := &models.ThreatIntel{
		Type:     "ip",
		Value:    ip,
		Severity: "none",
		Source:   "local",
	}
	s.putToCache("ip:"+ip, benign)

	return benign, nil
}

// CheckDomain checks if a domain is malicious
func (s *Service) CheckDomain(ctx context.Context, domain string) (*models.ThreatIntel, error) {
	cacheKey := "domain:" + domain
	
	if cached := s.getFromCache(cacheKey); cached != nil {
		return cached, nil
	}

	var intel models.ThreatIntel
	err := s.db.Where("type = ? AND value = ?", "domain", domain).First(&intel).Error
	if err == nil {
		s.putToCache(cacheKey, &intel)
		return &intel, nil
	}

	for _, source := range s.sources {
		if !source.Enabled {
			continue
		}

		result, err := s.querySource(ctx, source, "domain", domain)
		if err != nil {
			continue
		}

		if result != nil {
			s.db.Create(result)
			s.putToCache(cacheKey, result)
			return result, nil
		}
	}

	benign := &models.ThreatIntel{
		Type:     "domain",
		Value:    domain,
		Severity: "none",
		Source:   "local",
	}
	s.putToCache(cacheKey, benign)

	return benign, nil
}

// CheckHash checks if a file hash is malicious
func (s *Service) CheckHash(ctx context.Context, hash string) (*models.ThreatIntel, error) {
	cacheKey := "hash:" + hash
	
	if cached := s.getFromCache(cacheKey); cached != nil {
		return cached, nil
	}

	var intel models.ThreatIntel
	err := s.db.Where("type = ? AND value = ?", "hash", hash).First(&intel).Error
	if err == nil {
		s.putToCache(cacheKey, &intel)
		return &intel, nil
	}

	for _, source := range s.sources {
		if !source.Enabled {
			continue
		}

		result, err := s.querySource(ctx, source, "hash", hash)
		if err != nil {
			continue
		}

		if result != nil {
			s.db.Create(result)
			s.putToCache(cacheKey, result)
			return result, nil
		}
	}

	benign := &models.ThreatIntel{
		Type:     "hash",
		Value:    hash,
		Severity: "none",
		Source:   "local",
	}
	s.putToCache(cacheKey, benign)

	return benign, nil
}

// querySource queries external threat intelligence source
func (s *Service) querySource(ctx context.Context, source Source, iocType, value string) (*models.ThreatIntel, error) {
	// Simplified - actual implementation would vary per source
	client := &http.Client{Timeout: 10 * time.Second}
	
	req, err := http.NewRequestWithContext(ctx, "GET", source.URL, nil)
	if err != nil {
		return nil, err
	}

	if source.APIKey != "" {
		req.Header.Set("X-API-KEY", source.APIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse response (simplified)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Check if malicious
	if isMalicious, ok := result["malicious"].(bool); ok && isMalicious {
		return &models.ThreatIntel{
			Type:     iocType,
			Value:    value,
			Severity: "high",
			Source:   source.Name,
		}, nil
	}

	return nil, nil
}

// UpdateFeeds updates threat intelligence from all sources
func (s *Service) UpdateFeeds(ctx context.Context) error {
	s.logger.Info("Updating threat intelligence feeds...")

	for _, source := range s.sources {
		if !source.Enabled {
			continue
		}

		s.logger.Infof("Fetching from source: %s", source.Name)
		// Implementation would fetch bulk data from each source
	}

	return nil
}

// getFromCache retrieves entry from cache
func (s *Service) getFromCache(key string) *models.ThreatIntel {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	entry, exists := s.cache[key]
	if !exists {
		return nil
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil
	}

	return entry.Result
}

// putToCache stores entry in cache
func (s *Service) putToCache(key string, intel *models.ThreatIntel) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.cache[key] = &CacheEntry{
		Result:    intel,
		ExpiresAt: time.Now().Add(s.cacheTTL),
	}
}

// CleanCache removes expired entries
func (s *Service) CleanCache() {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	now := time.Now()
	for key, entry := range s.cache {
		if now.After(entry.ExpiresAt) {
			delete(s.cache, key)
		}
	}
}
