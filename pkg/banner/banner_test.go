// Package banner provides ASCII art banner for TelemetryFlow Collector.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 DevOpsCorner Indonesia. All rights reserved.
//
// LEGO Building Block - Self-contained within tfo-collector project.
package banner

import (
	"strings"
	"testing"
)

// TestDefaultConfig tests the DefaultConfig function
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ProductName != "TelemetryFlow Collector" {
		t.Errorf("Expected ProductName 'TelemetryFlow Collector', got '%s'", cfg.ProductName)
	}

	if cfg.Version != "1.1.0" {
		t.Errorf("Expected Version '1.1.0', got '%s'", cfg.Version)
	}

	if cfg.BasedOn != "OTEL Collector 0.142.0" {
		t.Errorf("Expected BasedOn 'OTEL Collector 0.142.0', got '%s'", cfg.BasedOn)
	}

	if cfg.Motto != "Community Enterprise Observability Platform (CEOP)" {
		t.Errorf("Expected Motto 'Community Enterprise Observability Platform (CEOP)', got '%s'", cfg.Motto)
	}

	if cfg.GitCommit != "unknown" {
		t.Errorf("Expected GitCommit 'unknown', got '%s'", cfg.GitCommit)
	}

	if cfg.BuildTime != "unknown" {
		t.Errorf("Expected BuildTime 'unknown', got '%s'", cfg.BuildTime)
	}

	if cfg.GoVersion != "unknown" {
		t.Errorf("Expected GoVersion 'unknown', got '%s'", cfg.GoVersion)
	}

	if cfg.Platform != "unknown" {
		t.Errorf("Expected Platform 'unknown', got '%s'", cfg.Platform)
	}

	if cfg.Vendor != "TelemetryFlow" {
		t.Errorf("Expected Vendor 'TelemetryFlow', got '%s'", cfg.Vendor)
	}

	if cfg.VendorURL != "https://telemetryflow.id" {
		t.Errorf("Expected VendorURL 'https://telemetryflow.id', got '%s'", cfg.VendorURL)
	}

	if cfg.Developer != "DevOpsCorner Indonesia" {
		t.Errorf("Expected Developer 'DevOpsCorner Indonesia', got '%s'", cfg.Developer)
	}

	if cfg.License != "Apache-2.0" {
		t.Errorf("Expected License 'Apache-2.0', got '%s'", cfg.License)
	}

	if cfg.SupportURL != "https://docs.telemetryflow.id" {
		t.Errorf("Expected SupportURL 'https://docs.telemetryflow.id', got '%s'", cfg.SupportURL)
	}

	if cfg.Copyright != "Copyright (c) 2024-2026 DevOpsCorner Indonesia" {
		t.Errorf("Expected Copyright 'Copyright (c) 2024-2026 DevOpsCorner Indonesia', got '%s'", cfg.Copyright)
	}
}

// TestGenerate tests the Generate function
func TestGenerate(t *testing.T) {
	cfg := DefaultConfig()
	banner := Generate(cfg)

	if banner == "" {
		t.Error("Expected non-empty banner")
	}

	// Check for ASCII art elements
	if !strings.Contains(banner, "___") {
		t.Error("Expected banner to contain ASCII art")
	}

	// Check for product information
	mustContain := []string{
		cfg.ProductName,
		cfg.Version,
		cfg.BasedOn,
		cfg.Motto,
		cfg.Platform,
		cfg.GoVersion,
		cfg.GitCommit,
		cfg.BuildTime,
		cfg.Vendor,
		cfg.VendorURL,
		cfg.Developer,
		cfg.License,
		cfg.SupportURL,
		cfg.Copyright,
	}

	for _, substr := range mustContain {
		if !strings.Contains(banner, substr) {
			t.Errorf("Expected banner to contain '%s'", substr)
		}
	}

	// Check for separator lines
	if !strings.Contains(banner, "═") {
		t.Error("Expected banner to contain '═' separator")
	}

	if !strings.Contains(banner, "─") {
		t.Error("Expected banner to contain '─' separator")
	}
}

// TestGenerateWithoutBasedOn tests Generate without BasedOn field
func TestGenerateWithoutBasedOn(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BasedOn = ""

	banner := Generate(cfg)

	if strings.Contains(banner, "(Based on") {
		t.Error("Expected banner to not contain '(Based on' when BasedOn is empty")
	}
}

// TestGenerateWithCustomConfig tests Generate with custom configuration
func TestGenerateWithCustomConfig(t *testing.T) {
	cfg := Config{
		ProductName: "Custom Product",
		Version:     "2.0.0",
		BasedOn:     "Custom Base",
		Motto:       "Custom Motto",
		GitCommit:   "abc123",
		BuildTime:   "2024-01-01",
		GoVersion:   "go1.21",
		Platform:    "linux/amd64",
		Vendor:      "Custom Vendor",
		VendorURL:   "https://custom.com",
		Developer:   "Custom Dev",
		License:     "MIT",
		SupportURL:  "https://support.custom.com",
		Copyright:   "Copyright Custom",
	}

	banner := Generate(cfg)

	if !strings.Contains(banner, "Custom Product") {
		t.Error("Expected banner to contain 'Custom Product'")
	}

	if !strings.Contains(banner, "2.0.0") {
		t.Error("Expected banner to contain '2.0.0'")
	}

	if !strings.Contains(banner, "(Based on Custom Base)") {
		t.Error("Expected banner to contain '(Based on Custom Base)'")
	}

	if !strings.Contains(banner, "abc123") {
		t.Error("Expected banner to contain 'abc123'")
	}
}

// TestGenerateCompact tests the GenerateCompact function
func TestGenerateCompact(t *testing.T) {
	cfg := DefaultConfig()
	banner := GenerateCompact(cfg)

	if banner == "" {
		t.Error("Expected non-empty compact banner")
	}

	// Compact banner should contain key information
	mustContain := []string{
		cfg.ProductName,
		cfg.Version,
		cfg.Motto,
		cfg.Copyright,
	}

	for _, substr := range mustContain {
		if !strings.Contains(banner, substr) {
			t.Errorf("Expected compact banner to contain '%s'", substr)
		}
	}

	// Check for separator lines
	if !strings.Contains(banner, "═") {
		t.Error("Expected compact banner to contain '═' separator")
	}
}

// TestGenerateCompactWithoutBasedOn tests GenerateCompact without BasedOn
func TestGenerateCompactWithoutBasedOn(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BasedOn = ""

	banner := GenerateCompact(cfg)

	if strings.Contains(banner, "(Based on") {
		t.Error("Expected compact banner to not contain '(Based on' when BasedOn is empty")
	}
}

// TestGenerateCompactShorterThanFull tests that compact is shorter
func TestGenerateCompactShorterThanFull(t *testing.T) {
	cfg := DefaultConfig()

	full := Generate(cfg)
	compact := GenerateCompact(cfg)

	if len(compact) >= len(full) {
		t.Error("Expected compact banner to be shorter than full banner")
	}
}

// TestConfigStruct tests Config struct initialization
func TestConfigStruct(t *testing.T) {
	cfg := Config{
		ProductName: "Test",
		Version:     "1.0",
	}

	if cfg.ProductName != "Test" {
		t.Errorf("Expected ProductName 'Test', got '%s'", cfg.ProductName)
	}

	if cfg.Version != "1.0" {
		t.Errorf("Expected Version '1.0', got '%s'", cfg.Version)
	}

	// Unset fields should be empty
	if cfg.BasedOn != "" {
		t.Errorf("Expected BasedOn to be empty, got '%s'", cfg.BasedOn)
	}
}

// TestBannerContainsASCIIArt tests that banner contains ASCII art
func TestBannerContainsASCIIArt(t *testing.T) {
	cfg := DefaultConfig()
	banner := Generate(cfg)

	// Check for specific ASCII art patterns
	asciiPatterns := []string{
		"_________",
		"/",
		"\\",
		"|",
		"___  >",
	}

	for _, pattern := range asciiPatterns {
		if !strings.Contains(banner, pattern) {
			t.Errorf("Expected banner to contain ASCII pattern '%s'", pattern)
		}
	}
}

// TestBannerNewlines tests that banner has proper line breaks
func TestBannerNewlines(t *testing.T) {
	cfg := DefaultConfig()
	banner := Generate(cfg)

	lines := strings.Split(banner, "\n")
	if len(lines) < 20 {
		t.Errorf("Expected banner to have at least 20 lines, got %d", len(lines))
	}
}

// Benchmark tests
func BenchmarkGenerate(b *testing.B) {
	cfg := DefaultConfig()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Generate(cfg)
	}
}

func BenchmarkGenerateCompact(b *testing.B) {
	cfg := DefaultConfig()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = GenerateCompact(cfg)
	}
}

func BenchmarkDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DefaultConfig()
	}
}
