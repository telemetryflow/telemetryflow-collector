// Package exporter_test provides unit tests for the debug exporter.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package exporter_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"

	"github.com/telemetryflow/telemetryflow-collector/internal/exporter/debug"
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
			cfg := debug.Config{Verbosity: tt.verbosity}

			e := debug.New(cfg, logger)

			require.NotNil(t, e)
		})
	}
}

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

			cfg := debug.Config{Verbosity: tt.verbosity}
			e := debug.New(cfg, logger)

			td := createTestTraces(tt.spanCount)
			err := e.ExportTraces(context.Background(), td)

			require.NoError(t, err)

			stats := e.Stats()
			assert.Equal(t, int64(tt.spanCount), stats.TracesExported)
			assert.NotEmpty(t, logs.All())
		})
	}
}

func TestExportTracesEmpty(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := debug.Config{Verbosity: "basic"}
	e := debug.New(cfg, logger)

	td := ptrace.NewTraces()
	err := e.ExportTraces(context.Background(), td)

	require.NoError(t, err)
	assert.Equal(t, int64(0), e.Stats().TracesExported)
}

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zap.InfoLevel)
			logger := zap.New(core)

			cfg := debug.Config{Verbosity: tt.verbosity}
			e := debug.New(cfg, logger)

			md := createTestMetrics(tt.dataPointCount)
			err := e.ExportMetrics(context.Background(), md)

			require.NoError(t, err)

			stats := e.Stats()
			assert.Equal(t, int64(tt.dataPointCount), stats.MetricsExported)
			assert.NotEmpty(t, logs.All())
		})
	}
}

func TestExportMetricsEmpty(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := debug.Config{Verbosity: "basic"}
	e := debug.New(cfg, logger)

	md := pmetric.NewMetrics()
	err := e.ExportMetrics(context.Background(), md)

	require.NoError(t, err)
	assert.Equal(t, int64(0), e.Stats().MetricsExported)
}

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zap.InfoLevel)
			logger := zap.New(core)

			cfg := debug.Config{Verbosity: tt.verbosity}
			e := debug.New(cfg, logger)

			ld := createTestLogs(tt.logRecordCount)
			err := e.ExportLogs(context.Background(), ld)

			require.NoError(t, err)

			stats := e.Stats()
			assert.Equal(t, int64(tt.logRecordCount), stats.LogsExported)
			assert.NotEmpty(t, logs.All())
		})
	}
}

func TestExportLogsEmpty(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := debug.Config{Verbosity: "basic"}
	e := debug.New(cfg, logger)

	ld := plog.NewLogs()
	err := e.ExportLogs(context.Background(), ld)

	require.NoError(t, err)
	assert.Equal(t, int64(0), e.Stats().LogsExported)
}

func TestStats(t *testing.T) {
	logger := zap.NewNop()
	cfg := debug.Config{Verbosity: "basic"}
	e := debug.New(cfg, logger)

	stats := e.Stats()
	assert.Equal(t, int64(0), stats.TracesExported)
	assert.Equal(t, int64(0), stats.MetricsExported)
	assert.Equal(t, int64(0), stats.LogsExported)

	ctx := context.Background()
	_ = e.ExportTraces(ctx, createTestTraces(3))
	_ = e.ExportMetrics(ctx, createTestMetrics(5))
	_ = e.ExportLogs(ctx, createTestLogs(2))

	stats = e.Stats()
	assert.Equal(t, int64(3), stats.TracesExported)
	assert.Equal(t, int64(5), stats.MetricsExported)
	assert.Equal(t, int64(2), stats.LogsExported)
}

func TestStatsAccumulate(t *testing.T) {
	logger := zap.NewNop()
	cfg := debug.Config{Verbosity: "basic"}
	e := debug.New(cfg, logger)

	ctx := context.Background()

	_ = e.ExportTraces(ctx, createTestTraces(2))
	_ = e.ExportTraces(ctx, createTestTraces(3))
	_ = e.ExportTraces(ctx, createTestTraces(5))

	stats := e.Stats()
	assert.Equal(t, int64(10), stats.TracesExported)
}

func TestExporterStatsStruct(t *testing.T) {
	stats := debug.ExporterStats{
		TracesExported:  100,
		MetricsExported: 200,
		LogsExported:    300,
	}

	assert.Equal(t, int64(100), stats.TracesExported)
	assert.Equal(t, int64(200), stats.MetricsExported)
	assert.Equal(t, int64(300), stats.LogsExported)
}

func TestConfigStruct(t *testing.T) {
	cfg := debug.Config{Verbosity: "detailed"}
	assert.Equal(t, "detailed", cfg.Verbosity)
}

func TestContextCancellation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := debug.Config{Verbosity: "basic"}
	e := debug.New(cfg, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := e.ExportTraces(ctx, createTestTraces(1))
	assert.NoError(t, err)

	err = e.ExportMetrics(ctx, createTestMetrics(1))
	assert.NoError(t, err)

	err = e.ExportLogs(ctx, createTestLogs(1))
	assert.NoError(t, err)
}

func TestMultipleResourceSpans(t *testing.T) {
	logger := zap.NewNop()
	cfg := debug.Config{Verbosity: "basic"}
	e := debug.New(cfg, logger)

	td := ptrace.NewTraces()

	rs1 := td.ResourceSpans().AppendEmpty()
	rs1.Resource().Attributes().PutStr("service.name", "service-1")
	ss1 := rs1.ScopeSpans().AppendEmpty()
	ss1.Spans().AppendEmpty().SetName("span-1")

	rs2 := td.ResourceSpans().AppendEmpty()
	rs2.Resource().Attributes().PutStr("service.name", "service-2")
	ss2 := rs2.ScopeSpans().AppendEmpty()
	ss2.Spans().AppendEmpty().SetName("span-2")

	err := e.ExportTraces(context.Background(), td)
	require.NoError(t, err)

	stats := e.Stats()
	assert.Equal(t, int64(2), stats.TracesExported)
}

// Benchmark tests
func BenchmarkExportTraces(b *testing.B) {
	logger := zap.NewNop()
	cfg := debug.Config{Verbosity: "basic"}
	e := debug.New(cfg, logger)

	td := createTestTraces(10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.ExportTraces(ctx, td)
	}
}

func BenchmarkExportMetrics(b *testing.B) {
	logger := zap.NewNop()
	cfg := debug.Config{Verbosity: "basic"}
	e := debug.New(cfg, logger)

	md := createTestMetrics(10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.ExportMetrics(ctx, md)
	}
}

func BenchmarkExportLogs(b *testing.B) {
	logger := zap.NewNop()
	cfg := debug.Config{Verbosity: "basic"}
	e := debug.New(cfg, logger)

	ld := createTestLogs(10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.ExportLogs(ctx, ld)
	}
}
