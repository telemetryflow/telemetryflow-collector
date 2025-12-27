# TelemetryFlow Collector - Exemplars Guide

This guide explains how to set up exemplars support for metrics-to-traces correlation in the TelemetryFlow Collector.

## What are Exemplars?

Exemplars are references from aggregated metric data points to individual trace spans. They allow you to:

- Click on a metric data point and jump directly to a related trace
- Understand which specific requests contributed to a latency percentile
- Debug issues by correlating metrics anomalies with trace data

```
┌─────────────────────────────────────────────────────────────┐
│                     Prometheus/Grafana                       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Latency Histogram (p99 = 250ms)                     │   │
│  │     ▲                                                │   │
│  │     │    *  ← Click exemplar                         │   │
│  │     │   * *                                          │   │
│  │     │  *   *                                         │   │
│  │     └──────────────────────────────────► time        │   │
│  └─────────────────────────────────────────────────────┘   │
│                           │                                  │
│                           ▼                                  │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Trace: service-a → service-b → database             │   │
│  │  Duration: 247ms                                     │   │
│  │  Trace ID: abc123...                                 │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Architecture Overview

```
┌─────────────┐     ┌──────────────────────────────────────────────┐
│ Application │────▶│              TelemetryFlow Collector         │
│  (traces)   │     │                                              │
└─────────────┘     │  ┌────────────┐    ┌─────────────────────┐  │
                    │  │   OTLP     │───▶│   spanmetrics       │  │
                    │  │  Receiver  │    │   connector         │  │
                    │  └────────────┘    │  (derives metrics   │  │
                    │        │           │   with exemplars)   │  │
                    │        │           └──────────┬──────────┘  │
                    │        │                      │              │
                    │        ▼                      ▼              │
                    │  ┌─────────────┐    ┌─────────────────────┐  │
                    │  │   Trace     │    │  Prometheus         │  │
                    │  │  Exporter   │    │  Exporter           │  │
                    │  │  (Jaeger)   │    │  (OpenMetrics)      │  │
                    │  └──────┬──────┘    └──────────┬──────────┘  │
                    └─────────┼─────────────────────┼──────────────┘
                              │                     │
                              ▼                     ▼
                    ┌─────────────────┐   ┌─────────────────────┐
                    │     Jaeger      │   │    Prometheus       │
                    │  (trace store)  │   │  (metrics store)    │
                    └─────────────────┘   └─────────────────────┘
```

## Configuration

### Complete Exemplars Configuration

```yaml
# =============================================================================
# RECEIVERS
# =============================================================================
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: "0.0.0.0:4317"
      http:
        endpoint: "0.0.0.0:4318"

# =============================================================================
# PROCESSORS
# =============================================================================
processors:
  memory_limiter:
    check_interval: 1s
    limit_percentage: 80
    spike_limit_percentage: 25

  batch:
    timeout: 200ms
    send_batch_size: 8192

# =============================================================================
# CONNECTORS - Key for Exemplars
# =============================================================================
connectors:
  # Span Metrics Connector - derives metrics from traces with EXEMPLARS
  spanmetrics:
    histogram:
      explicit:
        buckets: [1ms, 5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s, 10s]
    dimensions:
      - name: http.method
        default: GET
      - name: http.status_code
      - name: http.route
      - name: rpc.method
      - name: rpc.service
    exemplars:
      enabled: true  # <-- Enable exemplars
    namespace: traces
    metrics_flush_interval: 15s
    aggregation_temporality: AGGREGATION_TEMPORALITY_CUMULATIVE

  # Service Graph Connector - builds service dependency graphs
  servicegraph:
    latency_histogram_buckets: [1ms, 5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s, 10s]
    dimensions:
      - http.method
      - http.status_code
    store:
      ttl: 2s
      max_items: 1000
    cache_loop: 1s
    store_expiration_loop: 2s
    virtual_node_peer_attributes:
      - db.system
      - messaging.system
      - rpc.service

# =============================================================================
# EXPORTERS
# =============================================================================
exporters:
  # Debug exporter
  debug:
    verbosity: detailed

  # Prometheus exporter with OpenMetrics (required for exemplars)
  prometheus:
    endpoint: "0.0.0.0:8889"
    namespace: telemetryflow
    send_timestamps: true
    metric_expiration: 5m
    enable_open_metrics: true  # <-- Required for exemplars
    resource_to_telemetry_conversion:
      enabled: true

  # OTLP exporter for traces (to Jaeger or other backend)
  otlp/traces:
    endpoint: "jaeger:4317"
    tls:
      insecure: true

# =============================================================================
# EXTENSIONS
# =============================================================================
extensions:
  health_check:
    endpoint: "0.0.0.0:13133"
  zpages:
    endpoint: "0.0.0.0:55679"
  pprof:
    endpoint: "0.0.0.0:1777"

# =============================================================================
# SERVICE - Pipeline Configuration
# =============================================================================
service:
  extensions: [health_check, zpages, pprof]

  pipelines:
    # Traces pipeline - receives traces, exports to Jaeger AND connectors
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [debug, otlp/traces, spanmetrics, servicegraph]

    # Direct metrics pipeline - receives OTLP metrics
    metrics:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [debug, prometheus]

    # Derived metrics from spanmetrics connector (with exemplars)
    metrics/spanmetrics:
      receivers: [spanmetrics]
      processors: [memory_limiter, batch]
      exporters: [prometheus]

    # Derived metrics from servicegraph connector
    metrics/servicegraph:
      receivers: [servicegraph]
      processors: [memory_limiter, batch]
      exporters: [prometheus]

    # Logs pipeline
    logs:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [debug]

  telemetry:
    logs:
      level: info
      encoding: json
    metrics:
      level: detailed
      readers:
        - pull:
            exporter:
              prometheus:
                host: "0.0.0.0"
                port: 8888
```

## Metrics Generated by spanmetrics Connector

The `spanmetrics` connector generates these metrics with exemplars:

| Metric | Type | Description |
|--------|------|-------------|
| `traces_spanmetrics_duration_milliseconds` | Histogram | Request duration with exemplars |
| `traces_spanmetrics_calls_total` | Counter | Total number of calls |

Each metric includes dimensions from span attributes:
- `service.name`
- `span.name`
- `http.method`
- `http.status_code`
- `http.route`
- etc.

## Prometheus Configuration

Configure Prometheus to scrape exemplars:

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'tfo-collector'
    scrape_interval: 15s
    static_configs:
      - targets: ['tfo-collector:8889']
    # Enable exemplar scraping
    honor_labels: true
```

## Grafana Configuration

### 1. Add Data Sources

Configure both Prometheus and your trace backend (Jaeger/Tempo):

**Prometheus Data Source:**
```yaml
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    url: http://prometheus:9090
    access: proxy
    jsonData:
      exemplarTraceIdDestinations:
        - name: trace_id
          datasourceUid: jaeger  # Link to trace datasource
```

**Jaeger Data Source:**
```yaml
datasources:
  - name: Jaeger
    type: jaeger
    url: http://jaeger:16686
    access: proxy
    uid: jaeger
```

### 2. Create Dashboard with Exemplars

Example panel configuration:

```json
{
  "title": "Request Duration",
  "type": "timeseries",
  "datasource": "Prometheus",
  "targets": [
    {
      "expr": "histogram_quantile(0.99, sum(rate(traces_spanmetrics_duration_milliseconds_bucket{service_name=\"my-service\"}[5m])) by (le))",
      "legendFormat": "p99 latency",
      "exemplar": true  // Enable exemplars
    }
  ],
  "fieldConfig": {
    "defaults": {
      "links": [
        {
          "title": "View Trace",
          "url": "/explore?orgId=1&left=[\"now-1h\",\"now\",\"Jaeger\",{\"query\":\"${__data.fields.traceID}\"}]",
          "targetBlank": true
        }
      ]
    }
  }
}
```

## Service Graph Visualization

The `servicegraph` connector generates metrics for service dependencies:

| Metric | Description |
|--------|-------------|
| `traces_service_graph_request_total` | Total requests between services |
| `traces_service_graph_request_failed_total` | Failed requests |
| `traces_service_graph_request_duration_seconds` | Request duration histogram |

### Grafana Service Graph Panel

```json
{
  "type": "nodeGraph",
  "title": "Service Dependencies",
  "datasource": "Prometheus",
  "targets": [
    {
      "expr": "sum by (client, server) (rate(traces_service_graph_request_total[5m]))",
      "format": "table"
    }
  ]
}
```

## Docker Compose Example

Complete example with collector, Prometheus, Jaeger, and Grafana:

```yaml
version: '3.8'

services:
  tfo-collector:
    image: telemetryflow/telemetryflow-collector-ocb:latest
    ports:
      - "4317:4317"   # OTLP gRPC
      - "4318:4318"   # OTLP HTTP
      - "8888:8888"   # Self metrics
      - "8889:8889"   # Prometheus exporter
      - "13133:13133" # Health check
    volumes:
      - ./configs/exemplars-config.yaml:/etc/tfo-collector/otel-collector.yaml:ro

  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686" # UI
      - "14250:14250" # gRPC

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--enable-feature=exemplar-storage'  # Enable exemplar storage

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_AUTH_ANONYMOUS_ENABLED=true
      - GF_AUTH_ANONYMOUS_ORG_ROLE=Admin
    volumes:
      - ./grafana/datasources.yaml:/etc/grafana/provisioning/datasources/datasources.yaml
```

## Troubleshooting

### Exemplars Not Appearing

1. **Check OpenMetrics format is enabled:**
   ```yaml
   exporters:
     prometheus:
       enable_open_metrics: true
   ```

2. **Verify exemplars are enabled in spanmetrics:**
   ```yaml
   connectors:
     spanmetrics:
       exemplars:
         enabled: true
   ```

3. **Check Prometheus has exemplar storage enabled:**
   ```bash
   prometheus --enable-feature=exemplar-storage
   ```

4. **Verify metrics are being scraped with exemplars:**
   ```bash
   curl -H "Accept: application/openmetrics-text" http://localhost:8889/metrics | grep -A5 "duration_milliseconds"
   ```

### Service Graph Missing Nodes

1. Ensure `span.kind` is set correctly (CLIENT/SERVER)
2. Check that both client and server spans have `service.name` attribute
3. Verify `servicegraph` connector is in the traces pipeline exporters

## Related Documentation

- [OpenTelemetry Exemplars Spec](https://opentelemetry.io/docs/specs/otel/metrics/data-model/#exemplars)
- [Prometheus Exemplars](https://prometheus.io/docs/prometheus/latest/feature_flags/#exemplars-storage)
- [Grafana Exemplars](https://grafana.com/docs/grafana/latest/fundamentals/exemplars/)
- [OCB Build Guide](./OCB_BUILD.md)
- [Component Reference](./COMPONENTS.md)
- [Configuration Guide](./CONFIGURATION.md)
