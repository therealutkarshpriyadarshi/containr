# Contributing to Containr

Thank you for your interest in contributing to containr! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Project Structure](#project-structure)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Code Style](#code-style)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Issue Guidelines](#issue-guidelines)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

### Our Pledge

We are committed to providing a welcoming and inclusive environment for all contributors. We expect everyone to:

- Be respectful and considerate
- Welcome newcomers and help them learn
- Focus on constructive feedback
- Accept gracefully when others disagree
- Prioritize the project's best interests

### Unacceptable Behavior

- Harassment, discrimination, or offensive language
- Personal attacks or trolling
- Publishing others' private information
- Intentionally disrupting discussions
- Any conduct that would be inappropriate in a professional setting

## Getting Started

### Prerequisites

- **Operating System:** Linux (kernel 3.8+)
- **Go:** Version 1.16 or later
- **Git:** For version control
- **Root Access:** Required for testing namespace/cgroup operations

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:

```bash
git clone https://github.com/YOUR_USERNAME/containr.git
cd containr
```

3. Add upstream remote:

```bash
git remote add upstream https://github.com/therealutkarshpriyadarshi/containr.git
```

4. Keep your fork synchronized:

```bash
git fetch upstream
git checkout main
git merge upstream/main
```

## Development Environment

### Install Dependencies

```bash
# Install Go dependencies
make deps

# Install development tools
make install-tools
```

### Development Tools

The project uses several development tools:

- **golangci-lint:** Comprehensive linter
- **staticcheck:** Static analysis
- **gosec:** Security scanner
- **gofmt:** Code formatter

Install them with:

```bash
make install-tools
```

### Building

```bash
# Build the binary
make build

# Build with debug symbols
go build -o bin/containr ./cmd/containr

# Install system-wide
sudo make install
```

### Running Tests

```bash
# Run unit tests
make test-unit

# Run integration tests (requires root)
sudo make test-integration

# Run all tests with coverage
make test-coverage

# Run benchmarks
make bench
```

## Project Structure

```
containr/
├── cmd/
│   └── containr/          # CLI application
│       ├── main.go        # Entry point
│       ├── run.go         # Run command
│       ├── container.go   # Container commands
│       ├── image.go       # Image commands
│       ├── volume.go      # Volume commands
│       └── network.go     # Network commands
├── pkg/
│   ├── benchmark/         # Benchmarking utilities
│   ├── profiler/          # Profiling support
│   ├── runtime/           # OCI runtime spec
│   ├── version/           # Version information
│   ├── container/         # Container management
│   ├── namespace/         # Namespace handling
│   ├── cgroup/            # Cgroup management
│   ├── rootfs/            # Filesystem operations
│   ├── network/           # Networking
│   ├── image/             # Image management
│   ├── volume/            # Volume management
│   ├── registry/          # Registry client
│   ├── metrics/           # Metrics collection
│   ├── events/            # Event system
│   ├── health/            # Health checks
│   ├── restart/           # Restart policies
│   ├── build/             # Dockerfile parser
│   ├── capabilities/      # Capabilities management
│   ├── seccomp/           # Seccomp profiles
│   ├── security/          # LSM support
│   ├── logger/            # Structured logging
│   ├── errors/            # Error handling
│   ├── state/             # State persistence
│   └── userns/            # User namespaces
├── test/                  # Integration tests
├── examples/              # Example programs
├── docs/                  # Documentation
│   ├── PHASE1.md
│   ├── PHASE2.md
│   ├── PHASE3.md
│   ├── PHASE4.md
│   ├── ARCHITECTURE.md
│   ├── SECURITY.md
│   ├── TESTING.md
│   └── tutorials/
├── .github/
│   └── workflows/         # CI/CD workflows
├── Makefile              # Build automation
├── go.mod                # Go module definition
├── README.md             # Project overview
├── ROADMAP.md            # Development roadmap
└── CONTRIBUTING.md       # This file
```

## Making Changes

### Creating a Branch

Always create a feature branch for your changes:

```bash
# Feature branch
git checkout -b feature/my-new-feature

# Bug fix branch
git checkout -b fix/issue-123

# Documentation branch
git checkout -b docs/improve-readme
```

### Making Commits

Follow these guidelines:

1. **Small, focused commits:** Each commit should represent a single logical change
2. **Descriptive messages:** Explain what and why, not how
3. **Reference issues:** Include issue numbers when applicable

### Code Changes

1. **Write tests first** (TDD approach recommended)
2. **Implement the feature** or fix
3. **Run tests** to ensure everything passes
4. **Update documentation** if needed
5. **Run linters** and fix any issues

```bash
# Run pre-commit checks
make pre-commit

# Format code
make fmt

# Run linters
make lint

# Run tests
make test
```

## Testing

### Writing Tests

Every package should have comprehensive tests:

```go
// pkg/mypackage/mypackage_test.go
package mypackage

import (
    "testing"
)

func TestMyFunction(t *testing.T) {
    result := MyFunction("input")
    expected := "expected-output"

    if result != expected {
        t.Errorf("expected %s, got %s", expected, result)
    }
}

func BenchmarkMyFunction(b *testing.B) {
    for i := 0; i < b.N; i++ {
        MyFunction("input")
    }
}
```

### Test Coverage

Maintain high test coverage (>70%):

```bash
# Check coverage
make test-coverage

# View coverage report
open coverage/coverage.html
```

### Integration Tests

Integration tests require root privileges:

```bash
# Run as root
sudo make test-integration
```

## Code Style

### Go Style Guidelines

Follow standard Go conventions:

1. **Formatting:** Use `gofmt` (automated via `make fmt`)
2. **Naming:**
   - Exported names: `PublicFunction`, `PublicType`
   - Unexported names: `privateFunction`, `privateType`
   - Acronyms: `HTTPServer`, `URLPath` (all caps)
3. **Comments:**
   - Package comment at top of package file
   - Exported functions/types must have doc comments
   - Comments should be complete sentences

```go
// Package container provides container lifecycle management.
// It handles creation, starting, stopping, and removal of containers
// using Linux namespaces, cgroups, and filesystem isolation.
package container

// Config holds container configuration options.
type Config struct {
    Name     string
    Hostname string
    Command  []string
}

// New creates a new container with the given configuration.
// It validates the configuration and initializes internal state
// but does not start the container.
func New(config *Config) (*Container, error) {
    // Implementation
}
```

### Error Handling

Use the errors package for error handling:

```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/errors"

// Return descriptive errors
func DoSomething() error {
    if err := operation(); err != nil {
        return errors.Wrap(err, errors.ErrCodeInternal, "operation failed")
    }
    return nil
}

// Use error codes for programmatic handling
if errors.IsCode(err, errors.ErrCodeNotFound) {
    // Handle not found error
}
```

### Logging

Use structured logging:

```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/logger"

log := logger.New("component")
log.Info("Starting operation")
log.WithField("container", containerID).Debug("Container state changed")
log.WithError(err).Error("Operation failed")
```

## Commit Guidelines

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**

```
feat(container): add support for custom DNS servers

Implement custom DNS configuration for containers allowing
users to specify their own DNS servers via --dns flag.

Closes #123
```

```
fix(network): resolve port mapping race condition

Fix race condition in port mapping setup that could cause
containers to fail to start when multiple containers are
created simultaneously.

Fixes #456
```

```
docs(phase4): add performance tuning guide

Add comprehensive guide covering profiling, benchmarking,
and optimization techniques for container operations.
```

## Pull Request Process

### Before Submitting

1. **Ensure tests pass:**
   ```bash
   make test-all
   ```

2. **Run linters:**
   ```bash
   make lint
   ```

3. **Update documentation:**
   - Update README.md if adding features
   - Add/update doc comments
   - Update relevant docs/ files

4. **Add yourself to contributors:**
   ```bash
   # If there's a CONTRIBUTORS file
   ```

### Creating a Pull Request

1. **Push your branch:**
   ```bash
   git push origin feature/my-new-feature
   ```

2. **Open PR on GitHub**

3. **Fill out PR template:**
   - Description of changes
   - Related issues
   - Testing performed
   - Screenshots (if applicable)

### PR Template

```markdown
## Description
Brief description of changes

## Related Issues
Closes #123

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review performed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] Tests added/updated
```

### Review Process

1. **Automated checks:** CI must pass
2. **Code review:** At least one approval required
3. **Address feedback:** Make requested changes
4. **Maintainer merge:** Once approved

## Issue Guidelines

### Creating Issues

Use appropriate issue templates:

**Bug Report:**
```markdown
## Bug Description
Clear description of the bug

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: Ubuntu 22.04
- Containr version: 1.0.0
- Go version: 1.21.0

## Additional Context
Any other relevant information
```

**Feature Request:**
```markdown
## Feature Description
Clear description of the proposed feature

## Use Case
Why is this feature needed?

## Proposed Implementation
How could this be implemented?

## Alternatives Considered
Other approaches that were considered

## Additional Context
Any other relevant information
```

### Issue Labels

- `bug`: Something isn't working
- `enhancement`: New feature or request
- `documentation`: Documentation improvements
- `good first issue`: Good for newcomers
- `help wanted`: Extra attention needed
- `question`: Further information requested
- `wontfix`: This will not be worked on

## Documentation

### Documentation Standards

1. **Keep docs up to date** with code changes
2. **Use clear, simple language**
3. **Include code examples**
4. **Add diagrams** where helpful
5. **Link related docs**

### Documentation Types

1. **README.md:** Project overview and quick start
2. **docs/PHASE*.md:** Phase-specific features
3. **docs/ARCHITECTURE.md:** System architecture
4. **docs/tutorials/:** Step-by-step guides
5. **Code comments:** Package and function docs

### Writing Tutorials

Good tutorials should:
- Start with prerequisites
- Include complete examples
- Explain each step clearly
- Show expected output
- Include troubleshooting section

## Community

### Getting Help

- **GitHub Issues:** For bugs and feature requests
- **Discussions:** For questions and general discussion
- **Email:** For private inquiries

### Recognition

Contributors are recognized through:
- Commit history
- Release notes
- CONTRIBUTORS file (if maintained)
- Special mentions for significant contributions

## Development Tips

### Debugging

```bash
# Enable debug logging
containr run --debug alpine /bin/sh

# Use delve debugger
dlv debug ./cmd/containr -- run alpine /bin/sh

# Profile performance
containr run --profile ./profiles alpine /bin/sh
```

### Working with Namespaces

Namespace operations require root:

```bash
# Run with sudo
sudo containr run alpine /bin/sh

# Or use rootless mode (Phase 2.4)
containr run --userns-remap alpine /bin/sh
```

### Testing Network Features

```bash
# Create test network
sudo containr network create --subnet 172.30.0.0/24 testnet

# Run container in network
sudo containr run --network testnet alpine /bin/sh

# Cleanup
sudo containr network rm testnet
```

## Release Process

### Version Bumping

1. Update `pkg/version/version.go`
2. Update CHANGELOG.md
3. Create git tag
4. Push tag to trigger release

```bash
# Create tag
git tag -a v1.0.0 -m "Release version 1.0.0"

# Push tag
git push origin v1.0.0
```

### Release Checklist

- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version bumped
- [ ] Tag created
- [ ] Binaries built
- [ ] Release notes written

## Questions?

If you have questions not covered here:

1. Check existing documentation
2. Search GitHub issues
3. Ask in discussions
4. Create a new issue

## Thank You!

Thank you for contributing to containr! Your efforts help make container technology more accessible and understandable for everyone.

**Happy coding! 🚀**
