// Package collector_test provides unit tests for the collector package.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package collector_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/telemetryflow/telemetryflow-collector/internal/collector"
	"github.com/telemetryflow/telemetryflow-collector/internal/config"
)

// Helper function to get a free port
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// createTestConfig creates a test configuration with free ports
func createTestConfig(t *testing.T) *config.Config {
	grpcPort, err := getFreePort()
	if err != nil {
		t.Fatalf("Failed to get free port: %v", err)
	}

	httpPort, err := getFreePort()
	if err != nil {
		t.Fatalf("Failed to get free port: %v", err)
	}

	healthPort, err := getFreePort()
	if err != nil {
		t.Fatalf("Failed to get free port: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.Collector.ID = "test-collector"
	cfg.Collector.Hostname = "test-host"
	cfg.Collector.Name = "Test Collector"

	cfg.Receivers.OTLP.Enabled = true
	cfg.Receivers.OTLP.Protocols.GRPC.Enabled = true
	cfg.Receivers.OTLP.Protocols.GRPC.Endpoint = fmt.Sprintf("localhost:%d", grpcPort)
	cfg.Receivers.OTLP.Protocols.HTTP.Enabled = true
	cfg.Receivers.OTLP.Protocols.HTTP.Endpoint = fmt.Sprintf("localhost:%d", httpPort)

	cfg.Extensions.Health.Enabled = true
	cfg.Extensions.Health.Endpoint = fmt.Sprintf("localhost:%d", healthPort)
	cfg.Extensions.Health.Path = "/health"

	cfg.Exporters.Debug.Verbosity = "basic"

	return cfg
}

// createMinimalConfig creates a minimal test configuration
func createMinimalConfig(_ *testing.T) *config.Config {
	cfg := config.DefaultConfig()
	cfg.Collector.ID = "minimal-collector"
	cfg.Collector.Hostname = "minimal-host"

	cfg.Receivers.OTLP.Enabled = false
	cfg.Receivers.Prometheus.Enabled = false
	cfg.Extensions.Health.Enabled = false
	cfg.Exporters.Debug.Verbosity = ""

	return cfg
}

func TestNewCollector(t *testing.T) {
	t.Run("should create collector with valid config", func(t *testing.T) {
		cfg := config.DefaultConfig()
		logger, _ := zap.NewDevelopment()

		c, err := collector.New(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, c)
		assert.NotEmpty(t, c.ID())
	})

	t.Run("should generate ID if not provided", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Collector.ID = ""
		logger, _ := zap.NewDevelopment()

		c, err := collector.New(cfg, logger)
		require.NoError(t, err)
		assert.NotEmpty(t, c.ID())
	})

	t.Run("full config", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		cfg := createTestConfig(t)

		c, err := collector.New(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, c)
		assert.NotEmpty(t, c.ID())
	})

	t.Run("minimal config", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		cfg := createMinimalConfig(t)

		c, err := collector.New(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("with auto-generated ID", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		cfg := createMinimalConfig(t)
		cfg.Collector.ID = ""

		c, err := collector.New(cfg, logger)
		require.NoError(t, err)
		assert.NotEmpty(t, c.ID())
		assert.Len(t, c.ID(), 36) // UUID format
	})

	t.Run("with debug exporter enabled", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		cfg := createMinimalConfig(t)
		cfg.Exporters.Debug.Verbosity = "detailed"

		c, err := collector.New(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("without debug exporter", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		cfg := createMinimalConfig(t)
		cfg.Exporters.Debug.Verbosity = ""

		c, err := collector.New(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("with OTLP receiver", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		cfg := createTestConfig(t)

		c, err := collector.New(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, c)
	})

	t.Run("without OTLP receiver", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		cfg := createMinimalConfig(t)
		cfg.Receivers.OTLP.Enabled = false

		c, err := collector.New(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, c)
	})
}

func TestCollectorID(t *testing.T) {
	t.Run("returns configured ID", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		cfg := createMinimalConfig(t)
		cfg.Collector.ID = "test-id-12345"

		c, err := collector.New(cfg, logger)
		require.NoError(t, err)
		assert.Equal(t, "test-id-12345", c.ID())
	})

	t.Run("auto-generates UUID if not configured", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		cfg := createMinimalConfig(t)
		cfg.Collector.ID = ""

		c, err := collector.New(cfg, logger)
		require.NoError(t, err)
		assert.NotEmpty(t, c.ID())
		assert.Len(t, c.ID(), 36) // UUID format
	})
}

func TestCollectorIsRunning(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := createMinimalConfig(t)

	c, err := collector.New(cfg, logger)
	require.NoError(t, err)
	assert.False(t, c.IsRunning())
}

func TestCollectorUptime(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := createMinimalConfig(t)

	c, err := collector.New(cfg, logger)
	require.NoError(t, err)
	assert.Equal(t, time.Duration(0), c.Uptime())
}

func TestCollectorStats(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := createMinimalConfig(t)
	cfg.Collector.Hostname = "stats-test-host"

	c, err := collector.New(cfg, logger)
	require.NoError(t, err)

	stats := c.Stats()

	assert.Equal(t, c.ID(), stats.ID)
	assert.Equal(t, "stats-test-host", stats.Hostname)
	assert.False(t, stats.Running)
}

func TestCollectorRunAndShutdown(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := createTestConfig(t)

	c, err := collector.New(cfg, logger)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- c.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	assert.True(t, c.IsRunning())
	assert.Greater(t, c.Uptime(), time.Duration(0))

	cancel()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for collector to stop")
	}

	assert.False(t, c.IsRunning())
}

func TestCollectorRunAlreadyRunning(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := createTestConfig(t)

	c, err := collector.New(cfg, logger)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = c.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	err = c.Run(context.Background())
	assert.Error(t, err)
}

func TestCollectorHealthServer(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := createTestConfig(t)

	c, err := collector.New(cfg, logger)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = c.Run(ctx)
	}()

	time.Sleep(200 * time.Millisecond)

	healthURL := fmt.Sprintf("http://%s%s", cfg.Extensions.Health.Endpoint, cfg.Extensions.Health.Path)
	resp, err := http.Get(healthURL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	statsURL := fmt.Sprintf("http://%s/stats", cfg.Extensions.Health.Endpoint)
	resp, err = http.Get(statsURL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestCollectorStatsStruct(t *testing.T) {
	now := time.Now()
	stats := collector.CollectorStats{
		ID:       "test-id",
		Hostname: "test-host",
		Running:  true,
		Started:  now,
		Uptime:   5 * time.Minute,
		ReceiverStats: collector.ReceiverStats{
			TracesReceived:  100,
			MetricsReceived: 200,
			LogsReceived:    300,
		},
		PipelineStats: collector.PipelineStats{
			TracesProcessed:  90,
			MetricsProcessed: 190,
			LogsProcessed:    290,
		},
	}

	assert.Equal(t, "test-id", stats.ID)
	assert.Equal(t, "test-host", stats.Hostname)
	assert.True(t, stats.Running)
	assert.Equal(t, now, stats.Started)
	assert.Equal(t, 5*time.Minute, stats.Uptime)
	assert.Equal(t, int64(100), stats.ReceiverStats.TracesReceived)
	assert.Equal(t, int64(200), stats.ReceiverStats.MetricsReceived)
	assert.Equal(t, int64(300), stats.ReceiverStats.LogsReceived)
	assert.Equal(t, int64(90), stats.PipelineStats.TracesProcessed)
	assert.Equal(t, int64(190), stats.PipelineStats.MetricsProcessed)
	assert.Equal(t, int64(290), stats.PipelineStats.LogsProcessed)
}

func TestReceiverStatsStruct(t *testing.T) {
	stats := collector.ReceiverStats{
		TracesReceived:  1000,
		MetricsReceived: 2000,
		LogsReceived:    3000,
	}

	assert.Equal(t, int64(1000), stats.TracesReceived)
	assert.Equal(t, int64(2000), stats.MetricsReceived)
	assert.Equal(t, int64(3000), stats.LogsReceived)
}

func TestPipelineStatsStruct(t *testing.T) {
	stats := collector.PipelineStats{
		TracesProcessed:  900,
		MetricsProcessed: 1900,
		LogsProcessed:    2900,
	}

	assert.Equal(t, int64(900), stats.TracesProcessed)
	assert.Equal(t, int64(1900), stats.MetricsProcessed)
	assert.Equal(t, int64(2900), stats.LogsProcessed)
}

func TestStatsWhileRunning(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := createTestConfig(t)

	c, err := collector.New(cfg, logger)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = c.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	stats := c.Stats()

	assert.True(t, stats.Running)
	assert.Equal(t, c.ID(), stats.ID)
	assert.Greater(t, stats.Uptime, time.Duration(0))
}

func TestConcurrentStatsAccess(t *testing.T) {
	logger := zap.NewNop()
	cfg := createMinimalConfig(t)

	c, err := collector.New(cfg, logger)
	require.NoError(t, err)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = c.Stats()
				_ = c.IsRunning()
				_ = c.Uptime()
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// Benchmark tests
func BenchmarkNew(b *testing.B) {
	logger := zap.NewNop()
	cfg := config.DefaultConfig()
	cfg.Receivers.OTLP.Enabled = false
	cfg.Extensions.Health.Enabled = false

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = collector.New(cfg, logger)
	}
}

func BenchmarkStats(b *testing.B) {
	logger := zap.NewNop()
	cfg := config.DefaultConfig()
	cfg.Receivers.OTLP.Enabled = false
	cfg.Extensions.Health.Enabled = false

	c, _ := collector.New(cfg, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Stats()
	}
}

func BenchmarkIsRunning(b *testing.B) {
	logger := zap.NewNop()
	cfg := config.DefaultConfig()
	cfg.Receivers.OTLP.Enabled = false
	cfg.Extensions.Health.Enabled = false

	c, _ := collector.New(cfg, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.IsRunning()
	}
}
