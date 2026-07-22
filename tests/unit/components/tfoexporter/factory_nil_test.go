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
package tfoexporter_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter/exportertest"

	"github.com/telemetryflow/telemetryflow-collector/components/tfoexporter"
	"github.com/telemetryflow/telemetryflow-collector/components/tfootlpreceiver"
)

// TestFactory_NilConfig exercises the nil-config and wrong-type-config error
// branches that flow through resolveConfig + newTFOExporter for each signal.
func TestFactory_NilConfig(t *testing.T) {
	factory := tfoexporter.NewFactory()
	set := exportertest.NewNopSettings(component.MustNewType("tfo"))

	t.Run("traces_nil", func(t *testing.T) {
		_, err := factory.CreateTraces(context.Background(), set, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})
	t.Run("metrics_nil", func(t *testing.T) {
		_, err := factory.CreateMetrics(context.Background(), set, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})
	t.Run("logs_nil", func(t *testing.T) {
		_, err := factory.CreateLogs(context.Background(), set, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})
}

// TestFactory_WrongConfigType covers the comma-ok assertion branch where cfg
// is non-nil but not a *tfoexporter.Config. The receiver Config type from a
// different component is a suitable stand-in.
func TestFactory_WrongConfigType(t *testing.T) {
	factory := tfoexporter.NewFactory()
	set := exportertest.NewNopSettings(component.MustNewType("tfo"))

	wrongCfg := tfootlpreceiver.NewFactory().CreateDefaultConfig()

	t.Run("traces_wrong_type", func(t *testing.T) {
		_, err := factory.CreateTraces(context.Background(), set, wrongCfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})
	t.Run("metrics_wrong_type", func(t *testing.T) {
		_, err := factory.CreateMetrics(context.Background(), set, wrongCfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})
	t.Run("logs_wrong_type", func(t *testing.T) {
		_, err := factory.CreateLogs(context.Background(), set, wrongCfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "config cannot be nil")
	})
}
