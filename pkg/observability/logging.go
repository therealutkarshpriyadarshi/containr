package observability

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

// Logger provides structured logging with trace context
type Logger struct {
	logger *logrus.Logger
	config LoggingConfig
	output io.WriteCloser
	mu     sync.RWMutex
}

// LogLevel represents log level
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// NewLogger creates a new logger
func NewLogger(config LoggingConfig) (*Logger, error) {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	logger.SetLevel(level)

	// Set formatter
	if config.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
		})
	}

	// Set output
	var output io.WriteCloser
	switch config.Output {
	case "stdout", "":
		output = os.Stdout
		logger.SetOutput(os.Stdout)
	case "stderr":
		output = os.Stderr
		logger.SetOutput(os.Stderr)
	default:
		// File output
		f, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		output = f
		logger.SetOutput(f)
	}

	return &Logger{
		logger: logger,
		config: config,
		output: output,
	}, nil
}

// WithContext returns a logger with trace context
func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	entry := l.logger.WithContext(ctx)

	// Add trace context if available
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		entry = entry.WithFields(logrus.Fields{
			"trace_id": span.SpanContext().TraceID().String(),
			"span_id":  span.SpanContext().SpanID().String(),
		})
	}

	return entry
}

// WithField returns a logger with a single field
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.logger.WithField(key, value)
}

// WithFields returns a logger with multiple fields
func (l *Logger) WithFields(fields map[string]interface{}) *logrus.Entry {
	return l.logger.WithFields(fields)
}

// WithError returns a logger with an error field
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.logger.WithError(err)
}

// Debug logs a debug message
func (l *Logger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	l.logger.SetLevel(parsedLevel)
	return nil
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.logger.GetLevel().String()
}

// Close closes the logger
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Don't close stdout or stderr
	if l.output != os.Stdout && l.output != os.Stderr && l.output != nil {
		return l.output.Close()
	}

	return nil
}

// Helper functions for common logging patterns

// LogContainerOperation logs a container operation
func LogContainerOperation(logger *Logger, ctx context.Context, operation string, containerID string, containerName string) *logrus.Entry {
	return logger.WithContext(ctx).WithFields(logrus.Fields{
		"operation":      operation,
		"container_id":   containerID,
		"container_name": containerName,
	})
}

// LogImageOperation logs an image operation
func LogImageOperation(logger *Logger, ctx context.Context, operation string, imageID string, imageName string) *logrus.Entry {
	return logger.WithContext(ctx).WithFields(logrus.Fields{
		"operation":  operation,
		"image_id":   imageID,
		"image_name": imageName,
	})
}

// LogError logs an error with context
func LogError(logger *Logger, ctx context.Context, operation string, err error) {
	logger.WithContext(ctx).WithFields(logrus.Fields{
		"operation": operation,
		"error":     err.Error(),
	}).Error("Operation failed")
}

// ParseLogLevel parses a log level string
func ParseLogLevel(level string) (LogLevel, error) {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	default:
		return "", fmt.Errorf("invalid log level: %s", level)
	}
}
