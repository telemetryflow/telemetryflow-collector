// Package debug provides exports for testing.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// This file exports internal types and methods for testing purposes only.
package debug

// TestExporterExports provides access to unexported Exporter fields for testing.
type TestExporterExports struct {
	E *Exporter
}

// Config returns the exporter config.
func (e *TestExporterExports) Config() Config {
	return e.E.config
}

// ExportExporter wraps an Exporter for testing access.
func ExportExporter(exp *Exporter) *TestExporterExports {
	return &TestExporterExports{E: exp}
}
