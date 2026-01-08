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

// MockExporter is a mock implementation of the exporter interface
type MockExporter struct {
	mock.Mock
	running bool
}

// NewMockExporter creates a new mock exporter
func NewMockExporter() *MockExporter {
	return &MockExporter{}
}

// Start mocks starting the exporter
func (m *MockExporter) Start(ctx context.Context, host component.Host) error {
	args := m.Called(ctx, host)
	m.running = true
	return args.Error(0)
}

// Shutdown mocks shutting down the exporter
func (m *MockExporter) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	m.running = false
	return args.Error(0)
}

// IsRunning returns whether the exporter is running
func (m *MockExporter) IsRunning() bool {
	return m.running
}

// MockMetricsExporter is a mock implementation of exporter.Metrics
type MockMetricsExporter struct {
	mock.Mock
	Exported []pmetric.Metrics
	running  bool
}

// NewMockMetricsExporter creates a new mock metrics exporter
func NewMockMetricsExporter() *MockMetricsExporter {
	return &MockMetricsExporter{
		Exported: make([]pmetric.Metrics, 0),
	}
}

// Start starts the exporter
func (m *MockMetricsExporter) Start(ctx context.Context, host component.Host) error {
	args := m.Called(ctx, host)
	m.running = true
	return args.Error(0)
}

// Shutdown shuts down the exporter
func (m *MockMetricsExporter) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	m.running = false
	return args.Error(0)
}

// ConsumeMetrics exports metrics
func (m *MockMetricsExporter) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	m.Exported = append(m.Exported, md)
	args := m.Called(ctx, md)
	return args.Error(0)
}

// MockTracesExporter is a mock implementation of exporter.Traces
type MockTracesExporter struct {
	mock.Mock
	Exported []ptrace.Traces
	running  bool
}

// NewMockTracesExporter creates a new mock traces exporter
func NewMockTracesExporter() *MockTracesExporter {
	return &MockTracesExporter{
		Exported: make([]ptrace.Traces, 0),
	}
}

// Start starts the exporter
func (m *MockTracesExporter) Start(ctx context.Context, host component.Host) error {
	args := m.Called(ctx, host)
	m.running = true
	return args.Error(0)
}

// Shutdown shuts down the exporter
func (m *MockTracesExporter) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	m.running = false
	return args.Error(0)
}

// ConsumeTraces exports traces
func (m *MockTracesExporter) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	m.Exported = append(m.Exported, td)
	args := m.Called(ctx, td)
	return args.Error(0)
}

// MockLogsExporter is a mock implementation of exporter.Logs
type MockLogsExporter struct {
	mock.Mock
	Exported []plog.Logs
	running  bool
}

// NewMockLogsExporter creates a new mock logs exporter
func NewMockLogsExporter() *MockLogsExporter {
	return &MockLogsExporter{
		Exported: make([]plog.Logs, 0),
	}
}

// Start starts the exporter
func (m *MockLogsExporter) Start(ctx context.Context, host component.Host) error {
	args := m.Called(ctx, host)
	m.running = true
	return args.Error(0)
}

// Shutdown shuts down the exporter
func (m *MockLogsExporter) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	m.running = false
	return args.Error(0)
}

// ConsumeLogs exports logs
func (m *MockLogsExporter) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	m.Exported = append(m.Exported, ld)
	args := m.Called(ctx, ld)
	return args.Error(0)
}
