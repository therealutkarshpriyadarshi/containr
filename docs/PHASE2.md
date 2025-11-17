# Phase 2: Feature Completeness Documentation

**Version:** 2.0.0
**Status:** Complete
**Completion Date:** November 17, 2025

## Overview

Phase 2 of Containr implements comprehensive container runtime features, making it practical for real-world educational use. This phase focuses on feature completeness with a Docker-like CLI, volume management, registry integration, and rootless container support.

## What's New in Phase 2

### 1. Enhanced CLI (Phase 2.1) ✅

Containr now features a complete CLI built with Cobra, providing Docker-like commands for container lifecycle management.

#### Container Lifecycle Commands

**`containr run`** - Run a command in a new container
```bash
# Basic usage
sudo containr run alpine /bin/sh

# With name and volumes
sudo containr run --name mycontainer -v /host:/container alpine /bin/sh

# With resource limits
sudo containr run --memory 100m --cpus 0.5 alpine stress

# Auto-remove on exit
sudo containr run --rm alpine /bin/echo "Hello World"

# Detached mode
sudo containr run -d --name webserver nginx
```

**`containr ps`** - List containers
```bash
# List running containers
sudo containr ps

# List all containers (including stopped)
sudo containr ps -a

# Only show container IDs
sudo containr ps -q
```

**`containr start/stop`** - Manage container state
```bash
# Start a stopped container
sudo containr start mycontainer

# Stop a running container
sudo containr stop mycontainer

# Stop multiple containers
sudo containr stop container1 container2
```

**`containr rm`** - Remove containers
```bash
# Remove a stopped container
sudo containr rm mycontainer

# Force remove a running container
sudo containr rm -f mycontainer

# Remove multiple containers
sudo containr rm container1 container2 container3
```

**`containr exec`** - Execute command in running container
```bash
# Execute interactive shell
sudo containr exec -it mycontainer /bin/sh

# Execute single command
sudo containr exec mycontainer ls /app

# Execute as specific user
sudo containr exec --user nobody mycontainer whoami
```

**`containr logs`** - View container logs
```bash
# View logs
sudo containr logs mycontainer

# Follow log output
sudo containr logs -f mycontainer

# Show last N lines
sudo containr logs --tail 100 mycontainer
```

#### Image Management Commands

**`containr pull`** - Pull image from registry
```bash
# Pull from Docker Hub
sudo containr pull alpine

# Pull specific tag
sudo containr pull alpine:3.14

# Pull from custom registry
sudo containr pull gcr.io/project/image:tag

# Quiet mode
sudo containr pull -q ubuntu:latest
```

**`containr images`** - List images
```bash
# List all images
sudo containr images

# Show all images including intermediates
sudo containr images -a

# Only show image IDs
sudo containr images -q
```

**`containr rmi`** - Remove images
```bash
# Remove an image
sudo containr rmi alpine:latest

# Force remove
sudo containr rmi -f myimage:v1.0

# Remove multiple images
sudo containr rmi image1 image2 image3
```

**`containr import/export`** - Import/export images
```bash
# Import from tarball
sudo containr import rootfs.tar.gz myimage:latest

# Export container filesystem
sudo containr export mycontainer > mycontainer.tar
```

#### Inspection & Monitoring Commands

**`containr inspect`** - Display detailed container information
```bash
# Inspect a container (JSON output)
sudo containr inspect mycontainer

# Inspect multiple containers
sudo containr inspect container1 container2

# Custom format (template)
sudo containr inspect --format '{{.State}}' mycontainer
```

**`containr stats`** - Display resource usage statistics
```bash
# Show stats for running containers
sudo containr stats

# Show stats for specific container
sudo containr stats mycontainer

# Disable streaming (single snapshot)
sudo containr stats --no-stream mycontainer
```

**`containr top`** - Display running processes
```bash
# Show processes in container
sudo containr top mycontainer

# With custom ps options
sudo containr top mycontainer aux
```

### 2. Volume Management (Phase 2.2) ✅

Containr now supports persistent data storage with bind mounts, named volumes, and tmpfs mounts.

#### Volume Commands

**`containr volume create`** - Create a named volume
```bash
# Create a volume
sudo containr volume create myvolume

# Volume with auto-generated name
sudo containr volume create
```

**`containr volume ls`** - List volumes
```bash
# List all volumes
sudo containr volume ls

# Only show volume names
sudo containr volume ls -q
```

**`containr volume rm`** - Remove volumes
```bash
# Remove a volume
sudo containr volume rm myvolume

# Force remove
sudo containr volume rm -f myvolume

# Remove multiple volumes
sudo containr volume rm vol1 vol2 vol3
```

**`containr volume inspect`** - Inspect volumes
```bash
# Inspect volume details (JSON output)
sudo containr volume inspect myvolume

# Inspect multiple volumes
sudo containr volume inspect vol1 vol2
```

**`containr volume prune`** - Remove unused volumes
```bash
# Remove all unused volumes
sudo containr volume prune

# Without confirmation prompt
sudo containr volume prune -f
```

#### Using Volumes with Containers

**Bind Mounts**
```bash
# Mount host directory to container
sudo containr run -v /host/path:/container/path alpine ls /container/path

# Read-only mount
sudo containr run -v /host/path:/container/path:ro alpine cat /container/path/file.txt

# Multiple volumes
sudo containr run \
  -v /host/data:/app/data \
  -v /host/config:/app/config:ro \
  alpine /app/start.sh
```

**Named Volumes**
```bash
# Create and use named volume
sudo containr volume create appdata
sudo containr run -v appdata:/app/data alpine touch /app/data/test.txt

# Volume persists across container restarts
sudo containr run -v appdata:/app/data alpine ls /app/data
```

#### Volume Types

1. **Bind Mounts** - Mount host directory/file into container
   - Source: Absolute path on host
   - Example: `-v /host/path:/container/path`

2. **Named Volumes** - Managed by containr
   - Source: Volume name
   - Stored in: `/var/lib/containr/volumes/`
   - Example: `-v myvolume:/container/path`

3. **tmpfs Mounts** - In-memory filesystem
   - Temporary, fast storage
   - Cleared on container stop
   - Good for sensitive data

### 3. Registry Integration (Phase 2.3) ✅

Pull container images from Docker Hub and OCI-compatible registries.

#### Supported Features

- **Docker Hub Integration**: Pull official and user images
- **Custom Registries**: Support for private registries
- **OCI Image Format**: Full compliance with OCI image spec
- **Layer Extraction**: Automatic extraction to rootfs
- **Authentication**: Support for registry authentication (token-based)
- **Digest Verification**: Verify image integrity

#### Pulling Images

```bash
# Pull official image from Docker Hub
sudo containr pull alpine

# Pull specific version
sudo containr pull alpine:3.14

# Pull from custom registry
sudo containr pull gcr.io/my-project/my-image:v1.0

# Pull with digest
sudo containr pull alpine@sha256:abc123...
```

#### Image Storage

Images are stored in `/var/lib/containr/images/` with the following structure:

```
/var/lib/containr/images/
├── library/alpine/
│   ├── config.json          # Image configuration
│   ├── manifest.json        # Layer manifest
│   └── layers/              # Layer tarballs
│       ├── layer-0.tar.gz
│       ├── layer-1.tar.gz
│       └── layer-2.tar.gz
└── user/myimage/
    └── ...
```

Extracted rootfs are stored in `/var/lib/containr/rootfs/`:

```
/var/lib/containr/rootfs/
├── library-alpine-latest/   # Extracted filesystem
│   ├── bin/
│   ├── etc/
│   ├── lib/
│   └── ...
└── user-myimage-v1.0/
    └── ...
```

#### Registry Client API

The registry package provides a programmatic API:

```go
import (
    "github.com/therealutkarshpriyadarshi/containr/pkg/registry"
)

// Create registry client
client := registry.DefaultClient()

// Parse image reference
ref, err := registry.ParseImageReference("alpine:latest")

// Pull image
opts := &registry.PullOptions{
    DestDir: "/var/lib/containr/images/alpine",
    Verbose: true,
}
err = client.Pull(ref, opts)

// Extract to rootfs
err = registry.ExtractImageToRootFS("/var/lib/containr/images/alpine", "/rootfs/alpine")

// Load image config
config, err := registry.LoadImageConfig("/var/lib/containr/images/alpine")
```

### 4. User Namespace Support (Phase 2.4) ✅

Run containers without root privileges using user namespaces for enhanced security.

#### Rootless Containers

User namespaces allow running containers as a non-root user by remapping UIDs/GIDs.

**Benefits:**
- Enhanced security: Containers can't escalate privileges on host
- Multi-tenancy: Different users can run containers without interference
- Reduced attack surface: Root inside container is unprivileged on host

#### Configuration

**System Requirements:**
1. User namespaces enabled in kernel
2. Subordinate UID/GID ranges configured

**Check if supported:**
```bash
# Check if user namespaces are available
ls /proc/self/ns/user

# Check subordinate UID range
grep $USER /etc/subuid

# Check subordinate GID range
grep $USER /etc/subgid
```

**Configure subordinate IDs:**
```bash
# Add UID range for user (65536 IDs starting from 100000)
echo "$USER:100000:65536" | sudo tee -a /etc/subuid

# Add GID range for user
echo "$USER:100000:65536" | sudo tee -a /etc/subgid
```

#### Usage

**Run rootless container:**
```bash
# Run without sudo (as regular user)
containr run --userns alpine /bin/sh

# Inside container, you're root (UID 0)
# But on the host, you're remapped to your user's subuid
```

**User Namespace API:**

```go
import (
    "github.com/therealutkarshpriyadarshi/containr/pkg/userns"
)

// Get default rootless config
config, err := userns.RootlessConfig()

// Or create custom mappings
config := &userns.Config{
    UIDMappings: []userns.IDMap{
        {ContainerID: 0, HostID: 1000, Size: 1},      // Map container root to host user
        {ContainerID: 1, HostID: 100000, Size: 65536}, // Map rest of range
    },
    GIDMappings: []userns.IDMap{
        {ContainerID: 0, HostID: 1000, Size: 1},
        {ContainerID: 1, HostID: 100000, Size: 65536},
    },
    RootlessMode: true,
}

// Validate configuration
err = userns.ValidateConfig(config)

// Set up user namespace for process
err = userns.SetupUserNamespace(pid, config)
```

#### ID Mapping Example

```
Container UID/GID    →    Host UID/GID
─────────────────────────────────────
0 (root)            →    1000 (your user)
1                   →    100000
2                   →    100001
...
65536               →    165535
```

#### Limitations

- Some features require additional configuration:
  - Network setup may need `slirp4netns` for non-root networking
  - Certain mount operations have restrictions
  - Some capabilities are unavailable

- File ownership:
  - Files created in container appear with subuid/subgid on host
  - Use bind mounts with appropriate permissions

## Architecture Changes

### New Packages

1. **`pkg/state`** - Container state persistence
   - Stores container metadata
   - Tracks container lifecycle
   - Enables commands like `ps`, `start`, `stop`

2. **`pkg/volume`** - Volume management
   - Named volume creation and management
   - Bind mount handling
   - tmpfs support
   - Volume lifecycle management

3. **`pkg/registry`** - OCI registry client
   - Image pulling from registries
   - Authentication handling
   - Layer download and extraction
   - Manifest parsing

4. **`pkg/userns`** - User namespace support
   - UID/GID mapping configuration
   - Rootless container support
   - Subuid/subgid parsing
   - User namespace setup

### Directory Structure

```
/var/lib/containr/
├── state/              # Container state files
│   └── <container-id>/
│       └── state.json
├── volumes/            # Named volumes
│   ├── <volume-name>/
│   └── <volume-name>.json
├── images/             # Pulled images
│   └── <repo>/
│       ├── config.json
│       ├── manifest.json
│       └── layers/
└── rootfs/             # Extracted filesystems
    └── <image>/
```

## Migration from Phase 1

Phase 2 maintains backward compatibility with Phase 1. Existing code continues to work, with enhanced functionality available through new commands and options.

**Key Changes:**
- Old `containr run` syntax still works
- New commands added without breaking existing workflows
- Enhanced error handling provides better diagnostics

## Performance Considerations

### Resource Usage

- **Memory**: ~50MB per container overhead
- **Storage**: Images stored once, shared across containers
- **Startup Time**: <2 seconds for cached images

### Optimization Tips

1. **Use named volumes** for persistent data (faster than bind mounts)
2. **Pull images once** and reuse across containers
3. **Clean up unused volumes** with `volume prune`
4. **Remove stopped containers** with `rm` to free resources

## Security Best Practices

### Default Security

Phase 2 maintains all Phase 1 security features:
- Capability dropping
- Seccomp profiles
- AppArmor/SELinux support

### Additional Recommendations

1. **Use rootless containers** when possible
2. **Use read-only volumes** for configuration files
3. **Limit resource usage** with `--memory` and `--cpus`
4. **Use named volumes** instead of bind mounts to avoid host filesystem access
5. **Pull images from trusted registries**

### Volume Security

```bash
# Read-only configuration
sudo containr run -v /host/config:/app/config:ro app

# tmpfs for sensitive data (cleared on stop)
sudo containr run --tmpfs /app/secrets app

# Named volume (isolated from host)
sudo containr volume create appdata
sudo containr run -v appdata:/app/data app
```

## Troubleshooting

### Common Issues

**1. Volume mount permission denied**
```bash
# Check volume exists
sudo containr volume ls

# Check file permissions on host
ls -la /host/path

# Use rootless mode for better permission handling
containr run --userns -v /host:/container alpine
```

**2. Image pull fails**
```bash
# Check network connectivity
ping registry-1.docker.io

# Use verbose mode
sudo containr pull -v alpine

# Check authentication for private registries
cat ~/.docker/config.json
```

**3. User namespace not supported**
```bash
# Check kernel support
grep CONFIG_USER_NS /boot/config-$(uname -r)

# Enable user namespaces (if disabled)
sudo sysctl kernel.unprivileged_userns_clone=1

# Check subuid/subgid
cat /etc/subuid /etc/subgid
```

**4. Container not found**
```bash
# List all containers
sudo containr ps -a

# Check container name/ID
sudo containr inspect <container>
```

## Examples

### Complete Workflow Example

```bash
# 1. Pull an image
sudo containr pull alpine:latest

# 2. Create a named volume
sudo containr volume create mydata

# 3. Run container with volume
sudo containr run \
  --name myapp \
  -v mydata:/data \
  -v /host/config:/app/config:ro \
  --memory 100m \
  alpine /bin/sh

# 4. Check container status
sudo containr ps

# 5. Execute command in running container
sudo containr exec myapp ls /data

# 6. View logs
sudo containr logs myapp

# 7. Stop and remove
sudo containr stop myapp
sudo containr rm myapp

# 8. Clean up volumes
sudo containr volume rm mydata
```

### Multi-Container Application

```bash
# Database container with persistent volume
sudo containr volume create db-data
sudo containr run -d --name database \
  -v db-data:/var/lib/postgresql/data \
  postgres:latest

# Application container linked to database
sudo containr run -d --name webapp \
  --link database:db \
  -p 8080:8080 \
  myapp:latest

# Check status
sudo containr ps

# View logs
sudo containr logs webapp
sudo containr logs database
```

## API Reference

See individual package documentation:
- [State Management](../pkg/state/state.go)
- [Volume Management](../pkg/volume/volume.go)
- [Registry Integration](../pkg/registry/registry.go)
- [User Namespaces](../pkg/userns/userns.go)

## Future Enhancements (Phase 3)

Phase 2 lays the groundwork for Phase 3 features:
- Enhanced networking (CNI, port mapping, DNS)
- Container orchestration (multi-container management)
- Monitoring and observability (Prometheus metrics)
- Build capabilities (Dockerfile support)

## Contributing

Contributions to Phase 2 features are welcome! Areas for improvement:
- Enhanced registry authentication
- More volume drivers
- Better rootless networking
- Performance optimizations

## Changelog

### v2.0.0 (Phase 2 Complete)
- ✅ Enhanced CLI with Cobra framework
- ✅ Container lifecycle management (create, start, stop, rm, ps, logs, exec)
- ✅ Image management (pull, images, rmi, import, export)
- ✅ Inspection commands (inspect, stats, top)
- ✅ Volume management (create, ls, rm, inspect, prune)
- ✅ Registry integration (Docker Hub and OCI registries)
- ✅ User namespace support (rootless containers)
- ✅ Comprehensive test suite (70%+ coverage)
- ✅ Enhanced documentation

### v1.3.0 (Phase 1.3 Complete)
- Error handling and logging
- Structured logging with configurable levels
- Error codes and user-friendly messages

### v1.2.0 (Phase 1.2 Complete)
- Security foundations
- Capabilities management
- Seccomp profiles
- AppArmor/SELinux support

### v1.1.0 (Phase 1.1 Complete)
- Core containerization features
- Namespaces, cgroups, rootfs
- Basic CLI

---

**Phase 2 Status: Complete ✅**
**Next: Phase 3 - Advanced Features**
