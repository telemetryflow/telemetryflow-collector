// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package tfoexporter_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/exporter/exportertest"

	"github.com/telemetryflow/telemetryflow-collector/components/tfoexporter"
)

func TestNewFactory(t *testing.T) {
	factory := tfoexporter.NewFactory()

	assert.NotNil(t, factory)
	assert.Equal(t, component.MustNewType("tfo"), factory.Type())
}

func TestCreateDefaultConfig(t *testing.T) {
	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig()

	require.NotNil(t, cfg)
	assert.IsType(t, &tfoexporter.Config{}, cfg)

	oCfg := cfg.(*tfoexporter.Config)

	// Check default settings
	assert.True(t, oCfg.UseV2API)
	assert.Equal(t, tfoexporter.DefaultEndpoint, oCfg.Endpoint)

	// Check retry config defaults
	assert.True(t, oCfg.RetryConfig.Enabled)
}

func TestFactory_CreateTracesExporter(t *testing.T) {
	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)

	// Add auth for valid config
	cfg.Auth = &tfoexporter.AuthConfig{
		APIKeyID:     configopaque.String("tfk_test_key"),
		APIKeySecret: configopaque.String("tfs_test_secret"),
	}

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))

	exporter, err := factory.CreateTraces(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NotNil(t, exporter)
}

func TestFactory_CreateMetricsExporter(t *testing.T) {
	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)

	// Add auth for valid config
	cfg.Auth = &tfoexporter.AuthConfig{
		APIKeyID:     configopaque.String("tfk_test_key"),
		APIKeySecret: configopaque.String("tfs_test_secret"),
	}

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))

	exporter, err := factory.CreateMetrics(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NotNil(t, exporter)
}

func TestFactory_CreateLogsExporter(t *testing.T) {
	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)

	// Add auth for valid config
	cfg.Auth = &tfoexporter.AuthConfig{
		APIKeyID:     configopaque.String("tfk_test_key"),
		APIKeySecret: configopaque.String("tfs_test_secret"),
	}

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))

	exporter, err := factory.CreateLogs(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NotNil(t, exporter)
}

func TestFactory_StabilityLevel(t *testing.T) {
	factory := tfoexporter.NewFactory()

	// All signal types should be stable
	assert.Equal(t, component.StabilityLevelStable, factory.TracesStability())
	assert.Equal(t, component.StabilityLevelStable, factory.MetricsStability())
	assert.Equal(t, component.StabilityLevelStable, factory.LogsStability())
}

func TestFactory_TypeConstant(t *testing.T) {
	assert.Equal(t, "tfo", tfoexporter.TypeStr)
	assert.Equal(t, "https://api.telemetryflow.id", tfoexporter.DefaultEndpoint)
}
