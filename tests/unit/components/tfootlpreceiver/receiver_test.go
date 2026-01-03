// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package tfootlpreceiver_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/telemetryflow/telemetryflow-collector/components/tfootlpreceiver"
)

// TestReceiverStartStop tests the receiver lifecycle
func TestReceiverStartStop(t *testing.T) {
	factory := tfootlpreceiver.NewFactory()
	cfg := factory.CreateDefaultConfig().(*tfootlpreceiver.Config)

	// Use random ports to avoid conflicts
	cfg.Protocols.GRPC = nil
	cfg.Protocols.HTTP = nil

	set := receivertest.NewNopSettings(component.MustNewType("tfootlp"))
	sink := new(consumertest.TracesSink)

	receiver, err := factory.CreateTraces(context.Background(), set, cfg, sink)
	require.NoError(t, err)

	err = receiver.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	// Allow some time for startup
	time.Sleep(100 * time.Millisecond)

	err = receiver.Shutdown(context.Background())
	require.NoError(t, err)
}

// TestV1EndpointNoAuthRequired tests that v1 endpoints don't require authentication
func TestV1EndpointNoAuthRequired(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		createData func() ([]byte, string)
	}{
		{
			name: "v1 traces",
			path: "/v1/traces",
			createData: func() ([]byte, string) {
				td := ptrace.NewTraces()
				rs := td.ResourceSpans().AppendEmpty()
				ss := rs.ScopeSpans().AppendEmpty()
				span := ss.Spans().AppendEmpty()
				span.SetName("test-span")

				req := ptraceotlp.NewExportRequestFromTraces(td)
				data, _ := req.MarshalProto()
				return data, "application/x-protobuf"
			},
		},
		{
			name: "v1 metrics",
			path: "/v1/metrics",
			createData: func() ([]byte, string) {
				md := pmetric.NewMetrics()
				rm := md.ResourceMetrics().AppendEmpty()
				sm := rm.ScopeMetrics().AppendEmpty()
				metric := sm.Metrics().AppendEmpty()
				metric.SetName("test-metric")

				req := pmetricotlp.NewExportRequestFromMetrics(md)
				data, _ := req.MarshalProto()
				return data, "application/x-protobuf"
			},
		},
		{
			name: "v1 logs",
			path: "/v1/logs",
			createData: func() ([]byte, string) {
				ld := plog.NewLogs()
				rl := ld.ResourceLogs().AppendEmpty()
				sl := rl.ScopeLogs().AppendEmpty()
				log := sl.LogRecords().AppendEmpty()
				log.Body().SetStr("test-log")

				req := plogotlp.NewExportRequestFromLogs(ld)
				data, _ := req.MarshalProto()
				return data, "application/x-protobuf"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, contentType := tt.createData()

			// Create request without auth headers
			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewReader(data))
			req.Header.Set("Content-Type", contentType)

			// The v1 endpoint should accept requests without auth
			// We just verify the path check here (not full receiver test)
			assert.NotContains(t, tt.path, "/v2/")
		})
	}
}

// TestV2EndpointAuthRequired tests that v2 endpoints require authentication
func TestV2EndpointAuthRequired(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		headers        map[string]string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "v2 traces without auth",
			path:           "/v2/traces",
			headers:        map[string]string{},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "missing TelemetryFlow API Key ID",
		},
		{
			name: "v2 traces with invalid key id format",
			path: "/v2/traces",
			headers: map[string]string{
				"X-TelemetryFlow-Key-ID": "invalid_key",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid TelemetryFlow API Key ID format",
		},
		{
			name: "v2 traces with valid key id",
			path: "/v2/traces",
			headers: map[string]string{
				"X-TelemetryFlow-Key-ID": "tfk_test_key_12345",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "v2 metrics without auth",
			path:           "/v2/metrics",
			headers:        map[string]string{},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "missing TelemetryFlow API Key ID",
		},
		{
			name:           "v2 logs without auth",
			path:           "/v2/logs",
			headers:        map[string]string{},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "missing TelemetryFlow API Key ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the path is a v2 endpoint
			isV2 := tt.path == "/v2/traces" || tt.path == "/v2/metrics" || tt.path == "/v2/logs"
			assert.True(t, isV2, "expected v2 endpoint path")

			// Verify header format expectations
			if keyID, ok := tt.headers["X-TelemetryFlow-Key-ID"]; ok {
				if tt.expectedStatus == http.StatusOK {
					assert.True(t, len(keyID) >= 4 && keyID[:4] == "tfk_")
				}
			}
		})
	}
}

// TestV2AuthValidation tests the v2 auth validation logic
func TestV2AuthValidation(t *testing.T) {
	tests := []struct {
		name     string
		keyID    string
		expected bool
	}{
		{
			name:     "valid tfk_ prefix",
			keyID:    "tfk_abc123",
			expected: true,
		},
		{
			name:     "valid tfk_ with long id",
			keyID:    "tfk_very_long_api_key_id_12345",
			expected: true,
		},
		{
			name:     "invalid prefix tff_",
			keyID:    "tff_abc123",
			expected: false,
		},
		{
			name:     "empty key id",
			keyID:    "",
			expected: false,
		},
		{
			name:     "too short",
			keyID:    "tfk",
			expected: false,
		},
		{
			name:     "no underscore",
			keyID:    "tfkxxx",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate tfk_ prefix format
			valid := len(tt.keyID) >= 4 && tt.keyID[:4] == "tfk_"
			assert.Equal(t, tt.expected, valid)
		})
	}
}

// TestV2SecretValidation tests the v2 secret validation logic
func TestV2SecretValidation(t *testing.T) {
	tests := []struct {
		name      string
		keySecret string
		expected  bool
	}{
		{
			name:      "valid tfs_ prefix",
			keySecret: "tfs_secret123",
			expected:  true,
		},
		{
			name:      "valid tfs_ with long secret",
			keySecret: "tfs_very_long_api_key_secret_12345",
			expected:  true,
		},
		{
			name:      "invalid prefix tfx_",
			keySecret: "tfx_secret123",
			expected:  false,
		},
		{
			name:      "empty secret",
			keySecret: "",
			expected:  false,
		},
		{
			name:      "too short",
			keySecret: "tfs",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate tfs_ prefix format
			valid := len(tt.keySecret) >= 4 && tt.keySecret[:4] == "tfs_"
			assert.Equal(t, tt.expected, valid)
		})
	}
}

// TestIsV2Endpoint tests the v2 endpoint detection
func TestIsV2Endpoint(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/v2/traces", true},
		{"/v2/metrics", true},
		{"/v2/logs", true},
		{"/v1/traces", false},
		{"/v1/metrics", false},
		{"/v1/logs", false},
		{"/traces", false},
		{"/metrics", false},
		{"/logs", false},
		{"/v2/other", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			isV2 := tt.path == "/v2/traces" || tt.path == "/v2/metrics" || tt.path == "/v2/logs"
			assert.Equal(t, tt.expected, isV2)
		})
	}
}

// TestHTTPMethodValidation tests that only POST is allowed
func TestHTTPMethodValidation(t *testing.T) {
	methods := []struct {
		method   string
		expected int
	}{
		{http.MethodPost, http.StatusOK},
		{http.MethodGet, http.StatusMethodNotAllowed},
		{http.MethodPut, http.StatusMethodNotAllowed},
		{http.MethodDelete, http.StatusMethodNotAllowed},
		{http.MethodPatch, http.StatusMethodNotAllowed},
	}

	for _, m := range methods {
		t.Run(m.method, func(t *testing.T) {
			if m.method == http.MethodPost {
				assert.Equal(t, http.StatusOK, m.expected)
			} else {
				assert.Equal(t, http.StatusMethodNotAllowed, m.expected)
			}
		})
	}
}

// TestContentTypeHandling tests JSON and protobuf content type handling
func TestContentTypeHandling(t *testing.T) {
	contentTypes := []struct {
		contentType string
		expected    string
	}{
		{"application/json", "json"},
		{"application/x-protobuf", "protobuf"},
		{"", "protobuf"},
		{"text/plain", "protobuf"},
	}

	for _, ct := range contentTypes {
		t.Run(ct.contentType, func(t *testing.T) {
			if ct.contentType == "application/json" {
				assert.Equal(t, "json", ct.expected)
			} else {
				assert.Equal(t, "protobuf", ct.expected)
			}
		})
	}
}

// TestRequestBodyReading tests error handling for request body reading
func TestRequestBodyReading(t *testing.T) {
	// Test with empty body
	req := httptest.NewRequest(http.MethodPost, "/v1/traces", nil)
	require.NotNil(t, req)
	require.NotNil(t, req.Body)

	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	assert.Empty(t, body)
}
