package migrations

import (
	"fmt"
	"time"

	"github.com/Cxiyuan/NTA/pkg/models"
	"gorm.io/gorm"
)

type Migration struct {
	ID        uint      `gorm:"primaryKey"`
	Version   string    `gorm:"uniqueIndex"`
	AppliedAt time.Time
}

type Migrator struct {
	db *gorm.DB
}

func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) Initialize() error {
	return m.db.AutoMigrate(&Migration{})
}

func (m *Migrator) ApplyMigrations() error {
	migrations := []struct {
		version string
		fn      func(*gorm.DB) error
	}{
		{"001_initial_schema", m.migration001InitialSchema},
		{"002_add_indexes", m.migration002AddIndexes},
		{"003_add_tenant_support", m.migration003AddTenantSupport},
	}

	for _, migration := range migrations {
		var existing Migration
		err := m.db.Where("version = ?", migration.version).First(&existing).Error
		
		if err == gorm.ErrRecordNotFound {
			fmt.Printf("Applying migration: %s\n", migration.version)
			
			if err := migration.fn(m.db); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", migration.version, err)
			}

			m.db.Create(&Migration{
				Version:   migration.version,
				AppliedAt: time.Now(),
			})
			
			fmt.Printf("Migration %s applied successfully\n", migration.version)
		} else if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) migration001InitialSchema(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Alert{},
		&models.Asset{},
		&models.ThreatIntel{},
		&models.Probe{},
		&models.APTIndicator{},
		&models.AuditLog{},
	)
}

func (m *Migrator) migration002AddIndexes(db *gorm.DB) error {
	sqls := []string{
		"CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity)",
		"CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status)",
		"CREATE INDEX IF NOT EXISTS idx_alerts_timestamp ON alerts(timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_assets_last_seen ON assets(last_seen)",
		"CREATE INDEX IF NOT EXISTS idx_threat_intel_type ON threat_intels(type)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp)",
	}

	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) migration003AddTenantSupport(db *gorm.DB) error {
	sqls := []string{
		"ALTER TABLE alerts ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(64) DEFAULT ''",
		"ALTER TABLE assets ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(64) DEFAULT ''",
		"CREATE INDEX IF NOT EXISTS idx_alerts_tenant ON alerts(tenant_id)",
		"CREATE INDEX IF NOT EXISTS idx_assets_tenant ON assets(tenant_id)",
	}

	for _, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			return err
		}
	}

	return nil
}
