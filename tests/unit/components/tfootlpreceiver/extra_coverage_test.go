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
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configgrpc"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/consumer"
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

func newGRPCConfig(port int) configgrpc.ServerConfig {
	return configgrpc.ServerConfig{
		NetAddr: confignet.AddrConfig{
			Endpoint:  fmt.Sprintf("127.0.0.1:%d", port),
			Transport: confignet.TransportTypeTCP,
		},
	}
}

func newHTTPCfg(port int) confighttp.ServerConfig {
	cfg := confighttp.NewDefaultServerConfig()
	cfg.NetAddr.Endpoint = fmt.Sprintf("127.0.0.1:%d", port)
	cfg.NetAddr.Transport = confignet.TransportTypeTCP
	return cfg
}

// failingTracesConsumer always returns err on ConsumeTraces.
type failingTracesConsumer struct{ err error }

func (f *failingTracesConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}
func (f *failingTracesConsumer) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	return f.err
}

type failingMetricsConsumer struct{ err error }

func (f *failingMetricsConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}
func (f *failingMetricsConsumer) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	return f.err
}

type failingLogsConsumer struct{ err error }

func (f *failingLogsConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}
func (f *failingLogsConsumer) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	return f.err
}

// --- HTTP consumer-error paths (returns 500) ---

func TestReceiver_V1Traces_ConsumerError_500(t *testing.T) {
	cfg := httpOnlyCfg(t, true, false, nil)
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	r, err := factory.CreateTraces(context.Background(), set, cfg,
		&failingTracesConsumer{err: errors.New("downstream failure")})
	require.NoError(t, err)
	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { _ = r.Shutdown(context.Background()) })
	time.Sleep(80 * time.Millisecond)

	td := ptrace.NewTraces()
	sp := td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	sp.SetName("fail-trace")
	sp.SetTraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	sp.SetSpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	data, _ := ptraceotlp.NewExportRequestFromTraces(td).MarshalProto()

	url := fmt.Sprintf("http://%s/v1/traces", cfg.Protocols.HTTP.NetAddr.Endpoint)
	resp, body := doPost(t, url, nil, data)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), "Failed to process traces")
}

func TestReceiver_V1Metrics_ConsumerError_500(t *testing.T) {
	cfg := httpOnlyCfg(t, true, false, nil)
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	r, err := factory.CreateMetrics(context.Background(), set, cfg,
		&failingMetricsConsumer{err: errors.New("downstream failure")})
	require.NoError(t, err)
	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { _ = r.Shutdown(context.Background()) })
	time.Sleep(80 * time.Millisecond)

	md := pmetric.NewMetrics()
	m := md.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	m.SetName("fail-metric")
	m.SetEmptyGauge().DataPoints().AppendEmpty().SetIntValue(1)
	data, _ := pmetricotlp.NewExportRequestFromMetrics(md).MarshalProto()

	url := fmt.Sprintf("http://%s/v1/metrics", cfg.Protocols.HTTP.NetAddr.Endpoint)
	resp, body := doPost(t, url, nil, data)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), "Failed to process metrics")
}

func TestReceiver_V1Logs_ConsumerError_500(t *testing.T) {
	cfg := httpOnlyCfg(t, true, false, nil)
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	r, err := factory.CreateLogs(context.Background(), set, cfg,
		&failingLogsConsumer{err: errors.New("downstream failure")})
	require.NoError(t, err)
	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { _ = r.Shutdown(context.Background()) })
	time.Sleep(80 * time.Millisecond)

	ld := plog.NewLogs()
	ld.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty().Body().SetStr("fail-log")
	data, _ := plogotlp.NewExportRequestFromLogs(ld).MarshalProto()

	url := fmt.Sprintf("http://%s/v1/logs", cfg.Protocols.HTTP.NetAddr.Endpoint)
	resp, body := doPost(t, url, nil, data)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, string(body), "Failed to process logs")
}

// --- gRPC consumer-error paths ---

func TestReceiver_GRPC_Traces_ConsumerError(t *testing.T) {
	cfg := grpcHTTPCfg(t)
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	r, err := factory.CreateTraces(context.Background(), set, cfg,
		&failingTracesConsumer{err: errors.New("grpc downstream failure")})
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
	sp.SetName("grpc-fail")
	sp.SetTraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	sp.SetSpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8})

	_, err = client.Export(context.Background(), ptraceotlp.NewExportRequestFromTraces(td))
	require.Error(t, err)
}

func TestReceiver_GRPC_Metrics_ConsumerError(t *testing.T) {
	cfg := grpcHTTPCfg(t)
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	r, err := factory.CreateMetrics(context.Background(), set, cfg,
		&failingMetricsConsumer{err: errors.New("grpc downstream failure")})
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
	m.SetName("grpc-fail")
	m.SetEmptyGauge().DataPoints().AppendEmpty().SetIntValue(1)

	_, err = client.Export(context.Background(), pmetricotlp.NewExportRequestFromMetrics(md))
	require.Error(t, err)
}

func TestReceiver_GRPC_Logs_ConsumerError(t *testing.T) {
	cfg := grpcHTTPCfg(t)
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	r, err := factory.CreateLogs(context.Background(), set, cfg,
		&failingLogsConsumer{err: errors.New("grpc downstream failure")})
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
	ld.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty().Body().SetStr("grpc-fail")

	_, err = client.Export(context.Background(), plogotlp.NewExportRequestFromLogs(ld))
	require.Error(t, err)
}

// --- Start idempotency ---

func TestReceiver_Start_Idempotent(t *testing.T) {
	cfg := grpcHTTPCfg(t)
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	r, err := factory.CreateTraces(context.Background(), set, cfg, new(consumertest.TracesSink))
	require.NoError(t, err)
	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { _ = r.Shutdown(context.Background()) })

	// Second Start on an already-started receiver should be a no-op (returns nil).
	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
}

// --- Start with only gRPC (HTTP disabled) ---

func TestReceiver_Start_GRPCOnly(t *testing.T) {
	port := freePort(t)
	cfg := &tfootlpreceiver.Config{
		Protocols: tfootlpreceiver.ProtocolsConfig{
			GRPC: &tfootlpreceiver.GRPCConfig{
				ServerConfig: newGRPCConfig(port),
			},
			HTTP: nil,
		},
		EnableV2Endpoints: true,
		V2Auth:            tfootlpreceiver.V2AuthConfig{Required: false},
	}
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	r, err := factory.CreateTraces(context.Background(), set, cfg, new(consumertest.TracesSink))
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
	sp.SetName("grpc-only")
	sp.SetTraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16})
	sp.SetSpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	_, err = client.Export(context.Background(), ptraceotlp.NewExportRequestFromTraces(td))
	require.NoError(t, err)
}

// --- Start with only HTTP (gRPC disabled) — exercises startHTTP only ---

func TestReceiver_Start_HTTPOnly_NoV2(t *testing.T) {
	port := freePort(t)
	cfg := &tfootlpreceiver.Config{
		Protocols: tfootlpreceiver.ProtocolsConfig{
			HTTP: &tfootlpreceiver.HTTPConfig{
				ServerConfig:   newHTTPCfg(port),
				TracesURLPath:  "/v1/traces",
				MetricsURLPath: "/v1/metrics",
				LogsURLPath:    "/v1/logs",
			},
			GRPC: nil,
		},
		EnableV2Endpoints: false, // exercise the v2-disabled branch in startHTTP
		V2Auth:            tfootlpreceiver.V2AuthConfig{Required: false},
	}
	factory := tfootlpreceiver.NewFactory()
	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	r, err := factory.CreateTraces(context.Background(), set, cfg, new(consumertest.TracesSink))
	require.NoError(t, err)
	require.NoError(t, r.Start(context.Background(), componenttest.NewNopHost()))
	t.Cleanup(func() { _ = r.Shutdown(context.Background()) })
	time.Sleep(80 * time.Millisecond)

	// v1 should still work, v2 should 404 because EnableV2Endpoints=false.
	td := ptrace.NewTraces()
	td.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty().SetName("v2-disabled")
	data, _ := ptraceotlp.NewExportRequestFromTraces(td).MarshalProto()

	v1URL := fmt.Sprintf("http://%s/v1/traces", cfg.Protocols.HTTP.NetAddr.Endpoint)
	resp, _ := doPost(t, v1URL, nil, data)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// /v2/traces is not registered → 404 from the default mux.
	v2URL := fmt.Sprintf("http://%s/v2/traces", cfg.Protocols.HTTP.NetAddr.Endpoint)
	req, _ := http.NewRequest(http.MethodPost, v2URL, nil)
	v2resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = v2resp.Body.Close() }()
	assert.Equal(t, http.StatusNotFound, v2resp.StatusCode)
}
