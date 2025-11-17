package plugin

import (
	"context"
	"fmt"
	"sync"
)

// PluginType defines the category of a plugin
type PluginType string

const (
	// RuntimePlugin handles container lifecycle hooks
	RuntimePlugin PluginType = "runtime"
	// NetworkPlugin provides custom networking implementations
	NetworkPlugin PluginType = "network"
	// StoragePlugin provides custom volume drivers
	StoragePlugin PluginType = "storage"
	// LoggingPlugin provides custom log collectors
	LoggingPlugin PluginType = "logging"
	// MetricsPlugin provides custom metrics exporters
	MetricsPlugin PluginType = "metrics"
)

// Plugin represents a containr plugin
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Type returns the plugin type
	Type() PluginType

	// Version returns the plugin version
	Version() string

	// Init initializes the plugin with configuration
	Init(config map[string]interface{}) error

	// Start starts the plugin
	Start(ctx context.Context) error

	// Stop stops the plugin gracefully
	Stop(ctx context.Context) error

	// Health returns the health status of the plugin
	Health(ctx context.Context) error
}

// PluginInfo contains metadata about a plugin
type PluginInfo struct {
	Name        string                 `json:"name"`
	Type        PluginType             `json:"type"`
	Version     string                 `json:"version"`
	Enabled     bool                   `json:"enabled"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// PluginManager manages plugin lifecycle
type PluginManager struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
	enabled map[string]bool
	configs map[string]map[string]interface{}
}

// NewManager creates a new plugin manager
func NewManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
		enabled: make(map[string]bool),
		configs: make(map[string]map[string]interface{}),
	}
}

// Register registers a plugin with the manager
func (pm *PluginManager) Register(plugin Plugin) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	name := plugin.Name()
	if _, exists := pm.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	pm.plugins[name] = plugin
	pm.enabled[name] = false
	return nil
}

// Unregister removes a plugin from the manager
func (pm *PluginManager) Unregister(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Stop plugin if running
	if pm.enabled[name] {
		ctx := context.Background()
		if err := plugin.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop plugin %s: %w", name, err)
		}
	}

	delete(pm.plugins, name)
	delete(pm.enabled, name)
	delete(pm.configs, name)
	return nil
}

// Get retrieves a plugin by name
func (pm *PluginManager) Get(name string) (Plugin, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return plugin, nil
}

// List returns all registered plugins
func (pm *PluginManager) List() []PluginInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	infos := make([]PluginInfo, 0, len(pm.plugins))
	for name, plugin := range pm.plugins {
		info := PluginInfo{
			Name:    name,
			Type:    plugin.Type(),
			Version: plugin.Version(),
			Enabled: pm.enabled[name],
			Config:  pm.configs[name],
		}
		infos = append(infos, info)
	}

	return infos
}

// Enable enables a plugin and starts it
func (pm *PluginManager) Enable(ctx context.Context, name string, config map[string]interface{}) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if pm.enabled[name] {
		return fmt.Errorf("plugin %s is already enabled", name)
	}

	// Initialize with config
	if err := plugin.Init(config); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %w", name, err)
	}

	// Start the plugin
	if err := plugin.Start(ctx); err != nil {
		return fmt.Errorf("failed to start plugin %s: %w", name, err)
	}

	pm.enabled[name] = true
	pm.configs[name] = config
	return nil
}

// Disable disables a plugin and stops it
func (pm *PluginManager) Disable(ctx context.Context, name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if !pm.enabled[name] {
		return fmt.Errorf("plugin %s is not enabled", name)
	}

	// Stop the plugin
	if err := plugin.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop plugin %s: %w", name, err)
	}

	pm.enabled[name] = false
	return nil
}

// IsEnabled checks if a plugin is enabled
func (pm *PluginManager) IsEnabled(name string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.enabled[name]
}

// GetByType returns all plugins of a specific type
func (pm *PluginManager) GetByType(pluginType PluginType) []Plugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugins := make([]Plugin, 0)
	for _, plugin := range pm.plugins {
		if plugin.Type() == pluginType {
			plugins = append(plugins, plugin)
		}
	}

	return plugins
}

// HealthCheck checks the health of all enabled plugins
func (pm *PluginManager) HealthCheck(ctx context.Context) map[string]error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	results := make(map[string]error)
	for name, plugin := range pm.plugins {
		if pm.enabled[name] {
			results[name] = plugin.Health(ctx)
		}
	}

	return results
}

// BasePlugin provides a base implementation for plugins
type BasePlugin struct {
	name    string
	ptype   PluginType
	version string
	config  map[string]interface{}
}

// NewBasePlugin creates a new base plugin
func NewBasePlugin(name string, ptype PluginType, version string) *BasePlugin {
	return &BasePlugin{
		name:    name,
		ptype:   ptype,
		version: version,
		config:  make(map[string]interface{}),
	}
}

// Name returns the plugin name
func (bp *BasePlugin) Name() string {
	return bp.name
}

// Type returns the plugin type
func (bp *BasePlugin) Type() PluginType {
	return bp.ptype
}

// Version returns the plugin version
func (bp *BasePlugin) Version() string {
	return bp.version
}

// Init initializes the plugin (default implementation)
func (bp *BasePlugin) Init(config map[string]interface{}) error {
	bp.config = config
	return nil
}

// Start starts the plugin (default implementation)
func (bp *BasePlugin) Start(ctx context.Context) error {
	return nil
}

// Stop stops the plugin (default implementation)
func (bp *BasePlugin) Stop(ctx context.Context) error {
	return nil
}

// Health checks plugin health (default implementation)
func (bp *BasePlugin) Health(ctx context.Context) error {
	return nil
}

// GetConfig returns the plugin configuration
func (bp *BasePlugin) GetConfig() map[string]interface{} {
	return bp.config
}

// GetConfigValue retrieves a specific configuration value
func (bp *BasePlugin) GetConfigValue(key string) (interface{}, bool) {
	val, ok := bp.config[key]
	return val, ok
}
