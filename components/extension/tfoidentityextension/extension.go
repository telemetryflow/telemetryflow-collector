// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package tfoidentityextension

import (
	"context"
	"os"

	"github.com/google/uuid"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
	"go.uber.org/zap"
)

// tfoIdentityExtension provides collector identity management.
type tfoIdentityExtension struct {
	cfg      *Config
	settings *extension.Settings
	logger   *zap.Logger

	// Resolved identity
	collectorID string
	hostname    string
}

// newTFOIdentityExtension creates a new TFO identity extension.
func newTFOIdentityExtension(cfg *Config, set *extension.Settings) (*tfoIdentityExtension, error) {
	return &tfoIdentityExtension{
		cfg:      cfg,
		settings: set,
		logger:   set.Logger,
	}, nil
}

// Start implements component.Component.
func (e *tfoIdentityExtension) Start(ctx context.Context, host component.Host) error {
	// Resolve collector ID
	e.collectorID = e.cfg.ID
	if e.collectorID == "" {
		e.collectorID = uuid.New().String()
		e.logger.Info("Generated collector ID", zap.String("id", e.collectorID))
	}

	// Resolve hostname
	e.hostname = e.cfg.Hostname
	if e.hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			e.logger.Warn("Failed to detect hostname", zap.Error(err))
			e.hostname = "unknown"
		} else {
			e.hostname = hostname
		}
	}

	e.logger.Info("TFO identity extension started",
		zap.String("collector_id", e.collectorID),
		zap.String("hostname", e.hostname),
		zap.String("name", e.cfg.Name),
		zap.Any("tags", e.cfg.Tags),
		zap.Bool("enrich_resources", e.cfg.EnrichResources),
	)

	return nil
}

// Shutdown implements component.Component.
func (e *tfoIdentityExtension) Shutdown(ctx context.Context) error {
	e.logger.Info("TFO identity extension stopped")
	return nil
}

// GetCollectorID returns the collector ID.
// Implements the IdentityProvider interface for tfoexporter.
func (e *tfoIdentityExtension) GetCollectorID() string {
	return e.collectorID
}

// GetHostname returns the collector hostname.
func (e *tfoIdentityExtension) GetHostname() string {
	return e.hostname
}

// GetName returns the collector name.
func (e *tfoIdentityExtension) GetName() string {
	return e.cfg.Name
}

// GetDescription returns the collector description.
func (e *tfoIdentityExtension) GetDescription() string {
	return e.cfg.Description
}

// GetTags returns the collector tags.
func (e *tfoIdentityExtension) GetTags() map[string]string {
	return e.cfg.Tags
}

// ShouldEnrichResources returns whether resources should be enriched.
func (e *tfoIdentityExtension) ShouldEnrichResources() bool {
	return e.cfg.EnrichResources
}
