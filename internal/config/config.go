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
	"time"
)

// Config represents the complete collector configuration
type Config struct {
	Collector CollectorConfig `mapstructure:"collector"`
	Receivers ReceiversConfig `mapstructure:"receivers"`
	Processors ProcessorsConfig `mapstructure:"processors"`
	Exporters ExportersConfig `mapstructure:"exporters"`
	Pipelines PipelinesConfig `mapstructure:"pipelines"`
	Extensions ExtensionsConfig `mapstructure:"extensions"`
	Logging   LoggingConfig   `mapstructure:"logging"`
}

// CollectorConfig contains collector identification settings
type CollectorConfig struct {
	// ID is the unique collector identifier (auto-generated if empty)
	ID string `mapstructure:"id"`

	// Hostname is the collector hostname (auto-detected if empty)
	Hostname string `mapstructure:"hostname"`

	// Description is a human-readable description
	Description string `mapstructure:"description"`

	// Tags are custom key-value labels for the collector
	Tags map[string]string `mapstructure:"tags"`
}

// ReceiversConfig contains all receiver settings
type ReceiversConfig struct {
	// OTLP contains OTLP receiver settings
	OTLP OTLPReceiverConfig `mapstructure:"otlp"`

	// Prometheus contains Prometheus scrape receiver settings
	Prometheus PrometheusReceiverConfig `mapstructure:"prometheus"`
}

// OTLPReceiverConfig contains OTLP receiver settings
type OTLPReceiverConfig struct {
	// Enabled enables the OTLP receiver
	Enabled bool `mapstructure:"enabled"`

	// Protocols contains protocol-specific settings
	Protocols OTLPProtocolsConfig `mapstructure:"protocols"`
}

// OTLPProtocolsConfig contains OTLP protocol settings
type OTLPProtocolsConfig struct {
	// GRPC contains gRPC settings
	GRPC OTLPGRPCConfig `mapstructure:"grpc"`

	// HTTP contains HTTP settings
	HTTP OTLPHTTPConfig `mapstructure:"http"`
}

// OTLPGRPCConfig contains OTLP gRPC receiver settings
type OTLPGRPCConfig struct {
	// Enabled enables gRPC protocol
	Enabled bool `mapstructure:"enabled"`

	// Endpoint is the gRPC listener address
	Endpoint string `mapstructure:"endpoint"`

	// TLS contains TLS settings
	TLS TLSConfig `mapstructure:"tls"`

	// MaxRecvMsgSizeMiB is the maximum message size in MiB
	MaxRecvMsgSizeMiB int `mapstructure:"max_recv_msg_size_mib"`

	// MaxConcurrentStreams is the maximum concurrent streams
	MaxConcurrentStreams uint32 `mapstructure:"max_concurrent_streams"`

	// ReadBufferSize is the read buffer size
	ReadBufferSize int `mapstructure:"read_buffer_size"`

	// WriteBufferSize is the write buffer size
	WriteBufferSize int `mapstructure:"write_buffer_size"`

	// Keepalive contains keepalive settings
	Keepalive KeepaliveConfig `mapstructure:"keepalive"`
}

// OTLPHTTPConfig contains OTLP HTTP receiver settings
type OTLPHTTPConfig struct {
	// Enabled enables HTTP protocol
	Enabled bool `mapstructure:"enabled"`

	// Endpoint is the HTTP listener address
	Endpoint string `mapstructure:"endpoint"`

	// TLS contains TLS settings
	TLS TLSConfig `mapstructure:"tls"`

	// CORS contains CORS settings
	CORS CORSConfig `mapstructure:"cors"`

	// MaxRequestBodySize is the maximum request body size in bytes
	MaxRequestBodySize int64 `mapstructure:"max_request_body_size"`

	// IncludeMetadata includes client metadata in context
	IncludeMetadata bool `mapstructure:"include_metadata"`
}

// TLSConfig contains TLS settings
type TLSConfig struct {
	// Enabled enables TLS
	Enabled bool `mapstructure:"enabled"`

	// CertFile is the path to the TLS certificate
	CertFile string `mapstructure:"cert_file"`

	// KeyFile is the path to the TLS private key
	KeyFile string `mapstructure:"key_file"`

	// CAFile is the path to the CA certificate (for mTLS)
	CAFile string `mapstructure:"ca_file"`

	// ClientAuthType specifies client auth type (none, request, require, verify)
	ClientAuthType string `mapstructure:"client_auth_type"`

	// MinVersion is the minimum TLS version (1.2, 1.3)
	MinVersion string `mapstructure:"min_version"`
}

// CORSConfig contains CORS settings
type CORSConfig struct {
	// AllowedOrigins is a list of allowed origins
	AllowedOrigins []string `mapstructure:"allowed_origins"`

	// AllowedHeaders is a list of allowed headers
	AllowedHeaders []string `mapstructure:"allowed_headers"`

	// MaxAge is the max age for preflight cache in seconds
	MaxAge int `mapstructure:"max_age"`
}

// KeepaliveConfig contains gRPC keepalive settings
type KeepaliveConfig struct {
	// ServerParameters contains server-side keepalive settings
	ServerParameters KeepaliveServerConfig `mapstructure:"server_parameters"`

	// EnforcementPolicy contains enforcement policy settings
	EnforcementPolicy KeepaliveEnforcementConfig `mapstructure:"enforcement_policy"`
}

// KeepaliveServerConfig contains keepalive server parameters
type KeepaliveServerConfig struct {
	// MaxConnectionIdle is the max time a connection can be idle
	MaxConnectionIdle time.Duration `mapstructure:"max_connection_idle"`

	// MaxConnectionAge is the max age of a connection
	MaxConnectionAge time.Duration `mapstructure:"max_connection_age"`

	// MaxConnectionAgeGrace is the grace period after max age
	MaxConnectionAgeGrace time.Duration `mapstructure:"max_connection_age_grace"`

	// Time is the ping interval
	Time time.Duration `mapstructure:"time"`

	// Timeout is the ping timeout
	Timeout time.Duration `mapstructure:"timeout"`
}

// KeepaliveEnforcementConfig contains keepalive enforcement settings
type KeepaliveEnforcementConfig struct {
	// MinTime is the minimum time between pings
	MinTime time.Duration `mapstructure:"min_time"`

	// PermitWithoutStream allows pings without active streams
	PermitWithoutStream bool `mapstructure:"permit_without_stream"`
}

// PrometheusReceiverConfig contains Prometheus scrape receiver settings
type PrometheusReceiverConfig struct {
	// Enabled enables the Prometheus receiver
	Enabled bool `mapstructure:"enabled"`

	// ScrapeConfigs contains scrape configurations
	ScrapeConfigs []ScrapeConfig `mapstructure:"scrape_configs"`
}

// ScrapeConfig contains a single Prometheus scrape configuration
type ScrapeConfig struct {
	// JobName is the job name for this scrape config
	JobName string `mapstructure:"job_name"`

	// ScrapeInterval is the scrape interval
	ScrapeInterval time.Duration `mapstructure:"scrape_interval"`

	// ScrapeTimeout is the scrape timeout
	ScrapeTimeout time.Duration `mapstructure:"scrape_timeout"`

	// MetricsPath is the path to scrape
	MetricsPath string `mapstructure:"metrics_path"`

	// StaticConfigs contains static target configurations
	StaticConfigs []StaticTargetConfig `mapstructure:"static_configs"`
}

// StaticTargetConfig contains static target configuration
type StaticTargetConfig struct {
	// Targets is a list of target addresses
	Targets []string `mapstructure:"targets"`

	// Labels are additional labels to add
	Labels map[string]string `mapstructure:"labels"`
}

// ProcessorsConfig contains processor settings
type ProcessorsConfig struct {
	// Batch contains batch processor settings
	Batch BatchProcessorConfig `mapstructure:"batch"`

	// Memory contains memory limiter settings
	Memory MemoryLimiterConfig `mapstructure:"memory_limiter"`

	// Attributes contains attribute processor settings
	Attributes AttributesProcessorConfig `mapstructure:"attributes"`
}

// BatchProcessorConfig contains batch processor settings
type BatchProcessorConfig struct {
	// Enabled enables the batch processor
	Enabled bool `mapstructure:"enabled"`

	// SendBatchSize is the batch size
	SendBatchSize uint32 `mapstructure:"send_batch_size"`

	// SendBatchMaxSize is the maximum batch size
	SendBatchMaxSize uint32 `mapstructure:"send_batch_max_size"`

	// Timeout is the batch timeout
	Timeout time.Duration `mapstructure:"timeout"`
}

// MemoryLimiterConfig contains memory limiter settings
type MemoryLimiterConfig struct {
	// Enabled enables the memory limiter
	Enabled bool `mapstructure:"enabled"`

	// CheckInterval is the memory check interval
	CheckInterval time.Duration `mapstructure:"check_interval"`

	// LimitMiB is the memory limit in MiB
	LimitMiB uint32 `mapstructure:"limit_mib"`

	// SpikeLimitMiB is the spike limit in MiB
	SpikeLimitMiB uint32 `mapstructure:"spike_limit_mib"`

	// LimitPercentage is the memory limit as percentage of total
	LimitPercentage uint32 `mapstructure:"limit_percentage"`

	// SpikeLimitPercentage is the spike limit as percentage
	SpikeLimitPercentage uint32 `mapstructure:"spike_limit_percentage"`
}

// AttributesProcessorConfig contains attribute processor settings
type AttributesProcessorConfig struct {
	// Enabled enables the attributes processor
	Enabled bool `mapstructure:"enabled"`

	// Actions contains attribute actions
	Actions []AttributeAction `mapstructure:"actions"`
}

// AttributeAction represents an attribute action
type AttributeAction struct {
	// Key is the attribute key
	Key string `mapstructure:"key"`

	// Action is the action type (insert, update, upsert, delete, hash)
	Action string `mapstructure:"action"`

	// Value is the attribute value (for insert/update/upsert)
	Value interface{} `mapstructure:"value"`

	// FromAttribute is the source attribute (for copy)
	FromAttribute string `mapstructure:"from_attribute"`

	// Pattern is the regex pattern (for extract)
	Pattern string `mapstructure:"pattern"`
}

// ExportersConfig contains exporter settings
type ExportersConfig struct {
	// OTLP contains OTLP exporter settings
	OTLP OTLPExporterConfig `mapstructure:"otlp"`

	// Prometheus contains Prometheus exporter settings
	Prometheus PrometheusExporterConfig `mapstructure:"prometheus"`

	// Logging contains logging exporter settings (for debug)
	Logging LoggingExporterConfig `mapstructure:"logging"`

	// File contains file exporter settings
	File FileExporterConfig `mapstructure:"file"`
}

// OTLPExporterConfig contains OTLP exporter settings
type OTLPExporterConfig struct {
	// Enabled enables the OTLP exporter
	Enabled bool `mapstructure:"enabled"`

	// Endpoint is the destination endpoint
	Endpoint string `mapstructure:"endpoint"`

	// TLS contains TLS settings
	TLS TLSConfig `mapstructure:"tls"`

	// Headers contains headers to send
	Headers map[string]string `mapstructure:"headers"`

	// Compression is the compression type (gzip, none)
	Compression string `mapstructure:"compression"`

	// Timeout is the export timeout
	Timeout time.Duration `mapstructure:"timeout"`

	// RetryOnFailure enables retry on failure
	RetryOnFailure RetryConfig `mapstructure:"retry_on_failure"`

	// SendingQueue contains sending queue settings
	SendingQueue QueueConfig `mapstructure:"sending_queue"`
}

// RetryConfig contains retry settings
type RetryConfig struct {
	// Enabled enables retry
	Enabled bool `mapstructure:"enabled"`

	// InitialInterval is the initial retry interval
	InitialInterval time.Duration `mapstructure:"initial_interval"`

	// MaxInterval is the maximum retry interval
	MaxInterval time.Duration `mapstructure:"max_interval"`

	// MaxElapsedTime is the maximum elapsed time for retries
	MaxElapsedTime time.Duration `mapstructure:"max_elapsed_time"`
}

// QueueConfig contains sending queue settings
type QueueConfig struct {
	// Enabled enables the queue
	Enabled bool `mapstructure:"enabled"`

	// NumConsumers is the number of consumers
	NumConsumers int `mapstructure:"num_consumers"`

	// QueueSize is the queue size
	QueueSize int `mapstructure:"queue_size"`

	// StorageDir is the storage directory for persistent queue
	StorageDir string `mapstructure:"storage_dir"`
}

// PrometheusExporterConfig contains Prometheus exporter settings
type PrometheusExporterConfig struct {
	// Enabled enables the Prometheus exporter
	Enabled bool `mapstructure:"enabled"`

	// Endpoint is the Prometheus metrics endpoint
	Endpoint string `mapstructure:"endpoint"`

	// Namespace is the metric namespace
	Namespace string `mapstructure:"namespace"`

	// ConstLabels are constant labels to add
	ConstLabels map[string]string `mapstructure:"const_labels"`

	// SendTimestamps enables sending timestamps
	SendTimestamps bool `mapstructure:"send_timestamps"`

	// MetricExpiration is the metric expiration duration
	MetricExpiration time.Duration `mapstructure:"metric_expiration"`

	// ResourceToTelemetryConversion converts resource attributes to metric labels
	ResourceToTelemetryConversion bool `mapstructure:"resource_to_telemetry_conversion"`
}

// LoggingExporterConfig contains logging exporter settings
type LoggingExporterConfig struct {
	// Enabled enables the logging exporter
	Enabled bool `mapstructure:"enabled"`

	// LogLevel is the log level (debug, info, warn, error)
	LogLevel string `mapstructure:"loglevel"`

	// SamplingInitial is the initial sampling rate
	SamplingInitial int `mapstructure:"sampling_initial"`

	// SamplingThereafter is the subsequent sampling rate
	SamplingThereafter int `mapstructure:"sampling_thereafter"`
}

// FileExporterConfig contains file exporter settings
type FileExporterConfig struct {
	// Enabled enables the file exporter
	Enabled bool `mapstructure:"enabled"`

	// Path is the output file path
	Path string `mapstructure:"path"`

	// Rotation contains rotation settings
	Rotation FileRotationConfig `mapstructure:"rotation"`

	// Format is the output format (json, proto)
	Format string `mapstructure:"format"`

	// Compression is the compression type (none, gzip)
	Compression string `mapstructure:"compression"`

	// FlushInterval is the flush interval
	FlushInterval time.Duration `mapstructure:"flush_interval"`
}

// FileRotationConfig contains file rotation settings
type FileRotationConfig struct {
	// MaxMegabytes is the max file size in MB before rotation
	MaxMegabytes int `mapstructure:"max_megabytes"`

	// MaxDays is the max age in days before deletion
	MaxDays int `mapstructure:"max_days"`

	// MaxBackups is the max number of backup files
	MaxBackups int `mapstructure:"max_backups"`

	// LocalTime uses local time for timestamps
	LocalTime bool `mapstructure:"localtime"`
}

// PipelinesConfig contains pipeline configurations
type PipelinesConfig struct {
	// Metrics contains metrics pipeline configuration
	Metrics PipelineConfig `mapstructure:"metrics"`

	// Logs contains logs pipeline configuration
	Logs PipelineConfig `mapstructure:"logs"`

	// Traces contains traces pipeline configuration
	Traces PipelineConfig `mapstructure:"traces"`
}

// PipelineConfig contains a single pipeline configuration
type PipelineConfig struct {
	// Receivers is the list of receivers
	Receivers []string `mapstructure:"receivers"`

	// Processors is the list of processors (in order)
	Processors []string `mapstructure:"processors"`

	// Exporters is the list of exporters
	Exporters []string `mapstructure:"exporters"`
}

// ExtensionsConfig contains extension settings
type ExtensionsConfig struct {
	// Health contains health check extension settings
	Health HealthCheckConfig `mapstructure:"health_check"`

	// ZPages contains zPages extension settings
	ZPages ZPagesConfig `mapstructure:"zpages"`

	// PPROFConfig contains pprof extension settings
	PPROF PPROFConfig `mapstructure:"pprof"`
}

// HealthCheckConfig contains health check extension settings
type HealthCheckConfig struct {
	// Enabled enables the health check extension
	Enabled bool `mapstructure:"enabled"`

	// Endpoint is the health check endpoint
	Endpoint string `mapstructure:"endpoint"`

	// Path is the health check path
	Path string `mapstructure:"path"`
}

// ZPagesConfig contains zPages extension settings
type ZPagesConfig struct {
	// Enabled enables the zPages extension
	Enabled bool `mapstructure:"enabled"`

	// Endpoint is the zPages endpoint
	Endpoint string `mapstructure:"endpoint"`
}

// PPROFConfig contains pprof extension settings
type PPROFConfig struct {
	// Enabled enables the pprof extension
	Enabled bool `mapstructure:"enabled"`

	// Endpoint is the pprof endpoint
	Endpoint string `mapstructure:"endpoint"`

	// BlockProfileFraction is the block profile fraction
	BlockProfileFraction int `mapstructure:"block_profile_fraction"`

	// MutexProfileFraction is the mutex profile fraction
	MutexProfileFraction int `mapstructure:"mutex_profile_fraction"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	// Level is the log level (debug, info, warn, error)
	Level string `mapstructure:"level"`

	// Format is the log format (json, text)
	Format string `mapstructure:"format"`

	// File is the log file path (empty = stdout)
	File string `mapstructure:"file"`

	// MaxSizeMB is the max log file size before rotation
	MaxSizeMB int `mapstructure:"max_size_mb"`

	// MaxBackups is the number of old log files to keep
	MaxBackups int `mapstructure:"max_backups"`

	// MaxAgeDays is the max age in days for log files
	MaxAgeDays int `mapstructure:"max_age_days"`

	// Development enables development mode logging
	Development bool `mapstructure:"development"`

	// Sampling contains sampling settings for production
	Sampling LogSamplingConfig `mapstructure:"sampling"`
}

// LogSamplingConfig contains log sampling settings
type LogSamplingConfig struct {
	// Enabled enables log sampling
	Enabled bool `mapstructure:"enabled"`

	// Initial is the initial sampling rate
	Initial int `mapstructure:"initial"`

	// Thereafter is the subsequent sampling rate
	Thereafter int `mapstructure:"thereafter"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Collector: CollectorConfig{
			ID:       "",
			Hostname: "",
			Tags:     make(map[string]string),
		},
		Receivers: ReceiversConfig{
			OTLP: OTLPReceiverConfig{
				Enabled: true,
				Protocols: OTLPProtocolsConfig{
					GRPC: OTLPGRPCConfig{
						Enabled:              true,
						Endpoint:             "0.0.0.0:4317",
						MaxRecvMsgSizeMiB:    4,
						MaxConcurrentStreams: 100,
						ReadBufferSize:       512 * 1024,
						WriteBufferSize:      512 * 1024,
						Keepalive: KeepaliveConfig{
							ServerParameters: KeepaliveServerConfig{
								MaxConnectionIdle:     15 * time.Second,
								MaxConnectionAge:      30 * time.Second,
								MaxConnectionAgeGrace: 5 * time.Second,
								Time:                  10 * time.Second,
								Timeout:               5 * time.Second,
							},
						},
					},
					HTTP: OTLPHTTPConfig{
						Enabled:            true,
						Endpoint:           "0.0.0.0:4318",
						MaxRequestBodySize: 10 * 1024 * 1024, // 10MB
						IncludeMetadata:    true,
						CORS: CORSConfig{
							AllowedOrigins: []string{"*"},
							AllowedHeaders: []string{"*"},
							MaxAge:         7200,
						},
					},
				},
			},
			Prometheus: PrometheusReceiverConfig{
				Enabled: false,
			},
		},
		Processors: ProcessorsConfig{
			Batch: BatchProcessorConfig{
				Enabled:          true,
				SendBatchSize:    8192,
				SendBatchMaxSize: 0, // no limit
				Timeout:          200 * time.Millisecond,
			},
			Memory: MemoryLimiterConfig{
				Enabled:              true,
				CheckInterval:        1 * time.Second,
				LimitMiB:             0,
				SpikeLimitMiB:        0,
				LimitPercentage:      80,
				SpikeLimitPercentage: 25,
			},
		},
		Exporters: ExportersConfig{
			Logging: LoggingExporterConfig{
				Enabled:            true,
				LogLevel:           "info",
				SamplingInitial:    5,
				SamplingThereafter: 200,
			},
		},
		Pipelines: PipelinesConfig{
			Metrics: PipelineConfig{
				Receivers:  []string{"otlp"},
				Processors: []string{"memory_limiter", "batch"},
				Exporters:  []string{"logging"},
			},
			Logs: PipelineConfig{
				Receivers:  []string{"otlp"},
				Processors: []string{"memory_limiter", "batch"},
				Exporters:  []string{"logging"},
			},
			Traces: PipelineConfig{
				Receivers:  []string{"otlp"},
				Processors: []string{"memory_limiter", "batch"},
				Exporters:  []string{"logging"},
			},
		},
		Extensions: ExtensionsConfig{
			Health: HealthCheckConfig{
				Enabled:  true,
				Endpoint: "0.0.0.0:13133",
				Path:     "/",
			},
			ZPages: ZPagesConfig{
				Enabled:  false,
				Endpoint: "0.0.0.0:55679",
			},
			PPROF: PPROFConfig{
				Enabled:  false,
				Endpoint: "0.0.0.0:1777",
			},
		},
		Logging: LoggingConfig{
			Level:       "info",
			Format:      "json",
			File:        "",
			MaxSizeMB:   100,
			MaxBackups:  3,
			MaxAgeDays:  7,
			Development: false,
			Sampling: LogSamplingConfig{
				Enabled:    true,
				Initial:    100,
				Thereafter: 100,
			},
		},
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate receivers
	if !c.Receivers.OTLP.Enabled && !c.Receivers.Prometheus.Enabled {
		return ErrNoReceiversEnabled
	}

	if c.Receivers.OTLP.Enabled {
		if !c.Receivers.OTLP.Protocols.GRPC.Enabled && !c.Receivers.OTLP.Protocols.HTTP.Enabled {
			return ErrNoOTLPProtocolsEnabled
		}
	}

	return nil
}

// Errors
var (
	ErrNoReceiversEnabled     = configError("at least one receiver must be enabled")
	ErrNoOTLPProtocolsEnabled = configError("OTLP receiver is enabled but no protocols are configured")
)

type configError string

func (e configError) Error() string {
	return string(e)
}
