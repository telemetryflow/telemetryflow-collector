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
package tfootlpreceiver_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/telemetryflow/telemetryflow-collector/components/tfootlpreceiver"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  tfootlpreceiver.Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with defaults",
			config: tfootlpreceiver.Config{
				Protocols: tfootlpreceiver.ProtocolsConfig{
					GRPC: &tfootlpreceiver.GRPCConfig{},
					HTTP: &tfootlpreceiver.HTTPConfig{},
				},
				EnableV2Endpoints: true,
			},
			wantErr: false,
		},
		{
			name: "valid config with no protocols",
			config: tfootlpreceiver.Config{
				Protocols:         tfootlpreceiver.ProtocolsConfig{},
				EnableV2Endpoints: true,
			},
			wantErr: false,
		},
		{
			name: "valid config with grpc only",
			config: tfootlpreceiver.Config{
				Protocols: tfootlpreceiver.ProtocolsConfig{
					GRPC: &tfootlpreceiver.GRPCConfig{},
				},
				EnableV2Endpoints: true,
			},
			wantErr: false,
		},
		{
			name: "valid config with http only",
			config: tfootlpreceiver.Config{
				Protocols: tfootlpreceiver.ProtocolsConfig{
					HTTP: &tfootlpreceiver.HTTPConfig{},
				},
				EnableV2Endpoints: true,
			},
			wantErr: false,
		},
		{
			name: "valid config with v2 auth disabled",
			config: tfootlpreceiver.Config{
				Protocols: tfootlpreceiver.ProtocolsConfig{
					HTTP: &tfootlpreceiver.HTTPConfig{},
				},
				EnableV2Endpoints: true,
				V2Auth: tfootlpreceiver.V2AuthConfig{
					Required: false,
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with v2 auth enabled",
			config: tfootlpreceiver.Config{
				Protocols: tfootlpreceiver.ProtocolsConfig{
					HTTP: &tfootlpreceiver.HTTPConfig{},
				},
				EnableV2Endpoints: true,
				V2Auth: tfootlpreceiver.V2AuthConfig{
					Required:       true,
					ValidateSecret: false,
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with allowed api key ids",
			config: tfootlpreceiver.Config{
				Protocols: tfootlpreceiver.ProtocolsConfig{
					HTTP: &tfootlpreceiver.HTTPConfig{},
				},
				EnableV2Endpoints: true,
				V2Auth: tfootlpreceiver.V2AuthConfig{
					Required:       true,
					ValidAPIKeyIDs: []string{"tfk_test_key_1", "tfk_test_key_2"},
					ValidateSecret: false,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid config with validate secret and allowed ids",
			config: tfootlpreceiver.Config{
				Protocols: tfootlpreceiver.ProtocolsConfig{
					HTTP: &tfootlpreceiver.HTTPConfig{},
				},
				EnableV2Endpoints: true,
				V2Auth: tfootlpreceiver.V2AuthConfig{
					Required:       true,
					ValidAPIKeyIDs: []string{"tfk_test_key_1"},
					ValidateSecret: true,
				},
			},
			wantErr: true,
			errMsg:  "validate_secret with valid_api_key_ids requires tfoauth extension",
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

func TestV2AuthConfig_Defaults(t *testing.T) {
	cfg := tfootlpreceiver.V2AuthConfig{}

	// Default values should be false/empty
	assert.False(t, cfg.Required)
	assert.False(t, cfg.ValidateSecret)
	assert.Empty(t, cfg.ValidAPIKeyIDs)
}
