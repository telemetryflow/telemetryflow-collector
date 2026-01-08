# Test Fixtures

This directory contains test fixtures and test data for TelemetryFlow Collector tests.

## Directory Structure

```
fixtures/
├── configs/       # Test configuration files
├── otlp/          # OTLP test data (metrics, traces, logs)
└── responses/     # Expected response fixtures
```

## Usage

These fixtures are used by unit, integration, and e2e tests to provide consistent test data.

### Config Files
- `minimal.yaml` - Minimal valid configuration
- `full.yaml` - Full configuration with all options
- `invalid.yaml` - Invalid configuration for error testing

### OTLP Data
- `metrics.json` - Sample OTLP metrics data
- `traces.json` - Sample OTLP traces data
- `logs.json` - Sample OTLP logs data

### Response Fixtures
- `success.json` - Expected success response
- `error.json` - Expected error response

## License

TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
