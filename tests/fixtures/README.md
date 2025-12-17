# Test Fixtures

Test fixtures and sample data for TelemetryFlow Collector tests.

## Overview

This directory contains test fixtures, sample configurations, and test data for use in unit, integration, and E2E tests.

## Available Fixtures

```text
fixtures/
├── configs/           # Sample configuration files
│   ├── valid.yaml     # Valid minimal configuration
│   ├── full.yaml      # Complete configuration with all options
│   └── invalid.yaml   # Invalid configuration for error testing
├── otlp/              # Sample OTLP data
│   ├── metrics.json   # Sample metrics in OTLP format
│   ├── logs.json      # Sample logs in OTLP format
│   └── traces.json    # Sample traces in OTLP format
└── responses/         # Mock API responses
    ├── health.json    # Health check response
    └── error.json     # Error response
```

## Usage

```go
package mytest

import (
    "os"
    "testing"
    "path/filepath"
)

func TestWithFixture(t *testing.T) {
    // Load fixture file
    fixturePath := filepath.Join("testdata", "fixtures", "configs", "valid.yaml")
    data, err := os.ReadFile(fixturePath)
    if err != nil {
        t.Fatalf("failed to load fixture: %v", err)
    }

    // Use fixture data in test
    // ...
}
```

## Fixture Categories

### Configuration Fixtures

Sample YAML configuration files for testing config loading:

```yaml
# fixtures/configs/valid.yaml
collector:
  id: "test-collector"
receivers:
  otlp:
    enabled: true
    protocols:
      grpc:
        enabled: true
        endpoint: "0.0.0.0:4317"
```

### OTLP Fixtures

Sample OTLP data for testing receivers and processors:

```json
{
  "resourceMetrics": [
    {
      "resource": {
        "attributes": [
          {"key": "host.name", "value": {"stringValue": "test-host"}}
        ]
      },
      "scopeMetrics": [
        {
          "metrics": [
            {
              "name": "system.cpu.usage",
              "unit": "percent",
              "gauge": {"dataPoints": [{"asDouble": 45.5}]}
            }
          ]
        }
      ]
    }
  ]
}
```

### Response Fixtures

Mock API responses for testing:

```json
{
  "status": "healthy",
  "components": {
    "receiver": "ok",
    "processor": "ok",
    "exporter": "ok"
  }
}
```

## Best Practices

1. **Keep fixtures minimal**: Include only necessary data
2. **Version fixtures**: Update fixtures when API changes
3. **Document fixtures**: Explain what each fixture tests
4. **Use realistic data**: Use production-like values where possible
5. **Separate by type**: Organize fixtures by their purpose

## References

- [Go Testing](https://golang.org/pkg/testing/)
- [OTLP Specification](https://opentelemetry.io/docs/specs/otlp/)
