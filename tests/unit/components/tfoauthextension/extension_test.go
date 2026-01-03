// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package tfoauthextension_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/extension/extensiontest"

	"github.com/telemetryflow/telemetryflow-collector/components/extension/tfoauthextension"
)

func TestNewFactory(t *testing.T) {
	factory := tfoauthextension.NewFactory()

	assert.NotNil(t, factory)
	assert.Equal(t, component.MustNewType("tfoauth"), factory.Type())
}

func TestCreateDefaultConfig(t *testing.T) {
	factory := tfoauthextension.NewFactory()
	cfg := factory.CreateDefaultConfig()

	require.NotNil(t, cfg)
	assert.IsType(t, &tfoauthextension.Config{}, cfg)

	oCfg := cfg.(*tfoauthextension.Config)
	assert.Empty(t, string(oCfg.APIKeyID))
	assert.Empty(t, string(oCfg.APIKeySecret))
	assert.False(t, oCfg.ValidateOnStart)
	assert.Empty(t, oCfg.ValidationEndpoint)
}

func TestFactory_CreateExtension(t *testing.T) {
	factory := tfoauthextension.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoauthextension.Config)

	// Set required config
	cfg.APIKeyID = configopaque.String("tfk_test_key_12345")
	cfg.APIKeySecret = configopaque.String("tfs_test_secret_12345")

	set := extensiontest.NewNopSettings(component.MustNewType("tfoauth"))

	ext, err := factory.Create(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NotNil(t, ext)
}

func TestExtension_Lifecycle(t *testing.T) {
	factory := tfoauthextension.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoauthextension.Config)

	cfg.APIKeyID = configopaque.String("tfk_test_key_12345")
	cfg.APIKeySecret = configopaque.String("tfs_test_secret_12345")

	set := extensiontest.NewNopSettings(component.MustNewType("tfoauth"))

	ext, err := factory.Create(context.Background(), set, cfg)
	require.NoError(t, err)

	// Test start
	err = ext.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	// Test shutdown
	err = ext.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestExtension_GetCredentials(t *testing.T) {
	factory := tfoauthextension.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoauthextension.Config)

	cfg.APIKeyID = configopaque.String("tfk_test_key_12345")
	cfg.APIKeySecret = configopaque.String("tfs_test_secret_12345")

	set := extensiontest.NewNopSettings(component.MustNewType("tfoauth"))

	ext, err := factory.Create(context.Background(), set, cfg)
	require.NoError(t, err)

	err = ext.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	// Test that extension was created successfully
	assert.NotNil(t, ext)

	err = ext.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestFactory_StabilityLevel(t *testing.T) {
	factory := tfoauthextension.NewFactory()
	assert.Equal(t, component.StabilityLevelStable, factory.Stability())
}

func TestFactory_TypeConstant(t *testing.T) {
	assert.Equal(t, "tfoauth", tfoauthextension.TypeStr)
}
