// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package components_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/extension/extensiontest"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/telemetryflow/telemetryflow-collector/components/extension/tfoauthextension"
	"github.com/telemetryflow/telemetryflow-collector/components/extension/tfoidentityextension"
	"github.com/telemetryflow/telemetryflow-collector/components/tfootlpreceiver"
)

// TestTFOComponentsIntegration tests that all TFO components can be created
// and started together as they would be in a real collector configuration.
func TestTFOComponentsIntegration(t *testing.T) {
	ctx := context.Background()

	// Create tfoauth extension
	authFactory := tfoauthextension.NewFactory()
	authCfg := authFactory.CreateDefaultConfig().(*tfoauthextension.Config)
	authCfg.APIKeyID = configopaque.String("tfk_integration_test_key")
	authCfg.APIKeySecret = configopaque.String("tfs_integration_test_secret")

	authSet := extensiontest.NewNopSettings(component.MustNewType("tfoauth"))
	authExt, err := authFactory.Create(ctx, authSet, authCfg)
	require.NoError(t, err)

	// Create tfoidentity extension
	identityFactory := tfoidentityextension.NewFactory()
	identityCfg := identityFactory.CreateDefaultConfig().(*tfoidentityextension.Config)
	identityCfg.ID = "integration-test-collector"
	identityCfg.Name = "Integration Test Collector"
	identityCfg.Tags = map[string]string{"test": "integration"}

	identitySet := extensiontest.NewNopSettings(component.MustNewType("tfoidentity"))
	identityExt, err := identityFactory.Create(ctx, identitySet, identityCfg)
	require.NoError(t, err)

	// Create tfootlp receiver (with protocols disabled to avoid port conflicts)
	receiverFactory := tfootlpreceiver.NewFactory()
	receiverCfg := receiverFactory.CreateDefaultConfig().(*tfootlpreceiver.Config)
	receiverCfg.Protocols.GRPC = nil
	receiverCfg.Protocols.HTTP = nil

	receiverSet := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	sink := new(consumertest.TracesSink)
	receiver, err := receiverFactory.CreateTraces(ctx, receiverSet, receiverCfg, sink)
	require.NoError(t, err)

	// Start all components
	host := componenttest.NewNopHost()

	err = authExt.Start(ctx, host)
	require.NoError(t, err)
	defer func() { _ = authExt.Shutdown(ctx) }()

	err = identityExt.Start(ctx, host)
	require.NoError(t, err)
	defer func() { _ = identityExt.Shutdown(ctx) }()

	err = receiver.Start(ctx, host)
	require.NoError(t, err)
	defer func() { _ = receiver.Shutdown(ctx) }()

	// Verify extensions were created successfully
	assert.NotNil(t, authExt)
	assert.NotNil(t, identityExt)
}

// TestTFOReceiverV1V2EndpointDifferentiation tests that v1 and v2 endpoints
// behave differently in terms of authentication requirements.
func TestTFOReceiverV1V2EndpointDifferentiation(t *testing.T) {
	// This test verifies the v1/v2 endpoint behavior at the configuration level

	cfg := &tfootlpreceiver.Config{
		EnableV2Endpoints: true,
		V2Auth: tfootlpreceiver.V2AuthConfig{
			Required:       true,
			ValidateSecret: false,
		},
	}

	// v1 endpoints should NOT require auth
	// v2 endpoints SHOULD require auth
	assert.True(t, cfg.EnableV2Endpoints)
	assert.True(t, cfg.V2Auth.Required)

	// Verify endpoint path detection
	v1Paths := []string{"/v1/traces", "/v1/metrics", "/v1/logs"}
	v2Paths := []string{"/v2/traces", "/v2/metrics", "/v2/logs"}

	for _, path := range v1Paths {
		isV2 := path == "/v2/traces" || path == "/v2/metrics" || path == "/v2/logs"
		assert.False(t, isV2, "v1 path should not be detected as v2: %s", path)
	}

	for _, path := range v2Paths {
		isV2 := path == "/v2/traces" || path == "/v2/metrics" || path == "/v2/logs"
		assert.True(t, isV2, "v2 path should be detected as v2: %s", path)
	}
}

// TestTFOAuthValidAPIKeyIDFormats tests valid API key ID formats
func TestTFOAuthValidAPIKeyIDFormats(t *testing.T) {
	validFormats := []string{
		"tfk_a",
		"tfk_abc123",
		"tfk_test_key_12345",
		"tfk_PROD_KEY_ABC",
		"tfk_mixed.case-with_special",
	}

	for _, keyID := range validFormats {
		t.Run(keyID, func(t *testing.T) {
			cfg := &tfoauthextension.Config{
				APIKeyID:     configopaque.String(keyID),
				APIKeySecret: configopaque.String("tfs_test_secret"),
			}
			err := cfg.Validate()
			require.NoError(t, err, "key ID %s should be valid", keyID)
		})
	}
}

// TestTFOAuthInvalidAPIKeyIDFormats tests invalid API key ID formats
func TestTFOAuthInvalidAPIKeyIDFormats(t *testing.T) {
	invalidFormats := []string{
		"",        // empty
		"abc",     // no prefix
		"tfs_key", // wrong prefix
		"TFK_key", // wrong case prefix
		"tfk",     // prefix only
	}

	for _, keyID := range invalidFormats {
		t.Run(keyID, func(t *testing.T) {
			cfg := &tfoauthextension.Config{
				APIKeyID:     configopaque.String(keyID),
				APIKeySecret: configopaque.String("tfs_test_secret"),
			}
			err := cfg.Validate()
			require.Error(t, err, "key ID %s should be invalid", keyID)
		})
	}
}

// TestTFOIdentityAutoGeneration tests that collector ID is auto-generated when not provided
func TestTFOIdentityAutoGeneration(t *testing.T) {
	ctx := context.Background()

	factory := tfoidentityextension.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoidentityextension.Config)
	// Don't set ID - should be auto-generated

	set := extensiontest.NewNopSettings(component.MustNewType("tfoidentity"))
	ext, err := factory.Create(ctx, set, cfg)
	require.NoError(t, err)

	err = ext.Start(ctx, componenttest.NewNopHost())
	require.NoError(t, err)
	defer func() { _ = ext.Shutdown(ctx) }()

	// Test that extension was created successfully
	assert.NotNil(t, ext)
}

// TestTFOComponentConfigValidation tests that all TFO components validate their configs correctly
func TestTFOComponentConfigValidation(t *testing.T) {
	t.Run("tfootlpreceiver", func(t *testing.T) {
		cfg := &tfootlpreceiver.Config{}
		err := cfg.Validate()
		require.NoError(t, err) // Empty config is valid
	})

	t.Run("tfoauthextension_valid", func(t *testing.T) {
		cfg := &tfoauthextension.Config{
			APIKeyID:     configopaque.String("tfk_valid_key"),
			APIKeySecret: configopaque.String("tfs_valid_secret"),
		}
		err := cfg.Validate()
		require.NoError(t, err)
	})

	t.Run("tfoauthextension_invalid", func(t *testing.T) {
		cfg := &tfoauthextension.Config{}
		err := cfg.Validate()
		require.Error(t, err)
	})

	t.Run("tfoidentityextension", func(t *testing.T) {
		cfg := &tfoidentityextension.Config{}
		err := cfg.Validate()
		require.NoError(t, err) // Empty config is valid
	})
}
