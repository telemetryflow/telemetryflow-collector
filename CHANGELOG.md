<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-collector-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-collector-light.svg">
    <img src="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-collector-light.svg" alt="TelemetryFlow Logo" width="80%">
  </picture>

  <h3>TelemetryFlow Collector (OTEL Collector)</h3>

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org/)
[![OTEL](https://img.shields.io/badge/OpenTelemetry-0.142.0-blueviolet)](https://opentelemetry.io/)

</div>

---

# Changelog

All notable changes to TelemetryFlow Collector will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.1] - 2024-12-29

### Changed

- **Version Bump**: Updated version from 1.1.0 to 1.1.1 across all files
- **Documentation**: Added Mermaid diagrams to all documentation files
  - COMPONENTS.md - Component overview and data flow diagrams
  - CONFIGURATION.md - Config structure and file lookup flowcharts
  - OCB_BUILD.md - Build process and component workflow diagrams
  - EXEMPLARS.md - Converted ASCII art to Mermaid diagrams
- **Testing**: Reorganized unit tests following DDD (Domain-Driven Design) pattern
  - domain/ - Collector, pipeline, plugin tests
  - application/ - OTEL collector, config tests
  - infrastructure/ - Exporter, receiver, config loader tests
  - presentation/ - Banner, version tests
- **Documentation**: Added ARCHITECTURE.md and TESTING.md guides
- **Fixtures**: Added test fixtures for configuration and telemetry data

### Fixed

- Removed duplicate entries in Makefile
- Fixed markdown lint warnings in documentation files

## [1.1.0] - 2024-12-27

### Added

- **OpenTelemetry Collector v0.142.0**: Upgraded from v0.114.0 to v0.142.0
- **TelemetryFlow Configuration Section**: New `telemetryflow:` config section for platform authentication
  - `api_key_id` and `api_key_secret` for TelemetryFlow authentication
  - `endpoint` for backend connectivity
  - Environment variable substitution support
- **Standard OTEL Configuration**: Both Standalone and OCB builds now use standard OpenTelemetry Collector YAML format
- **New Documentation**:
  - [CONFIGURATION.md](docs/CONFIGURATION.md) - Comprehensive configuration guide
  - [COMPONENTS.md](docs/COMPONENTS.md) - Available receivers, processors, exporters reference
  - [EXEMPLARS.md](docs/EXEMPLARS.md) - Exemplars and metrics-to-traces correlation
  - [OCB_BUILD.md](docs/OCB_BUILD.md) - OpenTelemetry Collector Builder guide
- **Enhanced Connectors**: Added spanmetrics and servicegraph connectors for Exemplars support
- **Cloud Provider Receivers**: AWS CloudWatch, Azure Monitor, Google Cloud Pub/Sub
- **Database Receivers**: MySQL, PostgreSQL, MongoDB, Redis, Elasticsearch, OracleDB
- **Message Queue Support**: Kafka, RabbitMQ, Pulsar receivers and exporters
- **APM Exporters**: Datadog, Splunk HEC, SignalFx, Honeycomb, Coralogix, Logz.io

### Changed

- **Configuration Format**: Standalone config now uses standard OTEL format with optional TelemetryFlow extensions
- **Build System**: Updated OCB builder to v0.142.0
- **Environment Variables**: Standardized to use `TELEMETRYFLOW_*` prefix
- **GitHub Workflows**:
  - Updated CodeQL Action from v3 to v4
  - Enhanced Docker workflow with disk cleanup, Go version tracking
  - Renamed workflows from `docker-standalone.yml` to `docker-tfo.yml`
  - Renamed workflows from `release-standalone.yml` to `release-tfo.yml`

### Fixed

- Config loader now properly handles both legacy and new TelemetryFlow config sections
- OTEL Collector integration for standard config parsing

### Dependencies

- OpenTelemetry Collector: v0.142.0
- Go: 1.24+

### Breaking Changes

- Removed `enabled` flags from configuration - now uses standard OTEL service pipelines
- Renamed config files: `ocb-collector.yaml` → `otel-collector.yaml`

## [1.0.1] - 2024-12-17

### Added

- GitHub Actions workflow for Docker image building (Standalone)
- GitHub Actions workflow for Docker image building (OCB)
- Multi-platform Docker support (linux/amd64, linux/arm64)
- SBOM generation for Docker images
- Trivy security scanning in CI/CD pipeline
- GitHub Container Registry publishing
- Docker Hub publishing support
- GitHub Workflows documentation with Mermaid diagrams

### Changed

- Updated documentation structure with new GITHUB-WORKFLOWS.md

## [1.0.0] - 2024-12-17

### Added

- Initial release of TelemetryFlow Collector
- Dual build system: Standalone and OCB
- OpenTelemetry Collector v0.114.0 base

#### Standalone Build

- Custom Cobra CLI with TelemetryFlow branding
- ASCII art startup banner
- Custom configuration format with `enabled` flags
- Commands: `start`, `version`, `config validate`

#### OCB Build

- Standard OpenTelemetry Collector CLI
- Full OTEL ecosystem compatibility
- Component manifest (manifest.yaml)

#### Receivers

- OTLP (gRPC and HTTP)
- Host Metrics
- File Log
- Prometheus
- Kafka
- Kubernetes Cluster
- Kubernetes Events
- Syslog

#### Processors

- Batch
- Memory Limiter
- Attributes
- Resource
- Resource Detection
- Filter
- Transform
- K8s Attributes
- Tail Sampling

#### Exporters

- OTLP (gRPC)
- OTLP HTTP
- Debug
- Prometheus
- Prometheus Remote Write
- Kafka
- Loki
- Elasticsearch
- File

#### Extensions

- Health Check
- pprof
- zPages
- Basic Auth
- Bearer Token Auth
- File Storage

### Infrastructure

- Docker support (separate Dockerfiles for each build type)
- Docker Compose configurations
- Systemd service configuration
- RPM and DEB package builds
- macOS DMG installer
- Windows ZIP with PowerShell installer
- Kubernetes deployment examples

### Documentation

- README with quick start guide
- Installation guide for all platforms
- Build system comparison
- GitHub workflows documentation

---

## Version History

| Version | Date | Description |
|---------|------|-------------|
| 1.1.1 | 2024-12-29 | Documentation improvements, DDD test reorganization |
| 1.1.0 | 2024-12-27 | OTEL v0.142.0, TelemetryFlow config, standard OTEL format |
| 1.0.1 | 2024-12-17 | Docker workflows, SBOM, multi-platform support |
| 1.0.0 | 2024-12-17 | Initial release |

## Build Types

### Standalone vs OCB

| Feature | Standalone | OCB |
|---------|------------|-----|
| CLI | Custom Cobra | Standard OTEL |
| Config format | `enabled` flags | Standard OTEL |
| Start command | `start --config` | `--config` |
| Binary name | `tfo-collector` | `tfo-collector-ocb` |
| Docker image | `telemetryflow-collector` | `telemetryflow-collector-ocb` |

## Upgrade Guide

### From Pre-release to 1.0.0

This is the initial stable release. No upgrade steps required.

### Switching Build Types

#### Standalone to OCB

1. Update configuration from custom format to standard OTEL format
2. Change start command from `start --config` to `--config`
3. Pull OCB Docker image or install OCB package

#### OCB to Standalone

1. Update configuration to use `enabled` flags format
2. Change start command from `--config` to `start --config`
3. Pull Standalone Docker image or install Standalone package

## Support

- **Issues**: [GitHub Issues](https://github.com/telemetryflow/telemetryflow-platform/issues)
- **Documentation**: [https://docs.telemetryflow.id](https://docs.telemetryflow.id)
- **Email**: support@telemetryflow.id
