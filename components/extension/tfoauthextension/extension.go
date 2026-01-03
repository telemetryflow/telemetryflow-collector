// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package tfoauthextension

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
	"go.uber.org/zap"
)

// tfoAuthExtension provides TFO API key authentication.
type tfoAuthExtension struct {
	cfg      *Config
	settings *extension.Settings
	logger   *zap.Logger
	client   *http.Client
}

// newTFOAuthExtension creates a new TFO auth extension.
func newTFOAuthExtension(cfg *Config, set *extension.Settings) (*tfoAuthExtension, error) {
	return &tfoAuthExtension{
		cfg:      cfg,
		settings: set,
		logger:   set.Logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Start implements component.Component.
func (e *tfoAuthExtension) Start(ctx context.Context, host component.Host) error {
	e.logger.Info("TFO auth extension started",
		zap.String("api_key_id", maskAPIKey(string(e.cfg.APIKeyID))),
		zap.Bool("validate_on_start", e.cfg.ValidateOnStart),
	)

	if e.cfg.ValidateOnStart && e.cfg.ValidationEndpoint != "" {
		if err := e.validateCredentials(ctx); err != nil {
			return fmt.Errorf("API key validation failed: %w", err)
		}
		e.logger.Info("API key validated successfully")
	}

	return nil
}

// Shutdown implements component.Component.
func (e *tfoAuthExtension) Shutdown(ctx context.Context) error {
	e.logger.Info("TFO auth extension stopped")
	return nil
}

// GetAPIKeyID returns the API Key ID.
// Implements the AuthProvider interface for tfoexporter.
func (e *tfoAuthExtension) GetAPIKeyID() string {
	return string(e.cfg.APIKeyID)
}

// GetAPIKeySecret returns the API Key Secret.
// Implements the AuthProvider interface for tfoexporter.
func (e *tfoAuthExtension) GetAPIKeySecret() string {
	return string(e.cfg.APIKeySecret)
}

// validateCredentials validates the API key against the validation endpoint.
func (e *tfoAuthExtension) validateCredentials(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.cfg.ValidationEndpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	req.Header.Set("X-TelemetryFlow-Key-ID", string(e.cfg.APIKeyID))
	req.Header.Set("X-TelemetryFlow-Key-Secret", string(e.cfg.APIKeySecret))

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("validation request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read and discard body
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("invalid API credentials")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected validation response: %d", resp.StatusCode)
	}

	return nil
}

// maskAPIKey masks an API key for logging (shows first 8 chars only).
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:8] + "****"
}
