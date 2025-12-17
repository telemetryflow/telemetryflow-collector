// Package config_test provides unit tests for the configuration package.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/telemetryflow/telemetryflow-collector/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	t.Run("should return valid default configuration", func(t *testing.T) {
		cfg := config.DefaultConfig()

		require.NotNil(t, cfg)
		assert.True(t, cfg.Receivers.OTLP.Enabled)
		assert.Equal(t, "info", cfg.Logging.Level)
		assert.Equal(t, "json", cfg.Logging.Format)
	})

	t.Run("should have OTLP receiver enabled by default", func(t *testing.T) {
		cfg := config.DefaultConfig()

		assert.True(t, cfg.Receivers.OTLP.Enabled)
		assert.True(t, cfg.Receivers.OTLP.Protocols.GRPC.Enabled)
		assert.True(t, cfg.Receivers.OTLP.Protocols.HTTP.Enabled)
	})

	t.Run("should have correct OTLP gRPC defaults", func(t *testing.T) {
		cfg := config.DefaultConfig()

		grpc := cfg.Receivers.OTLP.Protocols.GRPC
		assert.Equal(t, "0.0.0.0:4317", grpc.Endpoint)
		assert.Equal(t, 4, grpc.MaxRecvMsgSizeMiB)
		assert.Equal(t, uint32(100), grpc.MaxConcurrentStreams)
	})

	t.Run("should have correct OTLP HTTP defaults", func(t *testing.T) {
		cfg := config.DefaultConfig()

		http := cfg.Receivers.OTLP.Protocols.HTTP
		assert.Equal(t, "0.0.0.0:4318", http.Endpoint)
		assert.Equal(t, int64(10*1024*1024), http.MaxRequestBodySize) // 10MB
		assert.True(t, http.IncludeMetadata)
	})

	t.Run("should have Prometheus receiver disabled by default", func(t *testing.T) {
		cfg := config.DefaultConfig()

		assert.False(t, cfg.Receivers.Prometheus.Enabled)
	})

	t.Run("should have batch processor enabled", func(t *testing.T) {
		cfg := config.DefaultConfig()

		assert.True(t, cfg.Processors.Batch.Enabled)
		assert.Equal(t, uint32(8192), cfg.Processors.Batch.SendBatchSize)
		assert.Equal(t, 200*time.Millisecond, cfg.Processors.Batch.Timeout)
	})

	t.Run("should have memory limiter enabled", func(t *testing.T) {
		cfg := config.DefaultConfig()

		assert.True(t, cfg.Processors.Memory.Enabled)
		assert.Equal(t, uint32(80), cfg.Processors.Memory.LimitPercentage)
		assert.Equal(t, uint32(25), cfg.Processors.Memory.SpikeLimitPercentage)
	})

	t.Run("should have health check enabled", func(t *testing.T) {
		cfg := config.DefaultConfig()

		assert.True(t, cfg.Extensions.Health.Enabled)
		assert.Equal(t, "0.0.0.0:13133", cfg.Extensions.Health.Endpoint)
		assert.Equal(t, "/", cfg.Extensions.Health.Path)
	})
}

func TestConfigValidation(t *testing.T) {
	t.Run("should pass validation with valid config", func(t *testing.T) {
		cfg := config.DefaultConfig()

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("should fail validation with no receivers enabled", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Receivers.OTLP.Enabled = false
		cfg.Receivers.Prometheus.Enabled = false

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Equal(t, config.ErrNoReceiversEnabled, err)
	})

	t.Run("should fail validation with OTLP enabled but no protocols", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Receivers.OTLP.Enabled = true
		cfg.Receivers.OTLP.Protocols.GRPC.Enabled = false
		cfg.Receivers.OTLP.Protocols.HTTP.Enabled = false

		err := cfg.Validate()
		assert.Error(t, err)
		assert.Equal(t, config.ErrNoOTLPProtocolsEnabled, err)
	})

	t.Run("should pass validation with only gRPC enabled", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Receivers.OTLP.Protocols.HTTP.Enabled = false

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("should pass validation with only HTTP enabled", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Receivers.OTLP.Protocols.GRPC.Enabled = false

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("should pass validation with Prometheus receiver only", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Receivers.OTLP.Enabled = false
		cfg.Receivers.Prometheus.Enabled = true

		err := cfg.Validate()
		assert.NoError(t, err)
	})
}

func TestCollectorConfig(t *testing.T) {
	t.Run("should allow empty collector ID", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Collector.ID = ""

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("should allow empty hostname", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Collector.Hostname = ""

		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("should preserve custom tags", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Collector.Tags = map[string]string{
			"environment": "production",
			"datacenter":  "us-east-1",
		}

		assert.Equal(t, "production", cfg.Collector.Tags["environment"])
		assert.Equal(t, "us-east-1", cfg.Collector.Tags["datacenter"])
	})
}

func TestTLSConfig(t *testing.T) {
	t.Run("should have TLS disabled by default for gRPC", func(t *testing.T) {
		cfg := config.DefaultConfig()

		assert.False(t, cfg.Receivers.OTLP.Protocols.GRPC.TLS.Enabled)
	})

	t.Run("should have TLS disabled by default for HTTP", func(t *testing.T) {
		cfg := config.DefaultConfig()

		assert.False(t, cfg.Receivers.OTLP.Protocols.HTTP.TLS.Enabled)
	})

	t.Run("should allow custom TLS settings", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Receivers.OTLP.Protocols.GRPC.TLS.Enabled = true
		cfg.Receivers.OTLP.Protocols.GRPC.TLS.CertFile = "/etc/ssl/cert.pem"
		cfg.Receivers.OTLP.Protocols.GRPC.TLS.KeyFile = "/etc/ssl/key.pem"
		cfg.Receivers.OTLP.Protocols.GRPC.TLS.MinVersion = "1.3"

		assert.True(t, cfg.Receivers.OTLP.Protocols.GRPC.TLS.Enabled)
		assert.Equal(t, "/etc/ssl/cert.pem", cfg.Receivers.OTLP.Protocols.GRPC.TLS.CertFile)
		assert.Equal(t, "1.3", cfg.Receivers.OTLP.Protocols.GRPC.TLS.MinVersion)
	})
}

func TestLoggingConfig(t *testing.T) {
	validLevels := []string{"debug", "info", "warn", "error"}

	for _, level := range validLevels {
		t.Run("should accept log level "+level, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Logging.Level = level

			err := cfg.Validate()
			assert.NoError(t, err)
		})
	}

	validFormats := []string{"json", "text"}

	for _, format := range validFormats {
		t.Run("should accept log format "+format, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Logging.Format = format

			err := cfg.Validate()
			assert.NoError(t, err)
		})
	}

	t.Run("should have development mode disabled by default", func(t *testing.T) {
		cfg := config.DefaultConfig()

		assert.False(t, cfg.Logging.Development)
	})
}

func TestExporterConfig(t *testing.T) {
	t.Run("should have logging exporter enabled by default", func(t *testing.T) {
		cfg := config.DefaultConfig()

		assert.True(t, cfg.Exporters.Logging.Enabled)
	})

	t.Run("should have OTLP exporter disabled by default", func(t *testing.T) {
		cfg := config.DefaultConfig()

		assert.False(t, cfg.Exporters.OTLP.Enabled)
	})

	t.Run("should have Prometheus exporter disabled by default", func(t *testing.T) {
		cfg := config.DefaultConfig()

		assert.False(t, cfg.Exporters.Prometheus.Enabled)
	})

	t.Run("should have File exporter disabled by default", func(t *testing.T) {
		cfg := config.DefaultConfig()

		assert.False(t, cfg.Exporters.File.Enabled)
	})
}

func TestExtensionsConfig(t *testing.T) {
	t.Run("should have zPages disabled by default", func(t *testing.T) {
		cfg := config.DefaultConfig()

		assert.False(t, cfg.Extensions.ZPages.Enabled)
	})

	t.Run("should have pprof disabled by default", func(t *testing.T) {
		cfg := config.DefaultConfig()

		assert.False(t, cfg.Extensions.PPROF.Enabled)
	})
}

func TestCORSConfig(t *testing.T) {
	t.Run("should have permissive CORS by default", func(t *testing.T) {
		cfg := config.DefaultConfig()

		cors := cfg.Receivers.OTLP.Protocols.HTTP.CORS
		assert.Contains(t, cors.AllowedOrigins, "*")
		assert.Contains(t, cors.AllowedHeaders, "*")
		assert.Equal(t, 7200, cors.MaxAge)
	})
}
