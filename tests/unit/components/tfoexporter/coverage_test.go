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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configopaque"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/extension/extensiontest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/telemetryflow/telemetryflow-collector/components/extension/tfoauthextension"
	"github.com/telemetryflow/telemetryflow-collector/components/extension/tfoidentityextension"
	"github.com/telemetryflow/telemetryflow-collector/components/tfoexporter"
)

// extensionsHost is a component.Host that returns a populated extensions map,
// allowing tfoexporter.start() to resolve tfoauth/tfoidentity extensions.
type extensionsHost struct {
	component.Host
	exts map[component.ID]component.Component
}

func (h *extensionsHost) GetExtensions() map[component.ID]component.Component {
	return h.exts
}

func newExtHost(exts map[component.ID]component.Component) component.Host {
	return &extensionsHost{Host: componenttest.NewNopHost(), exts: exts}
}

// recordingBackend is an httptest.Server that records the last request it saw.
type recordingBackend struct {
	srv       *httptest.Server
	lastReq   *http.Request
	lastBody  []byte
	status    int
	requested chan struct{}
}

func newRecordingBackend(status int) *recordingBackend {
	rb := &recordingBackend{status: status, requested: make(chan struct{}, 16)}
	rb.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, 0, len(r.Header.Get("Content-Length")))
		buf := make([]byte, 4096)
		for {
			n, err := r.Body.Read(buf)
			if n > 0 {
				body = append(body, buf[:n]...)
			}
			if err != nil {
				break
			}
		}
		rb.lastReq = r
		rb.lastBody = body
		w.WriteHeader(rb.status)
		select {
		case rb.requested <- struct{}{}:
		default:
		}
	}))
	return rb
}

func (rb *recordingBackend) URL() string { return rb.srv.URL }
func (rb *recordingBackend) Close()      { rb.srv.Close() }
func (rb *recordingBackend) wait() {
	select {
	case <-rb.requested:
	case <-time.After(2 * time.Second):
	}
}

func disableRetry(cfg *tfoexporter.Config) {
	cfg.RetryConfig.Enabled = false
}

// --- Tests ---

func TestExporter_PushTraces_V2_Success(t *testing.T) {
	backend := newRecordingBackend(http.StatusOK)
	t.Cleanup(backend.Close)

	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)
	cfg.Endpoint = backend.URL()
	cfg.UseV2API = true
	cfg.Auth = &tfoexporter.AuthConfig{
		APIKeyID:     configopaque.String("tfk_push_traces"),
		APIKeySecret: configopaque.String("tfs_push_traces"),
	}
	disableRetry(cfg)

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))
	tracesExp, err := factory.CreateTraces(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NoError(t, tracesExp.Start(context.Background(), newExtHost(nil)))
	t.Cleanup(func() { _ = tracesExp.Shutdown(context.Background()) })

	td := ptrace.NewTraces()
	td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty().SetName("span-1")

	require.NoError(t, tracesExp.ConsumeTraces(context.Background(), td))
	backend.wait()

	require.NotNil(t, backend.lastReq)
	assert.Equal(t, "/v2/traces", backend.lastReq.URL.Path)
	assert.Equal(t, "application/x-protobuf", backend.lastReq.Header.Get("Content-Type"))
	assert.Equal(t, "tfk_push_traces", backend.lastReq.Header.Get("X-TelemetryFlow-Key-ID"))
	assert.Equal(t, "tfs_push_traces", backend.lastReq.Header.Get("X-TelemetryFlow-Key-Secret"))
}

func TestExporter_PushTraces_V1_Success(t *testing.T) {
	backend := newRecordingBackend(http.StatusOK)
	t.Cleanup(backend.Close)

	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)
	cfg.Endpoint = backend.URL()
	cfg.UseV2API = false
	disableRetry(cfg)

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))
	tracesExp, err := factory.CreateTraces(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NoError(t, tracesExp.Start(context.Background(), newExtHost(nil)))
	t.Cleanup(func() { _ = tracesExp.Shutdown(context.Background()) })

	td := ptrace.NewTraces()
	td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty().SetName("v1-span")

	require.NoError(t, tracesExp.ConsumeTraces(context.Background(), td))
	backend.wait()
	require.NotNil(t, backend.lastReq)
	assert.Equal(t, "/v1/traces", backend.lastReq.URL.Path)
}

func TestExporter_PushTraces_CustomEndpoint(t *testing.T) {
	backend := newRecordingBackend(http.StatusOK)
	t.Cleanup(backend.Close)

	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)
	cfg.Endpoint = backend.URL()
	cfg.TracesEndpoint = "/custom/traces"
	disableRetry(cfg)

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))
	tracesExp, err := factory.CreateTraces(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NoError(t, tracesExp.Start(context.Background(), newExtHost(nil)))
	t.Cleanup(func() { _ = tracesExp.Shutdown(context.Background()) })

	td := ptrace.NewTraces()
	td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty().SetName("custom")

	require.NoError(t, tracesExp.ConsumeTraces(context.Background(), td))
	backend.wait()
	require.NotNil(t, backend.lastReq)
	assert.Equal(t, "/custom/traces", backend.lastReq.URL.Path)
}

func TestExporter_PushMetrics_Success(t *testing.T) {
	backend := newRecordingBackend(http.StatusOK)
	t.Cleanup(backend.Close)

	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)
	cfg.Endpoint = backend.URL()
	cfg.UseV2API = true
	cfg.Auth = &tfoexporter.AuthConfig{
		APIKeyID:     configopaque.String("tfk_metrics"),
		APIKeySecret: configopaque.String("tfs_metrics"),
	}
	disableRetry(cfg)

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))
	metricsExp, err := factory.CreateMetrics(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NoError(t, metricsExp.Start(context.Background(), newExtHost(nil)))
	t.Cleanup(func() { _ = metricsExp.Shutdown(context.Background()) })

	md := pmetric.NewMetrics()
	sm := md.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty()
	metric := sm.Metrics().AppendEmpty()
	metric.SetName("metric-1")
	metric.SetEmptyGauge().DataPoints().AppendEmpty().SetIntValue(1)

	require.NoError(t, metricsExp.ConsumeMetrics(context.Background(), md))
	backend.wait()
	require.NotNil(t, backend.lastReq)
	assert.Equal(t, "/v2/metrics", backend.lastReq.URL.Path)
}

func TestExporter_PushLogs_Success(t *testing.T) {
	backend := newRecordingBackend(http.StatusOK)
	t.Cleanup(backend.Close)

	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)
	cfg.Endpoint = backend.URL()
	cfg.UseV2API = true
	cfg.Auth = &tfoexporter.AuthConfig{
		APIKeyID:     configopaque.String("tfk_logs"),
		APIKeySecret: configopaque.String("tfs_logs"),
	}
	disableRetry(cfg)

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))
	logsExp, err := factory.CreateLogs(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NoError(t, logsExp.Start(context.Background(), newExtHost(nil)))
	t.Cleanup(func() { _ = logsExp.Shutdown(context.Background()) })

	ld := plog.NewLogs()
	ld.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty().Body().SetStr("hello")

	require.NoError(t, logsExp.ConsumeLogs(context.Background(), ld))
	backend.wait()
	require.NotNil(t, backend.lastReq)
	assert.Equal(t, "/v2/logs", backend.lastReq.URL.Path)
}

func TestExporter_SendData_Non2xx(t *testing.T) {
	backend := newRecordingBackend(http.StatusInternalServerError)
	t.Cleanup(backend.Close)

	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)
	cfg.Endpoint = backend.URL()
	disableRetry(cfg)

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))
	tracesExp, err := factory.CreateTraces(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NoError(t, tracesExp.Start(context.Background(), newExtHost(nil)))
	t.Cleanup(func() { _ = tracesExp.Shutdown(context.Background()) })

	td := ptrace.NewTraces()
	td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty().SetName("fail")

	err = tracesExp.ConsumeTraces(context.Background(), td)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code: 500")
}

func TestExporter_Start_WithAuthExtension(t *testing.T) {
	backend := newRecordingBackend(http.StatusOK)
	t.Cleanup(backend.Close)

	authCfg := tfoauthextension.NewFactory().CreateDefaultConfig().(*tfoauthextension.Config)
	authCfg.APIKeyID = configopaque.String("tfk_via_extension")
	authCfg.APIKeySecret = configopaque.String("tfs_via_extension")
	authSet := extensiontest.NewNopSettings(component.MustNewType("tfoauth"))
	authExt, err := tfoauthextension.NewFactory().Create(context.Background(), authSet, authCfg)
	require.NoError(t, err)
	require.NoError(t, authExt.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { _ = authExt.Shutdown(context.Background()) })

	authID := component.MustNewID("tfoauth")
	host := newExtHost(map[component.ID]component.Component{authID: authExt})

	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)
	cfg.Endpoint = backend.URL()
	cfg.Auth = &tfoexporter.AuthConfig{
		Extension: authID,
	}
	disableRetry(cfg)

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))
	tracesExp, err := factory.CreateTraces(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NoError(t, tracesExp.Start(context.Background(), host))
	t.Cleanup(func() { _ = tracesExp.Shutdown(context.Background()) })

	td := ptrace.NewTraces()
	td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty().SetName("ext-auth")

	require.NoError(t, tracesExp.ConsumeTraces(context.Background(), td))
	backend.wait()
	require.NotNil(t, backend.lastReq)
	assert.Equal(t, "tfk_via_extension", backend.lastReq.Header.Get("X-TelemetryFlow-Key-ID"))
	assert.Equal(t, "tfs_via_extension", backend.lastReq.Header.Get("X-TelemetryFlow-Key-Secret"))
}

func TestExporter_Start_AuthExtensionNotFound(t *testing.T) {
	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)
	cfg.Auth = &tfoexporter.AuthConfig{
		Extension: component.MustNewID("tfoauth"),
	}
	disableRetry(cfg)

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))
	tracesExp, err := factory.CreateTraces(context.Background(), set, cfg)
	require.NoError(t, err)

	host := newExtHost(nil) // no extensions registered
	err = tracesExp.Start(context.Background(), host)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `tfoauth extension "tfoauth" not found`)
}

func TestExporter_Start_WithIdentityExtension(t *testing.T) {
	backend := newRecordingBackend(http.StatusOK)
	t.Cleanup(backend.Close)

	idCfg := tfoidentityextension.NewFactory().CreateDefaultConfig().(*tfoidentityextension.Config)
	idCfg.ID = "collector-id-xyz"
	idCfg.Hostname = "host-xyz"
	idSet := extensiontest.NewNopSettings(component.MustNewType("tfoidentity"))
	idExt, err := tfoidentityextension.NewFactory().Create(context.Background(), idSet, idCfg)
	require.NoError(t, err)
	require.NoError(t, idExt.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { _ = idExt.Shutdown(context.Background()) })

	identityID := component.MustNewID("tfoidentity")
	host := newExtHost(map[component.ID]component.Component{identityID: idExt})

	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)
	cfg.Endpoint = backend.URL()
	cfg.CollectorIdentity = identityID
	disableRetry(cfg)

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))
	tracesExp, err := factory.CreateTraces(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NoError(t, tracesExp.Start(context.Background(), host))
	t.Cleanup(func() { _ = tracesExp.Shutdown(context.Background()) })

	td := ptrace.NewTraces()
	td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty().SetName("with-id")

	require.NoError(t, tracesExp.ConsumeTraces(context.Background(), td))
	backend.wait()
	require.NotNil(t, backend.lastReq)
	assert.Equal(t, "collector-id-xyz", backend.lastReq.Header.Get("X-TelemetryFlow-Collector-ID"))
}

func TestExporter_Start_IdentityExtensionMissing_Warns(t *testing.T) {
	// collector_identity references a non-existent extension — start() should
	// log a warning but NOT fail (start returns nil).
	factory := tfoexporter.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfoexporter.Config)
	cfg.Endpoint = "http://127.0.0.1:1"
	cfg.CollectorIdentity = component.MustNewID("tfoidentity")
	disableRetry(cfg)

	set := exportertest.NewNopSettings(component.MustNewType("tfo"))
	tracesExp, err := factory.CreateTraces(context.Background(), set, cfg)
	require.NoError(t, err)

	host := newExtHost(nil) // identity extension not registered
	require.NoError(t, tracesExp.Start(context.Background(), host))
	require.NoError(t, tracesExp.Shutdown(context.Background()))
}

func TestExporter_GetEndpointPath_DirectOverrides(t *testing.T) {
	// Direct cover of the three Get*Endpoint path-resolution branches.
	cfg := &tfoexporter.Config{}
	cfg.TracesEndpoint = "/custom/t"
	cfg.MetricsEndpoint = "/custom/m"
	cfg.LogsEndpoint = "/custom/l"
	assert.Equal(t, "/custom/t", cfg.GetTracesEndpoint())
	assert.Equal(t, "/custom/m", cfg.GetMetricsEndpoint())
	assert.Equal(t, "/custom/l", cfg.GetLogsEndpoint())

	cfg2 := &tfoexporter.Config{UseV2API: true}
	assert.Equal(t, "/v2/traces", cfg2.GetTracesEndpoint())
	assert.Equal(t, "/v2/metrics", cfg2.GetMetricsEndpoint())
	assert.Equal(t, "/v2/logs", cfg2.GetLogsEndpoint())

	cfg3 := &tfoexporter.Config{UseV2API: false}
	assert.Equal(t, "/v1/traces", cfg3.GetTracesEndpoint())
	assert.Equal(t, "/v1/metrics", cfg3.GetMetricsEndpoint())
	assert.Equal(t, "/v1/logs", cfg3.GetLogsEndpoint())
}

func TestExporter_ConfigValidate_Errors(t *testing.T) {
	// Empty endpoint → error.
	cfg := &tfoexporter.Config{}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "endpoint is required")

	// Auth present but neither direct creds nor extension → error.
	cfg.Endpoint = "http://localhost:4318"
	cfg.Auth = &tfoexporter.AuthConfig{}
	err = cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth requires either")
}

// Compile-time check that we use the exporter package types.
var _ exporter.Traces = (exporter.Traces)(nil)
