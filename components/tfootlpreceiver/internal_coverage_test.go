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
package tfootlpreceiver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
)

// TestNewTFOOTLPReceiver_NilArgs exercises the defensive nil-check branches
// in newTFOOTLPReceiver that cannot be reached via the public factory API.
// Each nil-check runs BEFORE the singleton lookup so it is safe to invoke
// directly without corrupting process state.
func TestNewTFOOTLPReceiver_NilArgs(t *testing.T) {
	goodSettings := func() *receiver.Settings {
		s := &receiver.Settings{}
		s.Logger = zap.NewNop()
		return s
	}

	t.Run("nil_cfg", func(t *testing.T) {
		_, err := newTFOOTLPReceiver(nil, goodSettings())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})

	t.Run("nil_settings", func(t *testing.T) {
		_, err := newTFOOTLPReceiver(&Config{}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "settings cannot be nil")
	})

	t.Run("nil_logger", func(t *testing.T) {
		_, err := newTFOOTLPReceiver(&Config{}, &receiver.Settings{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "settings.Logger cannot be nil")
	})
}

// TestResolveReceiverConfig_AllBranches covers all three branches of
// resolveReceiverConfig: nil interface, wrong type, and correct type.
func TestResolveReceiverConfig_AllBranches(t *testing.T) {
	t.Run("nil_interface", func(t *testing.T) {
		assert.Nil(t, resolveReceiverConfig(nil))
	})

	t.Run("wrong_type", func(t *testing.T) {
		assert.Nil(t, resolveReceiverConfig(wrongRcvConfig{}))
	})

	t.Run("correct_type", func(t *testing.T) {
		cfg := &Config{}
		got := resolveReceiverConfig(cfg)
		require.NotNil(t, got)
		assert.Same(t, cfg, got)
	})
}

type wrongRcvConfig struct{}

func (wrongRcvConfig) err() error { return nil }
