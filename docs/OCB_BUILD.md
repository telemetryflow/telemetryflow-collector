# TelemetryFlow Collector - OCB Build Guide

This guide explains how to build and use the TelemetryFlow Collector using the OpenTelemetry Collector Builder (OCB).

## Overview

The OCB (OpenTelemetry Collector Builder) is the official tool for building custom OpenTelemetry Collector distributions. It generates Go code from a manifest file (`manifest.yaml`) that defines which components to include.

```mermaid
flowchart LR
    subgraph Input["Input"]
        MANIFEST[manifest.yaml]
    end

    subgraph OCB["OCB Builder"]
        GEN[Code Generator]
    end

    subgraph Output["Generated Output"]
        MAIN[main.go]
        COMP[components.go]
        MOD[go.mod]
    end

    subgraph Binary["Final Binary"]
        BIN[tfo-collector]
    end

    MANIFEST --> GEN
    GEN --> MAIN
    GEN --> COMP
    GEN --> MOD
    MAIN --> BIN
    COMP --> BIN
    MOD --> BIN

    style OCB fill:#E3F2FD,stroke:#1565C0
    style Binary fill:#C8E6C9,stroke:#388E3C
```

### OCB-Native Build (v1.1.2+)

Since v1.1.2, TelemetryFlow Collector uses a single OCB-native build system. The legacy standalone build has been removed.

```mermaid
graph TB
    subgraph OCB_Build["OCB-Native Build"]
        O1[manifest.yaml]
        O2[OCB generates code]
        O3[go build]
        O4[tfo-collector]
        O1 --> O2 --> O3 --> O4
    end

    style OCB_Build fill:#E3F2FD,stroke:#1565C0
```

| Approach             | CLI Command                          | Config Format | Use Case            |
| -------------------- | ------------------------------------ | ------------- | ------------------- |
| **OCB-Native Build** | `tfo-collector --config config.yaml` | Standard OTEL | Full OTEL ecosystem |

## Prerequisites

- Go 1.26 or later
- OCB (OpenTelemetry Collector Builder) v0.152.1

## Installation

### Install OCB

```bash
# Install OCB matching your OTEL version
go install go.opentelemetry.io/collector/cmd/builder@v0.152.1

# Verify installation
builder version
```

### Alternative: Download Pre-built Binary

```bash
# Linux/macOS
curl -L -o ocb \
  "https://github.com/open-telemetry/opentelemetry-collector/releases/download/cmd%2Fbuilder%2Fv0.152.1/ocb_0.152.1_$(uname -s)_$(uname -m)"
chmod +x ocb
sudo mv ocb /usr/local/bin/
```

## Building the Collector

```mermaid
flowchart TD
    subgraph Step1["Step 1: Generate Code"]
        A1[Run: builder --config manifest.yaml]
        A2[OCB reads manifest]
        A3[Generates ./build/ocb/]
        A1 --> A2 --> A3
    end

    subgraph Step2["Step 2: Compile"]
        B1[cd build/ocb]
        B2[go build -o tfo-collector]
        B3[Binary created]
        B1 --> B2 --> B3
    end

    subgraph Step3["Step 3: Run"]
        C1[./tfo-collector]
        C2[--config config.yaml]
        C3[Collector Running]
        C1 --> C2 --> C3
    end

    Step1 --> Step2 --> Step3

    style Step1 fill:#BBDEFB,stroke:#1976D2
    style Step2 fill:#C8E6C9,stroke:#388E3C
    style Step3 fill:#FFE0B2,stroke:#F57C00
```

### Step 1: Generate Collector Code

```bash
# From project root
builder --config manifest.yaml

# Output will be in ./build/ocb/
```

### Step 2: Compile the Binary

```bash
cd build/ocb
go build -o tfo-collector .
```

### Step 3: Run the Collector

````bash
./tfo-collector --config configs/otel-collector.yaml

## Using Make Targets

The project includes convenient Make targets:

```bash
# Generate OCB code
make generate

# Build OCB collector
make build

# Build and run
make run

# Clean build artifacts
make clean
````

## Manifest Structure

The `manifest.yaml` defines which components to include:

```yaml
dist:
  name: tfo-collector
  description: TelemetryFlow Collector
  output_path: ./build
  module: github.com/telemetryflow/telemetryflow-collector
  skip_compilation: true

extensions:
  - gomod: go.opentelemetry.io/collector/extension/zpagesextension v0.152.1

receivers:
  - gomod: go.opentelemetry.io/collector/receiver/otlpreceiver v0.152.1

processors:
  - gomod: go.opentelemetry.io/collector/processor/batchprocessor v0.152.1

exporters:
  - gomod: go.opentelemetry.io/collector/exporter/otlpexporter v0.152.1

connectors:
  - gomod: go.opentelemetry.io/collector/connector/forwardconnector v0.152.1
```

## Adding New Components

```mermaid
flowchart TD
    FIND[1. Find Component] --> CHECK{In Core<br/>or Contrib?}
    CHECK -->|Core| CORE[go.opentelemetry.io/collector/...]
    CHECK -->|Contrib| CONTRIB[github.com/open-telemetry/opentelemetry-collector-contrib/...]

    CORE --> ADD[2. Add to manifest.yaml]
    CONTRIB --> ADD

    ADD --> REBUILD[3. Rebuild: builder --config manifest.yaml]
    REBUILD --> COMPILE[4. go build]
    COMPILE --> CONFIG[5. Configure in YAML]
    CONFIG --> USE[Component Ready]

    style FIND fill:#BBDEFB,stroke:#1976D2
    style USE fill:#C8E6C9,stroke:#388E3C
```

### 1. Find the Component

Browse available components:

- **Core**: https://github.com/open-telemetry/opentelemetry-collector
- **Contrib**: https://github.com/open-telemetry/opentelemetry-collector-contrib

### 2. Add to Manifest

```yaml
receivers:
  # Add new receiver
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbreceiver v0.152.1
```

### 3. Rebuild

```bash
builder --config manifest.yaml
cd build/ocb && go build -o tfo-collector .
```

### 4. Configure in YAML

```yaml
receivers:
  mongodb:
    hosts:
      - endpoint: localhost:27017
    username: admin
    password: ${env:MONGO_PASSWORD}
    collection_interval: 60s

service:
  pipelines:
    metrics:
      receivers: [mongodb]
      processors: [batch]
      exporters: [prometheus]
```

## Docker Build

### Build Image

```bash
docker build -f Dockerfile \
  --build-arg VERSION=1.2.2 \
  --build-arg OTEL_VERSION=0.152.1 \
  -t telemetryflow/telemetryflow-collector:1.2.2 .
```

### Run Container

```bash
docker run -d \
  --name tfo-collector \
  -p 4317:4317 \
  -p 4318:4318 \
  -p 8888:8888 \
  -p 8889:8889 \
  -p 13133:13133 \
  -v $(pwd)/configs/otel-collector.yaml:/etc/tfo-collector/otel-collector.yaml:ro \
  telemetryflow/telemetryflow-collector:1.2.2
```

### Docker Compose

```bash
docker-compose up -d
```

## Version Compatibility

| TelemetryFlow | OTEL Version | OCB Version |
| ------------- | ------------ | ----------- |
| 1.2.x         | 0.152.1      | v0.152.1    |
| 1.1.x         | 0.146.1      | v0.146.1    |
| 1.0.x         | 0.114.0      | v0.114.0    |

**Important**: All components in `manifest.yaml` must use the same OTEL version to ensure compatibility.

## OTLP HTTP Endpoints

### OCB Build (OTEL Community - v1 Only)

The OCB-native build uses the **standard OpenTelemetry OTLP receiver** which supports **v1 endpoints only**:

| Signal  | Endpoint      | Content-Type                                 |
| ------- | ------------- | -------------------------------------------- |
| Traces  | `/v1/traces`  | `application/x-protobuf`, `application/json` |
| Metrics | `/v1/metrics` | `application/x-protobuf`, `application/json` |
| Logs    | `/v1/logs`    | `application/x-protobuf`, `application/json` |

This ensures full compatibility with the OpenTelemetry specification and all OTEL SDKs.

### TFO OTLP Receiver (Dual Endpoints)

The TFO OTLP receiver (`tfootlp`) supports **both v1 and v2** endpoints:

| Version                 | Endpoint                                | Description                     |
| ----------------------- | --------------------------------------- | ------------------------------- |
| **v1** (OTEL Community) | `/v1/traces`, `/v1/metrics`, `/v1/logs` | Standard OpenTelemetry spec     |
| **v2** (TFO Platform)   | `/v2/traces`, `/v2/metrics`, `/v2/logs` | TelemetryFlow Platform-specific |

> **Recommendation:** Use **v2 endpoints** with the `tfootlp` receiver for TelemetryFlow Platform features. Use **v1 endpoints** when compatibility with standard OTEL tooling is required.

## Troubleshooting

### Build Errors

**"module not found"**

```bash
# Update Go module cache
cd build/ocb
go mod tidy
```

**"incompatible versions"**

```bash
# Ensure all components use same version in manifest.yaml
# All gomod entries should end with v0.152.1
```

### Runtime Errors

**"unknown receiver type"**

- Component not included in manifest.yaml
- Rebuild with the component added

**"failed to create pipeline"**

- Check YAML syntax in config file
- Validate with: `./tfo-collector validate --config config.yaml`

## Configuration Validation

```bash
# Validate configuration
./tfo-collector validate --config configs/otel-collector.yaml

# Show configuration (debug)
./tfo-collector --config configs/otel-collector.yaml --dry-run
```

## CLI Reference

```bash
# Basic usage
./tfo-collector --config <config-file>

# Common flags
--config          Path to configuration file
--set             Override config values (e.g., --set=service.telemetry.logs.level=debug)
--feature-gates   Enable/disable feature gates

# Validate configuration
./tfo-collector validate --config <config-file>
```

## Related Documentation

- [Component Reference](./COMPONENTS.md) - All available components
- [Configuration Guide](./CONFIGURATION.md) - Configuration examples
- [Exemplars Guide](./EXEMPLARS.md) - Setting up metrics-to-traces correlation

## External Resources

- [OpenTelemetry Collector Builder](https://github.com/open-telemetry/opentelemetry-collector/tree/main/cmd/builder)
- [OTEL Collector Configuration](https://opentelemetry.io/docs/collector/configuration/)
- [OTEL Collector Contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib)
