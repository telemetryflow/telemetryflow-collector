# Unit Tests

Unit tests for the TelemetryFlow Collector, organized following Domain-Driven Design (DDD) patterns.

## Overview

This directory contains unit tests organized by DDD architectural layers. Each layer has specific responsibilities and tests are grouped accordingly to ensure proper separation of concerns.

All tests use external test packages (`package <name>_test`) to ensure proper encapsulation and test the public API surface.

## DDD Architecture

```text
unit/
├── domain/                 # Core business logic and entities
│   ├── collector/          # Collector lifecycle and operations
│   ├── pipeline/           # Data processing pipelines (traces, metrics, logs)
│   └── plugin/             # Plugin registry and management
│
├── application/            # Use cases and orchestration
│   ├── collector/          # OTEL collector integration and startup
│   └── config/             # Configuration loading and validation
│
├── infrastructure/         # External systems and adapters
│   ├── exporter/           # OTLP exporters (debug, file, etc.)
│   ├── pkg_config/         # Public config loader utilities
│   └── receiver/           # OTLP receivers (gRPC/HTTP)
│
└── presentation/           # User interface and output
    ├── banner/             # Banner generation and display
    └── version/            # Version information display
```

## DDD Layers Explained

### Domain Layer (`domain/`)

Core business logic independent of external systems:

- **collector/** - Collector lifecycle, core operations
- **pipeline/** - Data processing pipelines for traces, metrics, logs
- **plugin/** - Plugin registry, factory pattern, component management

### Application Layer (`application/`)

Orchestration and use cases:

- **collector/** - OTEL collector wrapper, integration with OpenTelemetry
- **config/** - Configuration loading, validation, environment variables

### Infrastructure Layer (`infrastructure/`)

External system adapters:

- **exporter/** - Debug exporter, OTLP exporter implementations
- **pkg_config/** - Public configuration loader utilities
- **receiver/** - OTLP gRPC/HTTP receiver implementations

### Presentation Layer (`presentation/`)

User-facing output:

- **banner/** - ASCII art banner generation, startup display
- **version/** - Version information, build info display

## Running Tests

```bash
# Run all unit tests
go test ./tests/unit/...

# Run by DDD layer
go test ./tests/unit/domain/...
go test ./tests/unit/application/...
go test ./tests/unit/infrastructure/...
go test ./tests/unit/presentation/...

# Run specific domain
go test ./tests/unit/domain/pipeline/...

# Run with verbose output
go test -v ./tests/unit/...

# Run with race detection
go test -race ./tests/unit/...

# Run with coverage
go test -cover ./tests/unit/...

# Run with coverage report
go test -coverprofile=coverage.out ./tests/unit/...
go tool cover -html=coverage.out -o coverage.html
```

## Coverage Targets by Layer

### Domain Layer

| Package   | Target | Description                              |
|-----------|--------|------------------------------------------|
| collector | 90%    | Collector lifecycle and operations       |
| pipeline  | 85%    | Data processing pipelines                |
| plugin    | 85%    | Plugin registry and management           |

### Application Layer

| Package   | Target | Description                              |
|-----------|--------|------------------------------------------|
| collector | 90%    | OTEL collector integration               |
| config    | 95%    | Configuration loading and validation     |

### Infrastructure Layer

| Package    | Target | Description                              |
|------------|--------|------------------------------------------|
| exporter   | 90%    | OTLP and debug exporters                 |
| pkg_config | 90%    | Public config loader                     |
| receiver   | 90%    | OTLP gRPC/HTTP receivers                 |

### Presentation Layer

| Package | Target | Description                               |
|---------|--------|-------------------------------------------|
| banner  | 90%    | Banner generation and display             |
| version | 100%   | Version information                       |

## Test Naming Convention

- Test files: `*_test.go`
- Test functions: `TestFunctionName`
- Subtests: `t.Run("should do something", func(t *testing.T) {})`

## Example Test (Domain Layer)

```go
package pipeline_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/telemetryflow/telemetryflow-collector/internal/pipeline"
)

func TestPipeline(t *testing.T) {
    t.Run("should create pipeline with valid config", func(t *testing.T) {
        cfg := pipeline.Config{
            TracesEnabled:  true,
            MetricsEnabled: true,
            LogsEnabled:    true,
        }

        p, err := pipeline.New(cfg)
        require.NoError(t, err)
        require.NotNil(t, p)
    })

    t.Run("should process trace data", func(t *testing.T) {
        p := pipeline.New(pipeline.DefaultConfig())

        err := p.ProcessTraces(testTraceData)
        require.NoError(t, err)
    })
}
```

## Best Practices

1. **Domain isolation**: Domain tests should not depend on infrastructure
2. **Mock external dependencies**: Use mocks from `tests/mocks/`
3. **Table-driven tests**: Use table-driven tests for multiple scenarios
4. **Test error paths**: Cover both success and failure scenarios
5. **Use testify**: Use `github.com/stretchr/testify` for assertions
6. **External packages**: Use `package <name>_test` pattern for black-box testing
7. **Follow DDD boundaries**: Keep tests within their architectural layer

## References

- [Testing Documentation](../../docs/TESTING.md)
- [Test Fixtures](../fixtures/)
- [Test Mocks](../mocks/)
- [DDD Architecture Guide](../../docs/ARCHITECTURE.md)
