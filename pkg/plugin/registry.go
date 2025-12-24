// Package plugin provides a plugin registry system for TelemetryFlow Collector.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 DevOpsCorner Indonesia. All rights reserved.
//
// LEGO Building Block - Flexible plugin system for OTEL components.
package plugin

import (
	"fmt"
	"sync"
)

// ComponentType represents OTEL component types
type ComponentType string

const (
	// TypeReceiver is an OTEL receiver component
	TypeReceiver ComponentType = "receiver"

	// TypeProcessor is an OTEL processor component
	TypeProcessor ComponentType = "processor"

	// TypeExporter is an OTEL exporter component
	TypeExporter ComponentType = "exporter"

	// TypeExtension is an OTEL extension component
	TypeExtension ComponentType = "extension"

	// TypeConnector is an OTEL connector component
	TypeConnector ComponentType = "connector"
)

// ComponentInfo contains component metadata
type ComponentInfo struct {
	Name        string
	Type        ComponentType
	Version     string
	Description string
	Stability   string // "stable", "beta", "alpha", "development"
}

// Component is the interface for custom OTEL components
type Component interface {
	// Info returns component metadata
	Info() ComponentInfo

	// Validate validates the component configuration
	Validate(config map[string]interface{}) error
}

// ComponentFactory creates a new component
type ComponentFactory func() Component

// Registry holds registered components
type Registry struct {
	mu         sync.RWMutex
	components map[string]ComponentFactory
}

// NewRegistry creates a new component registry
func NewRegistry() *Registry {
	return &Registry{
		components: make(map[string]ComponentFactory),
	}
}

// Register adds a component factory to the registry
func (r *Registry) Register(name string, factory ComponentFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.components[name]; exists {
		return fmt.Errorf("component %s already registered", name)
	}

	r.components[name] = factory
	return nil
}

// Unregister removes a component from the registry
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.components, name)
}

// Get retrieves a component factory by name
func (r *Registry) Get(name string) (ComponentFactory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, exists := r.components[name]
	if !exists {
		return nil, fmt.Errorf("component %s not found", name)
	}

	return factory, nil
}

// Create creates a new component instance
func (r *Registry) Create(name string) (Component, error) {
	factory, err := r.Get(name)
	if err != nil {
		return nil, err
	}
	return factory(), nil
}

// List returns all registered component names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.components))
	for name := range r.components {
		names = append(names, name)
	}
	return names
}

// ListByType returns component names filtered by type
func (r *Registry) ListByType(t ComponentType) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name, factory := range r.components {
		component := factory()
		if component.Info().Type == t {
			names = append(names, name)
		}
	}
	return names
}

// Has checks if a component is registered
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.components[name]
	return exists
}

// Count returns the number of registered components
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.components)
}

// CountByType returns count of components by type
func (r *Registry) CountByType(t ComponentType) int {
	return len(r.ListByType(t))
}

// Global registries for each component type
var (
	Receivers  = NewRegistry()
	Processors = NewRegistry()
	Exporters  = NewRegistry()
	Extensions = NewRegistry()
	Connectors = NewRegistry()
)

// RegisterReceiver adds a receiver to the global registry
func RegisterReceiver(name string, factory ComponentFactory) error {
	return Receivers.Register(name, factory)
}

// RegisterProcessor adds a processor to the global registry
func RegisterProcessor(name string, factory ComponentFactory) error {
	return Processors.Register(name, factory)
}

// RegisterExporter adds an exporter to the global registry
func RegisterExporter(name string, factory ComponentFactory) error {
	return Exporters.Register(name, factory)
}

// RegisterExtension adds an extension to the global registry
func RegisterExtension(name string, factory ComponentFactory) error {
	return Extensions.Register(name, factory)
}

// RegisterConnector adds a connector to the global registry
func RegisterConnector(name string, factory ComponentFactory) error {
	return Connectors.Register(name, factory)
}

// Summary returns a summary of all registered components
func Summary() map[ComponentType]int {
	return map[ComponentType]int{
		TypeReceiver:  Receivers.Count(),
		TypeProcessor: Processors.Count(),
		TypeExporter:  Exporters.Count(),
		TypeExtension: Extensions.Count(),
		TypeConnector: Connectors.Count(),
	}
}
