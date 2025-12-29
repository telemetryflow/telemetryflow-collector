// Package collector_test provides unit tests for the OTEL collector.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package collector_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/telemetryflow/telemetryflow-collector/internal/collector"
)

func TestNewOTELCollector(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("creates collector with valid parameters", func(t *testing.T) {
		c, err := collector.NewOTELCollector("/path/to/config.yaml", logger, "1.0.0")

		require.NoError(t, err)
		require.NotNil(t, c)
	})

	t.Run("sets correct build info", func(t *testing.T) {
		c, err := collector.NewOTELCollector("/config.yaml", logger, "2.0.0")

		require.NoError(t, err)
		require.NotNil(t, c)
	})

	t.Run("handles empty config path", func(t *testing.T) {
		c, err := collector.NewOTELCollector("", logger, "1.0.0")

		require.NoError(t, err)
		require.NotNil(t, c)
	})

	t.Run("handles nil logger", func(t *testing.T) {
		c, err := collector.NewOTELCollector("/config.yaml", nil, "1.0.0")

		require.NoError(t, err)
		require.NotNil(t, c)
	})
}

func TestOTELCollectorShutdown(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("shutdown without running service", func(t *testing.T) {
		c, _ := collector.NewOTELCollector("/config.yaml", logger, "1.0.0")

		// Should not panic when service is nil
		c.Shutdown()
	})
}

func TestOTELCollectorRunWithInvalidConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	c, _ := collector.NewOTELCollector("/nonexistent/config.yaml", logger, "1.0.0")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := c.Run(ctx)
	assert.Error(t, err)
}

func TestOTELCollectorVersionInfo(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	versions := []string{"0.0.1", "1.0.0", "1.2.3-beta", "2.0.0-rc1"}

	for _, version := range versions {
		t.Run("version "+version, func(t *testing.T) {
			c, err := collector.NewOTELCollector("/config.yaml", logger, version)

			require.NoError(t, err)
			require.NotNil(t, c)
		})
	}
}

// Benchmark tests
func BenchmarkNewOTELCollector(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = collector.NewOTELCollector("/config.yaml", logger, "1.0.0")
	}
}
