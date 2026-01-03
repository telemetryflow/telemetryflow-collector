<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-collector-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-collector-light.svg">
    <img src="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-collector-light.svg" alt="TelemetryFlow Logo" width="80%">
  </picture>

  <h3>TelemetryFlow Collector (OTEL Collector)</h3>

[![Version](https://img.shields.io/badge/Version-1.1.2-orange.svg)](CHANGELOG.md)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org/)
[![OTEL](https://img.shields.io/badge/OpenTelemetry-0.142.0-blueviolet)](https://opentelemetry.io/)
[![OpenTelemetry](https://img.shields.io/badge/OTLP-100%25%20Compliant-success?logo=opentelemetry)](https://opentelemetry.io/)

</div>

---

# Changelog

All notable changes to TelemetryFlow Collector will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.2] - 2026-01-03

### Added

- **OCB-Native Architecture**: Unified single build system based on OpenTelemetry Collector Builder (OCB)
  - Single binary (`tfo-collector`) with 85+ OTEL community components
  - TFO custom components integrated natively into OCB build
  - Removed legacy standalone build completely
- **TFO Custom Components**: Production-ready custom components for TelemetryFlow Platform
  - `tfootlpreceiver` - OTLP receiver with dual v1/v2 endpoint support
  - `tfoexporter` - Exporter with auto TFO authentication header injection
  - `tfoauthextension` - Centralized API key management
  - `tfoidentityextension` - Collector identity and resource enrichment
- **OTEL 0.142.0 API Compatibility**: Updated all components for latest OTEL Collector API
  - Fixed factory map creation using new map-based approach
  - Updated exporter settings (removed deprecated QueueConfig)
  - Fixed `component.ID` API changes (String() vs Type())
  - Updated `ToClient()` signature for HTTP client creation

### Changed

- **Build System**: Simplified to OCB-native only
  - Makefile cleaned up - removed legacy/standalone targets
  - Single `make build` target for OCB-native build
  - `make ci-build` target for GitHub Actions CI
  - Unified binary name: `tfo-collector` (no more `-ocb` suffix)
- **GitHub Workflows**: Consolidated for OCB-native
  - Updated `ci.yml` for OCB-native only builds
  - Updated `release.yml` for unified release workflow
  - Updated `docker.yml` for single image build
  - Removed obsolete `docker-tfo.yml` and `release-tfo.yml`
- **Documentation**: Updated all docs for unified architecture
  - Removed dual build references
  - Updated configuration examples
  - Simplified installation instructions

### Removed

- **Legacy Standalone Build**: Fully deprecated
  - Removed custom Cobra CLI implementation
  - Removed `cmd/tfo-collector-legacy/` directory
  - Removed `Dockerfile.ocb` (merged into `Dockerfile`)
  - Removed `docker-compose.ocb.yml` (merged into `docker-compose.yml`)
  - Removed `internal/collector/`, `internal/cli/`, `internal/pipeline/` packages

### Fixed

- OTEL 0.142.0 API compatibility issues in components
- Build cache conflicts with Go version management

## [1.1.1] - 2025-01-01

### Added

- **OTLP Dual Endpoint Support**: Documented dual OTLP HTTP endpoint support
  - TFO Standalone: v1 (OTEL Community) + v2 (TelemetryFlow Platform)
  - OCB Build: v1 only (standard OpenTelemetry)
- **Endpoint Documentation**: Added OTLP endpoint tables to all relevant files
  - README.md - New "OTLP HTTP Endpoints" section
  - docs/README.md - OTLP Capabilities section
  - docs/COMPONENTS.md - HTTP API Endpoints table
  - docs/CONFIGURATION.md - OTLP Configuration section
  - docs/INSTALLATION.md - Endpoint version test examples
  - docs/OCB_BUILD.md - OTLP HTTP Endpoints section
  - docs/EXEMPLARS.md - OTLP Endpoints section
- **Config Documentation**: Added endpoint comments to all config files
  - configs/tfo-collector.yaml - v1+v2 endpoint documentation
  - configs/otel-collector.yaml - v1 only endpoint documentation
  - configs/otel-collector-minimal.yaml - v1 only endpoint documentation
- **Docker Compose Documentation**: Added endpoint info to all docker-compose files
  - docker-compose.yml - TFO Standalone (v1+v2)
  - docker-compose.ocb.yml - OCB (v1 only)
  - docker-compose.e2e.yml - E2E Testing (v1+v2)
- **GitHub Workflow Updates**: Added endpoint info to workflow summaries
  - docker-ocb.yml - v1 only endpoints in summary
  - docker-tfo.yml - v1+v2 endpoints in summary
  - release-ocb.yml - v1 only endpoints in release notes
  - release-tfo.yml - v1+v2 endpoints in release notes
- **CI/CD Build Type Selection**: Implemented tag-based build type selection
  - Tag `v*.*.*-standalone` → Standalone only builds
  - Tag `v*.*.*-ocb` → OCB only builds
  - Tag `v*.*.*` (no suffix) → Both builds run
  - Branch `main`/`master`/`release/*` → Both builds run
  - Prevents duplicate builds when using type-specific tags

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
- **GitHub Workflows**: Updated all CI/CD workflows with build type selection
  - ci.yml - Added `prepare` job with `build_standalone`/`build_ocb` outputs
  - docker-tfo.yml - Added `check-build-type` job, skips on `-ocb` tags
  - docker-ocb.yml - Added `check-build-type` job, skips on `-standalone` tags
  - release-tfo.yml - Uses `!endsWith(github.ref, '-ocb')` condition
  - release-ocb.yml - Uses `!endsWith(github.ref, '-standalone')` condition
- **Disk Space Optimization**: Added aggressive disk cleanup for OCB builds
  - Removes dotnet, android, ghc, CodeQL, boost, swift, AGENT_TOOLSDIRECTORY
  - Fixes "no space left on device" error during OCB compilation

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
| 1.1.2 | 2026-01-03 | OCB-native architecture, unified build system |
| 1.1.1 | 2025-01-01 | Documentation improvements, DDD test reorganization |
| 1.1.0 | 2024-12-27 | OTEL v0.142.0, TelemetryFlow config, standard OTEL format |
| 1.0.1 | 2024-12-17 | Docker workflows, SBOM, multi-platform support |
| 1.0.0 | 2024-12-17 | Initial release |

## Build Architecture

### OCB-Native Build (v1.1.2+)

| Feature | Description |
|---------|-------------|
| Binary | `tfo-collector` |
| Build | OpenTelemetry Collector Builder (OCB) |
| Components | 85+ OTEL community + TFO custom components |
| CLI | Standard OTEL CLI with TFO branding |
| Config | Standard OTEL YAML format |
| Docker | `telemetryflow/telemetryflow-collector` |

### TFO Custom Components

| Component | Type | Description |
|-----------|------|-------------|
| `tfootlp` | Receiver | OTLP with v1+v2 endpoint support |
| `tfo` | Exporter | Auto TFO auth header injection |
| `tfoauth` | Extension | API key management |
| `tfoidentity` | Extension | Collector identity |

## Upgrade Guide

### From v1.1.1 to v1.1.2

1. **Single Binary**: Replace both `tfo-collector` and `tfo-collector-ocb` with unified `tfo-collector`
2. **Config**: Standard OTEL format works unchanged
3. **Docker**: Use single image `telemetryflow/telemetryflow-collector`
4. **Commands**: Remove `start` prefix - use `tfo-collector --config config.yaml` directly

### From v1.0.x to v1.1.x

1. Update configuration from custom format to standard OTEL format
2. Use new `telemetryflow:` section for TFO-specific settings
3. Remove `enabled` flags - use `service.pipelines` instead

## Support

- **Issues**: [GitHub Issues](https://github.com/telemetryflow/telemetryflow-platform/issues)
- **Documentation**: [https://docs.telemetryflow.id](https://docs.telemetryflow.id)
- **Email**: support@telemetryflow.id
