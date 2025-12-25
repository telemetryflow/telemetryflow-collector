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
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/telemetryflow/telemetryflow-collector/internal/collector"
	"github.com/telemetryflow/telemetryflow-collector/internal/config"
	"github.com/telemetryflow/telemetryflow-collector/internal/version"
)

var (
	cfgFile   string
	logLevel  string
	logFormat string
)

func main() {
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
	rootCmd.AddCommand(startCmd())
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(configCmd())
	rootCmd.AddCommand(validateCmd())

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "", "log format (json, text)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// startCmd returns the start command
func startCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the TelemetryFlow collector",
		Long:  `Start the TelemetryFlow collector and begin receiving telemetry data.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCollector()
		},
	}
}

// versionCmd returns the version command
func versionCmd() *cobra.Command {
	var jsonOutput bool
	var shortOutput bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version and license information",
		Run: func(cmd *cobra.Command, args []string) {
			if jsonOutput {
				info := version.Get()
				fmt.Printf(`{"product":"%s","version":"%s","otelVersion":"%s","gitCommit":"%s","buildTime":"%s","goVersion":"%s","os":"%s","arch":"%s","vendor":"%s","developer":"%s","license":"%s"}`+"\n",
					info.Product, info.Version, info.OTELVersion, info.GitCommit, info.BuildTime, info.GoVersion, info.OS, info.Arch, info.Vendor, info.Developer, info.License)
			} else if shortOutput {
				fmt.Println(version.OneLiner())
			} else {
				fmt.Println(version.String())
			}
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	cmd.Flags().BoolVarP(&shortOutput, "short", "s", false, "output short version")
	return cmd
}

// configCmd returns the config command
func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management commands",
	}

	// config validate subcommand
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate the configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			loader := config.NewLoader()
			cfg, err := loader.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}
			fmt.Printf("Configuration is valid\n")
			fmt.Printf("  Collector ID: %s\n", cfg.Collector.ID)
			fmt.Printf("  Hostname: %s\n", cfg.Collector.Hostname)
			fmt.Printf("  OTLP gRPC: %s\n", cfg.Receivers.OTLP.Protocols.GRPC.Endpoint)
			fmt.Printf("  OTLP HTTP: %s\n", cfg.Receivers.OTLP.Protocols.HTTP.Endpoint)
			return nil
		},
	}

	// config show subcommand
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			loader := config.NewLoader()
			cfg, err := loader.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			printConfig(cfg)
			return nil
		},
	}

	cmd.AddCommand(validateCmd, showCmd)
	return cmd
}

// validateCmd returns the validate command (top-level shortcut)
func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate the configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			loader := config.NewLoader()
			_, err := loader.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}
			fmt.Println("Configuration is valid")
			return nil
		},
	}
}

// runCollector starts the collector
func runCollector() error {
	// Load configuration
	loader := config.NewLoader()
	cfg, err := loader.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override log settings from flags
	if logLevel != "" {
		cfg.Logging.Level = logLevel
	}
	if logFormat != "" {
		cfg.Logging.Format = logFormat
	}

	// Initialize logger
	logger, err := initLogger(cfg.Logging)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer func() { _ = logger.Sync() }()

	// Print startup banner
	fmt.Print(version.Banner())

	// Log startup info
	logger.Info("Starting TelemetryFlow Collector",
		zap.String("product", version.ProductName),
		zap.String("version", version.Short()),
		zap.String("otel_version", version.OTELVersion),
		zap.String("vendor", version.Vendor),
		zap.String("developer", version.Developer),
		zap.String("hostname", cfg.Collector.Hostname),
	)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create and start collector
	c, err := collector.New(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to create collector: %w", err)
	}

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Start collector in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- c.Run(ctx)
	}()

	// Wait for signals or error
	select {
	case sig := <-sigChan:
		logger.Info("Received signal, shutting down", zap.String("signal", sig.String()))
		cancel()
		// Wait for collector to finish
		if err := <-errChan; err != nil && err != context.Canceled {
			logger.Error("Collector error during shutdown", zap.Error(err))
		}
	case err := <-errChan:
		if err != nil && err != context.Canceled {
			logger.Error("Collector error", zap.Error(err))
			return err
		}
	}

	logger.Info("TelemetryFlow Collector stopped")
	return nil
}

// initLogger initializes the logger based on configuration
func initLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	var zapCfg zap.Config
	if cfg.Format == "json" {
		zapCfg = zap.NewProductionConfig()
	} else {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	zapCfg.Level = zap.NewAtomicLevelAt(level)

	if cfg.File != "" {
		zapCfg.OutputPaths = []string{cfg.File}
		zapCfg.ErrorOutputPaths = []string{cfg.File}
	}

	return zapCfg.Build()
}

// printConfig prints the configuration summary
func printConfig(cfg *config.Config) {
	fmt.Println("TelemetryFlow Collector Configuration")
	fmt.Println("======================================")
	fmt.Printf("\nCollector:\n")
	fmt.Printf("  ID:       %s\n", cfg.Collector.ID)
	fmt.Printf("  Hostname: %s\n", cfg.Collector.Hostname)

	fmt.Printf("\nReceivers:\n")
	fmt.Printf("  OTLP:\n")
	fmt.Printf("    Enabled:     %v\n", cfg.Receivers.OTLP.Enabled)
	fmt.Printf("    gRPC:\n")
	fmt.Printf("      Enabled:   %v\n", cfg.Receivers.OTLP.Protocols.GRPC.Enabled)
	fmt.Printf("      Endpoint:  %s\n", cfg.Receivers.OTLP.Protocols.GRPC.Endpoint)
	fmt.Printf("    HTTP:\n")
	fmt.Printf("      Enabled:   %v\n", cfg.Receivers.OTLP.Protocols.HTTP.Enabled)
	fmt.Printf("      Endpoint:  %s\n", cfg.Receivers.OTLP.Protocols.HTTP.Endpoint)
	fmt.Printf("  Prometheus:\n")
	fmt.Printf("    Enabled:     %v\n", cfg.Receivers.Prometheus.Enabled)

	fmt.Printf("\nProcessors:\n")
	fmt.Printf("  Batch:\n")
	fmt.Printf("    Enabled:     %v\n", cfg.Processors.Batch.Enabled)
	fmt.Printf("    Batch Size:  %d\n", cfg.Processors.Batch.SendBatchSize)
	fmt.Printf("    Timeout:     %s\n", cfg.Processors.Batch.Timeout)
	fmt.Printf("  Memory Limiter:\n")
	fmt.Printf("    Enabled:     %v\n", cfg.Processors.Memory.Enabled)
	fmt.Printf("    Limit %%:     %d\n", cfg.Processors.Memory.LimitPercentage)

	fmt.Printf("\nExporters:\n")
	fmt.Printf("  Logging:       enabled=%v\n", cfg.Exporters.Logging.Enabled)
	fmt.Printf("  OTLP:          enabled=%v\n", cfg.Exporters.OTLP.Enabled)
	fmt.Printf("  Prometheus:    enabled=%v\n", cfg.Exporters.Prometheus.Enabled)
	fmt.Printf("  File:          enabled=%v\n", cfg.Exporters.File.Enabled)

	fmt.Printf("\nExtensions:\n")
	fmt.Printf("  Health Check:  enabled=%v, endpoint=%s\n",
		cfg.Extensions.Health.Enabled, cfg.Extensions.Health.Endpoint)
	fmt.Printf("  ZPages:        enabled=%v\n", cfg.Extensions.ZPages.Enabled)
	fmt.Printf("  PPROF:         enabled=%v\n", cfg.Extensions.PPROF.Enabled)

	fmt.Printf("\nLogging:\n")
	fmt.Printf("  Level: %s, Format: %s\n", cfg.Logging.Level, cfg.Logging.Format)
}
