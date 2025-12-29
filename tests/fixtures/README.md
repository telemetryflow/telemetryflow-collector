# Test Fixtures

Test fixtures and sample data for the TelemetryFlow Collector.

## Overview

This directory contains test fixtures - predefined data files used in unit and integration tests. Fixtures provide consistent, reproducible test data across test runs.

## Directory Structure

```text
fixtures/
├── config/                 # Configuration fixtures
│   ├── valid-config.yaml   # Valid configuration
│   ├── invalid-config.yaml # Invalid configuration for error tests
│   └── minimal-config.yaml # Minimal required configuration
│
├── telemetry/              # Telemetry data fixtures
│   ├── traces/             # Sample trace data
│   ├── metrics/            # Sample metrics data
│   └── logs/               # Sample log data
│
└── README.md               # This file
```

## Usage

### Loading Fixtures in Tests

```go
package mytest

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/require"
    "gopkg.in/yaml.v3"
)

// LoadConfigFixture loads a YAML config fixture
func LoadConfigFixture(t *testing.T, name string) *Config {
    t.Helper()

    path := filepath.Join("../../fixtures/config", name)
    data, err := os.ReadFile(path)
    require.NoError(t, err, "failed to read fixture: %s", name)

    var cfg Config
    err = yaml.Unmarshal(data, &cfg)
    require.NoError(t, err, "failed to parse fixture: %s", name)

    return &cfg
}

// LoadJSONFixture loads a JSON fixture
func LoadJSONFixture(t *testing.T, path string) []byte {
    t.Helper()

    fullPath := filepath.Join("../../fixtures", path)
    data, err := os.ReadFile(fullPath)
    require.NoError(t, err, "failed to read fixture: %s", path)

    return data
}
```

### Example Test with Fixtures

```go
func TestConfigValidation(t *testing.T) {
    t.Run("should accept valid config", func(t *testing.T) {
        cfg := LoadConfigFixture(t, "valid-config.yaml")

        err := cfg.Validate()
        require.NoError(t, err)
    })

    t.Run("should reject invalid config", func(t *testing.T) {
        cfg := LoadConfigFixture(t, "invalid-config.yaml")

        err := cfg.Validate()
        require.Error(t, err)
    })
}
```

## Creating Fixtures

### Configuration Fixtures

Configuration fixtures should be valid YAML files:

```yaml
# fixtures/config/valid-config.yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"
      http:
        endpoint: "0.0.0.0:4318"

processors:
  batch:
    send_batch_size: 8192
    timeout: 200ms

exporters:
  debug:
    verbosity: detailed

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [debug]
```

### Telemetry Fixtures

Telemetry fixtures should follow OTLP JSON format:

```json
{
  "resourceSpans": [
    {
      "resource": {
        "attributes": [
          {
            "key": "service.name",
            "value": { "stringValue": "test-service" }
          }
        ]
      },
      "scopeSpans": [
        {
          "spans": [
            {
              "traceId": "5B8EFFF798038103D269B633813FC60C",
              "spanId": "EEE19B7EC3C1B174",
              "name": "test-span",
              "kind": 1
            }
          ]
        }
      ]
    }
  ]
}
```

## Best Practices

1. **Keep fixtures minimal**: Only include data needed for the test
2. **Name descriptively**: Use names that describe the fixture's purpose
3. **Version control**: Commit fixtures with tests
4. **Document fixtures**: Add comments explaining special cases
5. **Avoid duplication**: Reuse fixtures across tests when appropriate

## References

- [OTLP Protocol Specification](https://opentelemetry.io/docs/specs/otlp/)
- [Test Mocks](../mocks/)
- [Testing Documentation](../../docs/TESTING.md)

---

**Copyright (c) 2024-2026 DevOpsCorner Indonesia. All rights reserved.**
