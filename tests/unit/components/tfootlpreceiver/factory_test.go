// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package tfootlpreceiver_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/telemetryflow/telemetryflow-collector/components/tfootlpreceiver"
)

func TestNewFactory(t *testing.T) {
	factory := tfootlpreceiver.NewFactory()

	assert.NotNil(t, factory)
	assert.Equal(t, component.MustNewType("tfootlp"), factory.Type())
}

func TestCreateDefaultConfig(t *testing.T) {
	factory := tfootlpreceiver.NewFactory()
	cfg := factory.CreateDefaultConfig()

	require.NotNil(t, cfg)
	assert.IsType(t, &tfootlpreceiver.Config{}, cfg)

	oCfg := cfg.(*tfootlpreceiver.Config)

	// Check default protocols
	assert.NotNil(t, oCfg.Protocols.GRPC)
	assert.NotNil(t, oCfg.Protocols.HTTP)

	// Check default v2 settings
	assert.True(t, oCfg.EnableV2Endpoints)
	assert.True(t, oCfg.V2Auth.Required)
	assert.False(t, oCfg.V2Auth.ValidateSecret)
	assert.Empty(t, oCfg.V2Auth.ValidAPIKeyIDs)

	// Check default URL paths
	assert.Equal(t, "/v1/traces", oCfg.Protocols.HTTP.TracesURLPath)
	assert.Equal(t, "/v1/metrics", oCfg.Protocols.HTTP.MetricsURLPath)
	assert.Equal(t, "/v1/logs", oCfg.Protocols.HTTP.LogsURLPath)
}

func TestFactory_CreateTracesReceiver(t *testing.T) {
	factory := tfootlpreceiver.NewFactory()
	cfg := factory.CreateDefaultConfig()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))

	receiver, err := factory.CreateTraces(context.Background(), set, cfg, consumertest.NewNop())
	require.NoError(t, err)
	require.NotNil(t, receiver)

	// Verify it implements the receiver interface
	assert.NotNil(t, receiver)
}

func TestFactory_CreateMetricsReceiver(t *testing.T) {
	factory := tfootlpreceiver.NewFactory()
	cfg := factory.CreateDefaultConfig()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))

	receiver, err := factory.CreateMetrics(context.Background(), set, cfg, consumertest.NewNop())
	require.NoError(t, err)
	require.NotNil(t, receiver)

	assert.NotNil(t, receiver)
}

func TestFactory_CreateLogsReceiver(t *testing.T) {
	factory := tfootlpreceiver.NewFactory()
	cfg := factory.CreateDefaultConfig()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))

	receiver, err := factory.CreateLogs(context.Background(), set, cfg, consumertest.NewNop())
	require.NoError(t, err)
	require.NotNil(t, receiver)

	assert.NotNil(t, receiver)
}

func TestFactory_StabilityLevel(t *testing.T) {
	factory := tfootlpreceiver.NewFactory()

	// All signal types should be stable
	assert.Equal(t, component.StabilityLevelStable, factory.TracesStability())
	assert.Equal(t, component.StabilityLevelStable, factory.MetricsStability())
	assert.Equal(t, component.StabilityLevelStable, factory.LogsStability())
}

func TestFactory_TypeConstant(t *testing.T) {
	assert.Equal(t, "tfootlp", tfootlpreceiver.TypeStr)
	assert.Equal(t, "0.0.0.0:4317", tfootlpreceiver.DefaultGRPCEndpoint)
	assert.Equal(t, "0.0.0.0:4318", tfootlpreceiver.DefaultHTTPEndpoint)
}

func TestReceiverLifecycle(t *testing.T) {
	factory := tfootlpreceiver.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfootlpreceiver.Config)

	// Disable protocols to avoid port binding issues in tests
	cfg.Protocols.GRPC = nil
	cfg.Protocols.HTTP = nil

	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	receiver, err := factory.CreateTraces(context.Background(), set, cfg, consumertest.NewNop())
	require.NoError(t, err)

	err = receiver.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	err = receiver.Shutdown(context.Background())
	require.NoError(t, err)
}
