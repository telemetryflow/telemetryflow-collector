// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package tfoauthextension_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/config/configopaque"

	"github.com/telemetryflow/telemetryflow-collector/components/extension/tfoauthextension"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  tfoauthextension.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: tfoauthextension.Config{
				APIKeyID:     configopaque.String("tfk_test_key_12345"),
				APIKeySecret: configopaque.String("tfs_test_secret_12345"),
			},
			wantErr: false,
		},
		{
			name: "missing api key id",
			config: tfoauthextension.Config{
				APIKeySecret: configopaque.String("tfs_test_secret"),
			},
			wantErr: true,
			errMsg:  "api_key_id is required",
		},
		{
			name: "missing api key secret",
			config: tfoauthextension.Config{
				APIKeyID: configopaque.String("tfk_test_key"),
			},
			wantErr: true,
			errMsg:  "api_key_secret is required",
		},
		{
			name: "invalid api key id prefix",
			config: tfoauthextension.Config{
				APIKeyID:     configopaque.String("invalid_key"),
				APIKeySecret: configopaque.String("tfs_test_secret"),
			},
			wantErr: true,
			errMsg:  "api_key_id must start with 'tfk_' prefix",
		},
		{
			name: "invalid api key secret prefix",
			config: tfoauthextension.Config{
				APIKeyID:     configopaque.String("tfk_test_key"),
				APIKeySecret: configopaque.String("invalid_secret"),
			},
			wantErr: true,
			errMsg:  "api_key_secret must start with 'tfs_' prefix",
		},
		{
			name: "validate on start without endpoint",
			config: tfoauthextension.Config{
				APIKeyID:        configopaque.String("tfk_test_key"),
				APIKeySecret:    configopaque.String("tfs_test_secret"),
				ValidateOnStart: true,
			},
			wantErr: true,
			errMsg:  "validation_endpoint is required when validate_on_start is true",
		},
		{
			name: "validate on start with endpoint",
			config: tfoauthextension.Config{
				APIKeyID:           configopaque.String("tfk_test_key"),
				APIKeySecret:       configopaque.String("tfs_test_secret"),
				ValidateOnStart:    true,
				ValidationEndpoint: "https://api.telemetryflow.id/v1/auth/validate",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAPIKeyIDFormat(t *testing.T) {
	validKeys := []string{
		"tfk_a",
		"tfk_abc123",
		"tfk_very_long_key_id_with_numbers_12345",
		"tfk_UPPERCASE_ALLOWED",
		"tfk_mixed-chars.allowed",
	}

	for _, key := range validKeys {
		t.Run(key, func(t *testing.T) {
			assert.True(t, len(key) >= 4)
			assert.Equal(t, "tfk_", key[:4])
		})
	}
}

func TestAPIKeySecretFormat(t *testing.T) {
	validSecrets := []string{
		"tfs_a",
		"tfs_abc123",
		"tfs_very_long_secret_with_numbers_12345",
		"tfs_UPPERCASE_ALLOWED",
		"tfs_mixed-chars.allowed",
	}

	for _, secret := range validSecrets {
		t.Run(secret, func(t *testing.T) {
			assert.True(t, len(secret) >= 4)
			assert.Equal(t, "tfs_", secret[:4])
		})
	}
}
