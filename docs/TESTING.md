# Testing Guide for Containr

This document provides comprehensive guidance on testing the containr project, including how to run tests locally, write new tests, and understand the testing infrastructure.

## Table of Contents

- [Overview](#overview)
- [Test Structure](#test-structure)
- [Running Tests](#running-tests)
- [Writing Tests](#writing-tests)
- [Code Coverage](#code-coverage)
- [CI/CD Pipeline](#cicd-pipeline)
- [Troubleshooting](#troubleshooting)

## Overview

Containr uses a comprehensive testing strategy that includes:

- **Unit Tests**: Test individual packages and functions in isolation
- **Integration Tests**: Test end-to-end scenarios and interactions between components
- **Static Analysis**: Automated code quality and security checks
- **Coverage Reporting**: Track test coverage metrics

### Testing Goals

- **Coverage Target**: ≥70% code coverage
- **Quality**: All tests must pass in CI
- **Security**: Zero critical vulnerabilities
- **Performance**: Tests should complete in reasonable time

## Test Structure

```
containr/
├── pkg/                      # Source packages
│   ├── namespace/
│   │   ├── namespace.go
│   │   └── namespace_test.go # Unit tests
│   ├── cgroup/
│   │   ├── cgroup.go
│   │   └── cgroup_test.go
│   ├── container/
│   │   ├── container.go
│   │   └── container_test.go
│   ├── rootfs/
│   │   ├── rootfs.go
│   │   └── rootfs_test.go
│   ├── network/
│   │   ├── network.go
│   │   └── network_test.go
│   └── image/
│       ├── image.go
│       └── image_test.go
└── test/                     # Integration tests
    └── integration_test.go
```

### Package Test Coverage

Each package has comprehensive unit tests covering:

#### `pkg/namespace`
- Namespace flag generation and combinations
- Namespace creation and isolation
- Reexec functionality
- Configuration validation

#### `pkg/cgroup`
- Cgroup creation with different limits
- Process addition to cgroups
- Statistics collection
- Cleanup operations
- Both cgroup v1 and v2 support

#### `pkg/container`
- Container creation and configuration
- Namespace setup
- Lifecycle management
- Environment variable passing
- Error handling

#### `pkg/rootfs`
- Filesystem setup and teardown
- Overlay filesystem configuration
- Mount operations
- Device node creation
- Pivot root preparation

#### `pkg/network`
- Network configuration validation
- Bridge and veth pair creation
- IP address validation
- Interface configuration
- NAT and routing setup

#### `pkg/image`
- Image import/export
- Manifest handling
- Layer management
- JSON serialization
- Image naming and tagging

## Running Tests

### Prerequisites

Some tests require specific privileges or system features:

- **Root privileges**: Required for namespace, cgroup, and network tests
- **Linux kernel**: Tests must run on Linux (kernel 3.8+)
- **Cgroup filesystem**: `/sys/fs/cgroup` must be available
- **Network tools**: `ip`, `iptables` for network tests

### Quick Start

Run all tests:
```bash
make test
```

Run only unit tests:
```bash
make test-unit
```

Run only integration tests (requires root):
```bash
make test-integration
```

### Detailed Test Commands

#### Unit Tests

Run unit tests for all packages:
```bash
go test -v ./pkg/...
```

Run tests for a specific package:
```bash
go test -v ./pkg/namespace
go test -v ./pkg/cgroup
go test -v ./pkg/container
```

Run with race detection:
```bash
go test -v -race ./pkg/...
```

Run a specific test:
```bash
go test -v ./pkg/namespace -run TestGetNamespaceFlags
```

#### Integration Tests

Integration tests require root privileges:

```bash
# Run all integration tests
sudo go test -v ./test/...

# Run specific integration test
sudo go test -v ./test/... -run TestContainerRunsSuccessfully
```

#### Coverage Testing

Generate coverage report:
```bash
make test-coverage
```

This will:
1. Run all unit tests with coverage tracking
2. Generate HTML coverage report at `coverage/coverage.html`
3. Display overall coverage percentage
4. Create detailed coverage output

View coverage in terminal:
```bash
go test -cover ./pkg/...
```

Detailed coverage analysis:
```bash
go test -coverprofile=coverage.out ./pkg/...
go tool cover -func=coverage.out
```

Open HTML coverage report:
```bash
go tool cover -html=coverage.out
```

## Writing Tests

### Test File Naming

- Unit tests: `*_test.go` in the same directory as the source
- Integration tests: Place in `test/` directory
- Test functions: Must start with `Test` prefix

### Basic Test Structure

```go
package mypackage

import "testing"

func TestMyFunction(t *testing.T) {
    // Arrange
    input := "test"
    expected := "expected result"

    // Act
    result := MyFunction(input)

    // Assert
    if result != expected {
        t.Errorf("MyFunction(%q) = %q, want %q", input, result, expected)
    }
}
```

### Table-Driven Tests

Use table-driven tests for multiple test cases:

```go
func TestGetNamespaceFlags(t *testing.T) {
    tests := []struct {
        name     string
        types    []NamespaceType
        expected int
    }{
        {
            name:     "Single UTS namespace",
            types:    []NamespaceType{UTS},
            expected: syscall.CLONE_NEWUTS,
        },
        {
            name:     "Multiple namespaces",
            types:    []NamespaceType{UTS, PID},
            expected: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := GetNamespaceFlags(tt.types...)
            if result != tt.expected {
                t.Errorf("got %d, want %d", result, tt.expected)
            }
        })
    }
}
```

### Skipping Tests Conditionally

Skip tests that require specific privileges:

```go
func TestRequiresRoot(t *testing.T) {
    if os.Geteuid() != 0 {
        t.Skip("Skipping test that requires root privileges")
    }

    // Test code here
}
```

Skip tests for missing features:

```go
func TestCgroup(t *testing.T) {
    if _, err := os.Stat("/sys/fs/cgroup"); os.IsNotExist(err) {
        t.Skip("Skipping test: cgroup filesystem not available")
    }

    // Test code here
}
```

### Testing Error Cases

```go
func TestErrorHandling(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        expectError bool
    }{
        {"valid input", "valid", false},
        {"invalid input", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := MyFunction(tt.input)
            if (err != nil) != tt.expectError {
                t.Errorf("error = %v, expectError %v", err, tt.expectError)
            }
        })
    }
}
```

### Using Test Helpers

```go
func TestCleanup(t *testing.T) {
    tmpDir := t.TempDir() // Automatically cleaned up

    // Use tmpDir for test
}
```

### Integration Test Example

```go
func TestEndToEndContainer(t *testing.T) {
    if os.Geteuid() != 0 {
        t.Skip("Requires root")
    }

    // Create container
    config := &container.Config{
        Command:  []string{"/bin/echo", "test"},
        Hostname: "test-container",
    }

    c := container.New(config)
    if c == nil {
        t.Fatal("Failed to create container")
    }

    // Verify container properties
    if c.ID == "" {
        t.Error("Container ID should not be empty")
    }

    t.Logf("Container created: %s", c.ID)
}
```

## Code Coverage

### Coverage Targets

- **Overall**: ≥70% code coverage
- **Critical packages**: ≥80% coverage
  - `pkg/namespace`
  - `pkg/cgroup`
  - `pkg/container`

### Viewing Coverage

Generate and view coverage report:
```bash
# Generate coverage
make test-coverage

# Open HTML report in browser
open coverage/coverage.html  # macOS
xdg-open coverage/coverage.html  # Linux
```

### Coverage Analysis

View coverage by function:
```bash
go tool cover -func=coverage/coverage.out
```

Find uncovered code:
```bash
go tool cover -func=coverage/coverage.out | grep "0.0%"
```

### Improving Coverage

1. Identify untested code:
   ```bash
   go test -coverprofile=coverage.out ./pkg/...
   go tool cover -func=coverage.out | sort -k3 -n
   ```

2. Focus on low-coverage packages first
3. Add tests for error paths
4. Test edge cases and boundary conditions
5. Add integration tests for complex scenarios

## CI/CD Pipeline

### GitHub Actions Workflows

#### CI Workflow (`.github/workflows/ci.yml`)

Runs on every push and pull request:

1. **Test**: Run unit tests on multiple Go versions
2. **Integration Test**: Run integration tests with root privileges
3. **Lint**: Code formatting and linting checks
4. **Static Analysis**: Run staticcheck
5. **Build**: Build binaries for multiple architectures
6. **Coverage Report**: Generate and upload coverage reports
7. **Security Scan**: Run gosec security scanner

#### Release Workflow (`.github/workflows/release.yml`)

Triggers on version tags (`v*`):

1. Build binaries for all architectures
2. Generate checksums
3. Create GitHub release
4. Upload artifacts

### Running CI Checks Locally

Before pushing code, run:

```bash
# Quick pre-commit checks
make pre-commit

# Full CI checks
make ci
```

### CI Environment

Tests run in Ubuntu latest with:
- Go versions: 1.21, 1.22, 1.23
- Root privileges available for integration tests
- All necessary tools installed

## Troubleshooting

### Common Issues

#### "Operation not permitted"

**Problem**: Tests fail with permission errors

**Solution**: Run with root privileges
```bash
sudo go test -v ./pkg/namespace
```

#### "cgroup filesystem not available"

**Problem**: Cgroup tests are skipped

**Solution**:
- Verify cgroups are mounted: `mount | grep cgroup`
- On WSL2: Enable systemd or mount cgroups manually
- Tests will gracefully skip if unavailable

#### "command not found: /bin/echo"

**Problem**: Test commands not found

**Solution**: Ensure basic utilities are installed
```bash
which /bin/echo
which /bin/sh
```

#### Network tests failing

**Problem**: Network isolation tests fail

**Solution**:
- Ensure `CAP_NET_ADMIN` capability
- Check if `ip` and `iptables` tools are installed
- Run with root: `sudo -E go test -v ./pkg/network`

#### Race detector warnings

**Problem**: Data race warnings during tests

**Solution**:
```bash
# Run with race detector to identify issues
go test -race ./pkg/...
```

### Debugging Tests

Enable verbose output:
```bash
go test -v ./pkg/namespace
```

Run specific test with debugging:
```bash
go test -v ./pkg/namespace -run TestGetNamespaceFlags -test.v
```

Print test coverage during run:
```bash
go test -v -cover ./pkg/...
```

Run with additional logging:
```bash
go test -v ./pkg/... -args -test.v -test.parallel=1
```

### Test Timeouts

Set custom timeout for long-running tests:
```bash
go test -timeout 10m ./test/...
```

For integration tests in Makefile:
```bash
make test-integration TIMEOUT=10m
```

## Best Practices

### DO:

- ✅ Write tests for all new features
- ✅ Test both success and failure cases
- ✅ Use table-driven tests for multiple scenarios
- ✅ Add comments explaining complex test logic
- ✅ Skip tests gracefully when requirements aren't met
- ✅ Use `t.TempDir()` for temporary files
- ✅ Clean up resources in `defer` statements
- ✅ Run tests locally before pushing
- ✅ Aim for high coverage but focus on quality
- ✅ Use meaningful test names

### DON'T:

- ❌ Don't skip CI failures without understanding them
- ❌ Don't write tests that depend on external services
- ❌ Don't hardcode file paths (use temp directories)
- ❌ Don't write tests that modify system state without cleanup
- ❌ Don't ignore race detector warnings
- ❌ Don't test internal implementation details
- ❌ Don't write flaky tests that pass/fail randomly
- ❌ Don't commit failing tests

## Contributing Tests

When contributing tests:

1. **Follow existing patterns**: Match the style of existing tests
2. **Test edge cases**: Include boundary conditions and error cases
3. **Add documentation**: Comment complex test logic
4. **Verify coverage**: Ensure new code has adequate test coverage
5. **Run locally**: Test on your machine before submitting PR
6. **Check CI**: Ensure all CI checks pass

### Pull Request Checklist

- [ ] All unit tests pass locally
- [ ] Integration tests pass (if applicable)
- [ ] Code coverage maintained or improved
- [ ] No new linting errors
- [ ] Static analysis passes
- [ ] Tests added for new features
- [ ] Existing tests updated if behavior changed
- [ ] CI pipeline passes

## Resources

### Testing Documentation

- [Go Testing Package](https://pkg.go.dev/testing)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)

### Tools

- [golangci-lint](https://golangci-lint.run/) - Comprehensive Go linter
- [staticcheck](https://staticcheck.io/) - Go static analysis
- [gosec](https://github.com/securego/gosec) - Security scanner for Go

### CI/CD

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Codecov](https://about.codecov.io/) - Code coverage reporting

## Questions?

If you have questions about testing:

1. Check this documentation
2. Look at existing tests for examples
3. Open an issue on GitHub
4. Ask in discussions

---

**Last Updated**: November 17, 2025
**Maintainers**: Containr Core Team
**Status**: Active Development
