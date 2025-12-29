// Package collector provides exports for testing.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// This file exports internal types and methods for testing purposes only.
// It is only compiled during testing (due to _test.go suffix).
package collector

import (
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"

	"github.com/telemetryflow/telemetryflow-collector/internal/exporter/debug"
	"github.com/telemetryflow/telemetryflow-collector/internal/pipeline"
	"github.com/telemetryflow/telemetryflow-collector/internal/receiver/otlp"
)

// TestCollectorExports provides access to unexported Collector fields for testing.
type TestCollectorExports struct {
	C *Collector
}

// DebugExporter returns the debug exporter.
func (e *TestCollectorExports) DebugExporter() *debug.Exporter {
	return e.C.debugExporter
}

// Pipeline returns the pipeline.
func (e *TestCollectorExports) Pipeline() *pipeline.Pipeline {
	return e.C.pipeline
}

// OTLPReceiver returns the OTLP receiver.
func (e *TestCollectorExports) OTLPReceiver() *otlp.Receiver {
	return e.C.otlpReceiver
}

// ExportCollector wraps a Collector for testing access.
func ExportCollector(c *Collector) *TestCollectorExports {
	return &TestCollectorExports{C: c}
}

// TestOTELCollectorExports provides access to unexported OTELCollector fields for testing.
type TestOTELCollectorExports struct {
	C *OTELCollector
}

// ConfigPath returns the config path.
func (e *TestOTELCollectorExports) ConfigPath() string {
	return e.C.configPath
}

// Logger returns the logger.
func (e *TestOTELCollectorExports) Logger() *zap.Logger {
	return e.C.logger
}

// Version returns the build info.
func (e *TestOTELCollectorExports) Version() component.BuildInfo {
	return e.C.version
}

// Service returns the OTEL service.
func (e *TestOTELCollectorExports) Service() *otelcol.Collector {
	return e.C.service
}

// ExportOTELCollector wraps an OTELCollector for testing access.
func ExportOTELCollector(c *OTELCollector) *TestOTELCollectorExports {
	return &TestOTELCollectorExports{C: c}
}

// ExportComponents exposes the components method for testing.
func (c *OTELCollector) ExportComponents() (otelcol.Factories, error) {
	return c.components()
}
