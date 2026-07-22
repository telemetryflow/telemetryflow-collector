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
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/extension/extensiontest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/telemetryflow/telemetryflow-collector/components/extension/tfoidentityextension"
	"github.com/telemetryflow/telemetryflow-collector/components/tfoexporter"
)

// TestExporter_SendData_BadURL covers the http.NewRequestWithContext error
// branch in sendData. A space inside the host makes the URL invalid for
// http.NewRequest, which returns an error before any I/O is attempted.
func TestExporter_SendData_BadURL(t *testing.T) {
	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)
	// "http://bad host" — the embedded space is rejected by url.Parse.
	cfg.Endpoint = "http://bad host"
	disableRetry(cfg)

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))
	tracesExp, err := factory.CreateTraces(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NoError(t, tracesExp.Start(context.Background(), newExtHost(nil)))
	t.Cleanup(func() { _ = tracesExp.Shutdown(context.Background()) })

	td := ptrace.NewTraces()
	td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty().SetName("bad-url")

	err = tracesExp.ConsumeTraces(context.Background(), td)
	require.Error(t, err)
	// Either "failed to create request" (sendData) or wrapped retry error.
	assert.Contains(t, err.Error(), "failed to create request")
}

// TestExporter_SendData_BadURL_Metrics covers the same error path via the
// metrics signal.
func TestExporter_SendData_BadURL_Metrics(t *testing.T) {
	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)
	cfg.Endpoint = "http://bad host"
	disableRetry(cfg)

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))
	metricsExp, err := factory.CreateMetrics(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NoError(t, metricsExp.Start(context.Background(), newExtHost(nil)))
	t.Cleanup(func() { _ = metricsExp.Shutdown(context.Background()) })

	md := pmetric.NewMetrics()
	m := md.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	m.SetName("bad-url-metric")
	m.SetEmptyGauge().DataPoints().AppendEmpty().SetIntValue(1)

	err = metricsExp.ConsumeMetrics(context.Background(), md)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create request")
}

// TestExporter_SendData_BadURL_Logs covers the same error path via the logs
// signal.
func TestExporter_SendData_BadURL_Logs(t *testing.T) {
	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)
	cfg.Endpoint = "http://bad host"
	disableRetry(cfg)

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))
	logsExp, err := factory.CreateLogs(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NoError(t, logsExp.Start(context.Background(), newExtHost(nil)))
	t.Cleanup(func() { _ = logsExp.Shutdown(context.Background()) })

	ld := plog.NewLogs()
	ld.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty().Body().SetStr("bad-url-log")

	err = logsExp.ConsumeLogs(context.Background(), ld)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create request")
}

// TestExporter_NoHeaders_WhenAuthEmpty verifies that when neither direct creds
// nor extension auth is configured, sendData still functions but injects no
// auth headers (defensive coverage of the empty-credential branches).
func TestExporter_NoHeaders_WhenAuthEmpty(t *testing.T) {
	backend := newRecordingBackend(200)
	t.Cleanup(backend.Close)

	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)
	cfg.Endpoint = backend.URL()
	cfg.Auth = nil // no auth
	disableRetry(cfg)

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))
	tracesExp, err := factory.CreateTraces(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NoError(t, tracesExp.Start(context.Background(), newExtHost(nil)))
	t.Cleanup(func() { _ = tracesExp.Shutdown(context.Background()) })

	td := ptrace.NewTraces()
	td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty().SetName("no-auth")

	require.NoError(t, tracesExp.ConsumeTraces(context.Background(), td))
	backend.wait()
	require.NotNil(t, backend.lastReq)
	assert.Empty(t, backend.lastReq.Header.Get("X-TelemetryFlow-Key-ID"))
	assert.Empty(t, backend.lastReq.Header.Get("X-TelemetryFlow-Key-Secret"))
	assert.Empty(t, backend.lastReq.Header.Get("X-TelemetryFlow-Collector-ID"))
}

// TestExporter_ConfigureDirectAuth_OnlyKeyID exercises the half-configured
// direct-auth branch (only APIKeyID set, no secret) — start() should still
// resolve the partial creds and pass them through.
func TestExporter_ConfigureDirectAuth_OnlyKeyID(t *testing.T) {
	backend := newRecordingBackend(200)
	t.Cleanup(backend.Close)

	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)
	cfg.Endpoint = backend.URL()
	cfg.Auth = &tfoexporter.AuthConfig{
		APIKeyID: configopaque.String("tfk_only_id"),
		// Secret intentionally left empty.
	}
	disableRetry(cfg)

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))
	tracesExp, err := factory.CreateTraces(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NoError(t, tracesExp.Start(context.Background(), newExtHost(nil)))
	t.Cleanup(func() { _ = tracesExp.Shutdown(context.Background()) })

	td := ptrace.NewTraces()
	td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty().SetName("half-auth")

	require.NoError(t, tracesExp.ConsumeTraces(context.Background(), td))
	backend.wait()
	require.NotNil(t, backend.lastReq)
	assert.Equal(t, "tfk_only_id", backend.lastReq.Header.Get("X-TelemetryFlow-Key-ID"))
	// Secret header should be skipped because apiKeySecret resolves to "".
	assert.Empty(t, backend.lastReq.Header.Get("X-TelemetryFlow-Key-Secret"))
}

// TestExporter_AuthExtensionNotAuthProvider exercises the branch in start()
// where the referenced extension exists but does not implement AuthProvider.
func TestExporter_AuthExtensionNotAuthProvider(t *testing.T) {
	authID := component.MustNewID("tfoauth")

	// Build a real extension (the identity extension) that does not satisfy
	// the AuthProvider interface — so the type assertion in start() fails and
	// apiKeyID/apiKeySecret remain empty.
	idFactory := tfoidentityextension.NewFactory()
	idCfg := idFactory.CreateDefaultConfig()
	idSet := extensiontest.NewNopSettings(component.MustNewType("tfoidentity"))
	idExt, err := idFactory.Create(context.Background(), idSet, idCfg)
	require.NoError(t, err)
	require.NoError(t, idExt.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { _ = idExt.Shutdown(context.Background()) })

	host := newExtHost(map[component.ID]component.Component{authID: idExt})

	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)
	cfg.Endpoint = "http://127.0.0.1:1"
	cfg.Auth = &tfoexporter.AuthConfig{Extension: authID}
	disableRetry(cfg)

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))
	tracesExp, err := factory.CreateTraces(context.Background(), set, cfg)
	require.NoError(t, err)

	// Start should succeed (the extension simply doesn't provide creds),
	// but apiKeyID/apiKeySecret remain empty.
	require.NoError(t, tracesExp.Start(context.Background(), host))
	require.NoError(t, tracesExp.Shutdown(context.Background()))
}
