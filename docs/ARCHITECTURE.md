# Containr Architecture

## Overview

Containr is a minimal container runtime built from scratch using Linux primitives. It demonstrates the core concepts behind Docker and other container runtimes.

## Core Components

### 1. Namespaces (`pkg/namespace`)

Namespaces provide process isolation by creating separate views of system resources:

- **UTS Namespace**: Isolates hostname and domain name
- **PID Namespace**: Provides independent process ID space
- **Mount Namespace**: Isolates filesystem mount points
- **IPC Namespace**: Isolates inter-process communication resources
- **Network Namespace**: Provides independent network stack
- **User Namespace**: Isolates user and group IDs

#### Implementation

```go
// Create process with isolated namespaces
cmd.SysProcAttr = &syscall.SysProcAttr{
    Cloneflags: syscall.CLONE_NEWUTS |
                syscall.CLONE_NEWPID |
                syscall.CLONE_NEWNS |
                syscall.CLONE_NEWIPC |
                syscall.CLONE_NEWNET,
}
```

### 2. Cgroups (`pkg/cgroup`)

Control Groups (cgroups) enforce resource limits:

- **Memory**: Limit memory usage
- **CPU**: Control CPU allocation
- **PIDs**: Limit number of processes
- **Block I/O**: Control disk I/O

#### Resource Hierarchy

```
/sys/fs/cgroup/
├── memory/
│   └── container-name/
│       ├── memory.limit_in_bytes
│       ├── memory.usage_in_bytes
│       └── cgroup.procs
├── cpu/
│   └── container-name/
│       ├── cpu.shares
│       └── cgroup.procs
└── pids/
    └── container-name/
        ├── pids.max
        └── cgroup.procs
```

### 3. Filesystem Isolation (`pkg/rootfs`)

Multiple techniques for filesystem isolation:

#### a. Chroot
Basic filesystem isolation using `chroot()` syscall.

#### b. Pivot Root
Advanced isolation by changing the root mount:
```go
syscall.PivotRoot(newRoot, oldRoot)
```

#### c. Overlay Filesystem
Layered filesystem for efficient storage:
```
lowerdir: Read-only base layers
upperdir: Writable layer for changes
workdir:  Working directory for overlay
```

Benefits:
- Space efficient (shared base layers)
- Fast container creation
- Copy-on-write semantics

### 4. Networking (`pkg/network`)

Container networking using virtual ethernet pairs:

```
Host Network Namespace          Container Network Namespace
┌──────────────────┐           ┌───────────────────┐
│                  │           │                   │
│  ┌────────┐      │           │      ┌────────┐  │
│  │ Bridge │◄─────┼───────────┼─────►│ eth0   │  │
│  └────────┘      │  veth     │      └────────┘  │
│       │          │   pair    │                   │
│       ▼          │           │                   │
│   Internet       │           │                   │
└──────────────────┘           └───────────────────┘
```

Components:
- **Bridge**: Virtual switch on host
- **veth pair**: Virtual ethernet cable connecting namespaces
- **NAT**: Network address translation for internet access

### 5. Container Management (`pkg/container`)

High-level container lifecycle management:

1. **Create**: Set up configuration
2. **Start**: Fork process with namespaces
3. **Setup**: Configure resources inside container
4. **Run**: Execute container process
5. **Stop**: Clean up resources

### 6. Image Management (`pkg/image`)

Container image handling:

- **Import**: Load filesystem from tarball
- **Export**: Save container as tarball
- **Layers**: Support for layered images
- **Manifest**: Image metadata (OCI-compatible format)

## Process Flow

### Container Creation

```
1. Parse command-line arguments
   ↓
2. Create cgroup for resource limits
   ↓
3. Fork process with namespace flags
   ↓
4. In child process:
   - Set hostname
   - Mount /proc
   - Setup root filesystem
   - Configure network
   ↓
5. Execute container command
   ↓
6. Wait for completion
   ↓
7. Clean up resources
```

### Namespace Isolation Flow

```go
// Parent process
cmd := exec.Command("/proc/self/exe", "child", ...)
cmd.SysProcAttr = &syscall.SysProcAttr{
    Cloneflags: CLONE_NEWUTS | CLONE_NEWPID | ...,
}
cmd.Start()
cmd.Wait()

// Child process (re-executed)
if os.Args[1] == "child" {
    setupNamespace()
    exec.Command(actualCommand).Run()
}
```

## Security Considerations

### Capabilities

Linux capabilities provide fine-grained privilege control:
- Drop unnecessary capabilities
- Run with minimal privileges
- Use user namespaces for unprivileged containers

### Seccomp

System call filtering:
- Whitelist allowed syscalls
- Block dangerous operations
- Reduce attack surface

### AppArmor/SELinux

Mandatory access control:
- Restrict file access
- Limit network operations
- Enforce security policies

## Comparison with Docker

| Feature | Containr | Docker |
|---------|----------|--------|
| Namespaces | ✓ Basic | ✓ Full |
| Cgroups | ✓ Basic | ✓ Full |
| Overlay FS | ✓ Basic | ✓ Full |
| Networking | ✓ Basic | ✓ Advanced |
| Image Format | Simple | OCI-compliant |
| Registry | ✗ | ✓ Docker Hub |
| Volumes | ✗ | ✓ |
| Orchestration | ✗ | ✓ Swarm |

## Future Enhancements

1. **User Namespaces**: Run containers as non-root
2. **Seccomp Profiles**: System call filtering
3. **Volume Management**: Persistent storage
4. **Registry Support**: Pull/push images
5. **Logging**: Container log management
6. **Health Checks**: Container monitoring
7. **Resource Metrics**: CPU/memory usage tracking
8. **Multi-container Networks**: Container-to-container communication

## References

- [Linux Namespaces](https://man7.org/linux/man-pages/man7/namespaces.7.html)
- [Control Groups](https://www.kernel.org/doc/Documentation/cgroup-v2.txt)
- [Overlay Filesystem](https://www.kernel.org/doc/Documentation/filesystems/overlayfs.txt)
- [OCI Runtime Specification](https://github.com/opencontainers/runtime-spec)
