package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(ErrContainerCreate, "test error message")

	if err == nil {
		t.Fatal("Expected error to be created, got nil")
	}

	if err.Code != ErrContainerCreate {
		t.Errorf("Expected error code %s, got %s", ErrContainerCreate, err.Code)
	}

	if err.Message != "test error message" {
		t.Errorf("Expected message 'test error message', got '%s'", err.Message)
	}
}

func TestWrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := Wrap(ErrContainerStart, "wrapper message", cause)

	if err == nil {
		t.Fatal("Expected error to be created, got nil")
	}

	if err.Code != ErrContainerStart {
		t.Errorf("Expected error code %s, got %s", ErrContainerStart, err.Code)
	}

	if err.Message != "wrapper message" {
		t.Errorf("Expected message 'wrapper message', got '%s'", err.Message)
	}

	if err.Cause != cause {
		t.Error("Expected cause to be set")
	}
}

func TestErrorString(t *testing.T) {
	tests := []struct {
		name     string
		err      *ContainrError
		expected string
	}{
		{
			name:     "Error without cause",
			err:      New(ErrContainerCreate, "test error"),
			expected: "[CONTAINER_CREATE] test error",
		},
		{
			name:     "Error with cause",
			err:      Wrap(ErrContainerStart, "wrapper", errors.New("cause")),
			expected: "[CONTAINER_START] wrapper: cause",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Expected error string '%s', got '%s'", tt.expected, tt.err.Error())
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := Wrap(ErrContainerStart, "wrapper message", cause)

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Error("Expected Unwrap to return the cause")
	}
}

func TestWithHint(t *testing.T) {
	err := New(ErrPermissionDenied, "permission denied").
		WithHint("Try running with sudo")

	if err.Hint != "Try running with sudo" {
		t.Errorf("Expected hint 'Try running with sudo', got '%s'", err.Hint)
	}

	fullMsg := err.GetFullMessage()
	if !strings.Contains(fullMsg, "Hint: Try running with sudo") {
		t.Errorf("Expected full message to contain hint, got '%s'", fullMsg)
	}
}

func TestWithField(t *testing.T) {
	err := New(ErrContainerCreate, "test error").
		WithField("container_id", "abc123")

	if err.Fields == nil {
		t.Fatal("Expected fields map to be initialized")
	}

	if err.Fields["container_id"] != "abc123" {
		t.Errorf("Expected field 'container_id' to be 'abc123', got '%v'", err.Fields["container_id"])
	}
}

func TestWithMultipleFields(t *testing.T) {
	err := New(ErrContainerCreate, "test error").
		WithField("key1", "value1").
		WithField("key2", 123)

	if len(err.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(err.Fields))
	}

	if err.Fields["key1"] != "value1" {
		t.Errorf("Expected field 'key1' to be 'value1', got '%v'", err.Fields["key1"])
	}

	if err.Fields["key2"] != 123 {
		t.Errorf("Expected field 'key2' to be 123, got '%v'", err.Fields["key2"])
	}
}

func TestIsErrorCode(t *testing.T) {
	err := New(ErrContainerCreate, "test error")

	if !IsErrorCode(err, ErrContainerCreate) {
		t.Error("Expected IsErrorCode to return true for matching code")
	}

	if IsErrorCode(err, ErrContainerStart) {
		t.Error("Expected IsErrorCode to return false for non-matching code")
	}

	if IsErrorCode(nil, ErrContainerCreate) {
		t.Error("Expected IsErrorCode to return false for nil error")
	}

	stdErr := errors.New("standard error")
	if IsErrorCode(stdErr, ErrContainerCreate) {
		t.Error("Expected IsErrorCode to return false for standard error")
	}
}

func TestGetErrorCode(t *testing.T) {
	err := New(ErrContainerCreate, "test error")

	code := GetErrorCode(err)
	if code != ErrContainerCreate {
		t.Errorf("Expected error code %s, got %s", ErrContainerCreate, code)
	}

	nilCode := GetErrorCode(nil)
	if nilCode != "" {
		t.Errorf("Expected empty code for nil error, got %s", nilCode)
	}

	stdErr := errors.New("standard error")
	stdCode := GetErrorCode(stdErr)
	if stdCode != ErrInternal {
		t.Errorf("Expected ErrInternal for standard error, got %s", stdCode)
	}
}

func TestErrNotFound(t *testing.T) {
	err := ErrNotFound("container")

	if err.Code != ErrContainerNotFound {
		t.Errorf("Expected error code %s, got %s", ErrContainerNotFound, err.Code)
	}

	if !strings.Contains(err.Message, "container not found") {
		t.Errorf("Expected message to contain 'container not found', got '%s'", err.Message)
	}
}

func TestErrInvalidConfigError(t *testing.T) {
	err := ErrInvalidConfigError("invalid memory limit")

	if err.Code != ErrInvalidConfig {
		t.Errorf("Expected error code %s, got %s", ErrInvalidConfig, err.Code)
	}

	if err.Message != "invalid memory limit" {
		t.Errorf("Expected message 'invalid memory limit', got '%s'", err.Message)
	}

	if err.Hint == "" {
		t.Error("Expected hint to be set")
	}
}

func TestErrPermission(t *testing.T) {
	err := ErrPermission("cannot create namespace")

	if err.Code != ErrPermissionDenied {
		t.Errorf("Expected error code %s, got %s", ErrPermissionDenied, err.Code)
	}

	if !strings.Contains(err.Hint, "sudo") {
		t.Errorf("Expected hint to mention 'sudo', got '%s'", err.Hint)
	}
}

func TestErrInternalError(t *testing.T) {
	cause := errors.New("internal failure")
	err := ErrInternalError("unexpected error", cause)

	if err.Code != ErrInternal {
		t.Errorf("Expected error code %s, got %s", ErrInternal, err.Code)
	}

	if err.Cause != cause {
		t.Error("Expected cause to be set")
	}

	if !strings.Contains(err.Hint, "bug") {
		t.Errorf("Expected hint to mention 'bug', got '%s'", err.Hint)
	}
}

func TestErrorCodes(t *testing.T) {
	codes := []ErrorCode{
		ErrContainerCreate,
		ErrContainerStart,
		ErrContainerStop,
		ErrNamespaceCreate,
		ErrCgroupCreate,
		ErrRootFSNotFound,
		ErrNetworkCreate,
		ErrImageNotFound,
		ErrSecurityCapabilities,
		ErrInvalidConfig,
	}

	for _, code := range codes {
		if code == "" {
			t.Errorf("Error code should not be empty")
		}
	}
}

func TestGetFullMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      *ContainrError
		contains []string
	}{
		{
			name:     "Error without hint",
			err:      New(ErrContainerCreate, "test error"),
			contains: []string{"CONTAINER_CREATE", "test error"},
		},
		{
			name:     "Error with hint",
			err:      New(ErrPermissionDenied, "access denied").WithHint("Use sudo"),
			contains: []string{"PERMISSION_DENIED", "access denied", "Hint:", "Use sudo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullMsg := tt.err.GetFullMessage()
			for _, substr := range tt.contains {
				if !strings.Contains(fullMsg, substr) {
					t.Errorf("Expected full message to contain '%s', got '%s'", substr, fullMsg)
				}
			}
		})
	}
}
