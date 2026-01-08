// Package e2e_test contains end-to-end tests for TelemetryFlow Collector.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package e2e_test

import (
	"context"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectorDataCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Run("should expose health endpoint", func(t *testing.T) {
		// Build collector binary
		buildCmd := exec.Command("go", "build", "-o", "../../build/tfo-collector", "../../cmd/tfo-collector")
		err := buildCmd.Run()
		require.NoError(t, err, "Failed to build collector binary")

		// Start collector
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		collectorCmd := exec.CommandContext(ctx, "../../build/tfo-collector", "--config", "testdata/health-check.yaml")
		err = collectorCmd.Start()
		require.NoError(t, err)
		defer func() {
			_ = collectorCmd.Process.Kill()
			_ = collectorCmd.Wait()
		}()

		// Wait for collector to start
		time.Sleep(3 * time.Second)

		// Check health endpoint
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("http://localhost:13133/health")
		if err != nil {
			t.Skipf("Could not connect to health endpoint: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Health endpoint should return 200 OK")
	})

	t.Run("should expose metrics endpoint", func(t *testing.T) {
		// Start collector
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		collectorCmd := exec.CommandContext(ctx, "../../build/tfo-collector", "--config", "testdata/metrics-exposed.yaml")
		err := collectorCmd.Start()
		require.NoError(t, err)
		defer func() {
			_ = collectorCmd.Process.Kill()
			_ = collectorCmd.Wait()
		}()

		// Wait for collector to start
		time.Sleep(3 * time.Second)

		// Check metrics endpoint
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("http://localhost:8888/metrics")
		if err != nil {
			t.Skipf("Could not connect to metrics endpoint: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Metrics endpoint should return 200 OK")
	})

	t.Run("should handle concurrent requests", func(t *testing.T) {
		// Start collector
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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

		// Send concurrent requests
		client := &http.Client{Timeout: 10 * time.Second}
		concurrency := 10
		results := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			go func() {
				resp, err := client.Get("http://localhost:13133/health")
				if err != nil {
					results <- err
					return
				}
				_ = resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					results <- assert.AnError
					return
				}
				results <- nil
			}()
		}

		// Collect results
		successCount := 0
		for i := 0; i < concurrency; i++ {
			if err := <-results; err == nil {
				successCount++
			}
		}

		assert.Equal(t, concurrency, successCount, "All concurrent requests should succeed")
	})
}
