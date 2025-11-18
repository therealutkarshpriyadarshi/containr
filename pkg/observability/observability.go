package observability

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Manager manages observability for containr
type Manager struct {
	config         *Config
	tracerProvider *sdktrace.TracerProvider
	metricsManager *MetricsManager
	logger         *Logger
	exporters      *ExporterManager
	mu             sync.RWMutex
	shutdownFuncs  []func(context.Context) error
}

// Config holds observability configuration
type Config struct {
	ServiceName    string          `yaml:"service_name"`
	ServiceVersion string          `yaml:"service_version"`
	Environment    string          `yaml:"environment"`
	Enabled        bool            `yaml:"enabled"`
	Tracing        TracingConfig   `yaml:"tracing"`
	Metrics        MetricsConfig   `yaml:"metrics"`
	Logging        LoggingConfig   `yaml:"logging"`
	Exporters      ExporterConfig  `yaml:"exporters"`
}

// TracingConfig holds tracing configuration
type TracingConfig struct {
	Enabled      bool    `yaml:"enabled"`
	SamplingRate float64 `yaml:"sampling_rate"`
	MaxSpansPerTrace int `yaml:"max_spans_per_trace"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled         bool   `yaml:"enabled"`
	Port            int    `yaml:"port"`
	Path            string `yaml:"path"`
	CollectInterval int    `yaml:"collect_interval_seconds"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Enabled bool   `yaml:"enabled"`
	Level   string `yaml:"level"`
	Format  string `yaml:"format"` // json or text
	Output  string `yaml:"output"` // stdout, stderr, or file path
}

// ExporterConfig holds exporter configuration
type ExporterConfig struct {
	Prometheus PrometheusConfig `yaml:"prometheus"`
	Jaeger     JaegerConfig     `yaml:"jaeger"`
	OTLP       OTLPConfig       `yaml:"otlp"`
}

// PrometheusConfig holds Prometheus exporter configuration
type PrometheusConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
	Port     int    `yaml:"port"`
}

// JaegerConfig holds Jaeger exporter configuration
type JaegerConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
	AgentHost string `yaml:"agent_host"`
	AgentPort int    `yaml:"agent_port"`
}

// OTLPConfig holds OTLP exporter configuration
type OTLPConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
	Insecure bool   `yaml:"insecure"`
}

// NewManager creates a new observability manager
func NewManager(config *Config) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	mgr := &Manager{
		config:        config,
		shutdownFuncs: make([]func(context.Context) error, 0),
	}

	// Initialize components if enabled
	if !config.Enabled {
		return mgr, nil
	}

	// Initialize exporters
	if err := mgr.initExporters(); err != nil {
		return nil, fmt.Errorf("failed to initialize exporters: %w", err)
	}

	// Initialize tracing
	if config.Tracing.Enabled {
		if err := mgr.initTracing(); err != nil {
			return nil, fmt.Errorf("failed to initialize tracing: %w", err)
		}
	}

	// Initialize metrics
	if config.Metrics.Enabled {
		if err := mgr.initMetrics(); err != nil {
			return nil, fmt.Errorf("failed to initialize metrics: %w", err)
		}
	}

	// Initialize logging
	if config.Logging.Enabled {
		if err := mgr.initLogging(); err != nil {
			return nil, fmt.Errorf("failed to initialize logging: %w", err)
		}
	}

	return mgr, nil
}

// initExporters initializes all configured exporters
func (m *Manager) initExporters() error {
	var err error
	m.exporters, err = NewExporterManager(m.config.Exporters)
	if err != nil {
		return err
	}

	return nil
}

// initTracing initializes the tracing provider
func (m *Manager) initTracing() error {
	tp, err := NewTracerProvider(
		m.config.ServiceName,
		m.config.ServiceVersion,
		m.config.Environment,
		m.config.Tracing,
		m.exporters,
	)
	if err != nil {
		return err
	}

	m.tracerProvider = tp

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator for context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Register shutdown function
	m.shutdownFuncs = append(m.shutdownFuncs, func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	})

	return nil
}

// initMetrics initializes the metrics manager
func (m *Manager) initMetrics() error {
	metricsManager, err := NewMetricsManager(
		m.config.ServiceName,
		m.config.Metrics,
		m.exporters,
	)
	if err != nil {
		return err
	}

	m.metricsManager = metricsManager

	// Register shutdown function
	m.shutdownFuncs = append(m.shutdownFuncs, func(ctx context.Context) error {
		return metricsManager.Shutdown(ctx)
	})

	return nil
}

// initLogging initializes the logger
func (m *Manager) initLogging() error {
	logger, err := NewLogger(m.config.Logging)
	if err != nil {
		return err
	}

	m.logger = logger

	// Register shutdown function
	m.shutdownFuncs = append(m.shutdownFuncs, func(ctx context.Context) error {
		return logger.Close()
	})

	return nil
}

// GetTracer returns a tracer for the given name
func (m *Manager) GetTracer(name string) Tracer {
	if m.tracerProvider == nil {
		return &NoopTracer{}
	}
	return NewTracer(m.tracerProvider.Tracer(name))
}

// GetMetrics returns the metrics manager
func (m *Manager) GetMetrics() *MetricsManager {
	if m.metricsManager == nil {
		return &MetricsManager{}
	}
	return m.metricsManager
}

// GetLogger returns the logger
func (m *Manager) GetLogger() *Logger {
	if m.logger == nil {
		// Return a default logger
		logger, _ := NewLogger(LoggingConfig{
			Enabled: true,
			Level:   "info",
			Format:  "json",
			Output:  "stdout",
		})
		return logger
	}
	return m.logger
}

// Shutdown gracefully shuts down all observability components
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errors []error

	// Execute all shutdown functions in reverse order
	for i := len(m.shutdownFuncs) - 1; i >= 0; i-- {
		if err := m.shutdownFuncs[i](ctx); err != nil {
			errors = append(errors, err)
		}
	}

	// Shutdown exporters
	if m.exporters != nil {
		if err := m.exporters.Shutdown(ctx); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	return nil
}

// DefaultConfig returns the default observability configuration
func DefaultConfig() *Config {
	return &Config{
		ServiceName:    "containr",
		ServiceVersion: "1.0.0",
		Environment:    "development",
		Enabled:        true,
		Tracing: TracingConfig{
			Enabled:          true,
			SamplingRate:     1.0,
			MaxSpansPerTrace: 1000,
		},
		Metrics: MetricsConfig{
			Enabled:         true,
			Port:            9090,
			Path:            "/metrics",
			CollectInterval: 15,
		},
		Logging: LoggingConfig{
			Enabled: true,
			Level:   "info",
			Format:  "json",
			Output:  "stdout",
		},
		Exporters: ExporterConfig{
			Prometheus: PrometheusConfig{
				Enabled:  true,
				Endpoint: "localhost:9090",
				Port:     9090,
			},
			Jaeger: JaegerConfig{
				Enabled:   false,
				Endpoint:  "http://localhost:14268/api/traces",
				AgentHost: "localhost",
				AgentPort: 6831,
			},
			OTLP: OTLPConfig{
				Enabled:  false,
				Endpoint: "localhost:4317",
				Insecure: true,
			},
		},
	}
}
