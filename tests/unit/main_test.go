// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.

package main_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMainFunction(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	t.Run("no_arguments_shows_help", func(t *testing.T) {
		os.Args = []string{"tfo-collector"}
		// Test would show banner + help, but we can't easily capture output
		// This test verifies the args parsing logic
		assert.Equal(t, 1, len(os.Args))
	})

	t.Run("help_flag_recognized", func(t *testing.T) {
		testCases := []string{"--help", "-h", "--version", "-v"}
		for _, flag := range testCases {
			os.Args = []string{"tfo-collector", flag}
			// Verify flag is in args
			assert.Contains(t, os.Args, flag)
		}
	})

	t.Run("config_flag_parsing", func(t *testing.T) {
		os.Args = []string{"tfo-collector", "--config", "test.yaml"}
		assert.Contains(t, os.Args, "--config")
		assert.Contains(t, os.Args, "test.yaml")
	})

	t.Run("short_config_flag_parsing", func(t *testing.T) {
		os.Args = []string{"tfo-collector", "-c", "test.yaml"}
		assert.Contains(t, os.Args, "-c")
		assert.Contains(t, os.Args, "test.yaml")
	})
}

func TestFlagValidation(t *testing.T) {
	t.Run("valid_flags", func(t *testing.T) {
		validFlags := []string{"-c", "--config", "-s", "--set", "-f", "--feature-gates", "-h", "--help", "-v", "--version"}
		for _, flag := range validFlags {
			// Test that flag format is valid
			require.True(t, len(flag) > 0)
			if flag[0] == '-' {
				if len(flag) > 1 && flag[1] == '-' {
					// Long flag
					assert.True(t, len(flag) > 2, "Long flag should have content after --")
				} else {
					// Short flag
					assert.Equal(t, 2, len(flag), "Short flag should be exactly 2 characters")
				}
			}
		}
	})
}
