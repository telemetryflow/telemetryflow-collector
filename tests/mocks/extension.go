// Package mocks provides mock implementations for testing.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/collector/component"
)

// MockExtension is a mock implementation of the extension interface
type MockExtension struct {
	mock.Mock
	running bool
}

// NewMockExtension creates a new mock extension
func NewMockExtension() *MockExtension {
	return &MockExtension{}
}

// Start mocks starting the extension
func (m *MockExtension) Start(ctx context.Context, host component.Host) error {
	args := m.Called(ctx, host)
	m.running = true
	return args.Error(0)
}

// Shutdown mocks shutting down the extension
func (m *MockExtension) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	m.running = false
	return args.Error(0)
}

// IsRunning returns whether the extension is running
func (m *MockExtension) IsRunning() bool {
	return m.running
}

// MockAuthExtension is a mock implementation of auth extension
type MockAuthExtension struct {
	mock.Mock
	running bool
}

// NewMockAuthExtension creates a new mock auth extension
func NewMockAuthExtension() *MockAuthExtension {
	return &MockAuthExtension{}
}

// Start starts the extension
func (m *MockAuthExtension) Start(ctx context.Context, host component.Host) error {
	args := m.Called(ctx, host)
	m.running = true
	return args.Error(0)
}

// Shutdown shuts down the extension
func (m *MockAuthExtension) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	m.running = false
	return args.Error(0)
}

// Authenticate authenticates a request
func (m *MockAuthExtension) Authenticate(ctx context.Context, headers map[string][]string) (context.Context, error) {
	args := m.Called(ctx, headers)
	return args.Get(0).(context.Context), args.Error(1)
}

// MockHealthCheckExtension is a mock implementation of health check extension
type MockHealthCheckExtension struct {
	mock.Mock
	running bool
	healthy bool
}

// NewMockHealthCheckExtension creates a new mock health check extension
func NewMockHealthCheckExtension() *MockHealthCheckExtension {
	return &MockHealthCheckExtension{
		healthy: true,
	}
}

// Start starts the extension
func (m *MockHealthCheckExtension) Start(ctx context.Context, host component.Host) error {
	args := m.Called(ctx, host)
	m.running = true
	return args.Error(0)
}

// Shutdown shuts down the extension
func (m *MockHealthCheckExtension) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	m.running = false
	return args.Error(0)
}

// IsHealthy returns the health status
func (m *MockHealthCheckExtension) IsHealthy() bool {
	return m.healthy
}

// SetHealthy sets the health status
func (m *MockHealthCheckExtension) SetHealthy(healthy bool) {
	m.healthy = healthy
}
