// Package tfootlpreceiver provides a TelemetryFlow-enhanced OTLP receiver
// that supports both v1 (standard OTEL) and v2 (TFO Platform) HTTP endpoints.
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
