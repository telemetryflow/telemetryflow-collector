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
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
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

// ---------------------------------------------------------------------------
// readRequestBody – gzip decompression
// ---------------------------------------------------------------------------

func TestReadRequestBody_Uncompressed(t *testing.T) {
	payload := []byte(`{"resourceLogs":[]}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/logs", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	body, err := readRequestBody(req)
	require.NoError(t, err)
	assert.Equal(t, payload, body)
}

func TestReadRequestBody_GzipCompressed(t *testing.T) {
	payload := []byte(`{"resourceLogs":[]}`)

	var gzBuf bytes.Buffer
	gw := gzip.NewWriter(&gzBuf)
	_, err := gw.Write(payload)
	require.NoError(t, err)
	require.NoError(t, gw.Close())

	req := httptest.NewRequest(http.MethodPost, "/v1/logs", bytes.NewReader(gzBuf.Bytes()))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	body, err := readRequestBody(req)
	require.NoError(t, err)
	assert.Equal(t, payload, body)
}

func TestReadRequestBody_GzipCaseInsensitive(t *testing.T) {
	payload := []byte(`{"resourceLogs":[]}`)

	var gzBuf bytes.Buffer
	gw := gzip.NewWriter(&gzBuf)
	_, err := gw.Write(payload)
	require.NoError(t, err)
	require.NoError(t, gw.Close())

	req := httptest.NewRequest(http.MethodPost, "/v1/logs", bytes.NewReader(gzBuf.Bytes()))
	req.Header.Set("Content-Encoding", "GZIP")

	body, err := readRequestBody(req)
	require.NoError(t, err)
	assert.Equal(t, payload, body)
}

func TestReadRequestBody_InvalidGzip(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/logs", bytes.NewReader([]byte("not-gzip")))
	req.Header.Set("Content-Encoding", "gzip")

	_, err := readRequestBody(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gzip")
}

func TestReadRequestBody_NoEncodingHeader(t *testing.T) {
	payload := []byte{0x0a, 0x0b, 0x0c}
	req := httptest.NewRequest(http.MethodPost, "/v1/metrics", bytes.NewReader(payload))

	body, err := readRequestBody(req)
	require.NoError(t, err)
	assert.Equal(t, payload, body)
}

// ---------------------------------------------------------------------------
// getIdentityHeader – Agent-ID / Collector-ID alias
// ---------------------------------------------------------------------------

func TestGetIdentityHeader_AgentID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v2/logs", nil)
	req.Header.Set(headerAgentID, "agent-123")

	assert.Equal(t, "agent-123", getIdentityHeader(req))
}

func TestGetIdentityHeader_CollectorID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v2/logs", nil)
	req.Header.Set(headerCollectorID, "collector-456")

	assert.Equal(t, "collector-456", getIdentityHeader(req))
}

func TestGetIdentityHeader_AgentIDTakesPriority(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v2/logs", nil)
	req.Header.Set(headerAgentID, "agent-123")
	req.Header.Set(headerCollectorID, "collector-456")

	assert.Equal(t, "agent-123", getIdentityHeader(req))
}

func TestGetIdentityHeader_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v2/logs", nil)

	assert.Equal(t, "", getIdentityHeader(req))
}

// ---------------------------------------------------------------------------
// End-to-end handler tests: gzip-compressed OTLP data through handleX()
// ---------------------------------------------------------------------------

// newTestReceiver builds a fresh tfoOTLPReceiver with nop consumers for
// handler-level testing. It resets the singleton so parallel tests do not
// collide on shared state.
func newTestReceiver(t *testing.T) *tfoOTLPReceiver {
	t.Helper()
	receiverInstanceLock.Lock()
	receiverInstance = nil
	receiverInstanceLock.Unlock()

	settings := &receiver.Settings{}
	settings.Logger = zap.NewNop()
	settings.ID = component.MustNewID("tfootlp")

	r, err := newTFOOTLPReceiver(&Config{}, settings)
	require.NoError(t, err)
	return r
}

// gzipEncode gzip-compresses the given payload.
func gzipEncode(t *testing.T, payload []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write(payload)
	require.NoError(t, err)
	require.NoError(t, gw.Close())
	return buf.Bytes()
}

func TestHandleLogs_GzipProtobuf(t *testing.T) {
	r := newTestReceiver(t)
	sink := new(consumertest.LogsSink)
	r.registerLogsConsumer(sink)

	ld := plog.NewLogs()
	rl := ld.ResourceLogs().AppendEmpty()
	sl := rl.ScopeLogs().AppendEmpty()
	log := sl.LogRecords().AppendEmpty()
	log.Body().SetStr("gzip-test-log")

	exportReq := plogotlp.NewExportRequestFromLogs(ld)
	protoData, err := exportReq.MarshalProto()
	require.NoError(t, err)

	gzData := gzipEncode(t, protoData)

	req := httptest.NewRequest(http.MethodPost, "/v1/logs", bytes.NewReader(gzData))
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "gzip")

	rr := httptest.NewRecorder()
	r.handleLogs(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 1, sink.LogRecordCount())
}

func TestHandleMetrics_GzipProtobuf(t *testing.T) {
	r := newTestReceiver(t)
	sink := new(consumertest.MetricsSink)
	r.registerMetricsConsumer(sink)

	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()
	metric := sm.Metrics().AppendEmpty()
	metric.SetName("gzip-test-metric")
	metric.SetEmptyGauge()
	metric.Gauge().DataPoints().AppendEmpty().SetIntValue(42)

	exportReq := pmetricotlp.NewExportRequestFromMetrics(md)
	protoData, err := exportReq.MarshalProto()
	require.NoError(t, err)

	gzData := gzipEncode(t, protoData)

	req := httptest.NewRequest(http.MethodPost, "/v1/metrics", bytes.NewReader(gzData))
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "gzip")

	rr := httptest.NewRecorder()
	r.handleMetrics(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 1, sink.DataPointCount())
}

func TestHandleTraces_GzipProtobuf(t *testing.T) {
	r := newTestReceiver(t)
	sink := new(consumertest.TracesSink)
	r.registerTracesConsumer(sink)

	td := ptrace.NewTraces()
	rs := td.ResourceSpans().AppendEmpty()
	ss := rs.ScopeSpans().AppendEmpty()
	span := ss.Spans().AppendEmpty()
	span.SetName("gzip-test-span")

	exportReq := ptraceotlp.NewExportRequestFromTraces(td)
	protoData, err := exportReq.MarshalProto()
	require.NoError(t, err)

	gzData := gzipEncode(t, protoData)

	req := httptest.NewRequest(http.MethodPost, "/v1/traces", bytes.NewReader(gzData))
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "gzip")

	rr := httptest.NewRecorder()
	r.handleTraces(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 1, sink.SpanCount())
}

func TestHandleLogs_UncompressedStillWorks(t *testing.T) {
	r := newTestReceiver(t)
	sink := new(consumertest.LogsSink)
	r.registerLogsConsumer(sink)

	ld := plog.NewLogs()
	rl := ld.ResourceLogs().AppendEmpty()
	sl := rl.ScopeLogs().AppendEmpty()
	log := sl.LogRecords().AppendEmpty()
	log.Body().SetStr("plain-test-log")

	exportReq := plogotlp.NewExportRequestFromLogs(ld)
	protoData, err := exportReq.MarshalProto()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/logs", bytes.NewReader(protoData))
	req.Header.Set("Content-Type", "application/x-protobuf")

	rr := httptest.NewRecorder()
	r.handleLogs(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 1, sink.LogRecordCount())
}

func TestHandleLogs_GzipJSON(t *testing.T) {
	r := newTestReceiver(t)
	sink := new(consumertest.LogsSink)
	r.registerLogsConsumer(sink)

	ld := plog.NewLogs()
	rl := ld.ResourceLogs().AppendEmpty()
	sl := rl.ScopeLogs().AppendEmpty()
	log := sl.LogRecords().AppendEmpty()
	log.Body().SetStr("gzip-json-log")

	exportReq := plogotlp.NewExportRequestFromLogs(ld)
	jsonData, err := exportReq.MarshalJSON()
	require.NoError(t, err)

	gzData := gzipEncode(t, jsonData)

	req := httptest.NewRequest(http.MethodPost, "/v1/logs", bytes.NewReader(gzData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	rr := httptest.NewRecorder()
	r.handleLogs(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 1, sink.LogRecordCount())
}

func TestHandleLogs_GzipInvalidData(t *testing.T) {
	r := newTestReceiver(t)
	sink := new(consumertest.LogsSink)
	r.registerLogsConsumer(sink)

	req := httptest.NewRequest(http.MethodPost, "/v1/logs", bytes.NewReader([]byte("corrupt-data")))
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "gzip")

	rr := httptest.NewRecorder()
	r.handleLogs(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, 0, sink.LogRecordCount())
}

// TestHandleLogs_V2AuthWithAgentID verifies that v2 auth passes when the
// agent sends X-TelemetryFlow-Agent-ID instead of X-TelemetryFlow-Collector-ID.
func TestHandleLogs_V2AuthWithAgentID(t *testing.T) {
	r := newTestReceiver(t)
	r.cfg.V2Auth.Required = true
	r.cfg.V2Auth.ValidAPIKeyIDs = []string{"tfk_test_key_12345"}

	sink := new(consumertest.LogsSink)
	r.registerLogsConsumer(sink)

	ld := plog.NewLogs()
	ld.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty().Body().SetStr("v2-agent-id")

	exportReq := plogotlp.NewExportRequestFromLogs(ld)
	protoData, err := exportReq.MarshalProto()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v2/logs", bytes.NewReader(protoData))
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("X-TelemetryFlow-Key-ID", "tfk_test_key_12345")
	req.Header.Set("X-TelemetryFlow-Agent-ID", "agent-abc-123")

	rr := httptest.NewRecorder()
	r.handleLogs(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 1, sink.LogRecordCount())
}
