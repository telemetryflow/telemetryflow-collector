// Package tfootlpreceiver provides a TelemetryFlow-enhanced OTLP receiver
// that supports both v1 (standard OTEL) and v2 (TFO Platform) HTTP endpoints.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// The tfootlpreceiver extends the standard OTLP receiver to provide:
//   - v1 endpoints: /v1/traces, /v1/metrics, /v1/logs (OTEL standard)
//   - v2 endpoints: /v2/traces, /v2/metrics, /v2/logs (TFO Platform)
//   - Both endpoints served on the same port (4318)
//   - Full gRPC support on port 4317
//
// Configuration example:
//
//	receivers:
//	  tfootlp:
//	    protocols:
//	      grpc:
//	        endpoint: "0.0.0.0:4317"
//	      http:
//	        endpoint: "0.0.0.0:4318"
//	    enable_v2_endpoints: true
package tfootlpreceiver // import "github.com/telemetryflow/telemetryflow-collector/components/tfootlpreceiver"
