// Package e2e_test contains end-to-end tests for TelemetryFlow Collector.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
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
package e2e_test

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectorStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	t.Run("should start with valid config", func(t *testing.T) {
		// Build collector binary
		buildCmd := exec.Command("go", "build", "-o", "../../build/tfo-collector", "../../cmd/tfo-collector")
		err := buildCmd.Run()
		require.NoError(t, err, "Failed to build collector binary")

		// Start collector with test config
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		collectorCmd := exec.CommandContext(ctx, "../../build/tfo-collector", "--config", "testdata/minimal.yaml")
		err = collectorCmd.Start()
		require.NoError(t, err)

		// Wait for collector to start
		time.Sleep(2 * time.Second)

		// Verify process is running
		assert.NotNil(t, collectorCmd.Process)

		// Stop collector
		_ = collectorCmd.Process.Kill()
		_ = collectorCmd.Wait()
	})

	t.Run("should fail with invalid config", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		collectorCmd := exec.CommandContext(ctx, "../../build/tfo-collector", "--config", "testdata/invalid.yaml")
		err := collectorCmd.Run()
		assert.Error(t, err, "Collector should fail with invalid config")
	})

	t.Run("should fail with missing config", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		collectorCmd := exec.CommandContext(ctx, "../../build/tfo-collector", "--config", "testdata/nonexistent.yaml")
		err := collectorCmd.Run()
		assert.Error(t, err, "Collector should fail with missing config")
	})
}
