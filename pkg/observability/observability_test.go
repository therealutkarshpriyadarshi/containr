package observability

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name:      "default config",
			config:    nil,
			expectErr: false,
		},
		{
			name:      "custom config",
			config:    DefaultConfig(),
			expectErr: false,
		},
		{
			name: "disabled observability",
			config: &Config{
				Enabled: false,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewManager(tt.config)
			if (err != nil) != tt.expectErr {
				t.Errorf("NewManager() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if mgr != nil {
				defer mgr.Shutdown(context.Background())
			}
		})
	}
}

func TestManager_GetTracer(t *testing.T) {
	config := DefaultConfig()
	config.Tracing.Enabled = true
	config.Exporters.Prometheus.Enabled = false
	config.Exporters.Jaeger.Enabled = false
	config.Exporters.OTLP.Enabled = false

	mgr, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	tracer := mgr.GetTracer("test")
	if tracer == nil {
		t.Fatal("GetTracer returned nil")
	}

	// Test span creation
	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "test-span")
	defer span.End()

	// Test span attributes
	span.SetAttributes(
		attribute.String("test.key", "test.value"),
		attribute.Int("test.int", 42),
	)

	// Test span status
	span.SetStatus(codes.Ok, "success")

	// Test span event
	span.AddEvent("test-event")
}

func TestManager_GetMetrics(t *testing.T) {
	config := DefaultConfig()
	config.Metrics.Enabled = true
	config.Exporters.Prometheus.Enabled = false
	config.Exporters.Jaeger.Enabled = false
	config.Exporters.OTLP.Enabled = false

	mgr, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	metrics := mgr.GetMetrics()
	if metrics == nil {
		t.Fatal("GetMetrics returned nil")
	}

	ctx := context.Background()

	// Test container metrics
	metrics.RecordContainerCreated(ctx)
	metrics.RecordContainerStarted(ctx)
	metrics.RecordContainerStopped(ctx)
	metrics.RecordContainerDeleted(ctx)
	metrics.RecordContainerError(ctx)

	// Test image metrics
	metrics.RecordImagePulled(ctx)
	metrics.RecordImagePushed(ctx)
	metrics.RecordImageDeleted(ctx)

	// Test resource metrics
	metrics.RecordCPUUsage(ctx, 1.5)
	metrics.RecordMemoryUsage(ctx, 1024*1024*1024)
	metrics.RecordDiskUsage(ctx, 10*1024*1024*1024)
	metrics.RecordNetworkRx(ctx, 1024)
	metrics.RecordNetworkTx(ctx, 2048)

	// Test operation metrics
	metrics.RecordOperationDuration(ctx, 0.5)
	metrics.RecordOperation(ctx)
}

func TestManager_GetLogger(t *testing.T) {
	config := DefaultConfig()
	config.Logging.Enabled = true
	config.Logging.Level = "info"
	config.Logging.Format = "json"

	mgr, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	logger := mgr.GetLogger()
	if logger == nil {
		t.Fatal("GetLogger returned nil")
	}

	// Test logging methods
	logger.Info("test info message")
	logger.Infof("test info message: %s", "formatted")
	logger.Debug("test debug message")
	logger.Warn("test warning message")
	logger.Error("test error message")

	// Test context logging
	ctx := context.Background()
	logger.WithContext(ctx).Info("test context message")

	// Test field logging
	logger.WithField("key", "value").Info("test field message")
	logger.WithFields(map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}).Info("test fields message")
}

func TestTracer_SpanOperations(t *testing.T) {
	config := DefaultConfig()
	config.Tracing.Enabled = true
	config.Exporters.Prometheus.Enabled = false
	config.Exporters.Jaeger.Enabled = false
	config.Exporters.OTLP.Enabled = false

	mgr, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	tracer := mgr.GetTracer("test")
	ctx := context.Background()

	// Test container span
	ctx, span := StartContainerSpan(ctx, tracer, "container.create", "container-123", "test-container")
	span.SetAttributes(attribute.String("user", "testuser"))
	span.AddEvent("container created")
	FinishSpanWithError(span, nil)

	// Test image span
	ctx, span = StartImageSpan(ctx, tracer, "image.pull", "image-456", "test-image:latest")
	span.SetAttributes(attribute.String("registry", "docker.io"))
	FinishSpanWithError(span, nil)

	// Test span with error
	ctx, span = tracer.Start(ctx, "error-operation")
	testErr := &testError{msg: "test error"}
	FinishSpanWithError(span, testErr)
}

func TestMetricsManager_RecordMetrics(t *testing.T) {
	config := DefaultConfig()
	config.Metrics.Enabled = true
	config.Exporters.Prometheus.Enabled = false
	config.Exporters.OTLP.Enabled = false

	mgr, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	metrics := mgr.GetMetrics()
	ctx := context.Background()

	// Test metric recording with attributes
	attrs := []metric.AddOption{
		metric.WithAttributes(
			attribute.String("container.name", "test-container"),
			attribute.String("namespace", "default"),
		),
	}

	metrics.RecordContainerCreated(ctx, attrs...)
	metrics.RecordContainerStarted(ctx, attrs...)

	// Wait briefly for metrics to be recorded
	time.Sleep(100 * time.Millisecond)
}

func TestLogger_Levels(t *testing.T) {
	tests := []struct {
		name  string
		level string
		valid bool
	}{
		{"debug level", "debug", true},
		{"info level", "info", true},
		{"warn level", "warn", true},
		{"error level", "error", true},
		{"fatal level", "fatal", true},
		{"invalid level", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := LoggingConfig{
				Enabled: true,
				Level:   tt.level,
				Format:  "json",
				Output:  "stdout",
			}

			logger, err := NewLogger(config)
			if tt.valid {
				if err != nil {
					t.Errorf("NewLogger() error = %v, expected valid", err)
				} else {
					defer logger.Close()

					// Test level change
					newLevel := "info"
					if err := logger.SetLevel(newLevel); err != nil {
						t.Errorf("SetLevel() error = %v", err)
					}

					if logger.GetLevel() != newLevel {
						t.Errorf("GetLevel() = %v, expected %v", logger.GetLevel(), newLevel)
					}
				}
			} else {
				if err == nil {
					t.Error("NewLogger() expected error for invalid level")
					logger.Close()
				}
			}
		})
	}
}

func TestLogger_Formats(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"json format", "json"},
		{"text format", "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := LoggingConfig{
				Enabled: true,
				Level:   "info",
				Format:  tt.format,
				Output:  "stdout",
			}

			logger, err := NewLogger(config)
			if err != nil {
				t.Fatalf("NewLogger() error = %v", err)
			}
			defer logger.Close()

			// Test logging with different formats
			logger.Info("test message")
			logger.WithField("key", "value").Info("test with field")
		})
	}
}

func TestLogger_ContextLogging(t *testing.T) {
	config := DefaultConfig()
	config.Tracing.Enabled = true
	config.Logging.Enabled = true
	config.Exporters.Prometheus.Enabled = false
	config.Exporters.Jaeger.Enabled = false
	config.Exporters.OTLP.Enabled = false

	mgr, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer mgr.Shutdown(context.Background())

	logger := mgr.GetLogger()
	tracer := mgr.GetTracer("test")

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "test-operation")
	defer span.End()

	// Log with trace context
	logger.WithContext(ctx).Info("operation started")

	// Test helper functions
	LogContainerOperation(logger, ctx, "create", "container-123", "test-container").Info("container created")
	LogImageOperation(logger, ctx, "pull", "image-456", "test-image:latest").Info("image pulled")
	LogError(logger, ctx, "test-operation", &testError{msg: "test error"})
}

func TestExporterManager(t *testing.T) {
	tests := []struct {
		name      string
		config    ExporterConfig
		expectErr bool
	}{
		{
			name: "no exporters",
			config: ExporterConfig{
				Prometheus: PrometheusConfig{Enabled: false},
				Jaeger:     JaegerConfig{Enabled: false},
				OTLP:       OTLPConfig{Enabled: false},
			},
			expectErr: false,
		},
		{
			name: "prometheus only",
			config: ExporterConfig{
				Prometheus: PrometheusConfig{
					Enabled:  true,
					Endpoint: "localhost:9090",
					Port:     9090,
				},
				Jaeger: JaegerConfig{Enabled: false},
				OTLP:   OTLPConfig{Enabled: false},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em, err := NewExporterManager(tt.config)
			if (err != nil) != tt.expectErr {
				t.Errorf("NewExporterManager() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if em != nil {
				defer em.Shutdown(context.Background())

				// Verify exporters
				_ = em.GetTraceExporters()
				metricReaders := em.GetMetricReaders()

				if tt.config.Prometheus.Enabled && len(metricReaders) == 0 {
					t.Error("Expected Prometheus metric reader")
				}
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.ServiceName != "containr" {
		t.Errorf("Expected service name 'containr', got '%s'", config.ServiceName)
	}

	if !config.Enabled {
		t.Error("Expected observability to be enabled by default")
	}

	if !config.Tracing.Enabled {
		t.Error("Expected tracing to be enabled by default")
	}

	if !config.Metrics.Enabled {
		t.Error("Expected metrics to be enabled by default")
	}

	if !config.Logging.Enabled {
		t.Error("Expected logging to be enabled by default")
	}

	if config.Tracing.SamplingRate != 1.0 {
		t.Errorf("Expected sampling rate 1.0, got %f", config.Tracing.SamplingRate)
	}

	if config.Metrics.Port != 9090 {
		t.Errorf("Expected metrics port 9090, got %d", config.Metrics.Port)
	}

	if config.Logging.Level != "info" {
		t.Errorf("Expected log level 'info', got '%s'", config.Logging.Level)
	}
}

func TestNoopTracer(t *testing.T) {
	tracer := &NoopTracer{}
	ctx := context.Background()

	// Test that noop tracer doesn't panic
	ctx, span := tracer.Start(ctx, "test-span")
	span.SetAttributes(attribute.String("key", "value"))
	span.SetStatus(codes.Ok, "success")
	span.RecordError(&testError{msg: "test error"})
	span.AddEvent("test-event")
	span.End()

	// Test context propagation
	ctx = tracer.Extract(ctx, nil)
	tracer.Inject(ctx, nil)

	// Verify span context is invalid
	spanContext := span.SpanContext()
	if spanContext.IsValid() {
		t.Error("NoopSpan should have invalid span context")
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
		valid    bool
	}{
		{"debug", DebugLevel, true},
		{"info", InfoLevel, true},
		{"warn", WarnLevel, true},
		{"warning", WarnLevel, true},
		{"error", ErrorLevel, true},
		{"fatal", FatalLevel, true},
		{"DEBUG", DebugLevel, true},
		{"INFO", InfoLevel, true},
		{"invalid", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			level, err := ParseLogLevel(tt.input)
			if tt.valid {
				if err != nil {
					t.Errorf("ParseLogLevel(%s) error = %v, expected valid", tt.input, err)
				}
				if level != tt.expected {
					t.Errorf("ParseLogLevel(%s) = %v, expected %v", tt.input, level, tt.expected)
				}
			} else {
				if err == nil {
					t.Errorf("ParseLogLevel(%s) expected error", tt.input)
				}
			}
		})
	}
}

func TestManager_Shutdown(t *testing.T) {
	config := DefaultConfig()
	config.Exporters.Jaeger.Enabled = false
	config.Exporters.OTLP.Enabled = false

	mgr, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Use observability components
	tracer := mgr.GetTracer("test")
	metrics := mgr.GetMetrics()
	logger := mgr.GetLogger()

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "test-span")
	span.End()

	metrics.RecordContainerCreated(ctx)
	logger.Info("test message")

	// Shutdown
	if err := mgr.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Verify components still work (gracefully degrade)
	tracer.Start(context.Background(), "after-shutdown")
	metrics.RecordContainerCreated(context.Background())
	logger.Info("after shutdown")
}

func TestTracingSamplingRates(t *testing.T) {
	tests := []struct {
		name         string
		samplingRate float64
	}{
		{"always sample", 1.0},
		{"half sample", 0.5},
		{"never sample", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultConfig()
			config.Tracing.SamplingRate = tt.samplingRate
			config.Exporters.Prometheus.Enabled = false
			config.Exporters.Jaeger.Enabled = false
			config.Exporters.OTLP.Enabled = false

			mgr, err := NewManager(config)
			if err != nil {
				t.Fatalf("NewManager failed: %v", err)
			}
			defer mgr.Shutdown(context.Background())

			tracer := mgr.GetTracer("test")
			ctx := context.Background()

			// Create multiple spans to test sampling
			for i := 0; i < 10; i++ {
				_, span := tracer.Start(ctx, "test-span")
				span.End()
			}
		})
	}
}

// Helper types for testing

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
