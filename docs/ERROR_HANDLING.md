# Error Handling Guide

This document describes the error handling system in containr and best practices for handling errors effectively.

## Overview

Containr implements a comprehensive error handling system with:

- **Error codes**: Unique identifiers for different error types
- **Context-rich errors**: Errors include relevant context information
- **User-friendly messages**: Clear error messages with actionable hints
- **Error wrapping**: Preserve error chains for debugging
- **Integration with logging**: Automatic error logging

## Quick Start

### Creating Errors

```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/errors"

// Create a new error
err := errors.New(errors.ErrContainerCreate, "failed to create container")

// Wrap an existing error
if err := os.Open("/path"); err != nil {
    return errors.Wrap(errors.ErrRootFSNotFound, "cannot open rootfs", err)
}
```

### Adding Context

```go
// Add hint for users
err := errors.New(errors.ErrPermissionDenied, "cannot create namespace").
    WithHint("Try running with sudo or as root user")

// Add contextual fields
err := errors.New(errors.ErrContainerStart, "container failed").
    WithField("container_id", containerID).
    WithField("exit_code", exitCode)
```

### Checking Error Types

```go
// Check if error has specific code
if errors.IsErrorCode(err, errors.ErrPermissionDenied) {
    // Handle permission error
}

// Get error code
code := errors.GetErrorCode(err)
switch code {
case errors.ErrContainerNotFound:
    // Handle not found
case errors.ErrContainerStart:
    // Handle start failure
}
```

## Error Codes

Containr defines error codes for different failure scenarios:

### Container Errors

- `ErrContainerCreate`: Failed to create container
- `ErrContainerStart`: Failed to start container
- `ErrContainerStop`: Failed to stop container
- `ErrContainerNotFound`: Container not found
- `ErrContainerAlreadyExists`: Container already exists

### Namespace Errors

- `ErrNamespaceCreate`: Failed to create namespaces
- `ErrNamespaceSetup`: Failed to setup namespace

### Cgroup Errors

- `ErrCgroupCreate`: Failed to create cgroup
- `ErrCgroupApplyLimits`: Failed to apply resource limits
- `ErrCgroupAddProcess`: Failed to add process to cgroup
- `ErrCgroupRemove`: Failed to remove cgroup
- `ErrCgroupNotFound`: Cgroup not found

### RootFS Errors

- `ErrRootFSNotFound`: Root filesystem not found
- `ErrRootFSMount`: Failed to mount filesystem
- `ErrRootFSUnmount`: Failed to unmount filesystem
- `ErrRootFSPivot`: Failed to pivot root

### Network Errors

- `ErrNetworkCreate`: Failed to create network
- `ErrNetworkSetup`: Failed to setup network
- `ErrNetworkCleanup`: Failed to cleanup network
- `ErrNetworkNotFound`: Network not found

### Image Errors

- `ErrImageNotFound`: Image not found
- `ErrImageImport`: Failed to import image
- `ErrImageExport`: Failed to export image

### Security Errors

- `ErrSecurityCapabilities`: Failed to apply capabilities
- `ErrSecuritySeccomp`: Failed to apply seccomp profile
- `ErrSecurityLSM`: Failed to apply LSM configuration

### Generic Errors

- `ErrInvalidConfig`: Invalid configuration
- `ErrInvalidArgument`: Invalid argument
- `ErrPermissionDenied`: Permission denied
- `ErrNotImplemented`: Feature not implemented
- `ErrInternal`: Internal error

## Error Structure

### ContainrError

```go
type ContainrError struct {
    Code    ErrorCode              // Unique error code
    Message string                 // Human-readable message
    Cause   error                  // Underlying error (if any)
    Hint    string                 // Hint to help resolve the error
    Fields  map[string]interface{} // Contextual information
}
```

### Error Methods

```go
// Error returns the error string
func (e *ContainrError) Error() string

// Unwrap returns the underlying cause
func (e *ContainrError) Unwrap() error

// WithHint adds a hint
func (e *ContainrError) WithHint(hint string) *ContainrError

// WithField adds context
func (e *ContainrError) WithField(key string, value interface{}) *ContainrError

// GetFullMessage returns error with hint
func (e *ContainrError) GetFullMessage() string
```

## Best Practices

### 1. Use Appropriate Error Codes

```go
// Good - specific error code
return errors.New(errors.ErrCgroupCreate, "failed to create cgroup")

// Bad - generic error code
return errors.New(errors.ErrInternal, "failed to create cgroup")
```

### 2. Wrap Errors to Preserve Context

```go
// Good - preserves error chain
if err := syscall.Mount(src, dst, "bind", 0, ""); err != nil {
    return errors.Wrap(errors.ErrRootFSMount, "failed to mount rootfs", err)
}

// Bad - loses original error
if err := syscall.Mount(src, dst, "bind", 0, ""); err != nil {
    return errors.New(errors.ErrRootFSMount, "failed to mount rootfs")
}
```

### 3. Add Helpful Hints

```go
// Good - actionable hint
err := errors.New(errors.ErrPermissionDenied, "cannot create namespace").
    WithHint("Try running with sudo or as root user")

// Better - specific hint
err := errors.New(errors.ErrRootFSMount, "failed to mount rootfs").
    WithHint("Ensure you have CAP_SYS_ADMIN capability and the path is accessible")
```

### 4. Include Relevant Context

```go
// Good - includes context
return errors.Wrap(errors.ErrContainerStart, "container failed", err).
    WithField("container_id", c.ID).
    WithField("command", c.Command).
    WithField("exit_code", exitCode)
```

### 5. Clean Up on Errors

```go
func (c *Container) Setup() error {
    // Create resources
    if err := c.createCgroup(); err != nil {
        return err
    }

    if err := c.setupNetwork(); err != nil {
        // Clean up cgroup on error
        c.cleanupCgroup()
        return err
    }

    return nil
}
```

## Error Handling Patterns

### Pattern 1: Simple Error Creation

```go
func validateConfig(config *Config) error {
    if config.Memory < 0 {
        return errors.New(errors.ErrInvalidConfig, "memory limit cannot be negative").
            WithField("memory", config.Memory).
            WithHint("Set memory to a positive value or 0 for no limit")
    }
    return nil
}
```

### Pattern 2: Error Wrapping

```go
func (c *Container) startProcess() error {
    cmd := exec.Command(c.Command[0], c.Command[1:]...)
    if err := cmd.Start(); err != nil {
        return errors.Wrap(errors.ErrContainerStart, "failed to start process", err).
            WithField("container_id", c.ID).
            WithField("command", c.Command).
            WithHint("Ensure the command exists and you have sufficient privileges")
    }
    return nil
}
```

### Pattern 3: Multiple Error Handling

```go
func (c *Container) Cleanup() error {
    var cleanupErrors []error

    // Attempt all cleanups even if some fail
    if err := c.cleanupNetwork(); err != nil {
        cleanupErrors = append(cleanupErrors, err)
    }

    if err := c.cleanupCgroup(); err != nil {
        cleanupErrors = append(cleanupErrors, err)
    }

    if err := c.unmountRootFS(); err != nil {
        cleanupErrors = append(cleanupErrors, err)
    }

    if len(cleanupErrors) > 0 {
        return errors.New(errors.ErrInternal,
            fmt.Sprintf("cleanup encountered %d errors", len(cleanupErrors))).
            WithField("container_id", c.ID)
    }

    return nil
}
```

### Pattern 4: Error Code Checking

```go
func handleError(err error) {
    if errors.IsErrorCode(err, errors.ErrPermissionDenied) {
        fmt.Println("Please run with elevated privileges")
        os.Exit(1)
    }

    if errors.IsErrorCode(err, errors.ErrContainerNotFound) {
        fmt.Println("Container does not exist")
        os.Exit(1)
    }

    // Generic error handling
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

## Error Helpers

### Common Error Constructors

```go
// Not found error
err := errors.ErrNotFound("container")
// Returns: [CONTAINER_NOT_FOUND] container not found

// Invalid config error
err := errors.ErrInvalidConfigError("invalid memory limit")
// Returns: [INVALID_CONFIG] invalid memory limit
// Hint: Please check your configuration and try again

// Permission error
err := errors.ErrPermission("cannot create namespace")
// Returns: [PERMISSION_DENIED] cannot create namespace
// Hint: Try running with sudo or as root user

// Internal error
err := errors.ErrInternalError("unexpected failure", cause)
// Returns: [INTERNAL] unexpected failure: <cause>
// Hint: This is likely a bug. Please report it...
```

## Integration with Logging

Errors integrate seamlessly with the logging system:

```go
import (
    "github.com/therealutkarshpriyadarshi/containr/pkg/errors"
    "github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

log := logger.New("container")

// Create error with context
err := errors.Wrap(errors.ErrContainerStart, "failed to start", cause).
    WithField("container_id", c.ID).
    WithField("hostname", c.Hostname)

// Log the error with all context
log.WithError(err).
    WithFields(err.Fields).
    Error(err.Message)

// Or use the ContainrError directly
if ce, ok := err.(*errors.ContainrError); ok {
    log.WithFields(map[string]interface{}{
        "code": ce.Code,
        "fields": ce.Fields,
    }).Error(ce.Message)
}
```

## CLI Error Display

The CLI displays errors with full context:

```bash
$ sudo containr run /nonexistent
Error: [CONTAINER_START] failed to start container process: exec: "/nonexistent": stat /nonexistent: no such file or directory
Hint: Ensure the command exists and you have sufficient privileges
```

## User-Facing Error Messages

### Guidelines

1. **Be specific**: Describe what failed
2. **Be actionable**: Include hints on how to resolve
3. **Be concise**: Keep messages short and clear
4. **Be consistent**: Use similar phrasing for similar errors

### Examples

**Good**:
```
[PERMISSION_DENIED] cannot create namespace
Hint: Try running with sudo or as root user
```

**Bad**:
```
Error: operation failed
```

**Good**:
```
[ROOTFS_NOT_FOUND] root filesystem does not exist: /path/to/rootfs
Hint: Ensure the root filesystem path is correct and accessible
```

**Bad**:
```
Error: /path/to/rootfs
```

## Testing Error Handling

### Unit Tests

```go
func TestContainerStartError(t *testing.T) {
    // Test error creation
    err := errors.New(errors.ErrContainerStart, "test error")
    if err.Code != errors.ErrContainerStart {
        t.Errorf("Expected code %s, got %s", errors.ErrContainerStart, err.Code)
    }

    // Test error with context
    err = err.WithField("container_id", "test123")
    if err.Fields["container_id"] != "test123" {
        t.Error("Expected field to be set")
    }

    // Test error checking
    if !errors.IsErrorCode(err, errors.ErrContainerStart) {
        t.Error("Expected error code to match")
    }
}
```

### Integration Tests

```go
func TestErrorPropagation(t *testing.T) {
    // Simulate error condition
    err := container.Start()

    // Verify error code
    if !errors.IsErrorCode(err, errors.ErrContainerStart) {
        t.Errorf("Expected ErrContainerStart, got %v", err)
    }

    // Verify hint is present
    if ce, ok := err.(*errors.ContainrError); ok {
        if ce.Hint == "" {
            t.Error("Expected hint to be set")
        }
    }
}
```

## Debugging Errors

### Enable Debug Logging

```bash
sudo containr run --debug /bin/sh
```

This shows detailed logs leading up to errors.

### Inspect Error Details

```go
if ce, ok := err.(*errors.ContainrError); ok {
    fmt.Printf("Code: %s\n", ce.Code)
    fmt.Printf("Message: %s\n", ce.Message)
    fmt.Printf("Hint: %s\n", ce.Hint)
    fmt.Printf("Fields: %+v\n", ce.Fields)
    fmt.Printf("Cause: %v\n", ce.Cause)
}
```

### Unwrap Error Chain

```go
import "errors"

err := someOperation()

// Unwrap to see full chain
for err != nil {
    fmt.Printf("Error: %v\n", err)
    err = errors.Unwrap(err)
}
```

## Migration Guide

### From Standard Errors

**Before**:
```go
if err := operation(); err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

**After**:
```go
if err := operation(); err != nil {
    return errors.Wrap(errors.ErrContainerStart, "operation failed", err).
        WithField("container_id", c.ID).
        WithHint("Ensure prerequisites are met")
}
```

### From Error Strings

**Before**:
```go
if config.Invalid() {
    return fmt.Errorf("invalid config")
}
```

**After**:
```go
if config.Invalid() {
    return errors.New(errors.ErrInvalidConfig, "invalid memory limit").
        WithField("memory", config.Memory).
        WithHint("Set memory to a positive value")
}
```

## Error Handling Checklist

When adding error handling:

- [ ] Use appropriate error code
- [ ] Wrap errors to preserve context
- [ ] Add helpful hints for users
- [ ] Include relevant fields
- [ ] Log errors appropriately
- [ ] Clean up resources on errors
- [ ] Test error paths
- [ ] Document error conditions

## Examples

See [examples/error_handling.go](../examples/error_handling.go) for complete examples.

## Related Documentation

- [Logging Guide](LOGGING.md)
- [Architecture Documentation](ARCHITECTURE.md)
- [Contributing Guide](../CONTRIBUTING.md)
