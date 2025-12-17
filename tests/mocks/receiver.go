// Package mocks provides mock implementations for testing.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// TelemetryData represents generic telemetry data
type TelemetryData struct {
	Type       string                 `json:"type"` // metrics, logs, traces
	Resource   map[string]interface{} `json:"resource"`
	Scope      map[string]interface{} `json:"scope"`
	Data       []interface{}          `json:"data"`
	ReceivedAt time.Time              `json:"received_at"`
}

// MockReceiver is a mock implementation of the Receiver interface
type MockReceiver struct {
	mock.Mock
	name       string
	running    bool
	dataChan   chan *TelemetryData
}

// NewMockReceiver creates a new mock receiver
func NewMockReceiver(name string) *MockReceiver {
	return &MockReceiver{
		name:     name,
		dataChan: make(chan *TelemetryData, 100),
	}
}

// Name returns the receiver name
func (m *MockReceiver) Name() string {
	return m.name
}

// Start mocks starting the receiver
func (m *MockReceiver) Start(ctx context.Context) error {
	args := m.Called(ctx)
	m.running = true
	return args.Error(0)
}

// Stop mocks stopping the receiver
func (m *MockReceiver) Stop() error {
	args := m.Called()
	m.running = false
	close(m.dataChan)
	return args.Error(0)
}

// IsRunning returns whether the receiver is running
func (m *MockReceiver) IsRunning() bool {
	return m.running
}

// DataChannel returns the channel for received data
func (m *MockReceiver) DataChannel() <-chan *TelemetryData {
	return m.dataChan
}

// SimulateReceive simulates receiving telemetry data
func (m *MockReceiver) SimulateReceive(data *TelemetryData) {
	if m.running {
		data.ReceivedAt = time.Now()
		m.dataChan <- data
	}
}

// MockOTLPMetrics returns mock OTLP metrics for testing
func MockOTLPMetrics() *TelemetryData {
	return &TelemetryData{
		Type: "metrics",
		Resource: map[string]interface{}{
			"attributes": []map[string]interface{}{
				{"key": "host.name", "value": map[string]interface{}{"stringValue": "test-host"}},
				{"key": "service.name", "value": map[string]interface{}{"stringValue": "test-service"}},
			},
		},
		Scope: map[string]interface{}{
			"name":    "test-scope",
			"version": "1.0.0",
		},
		Data: []interface{}{
			map[string]interface{}{
				"name":        "system.cpu.usage",
				"unit":        "percent",
				"description": "CPU usage percentage",
				"gauge": map[string]interface{}{
					"dataPoints": []map[string]interface{}{
						{
							"asDouble":      45.5,
							"timeUnixNano": time.Now().UnixNano(),
						},
					},
				},
			},
		},
		ReceivedAt: time.Now(),
	}
}

// MockOTLPLogs returns mock OTLP logs for testing
func MockOTLPLogs() *TelemetryData {
	return &TelemetryData{
		Type: "logs",
		Resource: map[string]interface{}{
			"attributes": []map[string]interface{}{
				{"key": "host.name", "value": map[string]interface{}{"stringValue": "test-host"}},
				{"key": "service.name", "value": map[string]interface{}{"stringValue": "test-service"}},
			},
		},
		Scope: map[string]interface{}{
			"name":    "test-logger",
			"version": "1.0.0",
		},
		Data: []interface{}{
			map[string]interface{}{
				"timeUnixNano":         time.Now().UnixNano(),
				"severityNumber":       9, // INFO
				"severityText":         "INFO",
				"body":                 map[string]interface{}{"stringValue": "Test log message"},
				"traceId":              "",
				"spanId":               "",
			},
		},
		ReceivedAt: time.Now(),
	}
}

// MockOTLPTraces returns mock OTLP traces for testing
func MockOTLPTraces() *TelemetryData {
	return &TelemetryData{
		Type: "traces",
		Resource: map[string]interface{}{
			"attributes": []map[string]interface{}{
				{"key": "host.name", "value": map[string]interface{}{"stringValue": "test-host"}},
				{"key": "service.name", "value": map[string]interface{}{"stringValue": "test-service"}},
			},
		},
		Scope: map[string]interface{}{
			"name":    "test-tracer",
			"version": "1.0.0",
		},
		Data: []interface{}{
			map[string]interface{}{
				"traceId":           "0102030405060708090a0b0c0d0e0f10",
				"spanId":            "0102030405060708",
				"parentSpanId":      "",
				"name":              "test-span",
				"kind":              1, // INTERNAL
				"startTimeUnixNano": time.Now().Add(-time.Second).UnixNano(),
				"endTimeUnixNano":   time.Now().UnixNano(),
				"status": map[string]interface{}{
					"code":    1, // OK
					"message": "",
				},
			},
		},
		ReceivedAt: time.Now(),
	}
}
