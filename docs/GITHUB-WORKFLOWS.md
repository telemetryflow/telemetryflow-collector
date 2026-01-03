# TelemetryFlow Collector - GitHub Workflows

- **Version:** 1.1.2
- **Last Updated:** January 2026

This document describes the GitHub Actions workflows available for TelemetryFlow Collector with OCB-native architecture.

## Workflow Architecture

```mermaid
flowchart TB
    subgraph "GitHub Events"
        E1[Push to Branch]
        E2[Pull Request]
        E3[Push Tag v*.*.*]
        E4[Manual Dispatch]
    end

    subgraph "Workflows"
        CI[ci.yml<br/>CI - TFO Collector]
        DOCKER[docker.yml<br/>Docker Build]
        RELEASE[release.yml<br/>Release]
    end

    E1 --> CI
    E1 --> DOCKER
    E2 --> CI
    E3 --> DOCKER
    E3 --> RELEASE
    E4 --> CI
    E4 --> DOCKER
    E4 --> RELEASE
```

## OCB-Native Build

TelemetryFlow Collector now uses a **unified OCB-native build** - a single binary that includes all OTEL community components plus TFO custom components.

| Aspect | Description |
|--------|-------------|
| **Binary** | `tfo-collector` (single unified binary) |
| **Build** | OpenTelemetry Collector Builder (OCB) |
| **CLI** | Standard OTEL CLI with TFO branding |
| **Config** | Standard OTEL YAML format |
| **Docker Image** | `telemetryflow/telemetryflow-collector` |

---

## Workflow Files

| Workflow | File | Purpose |
|----------|------|---------|
| CI | `ci.yml` | Code quality, tests, build verification |
| Release | `release.yml` | Release binaries (RPM, DEB, DMG, ZIP) |
| Docker | `docker.yml` | Build & push Docker images |

---

## CI Workflow

**File:** `.github/workflows/ci.yml`

```mermaid
flowchart TD
    subgraph "Triggers"
        T1[Push: main, master, develop, feature/*, release/*]
        T2[Pull Request: main, master, develop]
        T3[Manual Dispatch]
    end

    subgraph "Lint & Code Quality"
        L[lint<br/>golangci-lint, staticcheck, fmt, vet]
    end

    subgraph "Tests"
        TU[test-unit<br/>Unit Tests]
        TI[test-integration<br/>Integration Tests]
    end

    subgraph "Build Verification"
        B[build<br/>Linux, macOS, Windows<br/>amd64, arm64]
    end

    subgraph "Security & Reports"
        SEC[security<br/>gosec, govulncheck]
        COV[coverage<br/>Coverage Report]
        SUM[summary<br/>CI Summary]
    end

    T1 --> L
    T2 --> L
    T3 --> L

    L --> TU
    L --> TI
    L --> B
    L --> SEC

    TU --> COV
    TI --> COV

    B --> SUM
    SEC --> SUM
    COV --> SUM
```

### CI Job Matrix

```mermaid
flowchart LR
    subgraph "build job"
        S1[linux/amd64<br/>ubuntu-latest]
        S2[linux/arm64<br/>ubuntu-latest]
        S3[darwin/amd64<br/>macos-latest]
        S4[darwin/arm64<br/>macos-latest]
        S5[windows/amd64<br/>windows-latest]
    end
```

---

## Release Workflow

**File:** `.github/workflows/release.yml`

```mermaid
flowchart TD
    subgraph "Triggers"
        T1[Push Tag: v*.*.*]
        T2[Manual Dispatch]
    end

    subgraph "Prepare"
        P[prepare<br/>Determine version, commit, branch]
    end

    subgraph "Build Binaries"
        BL[build-linux<br/>amd64, arm64]
        BW[build-windows<br/>amd64]
        BM[build-macos<br/>amd64, arm64]
    end

    subgraph "Package"
        PR[package-rpm<br/>amd64, arm64]
        PD[package-deb<br/>amd64, arm64]
        PW[package-windows<br/>ZIP with installer]
        PM[package-macos<br/>DMG amd64, arm64]
        PT[package-tarball<br/>linux, darwin<br/>amd64, arm64]
    end

    subgraph "Release"
        R[release<br/>Create GitHub Release<br/>Upload all artifacts]
    end

    T1 --> P
    T2 --> P

    P --> BL
    P --> BW
    P --> BM

    BL --> PR
    BL --> PD
    BL --> PT
    BW --> PW
    BM --> PM
    BM --> PT

    PR --> R
    PD --> R
    PW --> R
    PM --> R
    PT --> R
```

### Release Artifacts

```mermaid
flowchart LR
    subgraph "Linux"
        L1[tfo-collector-VERSION-1.x86_64.rpm]
        L2[tfo-collector-VERSION-1.aarch64.rpm]
        L3[tfo-collector_VERSION_amd64.deb]
        L4[tfo-collector_VERSION_arm64.deb]
        L5[tfo-collector-VERSION-linux-amd64.tar.gz]
        L6[tfo-collector-VERSION-linux-arm64.tar.gz]
    end

    subgraph "macOS"
        M1[tfo-collector-VERSION-darwin-amd64.dmg]
        M2[tfo-collector-VERSION-darwin-arm64.dmg]
        M3[tfo-collector-VERSION-darwin-amd64.tar.gz]
        M4[tfo-collector-VERSION-darwin-arm64.tar.gz]
    end

    subgraph "Windows"
        W1[tfo-collector-VERSION-windows-amd64.zip]
    end

    subgraph "Checksums"
        CS[checksums-sha256.txt]
    end

    L1 --> CS
    L2 --> CS
    L3 --> CS
    L4 --> CS
    L5 --> CS
    L6 --> CS
    M1 --> CS
    M2 --> CS
    M3 --> CS
    M4 --> CS
    W1 --> CS
```

---

## Docker Workflow

**File:** `.github/workflows/docker.yml`

```mermaid
flowchart TD
    subgraph "Triggers"
        T1[Push: main, master]
        T2[Push Tag: v*.*.*]
        T3[Pull Request]
        T4[Manual Dispatch]
    end

    subgraph "Prepare"
        P[prepare<br/>Docker metadata<br/>Tags & Labels]
    end

    subgraph "Build"
        B[build<br/>Multi-platform<br/>linux/amd64, linux/arm64<br/>SBOM generation]
    end

    subgraph "Security"
        S[scan<br/>Trivy vulnerability scanner]
    end

    subgraph "Report"
        R[summary<br/>Build Summary]
    end

    T1 --> P
    T2 --> P
    T3 --> P
    T4 --> P

    P --> B
    B --> S
    B --> R
```

### Docker Image Tags

```mermaid
flowchart TD
    subgraph "Git Tag v1.1.2"
        GT[v1.1.2]
    end

    subgraph "Docker Tags"
        T1[1.1.2]
        T2[1.1]
        T3[1]
        T4[latest]
        T5[sha-abc1234]
    end

    GT --> T1
    GT --> T2
    GT --> T3
    GT --> T4
    GT --> T5
```

### Docker Registry Flow

```mermaid
flowchart LR
    subgraph "Build"
        B[Docker Build<br/>Multi-platform]
    end

    subgraph "Registries"
        DH[Docker Hub<br/>telemetryflow/telemetryflow-collector]
        GH[GitHub Container Registry<br/>ghcr.io/telemetryflow/telemetryflow-collector]
    end

    subgraph "Security"
        SBOM[SBOM<br/>SPDX format]
        TRIVY[Trivy Scan<br/>CRITICAL, HIGH]
        PROV[Provenance<br/>Attestation]
    end

    B --> DH
    B --> GH
    B --> SBOM
    B --> TRIVY
    B --> PROV
```

---

## Supported Platforms

```mermaid
flowchart TB
    subgraph "Linux"
        LA[amd64<br/>x86_64]
        LR[arm64<br/>aarch64]
    end

    subgraph "macOS"
        MI[Intel<br/>amd64]
        MA[Apple Silicon<br/>arm64]
    end

    subgraph "Windows"
        WA[amd64<br/>64-bit]
    end

    subgraph "Packages"
        RPM[RPM<br/>RHEL, CentOS, Fedora]
        DEB[DEB<br/>Debian, Ubuntu]
        DMG[DMG<br/>macOS Installer]
        ZIP[ZIP<br/>Windows Portable]
        TAR[tar.gz<br/>Universal]
    end

    LA --> RPM
    LA --> DEB
    LA --> TAR
    LR --> RPM
    LR --> DEB
    LR --> TAR
    MI --> DMG
    MI --> TAR
    MA --> DMG
    MA --> TAR
    WA --> ZIP
```

---

## Environment & Secrets

```mermaid
flowchart LR
    subgraph "Secrets"
        GH[GITHUB_TOKEN<br/>Auto-provided]
        DH[DOCKERHUB_TOKEN<br/>Docker Hub push]
    end

    subgraph "Variables"
        DHU[DOCKERHUB_USERNAME<br/>Docker Hub username]
    end

    subgraph "Environment"
        GO[GO_VERSION: 1.24]
        OTEL[OTEL_VERSION: 0.142.0]
    end

    GH --> |Required| ALL[All Workflows]
    DH --> |Optional| DOCKER[Docker Workflow]
    DHU --> |Optional| DOCKER
    GO --> ALL
    OTEL --> ALL
```

---

## Security Features

```mermaid
flowchart TD
    subgraph "CI Security"
        GS[gosec<br/>Go Security Scanner]
        GV[govulncheck<br/>Vulnerability Check]
        CQL[CodeQL<br/>SARIF Upload]
    end

    subgraph "Docker Security"
        TRIVY[Trivy<br/>Container Scanning]
        SBOM[SBOM<br/>Software Bill of Materials]
        PROV[Provenance<br/>Build Attestation]
    end

    subgraph "Release Security"
        CS[SHA256 Checksums]
    end

    GS --> CQL
    GV --> CQL
    TRIVY --> CQL
```

---

## Exposed Ports

```mermaid
flowchart LR
    subgraph "OTLP Receivers"
        P1[4317<br/>gRPC]
        P2[4318<br/>HTTP]
    end

    subgraph "Metrics"
        P3[8888<br/>Self Metrics]
        P4[8889<br/>Prometheus Export]
    end

    subgraph "Extensions"
        P5[13133<br/>Health Check]
        P6[55679<br/>zPages]
        P7[1777<br/>pprof]
    end
```

---

## Quick Reference

### Release Commands

```bash
# Create release
git tag v1.1.2
git push origin v1.1.2
```

### Docker Pull Commands

```bash
# Latest version
docker pull telemetryflow/telemetryflow-collector:latest

# Specific version
docker pull telemetryflow/telemetryflow-collector:1.1.2
```

### Run Command

```bash
# Run collector
docker run -d \
  --name tfo-collector \
  -p 4317:4317 -p 4318:4318 -p 8888:8888 -p 13133:13133 \
  telemetryflow/telemetryflow-collector:1.1.2
```

---

## Configuration Files

```mermaid
flowchart TD
    subgraph "Build Files"
        DF[Dockerfile]
        MF[Makefile]
    end

    subgraph "Config Files"
        CF1[configs/tfo-collector.yaml]
        CF2[configs/otel-collector.yaml]
    end
```

---

## CLI Commands

```bash
# Run with config
./tfo-collector --config config.yaml

# Validate config
./tfo-collector validate --config config.yaml

# Show version
./tfo-collector --version

# List components
./tfo-collector components
```

---

## Links

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/)
- [Trivy Action](https://github.com/aquasecurity/trivy-action)
