// Package collector provides OTEL-based collector with full capabilities.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// This file provides integration with OpenTelemetry Collector framework
// to enable full OTEL capabilities (metrics, logs, traces, exemplars).
package collector

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	// Core receivers
	"go.opentelemetry.io/collector/receiver/otlpreceiver"

	// Core processors
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"

	// Core exporters
	"go.opentelemetry.io/collector/exporter/debugexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"

	// Core extensions
	"go.opentelemetry.io/collector/extension/zpagesextension"

	// Core connectors
	"go.opentelemetry.io/collector/connector/forwardconnector"

	// Contrib receivers
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jaegerreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zipkinreceiver"

	// Contrib processors
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"

	// Contrib exporters
	// Note: lokiexporter deprecated July 2024 - Loki now supports OTLP natively
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter"

	// Contrib extensions
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension"

	// Contrib connectors - Exemplars support
	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/countconnector"
	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/servicegraphconnector"
	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/spanmetricsconnector"
)

// OTELCollector wraps the OpenTelemetry Collector service
type OTELCollector struct {
	service    *otelcol.Collector
	configPath string
	logger     *zap.Logger
	version    component.BuildInfo
}

// NewOTELCollector creates a new OTEL-based collector with full capabilities
func NewOTELCollector(configPath string, logger *zap.Logger, version string) (*OTELCollector, error) {
	buildInfo := component.BuildInfo{
		Command:     "tfo-collector",
		Description: "TelemetryFlow Collector - Community Enterprise Observability Platform",
		Version:     version,
	}

	return &OTELCollector{
		configPath: configPath,
		logger:     logger,
		version:    buildInfo,
	}, nil
}

// components returns all registered OTEL component factories
func (c *OTELCollector) components() (otelcol.Factories, error) {
	var err error
	factories := otelcol.Factories{}

	// ==========================================================================
	// EXTENSIONS
	// ==========================================================================
	factories.Extensions, err = otelcol.MakeFactoryMap[extension.Factory](
		zpagesextension.NewFactory(),
		healthcheckextension.NewFactory(),
		pprofextension.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, fmt.Errorf("failed to create extension factories: %w", err)
	}

	// ==========================================================================
	// RECEIVERS - Metrics, Logs, Traces
	// ==========================================================================
	factories.Receivers, err = otelcol.MakeFactoryMap[receiver.Factory](
		// Core OTLP
		otlpreceiver.NewFactory(),
		// Traces
		jaegerreceiver.NewFactory(),
		zipkinreceiver.NewFactory(),
		// Metrics
		hostmetricsreceiver.NewFactory(),
		prometheusreceiver.NewFactory(),
		// Logs
		filelogreceiver.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, fmt.Errorf("failed to create receiver factories: %w", err)
	}

	// ==========================================================================
	// PROCESSORS - Transform, Sample, Enrich
	// ==========================================================================
	factories.Processors, err = otelcol.MakeFactoryMap[processor.Factory](
		// Core
		batchprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory(),
		// Attributes
		attributesprocessor.NewFactory(),
		resourceprocessor.NewFactory(),
		resourcedetectionprocessor.NewFactory(),
		// Transform
		filterprocessor.NewFactory(),
		transformprocessor.NewFactory(),
		// Sampling
		tailsamplingprocessor.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, fmt.Errorf("failed to create processor factories: %w", err)
	}

	// ==========================================================================
	// EXPORTERS - Send to Backends
	// ==========================================================================
	factories.Exporters, err = otelcol.MakeFactoryMap[exporter.Factory](
		// Core OTLP
		otlpexporter.NewFactory(),
		otlphttpexporter.NewFactory(),
		debugexporter.NewFactory(),
		// Metrics
		prometheusexporter.NewFactory(),
		prometheusremotewriteexporter.NewFactory(),
		// Files
		fileexporter.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, fmt.Errorf("failed to create exporter factories: %w", err)
	}

	// ==========================================================================
	// CONNECTORS - Pipeline Bridging & Exemplars
	// ==========================================================================
	factories.Connectors, err = otelcol.MakeFactoryMap[connector.Factory](
		forwardconnector.NewFactory(),
		// Exemplars support
		spanmetricsconnector.NewFactory(),
		servicegraphconnector.NewFactory(),
		countconnector.NewFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, fmt.Errorf("failed to create connector factories: %w", err)
	}

	return factories, nil
}

// Run starts the OTEL collector and blocks until shutdown
func (c *OTELCollector) Run(ctx context.Context) error {
	factories, err := c.components()
	if err != nil {
		return fmt.Errorf("failed to build component factories: %w", err)
	}

	// Configure providers
	configProviderSettings := otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs: []string{c.configPath},
			ProviderFactories: []confmap.ProviderFactory{
				fileprovider.NewFactory(),
				yamlprovider.NewFactory(),
			},
		},
	}

	// Create collector settings
	settings := otelcol.CollectorSettings{
		BuildInfo:              c.version,
		Factories:              func() (otelcol.Factories, error) { return factories, nil },
		ConfigProviderSettings: configProviderSettings,
	}

	// Create and run collector
	collector, err := otelcol.NewCollector(settings)
	if err != nil {
		return fmt.Errorf("failed to create collector: %w", err)
	}

	c.service = collector
	c.logger.Info("Starting TelemetryFlow OTEL Collector",
		zap.String("version", c.version.Version),
		zap.String("config", c.configPath),
	)

	return collector.Run(ctx)
}

// Shutdown gracefully stops the collector
func (c *OTELCollector) Shutdown() {
	if c.service != nil {
		c.service.Shutdown()
	}
}
