package observability

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Tracer wraps OpenTelemetry tracer with additional functionality
type Tracer interface {
	Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, Span)
	Extract(ctx context.Context, carrier interface{}) context.Context
	Inject(ctx context.Context, carrier interface{})
}

// Span wraps OpenTelemetry span with additional functionality
type Span interface {
	End()
	SetAttributes(attrs ...attribute.KeyValue)
	SetStatus(code codes.Code, description string)
	RecordError(err error, opts ...trace.EventOption)
	AddEvent(name string, opts ...trace.EventOption)
	SpanContext() trace.SpanContext
}

// TracerImpl implements the Tracer interface
type TracerImpl struct {
	tracer trace.Tracer
}

// SpanImpl implements the Span interface
type SpanImpl struct {
	span trace.Span
}

// NewTracerProvider creates a new tracer provider
func NewTracerProvider(
	serviceName string,
	serviceVersion string,
	environment string,
	config TracingConfig,
	exporters *ExporterManager,
) (*sdktrace.TracerProvider, error) {
	// Create resource
	resource, err := sdkresource.Merge(
		sdkresource.Default(),
		sdkresource.NewWithAttributes(
			"",
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			attribute.String("environment", environment),
		),
	)
	if err != nil {
		return nil, err
	}

	// Configure sampler
	sampler := sdktrace.ParentBased(
		sdktrace.TraceIDRatioBased(config.SamplingRate),
	)

	// Create span processor options
	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(resource),
	}

	// Add trace exporters
	if exporters != nil {
		for _, exporter := range exporters.GetTraceExporters() {
			opts = append(opts, sdktrace.WithBatcher(exporter))
		}
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(opts...)

	return tp, nil
}

// NewTracer creates a new tracer
func NewTracer(tracer trace.Tracer) *TracerImpl {
	return &TracerImpl{
		tracer: tracer,
	}
}

// Start starts a new span
func (t *TracerImpl) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, Span) {
	ctx, span := t.tracer.Start(ctx, name, opts...)
	return ctx, &SpanImpl{span: span}
}

// Extract extracts trace context from carrier (e.g., HTTP headers)
func (t *TracerImpl) Extract(ctx context.Context, carrier interface{}) context.Context {
	// This is handled by the global propagator
	return ctx
}

// Inject injects trace context into carrier (e.g., HTTP headers)
func (t *TracerImpl) Inject(ctx context.Context, carrier interface{}) {
	// This is handled by the global propagator
}

// End ends the span
func (s *SpanImpl) End() {
	s.span.End()
}

// SetAttributes sets attributes on the span
func (s *SpanImpl) SetAttributes(attrs ...attribute.KeyValue) {
	s.span.SetAttributes(attrs...)
}

// SetStatus sets the status of the span
func (s *SpanImpl) SetStatus(code codes.Code, description string) {
	s.span.SetStatus(code, description)
}

// RecordError records an error on the span
func (s *SpanImpl) RecordError(err error, opts ...trace.EventOption) {
	s.span.RecordError(err, opts...)
}

// AddEvent adds an event to the span
func (s *SpanImpl) AddEvent(name string, opts ...trace.EventOption) {
	s.span.AddEvent(name, opts...)
}

// SpanContext returns the span context
func (s *SpanImpl) SpanContext() trace.SpanContext {
	return s.span.SpanContext()
}

// NoopTracer is a no-op implementation of Tracer
type NoopTracer struct{}

// NoopSpan is a no-op implementation of Span
type NoopSpan struct{}

// Start starts a no-op span
func (t *NoopTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, Span) {
	return ctx, &NoopSpan{}
}

// Extract does nothing for no-op tracer
func (t *NoopTracer) Extract(ctx context.Context, carrier interface{}) context.Context {
	return ctx
}

// Inject does nothing for no-op tracer
func (t *NoopTracer) Inject(ctx context.Context, carrier interface{}) {}

// End does nothing for no-op span
func (s *NoopSpan) End() {}

// SetAttributes does nothing for no-op span
func (s *NoopSpan) SetAttributes(attrs ...attribute.KeyValue) {}

// SetStatus does nothing for no-op span
func (s *NoopSpan) SetStatus(code codes.Code, description string) {}

// RecordError does nothing for no-op span
func (s *NoopSpan) RecordError(err error, opts ...trace.EventOption) {}

// AddEvent does nothing for no-op span
func (s *NoopSpan) AddEvent(name string, opts ...trace.EventOption) {}

// SpanContext returns an invalid span context
func (s *NoopSpan) SpanContext() trace.SpanContext {
	return trace.SpanContext{}
}

// Common attribute keys for containr
var (
	ContainerIDKey   = attribute.Key("container.id")
	ContainerNameKey = attribute.Key("container.name")
	ImageIDKey       = attribute.Key("image.id")
	ImageNameKey     = attribute.Key("image.name")
	OperationKey     = attribute.Key("operation")
	UserKey          = attribute.Key("user")
	NamespaceKey     = attribute.Key("namespace")
)

// Helper functions for common span operations

// StartContainerSpan starts a span for a container operation
func StartContainerSpan(ctx context.Context, tracer Tracer, operation string, containerID string, containerName string) (context.Context, Span) {
	ctx, span := tracer.Start(ctx, operation)
	span.SetAttributes(
		OperationKey.String(operation),
		ContainerIDKey.String(containerID),
		ContainerNameKey.String(containerName),
	)
	return ctx, span
}

// StartImageSpan starts a span for an image operation
func StartImageSpan(ctx context.Context, tracer Tracer, operation string, imageID string, imageName string) (context.Context, Span) {
	ctx, span := tracer.Start(ctx, operation)
	span.SetAttributes(
		OperationKey.String(operation),
		ImageIDKey.String(imageID),
		ImageNameKey.String(imageName),
	)
	return ctx, span
}

// FinishSpanWithError finishes a span and records an error if present
func FinishSpanWithError(span Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End()
}
