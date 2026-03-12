// Package tfoexporter provides a TelemetryFlow Platform exporter that automatically
// injects TFO authentication headers and supports the TFO v2 API.
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
