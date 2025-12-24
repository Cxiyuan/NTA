package backup

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

type Service struct {
	backupDir  string
	dbPath     string
	retention  int
	logger     *logrus.Logger
}

func NewService(backupDir, dbPath string, retentionDays int, logger *logrus.Logger) *Service {
	return &Service{
		backupDir: backupDir,
		dbPath:    dbPath,
		retention: retentionDays,
		logger:    logger,
	}
}

func (s *Service) BackupDatabase() error {
	if err := os.MkdirAll(s.backupDir, 0700); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(s.backupDir, fmt.Sprintf("nta_backup_%s.db.gz", timestamp))

	s.logger.Infof("Creating database backup: %s", backupFile)

	inFile, err := os.Open(s.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer inFile.Close()

	outFile, err := os.Create(backupFile)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer outFile.Close()

	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	if _, err := io.Copy(gzWriter, inFile); err != nil {
		return fmt.Errorf("failed to compress backup: %w", err)
	}

	s.logger.Infof("Backup created successfully: %s", backupFile)

	return s.cleanOldBackups()
}

func (s *Service) RestoreDatabase(backupFile string) error {
	s.logger.Infof("Restoring database from: %s", backupFile)

	inFile, err := os.Open(backupFile)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer inFile.Close()

	gzReader, err := gzip.NewReader(inFile)
	if err != nil {
		return fmt.Errorf("failed to decompress backup: %w", err)
	}
	defer gzReader.Close()

	tempFile := s.dbPath + ".restore"
	outFile, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, gzReader); err != nil {
		return fmt.Errorf("failed to write database: %w", err)
	}

	if err := os.Rename(tempFile, s.dbPath); err != nil {
		return fmt.Errorf("failed to replace database: %w", err)
	}

	s.logger.Info("Database restored successfully")
	return nil
}

func (s *Service) cleanOldBackups() error {
	files, err := filepath.Glob(filepath.Join(s.backupDir, "nta_backup_*.db.gz"))
	if err != nil {
		return err
	}

	cutoff := time.Now().AddDate(0, 0, -s.retention)

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			s.logger.Infof("Removing old backup: %s", file)
			os.Remove(file)
		}
	}

	return nil
}

func (s *Service) ExportToSQL(outputFile string) error {
	s.logger.Infof("Exporting database to SQL: %s", outputFile)

	cmd := exec.Command("sqlite3", s.dbPath, ".dump")
	
	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	cmd.Stdout = outFile
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to export database: %w", err)
	}

	s.logger.Info("Database exported successfully")
	return nil
}

func (s *Service) StartPeriodicBackup(interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.BackupDatabase(); err != nil {
				s.logger.Errorf("Periodic backup failed: %v", err)
			}
		case <-stop:
			s.logger.Info("Stopping periodic backup")
			return
		}
	}
}
