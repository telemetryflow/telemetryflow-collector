// Package config provides exports for testing.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
//
// This file exports internal types and methods for testing purposes only.
// It is only compiled during testing (due to _test.go suffix).
package config

// TestLoaderExports provides access to unexported Loader fields for testing.
type TestLoaderExports struct {
	L *Loader
}

// ConfigPaths returns the config paths.
func (e *TestLoaderExports) ConfigPaths() []string {
	return e.L.configPaths
}

// EnvPrefix returns the environment prefix.
func (e *TestLoaderExports) EnvPrefix() string {
	return e.L.envPrefix
}

// ExportLoader wraps a Loader for testing access.
func ExportLoader(l *Loader) *TestLoaderExports {
	return &TestLoaderExports{L: l}
}
