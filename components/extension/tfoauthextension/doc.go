// Package tfoauthextension provides centralized TelemetryFlow API key management.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// The tfoauthextension provides:
//   - Centralized API key storage for TFO authentication
//   - API key validation (optional)
//   - Credential provider interface for tfoexporter
//
// Configuration example:
//
//	extensions:
//	  tfoauth:
//	    api_key_id: "${env:TELEMETRYFLOW_API_KEY_ID}"
//	    api_key_secret: "${env:TELEMETRYFLOW_API_KEY_SECRET}"
//	    validation_endpoint: "https://api.telemetryflow.id/v1/auth/validate"
package tfoauthextension // import "github.com/telemetryflow/telemetryflow-collector/components/extension/tfoauthextension"
