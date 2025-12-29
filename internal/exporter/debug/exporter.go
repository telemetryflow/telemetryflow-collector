// Package debug provides a debug exporter for the TelemetryFlow Collector.
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
package debug

import (
	"context"
	"sync/atomic"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
)

// Config holds the debug exporter configuration
type Config struct {
	Verbosity string // "basic", "normal", "detailed"
}

// Exporter is a debug exporter that logs telemetry data
type Exporter struct {
	config Config
	logger *zap.Logger

	// Stats
	tracesExported  atomic.Int64
	metricsExported atomic.Int64
	logsExported    atomic.Int64
}

// New creates a new debug exporter
func New(cfg Config, logger *zap.Logger) *Exporter {
	return &Exporter{
		config: cfg,
		logger: logger,
	}
}

// ExportTraces exports traces to the debug log
func (e *Exporter) ExportTraces(ctx context.Context, td ptrace.Traces) error {
	spanCount := td.SpanCount()
	e.tracesExported.Add(int64(spanCount))

	resourceSpans := td.ResourceSpans()
	for i := 0; i < resourceSpans.Len(); i++ {
		rs := resourceSpans.At(i)
		resource := rs.Resource()

		scopeSpans := rs.ScopeSpans()
		for j := 0; j < scopeSpans.Len(); j++ {
			ss := scopeSpans.At(j)
			scope := ss.Scope()

			spans := ss.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)

				if e.config.Verbosity == "detailed" {
					e.logger.Info("Trace",
						zap.String("trace_id", span.TraceID().String()),
						zap.String("span_id", span.SpanID().String()),
						zap.String("name", span.Name()),
						zap.String("kind", span.Kind().String()),
						zap.Time("start_time", span.StartTimestamp().AsTime()),
						zap.Time("end_time", span.EndTimestamp().AsTime()),
						zap.String("status", span.Status().Code().String()),
						zap.String("scope_name", scope.Name()),
						zap.Any("resource_attributes", resource.Attributes().AsRaw()),
					)
				} else {
					e.logger.Info("Trace",
						zap.String("trace_id", span.TraceID().String()),
						zap.String("name", span.Name()),
						zap.String("kind", span.Kind().String()),
					)
				}
			}
		}
	}

	e.logger.Debug("Exported traces",
		zap.Int("span_count", spanCount),
		zap.Int("resource_spans", resourceSpans.Len()),
	)

	return nil
}

// ExportMetrics exports metrics to the debug log
func (e *Exporter) ExportMetrics(ctx context.Context, md pmetric.Metrics) error {
	dataPointCount := md.DataPointCount()
	e.metricsExported.Add(int64(dataPointCount))

	resourceMetrics := md.ResourceMetrics()
	for i := 0; i < resourceMetrics.Len(); i++ {
		rm := resourceMetrics.At(i)
		resource := rm.Resource()

		scopeMetrics := rm.ScopeMetrics()
		for j := 0; j < scopeMetrics.Len(); j++ {
			sm := scopeMetrics.At(j)
			scope := sm.Scope()

			metrics := sm.Metrics()
			for k := 0; k < metrics.Len(); k++ {
				metric := metrics.At(k)

				if e.config.Verbosity == "detailed" {
					e.logger.Info("Metric",
						zap.String("name", metric.Name()),
						zap.String("description", metric.Description()),
						zap.String("unit", metric.Unit()),
						zap.String("type", metric.Type().String()),
						zap.String("scope_name", scope.Name()),
						zap.Any("resource_attributes", resource.Attributes().AsRaw()),
					)
				} else {
					e.logger.Info("Metric",
						zap.String("name", metric.Name()),
						zap.String("type", metric.Type().String()),
					)
				}
			}
		}
	}

	e.logger.Debug("Exported metrics",
		zap.Int("data_point_count", dataPointCount),
		zap.Int("resource_metrics", resourceMetrics.Len()),
	)

	return nil
}

// ExportLogs exports logs to the debug log
func (e *Exporter) ExportLogs(ctx context.Context, ld plog.Logs) error {
	logRecordCount := ld.LogRecordCount()
	e.logsExported.Add(int64(logRecordCount))

	resourceLogs := ld.ResourceLogs()
	for i := 0; i < resourceLogs.Len(); i++ {
		rl := resourceLogs.At(i)
		resource := rl.Resource()

		scopeLogs := rl.ScopeLogs()
		for j := 0; j < scopeLogs.Len(); j++ {
			sl := scopeLogs.At(j)
			scope := sl.Scope()

			logRecords := sl.LogRecords()
			for k := 0; k < logRecords.Len(); k++ {
				lr := logRecords.At(k)

				if e.config.Verbosity == "detailed" {
					e.logger.Info("Log",
						zap.String("severity", lr.SeverityText()),
						zap.Int("severity_number", int(lr.SeverityNumber())),
						zap.String("body", lr.Body().AsString()),
						zap.Time("timestamp", lr.Timestamp().AsTime()),
						zap.String("trace_id", lr.TraceID().String()),
						zap.String("span_id", lr.SpanID().String()),
						zap.String("scope_name", scope.Name()),
						zap.Any("resource_attributes", resource.Attributes().AsRaw()),
					)
				} else {
					e.logger.Info("Log",
						zap.String("severity", lr.SeverityText()),
						zap.String("body", lr.Body().AsString()),
					)
				}
			}
		}
	}

	e.logger.Debug("Exported logs",
		zap.Int("log_record_count", logRecordCount),
		zap.Int("resource_logs", resourceLogs.Len()),
	)

	return nil
}

// Stats returns exporter statistics
func (e *Exporter) Stats() ExporterStats {
	return ExporterStats{
		TracesExported:  e.tracesExported.Load(),
		MetricsExported: e.metricsExported.Load(),
		LogsExported:    e.logsExported.Load(),
	}
}

// ExporterStats contains exporter statistics
type ExporterStats struct {
	TracesExported  int64 `json:"traces_exported"`
	MetricsExported int64 `json:"metrics_exported"`
	LogsExported    int64 `json:"logs_exported"`
}
