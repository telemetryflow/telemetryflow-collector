// Package e2e_test contains end-to-end tests for TelemetryFlow Collector.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package e2e_test

import (
	"context"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCollectorShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Run("should shutdown gracefully on SIGTERM", func(t *testing.T) {
		// Build collector binary
		buildCmd := exec.Command("go", "build", "-o", "../../build/tfo-collector", "../../cmd/tfo-collector")
		err := buildCmd.Run()
		require.NoError(t, err, "Failed to build collector binary")

		// Start collector
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		collectorCmd := exec.CommandContext(ctx, "../../build/tfo-collector", "--config", "testdata/minimal.yaml")
		err = collectorCmd.Start()
		require.NoError(t, err)

		// Wait for collector to start
		time.Sleep(2 * time.Second)

		// Send SIGTERM
		err = collectorCmd.Process.Signal(syscall.SIGTERM)
		require.NoError(t, err)

		// Wait for graceful shutdown
		done := make(chan error, 1)
		go func() {
			done <- collectorCmd.Wait()
		}()

		select {
		case <-done:
			// Graceful shutdown completed (process terminated within timeout)
		case <-time.After(15 * time.Second):
			_ = collectorCmd.Process.Kill()
			t.Fatal("Collector did not shutdown within timeout")
		}
	})

	t.Run("should shutdown gracefully on SIGINT", func(t *testing.T) {
		// Build collector binary
		buildCmd := exec.Command("go", "build", "-o", "../../build/tfo-collector", "../../cmd/tfo-collector")
		err := buildCmd.Run()
		require.NoError(t, err, "Failed to build collector binary")

		// Start collector
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		collectorCmd := exec.CommandContext(ctx, "../../build/tfo-collector", "--config", "testdata/minimal.yaml")
		err = collectorCmd.Start()
		require.NoError(t, err)

		// Wait for collector to start
		time.Sleep(2 * time.Second)

		// Send SIGINT
		err = collectorCmd.Process.Signal(os.Interrupt)
		require.NoError(t, err)

		// Wait for graceful shutdown
		done := make(chan error, 1)
		go func() {
			done <- collectorCmd.Wait()
		}()

		select {
		case <-done:
			// Graceful shutdown completed
		case <-time.After(15 * time.Second):
			_ = collectorCmd.Process.Kill()
			t.Fatal("Collector did not shutdown within timeout")
		}
	})
}
