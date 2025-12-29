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

func TestDefaultConfigTelemetryFlow(t *testing.T) {
	cfg := config.DefaultConfig()

	assert.False(t, cfg.TelemetryFlow.Enabled)
	assert.Equal(t, "localhost:4317", cfg.TelemetryFlow.Endpoint)
	assert.True(t, cfg.TelemetryFlow.TLS.Enabled)
	assert.False(t, cfg.TelemetryFlow.TLS.InsecureSkipVerify)
}

func TestDefaultConfigCollector(t *testing.T) {
	cfg := config.DefaultConfig()

	assert.Empty(t, cfg.Collector.ID)
	assert.Equal(t, "TelemetryFlow Collector", cfg.Collector.Name)
	require.NotNil(t, cfg.Collector.Tags)
	assert.Equal(t, "production", cfg.Collector.Tags["environment"])
}

func TestDefaultConfigOTLPReceiver(t *testing.T) {
	cfg := config.DefaultConfig()

	assert.True(t, cfg.Receivers.OTLP.Enabled)

	// gRPC settings
	assert.True(t, cfg.Receivers.OTLP.Protocols.GRPC.Enabled)
	assert.Equal(t, "0.0.0.0:4317", cfg.Receivers.OTLP.Protocols.GRPC.Endpoint)
	assert.Equal(t, 4, cfg.Receivers.OTLP.Protocols.GRPC.MaxRecvMsgSizeMiB)
	assert.Equal(t, uint32(100), cfg.Receivers.OTLP.Protocols.GRPC.MaxConcurrentStreams)

	// HTTP settings
	assert.True(t, cfg.Receivers.OTLP.Protocols.HTTP.Enabled)
	assert.Equal(t, "0.0.0.0:4318", cfg.Receivers.OTLP.Protocols.HTTP.Endpoint)
	assert.Equal(t, int64(10*1024*1024), cfg.Receivers.OTLP.Protocols.HTTP.MaxRequestBodySize)
}

func TestDefaultConfigPrometheusReceiver(t *testing.T) {
	cfg := config.DefaultConfig()

	assert.False(t, cfg.Receivers.Prometheus.Enabled)
}

func TestDefaultConfigProcessors(t *testing.T) {
	cfg := config.DefaultConfig()

	// Batch processor
	assert.True(t, cfg.Processors.Batch.Enabled)
	assert.Equal(t, uint32(8192), cfg.Processors.Batch.SendBatchSize)
	assert.Equal(t, 200*time.Millisecond, cfg.Processors.Batch.Timeout)

	// Memory limiter
	assert.True(t, cfg.Processors.Memory.Enabled)
	assert.Equal(t, 1*time.Second, cfg.Processors.Memory.CheckInterval)
	assert.Equal(t, uint32(80), cfg.Processors.Memory.LimitPercentage)
}

func TestDefaultConfigExporters(t *testing.T) {
	cfg := config.DefaultConfig()

	// OTLP exporter
	assert.False(t, cfg.Exporters.OTLP.Enabled)
	assert.Equal(t, "localhost:4317", cfg.Exporters.OTLP.Endpoint)
	assert.Equal(t, "gzip", cfg.Exporters.OTLP.Compression)

	// Debug exporter
	assert.Equal(t, "detailed", cfg.Exporters.Debug.Verbosity)

	// Logging exporter (legacy)
	assert.True(t, cfg.Exporters.Logging.Enabled)
}

func TestDefaultConfigExtensions(t *testing.T) {
	cfg := config.DefaultConfig()

	// Health check
	assert.True(t, cfg.Extensions.Health.Enabled)
	assert.Equal(t, "0.0.0.0:13133", cfg.Extensions.Health.Endpoint)
	assert.Equal(t, "/", cfg.Extensions.Health.Path)

	// zPages
	assert.False(t, cfg.Extensions.ZPages.Enabled)

	// pprof
	assert.False(t, cfg.Extensions.PPROF.Enabled)
}

func TestDefaultConfigLogging(t *testing.T) {
	cfg := config.DefaultConfig()

	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, 100, cfg.Logging.MaxSizeMB)
}

func TestDefaultConfigService(t *testing.T) {
	cfg := config.DefaultConfig()

	assert.Len(t, cfg.Service.Extensions, 3)

	// Check traces pipeline
	assert.NotEmpty(t, cfg.Service.Pipelines.Traces.Receivers)
	assert.NotEmpty(t, cfg.Service.Pipelines.Traces.Processors)
	assert.NotEmpty(t, cfg.Service.Pipelines.Traces.Exporters)
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

	t.Run("TLS struct values", func(t *testing.T) {
		cfg := config.TLSConfig{
			Enabled:            true,
			InsecureSkipVerify: false,
			CertFile:           "/path/to/cert.pem",
			KeyFile:            "/path/to/key.pem",
			CAFile:             "/path/to/ca.pem",
			ClientAuthType:     "require",
			MinVersion:         "1.3",
		}

		assert.True(t, cfg.Enabled)
		assert.False(t, cfg.InsecureSkipVerify)
		assert.Equal(t, "/path/to/cert.pem", cfg.CertFile)
		assert.Equal(t, "1.3", cfg.MinVersion)
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

func TestConfigError(t *testing.T) {
	err := config.ErrNoReceiversEnabled
	assert.Equal(t, "at least one receiver must be enabled", err.Error())

	err2 := config.ErrNoOTLPProtocolsEnabled
	assert.Equal(t, "OTLP receiver is enabled but no protocols are configured", err2.Error())
}

func TestKeepaliveConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	keepalive := cfg.Receivers.OTLP.Protocols.GRPC.Keepalive

	assert.Equal(t, 15*time.Second, keepalive.ServerParameters.MaxConnectionIdle)
	assert.Equal(t, 30*time.Second, keepalive.ServerParameters.MaxConnectionAge)
}

func TestRetryConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	retry := cfg.Exporters.OTLP.RetryOnFailure

	assert.True(t, retry.Enabled)
	assert.Equal(t, 5*time.Second, retry.InitialInterval)
	assert.Equal(t, 30*time.Second, retry.MaxInterval)
	assert.Equal(t, 300*time.Second, retry.MaxElapsedTime)
}

func TestQueueConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	queue := cfg.Exporters.OTLP.SendingQueue

	assert.True(t, queue.Enabled)
	assert.Equal(t, 10, queue.NumConsumers)
	assert.Equal(t, 1000, queue.QueueSize)
}

func TestPipelineConfig(t *testing.T) {
	cfg := config.PipelineConfig{
		Receivers:  []string{"otlp"},
		Processors: []string{"batch", "memory_limiter"},
		Exporters:  []string{"debug", "prometheus"},
	}

	assert.Len(t, cfg.Receivers, 1)
	assert.Len(t, cfg.Processors, 2)
	assert.Len(t, cfg.Exporters, 2)
}

func TestSpanMetricsConnectorConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	spanMetrics := cfg.Connectors.SpanMetrics

	assert.Equal(t, "traces", spanMetrics.Namespace)
	assert.Equal(t, 15*time.Second, spanMetrics.MetricsFlushInterval)
	assert.True(t, spanMetrics.Exemplars.Enabled)
}

func TestServiceGraphConnectorConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	serviceGraph := cfg.Connectors.ServiceGraph

	assert.Equal(t, 2*time.Second, serviceGraph.Store.TTL)
	assert.Equal(t, 1000, serviceGraph.Store.MaxItems)
	assert.Equal(t, 1*time.Second, serviceGraph.CacheLoop)
}

func TestServiceTelemetryConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	telemetry := cfg.Service.Telemetry

	assert.Equal(t, "info", telemetry.Logs.Level)
	assert.Equal(t, "json", telemetry.Logs.Encoding)
	assert.Equal(t, "detailed", telemetry.Metrics.Level)
}

func TestLogSamplingConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	sampling := cfg.Logging.Sampling

	assert.True(t, sampling.Enabled)
	assert.Equal(t, 100, sampling.Initial)
	assert.Equal(t, 100, sampling.Thereafter)
}

func TestFileExporterConfig(t *testing.T) {
	cfg := config.FileExporterConfig{
		Enabled:     true,
		Path:        "/var/log/otel/traces.json",
		Format:      "json",
		Compression: "gzip",
		Rotation: config.FileRotationConfig{
			MaxMegabytes: 100,
			MaxDays:      7,
			MaxBackups:   3,
			LocalTime:    true,
		},
		FlushInterval: 5 * time.Second,
	}

	assert.True(t, cfg.Enabled)
	assert.Equal(t, "/var/log/otel/traces.json", cfg.Path)
	assert.Equal(t, 100, cfg.Rotation.MaxMegabytes)
	assert.Equal(t, 5*time.Second, cfg.FlushInterval)
}

func TestAttributeAction(t *testing.T) {
	action := config.AttributeAction{
		Key:           "environment",
		Action:        "upsert",
		Value:         "production",
		FromAttribute: "",
		Pattern:       "",
	}

	assert.Equal(t, "environment", action.Key)
	assert.Equal(t, "upsert", action.Action)
	assert.Equal(t, "production", action.Value)
}

func TestScrapeConfig(t *testing.T) {
	cfg := config.ScrapeConfig{
		JobName:        "my-app",
		ScrapeInterval: 15 * time.Second,
		ScrapeTimeout:  10 * time.Second,
		MetricsPath:    "/metrics",
		StaticConfigs: []config.StaticTargetConfig{
			{
				Targets: []string{"localhost:8080", "localhost:8081"},
				Labels:  map[string]string{"env": "dev"},
			},
		},
	}

	assert.Equal(t, "my-app", cfg.JobName)
	assert.Equal(t, 15*time.Second, cfg.ScrapeInterval)
	assert.Len(t, cfg.StaticConfigs, 1)
	assert.Len(t, cfg.StaticConfigs[0].Targets, 2)
}

// Benchmark tests
func BenchmarkDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = config.DefaultConfig()
	}
}

func BenchmarkValidate(b *testing.B) {
	cfg := config.DefaultConfig()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = cfg.Validate()
	}
}
