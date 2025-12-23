<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-collector-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-collector-light.svg">
    <img src="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-collector-light.svg" alt="TelemetryFlow Logo" width="80%">
  </picture>

  <h3>TelemetryFlow Collector (OTEL Collector)</h3>

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org/)
[![OTEL](https://img.shields.io/badge/OpenTelemetry-0.114.0-blueviolet)](https://opentelemetry.io/)

</div>

---

# Changelog

All notable changes to TelemetryFlow Collector will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
