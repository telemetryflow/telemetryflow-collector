// Package mocks provides mock implementations for testing.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// MockProcessor is a mock implementation of the processor interface
type MockProcessor struct {
	mock.Mock
	running bool
}

// NewMockProcessor creates a new mock processor
func NewMockProcessor() *MockProcessor {
	return &MockProcessor{}
}

// Start mocks starting the processor
func (m *MockProcessor) Start(ctx context.Context, host component.Host) error {
	args := m.Called(ctx, host)
	m.running = true
	return args.Error(0)
}

// Shutdown mocks shutting down the processor
func (m *MockProcessor) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	m.running = false
	return args.Error(0)
}

// IsRunning returns whether the processor is running
func (m *MockProcessor) IsRunning() bool {
	return m.running
}

// MockMetricsProcessor is a mock implementation of processor.Metrics
type MockMetricsProcessor struct {
	mock.Mock
	Processed []pmetric.Metrics
	running   bool
}

// NewMockMetricsProcessor creates a new mock metrics processor
func NewMockMetricsProcessor() *MockMetricsProcessor {
	return &MockMetricsProcessor{
		Processed: make([]pmetric.Metrics, 0),
	}
}

// Start starts the processor
func (m *MockMetricsProcessor) Start(ctx context.Context, host component.Host) error {
	args := m.Called(ctx, host)
	m.running = true
	return args.Error(0)
}

// Shutdown shuts down the processor
func (m *MockMetricsProcessor) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	m.running = false
	return args.Error(0)
}

// ConsumeMetrics processes metrics
func (m *MockMetricsProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	m.Processed = append(m.Processed, md)
	args := m.Called(ctx, md)
	return args.Error(0)
}

// MockTracesProcessor is a mock implementation of processor.Traces
type MockTracesProcessor struct {
	mock.Mock
	Processed []ptrace.Traces
	running   bool
}

// NewMockTracesProcessor creates a new mock traces processor
func NewMockTracesProcessor() *MockTracesProcessor {
	return &MockTracesProcessor{
		Processed: make([]ptrace.Traces, 0),
	}
}

// Start starts the processor
func (m *MockTracesProcessor) Start(ctx context.Context, host component.Host) error {
	args := m.Called(ctx, host)
	m.running = true
	return args.Error(0)
}

// Shutdown shuts down the processor
func (m *MockTracesProcessor) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	m.running = false
	return args.Error(0)
}

// ConsumeTraces processes traces
func (m *MockTracesProcessor) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	m.Processed = append(m.Processed, td)
	args := m.Called(ctx, td)
	return args.Error(0)
}

// MockLogsProcessor is a mock implementation of processor.Logs
type MockLogsProcessor struct {
	mock.Mock
	Processed []plog.Logs
	running   bool
}

// NewMockLogsProcessor creates a new mock logs processor
func NewMockLogsProcessor() *MockLogsProcessor {
	return &MockLogsProcessor{
		Processed: make([]plog.Logs, 0),
	}
}

// Start starts the processor
func (m *MockLogsProcessor) Start(ctx context.Context, host component.Host) error {
	args := m.Called(ctx, host)
	m.running = true
	return args.Error(0)
}

// Shutdown shuts down the processor
func (m *MockLogsProcessor) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	m.running = false
	return args.Error(0)
}

// ConsumeLogs processes logs
func (m *MockLogsProcessor) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	m.Processed = append(m.Processed, ld)
	args := m.Called(ctx, ld)
	return args.Error(0)
}
