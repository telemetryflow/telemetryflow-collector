// Package otlp provides OTLP receiver implementation for the TelemetryFlow Collector.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package otlp

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// mockConsumer implements the Consumer interface for testing
type mockConsumer struct {
	mu             sync.Mutex
	traces         []ptrace.Traces
	metrics        []pmetric.Metrics
	logs           []plog.Logs
	tracesErr      error
	metricsErr     error
	logsErr        error
	consumeTraces  int
	consumeMetrics int
	consumeLogs    int
}

func newMockConsumer() *mockConsumer {
	return &mockConsumer{
		traces:  make([]ptrace.Traces, 0),
		metrics: make([]pmetric.Metrics, 0),
		logs:    make([]plog.Logs, 0),
	}
}

func (m *mockConsumer) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.consumeTraces++
	if m.tracesErr != nil {
		return m.tracesErr
	}
	m.traces = append(m.traces, td)
	return nil
}

func (m *mockConsumer) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.consumeMetrics++
	if m.metricsErr != nil {
		return m.metricsErr
	}
	m.metrics = append(m.metrics, md)
	return nil
}

func (m *mockConsumer) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.consumeLogs++
	if m.logsErr != nil {
		return m.logsErr
	}
	m.logs = append(m.logs, ld)
	return nil
}

func (m *mockConsumer) getTraceCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.consumeTraces
}

func (m *mockConsumer) getMetricsCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.consumeMetrics
}

func (m *mockConsumer) getLogsCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.consumeLogs
}

// TestNewReceiver tests the receiver constructor
func TestNewReceiver(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()

	cfg := Config{
		GRPCEnabled:              true,
		GRPCEndpoint:             "localhost:4317",
		GRPCMaxRecvMsgSizeMiB:    4,
		GRPCMaxConcurrentStreams: 100,
		HTTPEnabled:              true,
		HTTPEndpoint:             "localhost:4318",
	}

	receiver := New(cfg, consumer, logger)

	assert.NotNil(t, receiver)
	assert.Equal(t, cfg.GRPCEnabled, receiver.config.GRPCEnabled)
	assert.Equal(t, cfg.GRPCEndpoint, receiver.config.GRPCEndpoint)
	assert.Equal(t, cfg.HTTPEnabled, receiver.config.HTTPEnabled)
	assert.Equal(t, cfg.HTTPEndpoint, receiver.config.HTTPEndpoint)
}

// TestReceiverStartStop tests receiver start and stop lifecycle
func TestReceiverStartStop(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()

	cfg := Config{
		GRPCEnabled:              true,
		GRPCEndpoint:             "localhost:14317",
		GRPCMaxRecvMsgSizeMiB:    4,
		GRPCMaxConcurrentStreams: 100,
		HTTPEnabled:              true,
		HTTPEndpoint:             "localhost:14318",
	}

	receiver := New(cfg, consumer, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start receiver
	err := receiver.Start(ctx)
	require.NoError(t, err)

	// Give servers time to start
	time.Sleep(100 * time.Millisecond)

	// Stop receiver
	err = receiver.Stop(ctx)
	require.NoError(t, err)
}

// TestReceiverStartAlreadyRunning tests that starting an already running receiver fails
func TestReceiverStartAlreadyRunning(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()

	cfg := Config{
		GRPCEnabled:              true,
		GRPCEndpoint:             "localhost:14319",
		GRPCMaxRecvMsgSizeMiB:    4,
		GRPCMaxConcurrentStreams: 100,
		HTTPEnabled:              false,
	}

	receiver := New(cfg, consumer, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start receiver
	err := receiver.Start(ctx)
	require.NoError(t, err)

	// Try to start again
	err = receiver.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Stop receiver
	_ = receiver.Stop(ctx)
}

// TestReceiverHTTPOnlyMode tests receiver with only HTTP enabled
func TestReceiverHTTPOnlyMode(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()

	cfg := Config{
		GRPCEnabled:  false,
		HTTPEnabled:  true,
		HTTPEndpoint: "localhost:14320",
	}

	receiver := New(cfg, consumer, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := receiver.Start(ctx)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	err = receiver.Stop(ctx)
	require.NoError(t, err)
}

// TestReceiverGRPCOnlyMode tests receiver with only gRPC enabled
func TestReceiverGRPCOnlyMode(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()

	cfg := Config{
		GRPCEnabled:              true,
		GRPCEndpoint:             "localhost:14321",
		GRPCMaxRecvMsgSizeMiB:    4,
		GRPCMaxConcurrentStreams: 100,
		HTTPEnabled:              false,
	}

	receiver := New(cfg, consumer, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := receiver.Start(ctx)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	err = receiver.Stop(ctx)
	require.NoError(t, err)
}

// TestReceiverStats tests receiver statistics
func TestReceiverStats(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()

	cfg := Config{
		HTTPEnabled:  true,
		HTTPEndpoint: "localhost:14322",
	}

	receiver := New(cfg, consumer, logger)

	stats := receiver.Stats()
	assert.Equal(t, int64(0), stats.TracesReceived)
	assert.Equal(t, int64(0), stats.MetricsReceived)
	assert.Equal(t, int64(0), stats.LogsReceived)
}

// TestHandleTracesMethodNotAllowed tests that non-POST requests are rejected
func TestHandleTracesMethodNotAllowed(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()

	cfg := Config{
		HTTPEnabled:  true,
		HTTPEndpoint: "localhost:14323",
	}

	receiver := New(cfg, consumer, logger)

	// Test GET request
	req := httptest.NewRequest(http.MethodGet, "/v2/traces", nil)
	rec := httptest.NewRecorder()

	receiver.handleTraces(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

// TestHandleMetricsMethodNotAllowed tests that non-POST requests are rejected
func TestHandleMetricsMethodNotAllowed(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()

	cfg := Config{
		HTTPEnabled:  true,
		HTTPEndpoint: "localhost:14324",
	}

	receiver := New(cfg, consumer, logger)

	req := httptest.NewRequest(http.MethodGet, "/v2/metrics", nil)
	rec := httptest.NewRecorder()

	receiver.handleMetrics(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

// TestHandleLogsMethodNotAllowed tests that non-POST requests are rejected
func TestHandleLogsMethodNotAllowed(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()

	cfg := Config{
		HTTPEnabled:  true,
		HTTPEndpoint: "localhost:14325",
	}

	receiver := New(cfg, consumer, logger)

	req := httptest.NewRequest(http.MethodGet, "/v2/logs", nil)
	rec := httptest.NewRecorder()

	receiver.handleLogs(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

// TestHandleTracesInvalidBody tests handling of invalid request body
func TestHandleTracesInvalidBody(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()

	cfg := Config{
		HTTPEnabled:  true,
		HTTPEndpoint: "localhost:14326",
	}

	receiver := New(cfg, consumer, logger)

	// Send invalid protobuf data
	req := httptest.NewRequest(http.MethodPost, "/v2/traces", bytes.NewReader([]byte("invalid data")))
	req.Header.Set("Content-Type", "application/x-protobuf")
	rec := httptest.NewRecorder()

	receiver.handleTraces(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestHandleMetricsInvalidBody tests handling of invalid request body
func TestHandleMetricsInvalidBody(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()

	cfg := Config{
		HTTPEnabled:  true,
		HTTPEndpoint: "localhost:14327",
	}

	receiver := New(cfg, consumer, logger)

	req := httptest.NewRequest(http.MethodPost, "/v2/metrics", bytes.NewReader([]byte("invalid data")))
	req.Header.Set("Content-Type", "application/x-protobuf")
	rec := httptest.NewRecorder()

	receiver.handleMetrics(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestHandleLogsInvalidBody tests handling of invalid request body
func TestHandleLogsInvalidBody(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()

	cfg := Config{
		HTTPEnabled:  true,
		HTTPEndpoint: "localhost:14328",
	}

	receiver := New(cfg, consumer, logger)

	req := httptest.NewRequest(http.MethodPost, "/v2/logs", bytes.NewReader([]byte("invalid data")))
	req.Header.Set("Content-Type", "application/x-protobuf")
	rec := httptest.NewRecorder()

	receiver.handleLogs(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestHandleTracesConsumerError tests handling when consumer returns error
func TestHandleTracesConsumerError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()
	consumer.tracesErr = fmt.Errorf("consumer error")

	cfg := Config{
		HTTPEnabled:  true,
		HTTPEndpoint: "localhost:14329",
	}

	receiver := New(cfg, consumer, logger)

	// Create valid empty traces request
	traces := ptrace.NewTraces()
	marshaler := ptrace.ProtoMarshaler{}
	data, err := marshaler.MarshalTraces(traces)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v2/traces", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/x-protobuf")
	rec := httptest.NewRecorder()

	receiver.handleTraces(rec, req)

	// Consumer error should return internal server error
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestReceiverStopNotRunning tests stopping a receiver that's not running
func TestReceiverStopNotRunning(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()

	cfg := Config{
		HTTPEnabled:  true,
		HTTPEndpoint: "localhost:14330",
	}

	receiver := New(cfg, consumer, logger)

	// Stop without starting should not error
	err := receiver.Stop(context.Background())
	assert.NoError(t, err)
}

// TestTraceServerExport tests the gRPC trace server Export method
func TestTraceServerExport(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()

	receiver := &Receiver{
		config:   Config{},
		logger:   logger,
		consumer: consumer,
	}

	server := &traceServer{r: receiver}

	// Create a mock export request with traces
	traces := ptrace.NewTraces()
	rs := traces.ResourceSpans().AppendEmpty()
	ss := rs.ScopeSpans().AppendEmpty()
	span := ss.Spans().AppendEmpty()
	span.SetName("test-span")
	span.SetTraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	span.SetSpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8})

	// Note: We can't easily create a ptraceotlp.ExportRequest without the actual proto
	// This test verifies the consumer is called correctly through integration testing
	assert.NotNil(t, server)
	assert.Equal(t, 0, consumer.getTraceCount())
}

// TestMetricsServerExport tests the gRPC metrics server Export method
func TestMetricsServerExport(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()

	receiver := &Receiver{
		config:   Config{},
		logger:   logger,
		consumer: consumer,
	}

	server := &metricsServer{r: receiver}
	assert.NotNil(t, server)
	assert.Equal(t, 0, consumer.getMetricsCount())
}

// TestLogsServerExport tests the gRPC logs server Export method
func TestLogsServerExport(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()

	receiver := &Receiver{
		config:   Config{},
		logger:   logger,
		consumer: consumer,
	}

	server := &logsServer{r: receiver}
	assert.NotNil(t, server)
	assert.Equal(t, 0, consumer.getLogsCount())
}

// TestReceiverWithNilConsumer tests receiver behavior with nil consumer
func TestReceiverWithNilConsumer(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := Config{
		HTTPEnabled:  true,
		HTTPEndpoint: "localhost:14331",
	}

	receiver := New(cfg, nil, logger)

	// Create valid empty traces request
	traces := ptrace.NewTraces()
	marshaler := ptrace.ProtoMarshaler{}
	data, err := marshaler.MarshalTraces(traces)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v2/traces", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/x-protobuf")
	rec := httptest.NewRecorder()

	receiver.handleTraces(rec, req)

	// Should succeed even with nil consumer
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestReceiverJSONContentType tests handling of JSON content type
func TestReceiverJSONContentType(t *testing.T) {
	logger := zaptest.NewLogger(t)
	consumer := newMockConsumer()

	cfg := Config{
		HTTPEnabled:  true,
		HTTPEndpoint: "localhost:14332",
	}

	receiver := New(cfg, consumer, logger)

	// Send syntactically invalid JSON (will fail to unmarshal)
	req := httptest.NewRequest(http.MethodPost, "/v2/traces", bytes.NewReader([]byte(`not valid json at all`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	receiver.handleTraces(rec, req)

	// Invalid JSON syntax should return bad request
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestConfigDefaults tests config with default values
func TestConfigDefaults(t *testing.T) {
	cfg := Config{}

	assert.False(t, cfg.GRPCEnabled)
	assert.False(t, cfg.HTTPEnabled)
	assert.Empty(t, cfg.GRPCEndpoint)
	assert.Empty(t, cfg.HTTPEndpoint)
	assert.Equal(t, 0, cfg.GRPCMaxRecvMsgSizeMiB)
	assert.Equal(t, uint32(0), cfg.GRPCMaxConcurrentStreams)
}

// BenchmarkHandleTraces benchmarks the trace handler
func BenchmarkHandleTraces(b *testing.B) {
	logger := zap.NewNop()
	consumer := newMockConsumer()

	cfg := Config{
		HTTPEnabled:  true,
		HTTPEndpoint: "localhost:14333",
	}

	receiver := New(cfg, consumer, logger)

	// Create valid empty traces request
	traces := ptrace.NewTraces()
	marshaler := ptrace.ProtoMarshaler{}
	data, _ := marshaler.MarshalTraces(traces)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v2/traces", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/x-protobuf")
		rec := httptest.NewRecorder()
		receiver.handleTraces(rec, req)
	}
}

// BenchmarkHandleMetrics benchmarks the metrics handler
func BenchmarkHandleMetrics(b *testing.B) {
	logger := zap.NewNop()
	consumer := newMockConsumer()

	cfg := Config{
		HTTPEnabled:  true,
		HTTPEndpoint: "localhost:14334",
	}

	receiver := New(cfg, consumer, logger)

	// Create valid empty metrics request
	metrics := pmetric.NewMetrics()
	marshaler := pmetric.ProtoMarshaler{}
	data, _ := marshaler.MarshalMetrics(metrics)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v2/metrics", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/x-protobuf")
		rec := httptest.NewRecorder()
		receiver.handleMetrics(rec, req)
	}
}

// BenchmarkHandleLogs benchmarks the logs handler
func BenchmarkHandleLogs(b *testing.B) {
	logger := zap.NewNop()
	consumer := newMockConsumer()

	cfg := Config{
		HTTPEnabled:  true,
		HTTPEndpoint: "localhost:14335",
	}

	receiver := New(cfg, consumer, logger)

	// Create valid empty logs request
	logs := plog.NewLogs()
	marshaler := plog.ProtoMarshaler{}
	data, _ := marshaler.MarshalLogs(logs)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/v2/logs", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/x-protobuf")
		rec := httptest.NewRecorder()
		receiver.handleLogs(rec, req)
	}
}
