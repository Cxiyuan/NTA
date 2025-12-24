package detector

import (
	"testing"
)

func TestDetectDGA(t *testing.T) {
	detector := NewAdvancedDetector(nil)

	tests := []struct {
		domain   string
		expected bool
	}{
		{"google.com", false},
		{"facebook.com", false},
		{"jksdh3kjh4kj3h4.com", true},
		{"randomdga123456.net", true},
		{"a1b2c3d4e5f6g7.xyz", true},
	}

	for _, tt := range tests {
		isDGA, score := detector.DetectDGA(tt.domain)
		if isDGA != tt.expected {
			t.Errorf("DetectDGA(%s) = %v (score: %.2f), want %v", tt.domain, isDGA, score, tt.expected)
		}
	}
}

func TestCalculateEntropy(t *testing.T) {
	detector := NewAdvancedDetector(nil)

	tests := []struct {
		input    string
		minScore float64
	}{
		{"aaaaaaa", 0.0},
		{"abcdefg", 2.5},
		{"j3kh4j5k6h7k8j9", 3.0},
	}

	for _, tt := range tests {
		entropy := detector.calculateEntropy(tt.input)
		if entropy < tt.minScore {
			t.Errorf("calculateEntropy(%s) = %.2f, want >= %.2f", tt.input, entropy, tt.minScore)
		}
	}
}
