# Containr 🚀

A minimal container runtime built from scratch using Linux primitives. This project demonstrates the core concepts behind Docker and other container runtimes by implementing process isolation using namespaces, resource limits with cgroups, and filesystem isolation.

**🎉 Phase 2 Complete!** Containr now features a full Docker-like CLI with volume management, registry integration, and rootless container support.

## Features

### ✨ Phase 2: Feature Completeness (NEW!)
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

### Build from Source

```bash
# Clone the repository
git clone https://github.com/therealutkarshpriyadarshi/containr.git
cd containr

# Build the binary
make build

# Install system-wide (optional)
sudo make install
```

The binary will be available at `bin/containr` or `/usr/local/bin/containr` after installation.

## Quick Start

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
│   ├── container/         # Container creation and management
│   ├── namespace/         # Namespace handling (UTS, PID, Mount, User, etc.)
│   ├── cgroup/           # Cgroup resource limits
│   ├── rootfs/           # Filesystem operations (overlay, pivot_root)
│   ├── network/          # Network setup (veth, bridges)
│   ├── image/            # Image import/export
│   ├── state/            # Container state persistence (Phase 2.1)
│   ├── volume/           # Volume management (Phase 2.2)
│   ├── registry/         # OCI registry client (Phase 2.3)
│   ├── userns/           # User namespace support (Phase 2.4)
│   ├── capabilities/     # Capabilities management (Phase 1.2)
│   ├── seccomp/          # Seccomp profiles (Phase 1.2)
│   ├── security/         # LSM support (Phase 1.2)
│   ├── logger/           # Structured logging (Phase 1.3)
│   └── errors/           # Error handling with codes (Phase 1.3)
├── examples/             # Example programs
├── docs/                 # Documentation
│   ├── ARCHITECTURE.md   # Detailed architecture guide
│   ├── PHASE2.md         # Phase 2 feature documentation (NEW!)
│   ├── LOGGING.md        # Logging guide (Phase 1.3)
│   ├── ERROR_HANDLING.md # Error handling guide (Phase 1.3)
│   ├── SECURITY.md       # Security guide (Phase 1.2)
│   ├── TESTING.md        # Testing guide
│   └── GETTING_STARTED.md # Getting started guide
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

Contributions are welcome! Areas for improvement:

- [ ] User namespace support for rootless containers
- [ ] Seccomp profiles for syscall filtering
- [ ] Volume management
- [ ] Image registry support (pull/push)
- [ ] Container networking improvements
- [ ] Logging and monitoring
- [ ] Better error handling
- [ ] Unit tests

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
