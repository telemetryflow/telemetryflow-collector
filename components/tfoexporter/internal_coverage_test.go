// TelemetryFlow Collector - AI-Powered Observability & Incident Response Management (IRM) Platform
// Copyright (c) 2024-2026 Telemetri Data Indonesia. All rights reserved.
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
package tfoexporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/exporter"
	"go.uber.org/zap"
)

// TestNewTFOExporter_NilArgs exercises the defensive nil-check branches in
// newTFOExporter that cannot be reached via the public factory API (which
// always passes a populated *exporter.Settings).
func TestNewTFOExporter_NilArgs(t *testing.T) {
	goodSettings := func() *exporter.Settings {
		s := &exporter.Settings{}
		s.Logger = zap.NewNop()
		return s
	}

	t.Run("nil_cfg", func(t *testing.T) {
		_, err := newTFOExporter(nil, goodSettings())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("nil_settings", func(t *testing.T) {
		_, err := newTFOExporter(&Config{}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "settings cannot be nil")
	})

	t.Run("nil_logger", func(t *testing.T) {
		// exporter.Settings embeds TelemetrySettings; an empty struct has a nil Logger.
		_, err := newTFOExporter(&Config{}, &exporter.Settings{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "settings.Logger cannot be nil")
	})

	t.Run("happy_path", func(t *testing.T) {
		exp, err := newTFOExporter(&Config{}, goodSettings())
		require.NoError(t, err)
		require.NotNil(t, exp)
	})
}

// TestResolveConfig_AllBranches covers all three branches of resolveConfig:
// nil interface, wrong type, and correct type.
func TestResolveConfig_AllBranches(t *testing.T) {
	t.Run("nil_interface", func(t *testing.T) {
		assert.Nil(t, resolveConfig(nil))
	})

	t.Run("wrong_type", func(t *testing.T) {
		// A *tfoexporter.Config is itself a component.Config, but here we
		// pass a different concrete type. Use a bare struct that satisfies
		// the (empty) component.Config interface.
		assert.Nil(t, resolveConfig(wrongConfig{}))
	})

	t.Run("correct_type", func(t *testing.T) {
		cfg := &Config{}
		got := resolveConfig(cfg)
		require.NotNil(t, got)
		assert.Same(t, cfg, got)
	})
}

type wrongConfig struct{}

func (wrongConfig) err() error { return nil }
