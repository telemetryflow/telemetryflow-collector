# TelemetryFlow Collector - GitHub Workflows

This document describes the GitHub Actions workflows available for TelemetryFlow Collector.

## Workflow Architecture

```mermaid
flowchart TB
    subgraph "GitHub Events"
        E1[Push to Branch]
        E2[Pull Request]
        E3[Push Tag v*.*.*]
        E4[Push Tag v*.*.*-ocb]
        E5[Manual Dispatch]
    end

    subgraph "Workflows"
        CI[ci.yml<br/>CI - TFO Collector]
        DT[docker-tfo.yml<br/>Docker Build Standalone]
        DO[docker-ocb.yml<br/>Docker Build OCB]
        RT[release-tfo.yml<br/>Release Standalone]
        RO[release-ocb.yml<br/>Release OCB]
    end

    E1 --> CI
    E1 --> DT
    E2 --> CI
    E2 --> DT
    E3 --> DT
    E3 --> DO
    E3 --> RT
    E3 --> RO
    E4 --> DO
    E4 --> RO
    E5 --> CI
    E5 --> DT
    E5 --> DO
    E5 --> RT
    E5 --> RO
```

## Build Types

```mermaid
flowchart LR
    subgraph "TelemetryFlow Collector"
        subgraph "Standalone Build"
            S1[Custom Cobra CLI]
            S2[TelemetryFlow Branding]
            S3[start --config]
        end

        subgraph "OCB Build"
            O1[OTEL Collector Builder]
            O2[Standard OTEL CLI]
            O3[--config]
        end
    end

    S1 --> S2 --> S3
    O1 --> O2 --> O3
```

| Build Type | Description | CLI Command | Config File |
|------------|-------------|-------------|-------------|
| **Standalone** | Custom Cobra CLI with TelemetryFlow branding | `tfo-collector start --config config.yaml` | `tfo-collector.yaml` |
| **OCB** | Standard OpenTelemetry Collector Builder | `tfo-collector --config config.yaml` | `otel-collector.yaml` |

---

## Build Type Selection

The CI/CD workflows use tag-based build type selection to prevent duplicate builds and allow selective releases:

```mermaid
flowchart TD
    subgraph "Git Reference"
        REF[Git Push/Tag]
    end

    subgraph "Tag Detection"
        CHECK{Check Tag Suffix}
    end

    subgraph "Build Decision"
        STANDALONE[Build Standalone Only]
        OCB[Build OCB Only]
        BOTH[Build Both]
    end

    REF --> CHECK
    CHECK -->|"v*.*.*-standalone"| STANDALONE
    CHECK -->|"v*.*.*-ocb"| OCB
    CHECK -->|"v*.*.* (no suffix)"| BOTH
    CHECK -->|"main/master branch"| BOTH
    CHECK -->|"release/* branch"| BOTH
```

### Selection Matrix

| Git Reference | Standalone Builds | OCB Builds | Example |
|---------------|-------------------|------------|---------|
| `v*.*.*-standalone` | ✅ | ❌ | `v1.1.1-standalone` |
| `v*.*.*-ocb` | ❌ | ✅ | `v1.1.1-ocb` |
| `v*.*.*` (no suffix) | ✅ | ✅ | `v1.1.1` |
| `main` / `master` branch | ✅ | ✅ | Push to main |
| `release/*` branch | ✅ | ✅ | `release/v1.2.0` |
| `workflow_dispatch` | Based on input | Based on input | Manual trigger |

### Implementation Patterns

Each workflow type uses a specific pattern for build type selection:

#### CI Workflow (`ci.yml`)

Uses a `prepare` job with outputs:

```yaml
prepare:
  outputs:
    build_standalone: ${{ steps.determine.outputs.build_standalone }}
    build_ocb: ${{ steps.determine.outputs.build_ocb }}
  steps:
    - name: Determine build type
      run: |
        if [[ "$REF" == refs/tags/*-standalone ]]; then
          BUILD_STANDALONE="true"
          BUILD_OCB="false"
        elif [[ "$REF" == refs/tags/*-ocb ]]; then
          BUILD_STANDALONE="false"
          BUILD_OCB="true"
        else
          BUILD_STANDALONE="true"
          BUILD_OCB="true"
        fi

build-standalone:
  needs: prepare
  if: needs.prepare.outputs.build_standalone == 'true'

build-ocb:
  needs: prepare
  if: needs.prepare.outputs.build_ocb == 'true'
```

#### Docker Workflows (`docker-tfo.yml`, `docker-ocb.yml`)

Uses a `check-build-type` job with `should_run` output:

```yaml
# docker-tfo.yml - Skips if -ocb tag
check-build-type:
  outputs:
    should_run: ${{ steps.check.outputs.should_run }}
  steps:
    - run: |
        if [[ "$REF" == refs/tags/*-ocb ]]; then
          echo "should_run=false" >> $GITHUB_OUTPUT
        else
          echo "should_run=true" >> $GITHUB_OUTPUT
        fi

# docker-ocb.yml - Skips if -standalone tag
check-build-type:
  steps:
    - run: |
        if [[ "$REF" == refs/tags/*-standalone ]]; then
          echo "should_run=false" >> $GITHUB_OUTPUT
        else
          echo "should_run=true" >> $GITHUB_OUTPUT
        fi
```

#### Release Workflows (`release-tfo.yml`, `release-ocb.yml`)

Uses job-level conditions with `endsWith()`:

```yaml
# release-tfo.yml
jobs:
  prepare:
    if: ${{ !endsWith(github.ref, '-ocb') }}

# release-ocb.yml
jobs:
  prepare:
    if: ${{ !endsWith(github.ref, '-standalone') }}
```

---

## Workflow Files

| Workflow | File | Build Type | Purpose |
|----------|------|------------|---------|
| CI | `ci.yml` | Both | Code quality, tests, build verification |
| Release Standalone | `release-tfo.yml` | Standalone | Release binaries (RPM, DEB, DMG, ZIP) |
| Release OCB | `release-ocb.yml` | OCB | Release binaries (RPM, DEB, DMG, ZIP) |
| Docker Standalone | `docker-tfo.yml` | Standalone | Build & push Docker images |
| Docker OCB | `docker-ocb.yml` | OCB | Build & push Docker images |

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
        TE[test-e2e<br/>E2E Tests<br/>optional]
    end

    subgraph "Build Verification"
        BS[build-standalone<br/>Linux, macOS, Windows<br/>amd64, arm64]
        BO[build-ocb<br/>Linux<br/>amd64, arm64]
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
    L --> BS
    L --> BO
    L --> SEC

    TU --> TE
    TI --> TE
    TU --> COV
    TI --> COV

    BS --> SUM
    BO --> SUM
    SEC --> SUM
    COV --> SUM
```

### CI Job Matrix

```mermaid
flowchart LR
    subgraph "build-standalone"
        S1[linux/amd64<br/>ubuntu-latest]
        S2[linux/arm64<br/>ubuntu-latest]
        S3[darwin/amd64<br/>macos-latest]
        S4[darwin/arm64<br/>macos-latest]
        S5[windows/amd64<br/>windows-latest]
    end

    subgraph "build-ocb"
        O1[linux/amd64<br/>ubuntu-latest]
        O2[linux/arm64<br/>ubuntu-latest]
    end
```

### CI Manual Dispatch Options

```mermaid
flowchart TD
    MD[workflow_dispatch]

    MD --> I1[run_e2e: boolean<br/>Run E2E tests]
    MD --> I2[skip_lint: boolean<br/>Skip linting]
    MD --> I3[build_type: choice<br/>all, standalone, ocb]
```

---

## Release Workflows

### Release Standalone

**File:** `.github/workflows/release-tfo.yml`

```mermaid
flowchart TD
    subgraph "Triggers"
        T1[Push Tag: v*.*.*]
        T2[Push Tag: v*.*.*-standalone]
        T3[Manual Dispatch]
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
    T3 --> P

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

### Release OCB

**File:** `.github/workflows/release-ocb.yml`

```mermaid
flowchart TD
    subgraph "Triggers"
        T1[Push Tag: v*.*.*]
        T2[Push Tag: v*.*.*-ocb]
        T3[Manual Dispatch]
    end

    subgraph "Prepare"
        P[prepare<br/>Determine version, OTEL version, commit]
    end

    subgraph "Build with OCB"
        BL[build-linux<br/>Install OCB + Build<br/>amd64, arm64]
        BW[build-windows<br/>Install OCB + Build<br/>amd64]
        BM[build-macos<br/>Install OCB + Build<br/>amd64, arm64]
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
    T3 --> P

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

### Release Artifact Flow

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

## Docker Workflows

### Docker Standalone

**File:** `.github/workflows/docker-tfo.yml`

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

### Docker OCB

**File:** `.github/workflows/docker-ocb.yml`

```mermaid
flowchart TD
    subgraph "Triggers"
        T1[Push Tag: v*.*.*]
        T2[Manual Dispatch]
    end

    subgraph "Prepare"
        P[prepare<br/>Docker metadata<br/>OTEL version]
    end

    subgraph "Build"
        B[build<br/>Multi-platform<br/>linux/amd64, linux/arm64<br/>OCB inside Docker]
    end

    subgraph "Security"
        S[scan<br/>Trivy vulnerability scanner]
    end

    subgraph "Report"
        R[summary<br/>Build Summary]
    end

    T1 --> P
    T2 --> P

    P --> B
    B --> S
    B --> R
```

### Docker Image Tags

```mermaid
flowchart TD
    subgraph "Git Tag v1.2.3"
        GT[v1.2.3]
    end

    subgraph "Standalone Tags"
        ST1[1.2.3]
        ST2[1.2]
        ST3[1]
        ST4[latest]
        ST5[standalone]
        ST6[sha-abc1234-standalone]
    end

    subgraph "OCB Tags"
        OT1[1.2.3-ocb]
        OT2[1.2-ocb]
        OT3[1-ocb]
        OT4[latest]
        OT5[ocb]
        OT6[sha-abc1234-ocb]
    end

    GT --> ST1
    GT --> ST2
    GT --> ST3
    GT --> ST4
    GT --> ST5
    GT --> ST6

    GT --> OT1
    GT --> OT2
    GT --> OT3
    GT --> OT4
    GT --> OT5
    GT --> OT6
```

### Docker Registry Flow

```mermaid
flowchart LR
    subgraph "Build"
        B[Docker Build<br/>Multi-platform]
    end

    subgraph "Registries"
        DH[Docker Hub<br/>telemetryflow/telemetryflow-collector]
    end

    subgraph "Security"
        SBOM[SBOM<br/>SPDX format]
        TRIVY[Trivy Scan<br/>CRITICAL, HIGH]
        PROV[Provenance<br/>Attestation]
    end

    B --> DH
    B --> SBOM
    B --> TRIVY
    B --> PROV
```

---

## Tag Routing

```mermaid
flowchart TD
    subgraph "Git Tags"
        V1[v1.1.1]
        V2[v1.1.1-standalone]
        V3[v1.1.1-ocb]
    end

    subgraph "Standalone Workflows"
        RST[release-tfo.yml]
        DST[docker-tfo.yml]
    end

    subgraph "OCB Workflows"
        ROCB[release-ocb.yml]
        DOCB[docker-ocb.yml]
    end

    subgraph "CI Workflow"
        CIW[ci.yml]
    end

    V1 --> RST
    V1 --> DST
    V1 --> ROCB
    V1 --> DOCB
    V2 --> RST
    V2 --> DST
    V3 --> ROCB
    V3 --> DOCB

    V1 -.-> CIW
    V2 -.-> CIW
    V3 -.-> CIW
```

---

## Version Handling

```mermaid
flowchart LR
    subgraph "Input"
        I1[Tag: v1.1.1-standalone]
        I2[Tag: v1.1.1-ocb]
        I3[Tag: v1.1.1]
        I4[Manual: 1.1.1]
    end

    subgraph "Processing"
        P[Strip suffixes<br/>-standalone, -ocb]
    end

    subgraph "Output"
        O[Clean Version: 1.1.1]
    end

    I1 --> P
    I2 --> P
    I3 --> P
    I4 --> P
    P --> O
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
        OCB[OCB_VERSION: 0.114.0]
        OTEL[OTEL_VERSION: 0.114.0]
    end

    GH --> |Required| ALL[All Workflows]
    DH --> |Optional| DOCKER[Docker Workflows]
    DHU --> |Optional| DOCKER
    GO --> ALL
    OCB --> |OCB Build| OCB_WF[OCB Workflows]
    OTEL --> |OCB Build| OCB_WF
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

### Tag Conventions

```mermaid
flowchart LR
    subgraph "Tagging"
        T1[v1.1.1 → Both builds]
        T2[v1.1.1-standalone → Standalone only]
        T3[v1.1.1-ocb → OCB only]
    end
```

### Release Commands

```bash
# Standalone release
git tag v1.1.1
git push origin v1.1.1

# Or explicit standalone
git tag v1.1.1-standalone
git push origin v1.1.1-standalone

# OCB-only release
git tag v1.1.1-ocb
git push origin v1.1.1-ocb
```

### Docker Pull Commands

```bash
# Standalone
docker pull telemetryflow/telemetryflow-collector:latest
docker pull telemetryflow/telemetryflow-collector:1.1.1

# OCB
docker pull telemetryflow/telemetryflow-collector-ocb:latest
docker pull telemetryflow/telemetryflow-collector-ocb:1.1.1-ocb
```

### Run Commands

```bash
# Standalone (uses 'start --config')
docker run -d \
  --name tfo-collector \
  -p 4317:4317 -p 4318:4318 -p 8888:8888 -p 13133:13133 \
  telemetryflow/telemetryflow-collector:latest

# OCB (uses '--config' directly)
docker run -d \
  --name tfo-collector-ocb \
  -p 4317:4317 -p 4318:4318 -p 8888:8888 -p 13133:13133 \
  telemetryflow/telemetryflow-collector-ocb:latest
```

---

## Configuration Files

```mermaid
flowchart TD
    subgraph "Standalone Build"
        DF1[Dockerfile]
        CF1[configs/tfo-collector.yaml]
    end

    subgraph "OCB Build"
        DF2[Dockerfile.ocb]
        MF[manifest.yaml]
        CF2[configs/otel-collector.yaml]
    end
```

---

## CLI Differences

```mermaid
flowchart LR
    subgraph "Standalone CLI"
        SC1[tfo-collector start --config config.yaml]
        SC2[tfo-collector version]
        SC3[tfo-collector config validate]
    end

    subgraph "OCB CLI"
        OC1[tfo-collector --config config.yaml]
        OC2[tfo-collector --version]
        OC3[tfo-collector validate --config config.yaml]
    end
```

---

## Links

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [OpenTelemetry Collector Builder](https://github.com/open-telemetry/opentelemetry-collector/tree/main/cmd/builder)
- [Trivy Action](https://github.com/aquasecurity/trivy-action)
