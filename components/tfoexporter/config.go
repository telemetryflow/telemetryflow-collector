// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package tfoexporter

import (
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/config/configretry"
)

// Config defines the configuration for the TFO exporter.
type Config struct {
	confighttp.ClientConfig `mapstructure:",squash"`

	// UseV2API enables the TFO Platform v2 API endpoints.
	// When true, exports to /v2/traces, /v2/metrics, /v2/logs instead of v1.
	// Default: true
	UseV2API bool `mapstructure:"use_v2_api"`

	// Auth configures authentication for the TFO Platform.
	Auth *AuthConfig `mapstructure:"auth"`

	// CollectorIdentity is a reference to a tfoidentity extension for collector metadata.
	CollectorIdentity component.ID `mapstructure:"collector_identity"`

	// RetryConfig configures retry on failure.
	RetryConfig configretry.BackOffConfig `mapstructure:"retry_on_failure"`

	// TracesEndpoint overrides the default traces endpoint path.
	TracesEndpoint string `mapstructure:"traces_endpoint"`

	// MetricsEndpoint overrides the default metrics endpoint path.
	MetricsEndpoint string `mapstructure:"metrics_endpoint"`

	// LogsEndpoint overrides the default logs endpoint path.
	LogsEndpoint string `mapstructure:"logs_endpoint"`
}

// AuthConfig defines authentication configuration.
type AuthConfig struct {
	// APIKeyID is the TelemetryFlow API Key ID (format: tfk_xxx).
	APIKeyID configopaque.String `mapstructure:"api_key_id"`

	// APIKeySecret is the TelemetryFlow API Key Secret (format: tfs_xxx).
	APIKeySecret configopaque.String `mapstructure:"api_key_secret"`

	// Extension is a reference to a tfoauth extension for authentication.
	// If set, takes precedence over APIKeyID/APIKeySecret.
	Extension component.ID `mapstructure:"extension"`
}

// Validate checks the configuration for errors.
func (cfg *Config) Validate() error {
	if cfg.Endpoint == "" {
		return errors.New("endpoint is required")
	}

	// Validate auth configuration
	if cfg.Auth != nil {
		hasDirectAuth := cfg.Auth.APIKeyID != "" && cfg.Auth.APIKeySecret != ""
		hasExtensionAuth := cfg.Auth.Extension.String() != ""

		if !hasDirectAuth && !hasExtensionAuth {
			return errors.New("auth requires either api_key_id/api_key_secret or extension reference")
		}
	}

	return nil
}

// GetTracesEndpoint returns the traces endpoint path.
func (cfg *Config) GetTracesEndpoint() string {
	if cfg.TracesEndpoint != "" {
		return cfg.TracesEndpoint
	}
	if cfg.UseV2API {
		return "/v2/traces"
	}
	return "/v1/traces"
}

// GetMetricsEndpoint returns the metrics endpoint path.
func (cfg *Config) GetMetricsEndpoint() string {
	if cfg.MetricsEndpoint != "" {
		return cfg.MetricsEndpoint
	}
	if cfg.UseV2API {
		return "/v2/metrics"
	}
	return "/v1/metrics"
}

// GetLogsEndpoint returns the logs endpoint path.
func (cfg *Config) GetLogsEndpoint() string {
	if cfg.LogsEndpoint != "" {
		return cfg.LogsEndpoint
	}
	if cfg.UseV2API {
		return "/v2/logs"
	}
	return "/v1/logs"
}
