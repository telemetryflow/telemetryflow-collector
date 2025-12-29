// Package pipeline provides exports for testing.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// This file exports internal types and methods for testing purposes only.
package pipeline

// TestPipelineExports provides access to unexported Pipeline fields for testing.
type TestPipelineExports struct {
	P *Pipeline
}

// TraceExportersLen returns the number of trace exporters.
func (e *TestPipelineExports) TraceExportersLen() int {
	return len(e.P.traceExporters)
}

// MetricsExportersLen returns the number of metrics exporters.
func (e *TestPipelineExports) MetricsExportersLen() int {
	return len(e.P.metricsExporters)
}

// LogsExportersLen returns the number of logs exporters.
func (e *TestPipelineExports) LogsExportersLen() int {
	return len(e.P.logsExporters)
}

// ExportPipeline wraps a Pipeline for testing access.
func ExportPipeline(p *Pipeline) *TestPipelineExports {
	return &TestPipelineExports{P: p}
}
