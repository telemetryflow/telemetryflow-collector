// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package tfoauthextension

import (
	"errors"
	"strings"

	"go.opentelemetry.io/collector/config/configopaque"
)

// Config defines the configuration for the TFO auth extension.
type Config struct {
	// APIKeyID is the TelemetryFlow API Key ID (format: tfk_xxx).
	APIKeyID configopaque.String `mapstructure:"api_key_id"`

	// APIKeySecret is the TelemetryFlow API Key Secret (format: tfs_xxx).
	APIKeySecret configopaque.String `mapstructure:"api_key_secret"`

	// ValidationEndpoint is the optional endpoint for validating API keys.
	// If set, the extension will validate credentials on startup.
	ValidationEndpoint string `mapstructure:"validation_endpoint"`

	// ValidateOnStart enables API key validation during extension startup.
	// Default: false
	ValidateOnStart bool `mapstructure:"validate_on_start"`
}

// Validate checks the configuration for errors.
// If API keys are not set (empty), the extension will start in passthrough mode
// where it doesn't inject authentication headers.
func (cfg *Config) Validate() error {
	// If both API key ID and secret are empty, allow passthrough mode
	// This enables the collector to start without TFO authentication configured
	if cfg.APIKeyID == "" && cfg.APIKeySecret == "" {
		return nil
	}

	// If one is set, both must be set
	if cfg.APIKeyID == "" {
		return errors.New("api_key_id is required when api_key_secret is set")
	}
	if cfg.APIKeySecret == "" {
		return errors.New("api_key_secret is required when api_key_id is set")
	}

	// Validate API key format
	keyID := string(cfg.APIKeyID)
	if !strings.HasPrefix(keyID, "tfk_") {
		return errors.New("api_key_id must start with 'tfk_' prefix")
	}

	keySecret := string(cfg.APIKeySecret)
	if !strings.HasPrefix(keySecret, "tfs_") {
		return errors.New("api_key_secret must start with 'tfs_' prefix")
	}

	if cfg.ValidateOnStart && cfg.ValidationEndpoint == "" {
		return errors.New("validation_endpoint is required when validate_on_start is true")
	}

	return nil
}
