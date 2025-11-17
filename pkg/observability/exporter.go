package observability

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials/insecure"
)

// ExporterManager manages all exporters
type ExporterManager struct {
	config         ExporterConfig
	traceExporters []sdktrace.SpanExporter
	metricReaders  []sdkmetric.Reader
	mu             sync.RWMutex
}

// NewExporterManager creates a new exporter manager
func NewExporterManager(config ExporterConfig) (*ExporterManager, error) {
	em := &ExporterManager{
		config:         config,
		traceExporters: make([]sdktrace.SpanExporter, 0),
		metricReaders:  make([]sdkmetric.Reader, 0),
	}

	// Initialize trace exporters
	if err := em.initTraceExporters(); err != nil {
		return nil, fmt.Errorf("failed to initialize trace exporters: %w", err)
	}

	// Initialize metric exporters
	if err := em.initMetricExporters(); err != nil {
		return nil, fmt.Errorf("failed to initialize metric exporters: %w", err)
	}

	return em, nil
}

// initTraceExporters initializes trace exporters
func (em *ExporterManager) initTraceExporters() error {
	// OTLP trace exporter
	if em.config.OTLP.Enabled {
		exporter, err := em.createOTLPTraceExporter()
		if err != nil {
			return fmt.Errorf("failed to create OTLP trace exporter: %w", err)
		}
		em.traceExporters = append(em.traceExporters, exporter)
	}

	// Jaeger trace exporter (via OTLP)
	if em.config.Jaeger.Enabled {
		exporter, err := em.createJaegerTraceExporter()
		if err != nil {
			return fmt.Errorf("failed to create Jaeger trace exporter: %w", err)
		}
		em.traceExporters = append(em.traceExporters, exporter)
	}

	// Add stdout exporter for development
	stdoutExporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
	if err == nil {
		em.traceExporters = append(em.traceExporters, stdoutExporter)
	}

	return nil
}

// initMetricExporters initializes metric exporters
func (em *ExporterManager) initMetricExporters() error {
	// Prometheus exporter
	if em.config.Prometheus.Enabled {
		reader, err := em.createPrometheusReader()
		if err != nil {
			return fmt.Errorf("failed to create Prometheus reader: %w", err)
		}
		em.metricReaders = append(em.metricReaders, reader)
	}

	// OTLP metric exporter
	if em.config.OTLP.Enabled {
		reader, err := em.createOTLPMetricReader()
		if err != nil {
			return fmt.Errorf("failed to create OTLP metric reader: %w", err)
		}
		em.metricReaders = append(em.metricReaders, reader)
	}

	return nil
}

// createOTLPTraceExporter creates an OTLP trace exporter
func (em *ExporterManager) createOTLPTraceExporter() (sdktrace.SpanExporter, error) {
	ctx := context.Background()

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(em.config.OTLP.Endpoint),
		otlptracegrpc.WithTimeout(30 * time.Second),
	}

	if em.config.OTLP.Insecure {
		opts = append(opts, otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return exporter, nil
}

// createJaegerTraceExporter creates a Jaeger trace exporter (via OTLP)
func (em *ExporterManager) createJaegerTraceExporter() (sdktrace.SpanExporter, error) {
	ctx := context.Background()

	// Jaeger supports OTLP protocol on port 4317
	endpoint := em.config.Jaeger.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("%s:4317", em.config.Jaeger.AgentHost)
	}

	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()),
		otlptracegrpc.WithTimeout(30*time.Second),
	)
	if err != nil {
		return nil, err
	}

	return exporter, nil
}

// createPrometheusReader creates a Prometheus reader
func (em *ExporterManager) createPrometheusReader() (sdkmetric.Reader, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	return exporter, nil
}

// createOTLPMetricReader creates an OTLP metric reader
func (em *ExporterManager) createOTLPMetricReader() (sdkmetric.Reader, error) {
	ctx := context.Background()

	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(em.config.OTLP.Endpoint),
		otlpmetricgrpc.WithTimeout(30 * time.Second),
	}

	if em.config.OTLP.Insecure {
		opts = append(opts, otlpmetricgrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	exporter, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	// Create periodic reader with 15 second interval
	reader := sdkmetric.NewPeriodicReader(
		exporter,
		sdkmetric.WithInterval(15*time.Second),
	)

	return reader, nil
}

// GetTraceExporters returns all trace exporters
func (em *ExporterManager) GetTraceExporters() []sdktrace.SpanExporter {
	em.mu.RLock()
	defer em.mu.RUnlock()

	return em.traceExporters
}

// GetMetricReaders returns all metric readers
func (em *ExporterManager) GetMetricReaders() []sdkmetric.Reader {
	em.mu.RLock()
	defer em.mu.RUnlock()

	return em.metricReaders
}

// Shutdown shuts down all exporters
func (em *ExporterManager) Shutdown(ctx context.Context) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	var errors []error

	// Shutdown trace exporters
	for _, exporter := range em.traceExporters {
		if err := exporter.Shutdown(ctx); err != nil {
			errors = append(errors, err)
		}
	}

	// Metric readers are shut down by the meter provider

	if len(errors) > 0 {
		return fmt.Errorf("exporter shutdown errors: %v", errors)
	}

	return nil
}
