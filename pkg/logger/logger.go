package logger

import (
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

// Level represents log levels
type Level uint32

const (
	// DebugLevel logs detailed debug information
	DebugLevel Level = iota
	// InfoLevel logs general informational messages
	InfoLevel
	// WarnLevel logs warnings
	WarnLevel
	// ErrorLevel logs errors
	ErrorLevel
	// FatalLevel logs fatal errors and exits
	FatalLevel
)

// Logger is a structured logger wrapper
type Logger struct {
	*logrus.Logger
	component string
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// init initializes the default logger
func init() {
	once.Do(func() {
		defaultLogger = newLogger("containr")
		defaultLogger.SetLevel(InfoLevel)
	})
}

// newLogger creates a new logger instance
func newLogger(component string) *Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   false,
	})
	log.SetOutput(os.Stdout)

	return &Logger{
		Logger:    log,
		component: component,
	}
}

// GetLogger returns the default logger
func GetLogger() *Logger {
	return defaultLogger
}

// New creates a new logger for a specific component
func New(component string) *Logger {
	return newLogger(component)
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level Level) {
	switch level {
	case DebugLevel:
		l.Logger.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		l.Logger.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		l.Logger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		l.Logger.SetLevel(logrus.ErrorLevel)
	case FatalLevel:
		l.Logger.SetLevel(logrus.FatalLevel)
	}
}

// SetOutput sets the output destination
func (l *Logger) SetOutput(out io.Writer) {
	l.Logger.SetOutput(out)
}

// WithField adds a single field to the logger
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		"component": l.component,
		key:         value,
	})
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *logrus.Entry {
	fields["component"] = l.component
	return l.Logger.WithFields(fields)
}

// WithError adds an error to the logger
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		"component": l.component,
	}).WithError(err)
}

// Debug logs a debug message
func (l *Logger) Debug(args ...interface{}) {
	l.WithField("component", l.component).Debug(args...)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.WithField("component", l.component).Debugf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(args ...interface{}) {
	l.WithField("component", l.component).Info(args...)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.WithField("component", l.component).Infof(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(args ...interface{}) {
	l.WithField("component", l.component).Warn(args...)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.WithField("component", l.component).Warnf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(args ...interface{}) {
	l.WithField("component", l.component).Error(args...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.WithField("component", l.component).Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(args ...interface{}) {
	l.WithField("component", l.component).Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.WithField("component", l.component).Fatalf(format, args...)
}

// SetFormatter sets the log formatter
func (l *Logger) SetFormatter(formatter logrus.Formatter) {
	l.Logger.SetFormatter(formatter)
}

// Package-level convenience functions using the default logger

// SetLevel sets the default logger level
func SetLevel(level Level) {
	defaultLogger.SetLevel(level)
}

// SetOutput sets the default logger output
func SetOutput(out io.Writer) {
	defaultLogger.SetOutput(out)
}

// Debug logs a debug message using the default logger
func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

// Debugf logs a formatted debug message using the default logger
func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

// Info logs an info message using the default logger
func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

// Infof logs a formatted info message using the default logger
func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

// Warn logs a warning message using the default logger
func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

// Warnf logs a formatted warning message using the default logger
func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

// Error logs an error message using the default logger
func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

// Errorf logs a formatted error message using the default logger
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// Fatal logs a fatal message using the default logger and exits
func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

// Fatalf logs a formatted fatal message using the default logger and exits
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

// WithField adds a field to the default logger
func WithField(key string, value interface{}) *logrus.Entry {
	return defaultLogger.WithField(key, value)
}

// WithFields adds fields to the default logger
func WithFields(fields map[string]interface{}) *logrus.Entry {
	return defaultLogger.WithFields(fields)
}

// WithError adds an error to the default logger
func WithError(err error) *logrus.Entry {
	return defaultLogger.WithError(err)
}
