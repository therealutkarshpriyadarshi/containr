# Observability Package

Comprehensive observability implementation for containr using OpenTelemetry, providing distributed tracing, metrics collection, and structured logging.

## Features

### Distributed Tracing
- OpenTelemetry-based distributed tracing
- Context propagation for distributed systems
- Support for Jaeger and OTLP exporters
- Configurable sampling rates
- Container and image operation tracking

### Metrics Collection
- Prometheus-compatible metrics export
- Container lifecycle metrics (create, start, stop, delete)
- Image operation metrics (pull, push, delete)
- Resource usage metrics (CPU, memory, disk, network)
- Operation duration and count tracking

### Structured Logging
- JSON and text format support
- Trace context integration
- Multiple log levels (debug, info, warn, error, fatal)
- Flexible output (stdout, stderr, file)
- Context-aware logging helpers

### Exporters
- Prometheus metrics exporter
- Jaeger/OTLP trace exporters
- OTLP metrics and traces support
- Stdout exporters for development

## Files

- **observability.go** (316 lines) - Main observability manager and configuration
- **tracing.go** (224 lines) - Distributed tracing implementation
- **metrics.go** (318 lines) - Prometheus metrics collection
- **logging.go** (248 lines) - Structured logging with trace context
- **exporter.go** (219 lines) - Exporters for Prometheus and Jaeger
- **observability_test.go** (568 lines) - Comprehensive test suite

## Usage

### Basic Setup

```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/observability"

// Create observability manager with default config
mgr, err := observability.NewManager(nil)
if err != nil {
    log.Fatal(err)
}
defer mgr.Shutdown(context.Background())
```

### Custom Configuration

```go
config := &observability.Config{
    ServiceName:    "containr",
    ServiceVersion: "1.0.0",
    Environment:    "production",
    Enabled:        true,
    Tracing: observability.TracingConfig{
        Enabled:      true,
        SamplingRate: 0.5, // Sample 50% of traces
    },
    Metrics: observability.MetricsConfig{
        Enabled: true,
        Port:    9090,
        Path:    "/metrics",
    },
    Logging: observability.LoggingConfig{
        Enabled: true,
        Level:   "info",
        Format:  "json",
        Output:  "stdout",
    },
    Exporters: observability.ExporterConfig{
        Prometheus: observability.PrometheusConfig{
            Enabled:  true,
            Port:     9090,
        },
        Jaeger: observability.JaegerConfig{
            Enabled:   true,
            Endpoint:  "http://localhost:14268/api/traces",
        },
    },
}

mgr, err := observability.NewManager(config)
```

### Distributed Tracing

```go
// Get tracer
tracer := mgr.GetTracer("container-runtime")

// Start a span
ctx, span := tracer.Start(ctx, "container.create")
defer span.End()

// Add attributes
span.SetAttributes(
    observability.ContainerIDKey.String("container-123"),
    observability.ContainerNameKey.String("my-container"),
)

// Record events
span.AddEvent("container created")

// Handle errors
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
}
```

### Helper Functions

```go
// Container operation span
ctx, span := observability.StartContainerSpan(
    ctx, tracer, "container.create",
    "container-123", "my-container",
)
defer observability.FinishSpanWithError(span, err)

// Image operation span
ctx, span := observability.StartImageSpan(
    ctx, tracer, "image.pull",
    "image-456", "nginx:latest",
)
defer observability.FinishSpanWithError(span, err)
```

### Metrics Collection

```go
// Get metrics manager
metrics := mgr.GetMetrics()

// Record container operations
ctx := context.Background()
metrics.RecordContainerCreated(ctx)
metrics.RecordContainerStarted(ctx)
metrics.RecordContainerStopped(ctx)
metrics.RecordContainerDeleted(ctx)

// Record image operations
metrics.RecordImagePulled(ctx)
metrics.RecordImagePushed(ctx)

// Record resource usage
metrics.RecordCPUUsage(ctx, 1.5)
metrics.RecordMemoryUsage(ctx, 1024*1024*1024) // 1GB
metrics.RecordNetworkRx(ctx, 1024)
metrics.RecordNetworkTx(ctx, 2048)

// Record operation metrics
metrics.RecordOperationDuration(ctx, 0.5) // 500ms
metrics.RecordOperation(ctx)
```

### Structured Logging

```go
// Get logger
logger := mgr.GetLogger()

// Basic logging
logger.Info("Container started")
logger.Infof("Container %s started", containerID)
logger.Error("Failed to start container")
logger.WithError(err).Error("Container error")

// Context logging (includes trace ID and span ID)
logger.WithContext(ctx).Info("Processing request")

// Field logging
logger.WithField("container_id", "123").Info("Container created")
logger.WithFields(map[string]interface{}{
    "container_id":   "123",
    "container_name": "my-container",
}).Info("Container created")
```

### Helper Functions for Logging

```go
// Container operation logging
observability.LogContainerOperation(
    logger, ctx, "create", "container-123", "my-container",
).Info("Container created successfully")

// Image operation logging
observability.LogImageOperation(
    logger, ctx, "pull", "image-456", "nginx:latest",
).Info("Image pulled successfully")

// Error logging
observability.LogError(logger, ctx, "container.create", err)
```

## Configuration

### Default Configuration

```yaml
service_name: containr
service_version: 1.0.0
environment: development
enabled: true

tracing:
  enabled: true
  sampling_rate: 1.0
  max_spans_per_trace: 1000

metrics:
  enabled: true
  port: 9090
  path: /metrics
  collect_interval_seconds: 15

logging:
  enabled: true
  level: info
  format: json
  output: stdout

exporters:
  prometheus:
    enabled: true
    endpoint: localhost:9090
    port: 9090
  jaeger:
    enabled: false
    endpoint: http://localhost:14268/api/traces
    agent_host: localhost
    agent_port: 6831
  otlp:
    enabled: false
    endpoint: localhost:4317
    insecure: true
```

## Metrics

### Container Metrics
- `containr_containers` - Current number of containers
- `containr_container_create_total` - Total container creations
- `containr_container_start_total` - Total container starts
- `containr_container_stop_total` - Total container stops
- `containr_container_delete_total` - Total container deletions
- `containr_container_error_total` - Total container errors

### Image Metrics
- `containr_images` - Current number of images
- `containr_image_pull_total` - Total image pulls
- `containr_image_push_total` - Total image pushes
- `containr_image_delete_total` - Total image deletions

### Resource Metrics
- `containr_cpu_usage` - CPU usage in cores
- `containr_memory_usage_bytes` - Memory usage in bytes
- `containr_disk_usage_bytes` - Disk usage in bytes
- `containr_network_rx_bytes` - Network bytes received
- `containr_network_tx_bytes` - Network bytes transmitted

### Operation Metrics
- `containr_operation_duration_seconds` - Operation duration
- `containr_operation_total` - Total operations

## Testing

Run tests:
```bash
go test ./pkg/observability/...
```

Run with coverage:
```bash
go test ./pkg/observability/... -cover
```

Current coverage: **71.0%**

## Integration

### Prometheus Integration

Metrics are exposed on the configured port (default: 9090) at the `/metrics` endpoint.

```bash
curl http://localhost:9090/metrics
```

### Jaeger Integration

Configure Jaeger endpoint in the configuration:

```yaml
exporters:
  jaeger:
    enabled: true
    endpoint: http://localhost:14268/api/traces
```

### OTLP Integration

For OpenTelemetry Collector:

```yaml
exporters:
  otlp:
    enabled: true
    endpoint: localhost:4317
    insecure: true
```

## Best Practices

1. **Always use context**: Pass context through the call chain to maintain trace propagation
2. **Defer span.End()**: Always defer span ending to ensure spans are completed
3. **Record errors**: Use `span.RecordError()` and `span.SetStatus()` for error tracking
4. **Use structured logging**: Prefer structured fields over string interpolation
5. **Graceful shutdown**: Always call `mgr.Shutdown()` to flush pending data

## Dependencies

- OpenTelemetry Go SDK v1.24.0
- Prometheus client
- Logrus for structured logging
- gRPC for OTLP exporters

## Architecture

The observability package follows a layered architecture:

1. **Manager Layer**: Central configuration and lifecycle management
2. **Component Layer**: Tracing, metrics, and logging implementations
3. **Exporter Layer**: Integration with external systems (Prometheus, Jaeger)
4. **Helper Layer**: Convenience functions for common operations

All components are thread-safe and can be used concurrently.

## License

Part of the containr project.
