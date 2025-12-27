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

# Contributing to TelemetryFlow Collector

Thank you for your interest in contributing to TelemetryFlow Collector! This document provides guidelines and instructions for contributing to this project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Build Options](#build-options)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Coding Standards](#coding-standards)
- [Adding Components](#adding-components)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you are expected to uphold this code. Please report unacceptable behavior to [support@devopscorner.id](mailto:support@devopscorner.id).

## Getting Started

### Prerequisites

- **Go 1.24** or later
- **Git**
- **Make**
- **Docker** (optional, for container builds)
- **golangci-lint** (for linting)
- **OpenTelemetry Collector Builder (OCB)** (for OCB builds)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:

```bash
git clone https://github.com/YOUR_USERNAME/telemetryflow-collector.git
cd telemetryflow-collector
```

3. Add the upstream remote:

```bash
git remote add upstream https://github.com/telemetryflow/telemetryflow-collector.git
```

## Development Setup

### Install Dependencies

```bash
# Download Go dependencies
make deps

# Or manually
go mod download
go mod tidy
```

### Install OCB (for OCB builds)

```bash
# Install OpenTelemetry Collector Builder
make install-ocb

# Or manually
go install go.opentelemetry.io/collector/cmd/builder@v0.142.0
```

### Build the Collector

```bash
# Build standalone collector (default)
make

# Build with OCB
make build

# Build for all platforms
make build-all
```

### Install Development Tools

```bash
# Install golangci-lint (macOS)
brew install golangci-lint

# Install golangci-lint (Linux)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

## Project Structure

```
telemetryflow-collector/
├── cmd/tfo-collector/        # Standalone CLI entry point
│   └── main.go               # Cobra CLI with banner
├── internal/
│   ├── collector/            # Core collector implementation
│   ├── config/               # Configuration management
│   └── version/              # Version and banner info
├── pkg/                      # LEGO Building Blocks (reusable)
│   ├── banner/               # Startup banner
│   ├── config/               # Config loader utilities
│   └── plugin/               # Component registry
├── configs/
│   ├── tfo-collector.yaml        # Standalone config (custom format)
│   ├── otel-collector.yaml       # OCB config (standard OTel format)
│   └── otel-collector-minimal.yaml
├── tests/
│   ├── unit/                 # Unit tests
│   ├── integration/          # Integration tests
│   ├── e2e/                  # End-to-end tests
│   ├── mocks/                # Mock implementations
│   └── fixtures/             # Test fixtures
├── build/                    # Build output directory
│   ├── tfo-collector         # Standalone binary
│   ├── tfo-collector-ocb     # OCB binary
│   └── ocb/                  # OCB generated code
├── manifest.yaml             # OCB manifest
├── Makefile
├── Dockerfile                # Standalone build
├── Dockerfile.ocb            # OCB build
├── docker-compose.yml        # Docker Compose (standalone)
└── docker-compose.ocb.yml    # Docker Compose (OCB)
```

### Key Packages

| Package | Description |
|---------|-------------|
| `cmd/tfo-collector` | Main entry point with Cobra CLI |
| `internal/collector` | Core collector implementation |
| `internal/config` | Configuration parsing and validation |
| `pkg/plugin` | Component registry for extensibility |

## Build Options

TelemetryFlow Collector supports two build modes:

| Build Type | Command | Binary | Description |
|------------|---------|--------|-------------|
| **Standalone** | `make` | `tfo-collector` | Custom CLI with Cobra commands |
| **OCB** | `make build` | `tfo-collector-ocb` | Standard OpenTelemetry Collector |

### When to Use Each Build

- **Standalone**: Custom features, simplified configuration with `enabled` flags
- **OCB**: Full OpenTelemetry ecosystem compatibility, standard OTEL config format

## Making Changes

### Branch Naming

Use descriptive branch names:

- `feature/add-elasticsearch-exporter`
- `fix/batch-processor-timeout`
- `docs/update-ocb-guide`
- `refactor/simplify-config-loader`

### Create a Feature Branch

```bash
# Sync with upstream
git fetch upstream
git checkout main
git merge upstream/main

# Create your branch
git checkout -b feature/your-feature-name
```

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:

```
feat(receiver): add syslog receiver support

fix(processor): resolve batch timeout issue

docs(readme): update OCB build instructions

chore(deps): update OpenTelemetry to v0.114.0
```

## Testing

### Run All Tests

```bash
# Run unit and integration tests
make test

# Run all tests including E2E
make test-all

# Run standalone tests
make test-standalone
```

### Run Specific Tests

```bash
# Unit tests only
make test-unit

# Integration tests only
make test-integration

# E2E tests only
make test-e2e

# Run short tests
make test-short
```

### Test Coverage

```bash
# Generate coverage report
make test-coverage

# View coverage in browser
go tool cover -html=coverage-unit.out
```

### Writing Tests

- Place unit tests in `tests/unit/` mirroring the package structure
- Place integration tests in `tests/integration/`
- Place E2E tests in `tests/e2e/`
- Use mocks from `tests/mocks/`
- Use fixtures from `tests/fixtures/`

Example test:

```go
func TestBatchProcessor_Process(t *testing.T) {
    tests := []struct {
        name      string
        batchSize int
        timeout   time.Duration
        input     []Span
        want      [][]Span
        wantErr   bool
    }{
        {
            name:      "batch by size",
            batchSize: 100,
            timeout:   time.Second,
            input:     generateSpans(250),
            want:      [][]Span{spans[:100], spans[100:200], spans[200:]},
        },
        // Add more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            p := NewBatchProcessor(tt.batchSize, tt.timeout)
            got, err := p.Process(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            // Assert results...
        })
    }
}
```

## Submitting Changes

### Code Quality Checks

Before submitting, ensure your code passes all checks:

```bash
# Format code
make fmt

# Run linter
make lint

# Run go vet
make vet

# Validate configuration
make validate-config

# Run all tests
make test-all
```

### Pull Request Process

1. Update documentation if needed
2. Add tests for new functionality
3. Ensure all tests pass
4. Update CHANGELOG.md if applicable
5. Submit a pull request to `main` branch

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update
- [ ] New component (receiver/processor/exporter)

## Build Type Affected
- [ ] Standalone build
- [ ] OCB build
- [ ] Both

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] E2E tests added/updated
- [ ] Configuration validated

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests pass locally
- [ ] manifest.yaml updated (if adding OCB components)
```

## Coding Standards

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Keep functions focused and small
- Use meaningful variable names

### Error Handling

```go
// Good: Handle errors explicitly
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// Good: Use error wrapping for context
if err := processor.Start(ctx); err != nil {
    return fmt.Errorf("starting processor: %w", err)
}
```

### Logging

Use structured logging with `zap`:

```go
logger.Info("starting receiver",
    zap.String("receiver", "otlp"),
    zap.String("endpoint", endpoint),
)

logger.Error("failed to process batch",
    zap.Error(err),
    zap.Int("batch_size", len(batch)),
)
```

### Configuration

- Use YAML for configuration files
- Standalone: Support `enabled` flags for components
- OCB: Follow standard OpenTelemetry config format
- Validate configuration on load

## Adding Components

### Adding Components to OCB Build

Edit `manifest.yaml` to add OpenTelemetry components:

```yaml
# Add a new receiver
receivers:
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/receiver/newreceiver v0.114.0

# Add a new processor
processors:
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/processor/newprocessor v0.114.0

# Add a new exporter
exporters:
  - gomod: github.com/open-telemetry/opentelemetry-collector-contrib/exporter/newexporter v0.114.0
```

Then rebuild:

```bash
make clean && make build
```

### Adding Components to Standalone Build

1. Create the component in the appropriate package under `internal/`
2. Register it in the plugin registry
3. Add configuration support
4. Add tests
5. Update documentation

Example plugin registration:

```go
import "github.com/telemetryflow/telemetryflow-collector/pkg/plugin"

func init() {
    plugin.Register("my-receiver", func() plugin.Plugin {
        return &MyReceiver{}
    })
}
```

## Documentation

### Code Documentation

- Add package-level documentation
- Document exported functions, types, and constants
- Use examples where helpful

```go
// Package collector provides the core telemetry collection functionality.
//
// It supports configuring and running receivers, processors, and exporters
// in a pipeline architecture for processing telemetry data.
package collector

// Collector manages the telemetry collection pipeline.
// It coordinates receivers, processors, and exporters.
type Collector struct {
    // ...
}

// NewCollector creates a new Collector with the given configuration.
// If config is nil, default values are used.
func NewCollector(config *Config) (*Collector, error) {
    // ...
}
```

### User Documentation

- Update README.md for user-facing changes
- Add/update docs in the `docs/` directory
- Include examples for new features
- Document both standalone and OCB configurations

## Community

### Getting Help

- **GitHub Issues**: Report bugs or request features
- **Discussions**: Ask questions and share ideas
- **Email**: [support@devopscorner.id](mailto:support@devopscorner.id)

### Recognition

Contributors are recognized in:
- Release notes
- CONTRIBUTORS.md file
- GitHub contributors page

## License

By contributing to TelemetryFlow Collector, you agree that your contributions will be licensed under the Apache License 2.0.

---

**Thank you for contributing to TelemetryFlow Collector!**

Copyright (c) 2024-2026 DevOpsCorner Indonesia. All rights reserved.
