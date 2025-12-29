// Package config provides configuration loading for TelemetryFlow Collector.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 DevOpsCorner Indonesia. All rights reserved.
//
// LEGO Building Block - Flexible config loading from multiple sources.
package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSourceConstants tests Source constants
func TestSourceConstants(t *testing.T) {
	if SourceFile != "file" {
		t.Errorf("Expected SourceFile 'file', got '%s'", SourceFile)
	}

	if SourceEnv != "env" {
		t.Errorf("Expected SourceEnv 'env', got '%s'", SourceEnv)
	}

	if SourceRemote != "remote" {
		t.Errorf("Expected SourceRemote 'remote', got '%s'", SourceRemote)
	}
}

// TestNewLoader tests the NewLoader function
func TestNewLoader(t *testing.T) {
	loader := NewLoader()

	if loader == nil {
		t.Fatal("Expected non-nil loader")
	}

	if loader.envPrefix != "TFO_COLLECTOR" {
		t.Errorf("Expected envPrefix 'TFO_COLLECTOR', got '%s'", loader.envPrefix)
	}

	if len(loader.sources) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(loader.sources))
	}

	if len(loader.configPaths) != 4 {
		t.Errorf("Expected 4 config paths, got %d", len(loader.configPaths))
	}
}

// TestWithEnvPrefix tests the WithEnvPrefix option
func TestWithEnvPrefix(t *testing.T) {
	loader := NewLoader(WithEnvPrefix("CUSTOM_PREFIX"))

	if loader.envPrefix != "CUSTOM_PREFIX" {
		t.Errorf("Expected envPrefix 'CUSTOM_PREFIX', got '%s'", loader.envPrefix)
	}
}

// TestWithConfigPaths tests the WithConfigPaths option
func TestWithConfigPaths(t *testing.T) {
	paths := []string{"/custom/path1", "/custom/path2"}
	loader := NewLoader(WithConfigPaths(paths...))

	if len(loader.configPaths) != 2 {
		t.Errorf("Expected 2 config paths, got %d", len(loader.configPaths))
	}

	if loader.configPaths[0] != "/custom/path1" {
		t.Errorf("Expected first path '/custom/path1', got '%s'", loader.configPaths[0])
	}
}

// TestWithSources tests the WithSources option
func TestWithSources(t *testing.T) {
	loader := NewLoader(WithSources(SourceEnv))

	if len(loader.sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(loader.sources))
	}

	if loader.sources[0] != SourceEnv {
		t.Errorf("Expected source 'env', got '%s'", loader.sources[0])
	}
}

// TestMultipleOptions tests combining multiple options
func TestMultipleOptions(t *testing.T) {
	loader := NewLoader(
		WithEnvPrefix("MY_APP"),
		WithConfigPaths("/etc/myapp", "/home/user/.myapp"),
		WithSources(SourceFile, SourceEnv, SourceRemote),
	)

	if loader.envPrefix != "MY_APP" {
		t.Errorf("Expected envPrefix 'MY_APP', got '%s'", loader.envPrefix)
	}

	if len(loader.configPaths) != 2 {
		t.Errorf("Expected 2 config paths, got %d", len(loader.configPaths))
	}

	if len(loader.sources) != 3 {
		t.Errorf("Expected 3 sources, got %d", len(loader.sources))
	}
}

// TestGetEnv tests the GetEnv function
func TestGetEnv(t *testing.T) {
	loader := NewLoader(WithEnvPrefix("TEST_LOADER"))

	// Set test env var
	os.Setenv("TEST_LOADER_MY_VAR", "test_value")
	defer os.Unsetenv("TEST_LOADER_MY_VAR")

	value := loader.GetEnv("MY_VAR")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}

	// Test unset variable
	value = loader.GetEnv("UNSET_VAR")
	if value != "" {
		t.Errorf("Expected empty string for unset var, got '%s'", value)
	}
}

// TestGetEnvOrDefault tests the GetEnvOrDefault function
func TestGetEnvOrDefault(t *testing.T) {
	loader := NewLoader(WithEnvPrefix("TEST_LOADER"))

	// Set test env var
	os.Setenv("TEST_LOADER_EXISTS", "exists_value")
	defer os.Unsetenv("TEST_LOADER_EXISTS")

	// Should return env value when set
	value := loader.GetEnvOrDefault("EXISTS", "default")
	if value != "exists_value" {
		t.Errorf("Expected 'exists_value', got '%s'", value)
	}

	// Should return default when not set
	value = loader.GetEnvOrDefault("NOT_EXISTS", "my_default")
	if value != "my_default" {
		t.Errorf("Expected 'my_default', got '%s'", value)
	}
}

// TestLoad tests the Load function
func TestLoad(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  port: 8080
  host: localhost
database:
  name: testdb
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Define a struct to load into
	type TestConfig struct {
		Server struct {
			Port int    `yaml:"port"`
			Host string `yaml:"host"`
		} `yaml:"server"`
		Database struct {
			Name string `yaml:"name"`
		} `yaml:"database"`
	}

	loader := NewLoader()
	var cfg TestConfig

	err := loader.Load(configFile, &cfg)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Server.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", cfg.Server.Host)
	}

	if cfg.Database.Name != "testdb" {
		t.Errorf("Expected database 'testdb', got '%s'", cfg.Database.Name)
	}
}

// TestLoadNonExistentFile tests loading from non-existent file
func TestLoadNonExistentFile(t *testing.T) {
	loader := NewLoader()
	var cfg struct{}

	err := loader.Load("/non/existent/file.yaml", &cfg)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// TestLoadWithEnvExpansion tests environment variable expansion in config
func TestLoadWithEnvExpansion(t *testing.T) {
	// Create a temporary config file with env var
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	os.Setenv("TEST_DB_HOST", "db.example.com")
	defer os.Unsetenv("TEST_DB_HOST")

	configContent := `
database:
  host: ${TEST_DB_HOST}
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	type TestConfig struct {
		Database struct {
			Host string `yaml:"host"`
		} `yaml:"database"`
	}

	loader := NewLoader()
	var cfg TestConfig

	err := loader.Load(configFile, &cfg)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Database.Host != "db.example.com" {
		t.Errorf("Expected host 'db.example.com', got '%s'", cfg.Database.Host)
	}
}

// TestLoadWithEmptyPath tests loading with no config file specified
func TestLoadWithEmptyPath(t *testing.T) {
	loader := NewLoader(WithConfigPaths(t.TempDir()))
	var cfg struct{}

	// Should not error even if no config found (uses defaults)
	err := loader.Load("", &cfg)
	if err != nil {
		t.Errorf("Expected no error for empty path, got: %v", err)
	}
}

// TestMustLoad tests the MustLoad function
func TestMustLoad(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `key: value`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := NewLoader()
	var cfg map[string]string

	// Should not panic
	loader.MustLoad(configFile, &cfg)

	if cfg["key"] != "value" {
		t.Errorf("Expected key 'value', got '%s'", cfg["key"])
	}
}

// TestMustLoadPanics tests that MustLoad panics on error
func TestMustLoadPanics(t *testing.T) {
	loader := NewLoader()
	var cfg struct{}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected MustLoad to panic on non-existent file")
		}
	}()

	loader.MustLoad("/non/existent/file.yaml", &cfg)
}

// TestValidate tests the Validate function
func TestValidate(t *testing.T) {
	cfg := struct {
		Name string
	}{
		Name: "test",
	}

	err := Validate(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// TestDefaultLoader tests the default loader package functions
func TestDefaultLoader(t *testing.T) {
	if defaultLoader == nil {
		t.Error("Expected defaultLoader to be initialized")
	}
}

// TestLoadPackageFunction tests the Load package function
func TestLoadPackageFunction(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `key: value`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	var cfg map[string]string

	err := Load(configFile, &cfg)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg["key"] != "value" {
		t.Errorf("Expected key 'value', got '%s'", cfg["key"])
	}
}

// TestMustLoadPackageFunction tests the MustLoad package function
func TestMustLoadPackageFunction(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `key: value`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	var cfg map[string]string

	// Should not panic
	MustLoad(configFile, &cfg)

	if cfg["key"] != "value" {
		t.Errorf("Expected key 'value', got '%s'", cfg["key"])
	}
}

// TestLoaderStruct tests Loader struct fields
func TestLoaderStruct(t *testing.T) {
	loader := &Loader{
		sources:     []Source{SourceFile},
		envPrefix:   "MY_PREFIX",
		configPaths: []string{"/path1", "/path2"},
	}

	if len(loader.sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(loader.sources))
	}

	if loader.envPrefix != "MY_PREFIX" {
		t.Errorf("Expected envPrefix 'MY_PREFIX', got '%s'", loader.envPrefix)
	}

	if len(loader.configPaths) != 2 {
		t.Errorf("Expected 2 config paths, got %d", len(loader.configPaths))
	}
}

// Benchmark tests
func BenchmarkNewLoader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewLoader()
	}
}

func BenchmarkNewLoaderWithOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewLoader(
			WithEnvPrefix("BENCH"),
			WithConfigPaths("/path1", "/path2"),
			WithSources(SourceFile, SourceEnv),
		)
	}
}

func BenchmarkGetEnv(b *testing.B) {
	loader := NewLoader(WithEnvPrefix("BENCH"))
	os.Setenv("BENCH_TEST_KEY", "test_value")
	defer os.Unsetenv("BENCH_TEST_KEY")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = loader.GetEnv("TEST_KEY")
	}
}

func BenchmarkGetEnvOrDefault(b *testing.B) {
	loader := NewLoader(WithEnvPrefix("BENCH"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = loader.GetEnvOrDefault("MISSING_KEY", "default")
	}
}
