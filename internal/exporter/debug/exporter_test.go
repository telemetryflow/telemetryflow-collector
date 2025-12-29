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
	"testing"
	"time"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

// Helper functions to create test data
func createTestTraces(spanCount int) ptrace.Traces {
	td := ptrace.NewTraces()
	rs := td.ResourceSpans().AppendEmpty()
	rs.Resource().Attributes().PutStr("service.name", "test-service")
	rs.Resource().Attributes().PutStr("deployment.environment", "test")

	ss := rs.ScopeSpans().AppendEmpty()
	ss.Scope().SetName("test-scope")
	ss.Scope().SetVersion("1.0.0")

	for i := 0; i < spanCount; i++ {
		span := ss.Spans().AppendEmpty()
		span.SetName("test-span")
		span.SetTraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
		span.SetSpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
		span.SetKind(ptrace.SpanKindServer)
		span.SetStartTimestamp(pcommon.NewTimestampFromTime(time.Now().Add(-time.Second)))
		span.SetEndTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		span.Status().SetCode(ptrace.StatusCodeOk)
	}

	return td
}

func createTestMetrics(dataPointCount int) pmetric.Metrics {
	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr("service.name", "test-service")
	rm.Resource().Attributes().PutStr("deployment.environment", "test")

	sm := rm.ScopeMetrics().AppendEmpty()
	sm.Scope().SetName("test-scope")
	sm.Scope().SetVersion("1.0.0")

	metric := sm.Metrics().AppendEmpty()
	metric.SetName("test-metric")
	metric.SetDescription("A test metric for unit testing")
	metric.SetUnit("1")

	gauge := metric.SetEmptyGauge()
	for i := 0; i < dataPointCount; i++ {
		dp := gauge.DataPoints().AppendEmpty()
		dp.SetDoubleValue(float64(i * 10))
		dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	}

	return md
}

func createTestLogs(logRecordCount int) plog.Logs {
	ld := plog.NewLogs()
	rl := ld.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutStr("service.name", "test-service")
	rl.Resource().Attributes().PutStr("deployment.environment", "test")

	sl := rl.ScopeLogs().AppendEmpty()
	sl.Scope().SetName("test-scope")
	sl.Scope().SetVersion("1.0.0")

	for i := 0; i < logRecordCount; i++ {
		lr := sl.LogRecords().AppendEmpty()
		lr.Body().SetStr("test log message")
		lr.SetSeverityText("INFO")
		lr.SetSeverityNumber(plog.SeverityNumberInfo)
		lr.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		lr.SetTraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
		lr.SetSpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	}

	return ld
}

// TestNew tests debug exporter creation
func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		verbosity string
	}{
		{"basic verbosity", "basic"},
		{"normal verbosity", "normal"},
		{"detailed verbosity", "detailed"},
		{"empty verbosity", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			cfg := Config{Verbosity: tt.verbosity}

			e := New(cfg, logger)

			if e == nil {
				t.Fatal("Expected non-nil exporter")
			}

			if e.config.Verbosity != tt.verbosity {
				t.Errorf("Expected verbosity %s, got %s", tt.verbosity, e.config.Verbosity)
			}

			if e.logger == nil {
				t.Error("Expected logger to be set")
			}
		})
	}
}

// TestExportTraces tests trace export
func TestExportTraces(t *testing.T) {
	tests := []struct {
		name      string
		verbosity string
		spanCount int
	}{
		{"basic single span", "basic", 1},
		{"basic multiple spans", "basic", 5},
		{"normal single span", "normal", 1},
		{"detailed single span", "detailed", 1},
		{"detailed multiple spans", "detailed", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zap.InfoLevel)
			logger := zap.New(core)

			cfg := Config{Verbosity: tt.verbosity}
			e := New(cfg, logger)

			td := createTestTraces(tt.spanCount)
			err := e.ExportTraces(context.Background(), td)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify stats
			stats := e.Stats()
			if stats.TracesExported != int64(tt.spanCount) {
				t.Errorf("Expected %d traces exported, got %d", tt.spanCount, stats.TracesExported)
			}

			// Verify logs were written
			if logs.Len() == 0 {
				t.Error("Expected log entries, got none")
			}

			// For detailed verbosity, verify more fields are logged
			if tt.verbosity == "detailed" {
				found := false
				for _, entry := range logs.All() {
					for _, field := range entry.Context {
						if field.Key == "resource_attributes" {
							found = true
							break
						}
					}
				}
				if !found {
					t.Error("Expected resource_attributes in detailed logs")
				}
			}
		})
	}
}

// TestExportTracesEmpty tests export with empty traces
func TestExportTracesEmpty(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := Config{Verbosity: "basic"}
	e := New(cfg, logger)

	td := ptrace.NewTraces()
	err := e.ExportTraces(context.Background(), td)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stats := e.Stats()
	if stats.TracesExported != 0 {
		t.Errorf("Expected 0 traces exported, got %d", stats.TracesExported)
	}
}

// TestExportMetrics tests metrics export
func TestExportMetrics(t *testing.T) {
	tests := []struct {
		name           string
		verbosity      string
		dataPointCount int
	}{
		{"basic single point", "basic", 1},
		{"basic multiple points", "basic", 5},
		{"normal single point", "normal", 1},
		{"detailed single point", "detailed", 1},
		{"detailed multiple points", "detailed", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zap.InfoLevel)
			logger := zap.New(core)

			cfg := Config{Verbosity: tt.verbosity}
			e := New(cfg, logger)

			md := createTestMetrics(tt.dataPointCount)
			err := e.ExportMetrics(context.Background(), md)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify stats
			stats := e.Stats()
			if stats.MetricsExported != int64(tt.dataPointCount) {
				t.Errorf("Expected %d metrics exported, got %d", tt.dataPointCount, stats.MetricsExported)
			}

			// Verify logs were written
			if logs.Len() == 0 {
				t.Error("Expected log entries, got none")
			}

			// For detailed verbosity, verify more fields are logged
			if tt.verbosity == "detailed" {
				found := false
				for _, entry := range logs.All() {
					for _, field := range entry.Context {
						if field.Key == "resource_attributes" {
							found = true
							break
						}
					}
				}
				if !found {
					t.Error("Expected resource_attributes in detailed logs")
				}
			}
		})
	}
}

// TestExportMetricsEmpty tests export with empty metrics
func TestExportMetricsEmpty(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := Config{Verbosity: "basic"}
	e := New(cfg, logger)

	md := pmetric.NewMetrics()
	err := e.ExportMetrics(context.Background(), md)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stats := e.Stats()
	if stats.MetricsExported != 0 {
		t.Errorf("Expected 0 metrics exported, got %d", stats.MetricsExported)
	}
}

// TestExportLogs tests logs export
func TestExportLogs(t *testing.T) {
	tests := []struct {
		name           string
		verbosity      string
		logRecordCount int
	}{
		{"basic single log", "basic", 1},
		{"basic multiple logs", "basic", 5},
		{"normal single log", "normal", 1},
		{"detailed single log", "detailed", 1},
		{"detailed multiple logs", "detailed", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zap.InfoLevel)
			logger := zap.New(core)

			cfg := Config{Verbosity: tt.verbosity}
			e := New(cfg, logger)

			ld := createTestLogs(tt.logRecordCount)
			err := e.ExportLogs(context.Background(), ld)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify stats
			stats := e.Stats()
			if stats.LogsExported != int64(tt.logRecordCount) {
				t.Errorf("Expected %d logs exported, got %d", tt.logRecordCount, stats.LogsExported)
			}

			// Verify logs were written
			if logs.Len() == 0 {
				t.Error("Expected log entries, got none")
			}

			// For detailed verbosity, verify more fields are logged
			if tt.verbosity == "detailed" {
				found := false
				for _, entry := range logs.All() {
					for _, field := range entry.Context {
						if field.Key == "resource_attributes" {
							found = true
							break
						}
					}
				}
				if !found {
					t.Error("Expected resource_attributes in detailed logs")
				}
			}
		})
	}
}

// TestExportLogsEmpty tests export with empty logs
func TestExportLogsEmpty(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := Config{Verbosity: "basic"}
	e := New(cfg, logger)

	ld := plog.NewLogs()
	err := e.ExportLogs(context.Background(), ld)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stats := e.Stats()
	if stats.LogsExported != 0 {
		t.Errorf("Expected 0 logs exported, got %d", stats.LogsExported)
	}
}

// TestStats tests exporter statistics
func TestStats(t *testing.T) {
	logger := zap.NewNop()
	cfg := Config{Verbosity: "basic"}
	e := New(cfg, logger)

	// Initial stats should be zero
	stats := e.Stats()
	if stats.TracesExported != 0 {
		t.Errorf("Expected 0 traces exported, got %d", stats.TracesExported)
	}
	if stats.MetricsExported != 0 {
		t.Errorf("Expected 0 metrics exported, got %d", stats.MetricsExported)
	}
	if stats.LogsExported != 0 {
		t.Errorf("Expected 0 logs exported, got %d", stats.LogsExported)
	}

	// Export some data
	ctx := context.Background()
	_ = e.ExportTraces(ctx, createTestTraces(3))
	_ = e.ExportMetrics(ctx, createTestMetrics(5))
	_ = e.ExportLogs(ctx, createTestLogs(2))

	// Verify stats
	stats = e.Stats()
	if stats.TracesExported != 3 {
		t.Errorf("Expected 3 traces exported, got %d", stats.TracesExported)
	}
	if stats.MetricsExported != 5 {
		t.Errorf("Expected 5 metrics exported, got %d", stats.MetricsExported)
	}
	if stats.LogsExported != 2 {
		t.Errorf("Expected 2 logs exported, got %d", stats.LogsExported)
	}
}

// TestStatsAccumulate tests that stats accumulate correctly
func TestStatsAccumulate(t *testing.T) {
	logger := zap.NewNop()
	cfg := Config{Verbosity: "basic"}
	e := New(cfg, logger)

	ctx := context.Background()

	// Export traces multiple times
	_ = e.ExportTraces(ctx, createTestTraces(2))
	_ = e.ExportTraces(ctx, createTestTraces(3))
	_ = e.ExportTraces(ctx, createTestTraces(5))

	stats := e.Stats()
	if stats.TracesExported != 10 {
		t.Errorf("Expected 10 traces exported, got %d", stats.TracesExported)
	}
}

// TestExporterStatsStruct tests ExporterStats struct fields
func TestExporterStatsStruct(t *testing.T) {
	stats := ExporterStats{
		TracesExported:  100,
		MetricsExported: 200,
		LogsExported:    300,
	}

	if stats.TracesExported != 100 {
		t.Errorf("Expected TracesExported 100, got %d", stats.TracesExported)
	}
	if stats.MetricsExported != 200 {
		t.Errorf("Expected MetricsExported 200, got %d", stats.MetricsExported)
	}
	if stats.LogsExported != 300 {
		t.Errorf("Expected LogsExported 300, got %d", stats.LogsExported)
	}
}

// TestConfigStruct tests Config struct
func TestConfigStruct(t *testing.T) {
	cfg := Config{
		Verbosity: "detailed",
	}

	if cfg.Verbosity != "detailed" {
		t.Errorf("Expected Verbosity 'detailed', got %s", cfg.Verbosity)
	}
}

// TestContextCancellation tests export with cancelled context
func TestContextCancellation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := Config{Verbosity: "basic"}
	e := New(cfg, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Export should still succeed (current implementation doesn't check context)
	// This test verifies the behavior is consistent
	err := e.ExportTraces(ctx, createTestTraces(1))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	err = e.ExportMetrics(ctx, createTestMetrics(1))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	err = e.ExportLogs(ctx, createTestLogs(1))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestMultipleResourceSpans tests handling multiple resource spans
func TestMultipleResourceSpans(t *testing.T) {
	logger := zap.NewNop()
	cfg := Config{Verbosity: "basic"}
	e := New(cfg, logger)

	td := ptrace.NewTraces()

	// Add first resource span
	rs1 := td.ResourceSpans().AppendEmpty()
	rs1.Resource().Attributes().PutStr("service.name", "service-1")
	ss1 := rs1.ScopeSpans().AppendEmpty()
	ss1.Spans().AppendEmpty().SetName("span-1")

	// Add second resource span
	rs2 := td.ResourceSpans().AppendEmpty()
	rs2.Resource().Attributes().PutStr("service.name", "service-2")
	ss2 := rs2.ScopeSpans().AppendEmpty()
	ss2.Spans().AppendEmpty().SetName("span-2")

	err := e.ExportTraces(context.Background(), td)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stats := e.Stats()
	if stats.TracesExported != 2 {
		t.Errorf("Expected 2 traces exported, got %d", stats.TracesExported)
	}
}

// TestMultipleScopeSpans tests handling multiple scope spans
func TestMultipleScopeSpans(t *testing.T) {
	logger := zap.NewNop()
	cfg := Config{Verbosity: "basic"}
	e := New(cfg, logger)

	td := ptrace.NewTraces()
	rs := td.ResourceSpans().AppendEmpty()
	rs.Resource().Attributes().PutStr("service.name", "test-service")

	// Add first scope span
	ss1 := rs.ScopeSpans().AppendEmpty()
	ss1.Scope().SetName("scope-1")
	ss1.Spans().AppendEmpty().SetName("span-1")

	// Add second scope span
	ss2 := rs.ScopeSpans().AppendEmpty()
	ss2.Scope().SetName("scope-2")
	ss2.Spans().AppendEmpty().SetName("span-2")

	err := e.ExportTraces(context.Background(), td)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	stats := e.Stats()
	if stats.TracesExported != 2 {
		t.Errorf("Expected 2 traces exported, got %d", stats.TracesExported)
	}
}

// Benchmark tests
func BenchmarkExportTraces(b *testing.B) {
	logger := zap.NewNop()
	cfg := Config{Verbosity: "basic"}
	e := New(cfg, logger)

	td := createTestTraces(10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.ExportTraces(ctx, td)
	}
}

func BenchmarkExportTracesDetailed(b *testing.B) {
	logger := zap.NewNop()
	cfg := Config{Verbosity: "detailed"}
	e := New(cfg, logger)

	td := createTestTraces(10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.ExportTraces(ctx, td)
	}
}

func BenchmarkExportMetrics(b *testing.B) {
	logger := zap.NewNop()
	cfg := Config{Verbosity: "basic"}
	e := New(cfg, logger)

	md := createTestMetrics(10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.ExportMetrics(ctx, md)
	}
}

func BenchmarkExportLogs(b *testing.B) {
	logger := zap.NewNop()
	cfg := Config{Verbosity: "basic"}
	e := New(cfg, logger)

	ld := createTestLogs(10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.ExportLogs(ctx, ld)
	}
}
