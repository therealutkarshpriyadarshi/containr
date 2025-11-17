package logger

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNew(t *testing.T) {
	log := New("test-component")

	if log == nil {
		t.Fatal("Expected logger to be created, got nil")
	}

	if log.component != "test-component" {
		t.Errorf("Expected component to be 'test-component', got '%s'", log.component)
	}
}

func TestGetLogger(t *testing.T) {
	log := GetLogger()

	if log == nil {
		t.Fatal("Expected default logger to be returned, got nil")
	}

	if log.component != "containr" {
		t.Errorf("Expected default component to be 'containr', got '%s'", log.component)
	}
}

func TestSetLevel(t *testing.T) {
	log := New("test")
	var buf bytes.Buffer
	log.SetOutput(&buf)

	tests := []struct {
		name          string
		level         Level
		expectedLevel logrus.Level
	}{
		{"Debug Level", DebugLevel, logrus.DebugLevel},
		{"Info Level", InfoLevel, logrus.InfoLevel},
		{"Warn Level", WarnLevel, logrus.WarnLevel},
		{"Error Level", ErrorLevel, logrus.ErrorLevel},
		{"Fatal Level", FatalLevel, logrus.FatalLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.SetLevel(tt.level)
			if log.Logger.Level != tt.expectedLevel {
				t.Errorf("Expected log level %v, got %v", tt.expectedLevel, log.Logger.Level)
			}
		})
	}
}

func TestWithField(t *testing.T) {
	log := New("test")
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(InfoLevel)

	entry := log.WithField("key", "value")

	if entry.Data["key"] != "value" {
		t.Errorf("Expected field 'key' to be 'value', got '%v'", entry.Data["key"])
	}

	if entry.Data["component"] != "test" {
		t.Errorf("Expected component field to be 'test', got '%v'", entry.Data["component"])
	}
}

func TestWithFields(t *testing.T) {
	log := New("test")
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(InfoLevel)

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	entry := log.WithFields(fields)

	if entry.Data["key1"] != "value1" {
		t.Errorf("Expected field 'key1' to be 'value1', got '%v'", entry.Data["key1"])
	}

	if entry.Data["key2"] != 123 {
		t.Errorf("Expected field 'key2' to be 123, got '%v'", entry.Data["key2"])
	}

	if entry.Data["component"] != "test" {
		t.Errorf("Expected component field to be 'test', got '%v'", entry.Data["component"])
	}
}

func TestDebugLogging(t *testing.T) {
	log := New("test")
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(DebugLevel)

	log.Debug("debug message")

	output := buf.String()
	if output == "" {
		t.Error("Expected debug message to be logged")
	}
}

func TestInfoLogging(t *testing.T) {
	log := New("test")
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(InfoLevel)

	log.Info("info message")

	output := buf.String()
	if output == "" {
		t.Error("Expected info message to be logged")
	}
}

func TestWarnLogging(t *testing.T) {
	log := New("test")
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(WarnLevel)

	log.Warn("warn message")

	output := buf.String()
	if output == "" {
		t.Error("Expected warn message to be logged")
	}
}

func TestErrorLogging(t *testing.T) {
	log := New("test")
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(ErrorLevel)

	log.Error("error message")

	output := buf.String()
	if output == "" {
		t.Error("Expected error message to be logged")
	}
}

func TestLogLevelFiltering(t *testing.T) {
	log := New("test")
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetLevel(InfoLevel)

	// Debug should not be logged
	log.Debug("debug message")
	if buf.Len() > 0 {
		t.Error("Expected debug message to be filtered out at Info level")
	}

	// Info should be logged
	buf.Reset()
	log.Info("info message")
	if buf.Len() == 0 {
		t.Error("Expected info message to be logged at Info level")
	}
}

func TestPackageLevelFunctions(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(InfoLevel)

	Info("test message")

	output := buf.String()
	if output == "" {
		t.Error("Expected package-level Info to log message")
	}
}
