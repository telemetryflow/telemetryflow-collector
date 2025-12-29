// Package main is the entry point for the TelemetryFlow Collector.
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
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/telemetryflow/telemetryflow-collector/internal/cli"
	"github.com/telemetryflow/telemetryflow-collector/internal/version"
)

func main() {
	opts := &cli.Options{}

	rootCmd := &cobra.Command{
		Use:   "tfo-collector",
		Short: "TelemetryFlow Collector - Enterprise Observability Platform",
		Long: fmt.Sprintf(`%s
TelemetryFlow Collector is an enterprise-grade OpenTelemetry collector
that receives, processes, and exports telemetry data (metrics, logs, traces)
using the OTLP protocol.

Features:
  • OTLP receiver (gRPC and HTTP)
  • Prometheus metrics receiver
  • Batch and memory limiter processors
  • Multiple exporters (OTLP, Prometheus, file)
  • Health check and monitoring extensions
  • Graceful shutdown with signal handling
  • Cross-platform support (Linux, macOS, Windows)

  `, version.Banner()),
	}

	// Add subcommands
	rootCmd.AddCommand(cli.NewStartCmd(opts))
	rootCmd.AddCommand(cli.NewVersionCmd())
	rootCmd.AddCommand(cli.NewConfigCmd(opts))
	rootCmd.AddCommand(cli.NewValidateCmd(opts))

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&opts.CfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().StringVar(&opts.LogLevel, "log-level", "", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&opts.LogFormat, "log-format", "", "log format (json, text)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
