# =============================================================================
# TelemetryFlow Collector - Dockerfile (OCB Native Build)
# =============================================================================
#
# TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
# Copyright (c) 2024-2026 DevOpsCorner Indonesia. All rights reserved.
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
# OCB Native Build: 100% OpenTelemetry Collector Builder with TFO Components
# =============================================================================
#
# This Dockerfile builds the TFO Collector using OCB (OpenTelemetry Collector
# Builder) with custom TFO components:
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
FROM golang:1.24-alpine AS builder

# Build arguments
ARG VERSION=1.1.2
ARG GIT_COMMIT=unknown
ARG GIT_BRANCH=unknown
ARG BUILD_TIME=unknown
ARG OTEL_VERSION=0.142.0

# Install build dependencies
RUN apk add --no-cache \
    git \
    make \
    ca-certificates \
    tzdata \
    curl

# Set working directory
WORKDIR /build

# Install OCB (OpenTelemetry Collector Builder)
RUN go install go.opentelemetry.io/collector/cmd/builder@v${OTEL_VERSION}

# Copy manifest and components
COPY manifest.yaml .
COPY components/ ./components/

# Create output directory for OCB
RUN mkdir -p ./build/ocb

# Generate collector code with OCB
RUN builder --config manifest.yaml

# Build OCB binary
WORKDIR /build/build/ocb
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w \
        -X 'main.Version=${VERSION}' \
        -X 'main.GitCommit=${GIT_COMMIT}' \
        -X 'main.BuildTime=${BUILD_TIME}'" \
    -o /tfo-collector .

# -----------------------------------------------------------------------------
# Stage 2: Runtime
# -----------------------------------------------------------------------------
FROM alpine:3.21

# Build arguments for labels
ARG VERSION=1.1.2
ARG OTEL_VERSION=0.142.0

# =============================================================================
# TelemetryFlow Metadata Labels (OCI Image Spec)
# =============================================================================
LABEL org.opencontainers.image.title="TelemetryFlow Collector" \
      org.opencontainers.image.description="Enterprise-grade OpenTelemetry Collector (OCB Native) - Community Enterprise Observability Platform (CEOP)" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.vendor="TelemetryFlow" \
      org.opencontainers.image.authors="DevOpsCorner Indonesia <support@telemetryflow.id>" \
      org.opencontainers.image.url="https://telemetryflow.id" \
      org.opencontainers.image.documentation="https://docs.telemetryflow.id" \
      org.opencontainers.image.source="https://github.com/telemetryflow/telemetryflow-collector" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.base.name="alpine:3.21" \
      # TelemetryFlow specific labels
      io.telemetryflow.product="TelemetryFlow Collector" \
      io.telemetryflow.component="tfo-collector" \
      io.telemetryflow.platform="CEOP" \
      io.telemetryflow.build.type="ocb-native" \
      io.telemetryflow.otel.version="${OTEL_VERSION}" \
      io.telemetryflow.maintainer="DevOpsCorner Indonesia"

# Update packages to get security patches (CVE fixes) and install runtime dependencies
RUN apk upgrade --no-cache && \
    apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    && rm -rf /var/cache/apk/*

# Create non-root user and group
RUN addgroup -g 10001 -S telemetryflow && \
    adduser -u 10001 -S telemetryflow -G telemetryflow -h /home/telemetryflow

# Create required directories
RUN mkdir -p \
    /etc/tfo-collector \
    /var/lib/tfo-collector/queue \
    /var/log/tfo-collector \
    && chown -R telemetryflow:telemetryflow \
    /etc/tfo-collector \
    /var/lib/tfo-collector \
    /var/log/tfo-collector

# Copy binary from builder
COPY --from=builder /tfo-collector /usr/local/bin/tfo-collector
RUN chmod +x /usr/local/bin/tfo-collector

# Copy default configuration
COPY configs/tfo-collector.yaml /etc/tfo-collector/tfo-collector.yaml
RUN chown telemetryflow:telemetryflow /etc/tfo-collector/tfo-collector.yaml

# Switch to non-root user
USER telemetryflow

# Set working directory
WORKDIR /home/telemetryflow

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
HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
    CMD curl -f http://localhost:13133/ || exit 1

# =============================================================================
# Entrypoint & Command (OCB Native - standard OTEL CLI)
# =============================================================================
ENTRYPOINT ["/usr/local/bin/tfo-collector"]
CMD ["--config", "/etc/tfo-collector/tfo-collector.yaml"]

# =============================================================================
# Build Information
# =============================================================================
# Build with:
#   docker build \
#     --build-arg VERSION=1.1.2 \
#     --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
#     --build-arg GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD) \
#     --build-arg BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ') \
#     --build-arg OTEL_VERSION=0.142.0 \
#     -t telemetryflow/telemetryflow-collector:1.1.2 .
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
#     telemetryflow/telemetryflow-collector:1.1.2
#
# Validate config:
#   docker run --rm \
#     -v /path/to/config.yaml:/etc/tfo-collector/tfo-collector.yaml:ro \
#     telemetryflow/telemetryflow-collector:1.1.2 \
#     validate --config /etc/tfo-collector/tfo-collector.yaml
# =============================================================================
