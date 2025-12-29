// Package config_test provides unit tests for the pkg/config loader.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 DevOpsCorner Indonesia. All rights reserved.
package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/telemetryflow/telemetryflow-collector/pkg/config"
)

func TestSourceConstants(t *testing.T) {
	assert.Equal(t, config.Source("file"), config.SourceFile)
	assert.Equal(t, config.Source("env"), config.SourceEnv)
	assert.Equal(t, config.Source("remote"), config.SourceRemote)
}

func TestNewLoader(t *testing.T) {
	loader := config.NewLoader()

	require.NotNil(t, loader)
}

func TestWithEnvPrefix(t *testing.T) {
	loader := config.NewLoader(config.WithEnvPrefix("CUSTOM_PREFIX"))

	require.NotNil(t, loader)
}

func TestWithConfigPaths(t *testing.T) {
	paths := []string{"/custom/path1", "/custom/path2"}
	loader := config.NewLoader(config.WithConfigPaths(paths...))

	require.NotNil(t, loader)
}

func TestWithSources(t *testing.T) {
	loader := config.NewLoader(config.WithSources(config.SourceEnv))

	require.NotNil(t, loader)
}

func TestMultipleOptions(t *testing.T) {
	loader := config.NewLoader(
		config.WithEnvPrefix("MY_APP"),
		config.WithConfigPaths("/etc/myapp", "/home/user/.myapp"),
		config.WithSources(config.SourceFile, config.SourceEnv, config.SourceRemote),
	)

	require.NotNil(t, loader)
}

func TestGetEnv(t *testing.T) {
	loader := config.NewLoader(config.WithEnvPrefix("TEST_LOADER"))

	require.NoError(t, os.Setenv("TEST_LOADER_MY_VAR", "test_value"))
	defer func() { _ = os.Unsetenv("TEST_LOADER_MY_VAR") }()

	value := loader.GetEnv("MY_VAR")
	assert.Equal(t, "test_value", value)

	value = loader.GetEnv("UNSET_VAR")
	assert.Empty(t, value)
}

func TestGetEnvOrDefault(t *testing.T) {
	loader := config.NewLoader(config.WithEnvPrefix("TEST_LOADER"))

	require.NoError(t, os.Setenv("TEST_LOADER_EXISTS", "exists_value"))
	defer func() { _ = os.Unsetenv("TEST_LOADER_EXISTS") }()

	value := loader.GetEnvOrDefault("EXISTS", "default")
	assert.Equal(t, "exists_value", value)

	value = loader.GetEnvOrDefault("NOT_EXISTS", "my_default")
	assert.Equal(t, "my_default", value)
}

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  port: 8080
  host: localhost
database:
  name: testdb
`
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	type TestConfig struct {
		Server struct {
			Port int    `yaml:"port"`
			Host string `yaml:"host"`
		} `yaml:"server"`
		Database struct {
			Name string `yaml:"name"`
		} `yaml:"database"`
	}

	loader := config.NewLoader()
	var cfg TestConfig

	err := loader.Load(configFile, &cfg)
	require.NoError(t, err)

	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, "testdb", cfg.Database.Name)
}

func TestLoadNonExistentFile(t *testing.T) {
	loader := config.NewLoader()
	var cfg struct{}

	err := loader.Load("/non/existent/file.yaml", &cfg)
	assert.Error(t, err)
}

func TestLoadWithEnvExpansion(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	require.NoError(t, os.Setenv("TEST_DB_HOST", "db.example.com"))
	defer func() { _ = os.Unsetenv("TEST_DB_HOST") }()

	configContent := `
database:
  host: ${TEST_DB_HOST}
`
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	type TestConfig struct {
		Database struct {
			Host string `yaml:"host"`
		} `yaml:"database"`
	}

	loader := config.NewLoader()
	var cfg TestConfig

	err := loader.Load(configFile, &cfg)
	require.NoError(t, err)

	assert.Equal(t, "db.example.com", cfg.Database.Host)
}

func TestLoadWithEmptyPath(t *testing.T) {
	loader := config.NewLoader(config.WithConfigPaths(t.TempDir()))
	var cfg struct{}

	err := loader.Load("", &cfg)
	assert.NoError(t, err)
}

func TestMustLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `key: value`
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	loader := config.NewLoader()
	var cfg map[string]string

	loader.MustLoad(configFile, &cfg)

	assert.Equal(t, "value", cfg["key"])
}

func TestMustLoadPanics(t *testing.T) {
	loader := config.NewLoader()
	var cfg struct{}

	defer func() {
		r := recover()
		assert.NotNil(t, r)
	}()

	loader.MustLoad("/non/existent/file.yaml", &cfg)
}

func TestValidate(t *testing.T) {
	cfg := struct {
		Name string
	}{
		Name: "test",
	}

	err := config.Validate(cfg)
	assert.NoError(t, err)
}

func TestLoadPackageFunction(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `key: value`
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	var cfg map[string]string

	err := config.Load(configFile, &cfg)
	require.NoError(t, err)

	assert.Equal(t, "value", cfg["key"])
}

func TestMustLoadPackageFunction(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `key: value`
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

	var cfg map[string]string

	config.MustLoad(configFile, &cfg)

	assert.Equal(t, "value", cfg["key"])
}

func TestLoaderOptions(t *testing.T) {
	loader := config.NewLoader(
		config.WithSources(config.SourceFile),
		config.WithEnvPrefix("MY_PREFIX"),
		config.WithConfigPaths("/path1", "/path2"),
	)

	require.NotNil(t, loader)
}

// Benchmark tests
func BenchmarkNewLoader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = config.NewLoader()
	}
}

func BenchmarkNewLoaderWithOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = config.NewLoader(
			config.WithEnvPrefix("BENCH"),
			config.WithConfigPaths("/path1", "/path2"),
			config.WithSources(config.SourceFile, config.SourceEnv),
		)
	}
}

func BenchmarkGetEnv(b *testing.B) {
	loader := config.NewLoader(config.WithEnvPrefix("BENCH"))
	_ = os.Setenv("BENCH_TEST_KEY", "test_value")
	defer func() { _ = os.Unsetenv("BENCH_TEST_KEY") }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = loader.GetEnv("TEST_KEY")
	}
}

func BenchmarkGetEnvOrDefault(b *testing.B) {
	loader := config.NewLoader(config.WithEnvPrefix("BENCH"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = loader.GetEnvOrDefault("MISSING_KEY", "default")
	}
}
