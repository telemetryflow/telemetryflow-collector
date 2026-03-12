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
package tfoidentityextension_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/telemetryflow/telemetryflow-collector/components/extension/tfoidentityextension"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name   string
		config tfoidentityextension.Config
	}{
		{
			name:   "empty config is valid",
			config: tfoidentityextension.Config{},
		},
		{
			name: "full config",
			config: tfoidentityextension.Config{
				ID:              "collector-001",
				Hostname:        "prod-collector-1.example.com",
				Name:            "Production Collector 1",
				Description:     "Main production collector for US-EAST region",
				Tags:            map[string]string{"environment": "production", "region": "us-east-1"},
				EnrichResources: true,
			},
		},
		{
			name: "minimal config with just ID",
			config: tfoidentityextension.Config{
				ID: "my-collector",
			},
		},
		{
			name: "config with tags only",
			config: tfoidentityextension.Config{
				Tags: map[string]string{
					"team":        "platform",
					"cost-center": "engineering",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			require.NoError(t, err)
		})
	}
}

func TestConfig_Defaults(t *testing.T) {
	cfg := tfoidentityextension.Config{}

	// All fields should be empty/false by default
	assert.Empty(t, cfg.ID)
	assert.Empty(t, cfg.Hostname)
	assert.Empty(t, cfg.Name)
	assert.Empty(t, cfg.Description)
	assert.Nil(t, cfg.Tags)
	assert.False(t, cfg.EnrichResources)
}

func TestConfig_TagsHandling(t *testing.T) {
	tests := []struct {
		name     string
		tags     map[string]string
		expected int
	}{
		{
			name:     "nil tags",
			tags:     nil,
			expected: 0,
		},
		{
			name:     "empty tags",
			tags:     map[string]string{},
			expected: 0,
		},
		{
			name:     "single tag",
			tags:     map[string]string{"env": "prod"},
			expected: 1,
		},
		{
			name: "multiple tags",
			tags: map[string]string{
				"environment": "production",
				"region":      "us-west-2",
				"team":        "platform",
				"version":     "1.0.0",
			},
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tfoidentityextension.Config{Tags: tt.tags}
			assert.Len(t, cfg.Tags, tt.expected)
		})
	}
}
