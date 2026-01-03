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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/service/telemetry/otelconftelemetry"

	"github.com/telemetryflow/telemetryflow-collector/internal/version"
)

func main() {
	// Create custom root command with Viper
	rootCmd := &cobra.Command{
		Use:   version.ProductShortName,
		Short: version.ProductName,
		Long: fmt.Sprintf(`%s

%s

Usage Examples:
  # Start with default config
  %s --config /etc/tfo-collector/config.yaml
  %s -c /etc/tfo-collector/config.yaml

  # Start with TFO Platform config (supports v2 endpoints)
  %s --config configs/tfo-collector.yaml
  %s -c configs/tfo-collector.yaml

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
			version.ProductShortName,
			version.SupportURL,
		),
		Run: runCollector,
	}

	// Add flags with short aliases using Viper
	rootCmd.Flags().StringSliceP("config", "c", []string{}, "Locations to the config file(s)")
	rootCmd.Flags().StringSliceP("set", "s", []string{}, "Set arbitrary component config property")
	rootCmd.Flags().StringSliceP("feature-gates", "f", []string{}, "Comma-delimited list of feature gate identifiers")

	// Bind flags to Viper
	if err := viper.BindPFlags(rootCmd.Flags()); err != nil {
		log.Fatal(err)
	}

	// Show banner and help if no arguments provided
	if len(os.Args) == 1 {
		fmt.Print(version.Banner())
		if err := rootCmd.Help(); err != nil {
			log.Fatal(err)
		}
		return
	}

	// Show banner for help/version commands
	for _, arg := range os.Args[1:] {
		if arg == "--help" || arg == "-h" || arg == "--version" || arg == "-v" {
			fmt.Print(version.Banner())
			break
		}
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runCollector(cmd *cobra.Command, args []string) {
	// Show banner when starting the collector
	fmt.Print(version.Banner())

	info := component.BuildInfo{
		Command:     version.ProductShortName,
		Description: version.ProductDescription,
		Version:     version.Version,
	}

	// Factories function that returns all component factories including telemetry
	factoriesFunc := func() (otelcol.Factories, error) {
		factories, err := components()
		if err != nil {
			return otelcol.Factories{}, err
		}
		factories.Telemetry = otelconftelemetry.NewFactory()
		return factories, nil
	}

	set := otelcol.CollectorSettings{
		BuildInfo: info,
		Factories: factoriesFunc,
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

	// Get config files from Viper
	configFiles := viper.GetStringSlice("config")
	if len(configFiles) == 0 {
		log.Fatal("at least one config file must be provided")
	}

	// Create OTEL collector command with config
	otelCmd := otelcol.NewCommand(set)
	// Pass config files to OTEL collector
	os.Args = append([]string{os.Args[0]}, "--config")
	os.Args = append(os.Args, configFiles...)

	if err := otelCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
