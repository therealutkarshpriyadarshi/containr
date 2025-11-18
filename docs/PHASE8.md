# Phase 8: Developer Experience & Advanced Tooling

**Status:** ✅ Complete
**Version:** 1.3.0
**Date:** November 18, 2025

## Overview

Phase 8 represents the culmination of containr's evolution with a focus on developer experience and advanced tooling. This phase transforms containr into a comprehensive development platform with IDE integration, advanced debugging capabilities, automated security scanning, GitOps workflows, hot reload functionality, and a complete testing framework.

## Goals

1. **IDE Integration**: Provide first-class IDE support with LSP and development tools
2. **Advanced Debugging**: Enable deep container debugging and profiling
3. **SBOM & Security**: Automate security scanning and compliance checking
4. **GitOps**: Implement continuous deployment from Git repositories
5. **Hot Reload**: Enable rapid development with automatic file sync and reload
6. **Testing Framework**: Provide comprehensive container testing utilities

---

## 🎯 Features

### 8.1 IDE Integration & LSP Support

Language Server Protocol support for intelligent code editing and validation.

#### What is LSP?

The Language Server Protocol provides IDE features like:
- **IntelliSense**: Smart code completion
- **Diagnostics**: Real-time error detection
- **Go-to-Definition**: Navigate to definitions
- **Hover Information**: Contextual documentation

#### Implementation

**Package**: `pkg/ide`

**Features**:
- Language Server Protocol implementation
- Dockerfile/Containrfile autocomplete
- Docker Compose YAML validation
- Real-time syntax checking
- Code snippets and templates
- Multi-editor support (VS Code, Vim, Emacs)

**LSP Server**:
```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/ide"

// Create LSP server
config := &ide.Config{
    LSPEnabled:   true,
    LSPPort:      4389,
    DebugEnabled: true,
    DebugPort:    4711,
}

integration := ide.NewIDEIntegration(config)

// Start LSP server
err := integration.StartLSPServer(ctx, 4389)
```

**Usage**:
```bash
# Start LSP server
containr ide lsp --port 4389

# Start debug adapter
containr ide debug --port 4711

# Generate VS Code extension
containr ide generate-extension --output containr-vscode

# List supported features
containr ide capabilities
```

**VS Code Integration**:
```json
{
  "name": "containr",
  "version": "1.0.0",
  "contributes": {
    "configuration": {
      "title": "Containr",
      "properties": {
        "containr.lsp.enabled": {
          "type": "boolean",
          "default": true
        },
        "containr.lsp.port": {
          "type": "number",
          "default": 4389
        }
      }
    },
    "languages": [
      {
        "id": "dockerfile",
        "extensions": [".Dockerfile", ".Containrfile"]
      }
    ]
  }
}
```

**Autocomplete Example**:
```dockerfile
# Type "FROM" and get suggestions:
FROM alpine:latest    # Base image suggestion
FROM ubuntu:22.04     # Alternative suggestions
FROM node:18-alpine   # Context-aware suggestions

# Type "RUN" and get snippets:
RUN apt-get update && apt-get install -y \
    package1 \
    package2

# Smart parameter completion:
EXPOSE 8080           # Common ports suggested
WORKDIR /app          # Common paths suggested
```

#### Educational Value

- **Protocol Design**: Learn Language Server Protocol architecture
- **Editor Integration**: Understand how IDEs communicate with language servers
- **Code Intelligence**: Practice building autocomplete and diagnostics
- **Developer Tools**: Learn tooling that improves developer productivity

---

### 8.2 Advanced Debugging & Profiling

Comprehensive debugging and performance profiling for containers.

#### What is Container Debugging?

Container debugging provides:
- **Interactive Debugging**: Breakpoints and step-through execution
- **System Call Tracing**: Monitor all syscalls in real-time
- **Performance Profiling**: CPU, memory, and I/O profiling
- **Live Inspection**: Inspect running container state

#### Implementation

**Package**: `pkg/debug`

**Features**:
- Interactive debugger with breakpoints
- System call tracing with ptrace
- CPU and memory profiling
- Goroutine analysis
- Allocation tracking
- Performance bottleneck detection

**Debugger Interface**:
```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/debug"

// Create debugger
config := &debug.Config{
    ContainerID:  "myapp",
    EnableTrace:  true,
    EnableProfile: true,
}

debugger := debug.NewDebugger(config)

// Attach to container
err := debugger.Attach(ctx, containerPID)

// Add breakpoints
bp, err := debugger.AddBreakpoint("syscall:open", debug.BreakpointTypeSyscall)

// Start syscall tracing
err = debugger.StartSyscallTrace(ctx, os.Stdout)

// Profile CPU
err = debugger.ProfileCPU(30*time.Second, cpuFile)

// Profile memory
err = debugger.ProfileMemory(time.Minute, memFile)
```

**Usage**:
```bash
# Attach debugger to container
containr debug attach myapp

# Interactive debug session
containr debug interactive myapp
> breakpoint add syscall:write
> breakpoint add network:connect
> continue

# Trace syscalls
containr debug trace myapp --syscalls all

# Profile CPU for 30 seconds
containr debug profile cpu myapp --duration 30s --output cpu.prof

# Profile memory
containr debug profile memory myapp --output mem.prof

# Analyze goroutines
containr debug goroutines myapp

# Trace allocations
containr debug trace alloc myapp --duration 1m

# View profile with pprof
go tool pprof cpu.prof
```

**Interactive Session**:
```
containr debug interactive myapp

Debug session started for container myapp (PID: 12345)

(containr-debug) help
Available commands:
  continue (c)     - Continue execution
  step (s)         - Step to next instruction
  next (n)         - Step over function call
  breakpoint (bp)  - Manage breakpoints
  backtrace (bt)   - Show call stack
  print (p)        - Print variable
  quit (q)         - Exit debugger

(containr-debug) bp add syscall:open
Breakpoint 1 added at syscall:open

(containr-debug) continue
Breakpoint 1 hit: syscall:open("/etc/passwd", O_RDONLY)

(containr-debug) bt
#0  open("/etc/passwd", O_RDONLY) at syscall.go:123
#1  readFile("/etc/passwd") at main.go:45
#2  main() at main.go:20
```

**Profiling Analysis**:
```bash
# Generate CPU profile
containr debug profile cpu myapp --duration 60s --output cpu.prof

# Analyze with pprof
go tool pprof -http=:8080 cpu.prof

# Top functions by CPU usage
go tool pprof -top cpu.prof

# View call graph
go tool pprof -web cpu.prof
```

**Configuration**:
```yaml
# /etc/containr/debug.yaml
debug:
  enabled: true

  tracing:
    syscalls: true
    network: true
    file_io: true

  profiling:
    cpu: true
    memory: true
    block: true
    mutex: true

  breakpoints:
    max_count: 100

  output:
    directory: /var/log/containr/debug
    format: json
```

#### Educational Value

- **System Calls**: Deep understanding of Linux syscalls
- **Performance Analysis**: Learn profiling and optimization
- **Debugging Techniques**: Practice debugging complex systems
- **Tool Building**: Build debugging tools from scratch

---

### 8.3 SBOM Generation & Security Scanning

Automated Software Bill of Materials generation and vulnerability scanning.

#### What is SBOM?

Software Bill of Materials provides:
- **Transparency**: Complete list of software components
- **Security**: Vulnerability tracking and management
- **Compliance**: License and regulatory compliance
- **Supply Chain**: Track software dependencies

#### Implementation

**Package**: `pkg/sbom`

**Features**:
- SBOM generation in multiple formats (SPDX, CycloneDX, Syft)
- Vulnerability scanning (Trivy, Grype, Clair, Anchore)
- License compliance checking
- Dependency analysis
- CVE tracking and remediation
- Compliance reporting

**SBOM Generation**:
```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/sbom"

// Create SBOM generator
generator := sbom.NewGenerator(sbom.FormatSPDX)

// Generate SBOM
sbomDoc, err := generator.Generate(ctx, "myapp:latest")

// Export SBOM
err = generator.Export(sbomDoc, "sbom.json")
```

**Vulnerability Scanning**:
```go
// Create scanner
scanner := sbom.NewScanner(sbom.BackendTrivy)

// Scan image
result, err := scanner.Scan(ctx, "myapp:latest")

// Filter critical vulnerabilities
critical := scanner.FilterBySeverity(result, "critical")

// Export scan results
err = scanner.ExportScanResult(result, "scan-results.json", "json")
```

**Usage**:
```bash
# Generate SBOM
containr sbom generate myapp:latest --format spdx --output sbom.json

# Generate SBOM in CycloneDX format
containr sbom generate myapp:latest --format cyclonedx --output sbom.xml

# Scan image for vulnerabilities
containr scan myapp:latest --scanner trivy

# Scan with specific severity
containr scan myapp:latest --min-severity high

# Generate compliance report
containr compliance check myapp:latest \
    --allowed-licenses MIT,Apache-2.0 \
    --denied-licenses GPL \
    --max-severity medium \
    --output report.pdf

# Scan and generate SBOM
containr scan myapp:latest --with-sbom --output scan-with-sbom.json

# List all vulnerabilities
containr scan list myapp:latest

# Show vulnerability details
containr scan detail CVE-2024-1234
```

**SBOM Format (SPDX)**:
```json
{
  "spdxVersion": "SPDX-2.3",
  "dataLicense": "CC0-1.0",
  "name": "myapp:latest",
  "documentNamespace": "https://example.com/sbom/myapp-latest",
  "packages": [
    {
      "name": "alpine-base",
      "versionInfo": "3.18.4",
      "licenseDeclared": "MIT",
      "filesAnalyzed": false,
      "externalRefs": [
        {
          "referenceCategory": "PACKAGE_MANAGER",
          "referenceType": "purl",
          "referenceLocator": "pkg:apk/alpine/alpine-base@3.18.4"
        }
      ]
    }
  ]
}
```

**Scan Results**:
```json
{
  "image": "myapp:latest",
  "timestamp": "2025-11-18T10:00:00Z",
  "scanner": "trivy",
  "vulnerabilities": [
    {
      "id": "CVE-2024-1234",
      "package": "openssl",
      "version": "1.1.1q",
      "severity": "critical",
      "description": "Buffer overflow in OpenSSL",
      "fixedIn": "1.1.1r",
      "cve": ["CVE-2024-1234"],
      "urls": ["https://nvd.nist.gov/vuln/detail/CVE-2024-1234"]
    }
  ],
  "summary": {
    "totalVulnerabilities": 15,
    "critical": 2,
    "high": 5,
    "medium": 6,
    "low": 2
  }
}
```

**Compliance Check**:
```bash
# Check compliance
containr compliance check myapp:latest \
    --policy /etc/containr/compliance-policy.yaml

# compliance-policy.yaml
licenses:
  allowed:
    - MIT
    - Apache-2.0
    - BSD-3-Clause
  denied:
    - GPL
    - AGPL

security:
  max_severity: medium
  block_on_critical: true

  exceptions:
    - CVE-2023-1234  # Known false positive
```

#### Educational Value

- **Supply Chain Security**: Learn software supply chain risks
- **Vulnerability Management**: Practice CVE tracking and remediation
- **Compliance**: Understand licensing and regulatory requirements
- **Security Tools**: Work with industry-standard security scanners

---

### 8.4 GitOps & CI/CD Integration

Continuous deployment using Git as the source of truth.

#### What is GitOps?

GitOps provides:
- **Declarative**: Infrastructure and apps defined as code
- **Versioned**: All changes tracked in Git
- **Automated**: Automatic synchronization and deployment
- **Auditable**: Complete deployment history

#### Implementation

**Package**: `pkg/gitops`

**Features**:
- Git repository monitoring
- Automatic deployment synchronization
- Declarative deployment manifests
- Rollback capabilities
- Multi-environment support
- CI/CD pipeline execution

**GitOps Controller**:
```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/gitops"

// Create GitOps controller
config := &gitops.Config{
    Repository: "https://github.com/example/deployments",
    Branch:     "main",
    Path:       "manifests",
    Interval:   30 * time.Second,
}

controller := gitops.NewController(config)

// Start controller
err := controller.Start(ctx)

// List deployments
deployments := controller.ListDeployments()

// Get specific deployment
deployment, err := controller.GetDeployment("myapp")

// Delete deployment
err = controller.DeleteDeployment(ctx, "myapp")
```

**Usage**:
```bash
# Initialize GitOps repository
containr gitops init https://github.com/example/deployments

# Start GitOps controller
containr gitops start \
    --repo https://github.com/example/deployments \
    --branch main \
    --interval 30s

# List deployments
containr gitops list

# Sync manually
containr gitops sync

# View deployment status
containr gitops status myapp

# Rollback deployment
containr gitops rollback myapp --revision abc123

# View deployment history
containr gitops history myapp
```

**Deployment Manifest**:
```yaml
# deployments/myapp.yaml
apiVersion: containr.io/v1
kind: Deployment
metadata:
  name: myapp
  namespace: production
spec:
  replicas: 3
  image: registry.example.com/myapp:v1.2.3
  command:
    - /app/server
  args:
    - --port=8080
  environment:
    PORT: "8080"
    LOG_LEVEL: "info"
  volumes:
    - name: data
      mountPath: /data
  resources:
    limits:
      cpu: "2"
      memory: "4Gi"
    requests:
      cpu: "1"
      memory: "2Gi"
```

**Pipeline Execution**:
```go
// Create pipeline
pipeline := &gitops.Pipeline{
    Name: "build-and-deploy",
    Stages: []*gitops.Stage{
        {
            Name: "build",
            Steps: []*gitops.Step{
                {
                    Name:    "compile",
                    Command: "make",
                    Args:    []string{"build"},
                    Image:   "golang:1.21",
                },
            },
        },
        {
            Name: "deploy",
            Steps: []*gitops.Step{
                {
                    Name:    "deploy",
                    Command: "kubectl",
                    Args:    []string{"apply", "-f", "manifests/"},
                },
            },
        },
    },
}

executor := gitops.NewPipelineExecutor()
err := executor.Execute(ctx, pipeline)
```

**Configuration**:
```yaml
# /etc/containr/gitops.yaml
gitops:
  enabled: true

  repository:
    url: https://github.com/example/deployments
    branch: main
    path: manifests
    auth:
      token: ${GITHUB_TOKEN}

  sync:
    interval: 30s
    auto_sync: true
    prune: true

  notifications:
    slack:
      webhook: ${SLACK_WEBHOOK}
      channel: "#deployments"
    email:
      smtp: smtp.example.com
      to: ops@example.com
```

#### Educational Value

- **GitOps Principles**: Learn declarative infrastructure management
- **Continuous Deployment**: Practice automated deployment workflows
- **Version Control**: Understand infrastructure as code
- **Reconciliation**: Learn desired vs actual state management

---

### 8.5 Hot Reload & Development Workflows

Rapid development with automatic file synchronization and container reload.

#### What is Hot Reload?

Hot reload provides:
- **Fast Iteration**: Instant code changes without rebuild
- **File Sync**: Automatic synchronization between host and container
- **Live Reload**: Automatic application restart on changes
- **Development Mode**: Optimized for development workflows

#### Implementation

**Package**: `pkg/hotreload`

**Features**:
- File system watching
- Bidirectional file synchronization
- Multiple reload strategies (restart, signal, exec)
- Ignore patterns and filters
- Debouncing and throttling
- Development environment templates

**Hot Reload Setup**:
```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/hotreload"

// Create watcher
config := &hotreload.Config{
    ContainerID:    "myapp-dev",
    WatchPaths:     []string{"/app/src"},
    IgnorePatterns: []string{"*.tmp", "*.log", "node_modules"},
    Debounce:       500 * time.Millisecond,
}

watcher := hotreload.NewWatcher(config)

// Start watching
err := watcher.Start(ctx)

// Create syncer
syncer := hotreload.NewSyncer("myapp-dev")
syncer.AddSyncPair(&hotreload.SyncPair{
    HostPath:      "./src",
    ContainerPath: "/app/src",
    Direction:     hotreload.SyncDirectionBidirectional,
})

// Create reload manager
reloader := hotreload.NewReloadManager("myapp-dev", hotreload.StrategySignal)
```

**Usage**:
```bash
# Start development mode with hot reload
containr dev myapp \
    --watch ./src:/app/src \
    --watch ./config:/app/config:ro \
    --reload-strategy restart \
    --ignore "*.tmp,*.log,node_modules"

# Use signal-based reload
containr dev myapp \
    --watch ./src:/app/src \
    --reload-strategy signal \
    --reload-signal SIGHUP

# Use exec-based reload
containr dev myapp \
    --watch ./src:/app/src \
    --reload-strategy exec \
    --reload-command "npm run reload"

# View reload statistics
containr dev stats myapp-dev

# Stop development mode
containr dev stop myapp-dev
```

**Development Environment**:
```bash
# Create dev environment from template
containr dev template create nodejs --name myapp-dev

# This creates:
# - Container with development tools
# - Hot reload configuration
# - Debug port forwarding
# - Volume mounts for source code

# Start dev environment
containr dev start myapp-dev

# View logs with auto-refresh
containr dev logs myapp-dev --follow

# Execute command in dev container
containr dev exec myapp-dev npm test
```

**Configuration**:
```yaml
# .containr/dev.yaml
name: myapp-dev
image: node:18-alpine

watch:
  - path: ./src
    target: /app/src
    ignore:
      - "*.tmp"
      - "*.log"
      - "node_modules"

  - path: ./public
    target: /app/public

reload:
  strategy: signal
  signal: SIGHUP
  debounce: 500ms

sync:
  interval: 2s
  direction: bidirectional

ports:
  - "3000:3000"   # App port
  - "9229:9229"   # Debug port

environment:
  NODE_ENV: development
  DEBUG: "*"
```

**Reload Strategies**:
```go
// Restart strategy - full container restart
hotreload.StrategyRestart

// Signal strategy - send signal to process
hotreload.StrategySignal

// Exec strategy - execute reload command
hotreload.StrategyExec

// Rolling strategy - rolling restart for multi-replica
hotreload.StrategyRolling
```

#### Educational Value

- **File Watching**: Learn file system event monitoring
- **Process Signals**: Understand Unix signals and process management
- **Development Workflows**: Practice efficient development setups
- **Synchronization**: Learn file synchronization algorithms

---

### 8.6 Container Testing Framework

Comprehensive testing utilities for containers and applications.

#### What is Container Testing?

Container testing provides:
- **Unit Tests**: Test individual containers
- **Integration Tests**: Test multi-container applications
- **Behavior Tests**: BDD-style testing
- **Assertions**: Rich assertion library
- **Test Fixtures**: Reusable test setups

#### Implementation

**Package**: `pkg/testframework`

**Features**:
- Container test runner
- Assertion framework
- Behavior-driven testing (BDD)
- Integration test support
- Test fixtures and helpers
- Parallel test execution

**Container Testing**:
```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/testframework"

// Create test
test := &testframework.ContainerTest{
    Name:    "web-server-test",
    Image:   "nginx:latest",
    Command: []string{"nginx", "-g", "daemon off;"},
    Timeout: 30 * time.Second,
    Assertions: []testframework.Assertion{
        testframework.AssertExitCode(0),
        testframework.AssertPortOpen(80),
        testframework.AssertFileExists("/etc/nginx/nginx.conf"),
        testframework.AssertOutputContains("stdout", "nginx"),
    },
}

// Run test
runner := testframework.NewTestRunner()
err := runner.RunTest(ctx, test)
```

**Usage**:
```bash
# Run container tests
containr test run ./tests/

# Run specific test
containr test run web-server-test

# Run with verbose output
containr test run -v ./tests/

# Run tests in parallel
containr test run --parallel 4 ./tests/

# Generate test report
containr test run --report junit.xml ./tests/
```

**Test File Example**:
```yaml
# tests/web-server.yaml
name: web-server-test
image: nginx:latest
command:
  - nginx
  - -g
  - daemon off;
timeout: 30s

assertions:
  - type: exit_code
    value: 0

  - type: port_open
    port: 80

  - type: file_exists
    path: /etc/nginx/nginx.conf

  - type: output_contains
    stream: stdout
    text: "nginx"
```

**Behavior-Driven Testing**:
```go
// Create behavior test
test := &testframework.BehaviorTest{
    Description: "Web server serves static files",
    Given: func(ctx context.Context) error {
        // Setup: Create test files
        return createTestFiles()
    },
    When: func(ctx context.Context) error {
        // Action: Start web server
        return startWebServer()
    },
    Then: func(ctx context.Context) error {
        // Assert: Check files are served
        return checkFilesServed()
    },
}

runner := testframework.NewBehaviorRunner()
err := runner.Run(ctx, test)
```

**Integration Testing**:
```go
// Create integration test
test := &testframework.IntegrationTest{
    Name: "microservices-integration",
    Containers: []*testframework.TestContainer{
        {Name: "api", Image: "myapp/api:latest"},
        {Name: "db", Image: "postgres:15"},
        {Name: "cache", Image: "redis:7"},
    },
    Network: "test-network",
    Setup: func(ctx context.Context) error {
        // Setup test environment
        return setupDatabase()
    },
    Test: func(ctx context.Context) error {
        // Run integration tests
        return testAPIWithDatabase()
    },
    Teardown: func(ctx context.Context) error {
        // Cleanup
        return cleanupTestData()
    },
}

runner := testframework.NewIntegrationRunner()
err := runner.Run(ctx, test)
```

**Assertions Available**:
```go
// Exit code assertion
testframework.AssertExitCode(0)

// Output contains assertion
testframework.AssertOutputContains("stdout", "success")

// Port open assertion
testframework.AssertPortOpen(8080)

// File exists assertion
testframework.AssertFileExists("/app/config.yaml")

// HTTP response assertion
testframework.AssertHTTPStatus("http://localhost:8080", 200)

// Custom assertion
testframework.AssertCustom(func(container *TestContainer) error {
    // Custom validation logic
    return nil
})
```

#### Educational Value

- **Test Design**: Learn testing best practices
- **Assertion Patterns**: Practice test assertion design
- **BDD Methodology**: Understand behavior-driven development
- **Integration Testing**: Learn multi-service testing

---

## 📦 Package Structure

```
pkg/
├── ide/                      # IDE Integration
│   ├── ide.go               # IDE integration manager
│   ├── lsp.go               # Language Server Protocol
│   ├── lsp_test.go          # LSP tests
│   └── README.md            # IDE documentation
├── debug/                   # Advanced Debugging
│   ├── debugger.go          # Interactive debugger
│   ├── debugger_test.go     # Debugger tests
│   └── profiler.go          # Performance profiler
├── sbom/                    # SBOM & Security Scanning
│   ├── sbom.go              # SBOM generator
│   ├── sbom_test.go         # SBOM tests
│   ├── scanner.go           # Vulnerability scanner
│   └── compliance.go        # Compliance checker
├── gitops/                  # GitOps & CI/CD
│   ├── gitops.go            # GitOps controller
│   ├── gitops_test.go       # GitOps tests
│   ├── reconciler.go        # State reconciler
│   └── pipeline.go          # CI/CD pipeline executor
├── hotreload/               # Hot Reload
│   ├── hotreload.go         # File watcher
│   ├── hotreload_test.go    # Hot reload tests
│   ├── syncer.go            # File syncer
│   └── reloader.go          # Reload manager
└── testframework/           # Testing Framework
    ├── testframework.go     # Test runner
    ├── testframework_test.go # Framework tests
    ├── assertions.go        # Assertion library
    └── integration.go       # Integration test support
```

## 🎨 CLI Commands

### IDE Commands
```bash
containr ide lsp                        # Start LSP server
containr ide debug                      # Start debug adapter
containr ide generate-extension         # Generate IDE extension
containr ide capabilities               # List IDE features
```

### Debug Commands
```bash
containr debug attach <container>       # Attach debugger
containr debug interactive <container>  # Interactive session
containr debug trace <container>        # Trace syscalls
containr debug profile cpu <container>  # CPU profiling
containr debug profile memory <container> # Memory profiling
containr debug goroutines <container>   # Analyze goroutines
```

### SBOM & Scanning Commands
```bash
containr sbom generate <image>          # Generate SBOM
containr scan <image>                   # Scan for vulnerabilities
containr compliance check <image>       # Check compliance
containr scan detail <CVE>              # Show CVE details
```

### GitOps Commands
```bash
containr gitops init <repo>             # Initialize GitOps
containr gitops start                   # Start controller
containr gitops sync                    # Manual sync
containr gitops status <deployment>     # View status
containr gitops rollback <deployment>   # Rollback deployment
```

### Hot Reload Commands
```bash
containr dev <container>                # Start dev mode
containr dev stats <container>          # View reload stats
containr dev stop <container>           # Stop dev mode
containr dev template create <type>     # Create template
```

### Testing Commands
```bash
containr test run <path>                # Run tests
containr test run -v <path>             # Verbose output
containr test run --parallel <n>        # Parallel execution
```

## 🧪 Testing

```bash
# Run all Phase 8 tests
make test-phase8

# Individual component tests
go test ./pkg/ide/...
go test ./pkg/debug/...
go test ./pkg/sbom/...
go test ./pkg/gitops/...
go test ./pkg/hotreload/...
go test ./pkg/testframework/...

# Integration tests
make test-ide-integration
make test-debug-integration
make test-gitops-integration

# E2E tests
make test-phase8-e2e
```

## 📚 Documentation

- [IDE Integration Guide](tutorials/15-ide-integration.md)
- [Advanced Debugging Guide](tutorials/16-advanced-debugging.md)
- [SBOM & Security Scanning Guide](tutorials/17-sbom-scanning.md)
- [GitOps Deployment Guide](tutorials/18-gitops-deployment.md)
- [Hot Reload Development Guide](tutorials/19-hot-reload-dev.md)
- [Container Testing Guide](tutorials/20-container-testing.md)

## 🚀 Development Workflow

Phase 8 enables modern development workflows:

```bash
# 1. Start development environment
containr dev myapp \
    --watch ./src:/app/src \
    --reload-strategy signal \
    --debug-port 9229

# 2. Edit code (changes automatically sync and reload)
vim src/server.js

# 3. Debug if needed
containr debug attach myapp
> breakpoint add /app/src/server.js:42
> continue

# 4. Run tests
containr test run ./tests/

# 5. Scan for vulnerabilities
containr scan myapp:dev

# 6. Commit and push (GitOps handles deployment)
git add .
git commit -m "Add new feature"
git push
```

## 🎓 Educational Value

Phase 8 teaches modern development practices:
- **IDE Integration**: Build development tools and language servers
- **Advanced Debugging**: Master debugging techniques and profiling
- **Security**: Practice supply chain security and vulnerability management
- **GitOps**: Learn modern deployment workflows
- **Developer Experience**: Design efficient development environments
- **Testing**: Comprehensive testing strategies

---

**Phase 8 Status**: ✅ Complete
**Version**: 1.3.0
**Next**: Community adoption and real-world use cases

This phase completes containr's transformation into a comprehensive, production-ready, developer-friendly container runtime and development platform! 🚀
