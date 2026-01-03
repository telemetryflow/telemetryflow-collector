// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package tfoexporter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configopaque"

	"github.com/telemetryflow/telemetryflow-collector/components/tfoexporter"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  tfoexporter.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with direct auth",
			config: tfoexporter.Config{
				UseV2API: true,
				Auth: &tfoexporter.AuthConfig{
					APIKeyID:     configopaque.String("tfk_test_key"),
					APIKeySecret: configopaque.String("tfs_test_secret"),
				},
			},
			wantErr: true, // Missing endpoint
			errMsg:  "endpoint is required",
		},
		{
			name: "valid config with extension auth",
			config: func() tfoexporter.Config {
				cfg := tfoexporter.Config{
					UseV2API: true,
					Auth: &tfoexporter.AuthConfig{
						Extension: component.MustNewID("tfoauth"),
					},
				}
				cfg.Endpoint = "https://api.telemetryflow.id"
				return cfg
			}(),
			wantErr: false,
		},
		{
			name: "missing endpoint",
			config: tfoexporter.Config{
				UseV2API: true,
			},
			wantErr: true,
			errMsg:  "endpoint is required",
		},
		{
			name: "missing auth configuration",
			config: func() tfoexporter.Config {
				cfg := tfoexporter.Config{
					UseV2API: true,
					Auth:     &tfoexporter.AuthConfig{},
				}
				cfg.Endpoint = "https://api.telemetryflow.id"
				return cfg
			}(),
			wantErr: true,
			errMsg:  "auth requires either api_key_id/api_key_secret or extension reference",
		},
		{
			name: "partial auth - only key id",
			config: func() tfoexporter.Config {
				cfg := tfoexporter.Config{
					UseV2API: true,
					Auth: &tfoexporter.AuthConfig{
						APIKeyID: configopaque.String("tfk_test_key"),
					},
				}
				cfg.Endpoint = "https://api.telemetryflow.id"
				return cfg
			}(),
			wantErr: true,
			errMsg:  "auth requires either api_key_id/api_key_secret or extension reference",
		},
		{
			name: "no auth config (nil)",
			config: func() tfoexporter.Config {
				cfg := tfoexporter.Config{
					UseV2API: true,
				}
				cfg.Endpoint = "https://api.telemetryflow.id"
				return cfg
			}(),
			wantErr: false, // Auth is optional
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

func TestConfig_GetTracesEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		config   tfoexporter.Config
		expected string
	}{
		{
			name: "v2 api enabled",
			config: tfoexporter.Config{
				UseV2API: true,
			},
			expected: "/v2/traces",
		},
		{
			name: "v2 api disabled",
			config: tfoexporter.Config{
				UseV2API: false,
			},
			expected: "/v1/traces",
		},
		{
			name: "custom endpoint",
			config: tfoexporter.Config{
				UseV2API:       true,
				TracesEndpoint: "/custom/traces",
			},
			expected: "/custom/traces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetTracesEndpoint()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_GetMetricsEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		config   tfoexporter.Config
		expected string
	}{
		{
			name: "v2 api enabled",
			config: tfoexporter.Config{
				UseV2API: true,
			},
			expected: "/v2/metrics",
		},
		{
			name: "v2 api disabled",
			config: tfoexporter.Config{
				UseV2API: false,
			},
			expected: "/v1/metrics",
		},
		{
			name: "custom endpoint",
			config: tfoexporter.Config{
				UseV2API:        true,
				MetricsEndpoint: "/custom/metrics",
			},
			expected: "/custom/metrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetMetricsEndpoint()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_GetLogsEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		config   tfoexporter.Config
		expected string
	}{
		{
			name: "v2 api enabled",
			config: tfoexporter.Config{
				UseV2API: true,
			},
			expected: "/v2/logs",
		},
		{
			name: "v2 api disabled",
			config: tfoexporter.Config{
				UseV2API: false,
			},
			expected: "/v1/logs",
		},
		{
			name: "custom endpoint",
			config: tfoexporter.Config{
				UseV2API:     true,
				LogsEndpoint: "/custom/logs",
			},
			expected: "/custom/logs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetLogsEndpoint()
			assert.Equal(t, tt.expected, result)
		})
	}
}
