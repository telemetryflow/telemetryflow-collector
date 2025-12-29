# Testing Documentation

Testing guide for the TelemetryFlow Collector.

- **Version:** 1.1.1
- **Last Updated:** December 2025

---

## Overview

TelemetryFlow Collector uses a comprehensive testing strategy with three levels of testing:

1. **Unit Tests** - Test individual components in isolation
2. **Integration Tests** - Test component interactions
3. **End-to-End Tests** - Test complete data flows

All tests follow Domain-Driven Design (DDD) patterns for organization.

---

## Test Structure

```text
tests/
├── unit/                   # Unit tests (DDD organized)
│   ├── domain/             # Core business logic tests
│   │   ├── collector/
│   │   ├── pipeline/
│   │   └── plugin/
│   ├── application/        # Use case tests
│   │   ├── collector/
│   │   └── config/
│   ├── infrastructure/     # External adapter tests
│   │   ├── exporter/
│   │   ├── pkg_config/
│   │   └── receiver/
│   └── presentation/       # UI/Output tests
│       ├── banner/
│       └── version/
├── integration/            # Integration tests
│   ├── collector/
│   └── exporter/
├── e2e/                    # End-to-end tests
│   ├── pipeline_test.go
│   ├── receiver_test.go
│   └── startup_test.go
├── fixtures/               # Test data and fixtures
└── mocks/                  # Mock implementations
```

---

## Running Tests

### All Tests

```bash
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run with race detection
go test -race ./...

# Run with coverage
go test -cover ./...
```

### Unit Tests

```bash
# All unit tests
go test ./tests/unit/...

# By DDD layer
go test ./tests/unit/domain/...
go test ./tests/unit/application/...
go test ./tests/unit/infrastructure/...
go test ./tests/unit/presentation/...

# Specific package
go test ./tests/unit/domain/pipeline/...
```

### Integration Tests

```bash
# All integration tests
go test ./tests/integration/...

# Specific integration test
go test ./tests/integration/collector/...
```

### End-to-End Tests

```bash
# All e2e tests
go test ./tests/e2e/...

# Run with timeout (e2e tests may take longer)
go test -timeout 5m ./tests/e2e/...
```

---

## Coverage

### Generate Coverage Report

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Open HTML report (macOS)
open coverage.html
```

### Coverage Targets

| Layer          | Package      | Target | Description                    |
|----------------|--------------|--------|--------------------------------|
| Domain         | collector    | 90%    | Core collector logic           |
| Domain         | pipeline     | 85%    | Data processing pipelines      |
| Domain         | plugin       | 85%    | Plugin registry                |
| Application    | collector    | 90%    | OTEL collector integration     |
| Application    | config       | 95%    | Configuration management       |
| Infrastructure | exporter     | 90%    | Data exporters                 |
| Infrastructure | receiver     | 90%    | Data receivers                 |
| Infrastructure | pkg_config   | 90%    | Config loader utilities        |
| Presentation   | banner       | 90%    | Banner display                 |
| Presentation   | version      | 100%   | Version information            |

---

## Writing Tests

### Test File Naming

- Unit tests: `*_test.go` in `tests/unit/<layer>/<package>/`
- Integration tests: `*_test.go` in `tests/integration/<package>/`
- E2E tests: `*_test.go` in `tests/e2e/`

### Test Package Pattern

Use external test packages for black-box testing:

```go
// Good: External test package
package pipeline_test

import (
    "github.com/telemetryflow/telemetryflow-collector/internal/pipeline"
)

// Bad: Same package (white-box testing)
package pipeline
```

### Test Function Naming

```go
// Test function: TestFunctionName
func TestPipelineProcess(t *testing.T) {
    // Subtests: t.Run("should do something", ...)
    t.Run("should process trace data", func(t *testing.T) {
        // Test implementation
    })

    t.Run("should handle empty input", func(t *testing.T) {
        // Test implementation
    })
}
```

### Example Unit Test

```go
package pipeline_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/telemetryflow/telemetryflow-collector/internal/pipeline"
)

func TestPipeline(t *testing.T) {
    t.Run("should create pipeline with valid config", func(t *testing.T) {
        // Given
        cfg := pipeline.Config{
            TracesEnabled:  true,
            MetricsEnabled: true,
            LogsEnabled:    true,
        }

        // When
        p, err := pipeline.New(cfg)

        // Then
        require.NoError(t, err)
        require.NotNil(t, p)
        assert.True(t, p.TracesEnabled())
    })

    t.Run("should return error for invalid config", func(t *testing.T) {
        // Given
        cfg := pipeline.Config{} // All disabled

        // When
        _, err := pipeline.New(cfg)

        // Then
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "at least one signal must be enabled")
    })
}
```

### Table-Driven Tests

```go
func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  Config
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid config",
            config:  DefaultConfig(),
            wantErr: false,
        },
        {
            name: "missing endpoint",
            config: Config{
                Endpoint: "",
            },
            wantErr: true,
            errMsg:  "endpoint is required",
        },
        {
            name: "invalid port",
            config: Config{
                Endpoint: "localhost:999999",
            },
            wantErr: true,
            errMsg:  "invalid port",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()

            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

---

## Using Mocks

### Import Mocks

```go
import (
    "github.com/telemetryflow/telemetryflow-collector/tests/mocks"
)
```

### Example with Mock

```go
func TestCollectorWithMockExporter(t *testing.T) {
    t.Run("should export data successfully", func(t *testing.T) {
        // Given
        mockExporter := mocks.NewMockExporter()
        mockExporter.On("Export", mock.Anything, mock.Anything).Return(nil)

        collector := NewCollector(WithExporter(mockExporter))

        // When
        err := collector.Process(testData)

        // Then
        require.NoError(t, err)
        mockExporter.AssertExpectations(t)
        mockExporter.AssertCalled(t, "Export", mock.Anything, mock.Anything)
    })
}
```

---

## Using Fixtures

### Load Test Fixtures

```go
import (
    "github.com/telemetryflow/telemetryflow-collector/tests/fixtures"
)

func TestWithFixture(t *testing.T) {
    // Load YAML fixture
    config := fixtures.LoadConfig(t, "valid-config.yaml")

    // Load JSON fixture
    traces := fixtures.LoadTraces(t, "sample-traces.json")

    // Use fixtures in test
    // ...
}
```

---

## Benchmark Tests

### Writing Benchmarks

```go
func BenchmarkPipelineProcess(b *testing.B) {
    pipeline := NewPipeline(DefaultConfig())
    data := generateTestData(1000)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = pipeline.Process(data)
    }
}

func BenchmarkPipelineProcessParallel(b *testing.B) {
    pipeline := NewPipeline(DefaultConfig())
    data := generateTestData(1000)

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _ = pipeline.Process(data)
        }
    })
}
```

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkPipelineProcess ./tests/unit/domain/pipeline/

# Run with memory allocation stats
go test -bench=. -benchmem ./...

# Run multiple times for accurate results
go test -bench=. -count=5 ./...
```

---

## CI/CD Integration

### GitHub Actions

Tests are automatically run on every push and pull request:

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - name: Run tests
        run: make test
      - name: Run coverage
        run: go test -coverprofile=coverage.out ./...
      - name: Upload coverage
        uses: codecov/codecov-action@v4
```

---

## Best Practices

1. **Test in isolation**: Use mocks for external dependencies
2. **Use table-driven tests**: For multiple test cases
3. **Test error paths**: Cover both success and failure scenarios
4. **Use testify**: Prefer `assert` and `require` from testify
5. **Follow DDD boundaries**: Keep tests within their architectural layer
6. **Use external packages**: `package <name>_test` for black-box testing
7. **Name tests descriptively**: `t.Run("should do something", ...)`
8. **Keep tests fast**: Unit tests should complete in milliseconds
9. **Clean up resources**: Use `t.Cleanup()` for teardown
10. **Avoid test interdependence**: Each test should be independent

---

## Troubleshooting

### Common Issues

**Tests timeout:**
```bash
# Increase timeout
go test -timeout 10m ./...
```

**Race conditions:**
```bash
# Detect races
go test -race ./...
```

**Flaky tests:**
```bash
# Run multiple times to detect flakiness
go test -count=10 ./...
```

**Coverage not accurate:**
```bash
# Include all packages
go test -coverpkg=./... -coverprofile=coverage.out ./...
```

---

## References

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify](https://github.com/stretchr/testify)
- [Test Mocks](../tests/mocks/)
- [Test Fixtures](../tests/fixtures/)
- [DDD Architecture Guide](ARCHITECTURE.md)

---

**Copyright (c) 2024-2026 DevOpsCorner Indonesia. All rights reserved.**
