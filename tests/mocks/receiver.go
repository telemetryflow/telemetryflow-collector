// Package mocks provides mock implementations for testing.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/receiver"
)

// MockReceiver is a mock implementation of the receiver interface
type MockReceiver struct {
	mock.Mock
	running bool
}

// NewMockReceiver creates a new mock receiver
func NewMockReceiver() *MockReceiver {
	return &MockReceiver{}
}

// Start mocks starting the receiver
func (m *MockReceiver) Start(ctx context.Context, host component.Host) error {
	args := m.Called(ctx, host)
	m.running = true
	return args.Error(0)
}

// Shutdown mocks shutting down the receiver
func (m *MockReceiver) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	m.running = false
	return args.Error(0)
}

// IsRunning returns whether the receiver is running
func (m *MockReceiver) IsRunning() bool {
	return m.running
}

// MockReceiverFactory is a mock implementation of receiver.Factory
type MockReceiverFactory struct {
	mock.Mock
}

// Type returns the receiver type
func (m *MockReceiverFactory) Type() component.Type {
	args := m.Called()
	return args.Get(0).(component.Type)
}

// CreateDefaultConfig creates default config
func (m *MockReceiverFactory) CreateDefaultConfig() component.Config {
	args := m.Called()
	return args.Get(0).(component.Config)
}

// CreateTracesReceiver creates a traces receiver
func (m *MockReceiverFactory) CreateTracesReceiver(
	ctx context.Context,
	set receiver.Settings,
	cfg component.Config,
	consumer consumer.Traces,
) (receiver.Traces, error) {
	args := m.Called(ctx, set, cfg, consumer)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(receiver.Traces), args.Error(1)
}

// CreateMetricsReceiver creates a metrics receiver
func (m *MockReceiverFactory) CreateMetricsReceiver(
	ctx context.Context,
	set receiver.Settings,
	cfg component.Config,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	args := m.Called(ctx, set, cfg, consumer)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(receiver.Metrics), args.Error(1)
}

// CreateLogsReceiver creates a logs receiver
func (m *MockReceiverFactory) CreateLogsReceiver(
	ctx context.Context,
	set receiver.Settings,
	cfg component.Config,
	consumer consumer.Logs,
) (receiver.Logs, error) {
	args := m.Called(ctx, set, cfg, consumer)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(receiver.Logs), args.Error(1)
}

// MockMetricsConsumer is a mock implementation of consumer.Metrics
type MockMetricsConsumer struct {
	mock.Mock
	Received []pmetric.Metrics
}

// NewMockMetricsConsumer creates a new mock metrics consumer
func NewMockMetricsConsumer() *MockMetricsConsumer {
	return &MockMetricsConsumer{
		Received: make([]pmetric.Metrics, 0),
	}
}

// ConsumeMetrics consumes metrics
func (m *MockMetricsConsumer) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	m.Received = append(m.Received, md)
	args := m.Called(ctx, md)
	return args.Error(0)
}

// Capabilities returns consumer capabilities
func (m *MockMetricsConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

// MockTracesConsumer is a mock implementation of consumer.Traces
type MockTracesConsumer struct {
	mock.Mock
	Received []ptrace.Traces
}

// NewMockTracesConsumer creates a new mock traces consumer
func NewMockTracesConsumer() *MockTracesConsumer {
	return &MockTracesConsumer{
		Received: make([]ptrace.Traces, 0),
	}
}

// ConsumeTraces consumes traces
func (m *MockTracesConsumer) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	m.Received = append(m.Received, td)
	args := m.Called(ctx, td)
	return args.Error(0)
}

// Capabilities returns consumer capabilities
func (m *MockTracesConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

// MockLogsConsumer is a mock implementation of consumer.Logs
type MockLogsConsumer struct {
	mock.Mock
	Received []plog.Logs
}

// NewMockLogsConsumer creates a new mock logs consumer
func NewMockLogsConsumer() *MockLogsConsumer {
	return &MockLogsConsumer{
		Received: make([]plog.Logs, 0),
	}
}

// ConsumeLogs consumes logs
func (m *MockLogsConsumer) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	m.Received = append(m.Received, ld)
	args := m.Called(ctx, ld)
	return args.Error(0)
}

// Capabilities returns consumer capabilities
func (m *MockLogsConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}
