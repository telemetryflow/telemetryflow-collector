// Package banner_test provides unit tests for the banner package.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 DevOpsCorner Indonesia. All rights reserved.
package banner_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/telemetryflow/telemetryflow-collector/pkg/banner"
)

func TestDefaultConfig(t *testing.T) {
	cfg := banner.DefaultConfig()

	assert.Equal(t, "TelemetryFlow Collector", cfg.ProductName)
	assert.Equal(t, "1.1.1", cfg.Version)
	assert.Equal(t, "OTEL Collector 0.142.0", cfg.BasedOn)
	assert.Equal(t, "Community Enterprise Observability Platform (CEOP)", cfg.Motto)
	assert.Equal(t, "unknown", cfg.GitCommit)
	assert.Equal(t, "unknown", cfg.BuildTime)
	assert.Equal(t, "unknown", cfg.GoVersion)
	assert.Equal(t, "unknown", cfg.Platform)
	assert.Equal(t, "TelemetryFlow", cfg.Vendor)
	assert.Equal(t, "https://telemetryflow.id", cfg.VendorURL)
	assert.Equal(t, "DevOpsCorner Indonesia", cfg.Developer)
	assert.Equal(t, "Apache-2.0", cfg.License)
	assert.Equal(t, "https://docs.telemetryflow.id", cfg.SupportURL)
	assert.Equal(t, "Copyright (c) 2024-2026 DevOpsCorner Indonesia", cfg.Copyright)
}

func TestGenerate(t *testing.T) {
	cfg := banner.DefaultConfig()
	b := banner.Generate(cfg)

	require.NotEmpty(t, b)

	assert.Contains(t, b, "___")

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
		assert.Contains(t, b, substr)
	}

	assert.Contains(t, b, "═")
	assert.Contains(t, b, "─")
}

func TestGenerateWithoutBasedOn(t *testing.T) {
	cfg := banner.DefaultConfig()
	cfg.BasedOn = ""

	b := banner.Generate(cfg)

	assert.NotContains(t, b, "(Based on")
}

func TestGenerateWithCustomConfig(t *testing.T) {
	cfg := banner.Config{
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

	b := banner.Generate(cfg)

	assert.Contains(t, b, "Custom Product")
	assert.Contains(t, b, "2.0.0")
	assert.Contains(t, b, "(Based on Custom Base)")
	assert.Contains(t, b, "abc123")
}

func TestGenerateCompact(t *testing.T) {
	cfg := banner.DefaultConfig()
	b := banner.GenerateCompact(cfg)

	require.NotEmpty(t, b)

	mustContain := []string{
		cfg.ProductName,
		cfg.Version,
		cfg.Motto,
		cfg.Copyright,
	}

	for _, substr := range mustContain {
		assert.Contains(t, b, substr)
	}

	assert.Contains(t, b, "═")
}

func TestGenerateCompactWithoutBasedOn(t *testing.T) {
	cfg := banner.DefaultConfig()
	cfg.BasedOn = ""

	b := banner.GenerateCompact(cfg)

	assert.NotContains(t, b, "(Based on")
}

func TestGenerateCompactShorterThanFull(t *testing.T) {
	cfg := banner.DefaultConfig()

	full := banner.Generate(cfg)
	compact := banner.GenerateCompact(cfg)

	assert.Less(t, len(compact), len(full))
}

func TestConfigStruct(t *testing.T) {
	cfg := banner.Config{
		ProductName: "Test",
		Version:     "1.0",
	}

	assert.Equal(t, "Test", cfg.ProductName)
	assert.Equal(t, "1.0", cfg.Version)
	assert.Empty(t, cfg.BasedOn)
}

func TestBannerContainsASCIIArt(t *testing.T) {
	cfg := banner.DefaultConfig()
	b := banner.Generate(cfg)

	asciiPatterns := []string{
		"_________",
		"/",
		"\\",
		"|",
		"___  >",
	}

	for _, pattern := range asciiPatterns {
		assert.Contains(t, b, pattern)
	}
}

func TestBannerNewlines(t *testing.T) {
	cfg := banner.DefaultConfig()
	b := banner.Generate(cfg)

	lines := strings.Split(b, "\n")
	assert.GreaterOrEqual(t, len(lines), 20)
}

// Benchmark tests
func BenchmarkGenerate(b *testing.B) {
	cfg := banner.DefaultConfig()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = banner.Generate(cfg)
	}
}

func BenchmarkGenerateCompact(b *testing.B) {
	cfg := banner.DefaultConfig()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = banner.GenerateCompact(cfg)
	}
}

func BenchmarkDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = banner.DefaultConfig()
	}
}
