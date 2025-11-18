package observability

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// MetricsManager manages metrics collection
type MetricsManager struct {
	config         MetricsConfig
	meterProvider  *sdkmetric.MeterProvider
	meter          metric.Meter

	// Container metrics
	containerCount       metric.Int64UpDownCounter
	containerCreateTotal metric.Int64Counter
	containerStartTotal  metric.Int64Counter
	containerStopTotal   metric.Int64Counter
	containerDeleteTotal metric.Int64Counter
	containerErrorTotal  metric.Int64Counter

	// Image metrics
	imageCount       metric.Int64UpDownCounter
	imagePullTotal   metric.Int64Counter
	imagePushTotal   metric.Int64Counter
	imageDeleteTotal metric.Int64Counter

	// Resource metrics
	cpuUsage    metric.Float64Histogram
	memoryUsage metric.Int64Histogram
	diskUsage   metric.Int64Histogram
	networkRx   metric.Int64Counter
	networkTx   metric.Int64Counter

	// Operation metrics
	operationDuration metric.Float64Histogram
	operationTotal    metric.Int64Counter

	mu sync.RWMutex
}

// NewMetricsManager creates a new metrics manager
func NewMetricsManager(serviceName string, config MetricsConfig, exporters *ExporterManager) (*MetricsManager, error) {
	mm := &MetricsManager{
		config: config,
	}

	// Create meter provider
	opts := []sdkmetric.Option{}

	// Add metric exporters
	if exporters != nil {
		for _, reader := range exporters.GetMetricReaders() {
			opts = append(opts, sdkmetric.WithReader(reader))
		}
	}

	mp := sdkmetric.NewMeterProvider(opts...)
	mm.meterProvider = mp

	// Get meter
	mm.meter = mp.Meter(serviceName)

	// Initialize metrics
	if err := mm.initMetrics(); err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	return mm, nil
}

// initMetrics initializes all metrics
func (mm *MetricsManager) initMetrics() error {
	var err error

	// Container metrics
	mm.containerCount, err = mm.meter.Int64UpDownCounter(
		"containr_containers",
		metric.WithDescription("Number of containers"),
	)
	if err != nil {
		return err
	}

	mm.containerCreateTotal, err = mm.meter.Int64Counter(
		"containr_container_create_total",
		metric.WithDescription("Total number of container creations"),
	)
	if err != nil {
		return err
	}

	mm.containerStartTotal, err = mm.meter.Int64Counter(
		"containr_container_start_total",
		metric.WithDescription("Total number of container starts"),
	)
	if err != nil {
		return err
	}

	mm.containerStopTotal, err = mm.meter.Int64Counter(
		"containr_container_stop_total",
		metric.WithDescription("Total number of container stops"),
	)
	if err != nil {
		return err
	}

	mm.containerDeleteTotal, err = mm.meter.Int64Counter(
		"containr_container_delete_total",
		metric.WithDescription("Total number of container deletions"),
	)
	if err != nil {
		return err
	}

	mm.containerErrorTotal, err = mm.meter.Int64Counter(
		"containr_container_error_total",
		metric.WithDescription("Total number of container errors"),
	)
	if err != nil {
		return err
	}

	// Image metrics
	mm.imageCount, err = mm.meter.Int64UpDownCounter(
		"containr_images",
		metric.WithDescription("Number of images"),
	)
	if err != nil {
		return err
	}

	mm.imagePullTotal, err = mm.meter.Int64Counter(
		"containr_image_pull_total",
		metric.WithDescription("Total number of image pulls"),
	)
	if err != nil {
		return err
	}

	mm.imagePushTotal, err = mm.meter.Int64Counter(
		"containr_image_push_total",
		metric.WithDescription("Total number of image pushes"),
	)
	if err != nil {
		return err
	}

	mm.imageDeleteTotal, err = mm.meter.Int64Counter(
		"containr_image_delete_total",
		metric.WithDescription("Total number of image deletions"),
	)
	if err != nil {
		return err
	}

	// Resource metrics
	mm.cpuUsage, err = mm.meter.Float64Histogram(
		"containr_cpu_usage",
		metric.WithDescription("CPU usage in cores"),
		metric.WithUnit("cores"),
	)
	if err != nil {
		return err
	}

	mm.memoryUsage, err = mm.meter.Int64Histogram(
		"containr_memory_usage_bytes",
		metric.WithDescription("Memory usage in bytes"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		return err
	}

	mm.diskUsage, err = mm.meter.Int64Histogram(
		"containr_disk_usage_bytes",
		metric.WithDescription("Disk usage in bytes"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		return err
	}

	mm.networkRx, err = mm.meter.Int64Counter(
		"containr_network_rx_bytes",
		metric.WithDescription("Network bytes received"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		return err
	}

	mm.networkTx, err = mm.meter.Int64Counter(
		"containr_network_tx_bytes",
		metric.WithDescription("Network bytes transmitted"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		return err
	}

	// Operation metrics
	mm.operationDuration, err = mm.meter.Float64Histogram(
		"containr_operation_duration_seconds",
		metric.WithDescription("Operation duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	mm.operationTotal, err = mm.meter.Int64Counter(
		"containr_operation_total",
		metric.WithDescription("Total number of operations"),
	)
	if err != nil {
		return err
	}

	return nil
}

// RecordContainerCreated records a container creation
func (mm *MetricsManager) RecordContainerCreated(ctx context.Context, attrs ...metric.AddOption) {
	mm.containerCount.Add(ctx, 1, attrs...)
	mm.containerCreateTotal.Add(ctx, 1, attrs...)
}

// RecordContainerDeleted records a container deletion
func (mm *MetricsManager) RecordContainerDeleted(ctx context.Context, attrs ...metric.AddOption) {
	mm.containerCount.Add(ctx, -1, attrs...)
	mm.containerDeleteTotal.Add(ctx, 1, attrs...)
}

// RecordContainerStarted records a container start
func (mm *MetricsManager) RecordContainerStarted(ctx context.Context, attrs ...metric.AddOption) {
	mm.containerStartTotal.Add(ctx, 1, attrs...)
}

// RecordContainerStopped records a container stop
func (mm *MetricsManager) RecordContainerStopped(ctx context.Context, attrs ...metric.AddOption) {
	mm.containerStopTotal.Add(ctx, 1, attrs...)
}

// RecordContainerError records a container error
func (mm *MetricsManager) RecordContainerError(ctx context.Context, attrs ...metric.AddOption) {
	mm.containerErrorTotal.Add(ctx, 1, attrs...)
}

// RecordImagePulled records an image pull
func (mm *MetricsManager) RecordImagePulled(ctx context.Context, attrs ...metric.AddOption) {
	mm.imageCount.Add(ctx, 1, attrs...)
	mm.imagePullTotal.Add(ctx, 1, attrs...)
}

// RecordImagePushed records an image push
func (mm *MetricsManager) RecordImagePushed(ctx context.Context, attrs ...metric.AddOption) {
	mm.imagePushTotal.Add(ctx, 1, attrs...)
}

// RecordImageDeleted records an image deletion
func (mm *MetricsManager) RecordImageDeleted(ctx context.Context, attrs ...metric.AddOption) {
	mm.imageCount.Add(ctx, -1, attrs...)
	mm.imageDeleteTotal.Add(ctx, 1, attrs...)
}

// RecordCPUUsage records CPU usage
func (mm *MetricsManager) RecordCPUUsage(ctx context.Context, usage float64, attrs ...metric.RecordOption) {
	mm.cpuUsage.Record(ctx, usage, attrs...)
}

// RecordMemoryUsage records memory usage
func (mm *MetricsManager) RecordMemoryUsage(ctx context.Context, usage int64, attrs ...metric.RecordOption) {
	mm.memoryUsage.Record(ctx, usage, attrs...)
}

// RecordDiskUsage records disk usage
func (mm *MetricsManager) RecordDiskUsage(ctx context.Context, usage int64, attrs ...metric.RecordOption) {
	mm.diskUsage.Record(ctx, usage, attrs...)
}

// RecordNetworkRx records network bytes received
func (mm *MetricsManager) RecordNetworkRx(ctx context.Context, bytes int64, attrs ...metric.AddOption) {
	mm.networkRx.Add(ctx, bytes, attrs...)
}

// RecordNetworkTx records network bytes transmitted
func (mm *MetricsManager) RecordNetworkTx(ctx context.Context, bytes int64, attrs ...metric.AddOption) {
	mm.networkTx.Add(ctx, bytes, attrs...)
}

// RecordOperationDuration records operation duration
func (mm *MetricsManager) RecordOperationDuration(ctx context.Context, duration float64, attrs ...metric.RecordOption) {
	mm.operationDuration.Record(ctx, duration, attrs...)
}

// RecordOperation records an operation
func (mm *MetricsManager) RecordOperation(ctx context.Context, attrs ...metric.AddOption) {
	mm.operationTotal.Add(ctx, 1, attrs...)
}

// Shutdown shuts down the metrics manager
func (mm *MetricsManager) Shutdown(ctx context.Context) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if mm.meterProvider != nil {
		return mm.meterProvider.Shutdown(ctx)
	}

	return nil
}
