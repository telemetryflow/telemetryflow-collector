// Package collector provides OTEL-based collector with full capabilities.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// LEGO Building Block - Self-contained within tfo-collector project.
package collector

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestNewOTELCollector tests the NewOTELCollector function
func TestNewOTELCollector(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("creates collector with valid parameters", func(t *testing.T) {
		collector, err := NewOTELCollector("/path/to/config.yaml", logger, "1.0.0")

		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if collector == nil {
			t.Fatal("Expected non-nil collector")
		}

		if collector.configPath != "/path/to/config.yaml" {
			t.Errorf("Expected configPath '/path/to/config.yaml', got '%s'", collector.configPath)
		}

		if collector.logger != logger {
			t.Error("Expected logger to be set")
		}

		if collector.version.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%s'", collector.version.Version)
		}
	})

	t.Run("sets correct build info", func(t *testing.T) {
		collector, _ := NewOTELCollector("/config.yaml", logger, "2.0.0")

		if collector.version.Command != "tfo-collector" {
			t.Errorf("Expected Command 'tfo-collector', got '%s'", collector.version.Command)
		}

		if collector.version.Description != "TelemetryFlow Collector - Community Enterprise Observability Platform" {
			t.Errorf("Unexpected Description: %s", collector.version.Description)
		}

		if collector.version.Version != "2.0.0" {
			t.Errorf("Expected Version '2.0.0', got '%s'", collector.version.Version)
		}
	})

	t.Run("handles empty config path", func(t *testing.T) {
		collector, err := NewOTELCollector("", logger, "1.0.0")

		if err != nil {
			t.Fatalf("Expected no error for empty config path, got: %v", err)
		}

		if collector.configPath != "" {
			t.Errorf("Expected empty configPath, got '%s'", collector.configPath)
		}
	})

	t.Run("handles nil logger", func(t *testing.T) {
		collector, err := NewOTELCollector("/config.yaml", nil, "1.0.0")

		if err != nil {
			t.Fatalf("Expected no error for nil logger, got: %v", err)
		}

		if collector.logger != nil {
			t.Error("Expected nil logger")
		}
	})
}

// TestOTELCollectorComponents tests the components method
func TestOTELCollectorComponents(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	collector, _ := NewOTELCollector("/config.yaml", logger, "1.0.0")

	factories, err := collector.components()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	t.Run("has extension factories", func(t *testing.T) {
		if len(factories.Extensions) == 0 {
			t.Error("Expected extension factories to be registered")
		}

		// Check for specific extensions by iterating over registered factories
		expectedExtensions := map[string]bool{"zpages": false, "health_check": false, "pprof": false}
		for compType := range factories.Extensions {
			expectedExtensions[compType.String()] = true
		}
		for ext, found := range expectedExtensions {
			if !found {
				t.Errorf("Expected extension '%s' to be registered", ext)
			}
		}
	})

	t.Run("has receiver factories", func(t *testing.T) {
		if len(factories.Receivers) == 0 {
			t.Error("Expected receiver factories to be registered")
		}

		// Check for specific receivers
		expectedReceivers := map[string]bool{"otlp": false, "jaeger": false, "zipkin": false, "hostmetrics": false, "prometheus": false, "filelog": false}
		for compType := range factories.Receivers {
			expectedReceivers[compType.String()] = true
		}
		for rcv, found := range expectedReceivers {
			if !found {
				t.Errorf("Expected receiver '%s' to be registered", rcv)
			}
		}
	})

	t.Run("has processor factories", func(t *testing.T) {
		if len(factories.Processors) == 0 {
			t.Error("Expected processor factories to be registered")
		}

		// Check for specific processors
		expectedProcessors := map[string]bool{"batch": false, "memory_limiter": false, "attributes": false, "resource": false, "resourcedetection": false, "filter": false, "transform": false, "tail_sampling": false}
		for compType := range factories.Processors {
			expectedProcessors[compType.String()] = true
		}
		for proc, found := range expectedProcessors {
			if !found {
				t.Errorf("Expected processor '%s' to be registered", proc)
			}
		}
	})

	t.Run("has exporter factories", func(t *testing.T) {
		if len(factories.Exporters) == 0 {
			t.Error("Expected exporter factories to be registered")
		}

		// Check for specific exporters
		expectedExporters := map[string]bool{"otlp": false, "otlphttp": false, "debug": false, "prometheus": false, "prometheusremotewrite": false, "file": false}
		for compType := range factories.Exporters {
			expectedExporters[compType.String()] = true
		}
		for exp, found := range expectedExporters {
			if !found {
				t.Errorf("Expected exporter '%s' to be registered", exp)
			}
		}
	})

	t.Run("has connector factories", func(t *testing.T) {
		if len(factories.Connectors) == 0 {
			t.Error("Expected connector factories to be registered")
		}

		// Check for specific connectors
		expectedConnectors := map[string]bool{"forward": false, "spanmetrics": false, "servicegraph": false, "count": false}
		for compType := range factories.Connectors {
			expectedConnectors[compType.String()] = true
		}
		for conn, found := range expectedConnectors {
			if !found {
				t.Errorf("Expected connector '%s' to be registered", conn)
			}
		}
	})
}

// TestOTELCollectorShutdown tests the Shutdown method
func TestOTELCollectorShutdown(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("shutdown without running service", func(t *testing.T) {
		collector, _ := NewOTELCollector("/config.yaml", logger, "1.0.0")

		// Should not panic when service is nil
		collector.Shutdown()
	})
}

// TestOTELCollectorRunWithInvalidConfig tests Run with invalid config
func TestOTELCollectorRunWithInvalidConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	collector, _ := NewOTELCollector("/nonexistent/config.yaml", logger, "1.0.0")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := collector.Run(ctx)
	if err == nil {
		t.Error("Expected error for nonexistent config file")
	}
}

// TestOTELCollectorStruct tests OTELCollector struct initialization
func TestOTELCollectorStruct(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("struct fields are set correctly", func(t *testing.T) {
		collector, _ := NewOTELCollector("/test/config.yaml", logger, "3.0.0")

		if collector.configPath != "/test/config.yaml" {
			t.Errorf("Expected configPath '/test/config.yaml', got '%s'", collector.configPath)
		}

		if collector.logger != logger {
			t.Error("Expected logger to match")
		}

		if collector.version.Version != "3.0.0" {
			t.Errorf("Expected version '3.0.0', got '%s'", collector.version.Version)
		}

		if collector.service != nil {
			t.Error("Expected service to be nil before Run")
		}
	})
}

// TestOTELCollectorVersionInfo tests version information
func TestOTELCollectorVersionInfo(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	versions := []string{"0.0.1", "1.0.0", "1.2.3-beta", "2.0.0-rc1"}

	for _, version := range versions {
		t.Run("version "+version, func(t *testing.T) {
			collector, _ := NewOTELCollector("/config.yaml", logger, version)

			if collector.version.Version != version {
				t.Errorf("Expected version '%s', got '%s'", version, collector.version.Version)
			}
		})
	}
}

// Benchmark tests
func BenchmarkNewOTELCollector(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewOTELCollector("/config.yaml", logger, "1.0.0")
	}
}

func BenchmarkOTELCollectorComponents(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	collector, _ := NewOTELCollector("/config.yaml", logger, "1.0.0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = collector.components()
	}
}
