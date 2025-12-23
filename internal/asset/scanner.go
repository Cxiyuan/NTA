package asset

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Scanner performs network asset discovery
type Scanner struct {
	db     *gorm.DB
	logger *logrus.Logger
	assets map[string]*models.Asset
	mu     sync.RWMutex
}

// NewScanner creates a new asset scanner
func NewScanner(db *gorm.DB, logger *logrus.Logger) *Scanner {
	return &Scanner{
		db:     db,
		logger: logger,
		assets: make(map[string]*models.Asset),
	}
}

// DiscoverFromTraffic discovers assets from network traffic
func (s *Scanner) DiscoverFromTraffic(ctx context.Context, iface string) error {
	handle, err := pcap.OpenLive(iface, 1600, true, pcap.BlockForever)
	if err != nil {
		return err
	}
	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for {
		select {
		case <-ctx.Done():
			return nil
		case packet := <-packetSource.Packets():
			s.processPacket(packet)
		}
	}
}

func (s *Scanner) processPacket(packet gopacket.Packet) {
	// Extract IP layer
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		return
	}

	ip, _ := ipLayer.(*layers.IPv4)
	
	// Update source IP
	s.updateAsset(ip.SrcIP.String(), "")
	
	// Update destination IP
	s.updateAsset(ip.DstIP.String(), "")
}

func (s *Scanner) updateAsset(ip, mac string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	asset, exists := s.assets[ip]
	if !exists {
		asset = &models.Asset{
			IP:        ip,
			MAC:       mac,
			FirstSeen: time.Now(),
			LastSeen:  time.Now(),
		}
		s.assets[ip] = asset
		
		// Resolve hostname
		go s.resolveHostname(asset)
	} else {
		asset.LastSeen = time.Now()
	}
}

func (s *Scanner) resolveHostname(asset *models.Asset) {
	names, err := net.LookupAddr(asset.IP)
	if err == nil && len(names) > 0 {
		s.mu.Lock()
		asset.Hostname = names[0]
		s.mu.Unlock()
	}
}

// SaveAssets persists discovered assets to database
func (s *Scanner) SaveAssets() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, asset := range s.assets {
		result := s.db.Where("ip = ?", asset.IP).FirstOrCreate(asset)
		if result.Error != nil {
			s.logger.Errorf("Failed to save asset %s: %v", asset.IP, result.Error)
		}
	}

	return nil
}

// GetAssets returns all discovered assets
func (s *Scanner) GetAssets() []*models.Asset {
	s.mu.RLock()
	defer s.mu.RUnlock()

	assets := make([]*models.Asset, 0, len(s.assets))
	for _, asset := range s.assets {
		assets = append(assets, asset)
	}

	return assets
}

// ScanNetwork performs active network scanning (nmap-style)
func (s *Scanner) ScanNetwork(network string) ([]*models.Asset, error) {
	// Parse CIDR
	_, ipnet, err := net.ParseCIDR(network)
	if err != nil {
		return nil, err
	}

	var assets []*models.Asset
	
	// Iterate through IPs in network
	for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		if isReachable(ip.String()) {
			asset := &models.Asset{
				IP:        ip.String(),
				FirstSeen: time.Now(),
				LastSeen:  time.Now(),
			}
			
			// Resolve hostname
			names, _ := net.LookupAddr(ip.String())
			if len(names) > 0 {
				asset.Hostname = names[0]
			}
			
			assets = append(assets, asset)
		}
	}

	return assets, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func isReachable(ip string) bool {
	timeout := 500 * time.Millisecond
	conn, err := net.DialTimeout("tcp", ip+":80", timeout)
	if err == nil {
		conn.Close()
		return true
	}
	return false
}
