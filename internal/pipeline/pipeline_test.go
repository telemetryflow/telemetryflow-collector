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
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// mockTraceExporter is a mock trace exporter for testing
type mockTraceExporter struct {
	mu           sync.Mutex
	traces       []ptrace.Traces
	exportErr    error
	exportCount  atomic.Int64
	shouldFail   bool
	failAfter    int
	exportCalled atomic.Bool
}

func (m *mockTraceExporter) ExportTraces(ctx context.Context, td ptrace.Traces) error {
	m.exportCalled.Store(true)
	count := m.exportCount.Add(1)

	if m.shouldFail && int(count) > m.failAfter {
		return m.exportErr
	}

	m.mu.Lock()
	m.traces = append(m.traces, td)
	m.mu.Unlock()

	if m.exportErr != nil && !m.shouldFail {
		return m.exportErr
	}
	return nil
}

func (m *mockTraceExporter) getTraces() []ptrace.Traces {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.traces
}

// mockMetricsExporter is a mock metrics exporter for testing
type mockMetricsExporter struct {
	mu           sync.Mutex
	metrics      []pmetric.Metrics
	exportErr    error
	exportCount  atomic.Int64
	shouldFail   bool
	failAfter    int
	exportCalled atomic.Bool
}

func (m *mockMetricsExporter) ExportMetrics(ctx context.Context, md pmetric.Metrics) error {
	m.exportCalled.Store(true)
	count := m.exportCount.Add(1)

	if m.shouldFail && int(count) > m.failAfter {
		return m.exportErr
	}

	m.mu.Lock()
	m.metrics = append(m.metrics, md)
	m.mu.Unlock()

	if m.exportErr != nil && !m.shouldFail {
		return m.exportErr
	}
	return nil
}

func (m *mockMetricsExporter) getMetrics() []pmetric.Metrics {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.metrics
}

// mockLogsExporter is a mock logs exporter for testing
type mockLogsExporter struct {
	mu           sync.Mutex
	logs         []plog.Logs
	exportErr    error
	exportCount  atomic.Int64
	shouldFail   bool
	failAfter    int
	exportCalled atomic.Bool
}

func (m *mockLogsExporter) ExportLogs(ctx context.Context, ld plog.Logs) error {
	m.exportCalled.Store(true)
	count := m.exportCount.Add(1)

	if m.shouldFail && int(count) > m.failAfter {
		return m.exportErr
	}

	m.mu.Lock()
	m.logs = append(m.logs, ld)
	m.mu.Unlock()

	if m.exportErr != nil && !m.shouldFail {
		return m.exportErr
	}
	return nil
}

func (m *mockLogsExporter) getLogs() []plog.Logs {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.logs
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

// TestNew tests pipeline creation
func TestNew(t *testing.T) {
	logger := zaptest.NewLogger(t)

	p := New(logger)

	if p == nil {
		t.Fatal("Expected non-nil pipeline")
	}

	if p.logger == nil {
		t.Error("Expected logger to be set")
	}

	if len(p.traceExporters) != 0 {
		t.Error("Expected empty trace exporters")
	}

	if len(p.metricsExporters) != 0 {
		t.Error("Expected empty metrics exporters")
	}

	if len(p.logsExporters) != 0 {
		t.Error("Expected empty logs exporters")
	}
}

// TestAddTraceExporter tests adding trace exporters
func TestAddTraceExporter(t *testing.T) {
	logger := zaptest.NewLogger(t)
	p := New(logger)

	exporter1 := &mockTraceExporter{}
	exporter2 := &mockTraceExporter{}

	p.AddTraceExporter(exporter1)
	if len(p.traceExporters) != 1 {
		t.Errorf("Expected 1 trace exporter, got %d", len(p.traceExporters))
	}

	p.AddTraceExporter(exporter2)
	if len(p.traceExporters) != 2 {
		t.Errorf("Expected 2 trace exporters, got %d", len(p.traceExporters))
	}
}

// TestAddMetricsExporter tests adding metrics exporters
func TestAddMetricsExporter(t *testing.T) {
	logger := zaptest.NewLogger(t)
	p := New(logger)

	exporter1 := &mockMetricsExporter{}
	exporter2 := &mockMetricsExporter{}

	p.AddMetricsExporter(exporter1)
	if len(p.metricsExporters) != 1 {
		t.Errorf("Expected 1 metrics exporter, got %d", len(p.metricsExporters))
	}

	p.AddMetricsExporter(exporter2)
	if len(p.metricsExporters) != 2 {
		t.Errorf("Expected 2 metrics exporters, got %d", len(p.metricsExporters))
	}
}

// TestAddLogsExporter tests adding logs exporters
func TestAddLogsExporter(t *testing.T) {
	logger := zaptest.NewLogger(t)
	p := New(logger)

	exporter1 := &mockLogsExporter{}
	exporter2 := &mockLogsExporter{}

	p.AddLogsExporter(exporter1)
	if len(p.logsExporters) != 1 {
		t.Errorf("Expected 1 logs exporter, got %d", len(p.logsExporters))
	}

	p.AddLogsExporter(exporter2)
	if len(p.logsExporters) != 2 {
		t.Errorf("Expected 2 logs exporters, got %d", len(p.logsExporters))
	}
}

// TestConsumeTraces tests trace consumption
func TestConsumeTraces(t *testing.T) {
	tests := []struct {
		name          string
		spanCount     int
		exporterCount int
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "single exporter single span",
			spanCount:     1,
			exporterCount: 1,
			wantErr:       false,
		},
		{
			name:          "multiple exporters multiple spans",
			spanCount:     5,
			exporterCount: 3,
			wantErr:       false,
		},
		{
			name:          "no exporters",
			spanCount:     1,
			exporterCount: 0,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			p := New(logger)

			var exporters []*mockTraceExporter
			for i := 0; i < tt.exporterCount; i++ {
				exporter := &mockTraceExporter{}
				exporters = append(exporters, exporter)
				p.AddTraceExporter(exporter)
			}

			td := createTestTraces(tt.spanCount)
			err := p.ConsumeTraces(context.Background(), td)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Verify all exporters received traces
			for i, exporter := range exporters {
				if !exporter.exportCalled.Load() {
					t.Errorf("Exporter %d was not called", i)
				}
				if len(exporter.getTraces()) != 1 {
					t.Errorf("Exporter %d expected 1 traces, got %d", i, len(exporter.getTraces()))
				}
			}

			// Verify stats
			stats := p.Stats()
			if stats.TracesProcessed != int64(tt.spanCount) {
				t.Errorf("Expected %d traces processed, got %d", tt.spanCount, stats.TracesProcessed)
			}

			// If no exporters, traces should be dropped
			if tt.exporterCount == 0 && stats.TracesDropped != int64(tt.spanCount) {
				t.Errorf("Expected %d traces dropped, got %d", tt.spanCount, stats.TracesDropped)
			}
		})
	}
}

// TestConsumeTracesWithError tests trace consumption with exporter errors
func TestConsumeTracesWithError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	p := New(logger)

	expectedErr := errors.New("export failed")
	exporter := &mockTraceExporter{exportErr: expectedErr}
	p.AddTraceExporter(exporter)

	td := createTestTraces(1)
	err := p.ConsumeTraces(context.Background(), td)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// TestConsumeMetrics tests metrics consumption
func TestConsumeMetrics(t *testing.T) {
	tests := []struct {
		name           string
		dataPointCount int
		exporterCount  int
		wantErr        bool
	}{
		{
			name:           "single exporter single data point",
			dataPointCount: 1,
			exporterCount:  1,
			wantErr:        false,
		},
		{
			name:           "multiple exporters multiple data points",
			dataPointCount: 5,
			exporterCount:  3,
			wantErr:        false,
		},
		{
			name:           "no exporters",
			dataPointCount: 1,
			exporterCount:  0,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			p := New(logger)

			var exporters []*mockMetricsExporter
			for i := 0; i < tt.exporterCount; i++ {
				exporter := &mockMetricsExporter{}
				exporters = append(exporters, exporter)
				p.AddMetricsExporter(exporter)
			}

			md := createTestMetrics(tt.dataPointCount)
			err := p.ConsumeMetrics(context.Background(), md)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Verify all exporters received metrics
			for i, exporter := range exporters {
				if !exporter.exportCalled.Load() {
					t.Errorf("Exporter %d was not called", i)
				}
				if len(exporter.getMetrics()) != 1 {
					t.Errorf("Exporter %d expected 1 metrics, got %d", i, len(exporter.getMetrics()))
				}
			}

			// Verify stats
			stats := p.Stats()
			if stats.MetricsProcessed != int64(tt.dataPointCount) {
				t.Errorf("Expected %d metrics processed, got %d", tt.dataPointCount, stats.MetricsProcessed)
			}

			// If no exporters, metrics should be dropped
			if tt.exporterCount == 0 && stats.MetricsDropped != int64(tt.dataPointCount) {
				t.Errorf("Expected %d metrics dropped, got %d", tt.dataPointCount, stats.MetricsDropped)
			}
		})
	}
}

// TestConsumeMetricsWithError tests metrics consumption with exporter errors
func TestConsumeMetricsWithError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	p := New(logger)

	expectedErr := errors.New("export failed")
	exporter := &mockMetricsExporter{exportErr: expectedErr}
	p.AddMetricsExporter(exporter)

	md := createTestMetrics(1)
	err := p.ConsumeMetrics(context.Background(), md)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// TestConsumeLogs tests logs consumption
func TestConsumeLogs(t *testing.T) {
	tests := []struct {
		name           string
		logRecordCount int
		exporterCount  int
		wantErr        bool
	}{
		{
			name:           "single exporter single log",
			logRecordCount: 1,
			exporterCount:  1,
			wantErr:        false,
		},
		{
			name:           "multiple exporters multiple logs",
			logRecordCount: 5,
			exporterCount:  3,
			wantErr:        false,
		},
		{
			name:           "no exporters",
			logRecordCount: 1,
			exporterCount:  0,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			p := New(logger)

			var exporters []*mockLogsExporter
			for i := 0; i < tt.exporterCount; i++ {
				exporter := &mockLogsExporter{}
				exporters = append(exporters, exporter)
				p.AddLogsExporter(exporter)
			}

			ld := createTestLogs(tt.logRecordCount)
			err := p.ConsumeLogs(context.Background(), ld)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Verify all exporters received logs
			for i, exporter := range exporters {
				if !exporter.exportCalled.Load() {
					t.Errorf("Exporter %d was not called", i)
				}
				if len(exporter.getLogs()) != 1 {
					t.Errorf("Exporter %d expected 1 logs, got %d", i, len(exporter.getLogs()))
				}
			}

			// Verify stats
			stats := p.Stats()
			if stats.LogsProcessed != int64(tt.logRecordCount) {
				t.Errorf("Expected %d logs processed, got %d", tt.logRecordCount, stats.LogsProcessed)
			}

			// If no exporters, logs should be dropped
			if tt.exporterCount == 0 && stats.LogsDropped != int64(tt.logRecordCount) {
				t.Errorf("Expected %d logs dropped, got %d", tt.logRecordCount, stats.LogsDropped)
			}
		})
	}
}

// TestConsumeLogsWithError tests logs consumption with exporter errors
func TestConsumeLogsWithError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	p := New(logger)

	expectedErr := errors.New("export failed")
	exporter := &mockLogsExporter{exportErr: expectedErr}
	p.AddLogsExporter(exporter)

	ld := createTestLogs(1)
	err := p.ConsumeLogs(context.Background(), ld)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// TestStats tests pipeline statistics
func TestStats(t *testing.T) {
	logger := zaptest.NewLogger(t)
	p := New(logger)

	// Initial stats should be zero
	stats := p.Stats()
	if stats.TracesProcessed != 0 {
		t.Errorf("Expected 0 traces processed, got %d", stats.TracesProcessed)
	}
	if stats.MetricsProcessed != 0 {
		t.Errorf("Expected 0 metrics processed, got %d", stats.MetricsProcessed)
	}
	if stats.LogsProcessed != 0 {
		t.Errorf("Expected 0 logs processed, got %d", stats.LogsProcessed)
	}

	// Add exporters
	p.AddTraceExporter(&mockTraceExporter{})
	p.AddMetricsExporter(&mockMetricsExporter{})
	p.AddLogsExporter(&mockLogsExporter{})

	// Process some data
	ctx := context.Background()
	_ = p.ConsumeTraces(ctx, createTestTraces(3))
	_ = p.ConsumeMetrics(ctx, createTestMetrics(5))
	_ = p.ConsumeLogs(ctx, createTestLogs(2))

	// Verify stats
	stats = p.Stats()
	if stats.TracesProcessed != 3 {
		t.Errorf("Expected 3 traces processed, got %d", stats.TracesProcessed)
	}
	if stats.MetricsProcessed != 5 {
		t.Errorf("Expected 5 metrics processed, got %d", stats.MetricsProcessed)
	}
	if stats.LogsProcessed != 2 {
		t.Errorf("Expected 2 logs processed, got %d", stats.LogsProcessed)
	}
}

// TestMultipleExportersPartialFailure tests that all exporters are called even when some fail
func TestMultipleExportersPartialFailure(t *testing.T) {
	logger := zaptest.NewLogger(t)
	p := New(logger)

	// Add multiple exporters, middle one fails
	exporter1 := &mockTraceExporter{}
	exporter2 := &mockTraceExporter{exportErr: errors.New("failed")}
	exporter3 := &mockTraceExporter{}

	p.AddTraceExporter(exporter1)
	p.AddTraceExporter(exporter2)
	p.AddTraceExporter(exporter3)

	td := createTestTraces(1)
	err := p.ConsumeTraces(context.Background(), td)

	// Should get error from failing exporter
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// All exporters should have been called
	if !exporter1.exportCalled.Load() {
		t.Error("Exporter 1 should have been called")
	}
	if !exporter2.exportCalled.Load() {
		t.Error("Exporter 2 should have been called")
	}
	if !exporter3.exportCalled.Load() {
		t.Error("Exporter 3 should have been called")
	}
}

// TestConcurrentAccess tests concurrent access to the pipeline
func TestConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	p := New(logger)

	exporter := &mockTraceExporter{}
	p.AddTraceExporter(exporter)

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrently consume traces
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
	if stats.TracesProcessed != int64(numGoroutines) {
		t.Errorf("Expected %d traces processed, got %d", numGoroutines, stats.TracesProcessed)
	}
}

// TestConcurrentAddExporter tests concurrent exporter addition
func TestConcurrentAddExporter(t *testing.T) {
	logger := zap.NewNop()
	p := New(logger)

	var wg sync.WaitGroup
	numGoroutines := 50

	// Concurrently add exporters
	for i := 0; i < numGoroutines; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			p.AddTraceExporter(&mockTraceExporter{})
		}()
		go func() {
			defer wg.Done()
			p.AddMetricsExporter(&mockMetricsExporter{})
		}()
		go func() {
			defer wg.Done()
			p.AddLogsExporter(&mockLogsExporter{})
		}()
	}

	wg.Wait()

	if len(p.traceExporters) != numGoroutines {
		t.Errorf("Expected %d trace exporters, got %d", numGoroutines, len(p.traceExporters))
	}
	if len(p.metricsExporters) != numGoroutines {
		t.Errorf("Expected %d metrics exporters, got %d", numGoroutines, len(p.metricsExporters))
	}
	if len(p.logsExporters) != numGoroutines {
		t.Errorf("Expected %d logs exporters, got %d", numGoroutines, len(p.logsExporters))
	}
}

// TestPipelineStatsStruct tests PipelineStats struct fields
func TestPipelineStatsStruct(t *testing.T) {
	stats := PipelineStats{
		TracesProcessed:  100,
		MetricsProcessed: 200,
		LogsProcessed:    300,
		TracesDropped:    10,
		MetricsDropped:   20,
		LogsDropped:      30,
	}

	if stats.TracesProcessed != 100 {
		t.Errorf("Expected TracesProcessed 100, got %d", stats.TracesProcessed)
	}
	if stats.MetricsProcessed != 200 {
		t.Errorf("Expected MetricsProcessed 200, got %d", stats.MetricsProcessed)
	}
	if stats.LogsProcessed != 300 {
		t.Errorf("Expected LogsProcessed 300, got %d", stats.LogsProcessed)
	}
	if stats.TracesDropped != 10 {
		t.Errorf("Expected TracesDropped 10, got %d", stats.TracesDropped)
	}
	if stats.MetricsDropped != 20 {
		t.Errorf("Expected MetricsDropped 20, got %d", stats.MetricsDropped)
	}
	if stats.LogsDropped != 30 {
		t.Errorf("Expected LogsDropped 30, got %d", stats.LogsDropped)
	}
}

// Benchmark tests
func BenchmarkConsumeTraces(b *testing.B) {
	logger := zap.NewNop()
	p := New(logger)
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
	p := New(logger)
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
	p := New(logger)
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
	p := New(logger)
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
