// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// This file registers all component factories for the TFO Collector.
// It includes both OCB-generated community components and TFO custom components.

package main

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"

	// ==========================================================================
	// TelemetryFlow Custom Components
	// ==========================================================================

	// TFO Extensions
	"github.com/telemetryflow/telemetryflow-collector/components/extension/tfoauthextension"
	"github.com/telemetryflow/telemetryflow-collector/components/extension/tfoidentityextension"

	// TFO Receiver
	"github.com/telemetryflow/telemetryflow-collector/components/tfootlpreceiver"

	// TFO Exporter
	"github.com/telemetryflow/telemetryflow-collector/components/tfoexporter"

	// ==========================================================================
	// OpenTelemetry Collector Core Components
	// ==========================================================================

	// Core Extensions
	"go.opentelemetry.io/collector/extension/zpagesextension"

	// Core Receivers
	"go.opentelemetry.io/collector/receiver/otlpreceiver"

	// Core Processors
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"

	// Core Exporters
	"go.opentelemetry.io/collector/exporter/debugexporter"
	"go.opentelemetry.io/collector/exporter/nopexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"

	// Core Connectors
	"go.opentelemetry.io/collector/connector/forwardconnector"

	// ==========================================================================
	// OpenTelemetry Collector Contrib Components
	// ==========================================================================

	// Contrib Extensions
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/basicauthextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/bearertokenauthextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension"

	// Contrib Receivers
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jaegerreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/zipkinreceiver"

	// Contrib Processors
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/tailsamplingprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"

	// Contrib Exporters
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter"

	// Contrib Connectors
	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/countconnector"
	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/servicegraphconnector"
	"github.com/open-telemetry/opentelemetry-collector-contrib/connector/spanmetricsconnector"
)

// components returns all component factories for the TFO Collector.
func components() (otelcol.Factories, error) {
	factories := otelcol.Factories{}

	// Extensions - build factory map manually
	factories.Extensions = make(map[component.Type]extension.Factory)
	for _, f := range []extension.Factory{
		// TFO Custom Extensions
		tfoauthextension.NewFactory(),
		tfoidentityextension.NewFactory(),

		// Core Extensions
		zpagesextension.NewFactory(),

		// Contrib Extensions
		healthcheckextension.NewFactory(),
		pprofextension.NewFactory(),
		basicauthextension.NewFactory(),
		bearertokenauthextension.NewFactory(),
	} {
		factories.Extensions[f.Type()] = f
	}

	// Receivers - build factory map manually
	factories.Receivers = make(map[component.Type]receiver.Factory)
	for _, f := range []receiver.Factory{
		// TFO Custom Receiver
		tfootlpreceiver.NewFactory(),

		// Core Receivers
		otlpreceiver.NewFactory(),

		// Contrib Receivers
		jaegerreceiver.NewFactory(),
		zipkinreceiver.NewFactory(),
		prometheusreceiver.NewFactory(),
		hostmetricsreceiver.NewFactory(),
		filelogreceiver.NewFactory(),
	} {
		factories.Receivers[f.Type()] = f
	}

	// Processors - build factory map manually
	factories.Processors = make(map[component.Type]processor.Factory)
	for _, f := range []processor.Factory{
		// Core Processors
		batchprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory(),

		// Contrib Processors
		attributesprocessor.NewFactory(),
		resourceprocessor.NewFactory(),
		resourcedetectionprocessor.NewFactory(),
		filterprocessor.NewFactory(),
		transformprocessor.NewFactory(),
		tailsamplingprocessor.NewFactory(),
	} {
		factories.Processors[f.Type()] = f
	}

	// Exporters - build factory map manually
	factories.Exporters = make(map[component.Type]exporter.Factory)
	for _, f := range []exporter.Factory{
		// TFO Custom Exporter
		tfoexporter.NewFactory(),

		// Core Exporters
		debugexporter.NewFactory(),
		nopexporter.NewFactory(),
		otlpexporter.NewFactory(),
		otlphttpexporter.NewFactory(),

		// Contrib Exporters
		prometheusexporter.NewFactory(),
		prometheusremotewriteexporter.NewFactory(),
		fileexporter.NewFactory(),
	} {
		factories.Exporters[f.Type()] = f
	}

	// Connectors - build factory map manually
	factories.Connectors = make(map[component.Type]connector.Factory)
	for _, f := range []connector.Factory{
		// Core Connectors
		forwardconnector.NewFactory(),

		// Contrib Connectors (for Exemplars support)
		spanmetricsconnector.NewFactory(),
		servicegraphconnector.NewFactory(),
		countconnector.NewFactory(),
	} {
		factories.Connectors[f.Type()] = f
	}

	return factories, nil
}
