# =============================================================================
# TelemetryFlow Collector - Dockerfile (Native Go Build)
# =============================================================================
#
# TelemetryFlow Collector v1.4.0 (Based on OpenTelemetry Collector Builder (OCB) 0.165.0)
# AI-Powered Observability & Incident Response Management (IRM) Platform
# Copyright (c) 2026 Telemetri Data Indonesia. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# =============================================================================
# Native Go Build with Custom TFO Components
# =============================================================================
#
# This Dockerfile builds the TFO Collector directly from source with custom
# TFO components:
#   - tfootlp receiver (v1/v2 endpoint support)
#   - tfo exporter (auto TFO auth injection)
#   - tfoauth extension (API key management)
#   - tfoidentity extension (collector identity)
#
# OTLP HTTP Endpoints:
#   v1 (Community/Open - NO AUTH): /v1/traces, /v1/metrics, /v1/logs
#   v2 (TFO Platform - AUTH):      /v2/traces, /v2/metrics, /v2/logs
#
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Builder
# -----------------------------------------------------------------------------
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

# Build arguments
ARG VERSION=1.4.0
ARG GIT_COMMIT=unknown
ARG GIT_BRANCH=unknown
ARG BUILD_TIME=unknown
ARG OTEL_VERSION=0.165.0
ARG TARGETOS=linux
ARG TARGETARCH

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    ca-certificates \
    tzdata

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Copy components (needed for replace directives in go.mod)
COPY components/ ./components/

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the binary
# Uses TARGETOS/TARGETARCH for multi-arch support (amd64, arm64)
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags "-s -w \
    -X 'github.com/telemetryflow/telemetryflow-collector/internal/version.Version=${VERSION}' \
    -X 'github.com/telemetryflow/telemetryflow-collector/internal/version.GitCommit=${GIT_COMMIT}' \
    -X 'github.com/telemetryflow/telemetryflow-collector/internal/version.GitBranch=${GIT_BRANCH}' \
    -X 'github.com/telemetryflow/telemetryflow-collector/internal/version.BuildTime=${BUILD_TIME}'" \
    -o /tfo-collector ./cmd/tfo-collector

# Build the static healthcheck probe.
# The distroless/static runtime ships no shell and no curl, so the container
# HEALTHCHECK runs this tiny statically linked binary instead of `curl`.
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags "-s -w" \
    -o /healthcheck ./cmd/healthcheck

# Stage the writable runtime directory tree and default config under /out.
# distroless has no shell, so these dirs (and their nonroot ownership) are
# created here and copied into the runtime image with `COPY --chown`.
RUN mkdir -p \
    /out/etc/tfo-collector \
    /out/var/lib/tfo-collector/queue \
    /out/var/log/tfo-collector \
    /out/tmp/tfo-collector
COPY configs/tfo-collector.yaml /out/etc/tfo-collector/tfo-collector.yaml

# -----------------------------------------------------------------------------
# Stage 2: Runtime (distroless static — minimal attack surface)
# -----------------------------------------------------------------------------
# The tfo-collector binary is built CGO_ENABLED=0 (fully static), so it needs no
# libc at all. gcr.io/distroless/static-debian12 contains only ca-certificates,
# tzdata, /etc/passwd (with the nonroot user, uid 65532) and a few base files —
# NO shell, NO package manager, NO OS utilities. This removes the base-image OS
# packages (openssl, sqlite, util-linux, coreutils, ...) that Trivy was flagging
# as CRITICAL/HIGH, rather than suppressing those findings.
FROM gcr.io/distroless/static-debian12:nonroot

# Build arguments for labels
ARG VERSION=1.4.0
ARG OTEL_VERSION=0.165.0

# =============================================================================
# TelemetryFlow Metadata Labels (OCI Image Spec)
# =============================================================================
LABEL org.opencontainers.image.title="TelemetryFlow Collector" \
    org.opencontainers.image.description="Enterprise-grade OpenTelemetry Collector - AI-Powered Observability & Incident Response Management (IRM) Platform" \
    org.opencontainers.image.version="${VERSION}" \
    org.opencontainers.image.vendor="TelemetryFlow" \
    org.opencontainers.image.authors="Telemetri Data Indonesia <support@telemetryflow.id>" \
    org.opencontainers.image.url="https://telemetryflow.id" \
    org.opencontainers.image.documentation="https://docs.telemetryflow.id" \
    org.opencontainers.image.source="https://github.com/telemetryflow/telemetryflow-collector" \
    org.opencontainers.image.licenses="Apache-2.0" \
    org.opencontainers.image.base.name="gcr.io/distroless/static-debian12:nonroot" \
    # TelemetryFlow specific labels
    io.telemetryflow.product="TelemetryFlow Collector" \
    io.telemetryflow.component="tfo-collector" \
    io.telemetryflow.platform="CEOP" \
    io.telemetryflow.build.type="native" \
    io.telemetryflow.otel.version="${OTEL_VERSION}" \
    io.telemetryflow.maintainer="Telemetri Data Indonesia"

# NOTE on security patching: distroless images carry no package manager or shell,
# so there is no `apt-get dist-upgrade` step. Instead, base-image freshness comes
# from rebuilding against the latest distroless/static-debian12 digest. The image
# contains only ca-certificates and tzdata (both needed by the collector for
# outbound TLS and timezone handling) — there are no openssl/sqlite/util-linux
# OS packages to patch, which is what eliminates the Trivy CRITICAL/HIGH findings.

# Copy the collector and the static healthcheck probe from the builder.
# Both are built CGO_ENABLED=0 and need no shared libraries.
COPY --from=builder /tfo-collector /usr/local/bin/tfo-collector
COPY --from=builder /healthcheck /usr/local/bin/healthcheck

# Copy the staged writable directory tree + default config, owned by the
# distroless nonroot user (uid/gid 65532). distroless has no shell, so the dirs
# and ownership were prepared in the builder stage above.
COPY --from=builder --chown=65532:65532 /out/etc /etc
COPY --from=builder --chown=65532:65532 /out/var /var
COPY --from=builder --chown=65532:65532 /out/tmp /tmp

# Run as the distroless nonroot user (uid/gid 65532)
USER 65532:65532

# Set working directory
WORKDIR /home/nonroot

# Set TMPDIR for Go runtime temp files
ENV TMPDIR=/tmp/tfo-collector

# =============================================================================
# Exposed Ports
# =============================================================================
# 4317  - OTLP gRPC receiver
# 4318  - OTLP HTTP receiver (v1 + v2 endpoints)
# 8888  - Prometheus metrics (self-observability)
# 8889  - Prometheus exporter
# 13133 - Health check endpoint
# 55679 - zPages debugging
# 1777  - pprof profiling
EXPOSE 4317 4318 8888 8889 13133 55679 1777

# =============================================================================
# Health Check
# =============================================================================
# distroless has no shell, so the healthcheck uses the exec form (JSON array)
# with the static /usr/local/bin/healthcheck probe instead of `curl`.
HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
    CMD ["/usr/local/bin/healthcheck"]

# =============================================================================
# Entrypoint & Command
# =============================================================================
ENTRYPOINT ["/usr/local/bin/tfo-collector"]
CMD ["-c", "/etc/tfo-collector/tfo-collector.yaml"]

# =============================================================================
# Build Information
# =============================================================================
# Build with:
#   docker build \
#     --build-arg VERSION=1.4.0 \
#     --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
#     --build-arg GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD) \
#     --build-arg BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ') \
#     --build-arg OTEL_VERSION=0.165.0 \
#     -t telemetryflow/telemetryflow-collector:1.4.0 .
#
# Run with:
#   docker run -d \
#     --name tfo-collector \
#     -p 4317:4317 \
#     -p 4318:4318 \
#     -p 8888:8888 \
#     -p 13133:13133 \
#     -e TELEMETRYFLOW_API_KEY_ID=tfk_your_key \
#     -e TELEMETRYFLOW_API_KEY_SECRET=tfs_your_secret \
#     -v /path/to/config.yaml:/etc/tfo-collector/tfo-collector.yaml:ro \
#     telemetryflow/telemetryflow-collector:1.4.0
#
# Validate config:
#   docker run --rm \
#     -v /path/to/config.yaml:/etc/tfo-collector/tfo-collector.yaml:ro \
#     telemetryflow/telemetryflow-collector:1.4.0 \
#     validate -c /etc/tfo-collector/tfo-collector.yaml
# =============================================================================
