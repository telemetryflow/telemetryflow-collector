// Package config provides configuration management for the TelemetryFlow Collector.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package config

import (
	"testing"
	"time"
)

// TestDefaultConfig tests that DefaultConfig returns a properly initialized config
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}
}

// TestDefaultConfigTelemetryFlow tests TelemetryFlow defaults
func TestDefaultConfigTelemetryFlow(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.TelemetryFlow.Enabled {
		t.Error("Expected TelemetryFlow.Enabled to be false by default")
	}

	if cfg.TelemetryFlow.Endpoint != "localhost:4317" {
		t.Errorf("Expected TelemetryFlow.Endpoint 'localhost:4317', got '%s'", cfg.TelemetryFlow.Endpoint)
	}

	if !cfg.TelemetryFlow.TLS.Enabled {
		t.Error("Expected TelemetryFlow.TLS.Enabled to be true by default")
	}

	if cfg.TelemetryFlow.TLS.InsecureSkipVerify {
		t.Error("Expected TelemetryFlow.TLS.InsecureSkipVerify to be false by default")
	}
}

// TestDefaultConfigCollector tests Collector defaults
func TestDefaultConfigCollector(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Collector.ID != "" {
		t.Errorf("Expected Collector.ID to be empty, got '%s'", cfg.Collector.ID)
	}

	if cfg.Collector.Name != "TelemetryFlow Collector" {
		t.Errorf("Expected Collector.Name 'TelemetryFlow Collector', got '%s'", cfg.Collector.Name)
	}

	if cfg.Collector.Tags == nil {
		t.Error("Expected Collector.Tags to be initialized")
	}

	if cfg.Collector.Tags["environment"] != "production" {
		t.Errorf("Expected Collector.Tags['environment'] 'production', got '%s'", cfg.Collector.Tags["environment"])
	}
}

// TestDefaultConfigOTLPReceiver tests OTLP receiver defaults
func TestDefaultConfigOTLPReceiver(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Receivers.OTLP.Enabled {
		t.Error("Expected OTLP receiver to be enabled by default")
	}

	// gRPC settings
	if !cfg.Receivers.OTLP.Protocols.GRPC.Enabled {
		t.Error("Expected OTLP gRPC to be enabled by default")
	}

	if cfg.Receivers.OTLP.Protocols.GRPC.Endpoint != "0.0.0.0:4317" {
		t.Errorf("Expected OTLP gRPC endpoint '0.0.0.0:4317', got '%s'", cfg.Receivers.OTLP.Protocols.GRPC.Endpoint)
	}

	if cfg.Receivers.OTLP.Protocols.GRPC.MaxRecvMsgSizeMiB != 4 {
		t.Errorf("Expected MaxRecvMsgSizeMiB 4, got %d", cfg.Receivers.OTLP.Protocols.GRPC.MaxRecvMsgSizeMiB)
	}

	if cfg.Receivers.OTLP.Protocols.GRPC.MaxConcurrentStreams != 100 {
		t.Errorf("Expected MaxConcurrentStreams 100, got %d", cfg.Receivers.OTLP.Protocols.GRPC.MaxConcurrentStreams)
	}

	// HTTP settings
	if !cfg.Receivers.OTLP.Protocols.HTTP.Enabled {
		t.Error("Expected OTLP HTTP to be enabled by default")
	}

	if cfg.Receivers.OTLP.Protocols.HTTP.Endpoint != "0.0.0.0:4318" {
		t.Errorf("Expected OTLP HTTP endpoint '0.0.0.0:4318', got '%s'", cfg.Receivers.OTLP.Protocols.HTTP.Endpoint)
	}

	if cfg.Receivers.OTLP.Protocols.HTTP.MaxRequestBodySize != 10*1024*1024 {
		t.Errorf("Expected MaxRequestBodySize 10MB, got %d", cfg.Receivers.OTLP.Protocols.HTTP.MaxRequestBodySize)
	}
}

// TestDefaultConfigPrometheusReceiver tests Prometheus receiver defaults
func TestDefaultConfigPrometheusReceiver(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Receivers.Prometheus.Enabled {
		t.Error("Expected Prometheus receiver to be disabled by default")
	}
}

// TestDefaultConfigProcessors tests processor defaults
func TestDefaultConfigProcessors(t *testing.T) {
	cfg := DefaultConfig()

	// Batch processor
	if !cfg.Processors.Batch.Enabled {
		t.Error("Expected Batch processor to be enabled by default")
	}

	if cfg.Processors.Batch.SendBatchSize != 8192 {
		t.Errorf("Expected SendBatchSize 8192, got %d", cfg.Processors.Batch.SendBatchSize)
	}

	if cfg.Processors.Batch.Timeout != 200*time.Millisecond {
		t.Errorf("Expected Timeout 200ms, got %v", cfg.Processors.Batch.Timeout)
	}

	// Memory limiter
	if !cfg.Processors.Memory.Enabled {
		t.Error("Expected Memory limiter to be enabled by default")
	}

	if cfg.Processors.Memory.CheckInterval != 1*time.Second {
		t.Errorf("Expected CheckInterval 1s, got %v", cfg.Processors.Memory.CheckInterval)
	}

	if cfg.Processors.Memory.LimitPercentage != 80 {
		t.Errorf("Expected LimitPercentage 80, got %d", cfg.Processors.Memory.LimitPercentage)
	}
}

// TestDefaultConfigExporters tests exporter defaults
func TestDefaultConfigExporters(t *testing.T) {
	cfg := DefaultConfig()

	// OTLP exporter
	if cfg.Exporters.OTLP.Enabled {
		t.Error("Expected OTLP exporter to be disabled by default")
	}

	if cfg.Exporters.OTLP.Endpoint != "localhost:4317" {
		t.Errorf("Expected OTLP endpoint 'localhost:4317', got '%s'", cfg.Exporters.OTLP.Endpoint)
	}

	if cfg.Exporters.OTLP.Compression != "gzip" {
		t.Errorf("Expected compression 'gzip', got '%s'", cfg.Exporters.OTLP.Compression)
	}

	// Debug exporter
	if cfg.Exporters.Debug.Verbosity != "detailed" {
		t.Errorf("Expected Debug.Verbosity 'detailed', got '%s'", cfg.Exporters.Debug.Verbosity)
	}

	// Logging exporter (legacy)
	if !cfg.Exporters.Logging.Enabled {
		t.Error("Expected Logging exporter to be enabled by default")
	}
}

// TestDefaultConfigExtensions tests extension defaults
func TestDefaultConfigExtensions(t *testing.T) {
	cfg := DefaultConfig()

	// Health check
	if !cfg.Extensions.Health.Enabled {
		t.Error("Expected Health check to be enabled by default")
	}

	if cfg.Extensions.Health.Endpoint != "0.0.0.0:13133" {
		t.Errorf("Expected Health endpoint '0.0.0.0:13133', got '%s'", cfg.Extensions.Health.Endpoint)
	}

	if cfg.Extensions.Health.Path != "/" {
		t.Errorf("Expected Health path '/', got '%s'", cfg.Extensions.Health.Path)
	}

	// zPages
	if cfg.Extensions.ZPages.Enabled {
		t.Error("Expected zPages to be disabled by default")
	}

	// pprof
	if cfg.Extensions.PPROF.Enabled {
		t.Error("Expected pprof to be disabled by default")
	}
}

// TestDefaultConfigLogging tests logging defaults
func TestDefaultConfigLogging(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Logging.Level != "info" {
		t.Errorf("Expected Logging.Level 'info', got '%s'", cfg.Logging.Level)
	}

	if cfg.Logging.Format != "json" {
		t.Errorf("Expected Logging.Format 'json', got '%s'", cfg.Logging.Format)
	}

	if cfg.Logging.MaxSizeMB != 100 {
		t.Errorf("Expected MaxSizeMB 100, got %d", cfg.Logging.MaxSizeMB)
	}
}

// TestDefaultConfigService tests service configuration defaults
func TestDefaultConfigService(t *testing.T) {
	cfg := DefaultConfig()

	if len(cfg.Service.Extensions) != 3 {
		t.Errorf("Expected 3 extensions, got %d", len(cfg.Service.Extensions))
	}

	// Check traces pipeline
	if len(cfg.Service.Pipelines.Traces.Receivers) == 0 {
		t.Error("Expected traces pipeline to have receivers")
	}

	if len(cfg.Service.Pipelines.Traces.Processors) == 0 {
		t.Error("Expected traces pipeline to have processors")
	}

	if len(cfg.Service.Pipelines.Traces.Exporters) == 0 {
		t.Error("Expected traces pipeline to have exporters")
	}
}

// TestValidate tests config validation
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Config)
		wantErr error
	}{
		{
			name:    "valid default config",
			setup:   func(c *Config) {},
			wantErr: nil,
		},
		{
			name: "no receivers enabled",
			setup: func(c *Config) {
				c.Receivers.OTLP.Enabled = false
				c.Receivers.Prometheus.Enabled = false
			},
			wantErr: ErrNoReceiversEnabled,
		},
		{
			name: "OTLP enabled but no protocols",
			setup: func(c *Config) {
				c.Receivers.OTLP.Enabled = true
				c.Receivers.OTLP.Protocols.GRPC.Enabled = false
				c.Receivers.OTLP.Protocols.HTTP.Enabled = false
			},
			wantErr: ErrNoOTLPProtocolsEnabled,
		},
		{
			name: "OTLP with only gRPC",
			setup: func(c *Config) {
				c.Receivers.OTLP.Enabled = true
				c.Receivers.OTLP.Protocols.GRPC.Enabled = true
				c.Receivers.OTLP.Protocols.HTTP.Enabled = false
			},
			wantErr: nil,
		},
		{
			name: "OTLP with only HTTP",
			setup: func(c *Config) {
				c.Receivers.OTLP.Enabled = true
				c.Receivers.OTLP.Protocols.GRPC.Enabled = false
				c.Receivers.OTLP.Protocols.HTTP.Enabled = true
			},
			wantErr: nil,
		},
		{
			name: "Prometheus receiver only",
			setup: func(c *Config) {
				c.Receivers.OTLP.Enabled = false
				c.Receivers.Prometheus.Enabled = true
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.setup(cfg)

			err := cfg.Validate()

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err != tt.wantErr {
					t.Errorf("Expected error %v, got: %v", tt.wantErr, err)
				}
			}
		})
	}
}

// TestConfigError tests the configError type
func TestConfigError(t *testing.T) {
	err := ErrNoReceiversEnabled

	if err.Error() != "at least one receiver must be enabled" {
		t.Errorf("Expected error message, got '%s'", err.Error())
	}

	err2 := ErrNoOTLPProtocolsEnabled

	if err2.Error() != "OTLP receiver is enabled but no protocols are configured" {
		t.Errorf("Expected error message, got '%s'", err2.Error())
	}
}

// TestTLSConfig tests TLS configuration struct
func TestTLSConfig(t *testing.T) {
	cfg := TLSConfig{
		Enabled:            true,
		InsecureSkipVerify: false,
		CertFile:           "/path/to/cert.pem",
		KeyFile:            "/path/to/key.pem",
		CAFile:             "/path/to/ca.pem",
		ClientAuthType:     "require",
		MinVersion:         "1.3",
	}

	if !cfg.Enabled {
		t.Error("Expected Enabled to be true")
	}

	if cfg.InsecureSkipVerify {
		t.Error("Expected InsecureSkipVerify to be false")
	}

	if cfg.CertFile != "/path/to/cert.pem" {
		t.Errorf("Expected CertFile '/path/to/cert.pem', got '%s'", cfg.CertFile)
	}

	if cfg.MinVersion != "1.3" {
		t.Errorf("Expected MinVersion '1.3', got '%s'", cfg.MinVersion)
	}
}

// TestCORSConfig tests CORS configuration struct
func TestCORSConfig(t *testing.T) {
	cfg := DefaultConfig()

	cors := cfg.Receivers.OTLP.Protocols.HTTP.CORS

	if len(cors.AllowedOrigins) == 0 {
		t.Error("Expected AllowedOrigins to be set")
	}

	if cors.AllowedOrigins[0] != "*" {
		t.Errorf("Expected AllowedOrigins[0] '*', got '%s'", cors.AllowedOrigins[0])
	}

	if cors.MaxAge != 7200 {
		t.Errorf("Expected MaxAge 7200, got %d", cors.MaxAge)
	}
}

// TestKeepaliveConfig tests keepalive configuration
func TestKeepaliveConfig(t *testing.T) {
	cfg := DefaultConfig()

	keepalive := cfg.Receivers.OTLP.Protocols.GRPC.Keepalive

	if keepalive.ServerParameters.MaxConnectionIdle != 15*time.Second {
		t.Errorf("Expected MaxConnectionIdle 15s, got %v", keepalive.ServerParameters.MaxConnectionIdle)
	}

	if keepalive.ServerParameters.MaxConnectionAge != 30*time.Second {
		t.Errorf("Expected MaxConnectionAge 30s, got %v", keepalive.ServerParameters.MaxConnectionAge)
	}
}

// TestRetryConfig tests retry configuration
func TestRetryConfig(t *testing.T) {
	cfg := DefaultConfig()

	retry := cfg.Exporters.OTLP.RetryOnFailure

	if !retry.Enabled {
		t.Error("Expected retry to be enabled by default")
	}

	if retry.InitialInterval != 5*time.Second {
		t.Errorf("Expected InitialInterval 5s, got %v", retry.InitialInterval)
	}

	if retry.MaxInterval != 30*time.Second {
		t.Errorf("Expected MaxInterval 30s, got %v", retry.MaxInterval)
	}

	if retry.MaxElapsedTime != 300*time.Second {
		t.Errorf("Expected MaxElapsedTime 300s, got %v", retry.MaxElapsedTime)
	}
}

// TestQueueConfig tests sending queue configuration
func TestQueueConfig(t *testing.T) {
	cfg := DefaultConfig()

	queue := cfg.Exporters.OTLP.SendingQueue

	if !queue.Enabled {
		t.Error("Expected queue to be enabled by default")
	}

	if queue.NumConsumers != 10 {
		t.Errorf("Expected NumConsumers 10, got %d", queue.NumConsumers)
	}

	if queue.QueueSize != 1000 {
		t.Errorf("Expected QueueSize 1000, got %d", queue.QueueSize)
	}
}

// TestPipelineConfig tests pipeline configuration struct
func TestPipelineConfig(t *testing.T) {
	cfg := PipelineConfig{
		Receivers:  []string{"otlp"},
		Processors: []string{"batch", "memory_limiter"},
		Exporters:  []string{"debug", "prometheus"},
	}

	if len(cfg.Receivers) != 1 {
		t.Errorf("Expected 1 receiver, got %d", len(cfg.Receivers))
	}

	if len(cfg.Processors) != 2 {
		t.Errorf("Expected 2 processors, got %d", len(cfg.Processors))
	}

	if len(cfg.Exporters) != 2 {
		t.Errorf("Expected 2 exporters, got %d", len(cfg.Exporters))
	}
}

// TestSpanMetricsConnectorConfig tests span metrics connector configuration
func TestSpanMetricsConnectorConfig(t *testing.T) {
	cfg := DefaultConfig()

	spanMetrics := cfg.Connectors.SpanMetrics

	if spanMetrics.Namespace != "traces" {
		t.Errorf("Expected Namespace 'traces', got '%s'", spanMetrics.Namespace)
	}

	if spanMetrics.MetricsFlushInterval != 15*time.Second {
		t.Errorf("Expected MetricsFlushInterval 15s, got %v", spanMetrics.MetricsFlushInterval)
	}

	if !spanMetrics.Exemplars.Enabled {
		t.Error("Expected Exemplars to be enabled by default")
	}
}

// TestServiceGraphConnectorConfig tests service graph connector configuration
func TestServiceGraphConnectorConfig(t *testing.T) {
	cfg := DefaultConfig()

	serviceGraph := cfg.Connectors.ServiceGraph

	if serviceGraph.Store.TTL != 2*time.Second {
		t.Errorf("Expected Store.TTL 2s, got %v", serviceGraph.Store.TTL)
	}

	if serviceGraph.Store.MaxItems != 1000 {
		t.Errorf("Expected Store.MaxItems 1000, got %d", serviceGraph.Store.MaxItems)
	}

	if serviceGraph.CacheLoop != 1*time.Second {
		t.Errorf("Expected CacheLoop 1s, got %v", serviceGraph.CacheLoop)
	}
}

// TestServiceTelemetryConfig tests service telemetry configuration
func TestServiceTelemetryConfig(t *testing.T) {
	cfg := DefaultConfig()

	telemetry := cfg.Service.Telemetry

	if telemetry.Logs.Level != "info" {
		t.Errorf("Expected Logs.Level 'info', got '%s'", telemetry.Logs.Level)
	}

	if telemetry.Logs.Encoding != "json" {
		t.Errorf("Expected Logs.Encoding 'json', got '%s'", telemetry.Logs.Encoding)
	}

	if telemetry.Metrics.Level != "detailed" {
		t.Errorf("Expected Metrics.Level 'detailed', got '%s'", telemetry.Metrics.Level)
	}
}

// TestLogSamplingConfig tests log sampling configuration
func TestLogSamplingConfig(t *testing.T) {
	cfg := DefaultConfig()

	sampling := cfg.Logging.Sampling

	if !sampling.Enabled {
		t.Error("Expected log sampling to be enabled by default")
	}

	if sampling.Initial != 100 {
		t.Errorf("Expected Initial 100, got %d", sampling.Initial)
	}

	if sampling.Thereafter != 100 {
		t.Errorf("Expected Thereafter 100, got %d", sampling.Thereafter)
	}
}

// TestFileExporterConfig tests file exporter configuration
func TestFileExporterConfig(t *testing.T) {
	cfg := FileExporterConfig{
		Enabled:     true,
		Path:        "/var/log/otel/traces.json",
		Format:      "json",
		Compression: "gzip",
		Rotation: FileRotationConfig{
			MaxMegabytes: 100,
			MaxDays:      7,
			MaxBackups:   3,
			LocalTime:    true,
		},
		FlushInterval: 5 * time.Second,
	}

	if !cfg.Enabled {
		t.Error("Expected Enabled to be true")
	}

	if cfg.Path != "/var/log/otel/traces.json" {
		t.Errorf("Expected Path '/var/log/otel/traces.json', got '%s'", cfg.Path)
	}

	if cfg.Rotation.MaxMegabytes != 100 {
		t.Errorf("Expected MaxMegabytes 100, got %d", cfg.Rotation.MaxMegabytes)
	}

	if cfg.FlushInterval != 5*time.Second {
		t.Errorf("Expected FlushInterval 5s, got %v", cfg.FlushInterval)
	}
}

// TestAttributeAction tests attribute action configuration
func TestAttributeAction(t *testing.T) {
	action := AttributeAction{
		Key:           "environment",
		Action:        "upsert",
		Value:         "production",
		FromAttribute: "",
		Pattern:       "",
	}

	if action.Key != "environment" {
		t.Errorf("Expected Key 'environment', got '%s'", action.Key)
	}

	if action.Action != "upsert" {
		t.Errorf("Expected Action 'upsert', got '%s'", action.Action)
	}

	if action.Value != "production" {
		t.Errorf("Expected Value 'production', got '%v'", action.Value)
	}
}

// TestScrapeConfig tests Prometheus scrape configuration
func TestScrapeConfig(t *testing.T) {
	cfg := ScrapeConfig{
		JobName:        "my-app",
		ScrapeInterval: 15 * time.Second,
		ScrapeTimeout:  10 * time.Second,
		MetricsPath:    "/metrics",
		StaticConfigs: []StaticTargetConfig{
			{
				Targets: []string{"localhost:8080", "localhost:8081"},
				Labels:  map[string]string{"env": "dev"},
			},
		},
	}

	if cfg.JobName != "my-app" {
		t.Errorf("Expected JobName 'my-app', got '%s'", cfg.JobName)
	}

	if cfg.ScrapeInterval != 15*time.Second {
		t.Errorf("Expected ScrapeInterval 15s, got %v", cfg.ScrapeInterval)
	}

	if len(cfg.StaticConfigs) != 1 {
		t.Errorf("Expected 1 static config, got %d", len(cfg.StaticConfigs))
	}

	if len(cfg.StaticConfigs[0].Targets) != 2 {
		t.Errorf("Expected 2 targets, got %d", len(cfg.StaticConfigs[0].Targets))
	}
}

// Benchmark tests
func BenchmarkDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DefaultConfig()
	}
}

func BenchmarkValidate(b *testing.B) {
	cfg := DefaultConfig()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = cfg.Validate()
	}
}
