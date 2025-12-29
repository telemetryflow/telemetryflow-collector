// Package plugin_test provides unit tests for the plugin registry.
//
// TelemetryFlow Collector - Community Enterprise Observability Platform (CEOP)
// Copyright (c) 2024-2026 DevOpsCorner Indonesia. All rights reserved.
package plugin_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/telemetryflow/telemetryflow-collector/pkg/plugin"
)

// mockComponent is a mock implementation of Component interface
type mockComponent struct {
	info plugin.ComponentInfo
}

func (m *mockComponent) Info() plugin.ComponentInfo {
	return m.info
}

func (m *mockComponent) Validate(_ map[string]interface{}) error {
	return nil
}

// newMockComponent creates a mock component factory
func newMockComponent(name string, t plugin.ComponentType) plugin.ComponentFactory {
	return func() plugin.Component {
		return &mockComponent{
			info: plugin.ComponentInfo{
				Name:        name,
				Type:        t,
				Version:     "1.0.0",
				Description: "Mock component for testing",
				Stability:   "stable",
			},
		}
	}
}

func TestComponentTypeConstants(t *testing.T) {
	assert.Equal(t, plugin.ComponentType("receiver"), plugin.TypeReceiver)
	assert.Equal(t, plugin.ComponentType("processor"), plugin.TypeProcessor)
	assert.Equal(t, plugin.ComponentType("exporter"), plugin.TypeExporter)
	assert.Equal(t, plugin.ComponentType("extension"), plugin.TypeExtension)
	assert.Equal(t, plugin.ComponentType("connector"), plugin.TypeConnector)
}

func TestComponentInfo(t *testing.T) {
	info := plugin.ComponentInfo{
		Name:        "test-component",
		Type:        plugin.TypeReceiver,
		Version:     "1.0.0",
		Description: "Test component",
		Stability:   "stable",
	}

	assert.Equal(t, "test-component", info.Name)
	assert.Equal(t, plugin.TypeReceiver, info.Type)
	assert.Equal(t, "1.0.0", info.Version)
	assert.Equal(t, "Test component", info.Description)
	assert.Equal(t, "stable", info.Stability)
}

func TestNewRegistry(t *testing.T) {
	r := plugin.NewRegistry()

	require.NotNil(t, r)
	assert.Equal(t, 0, r.Count())
}

func TestRegister(t *testing.T) {
	r := plugin.NewRegistry()

	factory := newMockComponent("test", plugin.TypeReceiver)
	err := r.Register("test", factory)

	require.NoError(t, err)
	assert.True(t, r.Has("test"))
}

func TestRegisterDuplicate(t *testing.T) {
	r := plugin.NewRegistry()

	factory := newMockComponent("test", plugin.TypeReceiver)
	_ = r.Register("test", factory)

	err := r.Register("test", factory)
	assert.Error(t, err)
}

func TestUnregister(t *testing.T) {
	r := plugin.NewRegistry()

	factory := newMockComponent("test", plugin.TypeReceiver)
	_ = r.Register("test", factory)

	r.Unregister("test")

	assert.False(t, r.Has("test"))
}

func TestGet(t *testing.T) {
	r := plugin.NewRegistry()

	factory := newMockComponent("test", plugin.TypeReceiver)
	_ = r.Register("test", factory)

	got, err := r.Get("test")

	require.NoError(t, err)
	require.NotNil(t, got)
}

func TestGetNotFound(t *testing.T) {
	r := plugin.NewRegistry()

	_, err := r.Get("nonexistent")
	assert.Error(t, err)
}

func TestCreate(t *testing.T) {
	r := plugin.NewRegistry()

	factory := newMockComponent("test", plugin.TypeReceiver)
	_ = r.Register("test", factory)

	component, err := r.Create("test")

	require.NoError(t, err)
	require.NotNil(t, component)

	info := component.Info()
	assert.Equal(t, "test", info.Name)
}

func TestCreateNotFound(t *testing.T) {
	r := plugin.NewRegistry()

	_, err := r.Create("nonexistent")
	assert.Error(t, err)
}

func TestList(t *testing.T) {
	r := plugin.NewRegistry()

	_ = r.Register("comp1", newMockComponent("comp1", plugin.TypeReceiver))
	_ = r.Register("comp2", newMockComponent("comp2", plugin.TypeProcessor))
	_ = r.Register("comp3", newMockComponent("comp3", plugin.TypeExporter))

	names := r.List()

	assert.Len(t, names, 3)
	assert.Contains(t, names, "comp1")
	assert.Contains(t, names, "comp2")
	assert.Contains(t, names, "comp3")
}

func TestListByType(t *testing.T) {
	r := plugin.NewRegistry()

	_ = r.Register("recv1", newMockComponent("recv1", plugin.TypeReceiver))
	_ = r.Register("recv2", newMockComponent("recv2", plugin.TypeReceiver))
	_ = r.Register("proc1", newMockComponent("proc1", plugin.TypeProcessor))
	_ = r.Register("exp1", newMockComponent("exp1", plugin.TypeExporter))

	receivers := r.ListByType(plugin.TypeReceiver)
	assert.Len(t, receivers, 2)

	processors := r.ListByType(plugin.TypeProcessor)
	assert.Len(t, processors, 1)

	exporters := r.ListByType(plugin.TypeExporter)
	assert.Len(t, exporters, 1)

	extensions := r.ListByType(plugin.TypeExtension)
	assert.Len(t, extensions, 0)
}

func TestHas(t *testing.T) {
	r := plugin.NewRegistry()

	_ = r.Register("test", newMockComponent("test", plugin.TypeReceiver))

	assert.True(t, r.Has("test"))
	assert.False(t, r.Has("nonexistent"))
}

func TestCount(t *testing.T) {
	r := plugin.NewRegistry()

	assert.Equal(t, 0, r.Count())

	_ = r.Register("comp1", newMockComponent("comp1", plugin.TypeReceiver))
	assert.Equal(t, 1, r.Count())

	_ = r.Register("comp2", newMockComponent("comp2", plugin.TypeProcessor))
	assert.Equal(t, 2, r.Count())

	r.Unregister("comp1")
	assert.Equal(t, 1, r.Count())
}

func TestCountByType(t *testing.T) {
	r := plugin.NewRegistry()

	_ = r.Register("recv1", newMockComponent("recv1", plugin.TypeReceiver))
	_ = r.Register("recv2", newMockComponent("recv2", plugin.TypeReceiver))
	_ = r.Register("proc1", newMockComponent("proc1", plugin.TypeProcessor))

	assert.Equal(t, 2, r.CountByType(plugin.TypeReceiver))
	assert.Equal(t, 1, r.CountByType(plugin.TypeProcessor))
	assert.Equal(t, 0, r.CountByType(plugin.TypeExporter))
}

func TestGlobalRegistries(t *testing.T) {
	assert.NotNil(t, plugin.Receivers)
	assert.NotNil(t, plugin.Processors)
	assert.NotNil(t, plugin.Exporters)
	assert.NotNil(t, plugin.Extensions)
	assert.NotNil(t, plugin.Connectors)
}

func TestRegisterReceiver(t *testing.T) {
	original := plugin.Receivers
	plugin.Receivers = plugin.NewRegistry()
	defer func() { plugin.Receivers = original }()

	err := plugin.RegisterReceiver("test-recv", newMockComponent("test-recv", plugin.TypeReceiver))

	require.NoError(t, err)
	assert.True(t, plugin.Receivers.Has("test-recv"))
}

func TestRegisterProcessor(t *testing.T) {
	original := plugin.Processors
	plugin.Processors = plugin.NewRegistry()
	defer func() { plugin.Processors = original }()

	err := plugin.RegisterProcessor("test-proc", newMockComponent("test-proc", plugin.TypeProcessor))

	require.NoError(t, err)
	assert.True(t, plugin.Processors.Has("test-proc"))
}

func TestRegisterExporter(t *testing.T) {
	original := plugin.Exporters
	plugin.Exporters = plugin.NewRegistry()
	defer func() { plugin.Exporters = original }()

	err := plugin.RegisterExporter("test-exp", newMockComponent("test-exp", plugin.TypeExporter))

	require.NoError(t, err)
	assert.True(t, plugin.Exporters.Has("test-exp"))
}

func TestRegisterExtension(t *testing.T) {
	original := plugin.Extensions
	plugin.Extensions = plugin.NewRegistry()
	defer func() { plugin.Extensions = original }()

	err := plugin.RegisterExtension("test-ext", newMockComponent("test-ext", plugin.TypeExtension))

	require.NoError(t, err)
	assert.True(t, plugin.Extensions.Has("test-ext"))
}

func TestRegisterConnector(t *testing.T) {
	original := plugin.Connectors
	plugin.Connectors = plugin.NewRegistry()
	defer func() { plugin.Connectors = original }()

	err := plugin.RegisterConnector("test-conn", newMockComponent("test-conn", plugin.TypeConnector))

	require.NoError(t, err)
	assert.True(t, plugin.Connectors.Has("test-conn"))
}

func TestSummary(t *testing.T) {
	origReceivers := plugin.Receivers
	origProcessors := plugin.Processors
	origExporters := plugin.Exporters
	origExtensions := plugin.Extensions
	origConnectors := plugin.Connectors

	plugin.Receivers = plugin.NewRegistry()
	plugin.Processors = plugin.NewRegistry()
	plugin.Exporters = plugin.NewRegistry()
	plugin.Extensions = plugin.NewRegistry()
	plugin.Connectors = plugin.NewRegistry()

	defer func() {
		plugin.Receivers = origReceivers
		plugin.Processors = origProcessors
		plugin.Exporters = origExporters
		plugin.Extensions = origExtensions
		plugin.Connectors = origConnectors
	}()

	_ = plugin.RegisterReceiver("recv1", newMockComponent("recv1", plugin.TypeReceiver))
	_ = plugin.RegisterReceiver("recv2", newMockComponent("recv2", plugin.TypeReceiver))
	_ = plugin.RegisterProcessor("proc1", newMockComponent("proc1", plugin.TypeProcessor))
	_ = plugin.RegisterExporter("exp1", newMockComponent("exp1", plugin.TypeExporter))

	summary := plugin.Summary()

	assert.Equal(t, 2, summary[plugin.TypeReceiver])
	assert.Equal(t, 1, summary[plugin.TypeProcessor])
	assert.Equal(t, 1, summary[plugin.TypeExporter])
	assert.Equal(t, 0, summary[plugin.TypeExtension])
	assert.Equal(t, 0, summary[plugin.TypeConnector])
}

func TestConcurrentAccess(t *testing.T) {
	r := plugin.NewRegistry()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := string(rune('a' + i%26))
			_ = r.Register(name, newMockComponent(name, plugin.TypeReceiver))
		}(i)
	}

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

func TestMockComponentInterface(t *testing.T) {
	factory := newMockComponent("test", plugin.TypeReceiver)
	component := factory()

	info := component.Info()
	assert.Equal(t, "test", info.Name)
	assert.Equal(t, plugin.TypeReceiver, info.Type)

	err := component.Validate(nil)
	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkRegister(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := plugin.NewRegistry()
		_ = r.Register("test", newMockComponent("test", plugin.TypeReceiver))
	}
}

func BenchmarkGet(b *testing.B) {
	r := plugin.NewRegistry()
	_ = r.Register("test", newMockComponent("test", plugin.TypeReceiver))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Get("test")
	}
}

func BenchmarkList(b *testing.B) {
	r := plugin.NewRegistry()
	for i := 0; i < 100; i++ {
		name := string(rune('a' + i%26))
		_ = r.Register(name, newMockComponent(name, plugin.TypeReceiver))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.List()
	}
}

func BenchmarkHas(b *testing.B) {
	r := plugin.NewRegistry()
	_ = r.Register("test", newMockComponent("test", plugin.TypeReceiver))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Has("test")
	}
}

func BenchmarkCount(b *testing.B) {
	r := plugin.NewRegistry()
	for i := 0; i < 100; i++ {
		name := string(rune('a' + i%26))
		_ = r.Register(name, newMockComponent(name, plugin.TypeReceiver))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Count()
	}
}
