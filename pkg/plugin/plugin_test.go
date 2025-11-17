package plugin

import (
	"context"
	"testing"
	"time"
)

// MockPlugin is a test plugin implementation
type MockPlugin struct {
	*BasePlugin
	startCalled bool
	stopCalled  bool
	healthErr   error
}

func NewMockPlugin(name string, ptype PluginType) *MockPlugin {
	return &MockPlugin{
		BasePlugin: NewBasePlugin(name, ptype, "1.0.0"),
	}
}

func (mp *MockPlugin) Start(ctx context.Context) error {
	mp.startCalled = true
	return mp.BasePlugin.Start(ctx)
}

func (mp *MockPlugin) Stop(ctx context.Context) error {
	mp.stopCalled = true
	return mp.BasePlugin.Stop(ctx)
}

func (mp *MockPlugin) Health(ctx context.Context) error {
	return mp.healthErr
}

func TestPluginManager_Register(t *testing.T) {
	pm := NewManager()
	plugin := NewMockPlugin("test-plugin", MetricsPlugin)

	err := pm.Register(plugin)
	if err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	// Test duplicate registration
	err = pm.Register(plugin)
	if err == nil {
		t.Fatal("Expected error when registering duplicate plugin")
	}
}

func TestPluginManager_Get(t *testing.T) {
	pm := NewManager()
	plugin := NewMockPlugin("test-plugin", MetricsPlugin)

	pm.Register(plugin)

	retrieved, err := pm.Get("test-plugin")
	if err != nil {
		t.Fatalf("Failed to get plugin: %v", err)
	}

	if retrieved.Name() != "test-plugin" {
		t.Errorf("Expected plugin name 'test-plugin', got '%s'", retrieved.Name())
	}

	// Test getting non-existent plugin
	_, err = pm.Get("non-existent")
	if err == nil {
		t.Fatal("Expected error when getting non-existent plugin")
	}
}

func TestPluginManager_List(t *testing.T) {
	pm := NewManager()

	plugins := []Plugin{
		NewMockPlugin("plugin1", MetricsPlugin),
		NewMockPlugin("plugin2", LoggingPlugin),
		NewMockPlugin("plugin3", NetworkPlugin),
	}

	for _, p := range plugins {
		pm.Register(p)
	}

	list := pm.List()
	if len(list) != 3 {
		t.Errorf("Expected 3 plugins, got %d", len(list))
	}
}

func TestPluginManager_Enable(t *testing.T) {
	pm := NewManager()
	plugin := NewMockPlugin("test-plugin", MetricsPlugin)
	pm.Register(plugin)

	ctx := context.Background()
	config := map[string]interface{}{"key": "value"}

	err := pm.Enable(ctx, "test-plugin", config)
	if err != nil {
		t.Fatalf("Failed to enable plugin: %v", err)
	}

	if !plugin.startCalled {
		t.Error("Expected Start to be called on plugin")
	}

	if !pm.IsEnabled("test-plugin") {
		t.Error("Expected plugin to be enabled")
	}

	// Test enabling already enabled plugin
	err = pm.Enable(ctx, "test-plugin", config)
	if err == nil {
		t.Fatal("Expected error when enabling already enabled plugin")
	}
}

func TestPluginManager_Disable(t *testing.T) {
	pm := NewManager()
	plugin := NewMockPlugin("test-plugin", MetricsPlugin)
	pm.Register(plugin)

	ctx := context.Background()
	config := map[string]interface{}{}

	pm.Enable(ctx, "test-plugin", config)

	err := pm.Disable(ctx, "test-plugin")
	if err != nil {
		t.Fatalf("Failed to disable plugin: %v", err)
	}

	if !plugin.stopCalled {
		t.Error("Expected Stop to be called on plugin")
	}

	if pm.IsEnabled("test-plugin") {
		t.Error("Expected plugin to be disabled")
	}

	// Test disabling already disabled plugin
	err = pm.Disable(ctx, "test-plugin")
	if err == nil {
		t.Fatal("Expected error when disabling already disabled plugin")
	}
}

func TestPluginManager_Unregister(t *testing.T) {
	pm := NewManager()
	plugin := NewMockPlugin("test-plugin", MetricsPlugin)
	pm.Register(plugin)

	ctx := context.Background()
	pm.Enable(ctx, "test-plugin", map[string]interface{}{})

	err := pm.Unregister("test-plugin")
	if err != nil {
		t.Fatalf("Failed to unregister plugin: %v", err)
	}

	if !plugin.stopCalled {
		t.Error("Expected Stop to be called before unregistering")
	}

	_, err = pm.Get("test-plugin")
	if err == nil {
		t.Fatal("Expected error when getting unregistered plugin")
	}
}

func TestPluginManager_GetByType(t *testing.T) {
	pm := NewManager()

	pm.Register(NewMockPlugin("metrics1", MetricsPlugin))
	pm.Register(NewMockPlugin("metrics2", MetricsPlugin))
	pm.Register(NewMockPlugin("logging1", LoggingPlugin))

	metricsPlugins := pm.GetByType(MetricsPlugin)
	if len(metricsPlugins) != 2 {
		t.Errorf("Expected 2 metrics plugins, got %d", len(metricsPlugins))
	}

	loggingPlugins := pm.GetByType(LoggingPlugin)
	if len(loggingPlugins) != 1 {
		t.Errorf("Expected 1 logging plugin, got %d", len(loggingPlugins))
	}
}

func TestPluginManager_HealthCheck(t *testing.T) {
	pm := NewManager()

	plugin1 := NewMockPlugin("healthy-plugin", MetricsPlugin)
	plugin2 := NewMockPlugin("unhealthy-plugin", MetricsPlugin)
	plugin2.healthErr = &PluginError{Plugin: "unhealthy-plugin", Err: "health check failed"}

	pm.Register(plugin1)
	pm.Register(plugin2)

	ctx := context.Background()
	pm.Enable(ctx, "healthy-plugin", map[string]interface{}{})
	pm.Enable(ctx, "unhealthy-plugin", map[string]interface{}{})

	results := pm.HealthCheck(ctx)

	if results["healthy-plugin"] != nil {
		t.Errorf("Expected healthy-plugin to be healthy, got error: %v", results["healthy-plugin"])
	}

	if results["unhealthy-plugin"] == nil {
		t.Error("Expected unhealthy-plugin to have error")
	}
}

func TestBasePlugin(t *testing.T) {
	plugin := NewBasePlugin("test", MetricsPlugin, "1.0.0")

	if plugin.Name() != "test" {
		t.Errorf("Expected name 'test', got '%s'", plugin.Name())
	}

	if plugin.Type() != MetricsPlugin {
		t.Errorf("Expected type MetricsPlugin, got %v", plugin.Type())
	}

	if plugin.Version() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", plugin.Version())
	}

	// Test configuration
	config := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	err := plugin.Init(config)
	if err != nil {
		t.Fatalf("Failed to init plugin: %v", err)
	}

	retrievedConfig := plugin.GetConfig()
	if len(retrievedConfig) != 2 {
		t.Errorf("Expected 2 config items, got %d", len(retrievedConfig))
	}

	val, ok := plugin.GetConfigValue("key1")
	if !ok {
		t.Error("Expected to find config value for key1")
	}
	if val != "value1" {
		t.Errorf("Expected value 'value1', got '%v'", val)
	}

	_, ok = plugin.GetConfigValue("non-existent")
	if ok {
		t.Error("Expected not to find non-existent config value")
	}
}

func TestPluginManager_Concurrent(t *testing.T) {
	pm := NewManager()
	ctx := context.Background()

	// Test concurrent registration
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			plugin := NewMockPlugin(string(rune('A'+id)), MetricsPlugin)
			pm.Register(plugin)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	list := pm.List()
	if len(list) != 10 {
		t.Errorf("Expected 10 plugins after concurrent registration, got %d", len(list))
	}

	// Test concurrent enable/disable
	for i := 0; i < 5; i++ {
		go func(id int) {
			name := string(rune('A' + id))
			pm.Enable(ctx, name, map[string]interface{}{})
			time.Sleep(10 * time.Millisecond)
			pm.Disable(ctx, name)
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}

// PluginError represents a plugin-specific error
type PluginError struct {
	Plugin string
	Err    string
}

func (e *PluginError) Error() string {
	return e.Plugin + ": " + e.Err
}
