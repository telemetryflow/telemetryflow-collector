// Package tfoexporter provides a TelemetryFlow Platform exporter that automatically
// injects TFO authentication headers and supports the TFO v2 API.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// The tfoexporter provides:
//   - Automatic injection of TFO authentication headers
//   - Support for both self-hosted and cloud SaaS endpoints
//   - v2 API endpoint support
//   - Integration with tfoauth and tfoidentity extensions
//
// Configuration example:
//
//	exporters:
//	  tfo:
//	    endpoint: "https://api.telemetryflow.id"
//	    use_v2_api: true
//	    auth:
//	      extension: tfoauth
//	    collector_identity: tfoidentity
//	    retry_on_failure:
//	      enabled: true
package tfoexporter // import "github.com/telemetryflow/telemetryflow-collector/components/tfoexporter"
