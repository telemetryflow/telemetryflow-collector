// Package config provides configuration management for the TelemetryFlow Collector.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// LEGO Building Block - Self-contained within tfo-collector project.
package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewLoader tests the NewLoader function
func TestNewLoader(t *testing.T) {
	loader := NewLoader()

	if loader == nil {
		t.Fatal("Expected non-nil loader")
	}

	// Check default config paths
	expectedPaths := []string{
		".",
		"./configs",
		"/etc/tfo-collector",
		"$HOME/.tfo-collector",
	}

	if len(loader.configPaths) != len(expectedPaths) {
		t.Errorf("Expected %d config paths, got %d", len(expectedPaths), len(loader.configPaths))
	}

	for i, path := range expectedPaths {
		if loader.configPaths[i] != path {
			t.Errorf("Expected path[%d] '%s', got '%s'", i, path, loader.configPaths[i])
		}
	}

	// Check default env prefix
	if loader.envPrefix != "TFCOLLECTOR" {
		t.Errorf("Expected envPrefix 'TFCOLLECTOR', got '%s'", loader.envPrefix)
	}
}

// TestWithConfigPaths tests the WithConfigPaths method
func TestWithConfigPaths(t *testing.T) {
	loader := NewLoader()
	originalCount := len(loader.configPaths)

	result := loader.WithConfigPaths("/custom/path1", "/custom/path2")

	// Should return same loader (fluent interface)
	if result != loader {
		t.Error("Expected WithConfigPaths to return the same loader instance")
	}

	// Should add paths
	expectedCount := originalCount + 2
	if len(loader.configPaths) != expectedCount {
		t.Errorf("Expected %d config paths, got %d", expectedCount, len(loader.configPaths))
	}

	// Verify new paths are added
	if loader.configPaths[originalCount] != "/custom/path1" {
		t.Errorf("Expected path '/custom/path1', got '%s'", loader.configPaths[originalCount])
	}

	if loader.configPaths[originalCount+1] != "/custom/path2" {
		t.Errorf("Expected path '/custom/path2', got '%s'", loader.configPaths[originalCount+1])
	}
}

// TestWithEnvPrefix tests the WithEnvPrefix method
func TestWithEnvPrefix(t *testing.T) {
	loader := NewLoader()

	result := loader.WithEnvPrefix("CUSTOM_PREFIX")

	// Should return same loader (fluent interface)
	if result != loader {
		t.Error("Expected WithEnvPrefix to return the same loader instance")
	}

	if loader.envPrefix != "CUSTOM_PREFIX" {
		t.Errorf("Expected envPrefix 'CUSTOM_PREFIX', got '%s'", loader.envPrefix)
	}
}

// TestFluentInterface tests the fluent interface pattern
func TestFluentInterface(t *testing.T) {
	loader := NewLoader().
		WithConfigPaths("/path1", "/path2").
		WithEnvPrefix("MY_APP")

	if loader.envPrefix != "MY_APP" {
		t.Errorf("Expected envPrefix 'MY_APP', got '%s'", loader.envPrefix)
	}

	// Original 4 + 2 new paths
	if len(loader.configPaths) != 6 {
		t.Errorf("Expected 6 config paths, got %d", len(loader.configPaths))
	}
}

// TestLoadWithValidConfig tests loading a valid config file
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
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(configFile)

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}

	if cfg.Collector.Name != "Test Collector" {
		t.Errorf("Expected name 'Test Collector', got '%s'", cfg.Collector.Name)
	}

	if cfg.Collector.Description != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", cfg.Collector.Description)
	}
}

// TestLoadWithEnvOverrides tests environment variable overrides
func TestLoadWithEnvOverrides(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("TELEMETRYFLOW_LOG_LEVEL", "debug")
	_ = os.Setenv("TELEMETRYFLOW_COLLECTOR_NAME", "Env Collector")
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
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(configFile)

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Environment variables should override file values
	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug' from env, got '%s'", cfg.Logging.Level)
	}

	if cfg.Collector.Name != "Env Collector" {
		t.Errorf("Expected name 'Env Collector' from env, got '%s'", cfg.Collector.Name)
	}
}

// TestLoadWithEmptyPath tests loading with no config file specified
func TestLoadWithEmptyPath(t *testing.T) {
	// Use temp dir to avoid finding any existing config
	tmpDir := t.TempDir()
	loader := NewLoader().WithConfigPaths(tmpDir)

	cfg, err := loader.Load("")

	// Should succeed with defaults when no config found
	if err != nil {
		t.Fatalf("Expected no error for empty path, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected non-nil config with defaults")
	}

	// Should have default values
	if cfg.Receivers.OTLP.Protocols.GRPC.Endpoint != "0.0.0.0:4317" {
		t.Errorf("Expected default GRPC endpoint '0.0.0.0:4317', got '%s'", cfg.Receivers.OTLP.Protocols.GRPC.Endpoint)
	}
}

// TestLoadWithInvalidYAML tests loading an invalid YAML file
func TestLoadWithInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	if err := os.WriteFile(configFile, []byte("invalid: yaml: content: ["), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := NewLoader()
	_, err := loader.Load(configFile)

	if err == nil {
		t.Error("Expected error for invalid YAML file")
	}
}

// TestLoadFromFile tests the LoadFromFile method
func TestLoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `
collector:
  name: "LoadFromFile Test"
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := NewLoader()

	t.Run("valid file path", func(t *testing.T) {
		cfg, err := loader.LoadFromFile(configFile)

		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if cfg.Collector.Name != "LoadFromFile Test" {
			t.Errorf("Expected name 'LoadFromFile Test', got '%s'", cfg.Collector.Name)
		}
	})

	t.Run("relative path conversion", func(t *testing.T) {
		// Create config in current directory
		cwd, _ := os.Getwd()
		relativeFile := filepath.Join(cwd, "test_config_temp.yaml")
		if err := os.WriteFile(relativeFile, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}
		defer func() { _ = os.Remove(relativeFile) }()

		cfg, err := loader.LoadFromFile("test_config_temp.yaml")
		if err != nil {
			t.Fatalf("Failed to load config from relative path: %v", err)
		}

		if cfg == nil {
			t.Fatal("Expected non-nil config")
		}
	})
}

// TestLoadWithHostnameAutoDetect tests hostname auto-detection
func TestLoadWithHostnameAutoDetect(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Config without hostname
	configContent := `
collector:
  name: "Test Collector"
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(configFile)

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Hostname should be auto-detected
	expectedHostname, _ := os.Hostname()
	if cfg.Collector.Hostname != expectedHostname {
		t.Errorf("Expected hostname '%s', got '%s'", expectedHostname, cfg.Collector.Hostname)
	}
}

// TestGetConfigFilePath tests the GetConfigFilePath function
func TestGetConfigFilePath(t *testing.T) {
	// Note: This function returns the path from viper's global state
	// We just verify it doesn't panic
	path := GetConfigFilePath()
	_ = path // Path may be empty if no config loaded
}

// TestLoaderStruct tests Loader struct initialization
func TestLoaderStruct(t *testing.T) {
	loader := &Loader{
		configPaths: []string{"/path1", "/path2"},
		envPrefix:   "TEST_PREFIX",
	}

	if len(loader.configPaths) != 2 {
		t.Errorf("Expected 2 config paths, got %d", len(loader.configPaths))
	}

	if loader.envPrefix != "TEST_PREFIX" {
		t.Errorf("Expected envPrefix 'TEST_PREFIX', got '%s'", loader.envPrefix)
	}
}

// TestLoadWithDefaultValues tests that defaults are applied
func TestLoadWithDefaultValues(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "minimal.yaml")

	// Minimal config
	configContent := `
collector:
  name: "Minimal"
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(configFile)

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check defaults are applied
	if !cfg.Receivers.OTLP.Enabled {
		t.Error("Expected OTLP receiver to be enabled by default")
	}

	if !cfg.Receivers.OTLP.Protocols.GRPC.Enabled {
		t.Error("Expected GRPC to be enabled by default")
	}

	if !cfg.Receivers.OTLP.Protocols.HTTP.Enabled {
		t.Error("Expected HTTP to be enabled by default")
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", cfg.Logging.Level)
	}

	if cfg.Logging.Format != "json" {
		t.Errorf("Expected default log format 'json', got '%s'", cfg.Logging.Format)
	}
}

// TestLoadWithTelemetryFlowConfig tests TelemetryFlow specific config
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
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.Load(configFile)

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if !cfg.TelemetryFlow.Enabled {
		t.Error("Expected TelemetryFlow to be enabled")
	}

	if cfg.TelemetryFlow.APIKeyID != "test-key-id" {
		t.Errorf("Expected API key ID 'test-key-id', got '%s'", cfg.TelemetryFlow.APIKeyID)
	}

	if cfg.TelemetryFlow.Endpoint != "https://api.telemetryflow.io" {
		t.Errorf("Expected endpoint 'https://api.telemetryflow.io', got '%s'", cfg.TelemetryFlow.Endpoint)
	}
}

// Benchmark tests
func BenchmarkNewLoader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewLoader()
	}
}

func BenchmarkLoaderWithOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewLoader().
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

	loader := NewLoader()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = loader.Load(configFile)
	}
}
