// Package plugin provides a plugin registry system for TelemetryFlow Collector.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 DevOpsCorner Indonesia. All rights reserved.
//
// LEGO Building Block - Flexible plugin system for OTEL components.
package plugin

import (
	"sync"
	"testing"
)

// mockComponent is a mock implementation of Component interface
type mockComponent struct {
	info ComponentInfo
}

func (m *mockComponent) Info() ComponentInfo {
	return m.info
}

func (m *mockComponent) Validate(config map[string]interface{}) error {
	return nil
}

// newMockComponent creates a mock component factory
func newMockComponent(name string, t ComponentType) ComponentFactory {
	return func() Component {
		return &mockComponent{
			info: ComponentInfo{
				Name:        name,
				Type:        t,
				Version:     "1.0.0",
				Description: "Mock component for testing",
				Stability:   "stable",
			},
		}
	}
}

// TestComponentTypeConstants tests ComponentType constants
func TestComponentTypeConstants(t *testing.T) {
	if TypeReceiver != "receiver" {
		t.Errorf("Expected TypeReceiver 'receiver', got '%s'", TypeReceiver)
	}

	if TypeProcessor != "processor" {
		t.Errorf("Expected TypeProcessor 'processor', got '%s'", TypeProcessor)
	}

	if TypeExporter != "exporter" {
		t.Errorf("Expected TypeExporter 'exporter', got '%s'", TypeExporter)
	}

	if TypeExtension != "extension" {
		t.Errorf("Expected TypeExtension 'extension', got '%s'", TypeExtension)
	}

	if TypeConnector != "connector" {
		t.Errorf("Expected TypeConnector 'connector', got '%s'", TypeConnector)
	}
}

// TestComponentInfo tests ComponentInfo struct
func TestComponentInfo(t *testing.T) {
	info := ComponentInfo{
		Name:        "test-component",
		Type:        TypeReceiver,
		Version:     "1.0.0",
		Description: "Test component",
		Stability:   "stable",
	}

	if info.Name != "test-component" {
		t.Errorf("Expected Name 'test-component', got '%s'", info.Name)
	}

	if info.Type != TypeReceiver {
		t.Errorf("Expected Type 'receiver', got '%s'", info.Type)
	}

	if info.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got '%s'", info.Version)
	}

	if info.Description != "Test component" {
		t.Errorf("Expected Description 'Test component', got '%s'", info.Description)
	}

	if info.Stability != "stable" {
		t.Errorf("Expected Stability 'stable', got '%s'", info.Stability)
	}
}

// TestNewRegistry tests the NewRegistry function
func TestNewRegistry(t *testing.T) {
	r := NewRegistry()

	if r == nil {
		t.Fatal("Expected non-nil registry")
	}

	if r.components == nil {
		t.Error("Expected components map to be initialized")
	}

	if len(r.components) != 0 {
		t.Errorf("Expected empty components map, got %d entries", len(r.components))
	}
}

// TestRegister tests the Register function
func TestRegister(t *testing.T) {
	r := NewRegistry()

	factory := newMockComponent("test", TypeReceiver)
	err := r.Register("test", factory)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !r.Has("test") {
		t.Error("Expected component to be registered")
	}
}

// TestRegisterDuplicate tests registering duplicate components
func TestRegisterDuplicate(t *testing.T) {
	r := NewRegistry()

	factory := newMockComponent("test", TypeReceiver)
	_ = r.Register("test", factory)

	err := r.Register("test", factory)
	if err == nil {
		t.Error("Expected error for duplicate registration")
	}
}

// TestUnregister tests the Unregister function
func TestUnregister(t *testing.T) {
	r := NewRegistry()

	factory := newMockComponent("test", TypeReceiver)
	_ = r.Register("test", factory)

	r.Unregister("test")

	if r.Has("test") {
		t.Error("Expected component to be unregistered")
	}
}

// TestGet tests the Get function
func TestGet(t *testing.T) {
	r := NewRegistry()

	factory := newMockComponent("test", TypeReceiver)
	_ = r.Register("test", factory)

	got, err := r.Get("test")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if got == nil {
		t.Error("Expected non-nil factory")
	}
}

// TestGetNotFound tests Get for non-existent component
func TestGetNotFound(t *testing.T) {
	r := NewRegistry()

	_, err := r.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent component")
	}
}

// TestCreate tests the Create function
func TestCreate(t *testing.T) {
	r := NewRegistry()

	factory := newMockComponent("test", TypeReceiver)
	_ = r.Register("test", factory)

	component, err := r.Create("test")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if component == nil {
		t.Error("Expected non-nil component")
	}

	info := component.Info()
	if info.Name != "test" {
		t.Errorf("Expected Name 'test', got '%s'", info.Name)
	}
}

// TestCreateNotFound tests Create for non-existent component
func TestCreateNotFound(t *testing.T) {
	r := NewRegistry()

	_, err := r.Create("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent component")
	}
}

// TestList tests the List function
func TestList(t *testing.T) {
	r := NewRegistry()

	_ = r.Register("comp1", newMockComponent("comp1", TypeReceiver))
	_ = r.Register("comp2", newMockComponent("comp2", TypeProcessor))
	_ = r.Register("comp3", newMockComponent("comp3", TypeExporter))

	names := r.List()

	if len(names) != 3 {
		t.Errorf("Expected 3 components, got %d", len(names))
	}

	// Check that all names are present
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	if !nameMap["comp1"] {
		t.Error("Expected 'comp1' in list")
	}
	if !nameMap["comp2"] {
		t.Error("Expected 'comp2' in list")
	}
	if !nameMap["comp3"] {
		t.Error("Expected 'comp3' in list")
	}
}

// TestListByType tests the ListByType function
func TestListByType(t *testing.T) {
	r := NewRegistry()

	_ = r.Register("recv1", newMockComponent("recv1", TypeReceiver))
	_ = r.Register("recv2", newMockComponent("recv2", TypeReceiver))
	_ = r.Register("proc1", newMockComponent("proc1", TypeProcessor))
	_ = r.Register("exp1", newMockComponent("exp1", TypeExporter))

	receivers := r.ListByType(TypeReceiver)
	if len(receivers) != 2 {
		t.Errorf("Expected 2 receivers, got %d", len(receivers))
	}

	processors := r.ListByType(TypeProcessor)
	if len(processors) != 1 {
		t.Errorf("Expected 1 processor, got %d", len(processors))
	}

	exporters := r.ListByType(TypeExporter)
	if len(exporters) != 1 {
		t.Errorf("Expected 1 exporter, got %d", len(exporters))
	}

	extensions := r.ListByType(TypeExtension)
	if len(extensions) != 0 {
		t.Errorf("Expected 0 extensions, got %d", len(extensions))
	}
}

// TestHas tests the Has function
func TestHas(t *testing.T) {
	r := NewRegistry()

	_ = r.Register("test", newMockComponent("test", TypeReceiver))

	if !r.Has("test") {
		t.Error("Expected Has to return true for registered component")
	}

	if r.Has("nonexistent") {
		t.Error("Expected Has to return false for non-existent component")
	}
}

// TestCount tests the Count function
func TestCount(t *testing.T) {
	r := NewRegistry()

	if r.Count() != 0 {
		t.Errorf("Expected count 0, got %d", r.Count())
	}

	_ = r.Register("comp1", newMockComponent("comp1", TypeReceiver))
	if r.Count() != 1 {
		t.Errorf("Expected count 1, got %d", r.Count())
	}

	_ = r.Register("comp2", newMockComponent("comp2", TypeProcessor))
	if r.Count() != 2 {
		t.Errorf("Expected count 2, got %d", r.Count())
	}

	r.Unregister("comp1")
	if r.Count() != 1 {
		t.Errorf("Expected count 1 after unregister, got %d", r.Count())
	}
}

// TestCountByType tests the CountByType function
func TestCountByType(t *testing.T) {
	r := NewRegistry()

	_ = r.Register("recv1", newMockComponent("recv1", TypeReceiver))
	_ = r.Register("recv2", newMockComponent("recv2", TypeReceiver))
	_ = r.Register("proc1", newMockComponent("proc1", TypeProcessor))

	if r.CountByType(TypeReceiver) != 2 {
		t.Errorf("Expected 2 receivers, got %d", r.CountByType(TypeReceiver))
	}

	if r.CountByType(TypeProcessor) != 1 {
		t.Errorf("Expected 1 processor, got %d", r.CountByType(TypeProcessor))
	}

	if r.CountByType(TypeExporter) != 0 {
		t.Errorf("Expected 0 exporters, got %d", r.CountByType(TypeExporter))
	}
}

// TestGlobalRegistries tests global registry instances
func TestGlobalRegistries(t *testing.T) {
	if Receivers == nil {
		t.Error("Expected Receivers registry to be initialized")
	}

	if Processors == nil {
		t.Error("Expected Processors registry to be initialized")
	}

	if Exporters == nil {
		t.Error("Expected Exporters registry to be initialized")
	}

	if Extensions == nil {
		t.Error("Expected Extensions registry to be initialized")
	}

	if Connectors == nil {
		t.Error("Expected Connectors registry to be initialized")
	}
}

// TestRegisterReceiver tests the RegisterReceiver helper
func TestRegisterReceiver(t *testing.T) {
	// Save original and restore after test
	original := Receivers
	Receivers = NewRegistry()
	defer func() { Receivers = original }()

	err := RegisterReceiver("test-recv", newMockComponent("test-recv", TypeReceiver))
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !Receivers.Has("test-recv") {
		t.Error("Expected receiver to be registered")
	}
}

// TestRegisterProcessor tests the RegisterProcessor helper
func TestRegisterProcessor(t *testing.T) {
	original := Processors
	Processors = NewRegistry()
	defer func() { Processors = original }()

	err := RegisterProcessor("test-proc", newMockComponent("test-proc", TypeProcessor))
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !Processors.Has("test-proc") {
		t.Error("Expected processor to be registered")
	}
}

// TestRegisterExporter tests the RegisterExporter helper
func TestRegisterExporter(t *testing.T) {
	original := Exporters
	Exporters = NewRegistry()
	defer func() { Exporters = original }()

	err := RegisterExporter("test-exp", newMockComponent("test-exp", TypeExporter))
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !Exporters.Has("test-exp") {
		t.Error("Expected exporter to be registered")
	}
}

// TestRegisterExtension tests the RegisterExtension helper
func TestRegisterExtension(t *testing.T) {
	original := Extensions
	Extensions = NewRegistry()
	defer func() { Extensions = original }()

	err := RegisterExtension("test-ext", newMockComponent("test-ext", TypeExtension))
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !Extensions.Has("test-ext") {
		t.Error("Expected extension to be registered")
	}
}

// TestRegisterConnector tests the RegisterConnector helper
func TestRegisterConnector(t *testing.T) {
	original := Connectors
	Connectors = NewRegistry()
	defer func() { Connectors = original }()

	err := RegisterConnector("test-conn", newMockComponent("test-conn", TypeConnector))
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !Connectors.Has("test-conn") {
		t.Error("Expected connector to be registered")
	}
}

// TestSummary tests the Summary function
func TestSummary(t *testing.T) {
	// Save originals and restore after test
	origReceivers := Receivers
	origProcessors := Processors
	origExporters := Exporters
	origExtensions := Extensions
	origConnectors := Connectors

	Receivers = NewRegistry()
	Processors = NewRegistry()
	Exporters = NewRegistry()
	Extensions = NewRegistry()
	Connectors = NewRegistry()

	defer func() {
		Receivers = origReceivers
		Processors = origProcessors
		Exporters = origExporters
		Extensions = origExtensions
		Connectors = origConnectors
	}()

	_ = RegisterReceiver("recv1", newMockComponent("recv1", TypeReceiver))
	_ = RegisterReceiver("recv2", newMockComponent("recv2", TypeReceiver))
	_ = RegisterProcessor("proc1", newMockComponent("proc1", TypeProcessor))
	_ = RegisterExporter("exp1", newMockComponent("exp1", TypeExporter))

	summary := Summary()

	if summary[TypeReceiver] != 2 {
		t.Errorf("Expected 2 receivers, got %d", summary[TypeReceiver])
	}

	if summary[TypeProcessor] != 1 {
		t.Errorf("Expected 1 processor, got %d", summary[TypeProcessor])
	}

	if summary[TypeExporter] != 1 {
		t.Errorf("Expected 1 exporter, got %d", summary[TypeExporter])
	}

	if summary[TypeExtension] != 0 {
		t.Errorf("Expected 0 extensions, got %d", summary[TypeExtension])
	}

	if summary[TypeConnector] != 0 {
		t.Errorf("Expected 0 connectors, got %d", summary[TypeConnector])
	}
}

// TestConcurrentAccess tests concurrent registry access
func TestConcurrentAccess(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := string(rune('a' + i%26))
			_ = r.Register(name, newMockComponent(name, TypeReceiver))
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = r.List()
			_ = r.Count()
		}()
	}

	wg.Wait()
}

// TestMockComponentInterface tests mock component implements interface
func TestMockComponentInterface(t *testing.T) {
	factory := newMockComponent("test", TypeReceiver)
	component := factory()

	info := component.Info()
	if info.Name != "test" {
		t.Errorf("Expected Name 'test', got '%s'", info.Name)
	}

	if info.Type != TypeReceiver {
		t.Errorf("Expected Type 'receiver', got '%s'", info.Type)
	}

	err := component.Validate(nil)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// Benchmark tests
func BenchmarkRegister(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := NewRegistry() // Reset to avoid duplicate errors
		_ = r.Register("test", newMockComponent("test", TypeReceiver))
	}
}

func BenchmarkGet(b *testing.B) {
	r := NewRegistry()
	_ = r.Register("test", newMockComponent("test", TypeReceiver))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Get("test")
	}
}

func BenchmarkList(b *testing.B) {
	r := NewRegistry()
	for i := 0; i < 100; i++ {
		name := string(rune('a' + i%26))
		_ = r.Register(name, newMockComponent(name, TypeReceiver))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.List()
	}
}

func BenchmarkHas(b *testing.B) {
	r := NewRegistry()
	_ = r.Register("test", newMockComponent("test", TypeReceiver))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Has("test")
	}
}

func BenchmarkCount(b *testing.B) {
	r := NewRegistry()
	for i := 0; i < 100; i++ {
		name := string(rune('a' + i%26))
		_ = r.Register(name, newMockComponent(name, TypeReceiver))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Count()
	}
}
