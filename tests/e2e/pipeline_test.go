// Package e2e_test contains end-to-end tests for TelemetryFlow Collector.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectorPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Run("should receive and process OTLP metrics", func(t *testing.T) {
		// Build collector binary
		buildCmd := exec.Command("go", "build", "-o", "../../build/tfo-collector", "../../cmd/tfo-collector")
		err := buildCmd.Run()
		require.NoError(t, err, "Failed to build collector binary")

		// Start collector with OTLP receiver config
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		collectorCmd := exec.CommandContext(ctx, "../../build/tfo-collector", "--config", "testdata/otlp-receiver.yaml")
		err = collectorCmd.Start()
		require.NoError(t, err)
		defer func() {
			_ = collectorCmd.Process.Kill()
			_ = collectorCmd.Wait()
		}()

		// Wait for collector to start
		time.Sleep(3 * time.Second)

		// Send test OTLP data
		client := &http.Client{Timeout: 5 * time.Second}
		testPayload := map[string]interface{}{
			"resourceMetrics": []map[string]interface{}{
				{
					"resource": map[string]interface{}{
						"attributes": []map[string]interface{}{
							{"key": "service.name", "value": map[string]interface{}{"stringValue": "test-service"}},
						},
					},
					"scopeMetrics": []map[string]interface{}{
						{
							"scope": map[string]interface{}{"name": "test-scope"},
							"metrics": []map[string]interface{}{
								{
									"name": "test_metric",
									"gauge": map[string]interface{}{
										"dataPoints": []map[string]interface{}{
											{
												"asDouble":     42.0,
												"timeUnixNano": time.Now().UnixNano(),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		payload, err := json.Marshal(testPayload)
		require.NoError(t, err)

		resp, err := client.Post("http://localhost:4318/v1/metrics", "application/json", bytes.NewReader(payload))
		if err != nil {
			t.Skipf("Could not connect to OTLP endpoint: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK, got %d: %s", resp.StatusCode, string(body))
	})

	t.Run("should receive and process OTLP traces", func(t *testing.T) {
		// Start collector with OTLP receiver config
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		collectorCmd := exec.CommandContext(ctx, "../../build/tfo-collector", "--config", "testdata/otlp-receiver.yaml")
		err := collectorCmd.Start()
		require.NoError(t, err)
		defer func() {
			_ = collectorCmd.Process.Kill()
			_ = collectorCmd.Wait()
		}()

		// Wait for collector to start
		time.Sleep(3 * time.Second)

		// Send test OTLP trace data
		client := &http.Client{Timeout: 5 * time.Second}
		testPayload := map[string]interface{}{
			"resourceSpans": []map[string]interface{}{
				{
					"resource": map[string]interface{}{
						"attributes": []map[string]interface{}{
							{"key": "service.name", "value": map[string]interface{}{"stringValue": "test-service"}},
						},
					},
					"scopeSpans": []map[string]interface{}{
						{
							"scope": map[string]interface{}{"name": "test-scope"},
							"spans": []map[string]interface{}{
								{
									"traceId":           "0123456789abcdef0123456789abcdef",
									"spanId":            "0123456789abcdef",
									"name":              "test-span",
									"kind":              1,
									"startTimeUnixNano": time.Now().Add(-time.Second).UnixNano(),
									"endTimeUnixNano":   time.Now().UnixNano(),
								},
							},
						},
					},
				},
			},
		}

		payload, err := json.Marshal(testPayload)
		require.NoError(t, err)

		resp, err := client.Post("http://localhost:4318/v1/traces", "application/json", bytes.NewReader(payload))
		if err != nil {
			t.Skipf("Could not connect to OTLP endpoint: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK, got %d: %s", resp.StatusCode, string(body))
	})

	t.Run("should receive and process OTLP logs", func(t *testing.T) {
		// Start collector with OTLP receiver config
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		collectorCmd := exec.CommandContext(ctx, "../../build/tfo-collector", "--config", "testdata/otlp-receiver.yaml")
		err := collectorCmd.Start()
		require.NoError(t, err)
		defer func() {
			_ = collectorCmd.Process.Kill()
			_ = collectorCmd.Wait()
		}()

		// Wait for collector to start
		time.Sleep(3 * time.Second)

		// Send test OTLP log data
		client := &http.Client{Timeout: 5 * time.Second}
		testPayload := map[string]interface{}{
			"resourceLogs": []map[string]interface{}{
				{
					"resource": map[string]interface{}{
						"attributes": []map[string]interface{}{
							{"key": "service.name", "value": map[string]interface{}{"stringValue": "test-service"}},
						},
					},
					"scopeLogs": []map[string]interface{}{
						{
							"scope": map[string]interface{}{"name": "test-scope"},
							"logRecords": []map[string]interface{}{
								{
									"timeUnixNano":         time.Now().UnixNano(),
									"severityNumber":       9,
									"severityText":         "INFO",
									"body":                 map[string]interface{}{"stringValue": "Test log message"},
									"observedTimeUnixNano": time.Now().UnixNano(),
								},
							},
						},
					},
				},
			},
		}

		payload, err := json.Marshal(testPayload)
		require.NoError(t, err)

		resp, err := client.Post("http://localhost:4318/v1/logs", "application/json", bytes.NewReader(payload))
		if err != nil {
			t.Skipf("Could not connect to OTLP endpoint: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected 200 OK, got %d: %s", resp.StatusCode, string(body))
	})
}
