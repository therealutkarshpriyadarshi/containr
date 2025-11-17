# Phase 6: Cloud-Native Integration & Advanced Runtime Features

**Status:** ✅ Complete
**Version:** 1.1.0
**Date:** November 17, 2025

## Overview

Phase 6 represents the evolution of containr from an educational container runtime into a cloud-native platform with advanced integration capabilities. This phase focuses on enterprise-grade features, extensibility, and integration with the broader cloud-native ecosystem.

## Goals

1. **Cloud-Native Integration**: Enable containr to work with Kubernetes and other orchestration platforms
2. **Extensibility**: Provide a plugin system for custom functionality
3. **Performance**: Implement snapshot support for fast container operations
4. **Complete Build System**: Full Dockerfile build implementation with advanced features
5. **Enterprise Features**: Add features needed for production-like educational environments

---

## 🎯 Features

### 6.1 CRI (Container Runtime Interface) Support

Enable containr to work as a Kubernetes container runtime through the CRI specification.

#### What is CRI?

The Container Runtime Interface (CRI) is a plugin interface that enables Kubernetes to use different container runtimes without needing to recompile. It defines APIs for container and image management.

#### Implementation

**Package**: `pkg/cri`

**Features**:
- gRPC server implementing CRI APIs
- RuntimeService (container lifecycle, exec, port-forward)
- ImageService (image pull, list, remove)
- Pod sandbox support
- Container metrics via CRI
- Streaming server for exec and logs

**API Endpoints**:
```go
// RuntimeService
RunPodSandbox(config *PodSandboxConfig) (string, error)
StopPodSandbox(podSandboxId string) error
RemovePodSandbox(podSandboxId string) error
CreateContainer(podSandboxId string, config *ContainerConfig) (string, error)
StartContainer(containerId string) error
StopContainer(containerId string, timeout int64) error
RemoveContainer(containerId string) error

// ImageService
ListImages(filter *ImageFilter) ([]*Image, error)
PullImage(image *ImageSpec, auth *AuthConfig) (string, error)
RemoveImage(image *ImageSpec) error
ImageStatus(image *ImageSpec) (*Image, error)
```

**Usage**:
```bash
# Start CRI server
sudo containr cri start --listen /var/run/containr.sock

# Configure kubelet to use containr
kubelet --container-runtime=remote \
        --container-runtime-endpoint=unix:///var/run/containr.sock
```

**Configuration**:
```yaml
# /etc/containr/cri-config.yaml
version: 1
runtime:
  endpoint: /var/run/containr.sock
  network_plugin: cni
  cni_conf_dir: /etc/cni/net.d
  cni_bin_dir: /opt/cni/bin
image_service:
  registry_mirrors:
    - https://mirror.gcr.io
  insecure_registries:
    - localhost:5000
```

#### Educational Value

- **Learn Kubernetes Internals**: Understand how Kubernetes communicates with container runtimes
- **gRPC & Protobuf**: Work with modern RPC frameworks
- **Pod Concepts**: Understand Kubernetes pod abstractions
- **Streaming Protocols**: Learn about container exec and logging mechanisms

---

### 6.2 Plugin System

Extensible architecture allowing custom plugins for various runtime components.

#### Architecture

**Package**: `pkg/plugin`

**Plugin Types**:
1. **Runtime Plugins** - Custom container lifecycle hooks
2. **Network Plugins** - Custom networking implementations (CNI compatible)
3. **Storage Plugins** - Custom volume drivers
4. **Logging Plugins** - Custom log collectors
5. **Metrics Plugins** - Custom metrics exporters

**Plugin Interface**:
```go
// Plugin represents a containr plugin
type Plugin interface {
    // Name returns the plugin name
    Name() string

    // Type returns the plugin type
    Type() PluginType

    // Init initializes the plugin
    Init(config map[string]interface{}) error

    // Start starts the plugin
    Start() error

    // Stop stops the plugin
    Stop() error
}

// PluginType defines plugin categories
type PluginType string

const (
    RuntimePlugin PluginType = "runtime"
    NetworkPlugin PluginType = "network"
    StoragePlugin PluginType = "storage"
    LoggingPlugin PluginType = "logging"
    MetricsPlugin PluginType = "metrics"
)
```

**Plugin Discovery**:
```go
// Plugins are discovered from:
// 1. Built-in plugins (compiled)
// 2. /etc/containr/plugins/ (socket-based)
// 3. /opt/containr/plugins/ (binary plugins)

type PluginManager struct {
    plugins map[string]Plugin
}

func (pm *PluginManager) Load(path string) error
func (pm *PluginManager) Register(plugin Plugin) error
func (pm *PluginManager) Get(name string) (Plugin, error)
func (pm *PluginManager) List() []Plugin
```

**Example Plugin**:
```go
// Custom metrics exporter plugin
type PrometheusPlugin struct {
    config map[string]interface{}
    server *http.Server
}

func (p *PrometheusPlugin) Name() string {
    return "prometheus-exporter"
}

func (p *PrometheusPlugin) Type() plugin.PluginType {
    return plugin.MetricsPlugin
}

func (p *PrometheusPlugin) Init(config map[string]interface{}) error {
    p.config = config
    return nil
}

func (p *PrometheusPlugin) Start() error {
    // Start Prometheus metrics server
    http.Handle("/metrics", promhttp.Handler())
    p.server = &http.Server{Addr: ":9090"}
    go p.server.ListenAndServe()
    return nil
}
```

**Usage**:
```bash
# List available plugins
containr plugin ls

# Install a plugin
containr plugin install ./prometheus-exporter.so

# Enable a plugin
containr plugin enable prometheus-exporter

# Configure a plugin
containr plugin configure prometheus-exporter --port 9090
```

**Plugin Configuration**:
```yaml
# /etc/containr/plugins.yaml
plugins:
  - name: prometheus-exporter
    type: metrics
    enabled: true
    config:
      port: 9090
      path: /metrics

  - name: custom-logger
    type: logging
    enabled: true
    config:
      output: /var/log/containr/custom.log
```

#### Educational Value

- **Plugin Architecture**: Learn how to design extensible systems
- **Dynamic Loading**: Understand Go plugins and dynamic libraries
- **Interface Design**: Practice clean API design
- **Separation of Concerns**: Modular architecture patterns

---

### 6.3 Snapshot Support

Fast container creation and migration using filesystem snapshots.

#### What are Snapshots?

Snapshots capture the state of a container's filesystem at a point in time, enabling:
- **Fast Container Creation**: Start new containers from snapshots instantly
- **Container Migration**: Move containers between hosts
- **Backup & Restore**: Save and restore container state
- **Deduplication**: Share common layers between snapshots

#### Implementation

**Package**: `pkg/snapshot`

**Features**:
- Multiple snapshot drivers (overlay2, btrfs, zfs)
- Copy-on-write (COW) optimization
- Snapshot chaining (parent-child relationships)
- Metadata management
- Snapshot diff computation

**Snapshot Interface**:
```go
type Snapshotter interface {
    // Prepare creates a snapshot for a container
    Prepare(ctx context.Context, key, parent string) ([]Mount, error)

    // Commit creates an immutable snapshot
    Commit(ctx context.Context, name, key string) error

    // Remove removes a snapshot
    Remove(ctx context.Context, key string) error

    // View creates a read-only view of a snapshot
    View(ctx context.Context, key, parent string) ([]Mount, error)

    // Walk iterates over all snapshots
    Walk(ctx context.Context, fn WalkFunc) error

    // Stat returns info about a snapshot
    Stat(ctx context.Context, key string) (Info, error)
}

type Info struct {
    Name      string
    Parent    string
    Kind      Kind
    CreatedAt time.Time
    UpdatedAt time.Time
    Labels    map[string]string
}
```

**Snapshot Drivers**:

1. **Overlay2 Driver** (default):
```go
type Overlay2Snapshotter struct {
    root string
}

// Uses Linux overlay filesystem
// Fast, efficient, widely supported
```

2. **Btrfs Driver**:
```go
type BtrfsSnapshotter struct {
    device string
}

// Native CoW filesystem
// Excellent snapshot performance
// Requires Btrfs filesystem
```

3. **ZFS Driver**:
```go
type ZFSSnapshotter struct {
    dataset string
}

// Enterprise-grade CoW
// Advanced features (compression, dedup)
// Requires ZFS
```

**Usage**:
```bash
# Create a snapshot of running container
containr snapshot create myapp snapshot1

# List snapshots
containr snapshot ls

# Create container from snapshot
containr run --snapshot snapshot1 alpine /bin/sh

# Export snapshot
containr snapshot export snapshot1 -o snapshot.tar.gz

# Import snapshot
containr snapshot import snapshot.tar.gz

# Remove snapshot
containr snapshot rm snapshot1

# Show snapshot diff
containr snapshot diff snapshot1 snapshot2
```

**Snapshot Metadata**:
```json
{
  "name": "snapshot1",
  "parent": "alpine:latest",
  "created_at": "2025-11-17T10:00:00Z",
  "size": 5242880,
  "labels": {
    "app": "myapp",
    "version": "1.0"
  },
  "driver": "overlay2"
}
```

**API Example**:
```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/snapshot"

// Initialize snapshotter
snap := snapshot.NewOverlay2("/var/lib/containr/snapshots")

// Create snapshot
mounts, err := snap.Prepare(ctx, "container1", "alpine")
if err != nil {
    log.Fatal(err)
}

// Use the mounts for container
// ...

// Commit snapshot
err = snap.Commit(ctx, "snapshot1", "container1")
```

#### Educational Value

- **Filesystem Concepts**: Learn about Copy-on-Write filesystems
- **Storage Drivers**: Understand different storage technologies
- **Data Structures**: Work with directed acyclic graphs (DAGs)
- **Performance Optimization**: Learn about caching and deduplication

---

### 6.4 Complete Image Build Engine

Full implementation of `containr build` command with Dockerfile support.

#### Build Features

**Package**: `pkg/builder`

**Capabilities**:
- Complete Dockerfile instruction support
- Multi-stage builds
- Build cache with layer caching
- BuildKit-compatible features
- Parallel build execution
- Build secrets and SSH forwarding
- Custom build platforms

**Supported Instructions**:
```dockerfile
# All standard Dockerfile instructions
FROM alpine:latest AS builder
LABEL maintainer="you@example.com"
ARG BUILD_VERSION=1.0
ENV APP_HOME=/app
WORKDIR /app
COPY . .
ADD archive.tar.gz /data
RUN apk add --no-cache gcc make
EXPOSE 8080
VOLUME /data
USER nobody
HEALTHCHECK --interval=30s CMD curl -f http://localhost:8080/health
ENTRYPOINT ["/app/start.sh"]
CMD ["--help"]
ONBUILD RUN echo "Triggered on child image build"
STOPSIGNAL SIGTERM
SHELL ["/bin/bash", "-c"]
```

**Build Engine Architecture**:
```go
type Builder struct {
    dockerfile *build.Dockerfile
    context    BuildContext
    cache      *BuildCache
    executor   Executor
}

// BuildContext represents build files
type BuildContext interface {
    ReadFile(path string) ([]byte, error)
    Walk(fn filepath.WalkFunc) error
}

// BuildCache manages layer caching
type BuildCache struct {
    backend CacheBackend
}

// Executor runs build steps
type Executor interface {
    Execute(ctx context.Context, step BuildStep) (Layer, error)
}
```

**Build Process**:
```go
// 1. Parse Dockerfile
dockerfile, err := build.ParseDockerfile("Dockerfile")

// 2. Create build context
ctx, err := build.NewContext(".")

// 3. Initialize builder
builder := build.NewBuilder(dockerfile, ctx)

// 4. Configure build
builder.SetTag("myapp:latest")
builder.SetBuildArgs(map[string]string{"VERSION": "1.0"})
builder.SetTarget("production")  // Multi-stage target

// 5. Execute build
image, err := builder.Build()
```

**Usage**:
```bash
# Build from Dockerfile
containr build -t myapp:latest .

# Build with arguments
containr build --build-arg VERSION=1.0 -t myapp:v1 .

# Multi-stage build (target specific stage)
containr build --target production -t myapp:prod .

# Build with no cache
containr build --no-cache -t myapp:latest .

# Build with cache from image
containr build --cache-from myapp:base -t myapp:latest .

# Build with secrets
containr build --secret id=mysecret,src=/path/to/secret -t myapp .

# Build for different platform
containr build --platform linux/arm64 -t myapp:arm .

# Show build progress
containr build --progress plain -t myapp .
```

**Build Cache**:
```go
// Cache key is computed from:
// 1. Parent layer hash
// 2. Build instruction
// 3. Build context (for COPY/ADD)
// 4. Build arguments

type CacheKey struct {
    ParentHash string
    Instruction string
    ContextHash string
    BuildArgs  map[string]string
}

// Cache lookup
func (c *BuildCache) Lookup(key CacheKey) (*Layer, bool)

// Cache store
func (c *BuildCache) Store(key CacheKey, layer *Layer) error
```

**Multi-Stage Build Example**:
```dockerfile
# Stage 1: Build
FROM golang:1.21 AS builder
WORKDIR /src
COPY . .
RUN go build -o app

# Stage 2: Production
FROM alpine:latest AS production
COPY --from=builder /src/app /app
ENTRYPOINT ["/app"]
```

**Build Configuration**:
```yaml
# .containr/build.yaml
build:
  dockerfile: Dockerfile
  context: .
  args:
    VERSION: 1.0
    BUILD_DATE: ${BUILD_DATE}
  cache:
    enabled: true
    ttl: 168h  # 7 days
  platforms:
    - linux/amd64
    - linux/arm64
  secrets:
    - id: github_token
      src: ~/.github/token
```

#### Educational Value

- **Build Systems**: Learn how container images are built
- **Parsing**: Understand lexing and parsing of DSLs
- **Caching Strategies**: Learn about build cache optimization
- **Multi-Stage Builds**: Understand image size optimization
- **Layer Management**: Work with image layers and hashing

---

## 📦 Package Structure

```
pkg/
├── cri/                    # Container Runtime Interface
│   ├── server.go          # CRI gRPC server
│   ├── runtime.go         # RuntimeService implementation
│   ├── image.go           # ImageService implementation
│   ├── streaming.go       # Exec/logs streaming
│   └── cri_test.go        # CRI tests
├── plugin/                # Plugin system
│   ├── plugin.go          # Plugin interface and manager
│   ├── loader.go          # Dynamic plugin loading
│   ├── registry.go        # Plugin registry
│   └── plugin_test.go     # Plugin tests
├── snapshot/              # Snapshot support
│   ├── snapshot.go        # Snapshot interface
│   ├── overlay2.go        # Overlay2 driver
│   ├── btrfs.go           # Btrfs driver (optional)
│   ├── metadata.go        # Snapshot metadata
│   └── snapshot_test.go   # Snapshot tests
└── builder/               # Image build engine
    ├── builder.go         # Main builder
    ├── parser.go          # Dockerfile parser (enhanced)
    ├── executor.go        # Build step executor
    ├── cache.go           # Build cache
    ├── context.go         # Build context
    ├── instructions.go    # Instruction handlers
    └── builder_test.go    # Builder tests
```

## 🎨 CLI Commands

### CRI Commands

```bash
containr cri start              # Start CRI server
containr cri stop               # Stop CRI server
containr cri status             # Check CRI server status
containr cri version            # Show CRI API version
```

### Plugin Commands

```bash
containr plugin ls              # List plugins
containr plugin install <path>  # Install plugin
containr plugin enable <name>   # Enable plugin
containr plugin disable <name>  # Disable plugin
containr plugin remove <name>   # Remove plugin
containr plugin info <name>     # Show plugin info
```

### Snapshot Commands

```bash
containr snapshot create <container> <name>  # Create snapshot
containr snapshot ls                         # List snapshots
containr snapshot rm <name>                  # Remove snapshot
containr snapshot inspect <name>             # Inspect snapshot
containr snapshot export <name> -o <file>    # Export snapshot
containr snapshot import <file>              # Import snapshot
containr snapshot diff <snap1> <snap2>       # Show differences
```

### Build Commands

```bash
containr build [OPTIONS] PATH                # Build image
containr build -t <tag> .                    # Build with tag
containr build --target <stage> .            # Multi-stage target
containr build --platform <platform> .       # Build for platform
containr buildx create                       # Create builder instance
containr buildx use <builder>                # Set active builder
```

## 🔧 Configuration

### CRI Configuration

```yaml
# /etc/containr/cri.yaml
version: 1
server:
  address: unix:///var/run/containr.sock
  stream_address: 0.0.0.0
  stream_port: 10010
runtime:
  pod_cidr: 10.244.0.0/16
  service_cidr: 10.96.0.0/12
network:
  plugin: cni
  cni_conf_dir: /etc/cni/net.d
  cni_bin_dir: /opt/cni/bin
```

### Plugin Configuration

```yaml
# /etc/containr/plugins.yaml
plugins:
  enabled:
    - prometheus-exporter
    - fluentd-logger
  config:
    prometheus-exporter:
      port: 9090
    fluentd-logger:
      host: localhost
      port: 24224
```

### Build Configuration

```yaml
# /etc/containr/builder.yaml
builder:
  workers: 4
  cache:
    enabled: true
    max_size: 10GB
    ttl: 168h
  registry:
    mirrors:
      - https://mirror.gcr.io
```

## 🧪 Testing

### Unit Tests

```bash
# Test CRI implementation
go test ./pkg/cri/...

# Test plugin system
go test ./pkg/plugin/...

# Test snapshot support
go test ./pkg/snapshot/...

# Test build engine
go test ./pkg/builder/...
```

### Integration Tests

```bash
# Test CRI with Kubernetes
make test-cri-integration

# Test plugins
make test-plugins

# Test snapshots
make test-snapshots

# Test builds
make test-builder
```

### E2E Tests

```bash
# Complete Phase 6 E2E tests
make test-phase6-e2e
```

## 📊 Performance Benchmarks

### Snapshot Performance

```bash
# Benchmark snapshot creation
go test -bench=BenchmarkSnapshotCreate ./pkg/snapshot/

# Benchmark snapshot commit
go test -bench=BenchmarkSnapshotCommit ./pkg/snapshot/
```

### Build Performance

```bash
# Benchmark build with cache
go test -bench=BenchmarkBuildCached ./pkg/builder/

# Benchmark build without cache
go test -bench=BenchmarkBuildNocache ./pkg/builder/
```

## 🎓 Educational Resources

### Tutorials

1. **Using containr with Kubernetes**
   - Setting up CRI
   - Deploying pods
   - Debugging CRI issues

2. **Writing containr Plugins**
   - Plugin architecture
   - Custom metrics exporter
   - Network plugin example

3. **Working with Snapshots**
   - Fast container creation
   - Container migration
   - Backup strategies

4. **Building Images**
   - Dockerfile best practices
   - Multi-stage builds
   - Build optimization

### Example Projects

- **Kubernetes Integration**: Deploy containr as a Kubernetes runtime
- **Custom Plugin**: Metrics exporter for Datadog
- **Snapshot Migration**: Live container migration demo
- **CI/CD Pipeline**: Build pipeline using containr build

## 🚀 Use Cases

### 1. Kubernetes Learning Environment

```bash
# Use containr as Kubernetes runtime
sudo containr cri start
kubelet --container-runtime=remote \
        --container-runtime-endpoint=unix:///var/run/containr.sock
```

### 2. Custom Metrics Collection

```go
// Custom plugin for metrics
plugin := &CustomMetricsPlugin{}
pm.Register(plugin)
```

### 3. Fast Development Workflow

```bash
# Use snapshots for quick iterations
containr snapshot create devenv base-snapshot
containr run --snapshot base-snapshot myapp
```

### 4. Automated Image Builds

```bash
# CI/CD pipeline
containr build -t myapp:${GIT_SHA} .
containr push myapp:${GIT_SHA}
```

## 🔒 Security Considerations

### CRI Security

- **Authentication**: mTLS for CRI connections
- **Authorization**: RBAC integration
- **Pod Security**: Pod Security Standards enforcement
- **Network Policies**: Support for NetworkPolicy

### Plugin Security

- **Signing**: Verify plugin signatures
- **Sandboxing**: Isolate plugin execution
- **Permissions**: Fine-grained capability control
- **Auditing**: Log all plugin operations

### Snapshot Security

- **Encryption**: Encrypt snapshots at rest
- **Integrity**: Verify snapshot checksums
- **Access Control**: Restrict snapshot operations
- **Scanning**: Security scan snapshots

### Build Security

- **Secrets Management**: Secure build secrets
- **Registry Auth**: Secure registry credentials
- **Supply Chain**: SBOM generation
- **Scanning**: Vulnerability scanning in builds

## 📈 Metrics & Monitoring

### CRI Metrics

```
cri_operations_total{operation="run_pod_sandbox"}
cri_operation_duration_seconds{operation="create_container"}
cri_errors_total{operation="start_container"}
```

### Plugin Metrics

```
plugin_loaded_total{type="metrics"}
plugin_errors_total{plugin="prometheus-exporter"}
plugin_calls_total{plugin="custom-logger"}
```

### Snapshot Metrics

```
snapshot_create_duration_seconds
snapshot_size_bytes{name="snapshot1"}
snapshot_operations_total{operation="commit"}
```

### Build Metrics

```
build_duration_seconds{stage="production"}
build_cache_hits_total
build_layers_created_total
```

## 🐛 Troubleshooting

### CRI Issues

```bash
# Check CRI server status
containr cri status

# View CRI logs
journalctl -u containr-cri

# Test CRI connectivity
crictl --runtime-endpoint unix:///var/run/containr.sock version
```

### Plugin Issues

```bash
# List plugin status
containr plugin ls

# View plugin logs
containr plugin logs prometheus-exporter

# Restart plugin
containr plugin restart prometheus-exporter
```

### Snapshot Issues

```bash
# Verify snapshot integrity
containr snapshot verify snapshot1

# Repair snapshot metadata
containr snapshot repair

# Clear snapshot cache
containr snapshot prune
```

### Build Issues

```bash
# Build with debug output
containr build --debug -t myapp .

# Clear build cache
containr builder prune

# View build history
containr build history myapp:latest
```

## 🎯 Future Enhancements

While Phase 6 is complete, potential future improvements include:

- **Advanced CRI Features**: GPU support, device plugins
- **Plugin Marketplace**: Central repository for plugins
- **Advanced Snapshots**: Incremental snapshots, deduplication
- **BuildKit Integration**: Full BuildKit compatibility
- **WebAssembly Support**: Run WASM containers
- **Edge Computing**: Optimizations for edge deployments

## 📚 References

- [CRI Specification](https://github.com/kubernetes/cri-api)
- [Kubernetes Runtime Class](https://kubernetes.io/docs/concepts/containers/runtime-class/)
- [containerd Snapshotter](https://github.com/containerd/containerd/tree/main/docs/snapshotters)
- [BuildKit](https://github.com/moby/buildkit)
- [CNI Specification](https://github.com/containernetworking/cni)

---

**Phase 6 Status**: ✅ Complete
**Next Steps**: Production deployment testing and community feedback

**Educational Impact**: Phase 6 teaches advanced cloud-native concepts, Kubernetes integration, plugin architecture, and enterprise-grade container features.
