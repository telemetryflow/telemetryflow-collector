// Package version_test provides unit tests for the version package.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package version_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/telemetryflow/telemetryflow-collector/internal/version"
)

func TestVersion(t *testing.T) {
	t.Run("should return version string", func(t *testing.T) {
		v := version.Version
		require.NotEmpty(t, v)
	})

	t.Run("should have a product name", func(t *testing.T) {
		assert.NotEmpty(t, version.ProductName)
		assert.Equal(t, "TelemetryFlow Collector", version.ProductName)
	})

	t.Run("should have a product short name", func(t *testing.T) {
		assert.NotEmpty(t, version.ProductShortName)
		assert.Equal(t, "tfo-collector", version.ProductShortName)
	})
}

func TestShort(t *testing.T) {
	t.Run("should return short version", func(t *testing.T) {
		short := version.Short()
		assert.NotEmpty(t, short)
	})
}

func TestFull(t *testing.T) {
	t.Run("should return full version info", func(t *testing.T) {
		full := version.Full()
		assert.NotEmpty(t, full)
		assert.Contains(t, full, version.ProductName)
	})
}

func TestBanner(t *testing.T) {
	t.Run("should return ASCII art banner", func(t *testing.T) {
		banner := version.Banner()

		require.NotEmpty(t, banner)
		assert.Contains(t, banner, version.ProductName)
		assert.Contains(t, banner, version.Short())
	})

	t.Run("should contain copyright notice", func(t *testing.T) {
		banner := version.Banner()

		assert.True(t, strings.Contains(banner, "TelemetryFlow") || strings.Contains(banner, "DevOpsCorner"))
	})
}

func TestBuildInfo(t *testing.T) {
	t.Run("should return build information map", func(t *testing.T) {
		info := version.BuildInfo()

		require.NotNil(t, info)
		assert.Contains(t, info, "version")
		assert.Contains(t, info, "product_name")
	})

	t.Run("should have git information", func(t *testing.T) {
		info := version.BuildInfo()

		assert.Contains(t, info, "git_commit")
		assert.Contains(t, info, "git_branch")
	})

	t.Run("should have build time", func(t *testing.T) {
		info := version.BuildInfo()

		assert.Contains(t, info, "build_time")
	})
}
