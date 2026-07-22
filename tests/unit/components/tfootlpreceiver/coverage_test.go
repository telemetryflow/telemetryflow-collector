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
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/telemetryflow/telemetryflow-collector/components/tfootlpreceiver"
)

// freePort returns a TCP port that is free at call time by binding to :0,
// reading the assigned port, then closing the listener.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port
}

// httpOnlyCfg returns a Config that has gRPC disabled and HTTP listening on an
// ephemeral port. V2 endpoints and auth settings are controlled by the caller.
func httpOnlyCfg(t *testing.T, v2Required bool, validateSecret bool, validIDs []string) *tfootlpreceiver.Config {
	t.Helper()
	port := freePort(t)
	httpServerCfg := confighttp.NewDefaultServerConfig()
	httpServerCfg.NetAddr.Endpoint = fmt.Sprintf("127.0.0.1:%d", port)
	httpServerCfg.NetAddr.Transport = confignet.TransportTypeTCP
	return &tfootlpreceiver.Config{
		Protocols: tfootlpreceiver.ProtocolsConfig{
			HTTP: &tfootlpreceiver.HTTPConfig{
				ServerConfig:   httpServerCfg,
				TracesURLPath:  "/v1/traces",
				MetricsURLPath: "/v1/metrics",
				LogsURLPath:    "/v1/logs",
			},
		},
		EnableV2Endpoints: true,
		V2Auth: tfootlpreceiver.V2AuthConfig{
			Required:       v2Required,
			ValidateSecret: validateSecret,
			ValidAPIKeyIDs: validIDs,
		},
	}
}

func grpcHTTPCfg(t *testing.T) *tfootlpreceiver.Config {
	t.Helper()
	grpcPort := freePort(t)
	httpPort := freePort(t)
	httpServerCfg := confighttp.NewDefaultServerConfig()
	httpServerCfg.NetAddr.Endpoint = fmt.Sprintf("127.0.0.1:%d", httpPort)
	httpServerCfg.NetAddr.Transport = confignet.TransportTypeTCP
	return &tfootlpreceiver.Config{
		Protocols: tfootlpreceiver.ProtocolsConfig{
			GRPC: &tfootlpreceiver.GRPCConfig{
				ServerConfig: configgrpc.ServerConfig{
					NetAddr: confignet.AddrConfig{
						Endpoint:  fmt.Sprintf("127.0.0.1:%d", grpcPort),
						Transport: confignet.TransportTypeTCP,
					},
				},
			},
			HTTP: &tfootlpreceiver.HTTPConfig{
				ServerConfig:   httpServerCfg,
				TracesURLPath:  "/v1/traces",
				MetricsURLPath: "/v1/metrics",
				LogsURLPath:    "/v1/logs",
			},
		},
		EnableV2Endpoints: true,
		V2Auth:            tfootlpreceiver.V2AuthConfig{Required: false},
	}
}

func startTracesReceiver(t *testing.T, cfg *tfootlpreceiver.Config, sink *consumertest.TracesSink) component.Component {
	t.Helper()
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	r, err := factory.CreateTraces(context.Background(), set, cfg, sink)
	require.NoError(t, err)
	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { _ = r.Shutdown(context.Background()) })
	time.Sleep(80 * time.Millisecond) // give goroutines a moment to bind
	return r
}

func startMetricsReceiver(t *testing.T, cfg *tfootlpreceiver.Config, sink *consumertest.MetricsSink) component.Component {
	t.Helper()
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	r, err := factory.CreateMetrics(context.Background(), set, cfg, sink)
	require.NoError(t, err)
	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { _ = r.Shutdown(context.Background()) })
	time.Sleep(80 * time.Millisecond)
	return r
}

func startLogsReceiver(t *testing.T, cfg *tfootlpreceiver.Config, sink *consumertest.LogsSink) component.Component {
	t.Helper()
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	r, err := factory.CreateLogs(context.Background(), set, cfg, sink)
	require.NoError(t, err)
	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { _ = r.Shutdown(context.Background()) })
	time.Sleep(80 * time.Millisecond)
	return r
}

// =============================================================================
// validateV2Auth — full HTTP integration tests
// =============================================================================

func TestReceiver_V2Auth_MissingKeyID_401(t *testing.T) {
	cfg := httpOnlyCfg(t, true, false, nil)
	sink := new(consumertest.TracesSink)
	startTracesReceiver(t, cfg, sink)

	url := fmt.Sprintf("http://%s/v2/traces", cfg.Protocols.HTTP.NetAddr.Endpoint)
	resp, body := doPost(t, url, nil, nil)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, string(body), "missing TelemetryFlow API Key ID")
}

func TestReceiver_V2Auth_InvalidKeyIDFormat_401(t *testing.T) {
	cfg := httpOnlyCfg(t, true, false, nil)
	sink := new(consumertest.TracesSink)
	startTracesReceiver(t, cfg, sink)

	url := fmt.Sprintf("http://%s/v2/traces", cfg.Protocols.HTTP.NetAddr.Endpoint)
	resp, body := doPost(t, url, map[string]string{"X-TelemetryFlow-Key-ID": "invalid_key"}, nil)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, string(body), "invalid TelemetryFlow API Key ID format")
}

func TestReceiver_V2Auth_KeyIDNotInAllowList_403(t *testing.T) {
	cfg := httpOnlyCfg(t, true, false, []string{"tfk_allowed_one"})
	sink := new(consumertest.TracesSink)
	startTracesReceiver(t, cfg, sink)

	url := fmt.Sprintf("http://%s/v2/traces", cfg.Protocols.HTTP.NetAddr.Endpoint)
	resp, body := doPost(t, url, map[string]string{"X-TelemetryFlow-Key-ID": "tfk_other"}, nil)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Contains(t, string(body), "API Key ID not authorized")
}

func TestReceiver_V2Auth_KeyIDInAllowList_OK(t *testing.T) {
	cfg := httpOnlyCfg(t, true, false, []string{"tfk_allowed_one"})
	sink := new(consumertest.TracesSink)
	startTracesReceiver(t, cfg, sink)

	td := ptrace.NewTraces()
	td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty().SetName("allowed")
	data, _ := ptraceotlp.NewExportRequestFromTraces(td).MarshalProto()

	url := fmt.Sprintf("http://%s/v2/traces", cfg.Protocols.HTTP.NetAddr.Endpoint)
	resp, _ := doPost(t, url, map[string]string{"X-TelemetryFlow-Key-ID": "tfk_allowed_one"}, data)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.Eventually(t, func() bool { return sink.SpanCount() == 1 }, time.Second, 10*time.Millisecond)
}

func TestReceiver_V2Auth_ValidateSecret_Missing_401(t *testing.T) {
	cfg := httpOnlyCfg(t, true, true, nil)
	sink := new(consumertest.TracesSink)
	startTracesReceiver(t, cfg, sink)

	url := fmt.Sprintf("http://%s/v2/traces", cfg.Protocols.HTTP.NetAddr.Endpoint)
	resp, body := doPost(t, url, map[string]string{"X-TelemetryFlow-Key-ID": "tfk_test"}, nil)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, string(body), "missing TelemetryFlow API Key Secret")
}

func TestReceiver_V2Auth_ValidateSecret_BadFormat_401(t *testing.T) {
	cfg := httpOnlyCfg(t, true, true, nil)
	sink := new(consumertest.TracesSink)
	startTracesReceiver(t, cfg, sink)

	url := fmt.Sprintf("http://%s/v2/traces", cfg.Protocols.HTTP.NetAddr.Endpoint)
	resp, body := doPost(t, url,
		map[string]string{
			"X-TelemetryFlow-Key-ID":     "tfk_test",
			"X-TelemetryFlow-Key-Secret": "bad_secret",
		}, nil)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, string(body), "invalid TelemetryFlow API Key Secret format")
}

func TestReceiver_V2Auth_NotRequired_OK(t *testing.T) {
	cfg := httpOnlyCfg(t, false, false, nil) // v2_auth.required = false
	sink := new(consumertest.TracesSink)
	startTracesReceiver(t, cfg, sink)

	td := ptrace.NewTraces()
	td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty().SetName("open")
	data, _ := ptraceotlp.NewExportRequestFromTraces(td).MarshalProto()

	url := fmt.Sprintf("http://%s/v2/traces", cfg.Protocols.HTTP.NetAddr.Endpoint)
	// No auth headers at all — should pass because required=false.
	resp, _ := doPost(t, url, nil, data)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// =============================================================================
// handle* — v1 happy paths for traces, metrics, logs
// =============================================================================

func TestReceiver_V1Traces_Success(t *testing.T) {
	cfg := httpOnlyCfg(t, true, false, nil)
	sink := new(consumertest.TracesSink)
	startTracesReceiver(t, cfg, sink)

	td := ptrace.NewTraces()
	rs := td.ResourceSpans().AppendEmpty()
	ss := rs.ScopeSpans().AppendEmpty()
	sp := ss.Spans().AppendEmpty()
	sp.SetName("v1-trace")
	sp.SetTraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	sp.SetSpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	data, _ := ptraceotlp.NewExportRequestFromTraces(td).MarshalProto()

	url := fmt.Sprintf("http://%s/v1/traces", cfg.Protocols.HTTP.NetAddr.Endpoint)
	resp, _ := doPost(t, url, nil, data)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.Eventually(t, func() bool { return sink.SpanCount() == 1 }, time.Second, 10*time.Millisecond)
}

func TestReceiver_V1Traces_JSON(t *testing.T) {
	cfg := httpOnlyCfg(t, true, false, nil)
	sink := new(consumertest.TracesSink)
	startTracesReceiver(t, cfg, sink)

	td := ptrace.NewTraces()
	td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty().SetName("json-trace")
	data, _ := ptraceotlp.NewExportRequestFromTraces(td).MarshalJSON()

	url := fmt.Sprintf("http://%s/v1/traces", cfg.Protocols.HTTP.NetAddr.Endpoint)
	resp, _ := doPostWithCT(t, url, nil, data, "application/json")
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReceiver_V1Traces_BadBody_400(t *testing.T) {
	cfg := httpOnlyCfg(t, true, false, nil)
	sink := new(consumertest.TracesSink)
	startTracesReceiver(t, cfg, sink)

	url := fmt.Sprintf("http://%s/v1/traces", cfg.Protocols.HTTP.NetAddr.Endpoint)
	resp, body := doPost(t, url, nil, []byte("not-protobuf"))
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), "Failed to unmarshal traces")
}

func TestReceiver_V1Traces_MethodNotAllowed(t *testing.T) {
	cfg := httpOnlyCfg(t, true, false, nil)
	sink := new(consumertest.TracesSink)
	startTracesReceiver(t, cfg, sink)

	url := fmt.Sprintf("http://%s/v1/traces", cfg.Protocols.HTTP.NetAddr.Endpoint)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestReceiver_V1Metrics_Success(t *testing.T) {
	cfg := httpOnlyCfg(t, true, false, nil)
	sink := new(consumertest.MetricsSink)
	startMetricsReceiver(t, cfg, sink)

	md := pmetric.NewMetrics()
	m := md.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	m.SetName("v1-metric")
	m.SetEmptyGauge().DataPoints().AppendEmpty().SetIntValue(42)
	data, _ := pmetricotlp.NewExportRequestFromMetrics(md).MarshalProto()

	url := fmt.Sprintf("http://%s/v1/metrics", cfg.Protocols.HTTP.NetAddr.Endpoint)
	resp, _ := doPost(t, url, nil, data)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReceiver_V1Metrics_BadBody_400(t *testing.T) {
	cfg := httpOnlyCfg(t, true, false, nil)
	sink := new(consumertest.MetricsSink)
	startMetricsReceiver(t, cfg, sink)

	url := fmt.Sprintf("http://%s/v1/metrics", cfg.Protocols.HTTP.NetAddr.Endpoint)
	resp, body := doPost(t, url, nil, []byte("garbage"))
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), "Failed to unmarshal metrics")
}

func TestReceiver_V1Metrics_MethodNotAllowed(t *testing.T) {
	cfg := httpOnlyCfg(t, true, false, nil)
	sink := new(consumertest.MetricsSink)
	startMetricsReceiver(t, cfg, sink)

	url := fmt.Sprintf("http://%s/v1/metrics", cfg.Protocols.HTTP.NetAddr.Endpoint)
	req, _ := http.NewRequest(http.MethodPut, url, strings.NewReader(""))
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestReceiver_V1Logs_Success(t *testing.T) {
	cfg := httpOnlyCfg(t, true, false, nil)
	sink := new(consumertest.LogsSink)
	startLogsReceiver(t, cfg, sink)

	ld := plog.NewLogs()
	lr := ld.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	lr.Body().SetStr("v1-log")
	data, _ := plogotlp.NewExportRequestFromLogs(ld).MarshalProto()

	url := fmt.Sprintf("http://%s/v1/logs", cfg.Protocols.HTTP.NetAddr.Endpoint)
	resp, _ := doPost(t, url, nil, data)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestReceiver_V1Logs_BadBody_400(t *testing.T) {
	cfg := httpOnlyCfg(t, true, false, nil)
	sink := new(consumertest.LogsSink)
	startLogsReceiver(t, cfg, sink)

	url := fmt.Sprintf("http://%s/v1/logs", cfg.Protocols.HTTP.NetAddr.Endpoint)
	resp, body := doPost(t, url, nil, []byte("garbage"))
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), "Failed to unmarshal logs")
}

func TestReceiver_V1Logs_MethodNotAllowed(t *testing.T) {
	cfg := httpOnlyCfg(t, true, false, nil)
	sink := new(consumertest.LogsSink)
	startLogsReceiver(t, cfg, sink)

	url := fmt.Sprintf("http://%s/v1/logs", cfg.Protocols.HTTP.NetAddr.Endpoint)
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

// =============================================================================
// gRPC Export handlers — traces, metrics, logs over real gRPC
// =============================================================================

func TestReceiver_GRPC_Traces(t *testing.T) {
	cfg := grpcHTTPCfg(t)
	sink := new(consumertest.TracesSink)

	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	r, err := factory.CreateTraces(context.Background(), set, cfg, sink)
	require.NoError(t, err)
	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { _ = r.Shutdown(context.Background()) })
	time.Sleep(100 * time.Millisecond)

	cc, err := grpc.NewClient(cfg.Protocols.GRPC.NetAddr.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer func() { _ = cc.Close() }()
	client := ptraceotlp.NewGRPCClient(cc)
	td := ptrace.NewTraces()
	sp := td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	sp.SetName("grpc-trace")
	sp.SetTraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	sp.SetSpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8})

	_, err = client.Export(context.Background(), ptraceotlp.NewExportRequestFromTraces(td))
	require.NoError(t, err)
	require.Eventually(t, func() bool { return sink.SpanCount() == 1 }, time.Second, 10*time.Millisecond)
}

func TestReceiver_GRPC_Metrics(t *testing.T) {
	cfg := grpcHTTPCfg(t)
	sink := new(consumertest.MetricsSink)

	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	r, err := factory.CreateMetrics(context.Background(), set, cfg, sink)
	require.NoError(t, err)
	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { _ = r.Shutdown(context.Background()) })
	time.Sleep(100 * time.Millisecond)

	cc, err := grpc.NewClient(cfg.Protocols.GRPC.NetAddr.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer func() { _ = cc.Close() }()
	client := pmetricotlp.NewGRPCClient(cc)
	md := pmetric.NewMetrics()
	m := md.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	m.SetName("grpc-metric")
	m.SetEmptyGauge().DataPoints().AppendEmpty().SetIntValue(7)

	_, err = client.Export(context.Background(), pmetricotlp.NewExportRequestFromMetrics(md))
	require.NoError(t, err)
	require.Eventually(t, func() bool { return sink.DataPointCount() == 1 }, time.Second, 10*time.Millisecond)
}

func TestReceiver_GRPC_Logs(t *testing.T) {
	cfg := grpcHTTPCfg(t)
	sink := new(consumertest.LogsSink)

	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	r, err := factory.CreateLogs(context.Background(), set, cfg, sink)
	require.NoError(t, err)
	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { _ = r.Shutdown(context.Background()) })
	time.Sleep(100 * time.Millisecond)

	cc, err := grpc.NewClient(cfg.Protocols.GRPC.NetAddr.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer func() { _ = cc.Close() }()
	client := plogotlp.NewGRPCClient(cc)
	ld := plog.NewLogs()
	ld.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty().Body().SetStr("grpc-log")

	_, err = client.Export(context.Background(), plogotlp.NewExportRequestFromLogs(ld))
	require.NoError(t, err)
	require.Eventually(t, func() bool { return sink.LogRecordCount() == 1 }, time.Second, 10*time.Millisecond)
}

// =============================================================================
// Start / Shutdown edge cases
// =============================================================================

func TestReceiver_Start_NoProtocols(t *testing.T) {
	// Both protocols nil — Start should succeed without binding anything.
	cfg := &tfootlpreceiver.Config{
		Protocols:         tfootlpreceiver.ProtocolsConfig{},
		EnableV2Endpoints: true,
	}
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	r, err := factory.CreateTraces(context.Background(), set, cfg, new(consumertest.TracesSink))
	require.NoError(t, err)
	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	require.NoError(t, r.Shutdown(context.Background()))
}

func TestReceiver_Shutdown_NotStarted(t *testing.T) {
	// Shutdown on a never-started receiver must not panic.
	cfg := &tfootlpreceiver.Config{
		Protocols:         tfootlpreceiver.ProtocolsConfig{},
		EnableV2Endpoints: true,
	}
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	r, err := factory.CreateTraces(context.Background(), set, cfg, new(consumertest.TracesSink))
	require.NoError(t, err)
	// Skip Start; go straight to Shutdown.
	require.NoError(t, r.Shutdown(context.Background()))
}

// =============================================================================
// Config validation
// =============================================================================

func TestConfig_Validate_SecretWithAllowList(t *testing.T) {
	cfg := &tfootlpreceiver.Config{
		Protocols: tfootlpreceiver.ProtocolsConfig{
			HTTP: &tfootlpreceiver.HTTPConfig{TracesURLPath: "/v1/traces"},
		},
		EnableV2Endpoints: true,
		V2Auth: tfootlpreceiver.V2AuthConfig{
			Required:       true,
			ValidateSecret: true,
			ValidAPIKeyIDs: []string{"tfk_one", "tfk_two"},
		},
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires tfoauth extension")
}

func TestConfig_Validate_OK(t *testing.T) {
	cfg := &tfootlpreceiver.Config{
		Protocols: tfootlpreceiver.ProtocolsConfig{
			HTTP: &tfootlpreceiver.HTTPConfig{TracesURLPath: "/v1/traces"},
		},
		EnableV2Endpoints: true,
		V2Auth:            tfootlpreceiver.V2AuthConfig{Required: true},
	}
	assert.NoError(t, cfg.Validate())
}

// =============================================================================
// helpers
// =============================================================================

func doPost(t *testing.T, url string, headers map[string]string, body []byte) (*http.Response, []byte) {
	t.Helper()
	return doPostWithCT(t, url, headers, body, "application/x-protobuf")
}

func doPostWithCT(t *testing.T, url string, headers map[string]string, body []byte, contentType string) (*http.Response, []byte) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(http.MethodPost, url, reader)
	require.NoError(t, err)
	req.Header.Set("Content-Type", contentType)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp, respBody
}
