# Logging Guide

This document describes the structured logging system in containr and how to use it effectively.

## Overview

Containr uses [logrus](https://github.com/sirupsen/logrus) for structured logging, providing:

- **Configurable log levels**: Debug, Info, Warn, Error, Fatal
- **Structured fields**: Context-rich logging with key-value pairs
- **Component-based logging**: Each component has its own logger
- **Debug mode**: Verbose output for troubleshooting
- **Flexible output**: Log to stdout, stderr, or files

## Quick Start

### Using the Logger in Code

```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/logger"

// Create a component-specific logger
log := logger.New("mycomponent")

// Log at different levels
log.Debug("Detailed debug information")
log.Info("General information")
log.Warn("Warning message")
log.Error("Error occurred")

// Add structured fields
log.WithField("container_id", "abc123").Info("Container started")

// Add multiple fields
log.WithFields(map[string]interface{}{
    "container_id": "abc123",
    "hostname": "myhost",
}).Info("Container configured")

// Log with error
if err != nil {
    log.WithError(err).Error("Operation failed")
}
```

### Using the CLI

```bash
# Normal operation (info level)
sudo containr run /bin/sh

# Enable debug mode
sudo containr run --debug /bin/sh

# Set specific log level
sudo containr run --log-level debug /bin/sh

# Log to stderr instead of stdout
sudo containr run --log-stderr /bin/sh
```

## Log Levels

Containr supports five log levels, from most to least verbose:

### 1. Debug Level

**When to use**: Development and troubleshooting

```go
log.Debug("Entering function")
log.Debugf("Processing %d items", count)
```

**CLI**: `--debug` or `--log-level debug`

**Output includes**:
- Function entry/exit points
- Variable values
- Detailed operation steps
- All lower-level logs (Info, Warn, Error)

### 2. Info Level (Default)

**When to use**: Normal operation, important events

```go
log.Info("Container started successfully")
log.Infof("Listening on port %d", port)
```

**CLI**: Default level, or `--log-level info`

**Output includes**:
- Container lifecycle events
- Configuration changes
- Successful operations
- All lower-level logs (Warn, Error)

### 3. Warn Level

**When to use**: Potential issues, recoverable errors

```go
log.Warn("Resource limit not set, using default")
log.Warnf("Failed to remove %s, continuing", path)
```

**CLI**: `--log-level warn`

**Output includes**:
- Non-critical errors
- Degraded functionality
- All lower-level logs (Error)

### 4. Error Level

**When to use**: Errors that prevent operation

```go
log.Error("Failed to start container")
log.Errorf("Invalid configuration: %s", reason)
```

**CLI**: `--log-level error`

**Output includes**:
- Operation failures
- Critical errors
- Fatal logs

### 5. Fatal Level

**When to use**: Unrecoverable errors (exits program)

```go
log.Fatal("Cannot initialize system")
```

**Note**: Fatal logs call `os.Exit(1)` after logging

## Structured Logging

### Adding Context with Fields

Fields provide context to log messages without cluttering the message text:

```go
// Single field
log.WithField("container_id", container.ID).Info("Starting container")

// Multiple fields
log.WithFields(map[string]interface{}{
    "container_id": container.ID,
    "hostname": container.Hostname,
    "pid": container.PID,
}).Info("Container running")

// With error
log.WithError(err).WithField("path", filepath).Error("Failed to mount")
```

**Output**:
```
time="2025-11-17 10:30:45" level=info msg="Starting container" component=container container_id=abc123
```

### Component-Based Logging

Each package should create its own logger:

```go
// In pkg/container/container.go
var log = logger.New("container")

// In pkg/cgroup/cgroup.go
var log = logger.New("cgroup")
```

This helps identify which component generated each log message.

## Best Practices

### 1. Use Appropriate Log Levels

```go
// Good
log.Debug("Parsing configuration file")
log.Info("Container started successfully")
log.Warn("Using default value for unspecified option")
log.Error("Failed to create namespace")

// Bad
log.Info("i = 42")  // Too detailed for Info
log.Error("Container started")  // Wrong level
```

### 2. Add Context with Fields

```go
// Good
log.WithFields(map[string]interface{}{
    "container_id": c.ID,
    "error_code": errCode,
}).Error("Container failed")

// Bad
log.Errorf("Container %s failed with code %d", c.ID, errCode)
```

### 3. Log at Decision Points

```go
// Log when entering important functions
log.WithField("container_id", c.ID).Debug("Starting container execution")

// Log before critical operations
log.WithField("path", rootfs).Debug("Mounting root filesystem")

// Log results
log.WithField("container_id", c.ID).Info("Container completed successfully")
```

### 4. Use Consistent Field Names

Common field names used in containr:

- `container_id`: Container identifier
- `cgroup_name`: Cgroup name
- `pid`: Process ID
- `path`: File or directory path
- `hostname`: Hostname
- `error_code`: Error code from errors package

### 5. Don't Log Sensitive Information

```go
// Bad - logs password
log.WithField("password", password).Debug("Authenticating")

// Good
log.Debug("Authenticating user")
```

## Debug Mode

Debug mode enables the most verbose logging for troubleshooting:

```bash
# CLI
sudo containr run --debug /bin/sh
```

Debug mode logs:
- All namespace operations
- Mount operations and paths
- Cgroup creation and limits
- Network setup steps
- Security policy application
- Process execution details

**Example debug output**:
```
time="2025-11-17 10:30:45" level=debug msg="Starting container execution" component=container container_id=container-12345
time="2025-11-17 10:30:45" level=debug msg="Namespace flags: 1426063360" component=container container_id=container-12345
time="2025-11-17 10:30:45" level=debug msg="Starting container process" component=container container_id=container-12345
time="2025-11-17 10:30:45" level=debug msg="Container process started with PID: 12346" component=container container_id=container-12345
```

## Logging Patterns

### Container Lifecycle

```go
// Start
log.WithField("container_id", c.ID).Info("Starting container")

// Progress
log.WithField("container_id", c.ID).Debug("Applying security policies")
log.WithField("container_id", c.ID).Debug("Setting up network")

// Completion
log.WithField("container_id", c.ID).Info("Container started successfully")

// Error
if err != nil {
    log.WithError(err).WithField("container_id", c.ID).Error("Container start failed")
}
```

### Resource Operations

```go
// Create
log.WithField("cgroup_name", name).Info("Creating cgroup")

// Configure
log.WithFields(map[string]interface{}{
    "cgroup_name": name,
    "memory_limit": limit,
}).Debug("Applying memory limit")

// Cleanup
log.WithField("cgroup_name", name).Debug("Removing cgroup")

// Error handling
if err != nil {
    log.WithError(err).WithField("cgroup_name", name).Warn("Failed to remove cgroup")
}
```

## Programmatic Configuration

### Set Log Level

```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/logger"

// Get default logger
log := logger.GetLogger()

// Set level
log.SetLevel(logger.DebugLevel)
```

### Change Output Destination

```go
import (
    "os"
    "github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

log := logger.GetLogger()

// Log to stderr
log.SetOutput(os.Stderr)

// Log to file
file, err := os.OpenFile("containr.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
if err == nil {
    log.SetOutput(file)
}
```

### Custom Formatter

```go
import (
    "github.com/sirupsen/logrus"
    "github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

log := logger.GetLogger()

// JSON formatter
log.SetFormatter(&logrus.JSONFormatter{})

// Custom text formatter
log.SetFormatter(&logrus.TextFormatter{
    FullTimestamp: true,
    DisableColors: true,
})
```

## Troubleshooting

### No Logs Appearing

Check the log level:
```bash
# Try debug mode
sudo containr run --debug /bin/sh
```

### Too Verbose

Reduce log level:
```bash
# Only show errors
sudo containr run --log-level error /bin/sh
```

### Logs Mixed with Application Output

Separate logging:
```bash
# Send logs to stderr, keep stdout for application
sudo containr run --log-stderr /bin/sh
```

## Integration with Error Handling

Logging works seamlessly with the error package:

```go
import (
    "github.com/therealutkarshpriyadarshi/containr/pkg/errors"
    "github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

log := logger.New("container")

if err := someOperation(); err != nil {
    // Log the error
    log.WithError(err).Error("Operation failed")

    // Return wrapped error with code
    return errors.Wrap(errors.ErrContainerStart, "operation failed", err)
}
```

See [ERROR_HANDLING.md](ERROR_HANDLING.md) for more details on error handling.

## Examples

### Example 1: Container Startup

```go
func (c *Container) Start() error {
    log.WithField("container_id", c.ID).Info("Starting container")

    log.WithField("container_id", c.ID).Debug("Creating namespaces")
    if err := c.createNamespaces(); err != nil {
        log.WithError(err).Error("Failed to create namespaces")
        return err
    }

    log.WithField("container_id", c.ID).Debug("Setting up cgroups")
    if err := c.setupCgroups(); err != nil {
        log.WithError(err).Error("Failed to setup cgroups")
        return err
    }

    log.WithField("container_id", c.ID).Info("Container started successfully")
    return nil
}
```

### Example 2: Resource Management

```go
func (c *Cgroup) ApplyLimits(config *Config) error {
    log.WithFields(map[string]interface{}{
        "cgroup_name": c.Name,
        "memory_limit": config.MemoryLimit,
        "cpu_shares": config.CPUShares,
    }).Debug("Applying resource limits")

    if config.MemoryLimit > 0 {
        log.WithField("cgroup_name", c.Name).
            Debugf("Setting memory limit: %d bytes", config.MemoryLimit)
        if err := c.setMemoryLimit(config.MemoryLimit); err != nil {
            log.WithError(err).Error("Failed to set memory limit")
            return err
        }
    }

    log.WithField("cgroup_name", c.Name).Info("Resource limits applied")
    return nil
}
```

## Related Documentation

- [Error Handling Guide](ERROR_HANDLING.md)
- [Architecture Documentation](ARCHITECTURE.md)
- [Security Guide](SECURITY.md)
