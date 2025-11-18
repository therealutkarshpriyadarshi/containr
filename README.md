# Containr 🚀

A minimal container runtime built from scratch using Linux primitives. This project demonstrates the core concepts behind Docker and other container runtimes by implementing process isolation using namespaces, resource limits with cgroups, and filesystem isolation.

**🎉 Phase 8 Complete - Developer Experience & Advanced Tooling!** Containr now includes IDE integration with LSP support, advanced debugging and profiling, SBOM generation and security scanning, GitOps continuous deployment, hot reload for rapid development, and a comprehensive container testing framework. All eight phases complete!

## Features

### 🛠️ Phase 8: Developer Experience & Advanced Tooling (NEW!)
- ✅ **IDE Integration & LSP Support**: First-class editor integration
  - **Language Server Protocol**: IntelliSense, autocomplete, and validation
  - **Real-time Diagnostics**: Instant error detection for Dockerfiles
  - **Code Snippets**: Smart templates for common patterns
  - **Multi-Editor Support**: VS Code, Vim, Emacs, and more
- ✅ **Advanced Debugging & Profiling**: Deep container inspection
  - **Interactive Debugger**: Breakpoints and step-through execution
  - **System Call Tracing**: Monitor all syscalls in real-time
  - **Performance Profiling**: CPU, memory, and I/O profiling
  - **Live Inspection**: Debug running containers without restart
- ✅ **SBOM & Security Scanning**: Automated security analysis
  - **SBOM Generation**: SPDX, CycloneDX, and Syft formats
  - **Vulnerability Scanning**: Trivy, Grype, Clair integration
  - **License Compliance**: Track and enforce license policies
  - **CVE Tracking**: Comprehensive vulnerability management
- ✅ **GitOps & CI/CD Integration**: Continuous deployment automation
  - **Git-Based Deployment**: Declarative infrastructure as code
  - **Automatic Synchronization**: Real-time deployment updates
  - **Pipeline Execution**: Integrated CI/CD workflows
  - **Rollback Support**: Easy rollback to previous versions
- ✅ **Hot Reload & Development Workflows**: Rapid development iteration
  - **File Watching**: Automatic detection of code changes
  - **Instant Sync**: Bidirectional file synchronization
  - **Multiple Reload Strategies**: Restart, signal, or exec-based
  - **Development Templates**: Pre-configured dev environments
- ✅ **Container Testing Framework**: Comprehensive testing utilities
  - **Unit Testing**: Test individual containers
  - **Behavior-Driven Testing**: BDD-style test framework
  - **Integration Testing**: Multi-container test scenarios
  - **Rich Assertions**: Extensive assertion library

### 🏢 Phase 7: Advanced Production Features & Enterprise Integration
- ✅ **RBAC & Multi-Tenancy**: Enterprise-grade access control
  - **Role-Based Access Control**: Fine-grained permissions and user management
  - **Resource Quotas**: CPU, memory, storage, and container limits per user
  - **Audit Logging**: Complete audit trail for compliance
  - **Built-in Roles**: Admin, Developer, Operator, Viewer roles
- ✅ **Advanced Observability**: OpenTelemetry integration
  - **Distributed Tracing**: Full request tracing with Jaeger/Zipkin
  - **Prometheus Metrics**: Comprehensive metrics collection and export
  - **Structured Logging**: Context-aware logging with trace correlation
  - **Performance Profiling**: CPU, memory, and trace profiling
- ✅ **Container Checkpointing & Migration**: CRIU integration
  - **Live Migration**: Move running containers between hosts
  - **Checkpoint/Restore**: Save and restore container state
  - **Pre-copy Migration**: Minimal downtime with iterative checkpointing
  - **State Persistence**: Complete process, memory, and network state
- ✅ **Service Mesh Integration**: Envoy sidecar support
  - **Traffic Management**: Load balancing, retries, circuit breakers
  - **mTLS**: Automatic mutual TLS for service-to-service communication
  - **Policy Enforcement**: Traffic policies and fault injection
  - **Observability**: Distributed tracing and metrics integration
- ✅ **Advanced Security**: Policy enforcement and supply chain security
  - **OPA Policies**: Policy-as-code with Open Policy Agent
  - **Image Signing**: Cosign/Sigstore integration for image verification
  - **Runtime Security**: Behavior monitoring and threat detection
  - **Vulnerability Scanning**: CVE scanning and compliance reporting
- ✅ **CSI Storage & Encryption**: Container Storage Interface support
  - **CSI Drivers**: Local, NFS storage drivers
  - **Volume Encryption**: LUKS encryption for data at rest
  - **Snapshots & Cloning**: Volume snapshot and clone support
  - **Dynamic Provisioning**: Automatic volume creation

### ☁️ Phase 6: Cloud-Native Integration & Advanced Runtime
- ✅ **CRI (Container Runtime Interface)**: Kubernetes integration
  - **RuntimeService**: Pod sandbox and container lifecycle management
  - **ImageService**: Image pull, list, and remove operations
  - **Pod Support**: Full Kubernetes pod abstraction
  - **Metrics Integration**: Container metrics via CRI API
- ✅ **Plugin System**: Extensible architecture for custom functionality
  - **Plugin Types**: Runtime, network, storage, logging, metrics plugins
  - **Dynamic Loading**: Load plugins at runtime
  - **Plugin Management**: Enable, disable, configure plugins
  - **Built-in Plugins**: Example plugins for common use cases
- ✅ **Snapshot Support**: Fast container creation and migration
  - **Multiple Drivers**: Overlay2, Btrfs, ZFS snapshot drivers
  - **Copy-on-Write**: Efficient storage with CoW optimization
  - **Snapshot Chain**: Parent-child snapshot relationships
  - **Metadata Management**: Snapshot info and labels
- ✅ **Complete Build Engine**: Full Dockerfile build implementation
  - **All Instructions**: Support for all Dockerfile instructions
  - **Multi-stage Builds**: Build stage targeting and optimization
  - **Build Cache**: Layer caching with smart invalidation
  - **Build Context**: Efficient context handling and hashing
  - **Build Arguments**: Dynamic build-time variables

### 🌱 Phase 5: Community & Growth
- ✅ **Community Building**: Vibrant and inclusive community
  - **Community Health**: Code of Conduct, Contributing Guide, Support docs
  - **Governance Model**: BDFL with path to community governance
  - **Recognition**: Contributor recognition and mentorship programs
  - **Communication**: GitHub Discussions, Issue templates, PR templates
- ✅ **Educational Partnerships**: Learning platform for students and educators
  - **Course Materials**: University course integration ready
  - **Tutorials**: Comprehensive step-by-step tutorials
  - **Documentation**: Complete educational documentation
  - **Examples**: Real-world examples and use cases
- ✅ **Sustainability**: Long-term project sustainability
  - **Security Policy**: Vulnerability reporting and patch process
  - **Maintainer Guide**: Clear maintainer roles and responsibilities
  - **Funding Infrastructure**: GitHub Sponsors and Open Collective ready
  - **Dependency Management**: Automated security updates
- ✅ **Project Management**: Professional project management
  - **Changelog**: Keep a Changelog compliant version history
  - **Release Process**: Semantic versioning and release automation
  - **Issue Templates**: Bug reports, feature requests, questions
  - **PR Templates**: Comprehensive pull request templates

### 🎯 Phase 4: Production Polish
- ✅ **Performance Optimization**: Production-ready performance tools
  - **Benchmarking**: Comprehensive benchmark suite for all operations
  - **Profiling**: CPU, memory, and execution trace profiling
  - **Performance Testing**: Automated performance regression testing
  - **Metrics Collection**: Detailed performance metrics and statistics
- ✅ **OCI Runtime Compliance**: Full OCI specification support
  - **OCI Runtime Spec**: Complete OCI Runtime Specification 1.0.2
  - **OCI State Management**: Container state tracking and persistence
  - **OCI Bundle Support**: Standard OCI bundle format
  - **Runtime Configuration**: Comprehensive config.json support
- ✅ **Version Management**: Professional version tracking
  - **Semantic Versioning**: Full SemVer support
  - **Build Information**: Git commit, build date, Go version
  - **Version Commands**: JSON and short output formats
  - **User Agent Strings**: HTTP user agent support
- ✅ **Release & Distribution**: Automated release pipeline
  - **Multi-platform Builds**: Linux amd64, arm64, arm
  - **GitHub Releases**: Automated release creation
  - **Checksums**: SHA256 verification for all binaries
  - **Installation Script**: One-line installation
  - **CI/CD Pipeline**: Comprehensive GitHub Actions workflows

### 🚀 Phase 3: Advanced Features
- ✅ **Enhanced Networking**: Production-ready networking capabilities
  - **Port Mapping**: TCP/UDP port exposure with iptables integration
  - **Network Modes**: Bridge, host, none, and container sharing
  - **DNS Resolution**: Automatic DNS configuration and hostname resolution
  - **Network Commands**: `network create`, `ls`, `rm`, `inspect`
- ✅ **Monitoring & Observability**: Comprehensive metrics and events
  - **Metrics Collection**: CPU, memory, network, disk I/O, and PID statistics
  - **Events API**: Container lifecycle event tracking and streaming
  - **Events Command**: Query and filter events by time and type
- ✅ **Health Checks**: Container health monitoring
  - **Configurable Checks**: Command-based health verification
  - **Health Status**: Track healthy, unhealthy, and starting states
  - **Health Events**: Automatic event emission on status changes
- ✅ **Restart Policies**: Automatic container restart
  - **Multiple Policies**: no, always, on-failure, unless-stopped
  - **Exponential Backoff**: Smart restart delays with configurable backoff
  - **Max Retries**: Limit restart attempts
- ✅ **Build Foundation**: Dockerfile parsing infrastructure
  - **Dockerfile Parser**: Parse standard Dockerfile syntax
  - **Multi-stage Support**: Build stage tracking and management
  - **Build Arguments**: ARG instruction support

### ✨ Phase 2: Feature Completeness
- ✅ **Enhanced CLI**: Docker-like commands with Cobra framework
  - **Container Lifecycle**: `run`, `create`, `start`, `stop`, `rm`, `ps`, `logs`, `exec`
  - **Image Management**: `pull`, `images`, `rmi`, `import`, `export`
  - **Inspection**: `inspect`, `stats`, `top`
- ✅ **Volume Management**: Persistent data storage
  - **Named Volumes**: Managed storage for containers
  - **Bind Mounts**: Mount host directories into containers
  - **Volume Commands**: `volume create`, `volume ls`, `volume rm`, `volume inspect`, `volume prune`
- ✅ **Registry Integration**: Pull images from Docker Hub and OCI registries
  - **Docker Hub Support**: Pull official and user images
  - **OCI Compliance**: Full OCI image format support
  - **Layer Extraction**: Automatic extraction to rootfs
- ✅ **User Namespace Support**: Rootless containers for enhanced security
  - **UID/GID Remapping**: Run as non-root with subuid/subgid
  - **Enhanced Security**: Root in container = unprivileged on host

### Phase 1: Foundation
- ✅ **Namespace Isolation**: UTS, PID, Mount, IPC, Network namespaces
- ✅ **Resource Limits**: CPU, memory, and PID limits using cgroups
- ✅ **Filesystem Isolation**: Chroot, pivot_root, and overlay filesystems
- ✅ **Network Isolation**: Virtual ethernet pairs and bridge networking
- ✅ **Security Features** (Phase 1.2):
  - **Capabilities Management**: Drop/add Linux capabilities with safe defaults
  - **Seccomp Profiles**: Syscall filtering with Docker-compatible profiles
  - **LSM Support**: AppArmor and SELinux integration with auto-detection
- ✅ **Error Handling & Logging** (Phase 1.3):
  - **Structured Logging**: Configurable log levels with context-rich output
  - **Error Codes**: Unique error identifiers for programmatic handling
  - **Debug Mode**: Verbose logging for troubleshooting
  - **User-Friendly Errors**: Clear error messages with actionable hints
  - **Resource Cleanup**: Automatic cleanup on error paths
- ✅ **Comprehensive Testing**: Unit and integration tests with >70% coverage

## Why Build This?

Understanding containers from first principles helps developers:
- Grasp how Docker and Kubernetes work under the hood
- Debug container issues more effectively
- Build custom container solutions
- Learn Linux kernel features

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                   Container Runtime                 │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐         │
│  │Namespaces│  │ Cgroups  │  │  RootFS  │         │
│  └──────────┘  └──────────┘  └──────────┘         │
│                                                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐         │
│  │ Network  │  │  Image   │  │Container │         │
│  └──────────┘  └──────────┘  └──────────┘         │
│                                                     │
└─────────────────────────────────────────────────────┘
                      ▼
        ┌────────────────────────┐
        │   Linux Kernel         │
        │  (Syscalls, cgroups,   │
        │   namespaces, etc.)    │
        └────────────────────────┘
```

## Requirements

- **OS**: Linux (kernel 3.8+)
- **Go**: 1.16 or later
- **Privileges**: Root access (for namespace and cgroup operations)
- **Optional**: Docker or debootstrap for creating root filesystems

## Installation

### Quick Install (Recommended)

Use our installation script for the latest release:

```bash
curl -fsSL https://raw.githubusercontent.com/therealutkarshpriyadarshi/containr/main/scripts/install.sh | bash
```

### Manual Installation

Download the latest release for your platform:

```bash
# Linux (amd64)
curl -LO https://github.com/therealutkarshpriyadarshi/containr/releases/latest/download/containr-linux-amd64
chmod +x containr-linux-amd64
sudo mv containr-linux-amd64 /usr/local/bin/containr

# Linux (arm64)
curl -LO https://github.com/therealutkarshpriyadarshi/containr/releases/latest/download/containr-linux-arm64
chmod +x containr-linux-arm64
sudo mv containr-linux-arm64 /usr/local/bin/containr
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/therealutkarshpriyadarshi/containr.git
cd containr

# Build the binary (with version info)
make build

# Install system-wide (optional)
sudo make install
```

The binary will be available at `bin/containr` or `/usr/local/bin/containr` after installation.

### Verify Installation

```bash
containr version
```

## Quick Start

### Phase 4: Performance & Production

```bash
# Check version with full information
containr version

# Check version (short)
containr version --short

# Get version in JSON
containr version --json

# Run benchmarks
make bench

# Generate performance profiles
make profile-cpu
make profile-mem
make profile-trace

# View profiles
make profile-view-cpu

# Build release binaries
make release
```

### Phase 3: Advanced Features

```bash
# Create a custom network
sudo ./bin/containr network create --subnet 172.30.0.0/24 mynetwork

# Run container with port mapping
sudo ./bin/containr run -p 8080:80 --name web nginx

# View container events
sudo ./bin/containr events

# List networks
sudo ./bin/containr network ls

# Inspect network
sudo ./bin/containr network inspect mynetwork
```

### Phase 2: Enhanced Features

```bash
# Pull an image from Docker Hub
sudo ./bin/containr pull alpine

# Run container with named volume
sudo ./bin/containr volume create mydata
sudo ./bin/containr run --name myapp -v mydata:/data alpine /bin/sh

# List running containers
sudo ./bin/containr ps

# Execute command in running container
sudo ./bin/containr exec myapp ls /data

# View container logs
sudo ./bin/containr logs myapp

# Inspect container details
sudo ./bin/containr inspect myapp

# Stop and remove
sudo ./bin/containr stop myapp
sudo ./bin/containr rm myapp
```

### Basic Usage

```bash
# Run a simple command in an isolated container
sudo ./bin/containr run alpine /bin/echo "Hello from container!"

# Run an interactive shell
sudo ./bin/containr run alpine /bin/sh

# Run with custom name and hostname
sudo ./bin/containr run --name mycontainer --hostname myhost alpine /bin/sh
```

### With Volumes

```bash
# Bind mount host directory
sudo ./bin/containr run -v /host/path:/container/path alpine ls /container/path

# Read-only mount
sudo ./bin/containr run -v /host/config:/app/config:ro alpine cat /app/config/file

# Multiple volumes
sudo ./bin/containr run \
  -v /host/data:/app/data \
  -v /host/config:/app/config:ro \
  alpine /app/start.sh
```

### With Resource Limits

```bash
# Run with memory and CPU limits
sudo ./bin/containr run --memory 100m --cpus 0.5 alpine /bin/sh

# With PID limit
sudo ./bin/containr run --pids 50 alpine /bin/sh
```

### With Debug Mode

```bash
# Enable verbose logging for troubleshooting
sudo ./bin/containr run --debug alpine /bin/sh

# Set specific log level
sudo ./bin/containr run --log-level debug alpine /bin/bash

# View detailed execution steps
sudo ./bin/containr run --debug --log-level debug alpine /bin/sh -c "hostname"
```

## Project Structure

```
containr/
├── cmd/
│   └── containr/          # Main CLI application (Cobra-based)
├── pkg/
│   ├── benchmark/         # Benchmarking utilities (Phase 4.1)
│   ├── profiler/          # Profiling support (Phase 4.1)
│   ├── runtime/           # OCI runtime spec (Phase 4.3)
│   ├── version/           # Version management (Phase 4.4)
│   ├── container/         # Container creation and management
│   ├── namespace/         # Namespace handling (UTS, PID, Mount, User, etc.)
│   ├── cgroup/           # Cgroup resource limits
│   ├── rootfs/           # Filesystem operations (overlay, pivot_root)
│   ├── network/          # Network setup (veth, bridges, port mapping, modes) (Phase 3)
│   ├── image/            # Image import/export
│   ├── state/            # Container state persistence (Phase 2.1)
│   ├── volume/           # Volume management (Phase 2.2)
│   ├── registry/         # OCI registry client (Phase 2.3)
│   ├── userns/           # User namespace support (Phase 2.4)
│   ├── metrics/          # Metrics collection and monitoring (Phase 3.3)
│   ├── events/           # Event tracking and streaming (Phase 3.3)
│   ├── health/           # Health check monitoring (Phase 3.2)
│   ├── restart/          # Restart policy management (Phase 3.2)
│   ├── build/            # Dockerfile parser (Phase 3.4)
│   ├── capabilities/     # Capabilities management (Phase 1.2)
│   ├── seccomp/          # Seccomp profiles (Phase 1.2)
│   ├── security/         # LSM support (Phase 1.2)
│   ├── logger/           # Structured logging (Phase 1.3)
│   └── errors/           # Error handling with codes (Phase 1.3)
├── examples/             # Example programs
├── docs/                 # Documentation
│   ├── PHASE4.md         # Phase 4 feature documentation (NEW!)
│   ├── PHASE3.md         # Phase 3 feature documentation
│   ├── PHASE2.md         # Phase 2 feature documentation
│   ├── ARCHITECTURE.md   # Detailed architecture guide
│   ├── LOGGING.md        # Logging guide (Phase 1.3)
│   ├── ERROR_HANDLING.md # Error handling guide (Phase 1.3)
│   ├── SECURITY.md       # Security guide (Phase 1.2)
│   ├── TESTING.md        # Testing guide
│   └── GETTING_STARTED.md # Getting started guide
├── scripts/              # Installation and utility scripts
│   └── install.sh        # One-line installation script
├── Makefile             # Build automation
└── README.md            # This file
```

## How It Works

### 1. Namespaces - Process Isolation

Namespaces create isolated views of system resources:

```go
// Create process with isolated namespaces
cmd.SysProcAttr = &syscall.SysProcAttr{
    Cloneflags: syscall.CLONE_NEWUTS |  // Hostname
                syscall.CLONE_NEWPID |  // Process IDs
                syscall.CLONE_NEWNS |   // Mount points
                syscall.CLONE_NEWIPC |  // IPC resources
                syscall.CLONE_NEWNET,   // Network stack
}
```

**See it in action:**
```bash
# Process inside container sees isolated PID namespace
sudo ./bin/containr run /bin/sh -c "ps aux"
# Only shows processes inside the container!
```

### 2. Cgroups - Resource Limits

Control groups enforce resource constraints:

```go
// Set memory limit to 100MB
cgroup.Config{
    Name:        "mycontainer",
    MemoryLimit: 100 * 1024 * 1024,
    CPUShares:   512,
    PIDLimit:    100,
}
```

### 3. Filesystem Isolation

Multiple techniques for isolating filesystems:

- **Chroot**: Basic directory isolation
- **Pivot Root**: Change root mount point
- **Overlay FS**: Layered, copy-on-write filesystem

```
┌─────────────────────────────────┐
│     Overlay Filesystem          │
├─────────────────────────────────┤
│  Upper (writable)               │
│  ├── /etc/hostname (modified)   │
│  └── /app/data (new)            │
├─────────────────────────────────┤
│  Lower (read-only base layers)  │
│  ├── Layer 3: App files         │
│  ├── Layer 2: Dependencies      │
│  └── Layer 1: Base OS           │
└─────────────────────────────────┘
```

### 4. Network Isolation

Virtual ethernet pairs connect containers to host network:

```
Host                     Container
┌─────────┐             ┌─────────┐
│ bridge0 │◄───veth────►│  eth0   │
└────┬────┘             └─────────┘
     │
     ▼
  Internet
```

## Examples

### Example 1: Basic Container

```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/container"

config := &container.Config{
    Command:  []string{"/bin/sh"},
    Hostname: "mycontainer",
    Isolate:  true,
}

c := container.New(config)
c.RunWithSetup()
```

### Example 2: With Resource Limits

```go
import (
    "github.com/therealutkarshpriyadarshi/containr/pkg/container"
    "github.com/therealutkarshpriyadarshi/containr/pkg/cgroup"
)

// Create cgroup with limits
cg, _ := cgroup.New(&cgroup.Config{
    Name:        "limited-container",
    MemoryLimit: 50 * 1024 * 1024,  // 50MB
    CPUShares:   256,                // 1/4 of CPU
})
defer cg.Remove()

cg.AddCurrentProcess()

// Run container
c := container.New(&container.Config{
    Command: []string{"/bin/sh"},
})
c.RunWithSetup()
```

### Example 3: Import and Run Image

```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/image"

// Import a rootfs tarball
img, _ := image.Import("/path/to/rootfs.tar.gz", "myimage", "latest")

// Use in container
config := &container.Config{
    RootFS:  img.GetRootFS(),
    Command: []string{"/bin/sh"},
}
```

## API Reference

### Namespace Package

```go
// Get namespace flags for multiple types
flags := namespace.GetNamespaceFlags(
    namespace.UTS,
    namespace.PID,
    namespace.Mount,
)

// Create isolated process
config := &namespace.Config{
    Flags:   flags,
    Command: "/bin/sh",
    Args:    []string{},
}
namespace.CreateNamespaces(config)
```

### Cgroup Package

```go
// Create cgroup
cg, err := cgroup.New(&cgroup.Config{
    Name:        "container-1",
    MemoryLimit: 100 * 1024 * 1024,
    CPUShares:   512,
    PIDLimit:    50,
})

// Add process
cg.AddProcess(pid)

// Get statistics
stats, _ := cg.GetStats()
fmt.Printf("Memory usage: %d bytes\n", stats.MemoryUsage)

// Cleanup
cg.Remove()
```

### RootFS Package

```go
// Setup overlay filesystem
rootfs := rootfs.New(&rootfs.Config{
    MountPoint: "/tmp/container/root",
    Layers: []string{
        "/layers/base",
        "/layers/app",
    },
})

rootfs.Setup()
rootfs.PivotRoot()
defer rootfs.Teardown()
```

## Learning Path

1. **Start Simple**: Run basic commands with namespace isolation
2. **Add Resources**: Experiment with cgroup limits
3. **Filesystem**: Try different root filesystem configurations
4. **Networking**: Set up network isolation and bridges
5. **Images**: Import and manage container images
6. **Security**: Configure capabilities, seccomp, and LSM

## Security

Containr implements multiple security layers to isolate and protect containers:

### Capabilities Management

Control Linux capabilities to follow the principle of least privilege:

```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/capabilities"

// Use default safe capabilities
config := &container.Config{
    Capabilities: &capabilities.Config{
        Drop: []capabilities.Capability{
            capabilities.CAP_NET_RAW,  // Drop raw networking
        },
    },
}
```

### Seccomp Profiles

Filter dangerous syscalls with seccomp:

```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/seccomp"

// Use default restrictive profile
config := &container.Config{
    Seccomp: &seccomp.Config{
        Profile: seccomp.DefaultProfile(),  // Blocks dangerous syscalls
    },
}

// Or load custom profile
config := &container.Config{
    Seccomp: &seccomp.Config{
        ProfilePath: "/path/to/custom-profile.json",
    },
}
```

### LSM Support (AppArmor/SELinux)

Automatic detection and application of Mandatory Access Control:

```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/security"

// Auto-detect and use available LSM
config := &container.Config{
    Security: &security.Config{}, // Auto-detects AppArmor or SELinux
}

// Check what's available
lsm := security.DetectLSM()
fmt.Printf("Available LSM: %s\n", lsm)
```

### Complete Security Example

```go
config := &container.Config{
    Command:  []string{"/bin/sh"},
    Hostname: "secure-container",
    Isolate:  true,
    Capabilities: &capabilities.Config{
        Drop: []capabilities.Capability{
            capabilities.CAP_NET_RAW,
            capabilities.CAP_MKNOD,
        },
    },
    Seccomp: &seccomp.Config{
        Profile: seccomp.DefaultProfile(),
    },
    Security: &security.Config{
        // Auto-detect LSM
    },
}

c := container.New(config)
c.RunWithSetup()
```

For detailed security information, see [docs/SECURITY.md](docs/SECURITY.md).

## Troubleshooting

### "Operation not permitted"

You need root privileges:
```bash
sudo ./bin/containr run /bin/sh
```

**With debug mode**:
```bash
sudo ./bin/containr run --debug /bin/sh
```
This will show detailed logs to help identify the permission issue.

### Cgroups not working

Check if cgroups v2 is enabled:
```bash
mount | grep cgroup
# Should show /sys/fs/cgroup
```

Enable debug logging to see detailed cgroup operations:
```bash
sudo ./bin/containr run --debug /bin/sh
```

### Network issues

Ensure you have CAP_NET_ADMIN capability:
```bash
sudo ./bin/containr run /bin/sh
```

### Debugging container failures

Use debug mode to see detailed execution steps:
```bash
sudo ./bin/containr run --debug /bin/sh
```

Error messages now include helpful hints:
```
Error: [PERMISSION_DENIED] cannot create namespace
Hint: Try running with sudo or as root user
```

For more troubleshooting help, see:
- [Logging Guide](docs/LOGGING.md)
- [Error Handling Guide](docs/ERROR_HANDLING.md)

## Comparison with Docker

| Feature | Containr | Docker |
|---------|----------|--------|
| **Purpose** | Educational | Production |
| **Namespaces** | ✅ Full support | ✅ Full support |
| **Cgroups** | ✅ v1 & v2 support | ✅ Advanced limits |
| **Images** | ⚠️ Import/export only | ✅ Registry, layers, build |
| **Networking** | ✅ Basic isolation | ✅ Advanced (overlay, macvlan) |
| **Security** | ✅ **Phase 1.2 Complete** | ✅ Full support |
| - Capabilities | ✅ Drop/add with defaults | ✅ Full control |
| - Seccomp | ✅ Default + custom profiles | ✅ Full support |
| - AppArmor/SELinux | ✅ Auto-detect + apply | ✅ Full support |
| **Orchestration** | ❌ None (planned Phase 3) | ✅ Swarm, Kubernetes |
| **User Namespaces** | ❌ Planned Phase 2.4 | ✅ Rootless mode |

## Further Reading

### Containr Documentation
- 🛠️ [Phase 8 Documentation](docs/PHASE8.md) - Developer experience & advanced tooling (Phase 8) **NEW!**
- 🏢 [Phase 7 Documentation](docs/PHASE7.md) - Enterprise production features (Phase 7)
- ☁️ [Phase 6 Documentation](docs/PHASE6.md) - Cloud-native integration guide (Phase 6)
- 🌱 [Phase 5 Documentation](docs/PHASE5.md) - Community & growth guide (Phase 5)
- 🎯 [Phase 4 Documentation](docs/PHASE4.md) - Production polish guide (Phase 4)
- 🤝 [Contributing Guide](CONTRIBUTING.md) - How to contribute
- 🚀 [Phase 3 Documentation](docs/PHASE3.md) - Advanced features guide (Phase 3)
- 📦 [Phase 2 Documentation](docs/PHASE2.md) - Feature completeness guide (Phase 2)
- 📖 [Architecture Documentation](docs/ARCHITECTURE.md) - Detailed architecture overview
- 🔒 [Security Guide](docs/SECURITY.md) - Comprehensive security documentation
- 📝 [Logging Guide](docs/LOGGING.md) - Structured logging and debug mode (Phase 1.3)
- ⚠️ [Error Handling Guide](docs/ERROR_HANDLING.md) - Error codes and best practices (Phase 1.3)

### External Resources
- 🔧 [Linux Namespaces Man Page](https://man7.org/linux/man-pages/man7/namespaces.7.html)
- 📚 [Cgroups Documentation](https://www.kernel.org/doc/Documentation/cgroup-v2.txt)
- 🔐 [Linux Capabilities](https://man7.org/linux/man-pages/man7/capabilities.7.html)
- 🛡️ [Seccomp](https://www.kernel.org/doc/html/latest/userspace-api/seccomp_filter.html)
- 🐋 [OCI Runtime Spec](https://github.com/opencontainers/runtime-spec)

## Contributing

Contributions are welcome! We have a vibrant community and comprehensive guidelines to help you get started.

### Quick Links

- 📖 [Contributing Guide](CONTRIBUTING.md) - How to contribute
- 💬 [GitHub Discussions](https://github.com/therealutkarshpriyadarshi/containr/discussions) - Ask questions
- 🐛 [Issue Tracker](https://github.com/therealutkarshpriyadarshi/containr/issues) - Report bugs
- 📚 [Tutorial](docs/tutorials/01-community-contribution.md) - First contribution guide
- 📋 [Code of Conduct](CODE_OF_CONDUCT.md) - Community guidelines
- 🏛️ [Governance](GOVERNANCE.md) - How we make decisions
- 🙋 [Support](SUPPORT.md) - Getting help

### Quick Start for Contributors

```bash
# Fork and clone the repository
git clone https://github.com/YOUR_USERNAME/containr.git
cd containr

# Install development tools
make install-tools

# Run tests
make test-unit

# Run benchmarks
make bench

# Format and lint code
make fmt
make lint

# Run all pre-commit checks
make pre-commit
```

### Future Enhancements

While all 5 phases are complete, potential future improvements include:

- [ ] CRI (Container Runtime Interface) support
- [ ] BuildKit integration for advanced builds
- [ ] Enhanced metrics (Prometheus/Grafana integration)
- [ ] Performance optimizations for extreme scale
- [ ] Windows container support (educational)
- [ ] Advanced orchestration features

## License

MIT License - feel free to use this for learning and education.

## Acknowledgments

Inspired by:
- Docker and containerd
- Linux kernel developers
- ["Containers from Scratch"](https://ericchiang.github.io/post/containers-from-scratch/) by Eric Chiang
- runc and the OCI runtime specification

---

**Note**: This is an educational project. For production use, consider established runtimes like Docker, containerd, or CRI-O.

## Getting Help

- 📖 Read the [architecture docs](docs/ARCHITECTURE.md)
- 🐛 [Open an issue](https://github.com/therealutkarshpriyadarshi/containr/issues)
- 💬 Check existing issues for solutions

**Happy containerizing! 🎉**
