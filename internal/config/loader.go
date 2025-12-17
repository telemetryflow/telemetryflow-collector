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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Loader handles configuration loading from multiple sources
type Loader struct {
	configPaths []string
	envPrefix   string
}

// NewLoader creates a new configuration loader
func NewLoader() *Loader {
	return &Loader{
		configPaths: []string{
			".",
			"./configs",
			"/etc/tfo-collector",
			"$HOME/.tfo-collector",
		},
		envPrefix: "TFCOLLECTOR",
	}
}

// WithConfigPaths adds additional config search paths
func (l *Loader) WithConfigPaths(paths ...string) *Loader {
	l.configPaths = append(l.configPaths, paths...)
	return l
}

// WithEnvPrefix sets the environment variable prefix
func (l *Loader) WithEnvPrefix(prefix string) *Loader {
	l.envPrefix = prefix
	return l
}

// Load loads the configuration from file and environment
func (l *Loader) Load(configFile string) (*Config, error) {
	v := viper.New()

	// Set defaults
	l.setDefaults(v)

	// Configure file search
	v.SetConfigName("tfo-collector")
	v.SetConfigType("yaml")

	// Add config paths
	for _, path := range l.configPaths {
		expandedPath := os.ExpandEnv(path)
		v.AddConfigPath(expandedPath)
	}

	// If explicit config file provided, use it
	if configFile != "" {
		v.SetConfigFile(configFile)
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults + env
	}

	// Configure environment variables
	v.SetEnvPrefix(l.envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	// Bind environment variables explicitly for nested configs
	l.bindEnvVars(v)

	// Unmarshal into config struct
	cfg := DefaultConfig()
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Auto-detect hostname if not set
	if cfg.Collector.Hostname == "" {
		hostname, err := os.Hostname()
		if err == nil {
			cfg.Collector.Hostname = hostname
		}
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// LoadFromFile loads configuration from a specific file
func (l *Loader) LoadFromFile(path string) (*Config, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config path: %w", err)
	}
	return l.Load(absPath)
}

// setDefaults sets default values in viper
func (l *Loader) setDefaults(v *viper.Viper) {
	defaults := DefaultConfig()

	// Collector
	v.SetDefault("collector.id", defaults.Collector.ID)
	v.SetDefault("collector.hostname", defaults.Collector.Hostname)

	// Receivers - OTLP
	v.SetDefault("receivers.otlp.enabled", defaults.Receivers.OTLP.Enabled)
	v.SetDefault("receivers.otlp.protocols.grpc.enabled", defaults.Receivers.OTLP.Protocols.GRPC.Enabled)
	v.SetDefault("receivers.otlp.protocols.grpc.endpoint", defaults.Receivers.OTLP.Protocols.GRPC.Endpoint)
	v.SetDefault("receivers.otlp.protocols.grpc.max_recv_msg_size_mib", defaults.Receivers.OTLP.Protocols.GRPC.MaxRecvMsgSizeMiB)
	v.SetDefault("receivers.otlp.protocols.http.enabled", defaults.Receivers.OTLP.Protocols.HTTP.Enabled)
	v.SetDefault("receivers.otlp.protocols.http.endpoint", defaults.Receivers.OTLP.Protocols.HTTP.Endpoint)
	v.SetDefault("receivers.otlp.protocols.http.max_request_body_size", defaults.Receivers.OTLP.Protocols.HTTP.MaxRequestBodySize)

	// Processors
	v.SetDefault("processors.batch.enabled", defaults.Processors.Batch.Enabled)
	v.SetDefault("processors.batch.send_batch_size", defaults.Processors.Batch.SendBatchSize)
	v.SetDefault("processors.batch.timeout", defaults.Processors.Batch.Timeout)
	v.SetDefault("processors.memory_limiter.enabled", defaults.Processors.Memory.Enabled)
	v.SetDefault("processors.memory_limiter.check_interval", defaults.Processors.Memory.CheckInterval)
	v.SetDefault("processors.memory_limiter.limit_percentage", defaults.Processors.Memory.LimitPercentage)
	v.SetDefault("processors.memory_limiter.spike_limit_percentage", defaults.Processors.Memory.SpikeLimitPercentage)

	// Exporters
	v.SetDefault("exporters.logging.enabled", defaults.Exporters.Logging.Enabled)
	v.SetDefault("exporters.logging.loglevel", defaults.Exporters.Logging.LogLevel)

	// Extensions
	v.SetDefault("extensions.health_check.enabled", defaults.Extensions.Health.Enabled)
	v.SetDefault("extensions.health_check.endpoint", defaults.Extensions.Health.Endpoint)
	v.SetDefault("extensions.health_check.path", defaults.Extensions.Health.Path)
	v.SetDefault("extensions.zpages.enabled", defaults.Extensions.ZPages.Enabled)
	v.SetDefault("extensions.zpages.endpoint", defaults.Extensions.ZPages.Endpoint)
	v.SetDefault("extensions.pprof.enabled", defaults.Extensions.PPROF.Enabled)
	v.SetDefault("extensions.pprof.endpoint", defaults.Extensions.PPROF.Endpoint)

	// Logging
	v.SetDefault("logging.level", defaults.Logging.Level)
	v.SetDefault("logging.format", defaults.Logging.Format)
	v.SetDefault("logging.max_size_mb", defaults.Logging.MaxSizeMB)
	v.SetDefault("logging.max_backups", defaults.Logging.MaxBackups)
	v.SetDefault("logging.max_age_days", defaults.Logging.MaxAgeDays)
	v.SetDefault("logging.development", defaults.Logging.Development)
}

// bindEnvVars explicitly binds environment variables
func (l *Loader) bindEnvVars(v *viper.Viper) {
	// Critical env vars that need explicit binding
	envBindings := map[string]string{
		// Collector
		"collector.id":       "TELEMETRYFLOW_COLLECTOR_ID",
		"collector.hostname": "TELEMETRYFLOW_HOSTNAME",

		// OTLP Receiver
		"receivers.otlp.protocols.grpc.endpoint": "TELEMETRYFLOW_OTLP_GRPC_ENDPOINT",
		"receivers.otlp.protocols.http.endpoint": "TELEMETRYFLOW_OTLP_HTTP_ENDPOINT",

		// Health check
		"extensions.health_check.endpoint": "TELEMETRYFLOW_HEALTH_ENDPOINT",

		// Logging
		"logging.level":  "TELEMETRYFLOW_LOG_LEVEL",
		"logging.format": "TELEMETRYFLOW_LOG_FORMAT",

		// TLS
		"receivers.otlp.protocols.grpc.tls.cert_file": "TELEMETRYFLOW_TLS_CERT_FILE",
		"receivers.otlp.protocols.grpc.tls.key_file":  "TELEMETRYFLOW_TLS_KEY_FILE",
		"receivers.otlp.protocols.grpc.tls.ca_file":   "TELEMETRYFLOW_TLS_CA_FILE",
	}

	for key, env := range envBindings {
		_ = v.BindEnv(key, env)
	}
}

// GetConfigFilePath returns the path of the loaded config file
func GetConfigFilePath() string {
	return viper.ConfigFileUsed()
}
