package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Zeek        ZeekConfig        `yaml:"zeek"`
	Redis       RedisConfig       `yaml:"redis"`
	Database    DatabaseConfig    `yaml:"database"`
	Detection   DetectionConfig   `yaml:"detection"`
	ThreatIntel ThreatIntelConfig `yaml:"threat_intel"`
	License     LicenseConfig     `yaml:"license"`
	Security    SecurityConfig    `yaml:"security"`
	Backup      BackupConfig      `yaml:"backup"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"` // debug, release
}

type ZeekConfig struct {
	LogDir    string `yaml:"log_dir"`
	ScriptDir string `yaml:"script_dir"`
	Interface string `yaml:"interface"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type DatabaseConfig struct {
	Type string `yaml:"type"` // postgres
	DSN  string `yaml:"dsn"`
}

type DetectionConfig struct {
	Scan ScanConfig `yaml:"scan"`
	Auth AuthConfig `yaml:"auth"`
	ML   MLConfig   `yaml:"ml"`
}

type ScanConfig struct {
	Threshold   int     `yaml:"threshold"`
	TimeWindow  int     `yaml:"time_window"`
	MinFailRate float64 `yaml:"min_fail_rate"`
}

type AuthConfig struct {
	FailThreshold int `yaml:"fail_threshold"`
	PTHWindow     int `yaml:"pth_window"`
}

type MLConfig struct {
	Enabled       bool    `yaml:"enabled"`
	Contamination float64 `yaml:"contamination"`
}

type ThreatIntelConfig struct {
	Sources      []ThreatSource `yaml:"sources"`
	UpdateInterval int          `yaml:"update_interval"`
	LocalFeedPath string        `yaml:"local_feed_path"`
}

type ThreatSource struct {
	Name    string `yaml:"name"`
	URL     string `yaml:"url"`
	APIKey  string `yaml:"api_key"`
	Enabled bool   `yaml:"enabled"`
}

type LicenseConfig struct {
	LicenseFile   string `yaml:"license_file"`
	PublicKeyFile string `yaml:"public_key_file"`
}

type SecurityConfig struct {
	JWTSecret          string `yaml:"jwt_secret"`
	EnableTLS          bool   `yaml:"enable_tls"`
	TLSCertFile        string `yaml:"tls_cert_file"`
	TLSKeyFile         string `yaml:"tls_key_file"`
	RateLimitRequests  int    `yaml:"rate_limit_requests"`
	RateLimitWindow    int    `yaml:"rate_limit_window"`
	AllowedOrigins     []string `yaml:"allowed_origins"`
}

type BackupConfig struct {
	Enabled        bool   `yaml:"enabled"`
	BackupDir      string `yaml:"backup_dir"`
	Interval       int    `yaml:"interval_hours"`
	RetentionDays  int    `yaml:"retention_days"`
}

// LoadConfig loads configuration from YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return errors.New("invalid server port")
	}

	if c.Server.Host == "0.0.0.0" && c.Server.Mode == "release" && !c.Security.EnableTLS {
		return errors.New("binding to 0.0.0.0 in release mode without TLS is insecure")
	}

	if c.Database.Type != "postgres" {
		return errors.New("only postgres database is supported")
	}

	if c.Database.DSN == "" {
		return errors.New("database DSN is required")
	}

	if c.Security.JWTSecret == "" || len(c.Security.JWTSecret) < 32 {
		return errors.New("JWT secret must be at least 32 characters")
	}

	if c.Security.EnableTLS {
		if c.Security.TLSCertFile == "" || c.Security.TLSKeyFile == "" {
			return errors.New("TLS cert and key files are required when TLS is enabled")
		}
	}

	if c.Detection.Scan.Threshold < 1 {
		return errors.New("scan threshold must be positive")
	}

	if c.Backup.Enabled && c.Backup.BackupDir == "" {
		return errors.New("backup directory is required when backup is enabled")
	}

	return nil
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 8080,
			Mode: "release",
		},
		Zeek: ZeekConfig{
			LogDir:    "/var/spool/zeek",
			ScriptDir: "/opt/nta-probe/zeek-scripts",
			Interface: "eth0",
		},
		Redis: RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		Database: DatabaseConfig{
			Type: "postgres",
			DSN:  "host=localhost user=nta password=nta_password dbname=nta port=5432 sslmode=disable",
		},
		Detection: DetectionConfig{
			Scan: ScanConfig{
				Threshold:   20,
				TimeWindow:  300,
				MinFailRate: 0.6,
			},
			Auth: AuthConfig{
				FailThreshold: 5,
				PTHWindow:     3600,
			},
			ML: MLConfig{
				Enabled:       true,
				Contamination: 0.01,
			},
		},
		ThreatIntel: ThreatIntelConfig{
			Sources: []ThreatSource{
				{
					Name:    "threatfox",
					URL:     "https://threatfox-api.abuse.ch/api/v1/",
					Enabled: true,
				},
			},
			UpdateInterval: 3600,
			LocalFeedPath:  "/opt/nta-probe/config/threat_feed.json",
		},
		License: LicenseConfig{
			LicenseFile:   "/opt/nta-probe/config/license.key",
			PublicKeyFile: "/opt/nta-probe/config/public.pem",
		},
		Security: SecurityConfig{
			JWTSecret:         "change-this-secret-in-production-min-32-chars",
			EnableTLS:         false,
			RateLimitRequests: 100,
			RateLimitWindow:   60,
			AllowedOrigins:    []string{"*"},
		},
		Backup: BackupConfig{
			Enabled:       true,
			BackupDir:     "/opt/nta-probe/backups",
			Interval:      24,
			RetentionDays: 7,
		},
	}
}