// Package config_test provides unit tests for the config loader.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/telemetryflow/telemetryflow-collector/internal/config"
)

func TestNewLoader(t *testing.T) {
	loader := config.NewLoader()

	require.NotNil(t, loader)
}

func TestWithConfigPaths(t *testing.T) {
	loader := config.NewLoader()

	result := loader.WithConfigPaths("/custom/path1", "/custom/path2")

	// Should return same loader (fluent interface)
	assert.Equal(t, loader, result)
}

func TestWithEnvPrefix(t *testing.T) {
	loader := config.NewLoader()

	result := loader.WithEnvPrefix("CUSTOM_PREFIX")

	// Should return same loader (fluent interface)
	assert.Equal(t, loader, result)
}

func TestFluentInterface(t *testing.T) {
	loader := config.NewLoader().
		WithConfigPaths("/path1", "/path2").
		WithEnvPrefix("MY_APP")

	require.NotNil(t, loader)
}

func TestLoadWithValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "tfo-collector.yaml")

	configContent := `
collector:
  name: "Test Collector"
  description: "Test Description"
receivers:
  otlp:
    enabled: true
    protocols:
      grpc:
        enabled: true
        endpoint: "0.0.0.0:4317"
      http:
        enabled: true
        endpoint: "0.0.0.0:4318"
`
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	loader := config.NewLoader()
	cfg, err := loader.Load(configFile)

	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "Test Collector", cfg.Collector.Name)
	assert.Equal(t, "Test Description", cfg.Collector.Description)
}

func TestLoadWithEnvOverrides(t *testing.T) {
	// Set environment variables
	require.NoError(t, os.Setenv("TELEMETRYFLOW_LOG_LEVEL", "debug"))
	require.NoError(t, os.Setenv("TELEMETRYFLOW_COLLECTOR_NAME", "Env Collector"))
	defer func() {
		_ = os.Unsetenv("TELEMETRYFLOW_LOG_LEVEL")
		_ = os.Unsetenv("TELEMETRYFLOW_COLLECTOR_NAME")
	}()

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "tfo-collector.yaml")

	configContent := `
collector:
  name: "File Collector"
logging:
  level: "info"
`
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	loader := config.NewLoader()
	cfg, err := loader.Load(configFile)

	require.NoError(t, err)

	// Environment variables should override file values
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "Env Collector", cfg.Collector.Name)
}

func TestLoadWithEmptyPath(t *testing.T) {
	// Use temp dir to avoid finding any existing config
	tmpDir := t.TempDir()
	loader := config.NewLoader().WithConfigPaths(tmpDir)

	cfg, err := loader.Load("")

	// Should succeed with defaults when no config found
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Should have default values
	assert.Equal(t, "0.0.0.0:4317", cfg.Receivers.OTLP.Protocols.GRPC.Endpoint)
}

func TestLoadWithInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	require.NoError(t, os.WriteFile(configFile, []byte("invalid: yaml: content: ["), 0644))

	loader := config.NewLoader()
	_, err := loader.Load(configFile)

	assert.Error(t, err)
}

func TestLoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `
collector:
  name: "LoadFromFile Test"
`
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	loader := config.NewLoader()

	t.Run("valid file path", func(t *testing.T) {
		cfg, err := loader.LoadFromFile(configFile)

		require.NoError(t, err)
		assert.Equal(t, "LoadFromFile Test", cfg.Collector.Name)
	})

	t.Run("relative path conversion", func(t *testing.T) {
		// Create config in current directory
		cwd, _ := os.Getwd()
		relativeFile := filepath.Join(cwd, "test_config_temp.yaml")
		require.NoError(t, os.WriteFile(relativeFile, []byte(configContent), 0644))
		defer func() { _ = os.Remove(relativeFile) }()

		cfg, err := loader.LoadFromFile("test_config_temp.yaml")
		require.NoError(t, err)
		require.NotNil(t, cfg)
	})
}

func TestLoadWithHostnameAutoDetect(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Config without hostname
	configContent := `
collector:
  name: "Test Collector"
`
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	loader := config.NewLoader()
	cfg, err := loader.Load(configFile)

	require.NoError(t, err)

	// Hostname should be auto-detected
	expectedHostname, _ := os.Hostname()
	assert.Equal(t, expectedHostname, cfg.Collector.Hostname)
}

func TestGetConfigFilePath(t *testing.T) {
	// Note: This function returns the path from viper's global state
	// We just verify it doesn't panic
	path := config.GetConfigFilePath()
	_ = path // Path may be empty if no config loaded
}

func TestLoadWithDefaultValues(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "minimal.yaml")

	// Minimal config
	configContent := `
collector:
  name: "Minimal"
`
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	loader := config.NewLoader()
	cfg, err := loader.Load(configFile)

	require.NoError(t, err)

	// Check defaults are applied
	assert.True(t, cfg.Receivers.OTLP.Enabled)
	assert.True(t, cfg.Receivers.OTLP.Protocols.GRPC.Enabled)
	assert.True(t, cfg.Receivers.OTLP.Protocols.HTTP.Enabled)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
}

func TestLoadWithTelemetryFlowConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "telemetryflow.yaml")

	configContent := `
telemetryflow:
  enabled: true
  api_key_id: "test-key-id"
  api_key_secret: "test-key-secret"
  endpoint: "https://api.telemetryflow.io"
collector:
  name: "TF Collector"
`
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	loader := config.NewLoader()
	cfg, err := loader.Load(configFile)

	require.NoError(t, err)

	assert.True(t, cfg.TelemetryFlow.Enabled)
	assert.Equal(t, "test-key-id", cfg.TelemetryFlow.APIKeyID)
	assert.Equal(t, "https://api.telemetryflow.io", cfg.TelemetryFlow.Endpoint)
}

// Benchmark tests
func BenchmarkNewLoader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = config.NewLoader()
	}
}

func BenchmarkLoaderWithOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = config.NewLoader().
			WithConfigPaths("/path1", "/path2").
			WithEnvPrefix("BENCH")
	}
}

func BenchmarkLoad(b *testing.B) {
	tmpDir := b.TempDir()
	configFile := filepath.Join(tmpDir, "bench.yaml")

	configContent := `
collector:
  name: "Benchmark"
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		b.Fatalf("Failed to write config file: %v", err)
	}

	loader := config.NewLoader()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = loader.Load(configFile)
	}
}
