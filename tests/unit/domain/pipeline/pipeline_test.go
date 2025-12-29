// Package pipeline_test provides unit tests for the pipeline package.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package pipeline_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/telemetryflow/telemetryflow-collector/internal/pipeline"
)

// mockTraceExporter is a mock trace exporter for testing
type mockTraceExporter struct {
	mu           sync.Mutex
	traces       []ptrace.Traces
	exportErr    error
	exportCount  atomic.Int64
	exportCalled atomic.Bool
}

func (m *mockTraceExporter) ExportTraces(_ context.Context, td ptrace.Traces) error {
	m.exportCalled.Store(true)
	m.exportCount.Add(1)

	if m.exportErr != nil {
		return m.exportErr
	}

	m.mu.Lock()
	m.traces = append(m.traces, td)
	m.mu.Unlock()

	return nil
}

// mockMetricsExporter is a mock metrics exporter for testing
type mockMetricsExporter struct {
	mu           sync.Mutex
	metrics      []pmetric.Metrics
	exportErr    error
	exportCount  atomic.Int64
	exportCalled atomic.Bool
}

func (m *mockMetricsExporter) ExportMetrics(_ context.Context, md pmetric.Metrics) error {
	m.exportCalled.Store(true)
	m.exportCount.Add(1)

	if m.exportErr != nil {
		return m.exportErr
	}

	m.mu.Lock()
	m.metrics = append(m.metrics, md)
	m.mu.Unlock()

	return nil
}

// mockLogsExporter is a mock logs exporter for testing
type mockLogsExporter struct {
	mu           sync.Mutex
	logs         []plog.Logs
	exportErr    error
	exportCount  atomic.Int64
	exportCalled atomic.Bool
}

func (m *mockLogsExporter) ExportLogs(_ context.Context, ld plog.Logs) error {
	m.exportCalled.Store(true)
	m.exportCount.Add(1)

	if m.exportErr != nil {
		return m.exportErr
	}

	m.mu.Lock()
	m.logs = append(m.logs, ld)
	m.mu.Unlock()

	return nil
}

// Helper functions to create test data
func createTestTraces(spanCount int) ptrace.Traces {
	td := ptrace.NewTraces()
	rs := td.ResourceSpans().AppendEmpty()
	rs.Resource().Attributes().PutStr("service.name", "test-service")

	ss := rs.ScopeSpans().AppendEmpty()
	ss.Scope().SetName("test-scope")

	for i := 0; i < spanCount; i++ {
		span := ss.Spans().AppendEmpty()
		span.SetName("test-span")
		span.SetTraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
		span.SetSpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	}

	return td
}

func createTestMetrics(dataPointCount int) pmetric.Metrics {
	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr("service.name", "test-service")

	sm := rm.ScopeMetrics().AppendEmpty()
	sm.Scope().SetName("test-scope")

	metric := sm.Metrics().AppendEmpty()
	metric.SetName("test-metric")
	metric.SetDescription("A test metric")
	metric.SetUnit("1")

	gauge := metric.SetEmptyGauge()
	for i := 0; i < dataPointCount; i++ {
		dp := gauge.DataPoints().AppendEmpty()
		dp.SetDoubleValue(float64(i))
	}

	return md
}

func createTestLogs(logRecordCount int) plog.Logs {
	ld := plog.NewLogs()
	rl := ld.ResourceLogs().AppendEmpty()
	rl.Resource().Attributes().PutStr("service.name", "test-service")

	sl := rl.ScopeLogs().AppendEmpty()
	sl.Scope().SetName("test-scope")

	for i := 0; i < logRecordCount; i++ {
		lr := sl.LogRecords().AppendEmpty()
		lr.Body().SetStr("test log message")
		lr.SetSeverityText("INFO")
	}

	return ld
}

func TestNew(t *testing.T) {
	logger := zaptest.NewLogger(t)

	p := pipeline.New(logger)

	require.NotNil(t, p)

	// Verify pipeline stats are zero initially
	stats := p.Stats()
	assert.Equal(t, int64(0), stats.TracesProcessed)
	assert.Equal(t, int64(0), stats.MetricsProcessed)
	assert.Equal(t, int64(0), stats.LogsProcessed)
}

func TestAddTraceExporter(t *testing.T) {
	logger := zaptest.NewLogger(t)
	p := pipeline.New(logger)

	exporter1 := &mockTraceExporter{}
	exporter2 := &mockTraceExporter{}

	p.AddTraceExporter(exporter1)
	p.AddTraceExporter(exporter2)

	// Verify by consuming traces and checking both exporters received them
	td := createTestTraces(1)
	err := p.ConsumeTraces(context.Background(), td)
	require.NoError(t, err)

	assert.True(t, exporter1.exportCalled.Load())
	assert.True(t, exporter2.exportCalled.Load())
}

func TestAddMetricsExporter(t *testing.T) {
	logger := zaptest.NewLogger(t)
	p := pipeline.New(logger)

	exporter1 := &mockMetricsExporter{}
	exporter2 := &mockMetricsExporter{}

	p.AddMetricsExporter(exporter1)
	p.AddMetricsExporter(exporter2)

	// Verify by consuming metrics and checking both exporters received them
	md := createTestMetrics(1)
	err := p.ConsumeMetrics(context.Background(), md)
	require.NoError(t, err)

	assert.True(t, exporter1.exportCalled.Load())
	assert.True(t, exporter2.exportCalled.Load())
}

func TestAddLogsExporter(t *testing.T) {
	logger := zaptest.NewLogger(t)
	p := pipeline.New(logger)

	exporter1 := &mockLogsExporter{}
	exporter2 := &mockLogsExporter{}

	p.AddLogsExporter(exporter1)
	p.AddLogsExporter(exporter2)

	// Verify by consuming logs and checking both exporters received them
	ld := createTestLogs(1)
	err := p.ConsumeLogs(context.Background(), ld)
	require.NoError(t, err)

	assert.True(t, exporter1.exportCalled.Load())
	assert.True(t, exporter2.exportCalled.Load())
}

func TestConsumeTraces(t *testing.T) {
	tests := []struct {
		name          string
		spanCount     int
		exporterCount int
	}{
		{"single exporter single span", 1, 1},
		{"multiple exporters multiple spans", 5, 3},
		{"no exporters", 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			p := pipeline.New(logger)

			var exporters []*mockTraceExporter
			for i := 0; i < tt.exporterCount; i++ {
				exporter := &mockTraceExporter{}
				exporters = append(exporters, exporter)
				p.AddTraceExporter(exporter)
			}

			td := createTestTraces(tt.spanCount)
			err := p.ConsumeTraces(context.Background(), td)

			require.NoError(t, err)

			for i, exporter := range exporters {
				assert.True(t, exporter.exportCalled.Load(), "Exporter %d was not called", i)
			}

			stats := p.Stats()
			assert.Equal(t, int64(tt.spanCount), stats.TracesProcessed)

			if tt.exporterCount == 0 {
				assert.Equal(t, int64(tt.spanCount), stats.TracesDropped)
			}
		})
	}
}

func TestConsumeTracesWithError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	p := pipeline.New(logger)

	expectedErr := errors.New("export failed")
	exporter := &mockTraceExporter{exportErr: expectedErr}
	p.AddTraceExporter(exporter)

	td := createTestTraces(1)
	err := p.ConsumeTraces(context.Background(), td)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
}

func TestConsumeMetrics(t *testing.T) {
	tests := []struct {
		name           string
		dataPointCount int
		exporterCount  int
	}{
		{"single exporter single data point", 1, 1},
		{"multiple exporters multiple data points", 5, 3},
		{"no exporters", 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			p := pipeline.New(logger)

			var exporters []*mockMetricsExporter
			for i := 0; i < tt.exporterCount; i++ {
				exporter := &mockMetricsExporter{}
				exporters = append(exporters, exporter)
				p.AddMetricsExporter(exporter)
			}

			md := createTestMetrics(tt.dataPointCount)
			err := p.ConsumeMetrics(context.Background(), md)

			require.NoError(t, err)

			for i, exporter := range exporters {
				assert.True(t, exporter.exportCalled.Load(), "Exporter %d was not called", i)
			}

			stats := p.Stats()
			assert.Equal(t, int64(tt.dataPointCount), stats.MetricsProcessed)

			if tt.exporterCount == 0 {
				assert.Equal(t, int64(tt.dataPointCount), stats.MetricsDropped)
			}
		})
	}
}

func TestConsumeMetricsWithError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	p := pipeline.New(logger)

	expectedErr := errors.New("export failed")
	exporter := &mockMetricsExporter{exportErr: expectedErr}
	p.AddMetricsExporter(exporter)

	md := createTestMetrics(1)
	err := p.ConsumeMetrics(context.Background(), md)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
}

func TestConsumeLogs(t *testing.T) {
	tests := []struct {
		name           string
		logRecordCount int
		exporterCount  int
	}{
		{"single exporter single log", 1, 1},
		{"multiple exporters multiple logs", 5, 3},
		{"no exporters", 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			p := pipeline.New(logger)

			var exporters []*mockLogsExporter
			for i := 0; i < tt.exporterCount; i++ {
				exporter := &mockLogsExporter{}
				exporters = append(exporters, exporter)
				p.AddLogsExporter(exporter)
			}

			ld := createTestLogs(tt.logRecordCount)
			err := p.ConsumeLogs(context.Background(), ld)

			require.NoError(t, err)

			for i, exporter := range exporters {
				assert.True(t, exporter.exportCalled.Load(), "Exporter %d was not called", i)
			}

			stats := p.Stats()
			assert.Equal(t, int64(tt.logRecordCount), stats.LogsProcessed)

			if tt.exporterCount == 0 {
				assert.Equal(t, int64(tt.logRecordCount), stats.LogsDropped)
			}
		})
	}
}

func TestConsumeLogsWithError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	p := pipeline.New(logger)

	expectedErr := errors.New("export failed")
	exporter := &mockLogsExporter{exportErr: expectedErr}
	p.AddLogsExporter(exporter)

	ld := createTestLogs(1)
	err := p.ConsumeLogs(context.Background(), ld)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, expectedErr))
}

func TestStats(t *testing.T) {
	logger := zaptest.NewLogger(t)
	p := pipeline.New(logger)

	stats := p.Stats()
	assert.Equal(t, int64(0), stats.TracesProcessed)
	assert.Equal(t, int64(0), stats.MetricsProcessed)
	assert.Equal(t, int64(0), stats.LogsProcessed)

	p.AddTraceExporter(&mockTraceExporter{})
	p.AddMetricsExporter(&mockMetricsExporter{})
	p.AddLogsExporter(&mockLogsExporter{})

	ctx := context.Background()
	_ = p.ConsumeTraces(ctx, createTestTraces(3))
	_ = p.ConsumeMetrics(ctx, createTestMetrics(5))
	_ = p.ConsumeLogs(ctx, createTestLogs(2))

	stats = p.Stats()
	assert.Equal(t, int64(3), stats.TracesProcessed)
	assert.Equal(t, int64(5), stats.MetricsProcessed)
	assert.Equal(t, int64(2), stats.LogsProcessed)
}

func TestMultipleExportersPartialFailure(t *testing.T) {
	logger := zaptest.NewLogger(t)
	p := pipeline.New(logger)

	exporter1 := &mockTraceExporter{}
	exporter2 := &mockTraceExporter{exportErr: errors.New("failed")}
	exporter3 := &mockTraceExporter{}

	p.AddTraceExporter(exporter1)
	p.AddTraceExporter(exporter2)
	p.AddTraceExporter(exporter3)

	td := createTestTraces(1)
	err := p.ConsumeTraces(context.Background(), td)

	assert.Error(t, err)
	assert.True(t, exporter1.exportCalled.Load())
	assert.True(t, exporter2.exportCalled.Load())
	assert.True(t, exporter3.exportCalled.Load())
}

func TestConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	p := pipeline.New(logger)

	exporter := &mockTraceExporter{}
	p.AddTraceExporter(exporter)

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			td := createTestTraces(1)
			_ = p.ConsumeTraces(context.Background(), td)
		}()
	}

	wg.Wait()

	stats := p.Stats()
	assert.Equal(t, int64(numGoroutines), stats.TracesProcessed)
}

func TestConcurrentAddExporter(t *testing.T) {
	logger := zap.NewNop()
	p := pipeline.New(logger)

	var wg sync.WaitGroup
	numGoroutines := 50

	traceExporters := make([]*mockTraceExporter, numGoroutines)
	metricsExporters := make([]*mockMetricsExporter, numGoroutines)
	logsExporters := make([]*mockLogsExporter, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		traceExporters[i] = &mockTraceExporter{}
		metricsExporters[i] = &mockMetricsExporter{}
		logsExporters[i] = &mockLogsExporter{}
	}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(3)
		idx := i
		go func() {
			defer wg.Done()
			p.AddTraceExporter(traceExporters[idx])
		}()
		go func() {
			defer wg.Done()
			p.AddMetricsExporter(metricsExporters[idx])
		}()
		go func() {
			defer wg.Done()
			p.AddLogsExporter(logsExporters[idx])
		}()
	}

	wg.Wait()

	// Verify all exporters work by consuming data
	ctx := context.Background()
	_ = p.ConsumeTraces(ctx, createTestTraces(1))
	_ = p.ConsumeMetrics(ctx, createTestMetrics(1))
	_ = p.ConsumeLogs(ctx, createTestLogs(1))

	// All exporters should have been called
	for i := 0; i < numGoroutines; i++ {
		assert.True(t, traceExporters[i].exportCalled.Load())
		assert.True(t, metricsExporters[i].exportCalled.Load())
		assert.True(t, logsExporters[i].exportCalled.Load())
	}
}

func TestPipelineStatsStruct(t *testing.T) {
	stats := pipeline.PipelineStats{
		TracesProcessed:  100,
		MetricsProcessed: 200,
		LogsProcessed:    300,
		TracesDropped:    10,
		MetricsDropped:   20,
		LogsDropped:      30,
	}

	assert.Equal(t, int64(100), stats.TracesProcessed)
	assert.Equal(t, int64(200), stats.MetricsProcessed)
	assert.Equal(t, int64(300), stats.LogsProcessed)
	assert.Equal(t, int64(10), stats.TracesDropped)
	assert.Equal(t, int64(20), stats.MetricsDropped)
	assert.Equal(t, int64(30), stats.LogsDropped)
}

// Benchmark tests
func BenchmarkConsumeTraces(b *testing.B) {
	logger := zap.NewNop()
	p := pipeline.New(logger)
	p.AddTraceExporter(&mockTraceExporter{})

	td := createTestTraces(10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.ConsumeTraces(ctx, td)
	}
}

func BenchmarkConsumeMetrics(b *testing.B) {
	logger := zap.NewNop()
	p := pipeline.New(logger)
	p.AddMetricsExporter(&mockMetricsExporter{})

	md := createTestMetrics(10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.ConsumeMetrics(ctx, md)
	}
}

func BenchmarkConsumeLogs(b *testing.B) {
	logger := zap.NewNop()
	p := pipeline.New(logger)
	p.AddLogsExporter(&mockLogsExporter{})

	ld := createTestLogs(10)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.ConsumeLogs(ctx, ld)
	}
}

func BenchmarkConsumeTracesParallel(b *testing.B) {
	logger := zap.NewNop()
	p := pipeline.New(logger)
	p.AddTraceExporter(&mockTraceExporter{})

	td := createTestTraces(10)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = p.ConsumeTraces(ctx, td)
		}
	})
}
