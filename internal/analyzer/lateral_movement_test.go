package analyzer

import (
	"testing"
	"time"

	"github.com/Cxiyuan/NTA/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDetectScan(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	detector := NewLateralMovementDetector(logger, 5, 300)

	for i := 0; i < 6; i++ {
		conn := &models.Connection{
			SrcIP:     "192.168.1.100",
			DstIP:     "192.168.1." + string(rune(i+1)),
			Timestamp: time.Now(),
		}
		alert := detector.DetectScan(conn)
		
		if i < 4 {
			assert.Nil(t, alert, "Should not trigger alert for %d targets", i+1)
		} else {
			assert.NotNil(t, alert, "Should trigger alert for 5+ targets")
			if alert != nil {
				assert.Equal(t, "lateral_scan", alert.Type)
				assert.Equal(t, "high", alert.Severity)
			}
		}
	}
}

func TestDetectPTH(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	detector := NewLateralMovementDetector(logger, 20, 300)

	srcIP := "192.168.1.100"
	hash := "ntlm:hash123456"

	alert := detector.DetectPTH(srcIP, hash, "192.168.1.1")
	assert.Nil(t, alert)

	alert = detector.DetectPTH(srcIP, hash, "192.168.1.2")
	assert.Nil(t, alert)

	alert = detector.DetectPTH(srcIP, hash, "192.168.1.3")
	assert.NotNil(t, alert)
	assert.Equal(t, "pass_the_hash", alert.Type)
	assert.Equal(t, "critical", alert.Severity)
}

func TestDetectRemoteExec(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	detector := NewLateralMovementDetector(logger, 20, 300)

	srcIP := "192.168.1.100"
	dstIP := "192.168.1.200"

	alert := detector.DetectRemoteExec(srcIP, dstIP, "admin_share")
	assert.Nil(t, alert)

	alert = detector.DetectRemoteExec(srcIP, dstIP, "svcctl")
	assert.NotNil(t, alert)
	assert.Equal(t, "psexec", alert.Type)
	assert.Equal(t, "critical", alert.Severity)
}

func TestCleanup(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	detector := NewLateralMovementDetector(logger, 20, 300)

	conn := &models.Connection{
		SrcIP:     "192.168.1.100",
		DstIP:     "192.168.1.1",
		Timestamp: time.Now().Add(-2 * time.Hour),
	}
	detector.DetectScan(conn)

	assert.Equal(t, 1, len(detector.scanTracker))

	detector.Cleanup()

	assert.Equal(t, 0, len(detector.scanTracker))
}
