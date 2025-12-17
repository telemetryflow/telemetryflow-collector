# Unit Tests

Unit tests for the TelemetryFlow Collector.

## Overview

This directory contains unit tests for core packages, configuration, version information, and business logic. Unit tests should be isolated from external dependencies using mocks.

## Test Structure

```text
unit/
├── config/        # Tests for configuration loading and validation
├── version/       # Tests for version package
├── collector/     # Tests for collector core logic
└── receiver/      # Tests for OTLP receivers
```

## Running Tests

```bash
# Run all unit tests
go test ./tests/unit/...

# Run specific package tests
go test ./tests/unit/config/...

# Run with verbose output
go test -v ./tests/unit/...

# Run with coverage
go test -cover ./tests/unit/...

# Run with coverage report
go test -coverprofile=coverage.out ./tests/unit/...
go tool cover -html=coverage.out -o coverage.html
```

## Coverage Targets

- **Config**: 95% coverage
- **Version**: 100% coverage
- **Collector**: 90% coverage
- **Receiver**: 90% coverage

## Test Naming Convention

- Test files: `*_test.go`
- Test functions: `TestFunctionName`
- Subtests: `t.Run("should do something", func(t *testing.T) {})`

## Example Test

```go
package config_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/telemetryflow/telemetryflow-collector/internal/config"
)

func TestReceiverConfig(t *testing.T) {
    t.Run("should have OTLP enabled by default", func(t *testing.T) {
        cfg := config.DefaultConfig()

        assert.True(t, cfg.Receivers.OTLP.Enabled)
        assert.True(t, cfg.Receivers.OTLP.Protocols.GRPC.Enabled)
        assert.True(t, cfg.Receivers.OTLP.Protocols.HTTP.Enabled)
    })

    t.Run("should fail validation with no receivers", func(t *testing.T) {
        cfg := config.DefaultConfig()
        cfg.Receivers.OTLP.Enabled = false
        cfg.Receivers.Prometheus.Enabled = false

        err := cfg.Validate()
        assert.Error(t, err)
    })
}
```

## Best Practices

1. **Test in isolation**: Use mocks for all external dependencies
2. **Table-driven tests**: Use table-driven tests for multiple scenarios
3. **Test error paths**: Cover both success and failure scenarios
4. **Use testify**: Use `github.com/stretchr/testify` for assertions
5. **Use mocks**: Import from `tests/mocks/` for mock implementations
6. **Use fixtures**: Import from `tests/fixtures/` for test data

## References

- [Testing Documentation](../../docs/TESTING.md)
- [Test Fixtures](../fixtures/)
- [Test Mocks](../mocks/)
