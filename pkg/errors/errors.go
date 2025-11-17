package errors

import (
	"fmt"
)

// ErrorCode represents a unique error code
type ErrorCode string

const (
	// Container errors
	ErrContainerCreate     ErrorCode = "CONTAINER_CREATE"
	ErrContainerStart      ErrorCode = "CONTAINER_START"
	ErrContainerStop       ErrorCode = "CONTAINER_STOP"
	ErrContainerNotFound   ErrorCode = "CONTAINER_NOT_FOUND"
	ErrContainerAlreadyExists ErrorCode = "CONTAINER_ALREADY_EXISTS"

	// Namespace errors
	ErrNamespaceCreate     ErrorCode = "NAMESPACE_CREATE"
	ErrNamespaceSetup      ErrorCode = "NAMESPACE_SETUP"

	// Cgroup errors
	ErrCgroupCreate        ErrorCode = "CGROUP_CREATE"
	ErrCgroupApplyLimits   ErrorCode = "CGROUP_APPLY_LIMITS"
	ErrCgroupAddProcess    ErrorCode = "CGROUP_ADD_PROCESS"
	ErrCgroupRemove        ErrorCode = "CGROUP_REMOVE"
	ErrCgroupNotFound      ErrorCode = "CGROUP_NOT_FOUND"

	// RootFS errors
	ErrRootFSNotFound      ErrorCode = "ROOTFS_NOT_FOUND"
	ErrRootFSMount         ErrorCode = "ROOTFS_MOUNT"
	ErrRootFSUnmount       ErrorCode = "ROOTFS_UNMOUNT"
	ErrRootFSPivot         ErrorCode = "ROOTFS_PIVOT"

	// Network errors
	ErrNetworkCreate       ErrorCode = "NETWORK_CREATE"
	ErrNetworkSetup        ErrorCode = "NETWORK_SETUP"
	ErrNetworkCleanup      ErrorCode = "NETWORK_CLEANUP"
	ErrNetworkNotFound     ErrorCode = "NETWORK_NOT_FOUND"

	// Image errors
	ErrImageNotFound       ErrorCode = "IMAGE_NOT_FOUND"
	ErrImageImport         ErrorCode = "IMAGE_IMPORT"
	ErrImageExport         ErrorCode = "IMAGE_EXPORT"

	// Security errors
	ErrSecurityCapabilities ErrorCode = "SECURITY_CAPABILITIES"
	ErrSecuritySeccomp     ErrorCode = "SECURITY_SECCOMP"
	ErrSecurityLSM         ErrorCode = "SECURITY_LSM"

	// Generic errors
	ErrInvalidConfig       ErrorCode = "INVALID_CONFIG"
	ErrInvalidArgument     ErrorCode = "INVALID_ARGUMENT"
	ErrPermissionDenied    ErrorCode = "PERMISSION_DENIED"
	ErrNotImplemented      ErrorCode = "NOT_IMPLEMENTED"
	ErrInternal            ErrorCode = "INTERNAL"
)

// ContainrError is a custom error type with error code and context
type ContainrError struct {
	Code    ErrorCode
	Message string
	Cause   error
	Hint    string
	Fields  map[string]interface{}
}

// Error implements the error interface
func (e *ContainrError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *ContainrError) Unwrap() error {
	return e.Cause
}

// WithHint adds a hint to help users resolve the error
func (e *ContainrError) WithHint(hint string) *ContainrError {
	e.Hint = hint
	return e
}

// WithField adds a context field to the error
func (e *ContainrError) WithField(key string, value interface{}) *ContainrError {
	if e.Fields == nil {
		e.Fields = make(map[string]interface{})
	}
	e.Fields[key] = value
	return e
}

// GetFullMessage returns the full error message with hint
func (e *ContainrError) GetFullMessage() string {
	msg := e.Error()
	if e.Hint != "" {
		msg += fmt.Sprintf("\nHint: %s", e.Hint)
	}
	return msg
}

// New creates a new ContainrError
func New(code ErrorCode, message string) *ContainrError {
	return &ContainrError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an existing error with a ContainrError
func Wrap(code ErrorCode, message string, cause error) *ContainrError {
	return &ContainrError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// IsErrorCode checks if an error has a specific error code
func IsErrorCode(err error, code ErrorCode) bool {
	if err == nil {
		return false
	}
	if ce, ok := err.(*ContainrError); ok {
		return ce.Code == code
	}
	return false
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) ErrorCode {
	if err == nil {
		return ""
	}
	if ce, ok := err.(*ContainrError); ok {
		return ce.Code
	}
	return ErrInternal
}

// Common error constructors for convenience

// ErrNotFound creates a not found error
func ErrNotFound(resource string) *ContainrError {
	return New(ErrContainerNotFound, fmt.Sprintf("%s not found", resource))
}

// ErrInvalidConfigError creates an invalid config error
func ErrInvalidConfigError(message string) *ContainrError {
	return New(ErrInvalidConfig, message).WithHint("Please check your configuration and try again")
}

// ErrPermission creates a permission denied error
func ErrPermission(message string) *ContainrError {
	return New(ErrPermissionDenied, message).WithHint("Try running with sudo or as root user")
}

// ErrInternalError creates an internal error
func ErrInternalError(message string, cause error) *ContainrError {
	return Wrap(ErrInternal, message, cause).WithHint("This is likely a bug. Please report it at https://github.com/therealutkarshpriyadarshi/containr/issues")
}
