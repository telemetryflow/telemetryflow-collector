# Test Mocks

Mock implementations for testing the TelemetryFlow Collector.

## Overview

This directory contains mock implementations of interfaces and external dependencies for use in unit and integration tests.

## Available Mocks

```text
mocks/
├── receiver.go         # Mock OTLP receiver
├── processor.go        # Mock data processor
├── exporter.go         # Mock data exporter
└── logger.go           # Mock logger
```

## Usage

```go
package mytest

import (
    "testing"
    "github.com/telemetryflow/telemetryflow-collector/tests/mocks"
)

func TestWithMock(t *testing.T) {
    mockExporter := mocks.NewMockExporter()
    mockExporter.On("Export", mock.Anything, mock.Anything).Return(nil)

    // Use mockExporter in your test
    // ...

    mockExporter.AssertExpectations(t)
}
```

## Mock Generation

Mocks are generated using `mockery`:

```bash
# Install mockery
go install github.com/vektra/mockery/v2@latest

# Generate mocks
mockery --name=Receiver --dir=internal/receiver --output=tests/mocks
mockery --name=Exporter --dir=internal/exporter --output=tests/mocks
```

## Best Practices

1. **Use interfaces**: Mock interfaces, not concrete types
2. **Keep mocks simple**: Only mock what's needed for the test
3. **Verify expectations**: Use `AssertExpectations` to verify mock calls
4. **Reset between tests**: Reset mock state between test cases
5. **Document behavior**: Comment mock methods explaining their behavior

## Mock Types

### MockReceiver

```go
type MockReceiver struct {
    mock.Mock
}

func (m *MockReceiver) Start(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}

func (m *MockReceiver) Stop() error {
    args := m.Called()
    return args.Error(0)
}
```

### MockExporter

```go
type MockExporter struct {
    mock.Mock
}

func (m *MockExporter) Export(ctx context.Context, data interface{}) error {
    args := m.Called(ctx, data)
    return args.Error(0)
}
```

## References

- [Testify Mock](https://github.com/stretchr/testify#mock-package)
- [Mockery](https://github.com/vektra/mockery)
