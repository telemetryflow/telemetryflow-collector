// Package config provides exports for testing.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 DevOpsCorner Indonesia. All rights reserved.
//
// This file exports internal types and methods for testing purposes only.
package config

// TestLoaderExports provides access to unexported Loader fields for testing.
type TestLoaderExports struct {
	L *Loader
}

// Sources returns the loader sources.
func (e *TestLoaderExports) Sources() []Source {
	return e.L.sources
}

// EnvPrefix returns the env prefix.
func (e *TestLoaderExports) EnvPrefix() string {
	return e.L.envPrefix
}

// ConfigPaths returns the config paths.
func (e *TestLoaderExports) ConfigPaths() []string {
	return e.L.configPaths
}

// ExportLoader wraps a Loader for testing access.
func ExportLoader(l *Loader) *TestLoaderExports {
	return &TestLoaderExports{L: l}
}

// ExportDefaultLoader returns the default loader for testing.
func ExportDefaultLoader() *Loader {
	return defaultLoader
}
