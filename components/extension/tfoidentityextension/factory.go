// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package tfoidentityextension

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

const (
	// TypeStr is the type string identifier for the TFO identity extension.
	TypeStr = "tfoidentity"
)

// NewFactory creates a new factory for the TFO identity extension.
func NewFactory() extension.Factory {
	return extension.NewFactory(
		component.MustNewType(TypeStr),
		createDefaultConfig,
		createExtension,
		component.StabilityLevelStable,
	)
}

// createDefaultConfig creates the default configuration for the extension.
func createDefaultConfig() component.Config {
	return &Config{
		Name:            "TelemetryFlow Collector",
		Description:     "TelemetryFlow Collector - Community Enterprise Observability Platform",
		Tags:            make(map[string]string),
		EnrichResources: true,
	}
}

// createExtension creates the TFO identity extension.
func createExtension(
	ctx context.Context,
	set extension.Settings,
	cfg component.Config,
) (extension.Extension, error) {
	return newTFOIdentityExtension(cfg.(*Config), &set)
}
