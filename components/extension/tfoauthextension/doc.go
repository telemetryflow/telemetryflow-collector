// Package tfoauthextension provides centralized TelemetryFlow API key management.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
