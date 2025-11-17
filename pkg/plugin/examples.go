package plugin

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

// PrometheusPlugin is an example metrics plugin that exports Prometheus metrics
type PrometheusPlugin struct {
	*BasePlugin
	server *http.Server
	port   int
	mu     sync.Mutex
	log    *logger.Logger
}

// NewPrometheusPlugin creates a new Prometheus exporter plugin
func NewPrometheusPlugin() *PrometheusPlugin {
	return &PrometheusPlugin{
		BasePlugin: NewBasePlugin("prometheus-exporter", MetricsPlugin, "1.0.0"),
		port:       9090,
		log:        logger.New("prometheus-plugin"),
	}
}

// Init initializes the Prometheus plugin
func (p *PrometheusPlugin) Init(config map[string]interface{}) error {
	p.BasePlugin.Init(config)

	// Get port from config if provided
	if port, ok := config["port"].(int); ok {
		p.port = port
	} else if portFloat, ok := config["port"].(float64); ok {
		p.port = int(portFloat)
	}

	p.log.Infof("Initialized Prometheus plugin on port %d", p.port)
	return nil
}

// Start starts the Prometheus HTTP server
func (p *PrometheusPlugin) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.server != nil {
		return fmt.Errorf("server already running")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", p.metricsHandler)
	mux.HandleFunc("/health", p.healthHandler)

	p.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", p.port),
		Handler: mux,
	}

	go func() {
		p.log.Infof("Starting Prometheus exporter on :%d", p.port)
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			p.log.WithError(err).Error("Server error")
		}
	}()

	p.log.Info("Prometheus plugin started")
	return nil
}

// Stop stops the Prometheus HTTP server
func (p *PrometheusPlugin) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.server == nil {
		return nil
	}

	p.log.Info("Stopping Prometheus plugin")

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := p.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	p.server = nil
	p.log.Info("Prometheus plugin stopped")
	return nil
}

// Health checks the health of the plugin
func (p *PrometheusPlugin) Health(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.server == nil {
		return fmt.Errorf("server not running")
	}

	return nil
}

func (p *PrometheusPlugin) metricsHandler(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would export actual metrics
	metrics := `# HELP containr_containers_total Total number of containers
# TYPE containr_containers_total gauge
containr_containers_total 0

# HELP containr_images_total Total number of images
# TYPE containr_images_total gauge
containr_images_total 0

# HELP containr_plugin_calls_total Total number of plugin calls
# TYPE containr_plugin_calls_total counter
containr_plugin_calls_total{plugin="prometheus-exporter"} 1
`
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(metrics))
}

func (p *PrometheusPlugin) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// LoggerPlugin is an example logging plugin
type LoggerPlugin struct {
	*BasePlugin
	output string
	log    *logger.Logger
}

// NewLoggerPlugin creates a new custom logger plugin
func NewLoggerPlugin() *LoggerPlugin {
	return &LoggerPlugin{
		BasePlugin: NewBasePlugin("custom-logger", LoggingPlugin, "1.0.0"),
		output:     "/var/log/containr/custom.log",
		log:        logger.New("logger-plugin"),
	}
}

// Init initializes the logger plugin
func (l *LoggerPlugin) Init(config map[string]interface{}) error {
	l.BasePlugin.Init(config)

	// Get output path from config
	if output, ok := config["output"].(string); ok {
		l.output = output
	}

	l.log.Infof("Initialized logger plugin with output: %s", l.output)
	return nil
}

// Start starts the logger plugin
func (l *LoggerPlugin) Start(ctx context.Context) error {
	l.log.Infof("Starting logger plugin, logging to: %s", l.output)
	return nil
}

// Stop stops the logger plugin
func (l *LoggerPlugin) Stop(ctx context.Context) error {
	l.log.Info("Stopping logger plugin")
	return nil
}

// Health checks the health of the logger plugin
func (l *LoggerPlugin) Health(ctx context.Context) error {
	// Check if output file is writable
	return nil
}

// CustomNetworkPlugin is an example network plugin
type CustomNetworkPlugin struct {
	*BasePlugin
	networkMode string
	log         *logger.Logger
}

// NewNetworkPlugin creates a new network plugin
func NewNetworkPlugin() *CustomNetworkPlugin {
	return &CustomNetworkPlugin{
		BasePlugin:  NewBasePlugin("custom-network", NetworkPlugin, "1.0.0"),
		networkMode: "bridge",
		log:         logger.New("network-plugin"),
	}
}

// Init initializes the network plugin
func (n *CustomNetworkPlugin) Init(config map[string]interface{}) error {
	n.BasePlugin.Init(config)

	if mode, ok := config["mode"].(string); ok {
		n.networkMode = mode
	}

	n.log.Infof("Initialized network plugin with mode: %s", n.networkMode)
	return nil
}

// Start starts the network plugin
func (n *CustomNetworkPlugin) Start(ctx context.Context) error {
	n.log.Infof("Starting network plugin in %s mode", n.networkMode)
	return nil
}

// Stop stops the network plugin
func (n *CustomNetworkPlugin) Stop(ctx context.Context) error {
	n.log.Info("Stopping network plugin")
	return nil
}

// Health checks the health of the network plugin
func (n *CustomNetworkPlugin) Health(ctx context.Context) error {
	return nil
}

// RegisterBuiltinPlugins registers all built-in example plugins
func RegisterBuiltinPlugins(pm *PluginManager) error {
	plugins := []Plugin{
		NewPrometheusPlugin(),
		NewLoggerPlugin(),
		NewNetworkPlugin(),
	}

	for _, plugin := range plugins {
		if err := pm.Register(plugin); err != nil {
			return fmt.Errorf("failed to register plugin %s: %w", plugin.Name(), err)
		}
	}

	return nil
}
