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
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/telemetryflow/telemetryflow-collector/components/tfootlpreceiver"
)

// TestReceiver_NewReceiver_AlreadyStarted exercises the "if started" branch
// inside newTFOOTLPReceiver. We create and start a traces receiver, then
// create a metrics receiver — the second call should return the same shared
// started instance.
func TestReceiver_NewReceiver_AlreadyStarted(t *testing.T) {
	cfg := httpOnlyCfg(t, false, false, nil)

	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))

	tracesSink := new(consumertest.TracesSink)
	rt, err := factory.CreateTraces(context.Background(), set, cfg, tracesSink)
	require.NoError(t, err)
	require.NoError(t, rt.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { _ = rt.Shutdown(context.Background()) })

	// Now the shared instance is started. A subsequent Create* call should
	// take the "if started" branch and reuse it without overwriting the cfg.
	metricsSink := new(consumertest.MetricsSink)
	rm, err := factory.CreateMetrics(context.Background(), set, cfg, metricsSink)
	require.NoError(t, err)
	require.NotNil(t, rm)

	logsSink := new(consumertest.LogsSink)
	rl, err := factory.CreateLogs(context.Background(), set, cfg, logsSink)
	require.NoError(t, err)
	require.NotNil(t, rl)
}

// TestReceiver_NewReceiver_NotStarted_OverwriteCfg exercises the "not started
// yet" branch where an existing instance is reused but its cfg is updated.
// We create a traces receiver (instance cached but not started), then create
// a metrics receiver with the SAME cfg — should overwrite the cached cfg.
func TestReceiver_NewReceiver_NotStarted_OverwriteCfg(t *testing.T) {
	cfg := httpOnlyCfg(t, false, false, nil)

	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))

	rt, err := factory.CreateTraces(context.Background(), set, cfg, new(consumertest.TracesSink))
	require.NoError(t, err)
	// Do NOT start — instance exists but is not started yet.

	rm, err := factory.CreateMetrics(context.Background(), set, cfg, new(consumertest.MetricsSink))
	require.NoError(t, err)
	require.NotNil(t, rm)

	// Cleanup without start.
	require.NoError(t, rt.Shutdown(context.Background()))
}

// TestReceiver_DefaultGRPCAndHTTPConstants sanity-checks that the exported
// default constants match what the receiver uses internally.
func TestReceiver_DefaultGRPCAndHTTPConstants(t *testing.T) {
	assert.Equal(t, "0.0.0.0:4317", tfootlpreceiver.DefaultGRPCEndpoint)
	assert.Equal(t, "0.0.0.0:4318", tfootlpreceiver.DefaultHTTPEndpoint)
	assert.Equal(t, "tfootlp", tfootlpreceiver.TypeStr)
}
