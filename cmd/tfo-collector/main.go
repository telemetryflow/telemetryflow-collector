// Package main is the entry point for the TelemetryFlow Collector.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// This is the OCB-native build that includes:
// - All 85+ OpenTelemetry Collector community components
// - TFO custom components (tfootlpreceiver, tfoexporter, tfoauth, tfoidentity)
// - TFO branding and CLI experience
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
package main

import (
	"fmt"
	"log"
	"os"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/otelcol"

	"github.com/telemetryflow/telemetryflow-collector/internal/version"
)

func main() {
	info := component.BuildInfo{
		Command:     version.ProductShortName,
		Description: version.ProductDescription,
		Version:     version.Version,
	}

	// Configure the collector settings
	// Pass the components function directly (not called) - OTEL 0.142.0 API
	set := otelcol.CollectorSettings{
		BuildInfo: info,
		Factories: components,
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				ProviderFactories: []confmap.ProviderFactory{
					fileprovider.NewFactory(),
					yamlprovider.NewFactory(),
					envprovider.NewFactory(),
				},
			},
		},
	}

	// Create and run the collector
	cmd := otelcol.NewCommand(set)
	cmd.Use = version.ProductShortName
	cmd.Short = version.ProductName
	cmd.Long = fmt.Sprintf(`%s

%s

Usage Examples:
  # Start with default config
  %s --config /etc/tfo-collector/config.yaml

  # Start with TFO Platform config (supports v2 endpoints)
  %s --config configs/tfo-collector.yaml

  # Start with multiple configs
  %s --config base.yaml --config overrides.yaml

TFO Custom Components:
  Receivers:
    tfootlp   - OTLP receiver with v1 and v2 endpoint support

  Exporters:
    tfo       - TFO Platform exporter with auto-auth injection

  Extensions:
    tfoauth     - TFO API key management
    tfoidentity - Collector identity and resource enrichment

Environment Variables:
  TELEMETRYFLOW_API_KEY_ID      - TFO API Key ID (tfk_xxx)
  TELEMETRYFLOW_API_KEY_SECRET  - TFO API Key Secret (tfs_xxx)
  TELEMETRYFLOW_ENDPOINT        - TFO Platform endpoint
  TELEMETRYFLOW_COLLECTOR_ID    - Unique collector identifier
  TELEMETRYFLOW_COLLECTOR_NAME  - Human-readable collector name
  TELEMETRYFLOW_ENVIRONMENT     - Deployment environment

For more information, visit: %s`,
		version.ProductName,
		version.Motto,
		version.ProductShortName,
		version.ProductShortName,
		version.ProductShortName,
		version.SupportURL,
	)

	// Add short flag aliases before execution
	if f := cmd.Flags().Lookup("config"); f != nil {
		f.Shorthand = "c"
	}
	if f := cmd.Flags().Lookup("set"); f != nil {
		f.Shorthand = "s"
	}
	if f := cmd.Flags().Lookup("feature-gates"); f != nil {
		f.Shorthand = "f"
	}

	// Show banner and help if no arguments provided
	if len(os.Args) == 1 {
		fmt.Print(version.Banner())
		cmd.Help()
		return
	}

	// Show banner for help/version commands
	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" || arg == "--version" || arg == "-v" {
			fmt.Print(version.Banner())
			break
		}
	}

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
