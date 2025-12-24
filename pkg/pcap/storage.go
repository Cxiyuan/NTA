package pcap

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pcapgo"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Storage struct {
	db         *gorm.DB
	logger     *logrus.Logger
	storageDir string
	writers    map[string]*pcapgo.Writer
	files      map[string]*os.File
	mu         sync.RWMutex
	maxSize    int64
	retention  time.Duration
}

func NewStorage(db *gorm.DB, logger *logrus.Logger, storageDir string) *Storage {
	os.MkdirAll(storageDir, 0755)

	s := &Storage{
		db:         db,
		logger:     logger,
		storageDir: storageDir,
		writers:    make(map[string]*pcapgo.Writer),
		files:      make(map[string]*os.File),
		maxSize:    1024 * 1024 * 1024,
		retention:  30 * 24 * time.Hour,
	}

	go s.cleanupLoop()

	return s
}

func (s *Storage) CapturePacket(packet gopacket.Packet) error {
	networkLayer := packet.NetworkLayer()
	if networkLayer == nil {
		return nil
	}

	var srcIP, dstIP string
	var srcPort, dstPort int
	var protocol string

	if ipv4Layer := packet.Layer(layers.LayerTypeIPv4); ipv4Layer != nil {
		ipv4, _ := ipv4Layer.(*layers.IPv4)
		srcIP = ipv4.SrcIP.String()
		dstIP = ipv4.DstIP.String()
		protocol = ipv4.Protocol.String()
	}

	if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.TCP)
		srcPort = int(tcp.SrcPort)
		dstPort = int(tcp.DstPort)
	} else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp, _ := udpLayer.(*layers.UDP)
		srcPort = int(udp.SrcPort)
		dstPort = int(udp.DstPort)
	}

	sessionID := fmt.Sprintf("%s:%d-%s:%d-%s", srcIP, srcPort, dstIP, dstPort, protocol)

	s.mu.Lock()
	writer, exists := s.writers[sessionID]
	if !exists {
		writer, err := s.createWriter(sessionID)
		if err != nil {
			s.mu.Unlock()
			return err
		}
		s.writers[sessionID] = writer
	}
	s.mu.Unlock()

	if writer != nil {
		err := writer.WritePacket(packet.Metadata().CaptureInfo, packet.Data())
		if err != nil {
			s.logger.Errorf("Failed to write packet: %v", err)
		}
	}

	return nil
}

func (s *Storage) createWriter(sessionID string) (*pcapgo.Writer, error) {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.pcap", sessionID, timestamp)
	filePath := filepath.Join(s.storageDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	writer := pcapgo.NewWriter(file)
	if err := writer.WriteFileHeader(65536, layers.LinkTypeEthernet); err != nil {
		file.Close()
		return nil, err
	}

	s.files[sessionID] = file

	session := &models.PCAPSession{
		SessionID:   sessionID,
		FilePath:    filePath,
		StartTime:   time.Now(),
		PacketCount: 0,
		BytesTotal:  0,
	}
	s.db.Create(session)

	return writer, nil
}

func (s *Storage) CloseSession(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if file, exists := s.files[sessionID]; exists {
		file.Close()
		delete(s.files, sessionID)
		delete(s.writers, sessionID)

		s.db.Model(&models.PCAPSession{}).
			Where("session_id = ?", sessionID).
			Update("end_time", time.Now())
	}

	return nil
}

func (s *Storage) SearchSessions(srcIP, dstIP string, startTime, endTime time.Time, limit int) ([]models.PCAPSession, error) {
	var sessions []models.PCAPSession

	query := s.db.Model(&models.PCAPSession{})

	if srcIP != "" {
		query = query.Where("src_ip = ?", srcIP)
	}
	if dstIP != "" {
		query = query.Where("dst_ip = ?", dstIP)
	}
	if !startTime.IsZero() {
		query = query.Where("start_time >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("end_time <= ?", endTime)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Order("start_time DESC").Find(&sessions).Error; err != nil {
		return nil, err
	}

	return sessions, nil
}

func (s *Storage) GetSessionFile(sessionID string) (string, error) {
	var session models.PCAPSession
	if err := s.db.Where("session_id = ?", sessionID).First(&session).Error; err != nil {
		return "", err
	}

	if _, err := os.Stat(session.FilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("PCAP file not found")
	}

	return session.FilePath, nil
}

func (s *Storage) cleanupLoop() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanup()
	}
}

func (s *Storage) cleanup() {
	cutoffTime := time.Now().Add(-s.retention)

	var oldSessions []models.PCAPSession
	if err := s.db.Where("start_time < ?", cutoffTime).Find(&oldSessions).Error; err != nil {
		s.logger.Errorf("Failed to query old sessions: %v", err)
		return
	}

	for _, session := range oldSessions {
		if err := os.Remove(session.FilePath); err != nil {
			s.logger.Warnf("Failed to delete PCAP file %s: %v", session.FilePath, err)
		}

		if err := s.db.Delete(&session).Error; err != nil {
			s.logger.Errorf("Failed to delete session record: %v", err)
		}
	}

	s.logger.Infof("Cleaned up %d old PCAP sessions", len(oldSessions))
}

func (s *Storage) StartCapture(device string) error {
	handle, err := pcap.OpenLive(device, 1600, true, pcap.BlockForever)
	if err != nil {
		return err
	}
	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	s.logger.Infof("Started PCAP capture on device %s", device)

	for packet := range packetSource.Packets() {
		if err := s.CapturePacket(packet); err != nil {
			s.logger.Errorf("Failed to capture packet: %v", err)
		}
	}

	return nil
}
