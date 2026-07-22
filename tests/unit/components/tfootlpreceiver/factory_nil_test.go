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
package tfootlpreceiver_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/telemetryflow/telemetryflow-collector/components/tfoexporter"
	"github.com/telemetryflow/telemetryflow-collector/components/tfootlpreceiver"
)

// TestReceiver_Factory_NilConfig exercises the nil-config and wrong-type
// branches that flow through resolveReceiverConfig + newTFOOTLPReceiver. The
// nil check in newTFOOTLPReceiver runs BEFORE the singleton lookup, so the
// test is hermetic regardless of prior receiver state.
func TestReceiver_Factory_NilConfig(t *testing.T) {
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	sink := new(consumertest.TracesSink)

	_, err := factory.CreateTraces(context.Background(), set, nil, sink)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")
}

func TestReceiver_Factory_WrongConfigType(t *testing.T) {
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	sink := new(consumertest.TracesSink)

	// A *tfoexporter.Config is a non-nil component.Config but the wrong type.
	wrongCfg := tfoexporter.NewFactory().CreateDefaultConfig()

	_, err := factory.CreateTraces(context.Background(), set, wrongCfg, sink)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")
}

func TestReceiver_Factory_NilConfig_Metrics(t *testing.T) {
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	sink := new(consumertest.MetricsSink)

	_, err := factory.CreateMetrics(context.Background(), set, nil, sink)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")
}

func TestReceiver_Factory_NilConfig_Logs(t *testing.T) {
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	sink := new(consumertest.LogsSink)

	_, err := factory.CreateLogs(context.Background(), set, nil, sink)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")
}
