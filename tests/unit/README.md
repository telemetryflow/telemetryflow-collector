# Unit Tests

Unit tests for the TelemetryFlow Collector.

## Overview

This directory contains unit tests for core packages, configuration, version information, and business logic. Unit tests should be isolated from external dependencies using mocks.

All tests use external test packages (`package <name>_test`) to ensure proper encapsulation and test the public API surface.

## Test Structure

```text
unit/
├── banner/        # Tests for banner generation and display
├── collector/     # Tests for collector core logic and OTel integration
├── config/        # Tests for internal configuration loading and validation
├── exporter/      # Tests for telemetry exporters (OTLP, etc.)
├── pipeline/      # Tests for data processing pipelines
├── pkg_config/    # Tests for pkg/config loader
├── plugin/        # Tests for plugin registry and management
├── receiver/      # Tests for OTLP receivers (gRPC/HTTP)
└── version/       # Tests for version package
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

| Package     | Target | Description                              |
|-------------|--------|------------------------------------------|
| banner      | 90%    | Banner generation and formatting         |
| collector   | 90%    | Core collector logic and OTel wrapper    |
| config      | 95%    | Configuration loading and validation     |
| exporter    | 90%    | OTLP and other exporters                 |
| pipeline    | 85%    | Data processing pipelines                |
| pkg_config  | 90%    | Public config loader                     |
| plugin      | 85%    | Plugin registry and lifecycle            |
| receiver    | 90%    | OTLP gRPC/HTTP receivers                 |
| version     | 100%   | Version information                      |

## Test Naming Convention

- Test files: `*_test.go`
- Test functions: `TestFunctionName`
- Subtests: `t.Run("should do something", func(t *testing.T) {})`

## Example Test

```go
package receiver_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/telemetryflow/telemetryflow-collector/internal/receiver/otlp"
)

func TestOTLPReceiverConfig(t *testing.T) {
    t.Run("should create receiver with default config", func(t *testing.T) {
        cfg := otlp.DefaultConfig()

        require.NotNil(t, cfg)
        assert.True(t, cfg.GRPC.Enabled)
        assert.Equal(t, ":4317", cfg.GRPC.Endpoint)
        assert.True(t, cfg.HTTP.Enabled)
        assert.Equal(t, ":4318", cfg.HTTP.Endpoint)
    })

    t.Run("should validate config", func(t *testing.T) {
        cfg := otlp.DefaultConfig()
        cfg.GRPC.Enabled = false
        cfg.HTTP.Enabled = false

        err := cfg.Validate()
        assert.Error(t, err, "at least one protocol must be enabled")
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
