// Package plugin provides exports for testing.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 DevOpsCorner Indonesia. All rights reserved.
//
// This file exports internal types and methods for testing purposes only.
package plugin

// TestRegistryExports provides access to unexported Registry fields for testing.
type TestRegistryExports struct {
	R *Registry
}

// ComponentsLen returns the number of registered components.
func (e *TestRegistryExports) ComponentsLen() int {
	return len(e.R.components)
}

// ExportRegistry wraps a Registry for testing access.
func ExportRegistry(r *Registry) *TestRegistryExports {
	return &TestRegistryExports{R: r}
}
