# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.0] - 2025-11-17

### Phase 6: Cloud-Native Integration & Advanced Runtime

#### Added - CRI (Container Runtime Interface)
- **CRI Service** (`pkg/cri`)
  - RuntimeService implementation for pod and container lifecycle
  - ImageService implementation for image operations
  - Pod sandbox support with full Kubernetes abstraction
  - Container creation, start, stop, and removal via CRI
  - Image pull, list, and remove operations
  - Filesystem usage reporting
- **CRI API Support**
  - Pod sandbox configuration and status
  - Container configuration with Linux-specific options
  - Security context support (capabilities, seccomp, AppArmor)
  - Resource limits (CPU, memory) via CRI
  - Mount and device configuration
  - Network and namespace options

#### Added - Plugin System
- **Plugin Framework** (`pkg/plugin`)
  - Plugin interface with lifecycle management
  - Plugin types: runtime, network, storage, logging, metrics
  - Plugin manager with enable/disable/configure capabilities
  - BasePlugin implementation for common functionality
  - Plugin health checking
  - Dynamic plugin configuration
- **Plugin Management**
  - Plugin registration and unregistration
  - Filter plugins by type
  - Plugin metadata and versioning
  - Thread-safe plugin operations

#### Added - Snapshot Support
- **Snapshot Interface** (`pkg/snapshot`)
  - Snapshotter interface for multiple drivers
  - Snapshot types: active, committed, view
  - Snapshot metadata management
  - Snapshot lifecycle (prepare, commit, remove)
  - Snapshot walking and filtering
  - Storage usage tracking
- **Overlay2 Driver**
  - Overlay2 filesystem snapshotter implementation
  - Parent-child snapshot relationships
  - Copy-on-write optimization
  - Snapshot export and import capabilities
  - Metadata persistence
  - Size calculation and reporting
- **Snapshot Operations**
  - Create active snapshots from parent
  - Commit active snapshots to immutable state
  - Create read-only views of snapshots
  - Update snapshot labels and metadata
  - Calculate and report snapshot usage

#### Added - Complete Build Engine
- **Builder Package** (`pkg/build`)
  - Complete Dockerfile build implementation
  - BuildContext with file hashing and indexing
  - Multi-stage build support
  - Build cache with intelligent invalidation
  - Layer generation and management
  - OCI image manifest generation
- **Build Features**
  - Support for all Dockerfile instructions (FROM, RUN, COPY, ADD, ENV, etc.)
  - Build argument substitution
  - Build context hashing for cache keys
  - Layer-by-layer execution with caching
  - Target stage selection for multi-stage builds
  - Build configuration (tags, args, platforms)
- **Build Cache**
  - Cache key generation from parent, instruction, and context
  - Cache lookup and storage
  - Cache pruning by age
  - Per-instruction caching
  - Build context-aware cache invalidation

#### Added - Documentation
- **Phase 6 Documentation** (`docs/PHASE6.md`)
  - CRI implementation guide
  - Plugin development guide
  - Snapshot usage and driver documentation
  - Build engine architecture and usage
  - Security considerations for Phase 6 features
  - Examples and tutorials for all features

#### Added - Tests
- **CRI Tests** (`pkg/cri/cri_test.go`)
  - Pod sandbox lifecycle tests
  - Container lifecycle tests
  - Image service tests
  - Filter and list operation tests
- **Plugin Tests** (`pkg/plugin/plugin_test.go`)
  - Plugin registration and lifecycle tests
  - Plugin manager tests
  - Concurrent plugin operation tests
  - Health check tests
- **Snapshot Tests** (`pkg/snapshot/snapshot_test.go`)
  - Overlay2 driver tests
  - Snapshot lifecycle tests
  - Parent chain tests
  - Metadata persistence tests
- **Builder Tests** (`pkg/build/builder_test.go`)
  - Build context tests
  - Build cache tests
  - Multi-stage build tests
  - Layer generation tests

### Phase 5: Community & Growth
- Community health files and governance
- Educational resources and examples
- Sustainability infrastructure

## [1.0.0] - 2025-11-17

### Phase 4: Production Polish

#### Added - Performance Optimization
- **Benchmarking Package** (`pkg/benchmark`)
  - Comprehensive benchmark suite for all operations
  - Benchmark result formatting and comparison
  - Performance regression testing utilities
  - Quick timing functions for profiling
- **Profiling Package** (`pkg/profiler`)
  - CPU profiling support
  - Memory profiling with statistics
  - Execution trace profiling
  - Integrated profiling configuration
- **Performance Tests**
  - Container startup benchmarks
  - Network operation benchmarks
  - Cgroup operation benchmarks
  - Scalability tests (100+ containers)

#### Added - OCI Runtime Compliance
- **Runtime Package** (`pkg/runtime`)
  - Full OCI Runtime Specification 1.0.2 implementation
  - OCI container state management
  - OCI bundle format support
  - Runtime configuration (config.json) handling
  - Spec validation and loading
- **OCI Features**
  - Container lifecycle states (creating, created, running, stopped)
  - Process configuration
  - Resource limits specification
  - Namespace configuration
  - Mount point specification

#### Added - Version Management
- **Version Package** (`pkg/version`)
  - Semantic versioning support
  - Build metadata (git commit, build date, Go version)
  - Version formatting (full, short, user-agent)
  - JSON output support
- **CLI Commands**
  - `containr version` - Full version information
  - `containr version --short` - Short version
  - `containr version --json` - JSON output

#### Added - Release & Distribution
- **Release Automation**
  - GitHub Actions release workflow
  - Multi-platform builds (linux/amd64, linux/arm64, linux/arm)
  - Automated checksums (SHA256)
  - Release notes generation
- **Installation**
  - Installation script (`scripts/install.sh`)
  - Binary distribution via GitHub Releases
  - Makefile targets for building and installation

#### Added - Documentation
- **Phase 4 Documentation** (`docs/PHASE4.md`)
  - Performance optimization guide
  - OCI compliance documentation
  - Version management guide
  - Release process documentation
- **Contributing Guide** (`CONTRIBUTING.md`)
  - Development environment setup
  - Code style guidelines
  - Testing requirements
  - PR process

### Phase 3: Advanced Features

#### Added - Enhanced Networking
- **Network Modes**
  - Bridge networking (default)
  - Host networking
  - None (no networking)
  - Container networking (share with another container)
- **Port Mapping**
  - TCP/UDP port exposure
  - Multiple port mappings
  - iptables integration
  - Random port assignment
- **DNS Resolution**
  - Automatic DNS configuration
  - Custom DNS servers
  - Container hostname resolution
- **Network Commands**
  - `containr network create` - Create custom networks
  - `containr network ls` - List networks
  - `containr network rm` - Remove networks
  - `containr network inspect` - Inspect network details

#### Added - Monitoring & Observability
- **Metrics Collection** (`pkg/metrics`)
  - CPU usage metrics
  - Memory usage statistics
  - Network I/O metrics
  - Disk I/O statistics
  - PID count and limits
- **Events System** (`pkg/events`)
  - Container lifecycle events
  - Event streaming API
  - Event filtering and querying
  - Time-based event queries
- **Commands**
  - `containr events` - Stream container events
  - `containr stats` - Real-time resource usage

#### Added - Health Checks & Restart Policies
- **Health Checks** (`pkg/health`)
  - Command-based health verification
  - Configurable intervals and timeouts
  - Health status tracking (healthy, unhealthy, starting)
  - Automatic health event emission
- **Restart Policies** (`pkg/restart`)
  - Multiple policies (no, always, on-failure, unless-stopped)
  - Exponential backoff
  - Maximum retry configuration
  - Smart restart delay management

#### Added - Build Foundation
- **Dockerfile Parser** (`pkg/build`)
  - Parse standard Dockerfile syntax
  - Multi-stage build support
  - Build argument (ARG) support
  - Build stage tracking

### Phase 2: Feature Completeness

#### Added - Enhanced CLI
- **Cobra Framework Integration**
  - Docker-like command structure
  - Subcommand organization
  - Flag parsing and validation
  - Help documentation
- **Container Lifecycle Commands**
  - `containr create` - Create container without starting
  - `containr start` - Start existing container
  - `containr stop` - Stop running container
  - `containr rm` - Remove container
  - `containr ps` - List containers
  - `containr logs` - View container logs
  - `containr exec` - Execute command in container
- **Image Management Commands**
  - `containr pull` - Pull images from registry
  - `containr images` - List images
  - `containr rmi` - Remove images
  - `containr import` - Import image from tarball
  - `containr export` - Export container to tarball
- **Inspection Commands**
  - `containr inspect` - Detailed container/image info
  - `containr stats` - Live resource usage
  - `containr top` - Show container processes

#### Added - Volume Management
- **Volume Support** (`pkg/volume`)
  - Named volumes
  - Bind mounts
  - Read-only mounts
  - tmpfs mounts
  - Volume lifecycle management
- **Volume Commands**
  - `containr volume create` - Create volume
  - `containr volume ls` - List volumes
  - `containr volume rm` - Remove volume
  - `containr volume inspect` - Volume details
  - `containr volume prune` - Remove unused volumes

#### Added - Registry Integration
- **Registry Client** (`pkg/registry`)
  - Docker Hub support
  - OCI registry compatibility
  - Authentication support
  - Layer downloading with progress
  - Parallel layer extraction
- **OCI Image Format**
  - Full OCI image specification support
  - Layer management
  - Config JSON handling
  - Manifest parsing
  - Content-addressable storage

#### Added - User Namespace Support
- **Rootless Containers** (`pkg/userns`)
  - UID/GID mapping
  - subuid/subgid parsing
  - User namespace creation
  - Root-in-container to unprivileged-on-host mapping
  - Enhanced security through isolation

### Phase 1: Foundation

#### Added - Core Container Runtime
- **Namespace Isolation** (`pkg/namespace`)
  - UTS namespace (hostname isolation)
  - PID namespace (process isolation)
  - Mount namespace (filesystem isolation)
  - IPC namespace (inter-process communication isolation)
  - Network namespace (network isolation)
  - User namespace (user isolation)
- **Resource Management** (`pkg/cgroup`)
  - Cgroup v1 and v2 support
  - CPU limits
  - Memory limits
  - PID limits
  - Resource statistics
- **Filesystem Isolation** (`pkg/rootfs`)
  - Chroot support
  - Pivot root implementation
  - Overlay filesystem support
  - Layer management
  - Mount propagation
- **Networking** (`pkg/network`)
  - Virtual ethernet pairs (veth)
  - Bridge networking
  - Network namespace setup
  - Basic IP configuration

#### Added - Security Features
- **Capabilities Management** (`pkg/capabilities`)
  - Linux capability dropping
  - Safe default capabilities
  - Configurable capability sets
  - Capability inheritance control
- **Seccomp Profiles** (`pkg/seccomp`)
  - Default restrictive profile
  - Custom profile support
  - Docker-compatible profiles
  - Syscall filtering
- **LSM Support** (`pkg/security`)
  - AppArmor integration
  - SELinux integration
  - Automatic LSM detection
  - Profile management

#### Added - Error Handling & Logging
- **Structured Logging** (`pkg/logger`)
  - Configurable log levels (debug, info, warn, error)
  - Context-rich logging
  - Structured fields
  - Lifecycle event logging
- **Error Management** (`pkg/errors`)
  - Unique error codes
  - Context-wrapped errors
  - User-friendly error messages
  - Actionable hints
  - Programmatic error handling

#### Added - Testing Infrastructure
- **Unit Tests**
  - Comprehensive package coverage
  - >70% code coverage
  - Mock interfaces
  - Table-driven tests
- **Integration Tests**
  - End-to-end scenarios
  - Privileged test support
  - Cleanup verification
  - Multi-container tests
- **CI/CD Pipeline**
  - GitHub Actions workflows
  - Automated testing
  - Code coverage reporting
  - Static analysis
  - Security scanning

## Version History

### Version Scheme

Containr follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

### Release Schedule

- **Major Releases:** Annually (or when major milestones reached)
- **Minor Releases:** Quarterly (every 3-4 months)
- **Patch Releases:** As needed (security: immediately)

### Support Policy

- **Current Major Version:** Full support
- **Previous Major Version:** Security fixes only (6 months)
- **Older Versions:** Unsupported

## Upgrade Guide

### Upgrading to 1.0.0

**From 0.x versions:**

1. **New Configuration Format:**
   - OCI spec is now the standard format
   - Legacy config format deprecated

2. **API Changes:**
   - Some package APIs have changed for OCI compliance
   - See [MIGRATION.md](docs/MIGRATION.md) for details

3. **New Features:**
   - Take advantage of performance tools
   - Use OCI bundles for portability
   - Explore new networking modes

### Breaking Changes

None in 1.0.0 (initial stable release)

## Migration Guides

### From Docker

See [docs/MIGRATION_FROM_DOCKER.md](docs/MIGRATION_FROM_DOCKER.md) for migrating from Docker to Containr for educational purposes.

### From Previous Versions

- **0.x to 1.0:** See upgrade guide above

## Contributors

Special thanks to all contributors who helped make Containr possible!

See [CONTRIBUTORS.md](CONTRIBUTORS.md) for a full list of contributors.

## Links

- **Repository:** https://github.com/therealutkarshpriyadarshi/containr
- **Documentation:** [docs/](docs/)
- **Issue Tracker:** https://github.com/therealutkarshpriyadarshi/containr/issues
- **Discussions:** https://github.com/therealutkarshpriyadarshi/containr/discussions
- **Releases:** https://github.com/therealutkarshpriyadarshi/containr/releases

---

**Note:** This changelog is maintained manually. For detailed commit history, see the Git log.

[Unreleased]: https://github.com/therealutkarshpriyadarshi/containr/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/therealutkarshpriyadarshi/containr/releases/tag/v1.0.0
