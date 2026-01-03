// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package tfootlpreceiver

import (
	"errors"

	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/confighttp"
)

// Config defines the configuration for the TFO OTLP receiver.
//
// The receiver supports two modes:
//   - v1 endpoints: Standard OTEL Community endpoints (no auth required)
//   - v2 endpoints: TFO Platform endpoints (requires TFO auth headers)
type Config struct {
	// Protocols defines the protocols (grpc and/or http) that the receiver will use.
	Protocols ProtocolsConfig `mapstructure:"protocols"`

	// EnableV2Endpoints enables TFO Platform v2 endpoints (/v2/traces, /v2/metrics, /v2/logs)
	// in addition to standard v1 endpoints. v2 endpoints require TFO authentication.
	// Default: true
	EnableV2Endpoints bool `mapstructure:"enable_v2_endpoints"`

	// V2Auth configures authentication for v2 endpoints.
	// Only applies when EnableV2Endpoints is true.
	V2Auth V2AuthConfig `mapstructure:"v2_auth"`
}

// V2AuthConfig defines authentication settings for v2 endpoints.
type V2AuthConfig struct {
	// Required when true, v2 endpoints will reject requests without valid TFO auth headers.
	// Default: true
	Required bool `mapstructure:"required"`

	// ValidAPIKeyIDs is a list of valid API Key IDs for authentication.
	// If empty and Required is true, any non-empty API Key ID is accepted.
	ValidAPIKeyIDs []string `mapstructure:"valid_api_key_ids"`

	// ValidateSecret when true, also validates the API Key Secret.
	// Default: false (only validates API Key ID presence)
	ValidateSecret bool `mapstructure:"validate_secret"`
}

// ProtocolsConfig defines the protocol configurations.
type ProtocolsConfig struct {
	// GRPC configures the gRPC server settings.
	GRPC *GRPCConfig `mapstructure:"grpc"`

	// HTTP configures the HTTP server settings.
	HTTP *HTTPConfig `mapstructure:"http"`
}

// GRPCConfig defines the gRPC protocol configuration.
type GRPCConfig struct {
	configgrpc.ServerConfig `mapstructure:",squash"`
}

// HTTPConfig defines the HTTP protocol configuration with TFO-specific settings.
type HTTPConfig struct {
	confighttp.ServerConfig `mapstructure:",squash"`

	// TracesURLPath overrides the default traces path. Default: /v1/traces
	TracesURLPath string `mapstructure:"traces_url_path"`

	// MetricsURLPath overrides the default metrics path. Default: /v1/metrics
	MetricsURLPath string `mapstructure:"metrics_url_path"`

	// LogsURLPath overrides the default logs path. Default: /v1/logs
	LogsURLPath string `mapstructure:"logs_url_path"`
}

// Validate checks the configuration for errors.
func (cfg *Config) Validate() error {
	if cfg.Protocols.GRPC == nil && cfg.Protocols.HTTP == nil {
		// At least one protocol must be enabled - but we'll use defaults
		return nil
	}

	// Validate V2Auth if v2 endpoints are enabled
	if cfg.EnableV2Endpoints && cfg.V2Auth.Required {
		if cfg.V2Auth.ValidateSecret && len(cfg.V2Auth.ValidAPIKeyIDs) > 0 {
			// When validating secrets with specific key IDs, we need a way to store secrets
			// For now, this is a configuration error - use extension-based auth instead
			return errors.New("validate_secret with valid_api_key_ids requires tfoauth extension")
		}
	}

	return nil
}
