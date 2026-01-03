// Package tfoidentityextension provides collector identity management and resource enrichment.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// The tfoidentityextension provides:
//   - Collector identification (ID, hostname, name)
//   - Custom tags for labeling and filtering
//   - Resource enrichment for telemetry data
//   - Identity provider interface for tfoexporter
//
// Configuration example:
//
//	extensions:
//	  tfoidentity:
//	    id: "${env:TELEMETRYFLOW_COLLECTOR_ID}"
//	    name: "Production Edge Collector"
//	    tags:
//	      environment: production
//	      datacenter: us-west-2
//	    enrich_resources: true
package tfoidentityextension // import "github.com/telemetryflow/telemetryflow-collector/components/extension/tfoidentityextension"
