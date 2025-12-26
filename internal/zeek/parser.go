package zeek

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/sirupsen/logrus"
)

// LogParser parses Zeek log files
type LogParser struct {
	logDir string
	logger *logrus.Logger
}

// NewLogParser creates a new Zeek log parser
func NewLogParser(logDir string, logger *logrus.Logger) *LogParser {
	return &LogParser{
		logDir: logDir,
		logger: logger,
	}
}

// ParseConnLog parses conn.log file
func (p *LogParser) ParseConnLog(filePath string) ([]*models.Connection, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var conns []*models.Connection
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		
		// Skip comments and headers
		if strings.HasPrefix(line, "#") {
			continue
		}

		conn, err := p.parseConnLine(line)
		if err != nil {
			p.logger.Warnf("Failed to parse conn line: %v", err)
			continue
		}

		conns = append(conns, conn)
	}

	return conns, scanner.Err()
}

// parseConnLine parses a single line from conn.log
func (p *LogParser) parseConnLine(line string) (*models.Connection, error) {
	fields := strings.Split(line, "\t")
	if len(fields) < 11 {
		return nil, nil
	}

	ts, _ := time.Parse(time.RFC3339, fields[0])
	srcPort := 0
	dstPort := 0
	
	// Parse ports (simplified)
	if len(fields) > 3 {
		// srcPort = parseInt(fields[3])
	}
	if len(fields) > 5 {
		// dstPort = parseInt(fields[5])
	}

	conn := &models.Connection{
		UID:       fields[1],
		Timestamp: ts,
		SrcIP:     fields[2],
		SrcPort:   srcPort,
		DstIP:     fields[4],
		DstPort:   dstPort,
		Protocol:  fields[6],
		Service:   fields[7],
	}

	return conn, nil
}

// ParseSSLLog parses ssl.log file
func (p *LogParser) ParseSSLLog(filePath string) ([]*models.TLSHandshake, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var handshakes []*models.TLSHandshake
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		
		if strings.HasPrefix(line, "#") {
			continue
		}

		hs, err := p.parseSSLLine(line)
		if err != nil {
			continue
		}

		handshakes = append(handshakes, hs)
	}

	return handshakes, scanner.Err()
}

func (p *LogParser) parseSSLLine(line string) (*models.TLSHandshake, error) {
	fields := strings.Split(line, "\t")
	if len(fields) < 10 {
		return nil, nil
	}

	ts, _ := time.Parse(time.RFC3339, fields[0])

	hs := &models.TLSHandshake{
		Timestamp:  ts,
		UID:        fields[1],
		SrcIP:      fields[2],
		DstIP:      fields[4],
		Version:    fields[6],
		ServerName: fields[9],
	}

	return hs, nil
}

// WatchLogDir watches Zeek log directory for new files
func (p *LogParser) WatchLogDir(callback func(string, string)) error {
	// Simple implementation - scan for current logs
	currentDir := filepath.Join(p.logDir, "current")
	
	files, err := os.ReadDir(currentDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.HasSuffix(file.Name(), ".log") {
			fullPath := filepath.Join(currentDir, file.Name())
			callback(file.Name(), fullPath)
		}
	}

	return nil
}

// ParseJSON parses JSON formatted Zeek logs
func (p *LogParser) ParseJSON(filePath string, result interface{}) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, result)
}