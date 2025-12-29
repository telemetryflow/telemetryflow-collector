// Package pipeline provides telemetry data pipeline for the TelemetryFlow Collector.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package pipeline

import (
	"context"
	"sync"
	"sync/atomic"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

// TraceExporter exports trace data
type TraceExporter interface {
	ExportTraces(ctx context.Context, td ptrace.Traces) error
}

// MetricsExporter exports metrics data
type MetricsExporter interface {
	ExportMetrics(ctx context.Context, md pmetric.Metrics) error
}

// LogsExporter exports logs data
type LogsExporter interface {
	ExportLogs(ctx context.Context, ld plog.Logs) error
}

// Pipeline manages the flow of telemetry data from receivers to exporters
type Pipeline struct {
	logger *zap.Logger

	// Exporters
	traceExporters   []TraceExporter
	metricsExporters []MetricsExporter
	logsExporters    []LogsExporter

	mu sync.RWMutex

	// Stats
	tracesProcessed  atomic.Int64
	metricsProcessed atomic.Int64
	logsProcessed    atomic.Int64
	tracesDropped    atomic.Int64
	metricsDropped   atomic.Int64
	logsDropped      atomic.Int64
}

// New creates a new pipeline
func New(logger *zap.Logger) *Pipeline {
	return &Pipeline{
		logger:           logger,
		traceExporters:   make([]TraceExporter, 0),
		metricsExporters: make([]MetricsExporter, 0),
		logsExporters:    make([]LogsExporter, 0),
	}
}

// AddTraceExporter adds a trace exporter to the pipeline
func (p *Pipeline) AddTraceExporter(e TraceExporter) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.traceExporters = append(p.traceExporters, e)
}

// AddMetricsExporter adds a metrics exporter to the pipeline
func (p *Pipeline) AddMetricsExporter(e MetricsExporter) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.metricsExporters = append(p.metricsExporters, e)
}

// AddLogsExporter adds a logs exporter to the pipeline
func (p *Pipeline) AddLogsExporter(e LogsExporter) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.logsExporters = append(p.logsExporters, e)
}

// ConsumeTraces implements the Consumer interface for traces
func (p *Pipeline) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	p.mu.RLock()
	exporters := p.traceExporters
	p.mu.RUnlock()

	spanCount := td.SpanCount()
	p.tracesProcessed.Add(int64(spanCount))

	if len(exporters) == 0 {
		p.logger.Debug("No trace exporters configured, dropping traces",
			zap.Int("span_count", spanCount),
		)
		p.tracesDropped.Add(int64(spanCount))
		return nil
	}

	var lastErr error
	for _, exporter := range exporters {
		if err := exporter.ExportTraces(ctx, td); err != nil {
			p.logger.Error("Failed to export traces", zap.Error(err))
			lastErr = err
		}
	}

	return lastErr
}

// ConsumeMetrics implements the Consumer interface for metrics
func (p *Pipeline) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	p.mu.RLock()
	exporters := p.metricsExporters
	p.mu.RUnlock()

	dataPointCount := md.DataPointCount()
	p.metricsProcessed.Add(int64(dataPointCount))

	if len(exporters) == 0 {
		p.logger.Debug("No metrics exporters configured, dropping metrics",
			zap.Int("data_point_count", dataPointCount),
		)
		p.metricsDropped.Add(int64(dataPointCount))
		return nil
	}

	var lastErr error
	for _, exporter := range exporters {
		if err := exporter.ExportMetrics(ctx, md); err != nil {
			p.logger.Error("Failed to export metrics", zap.Error(err))
			lastErr = err
		}
	}

	return lastErr
}

// ConsumeLogs implements the Consumer interface for logs
func (p *Pipeline) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	p.mu.RLock()
	exporters := p.logsExporters
	p.mu.RUnlock()

	logRecordCount := ld.LogRecordCount()
	p.logsProcessed.Add(int64(logRecordCount))

	if len(exporters) == 0 {
		p.logger.Debug("No logs exporters configured, dropping logs",
			zap.Int("log_record_count", logRecordCount),
		)
		p.logsDropped.Add(int64(logRecordCount))
		return nil
	}

	var lastErr error
	for _, exporter := range exporters {
		if err := exporter.ExportLogs(ctx, ld); err != nil {
			p.logger.Error("Failed to export logs", zap.Error(err))
			lastErr = err
		}
	}

	return lastErr
}

// Stats returns pipeline statistics
func (p *Pipeline) Stats() PipelineStats {
	return PipelineStats{
		TracesProcessed:  p.tracesProcessed.Load(),
		MetricsProcessed: p.metricsProcessed.Load(),
		LogsProcessed:    p.logsProcessed.Load(),
		TracesDropped:    p.tracesDropped.Load(),
		MetricsDropped:   p.metricsDropped.Load(),
		LogsDropped:      p.logsDropped.Load(),
	}
}

// PipelineStats contains pipeline statistics
type PipelineStats struct {
	TracesProcessed  int64 `json:"traces_processed"`
	MetricsProcessed int64 `json:"metrics_processed"`
	LogsProcessed    int64 `json:"logs_processed"`
	TracesDropped    int64 `json:"traces_dropped"`
	MetricsDropped   int64 `json:"metrics_dropped"`
	LogsDropped      int64 `json:"logs_dropped"`
}
