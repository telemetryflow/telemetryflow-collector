// Package tfoidentityextension provides collector identity management and resource enrichment.
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
