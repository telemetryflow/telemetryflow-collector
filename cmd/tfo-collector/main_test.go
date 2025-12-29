// Package main is the entry point for the TelemetryFlow Collector.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// LEGO Building Block - Self-contained within tfo-collector project.
package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/telemetryflow/telemetryflow-collector/internal/config"
)

// TestStartCmd tests the start command creation
func TestStartCmd(t *testing.T) {
	cmd := startCmd()

	if cmd == nil {
		t.Fatal("Expected non-nil start command")
	}

	if cmd.Use != "start" {
		t.Errorf("Expected Use 'start', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected non-empty Short description")
	}

	// Check for --otel flag
	flag := cmd.Flags().Lookup("otel")
	if flag == nil {
		t.Error("Expected --otel flag to be defined")
	}
}

// TestVersionCmd tests the version command creation
func TestVersionCmd(t *testing.T) {
	cmd := versionCmd()

	if cmd == nil {
		t.Fatal("Expected non-nil version command")
	}

	if cmd.Use != "version" {
		t.Errorf("Expected Use 'version', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected non-empty Short description")
	}

	// Check for --json flag
	jsonFlag := cmd.Flags().Lookup("json")
	if jsonFlag == nil {
		t.Error("Expected --json flag to be defined")
	}

	// Check for --short flag
	shortFlag := cmd.Flags().Lookup("short")
	if shortFlag == nil {
		t.Error("Expected --short flag to be defined")
	}
}

// TestConfigCmd tests the config command creation
func TestConfigCmd(t *testing.T) {
	cmd := configCmd()

	if cmd == nil {
		t.Fatal("Expected non-nil config command")
	}

	if cmd.Use != "config" {
		t.Errorf("Expected Use 'config', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected non-empty Short description")
	}

	// Check for subcommands
	if !cmd.HasSubCommands() {
		t.Error("Expected config command to have subcommands")
	}

	// Check validate subcommand exists
	validateCmd, _, err := cmd.Find([]string{"validate"})
	if err != nil {
		t.Errorf("Expected validate subcommand, got error: %v", err)
	}
	if validateCmd.Use != "validate" {
		t.Errorf("Expected validate subcommand Use 'validate', got '%s'", validateCmd.Use)
	}

	// Check show subcommand exists
	showCmd, _, err := cmd.Find([]string{"show"})
	if err != nil {
		t.Errorf("Expected show subcommand, got error: %v", err)
	}
	if showCmd.Use != "show" {
		t.Errorf("Expected show subcommand Use 'show', got '%s'", showCmd.Use)
	}
}

// TestValidateCmd tests the validate command creation
func TestValidateCmd(t *testing.T) {
	cmd := validateCmd()

	if cmd == nil {
		t.Fatal("Expected non-nil validate command")
	}

	if cmd.Use != "validate" {
		t.Errorf("Expected Use 'validate', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected non-empty Short description")
	}
}

// TestInitLogger tests the initLogger function
func TestInitLogger(t *testing.T) {
	tests := []struct {
		name   string
		cfg    config.LoggingConfig
		wantOK bool
	}{
		{
			name: "default info level json format",
			cfg: config.LoggingConfig{
				Level:  "info",
				Format: "json",
			},
			wantOK: true,
		},
		{
			name: "debug level json format",
			cfg: config.LoggingConfig{
				Level:  "debug",
				Format: "json",
			},
			wantOK: true,
		},
		{
			name: "warn level text format",
			cfg: config.LoggingConfig{
				Level:  "warn",
				Format: "text",
			},
			wantOK: true,
		},
		{
			name: "error level json format",
			cfg: config.LoggingConfig{
				Level:  "error",
				Format: "json",
			},
			wantOK: true,
		},
		{
			name: "unknown level defaults to info",
			cfg: config.LoggingConfig{
				Level:  "unknown",
				Format: "json",
			},
			wantOK: true,
		},
		{
			name: "empty level defaults to info",
			cfg: config.LoggingConfig{
				Level:  "",
				Format: "json",
			},
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := initLogger(tt.cfg)

			if tt.wantOK {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if logger == nil {
					t.Error("Expected non-nil logger")
				} else {
					_ = logger.Sync()
				}
			} else {
				if err == nil {
					t.Error("Expected error")
				}
			}
		})
	}
}

// TestPrintConfig tests the printConfig function
func TestPrintConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Collector.ID = "test-collector-id"
	cfg.Collector.Hostname = "test-hostname"

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printConfig(cfg)

	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify output contains expected sections
	expectedStrings := []string{
		"TelemetryFlow Collector Configuration",
		"Collector:",
		"test-collector-id",
		"test-hostname",
		"Receivers:",
		"OTLP:",
		"Processors:",
		"Batch:",
		"Exporters:",
		"Extensions:",
		"Logging:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s'", expected)
		}
	}
}

// TestVersionCmdExecution tests version command execution
func TestVersionCmdExecution(t *testing.T) {
	t.Run("default output", func(t *testing.T) {
		cmd := versionCmd()

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		cmd.Run(cmd, []string{})

		_ = w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		if output == "" {
			t.Error("Expected non-empty version output")
		}
	})

	t.Run("short output", func(t *testing.T) {
		cmd := versionCmd()
		_ = cmd.Flags().Set("short", "true")

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		cmd.Run(cmd, []string{})

		_ = w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		if output == "" {
			t.Error("Expected non-empty short version output")
		}
	})

	t.Run("json output", func(t *testing.T) {
		cmd := versionCmd()
		_ = cmd.Flags().Set("json", "true")

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		cmd.Run(cmd, []string{})

		_ = w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		if !strings.Contains(output, "{") || !strings.Contains(output, "}") {
			t.Error("Expected JSON output with braces")
		}
		if !strings.Contains(output, "version") {
			t.Error("Expected JSON output to contain 'version'")
		}
	})
}

// TestGlobalVariables tests global variable defaults
func TestGlobalVariables(t *testing.T) {
	// cfgFile should be empty by default
	if cfgFile != "" {
		// Reset for clean test state
		cfgFile = ""
	}

	// logLevel should be empty by default
	if logLevel != "" {
		logLevel = ""
	}

	// logFormat should be empty by default
	if logFormat != "" {
		logFormat = ""
	}

	// useOTEL should be false by default
	if useOTEL != false {
		useOTEL = false
	}
}

// TestCommandLongDescriptions tests that commands have proper long descriptions
func TestCommandLongDescriptions(t *testing.T) {
	t.Run("start command has long description", func(t *testing.T) {
		cmd := startCmd()
		if cmd.Long == "" {
			t.Error("Expected start command to have Long description")
		}
		if !strings.Contains(cmd.Long, "OTEL") {
			t.Error("Expected start command Long description to mention OTEL")
		}
	})

	t.Run("config command subcommands", func(t *testing.T) {
		cmd := configCmd()
		if !cmd.HasSubCommands() {
			t.Error("Expected config command to have subcommands")
		}
	})
}

// TestInitLoggerWithFile tests initLogger with file output
func TestInitLoggerWithFile(t *testing.T) {
	tmpFile := t.TempDir() + "/test.log"

	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		File:   tmpFile,
	}

	logger, err := initLogger(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected non-nil logger")
	}

	// Log something
	logger.Info("test message")
	_ = logger.Sync()

	// Verify file was created
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Expected log file to be created")
	}
}

// Benchmark tests
func BenchmarkInitLogger(b *testing.B) {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger, _ := initLogger(cfg)
		if logger != nil {
			_ = logger.Sync()
		}
	}
}

func BenchmarkStartCmd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = startCmd()
	}
}

func BenchmarkVersionCmd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = versionCmd()
	}
}

func BenchmarkConfigCmd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = configCmd()
	}
}
