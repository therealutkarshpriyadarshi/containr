# Phase 7: Advanced Production Features & Enterprise Integration

**Status:** ✅ Complete
**Version:** 1.2.0
**Date:** November 17, 2025

## Overview

Phase 7 represents the evolution of containr into an enterprise-ready container runtime with advanced production features, comprehensive security, and deep integration with cloud-native ecosystems. This phase focuses on multi-tenancy, advanced observability, container migration, service mesh integration, and enterprise-grade security capabilities.

## Goals

1. **Multi-Tenancy & RBAC**: Enable secure multi-user environments with role-based access control
2. **Advanced Observability**: Comprehensive distributed tracing and metrics with OpenTelemetry
3. **Container Migration**: Live container migration and checkpointing using CRIU
4. **Service Mesh Integration**: Native Envoy sidecar support and traffic management
5. **Enterprise Security**: Policy enforcement, image signing, runtime security monitoring
6. **Storage Innovation**: CSI driver support and volume encryption

---

## 🎯 Features

### 7.1 RBAC & Multi-Tenancy

Enterprise-grade access control and resource isolation for multi-user environments.

#### What is RBAC?

Role-Based Access Control (RBAC) is an authorization model that restricts system access based on user roles and permissions. In containr, RBAC enables:
- **User Isolation**: Each user has their own namespace
- **Resource Quotas**: Limit resources per user/team
- **Fine-grained Permissions**: Control who can create, start, stop containers
- **Audit Logging**: Track all user operations

#### Implementation

**Package**: `pkg/rbac`

**Features**:
- Role and permission management
- User authentication and authorization
- Resource quotas per user/namespace
- Audit logging for all operations
- Integration with external identity providers (LDAP, OAuth)

**RBAC Model**:
```go
// Role defines a set of permissions
type Role struct {
    Name        string
    Permissions []Permission
}

// Permission defines an allowed operation
type Permission struct {
    Resource string   // container, image, volume, etc.
    Actions  []Action // create, read, update, delete
}

// User represents an authenticated user
type User struct {
    ID       string
    Username string
    Roles    []Role
    Quota    *ResourceQuota
}

// ResourceQuota limits user resources
type ResourceQuota struct {
    MaxContainers int
    MaxCPU        float64
    MaxMemory     int64
    MaxStorage    int64
}
```

**Built-in Roles**:
```go
const (
    RoleAdmin      = "admin"       // Full access
    RoleDeveloper  = "developer"   // Create/manage own containers
    RoleOperator   = "operator"    // Start/stop containers
    RoleViewer     = "viewer"      // Read-only access
)
```

**Usage**:
```bash
# Create user
containr user create alice --role developer --quota cpu=2,memory=4Gi

# List users
containr user ls

# Grant role
containr user grant alice --role operator

# Revoke role
containr user revoke alice --role developer

# Set quota
containr quota set alice --max-containers 10 --max-cpu 4 --max-memory 8Gi

# Run container as user
containr run --user alice alpine /bin/sh

# View user's containers
containr ps --user alice

# Audit log
containr audit --user alice --action create
```

**Configuration**:
```yaml
# /etc/containr/rbac.yaml
rbac:
  enabled: true
  default_role: viewer

  roles:
    admin:
      permissions:
        - resource: "*"
          actions: ["*"]

    developer:
      permissions:
        - resource: "container"
          actions: ["create", "read", "update", "delete"]
        - resource: "image"
          actions: ["read", "pull"]
        - resource: "volume"
          actions: ["create", "read", "delete"]

    operator:
      permissions:
        - resource: "container"
          actions: ["read", "start", "stop"]

    viewer:
      permissions:
        - resource: "*"
          actions: ["read"]

  quotas:
    default:
      max_containers: 5
      max_cpu: 2.0
      max_memory: 4294967296  # 4Gi
      max_storage: 10737418240  # 10Gi

    developer:
      max_containers: 20
      max_cpu: 8.0
      max_memory: 17179869184  # 16Gi
      max_storage: 53687091200  # 50Gi

  authentication:
    providers:
      - type: local
        enabled: true
      - type: ldap
        enabled: false
        server: ldap://ldap.example.com
        base_dn: dc=example,dc=com
      - type: oauth
        enabled: false
        provider: github
        client_id: ${OAUTH_CLIENT_ID}
```

**API Example**:
```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/rbac"

// Initialize RBAC manager
rbacMgr := rbac.NewManager("/etc/containr/rbac.yaml")

// Create user
user := &rbac.User{
    Username: "alice",
    Roles:    []rbac.Role{rbac.RoleDeveloper},
    Quota: &rbac.ResourceQuota{
        MaxContainers: 10,
        MaxCPU:        4.0,
        MaxMemory:     8 * 1024 * 1024 * 1024,
    },
}
err := rbacMgr.CreateUser(user)

// Check permission
allowed, err := rbacMgr.CheckPermission("alice", "container", "create")
if !allowed {
    return errors.New("permission denied")
}

// Enforce quota
err = rbacMgr.EnforceQuota("alice")
```

#### Educational Value

- **Authorization Models**: Learn about RBAC and policy-based access control
- **Multi-Tenancy**: Understand resource isolation and fair sharing
- **Security Design**: Practice least-privilege principles
- **Identity Management**: Work with authentication providers

---

### 7.2 Advanced Observability with OpenTelemetry

Comprehensive distributed tracing, metrics, and logging for production observability.

#### What is OpenTelemetry?

OpenTelemetry is a unified observability framework providing:
- **Traces**: Distributed request tracking across services
- **Metrics**: Performance and resource measurements
- **Logs**: Structured event logging
- **Context Propagation**: Correlation across distributed systems

#### Implementation

**Package**: `pkg/observability`

**Features**:
- OpenTelemetry integration (traces, metrics, logs)
- Prometheus metrics exporter
- Jaeger/Zipkin trace exporter
- Distributed context propagation
- Custom metrics and spans
- Performance profiling integration

**Observability Architecture**:
```go
// Tracer provides distributed tracing
type Tracer struct {
    provider trace.TracerProvider
    exporter trace.SpanExporter
}

// MetricsCollector collects and exports metrics
type MetricsCollector struct {
    provider metric.MeterProvider
    exporter metric.Exporter
}

// Logger provides structured logging
type Logger struct {
    logger *slog.Logger
    exporter slog.Handler
}
```

**Instrumentation**:
```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
    "go.opentelemetry.io/otel/metric"
)

// Trace container operations
func (c *Container) Start(ctx context.Context) error {
    ctx, span := otel.Tracer("containr").Start(ctx, "container.start")
    defer span.End()

    span.SetAttributes(
        attribute.String("container.id", c.ID),
        attribute.String("container.name", c.Name),
    )

    // Record start time metric
    startTime := time.Now()
    defer func() {
        duration := time.Since(startTime)
        containerStartDuration.Record(ctx, duration.Seconds())
    }()

    // ... start logic
}

// Metrics
var (
    containerStartDuration = metric.Float64Histogram(
        "container.start.duration",
        metric.WithDescription("Container start duration in seconds"),
        metric.WithUnit("s"),
    )

    containerCount = metric.Int64Counter(
        "container.count",
        metric.WithDescription("Total number of containers"),
    )

    containerCPUUsage = metric.Float64ObservableGauge(
        "container.cpu.usage",
        metric.WithDescription("Container CPU usage"),
    )
)
```

**Usage**:
```bash
# Start with OpenTelemetry
containr run --trace-endpoint localhost:4317 alpine /bin/sh

# Export metrics to Prometheus
containr metrics export --format prometheus --port 9090

# View traces
containr trace ls
containr trace inspect <trace-id>

# Export to Jaeger
containr config set observability.trace.exporter jaeger
containr config set observability.trace.endpoint http://jaeger:14268/api/traces
```

**Configuration**:
```yaml
# /etc/containr/observability.yaml
observability:
  tracing:
    enabled: true
    provider: otlp
    endpoint: localhost:4317
    sampling_rate: 1.0
    exporters:
      - type: jaeger
        endpoint: http://jaeger:14268/api/traces
      - type: zipkin
        endpoint: http://zipkin:9411/api/v2/spans

  metrics:
    enabled: true
    provider: prometheus
    port: 9090
    path: /metrics
    interval: 15s
    exporters:
      - type: prometheus
        port: 9090
      - type: otlp
        endpoint: localhost:4317

  logging:
    enabled: true
    level: info
    format: json
    output: stdout
    exporters:
      - type: loki
        endpoint: http://loki:3100/loki/api/v1/push
```

**Prometheus Metrics**:
```
# Container metrics
container_start_duration_seconds{status="success"}
container_count_total{state="running"}
container_cpu_usage_percent{id="abc123"}
container_memory_usage_bytes{id="abc123"}
container_network_rx_bytes{id="abc123"}
container_network_tx_bytes{id="abc123"}

# Build metrics
build_duration_seconds{stage="production"}
build_cache_hit_ratio
build_layer_size_bytes{layer="sha256:..."}

# CRI metrics
cri_operation_duration_seconds{operation="run_pod_sandbox"}
cri_error_total{operation="create_container"}

# System metrics
containr_cpu_usage_percent
containr_memory_usage_bytes
containr_goroutines_count
```

**Grafana Dashboard**:
```json
{
  "dashboard": {
    "title": "Containr Metrics",
    "panels": [
      {
        "title": "Container Count",
        "targets": [{"expr": "container_count_total"}]
      },
      {
        "title": "CPU Usage",
        "targets": [{"expr": "container_cpu_usage_percent"}]
      },
      {
        "title": "Memory Usage",
        "targets": [{"expr": "container_memory_usage_bytes"}]
      }
    ]
  }
}
```

#### Educational Value

- **Observability Patterns**: Learn the three pillars of observability
- **Distributed Tracing**: Understand request flow across systems
- **Metrics Design**: Practice SLI/SLO-based monitoring
- **Performance Analysis**: Learn profiling and optimization

---

### 7.3 Container Checkpointing & Live Migration

Save and restore container state using CRIU (Checkpoint/Restore In Userspace).

#### What is Container Checkpointing?

Checkpointing allows you to:
- **Save State**: Freeze a running container and save its state to disk
- **Restore**: Resume a container from a checkpoint
- **Migrate**: Move containers between hosts without downtime
- **Fast Restart**: Instant container startup from checkpoint

#### Implementation

**Package**: `pkg/checkpoint`

**Features**:
- CRIU integration for process checkpointing
- Container state serialization
- Live migration support
- Pre-copy and post-copy migration
- Network state preservation
- File descriptor migration

**Checkpoint Interface**:
```go
// Checkpointer handles container checkpointing
type Checkpointer interface {
    // Checkpoint saves container state
    Checkpoint(ctx context.Context, container string, opts CheckpointOptions) error

    // Restore restores container from checkpoint
    Restore(ctx context.Context, checkpointPath string, opts RestoreOptions) error

    // Migrate migrates container to another host
    Migrate(ctx context.Context, container string, destination string) error
}

// CheckpointOptions configures checkpoint behavior
type CheckpointOptions struct {
    Name            string
    ImagePath       string
    LeaveRunning    bool
    PreDump         bool
    TrackMemory     bool
    AutoDedup       bool
    LazyPages       bool
}

// RestoreOptions configures restore behavior
type RestoreOptions struct {
    CheckpointPath  string
    Name            string
    DetachOnRestore bool
    RestoreNetwork  bool
}
```

**CRIU Integration**:
```go
import "github.com/checkpoint-restore/go-criu/v6"

// CRIUCheckpointer implements checkpointing with CRIU
type CRIUCheckpointer struct {
    criuPath string
}

func (c *CRIUCheckpointer) Checkpoint(ctx context.Context, container string, opts CheckpointOptions) error {
    // Get container process
    proc, err := getContainerProcess(container)
    if err != nil {
        return err
    }

    // Setup CRIU options
    criuOpts := &criu.CriuOpts{
        ImagesDirFd:     proto.Int32(int32(imageDirFd)),
        LogLevel:        proto.Int32(4),
        LogFile:         proto.String("checkpoint.log"),
        LeaveRunning:    proto.Bool(opts.LeaveRunning),
        TcpEstablished:  proto.Bool(true),
        ExtUnixSk:       proto.Bool(true),
        ShellJob:        proto.Bool(true),
        FileLocks:       proto.Bool(true),
    }

    // Perform checkpoint
    err = criu.Notify.Dump(criuOpts, proc.Pid)
    return err
}
```

**Usage**:
```bash
# Checkpoint a running container
containr checkpoint create myapp checkpoint1

# List checkpoints
containr checkpoint ls

# Restore from checkpoint
containr checkpoint restore checkpoint1

# Live migration
containr migrate myapp --to host2.example.com

# Pre-dump for faster migration
containr checkpoint create myapp pre1 --pre-dump
containr checkpoint create myapp pre2 --pre-dump --prev pre1
containr checkpoint create myapp final --prev pre2
containr checkpoint restore final

# Export checkpoint
containr checkpoint export checkpoint1 -o checkpoint.tar

# Import checkpoint
containr checkpoint import checkpoint.tar

# Remove checkpoint
containr checkpoint rm checkpoint1
```

**Migration Example**:
```bash
# Source host
containr migrate myapp \
    --to host2.example.com \
    --method pre-copy \
    --iterations 3

# Destination host automatically receives and restores
```

**Configuration**:
```yaml
# /etc/containr/checkpoint.yaml
checkpoint:
  enabled: true
  storage_dir: /var/lib/containr/checkpoints

  criu:
    path: /usr/sbin/criu
    log_level: 4
    tcp_established: true
    external_unix_sockets: true
    file_locks: true

  migration:
    enabled: true
    method: pre-copy
    max_iterations: 5
    bandwidth_limit: 100  # MB/s
    compression: true
```

**API Example**:
```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/checkpoint"

// Create checkpointer
cp := checkpoint.NewCRIUCheckpointer()

// Checkpoint container
opts := checkpoint.CheckpointOptions{
    Name:         "checkpoint1",
    ImagePath:    "/var/lib/containr/checkpoints/checkpoint1",
    LeaveRunning: false,
}
err := cp.Checkpoint(ctx, "myapp", opts)

// Restore container
restoreOpts := checkpoint.RestoreOptions{
    CheckpointPath: "/var/lib/containr/checkpoints/checkpoint1",
    Name:           "myapp-restored",
}
err = cp.Restore(ctx, "/var/lib/containr/checkpoints/checkpoint1", restoreOpts)
```

#### Educational Value

- **Process State**: Learn how Linux processes work
- **System Calls**: Understand process state capture
- **Migration Strategies**: Learn pre-copy vs post-copy
- **State Serialization**: Practice data structure serialization

---

### 7.4 Service Mesh Integration

Native Envoy sidecar support for traffic management and observability.

#### What is a Service Mesh?

A service mesh provides:
- **Traffic Management**: Load balancing, routing, retries
- **Security**: mTLS, authentication, authorization
- **Observability**: Distributed tracing, metrics
- **Resilience**: Circuit breaking, fault injection

#### Implementation

**Package**: `pkg/servicemesh`

**Features**:
- Automatic Envoy sidecar injection
- Traffic routing and load balancing
- mTLS for service-to-service communication
- Circuit breaking and retries
- Distributed tracing integration
- Traffic mirroring and splitting

**Service Mesh Architecture**:
```go
// SidecarConfig defines Envoy sidecar configuration
type SidecarConfig struct {
    Image          string
    AdminPort      int
    ProxyPort      int
    StatsPort      int
    TracingEnabled bool
    MTLSEnabled    bool
}

// TrafficPolicy defines routing rules
type TrafficPolicy struct {
    LoadBalancer LoadBalancerPolicy
    Retries      *RetryPolicy
    Timeout      time.Duration
    CircuitBreaker *CircuitBreakerPolicy
}

// LoadBalancerPolicy defines load balancing strategy
type LoadBalancerPolicy struct {
    Type   string  // round_robin, least_request, random
    Hosts  []string
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
    Attempts      int
    PerTryTimeout time.Duration
    RetryOn       []string
}

// CircuitBreakerPolicy defines circuit breaking
type CircuitBreakerPolicy struct {
    MaxConnections     int
    MaxPendingRequests int
    MaxRequests        int
    MaxRetries         int
}
```

**Envoy Configuration**:
```yaml
# Generated Envoy config
static_resources:
  listeners:
    - name: listener_0
      address:
        socket_address:
          address: 0.0.0.0
          port_value: 15001
      filter_chains:
        - filters:
            - name: envoy.filters.network.http_connection_manager
              typed_config:
                "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                stat_prefix: ingress_http
                route_config:
                  name: local_route
                  virtual_hosts:
                    - name: backend
                      domains: ["*"]
                      routes:
                        - match:
                            prefix: "/"
                          route:
                            cluster: backend_cluster
                http_filters:
                  - name: envoy.filters.http.router

  clusters:
    - name: backend_cluster
      connect_timeout: 0.25s
      type: STRICT_DNS
      lb_policy: ROUND_ROBIN
      load_assignment:
        cluster_name: backend_cluster
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address:
                      address: 127.0.0.1
                      port_value: 8080
```

**Usage**:
```bash
# Run container with Envoy sidecar
containr run --sidecar envoy --service-name myapp alpine /app/server

# Configure traffic policy
containr servicemesh policy set myapp \
    --load-balancer round-robin \
    --retries 3 \
    --timeout 5s \
    --circuit-breaker max-connections=100

# Enable mTLS
containr servicemesh mtls enable myapp

# Traffic mirroring
containr servicemesh mirror myapp --to myapp-v2 --percent 10

# Traffic splitting
containr servicemesh split myapp --v1 90 --v2 10

# Inject fault
containr servicemesh fault inject myapp --delay 100ms --percent 10

# View service mesh metrics
containr servicemesh stats myapp
```

**Configuration**:
```yaml
# /etc/containr/servicemesh.yaml
servicemesh:
  enabled: true
  sidecar:
    image: envoyproxy/envoy:v1.28.0
    admin_port: 15000
    proxy_port: 15001
    stats_port: 15002

  tracing:
    enabled: true
    provider: jaeger
    endpoint: jaeger:9411

  mtls:
    enabled: true
    ca_cert: /etc/containr/certs/ca.crt
    cert_dir: /etc/containr/certs

  policies:
    default:
      load_balancer: round_robin
      retries:
        attempts: 3
        per_try_timeout: 1s
      timeout: 10s
      circuit_breaker:
        max_connections: 1000
        max_pending_requests: 100
```

**API Example**:
```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/servicemesh"

// Create service mesh manager
sm := servicemesh.NewManager()

// Inject sidecar
sidecarConfig := &servicemesh.SidecarConfig{
    Image:          "envoyproxy/envoy:v1.28.0",
    AdminPort:      15000,
    ProxyPort:      15001,
    TracingEnabled: true,
    MTLSEnabled:    true,
}
err := sm.InjectSidecar(ctx, "myapp", sidecarConfig)

// Configure traffic policy
policy := &servicemesh.TrafficPolicy{
    LoadBalancer: servicemesh.LoadBalancerPolicy{
        Type: "round_robin",
    },
    Retries: &servicemesh.RetryPolicy{
        Attempts:      3,
        PerTryTimeout: time.Second,
    },
    CircuitBreaker: &servicemesh.CircuitBreakerPolicy{
        MaxConnections: 1000,
    },
}
err = sm.SetPolicy(ctx, "myapp", policy)
```

#### Educational Value

- **Service Mesh Concepts**: Learn Envoy and Istio patterns
- **Traffic Management**: Understand load balancing and routing
- **Security**: Practice mTLS and zero-trust networking
- **Resilience**: Learn circuit breakers and retries

---

### 7.5 Advanced Security

Enterprise-grade security with OPA policies, image signing, and runtime protection.

#### Security Components

1. **OPA (Open Policy Agent)** - Policy-based access control
2. **Cosign** - Container image signing and verification
3. **Runtime Security** - Behavior monitoring and threat detection
4. **Security Scanning** - Vulnerability and compliance scanning

#### Implementation

**Package**: `pkg/security/advanced`

**Features**:
- OPA policy enforcement for admission control
- Image signing and verification with Sigstore/Cosign
- Runtime security monitoring
- Vulnerability scanning integration
- Compliance reporting (CIS, PCI-DSS)
- Security event auditing

**OPA Integration**:
```go
// PolicyEngine evaluates OPA policies
type PolicyEngine struct {
    rego *rego.Rego
}

// Policy represents a security policy
type Policy struct {
    Name        string
    Description string
    Rego        string
    Severity    string
}

// Evaluate evaluates policy against input
func (pe *PolicyEngine) Evaluate(ctx context.Context, input interface{}) (bool, error) {
    results, err := pe.rego.Eval(ctx, rego.EvalInput(input))
    if err != nil {
        return false, err
    }

    // Check if policy allows
    allowed := results.Allowed()
    return allowed, nil
}
```

**Example OPA Policy**:
```rego
# /etc/containr/policies/container-security.rego
package containr.container

# Deny privileged containers
deny[msg] {
    input.container.privileged == true
    msg := "Privileged containers are not allowed"
}

# Require non-root user
deny[msg] {
    input.container.user == "root"
    msg := "Containers must run as non-root user"
}

# Restrict capabilities
deny[msg] {
    cap := input.container.capabilities.add[_]
    cap == "SYS_ADMIN"
    msg := "SYS_ADMIN capability is not allowed"
}

# Enforce image signing
deny[msg] {
    not input.image.signed
    msg := "Images must be signed with Cosign"
}

# Require specific registry
deny[msg] {
    not startswith(input.image.name, "registry.example.com/")
    msg := "Images must come from approved registry"
}
```

**Cosign Integration**:
```go
import "github.com/sigstore/cosign/v2/pkg/cosign"

// ImageVerifier verifies image signatures
type ImageVerifier struct {
    publicKey []byte
}

// Verify verifies image signature
func (iv *ImageVerifier) Verify(ctx context.Context, imageRef string) error {
    // Verify signature
    _, err := cosign.VerifyImageSignature(ctx, imageRef, iv.publicKey)
    if err != nil {
        return fmt.Errorf("image signature verification failed: %w", err)
    }

    // Verify attestations
    attestations, err := cosign.FetchAttestations(ctx, imageRef)
    if err != nil {
        return err
    }

    // Check SBOM attestation
    sbomFound := false
    for _, att := range attestations {
        if att.Type == "application/vnd.in-toto+json" {
            sbomFound = true
            break
        }
    }

    if !sbomFound {
        return errors.New("SBOM attestation not found")
    }

    return nil
}
```

**Runtime Security**:
```go
// RuntimeMonitor monitors container behavior
type RuntimeMonitor struct {
    rules []SecurityRule
}

// SecurityRule defines runtime behavior rules
type SecurityRule struct {
    Name        string
    Description string
    Severity    string
    Condition   func(event Event) bool
    Action      func(event Event) error
}

// Monitor monitors container events
func (rm *RuntimeMonitor) Monitor(ctx context.Context, containerID string) error {
    // Monitor system calls
    go rm.monitorSyscalls(ctx, containerID)

    // Monitor network connections
    go rm.monitorNetwork(ctx, containerID)

    // Monitor file access
    go rm.monitorFileAccess(ctx, containerID)

    return nil
}
```

**Usage**:
```bash
# Load OPA policies
containr policy load /etc/containr/policies/

# List policies
containr policy ls

# Test policy
containr policy test container-security --input test.json

# Sign image
cosign sign --key cosign.key registry.example.com/myapp:latest

# Verify image before run
containr run --verify-signature registry.example.com/myapp:latest /app

# Enable runtime security
containr security runtime enable --rules /etc/containr/security-rules.yaml

# Scan image for vulnerabilities
containr security scan myapp:latest

# Generate compliance report
containr security compliance report --standard cis --output report.pdf

# View security events
containr security events --severity high

# Block suspicious container
containr security block <container-id>
```

**Configuration**:
```yaml
# /etc/containr/security-advanced.yaml
security:
  opa:
    enabled: true
    policy_dir: /etc/containr/policies
    decision_logs: true

  image_verification:
    enabled: true
    require_signature: true
    cosign_public_key: /etc/containr/cosign.pub
    allowed_registries:
      - registry.example.com
      - gcr.io

  runtime:
    enabled: true
    rules: /etc/containr/security-rules.yaml
    actions:
      - type: log
      - type: alert
      - type: block

  scanning:
    enabled: true
    scanner: trivy
    scan_on_pull: true
    max_severity: high

  compliance:
    standards:
      - cis
      - pci-dss
    reporting:
      enabled: true
      schedule: "0 0 * * *"
```

**Security Rules**:
```yaml
# /etc/containr/security-rules.yaml
rules:
  - name: detect-privilege-escalation
    description: Detect privilege escalation attempts
    severity: critical
    syscalls:
      - setuid
      - setgid
      - capset
    action: block

  - name: detect-suspicious-network
    description: Detect connections to suspicious IPs
    severity: high
    network:
      blacklist_ips:
        - 192.0.2.0/24
    action: alert

  - name: detect-crypto-mining
    description: Detect cryptocurrency mining
    severity: high
    processes:
      - xmrig
      - ethminer
    action: block
```

#### Educational Value

- **Policy as Code**: Learn declarative security policies
- **Supply Chain Security**: Understand image signing and verification
- **Runtime Security**: Learn behavior-based threat detection
- **Compliance**: Practice security standards implementation

---

### 7.6 CSI Driver Support & Volume Encryption

Container Storage Interface (CSI) driver support and encrypted volumes.

#### What is CSI?

Container Storage Interface (CSI) is a standard for exposing storage systems to container orchestrators. Benefits:
- **Vendor Neutral**: Works with any storage provider
- **Dynamic Provisioning**: Automatic volume creation
- **Snapshots**: Volume snapshot support
- **Cloning**: Volume cloning capabilities
- **Expansion**: Dynamic volume resizing

#### Implementation

**Package**: `pkg/storage/csi`

**Features**:
- CSI driver plugin architecture
- Volume encryption with LUKS
- Dynamic volume provisioning
- Volume snapshots and clones
- Volume expansion
- Multiple backend support (local, NFS, Ceph, AWS EBS)

**CSI Interface**:
```go
// CSIDriver implements CSI specification
type CSIDriver interface {
    // Identity Service
    GetPluginInfo(ctx context.Context) (*PluginInfo, error)
    GetPluginCapabilities(ctx context.Context) (*PluginCapabilities, error)

    // Controller Service
    CreateVolume(ctx context.Context, req *CreateVolumeRequest) (*Volume, error)
    DeleteVolume(ctx context.Context, volumeID string) error
    ControllerPublishVolume(ctx context.Context, req *ControllerPublishRequest) error
    ControllerUnpublishVolume(ctx context.Context, volumeID, nodeID string) error

    // Node Service
    NodeStageVolume(ctx context.Context, req *NodeStageRequest) error
    NodeUnstageVolume(ctx context.Context, volumeID, stagingPath string) error
    NodePublishVolume(ctx context.Context, req *NodePublishRequest) error
    NodeUnpublishVolume(ctx context.Context, volumeID, targetPath string) error
}

// Volume represents a CSI volume
type Volume struct {
    ID         string
    Name       string
    Size       int64
    Encrypted  bool
    SnapshotID string
    Labels     map[string]string
}
```

**Volume Encryption**:
```go
// EncryptedVolume provides LUKS encryption
type EncryptedVolume struct {
    device     string
    mountPoint string
    keyFile    string
}

// Encrypt encrypts a volume with LUKS
func (ev *EncryptedVolume) Encrypt(ctx context.Context, passphrase string) error {
    // Format with LUKS
    cmd := exec.CommandContext(ctx,
        "cryptsetup",
        "luksFormat",
        "--type", "luks2",
        ev.device,
        ev.keyFile,
    )

    if err := cmd.Run(); err != nil {
        return fmt.Errorf("LUKS format failed: %w", err)
    }

    return nil
}

// Open opens encrypted volume
func (ev *EncryptedVolume) Open(ctx context.Context) error {
    mappedName := filepath.Base(ev.device) + "-encrypted"

    cmd := exec.CommandContext(ctx,
        "cryptsetup",
        "luksOpen",
        ev.device,
        mappedName,
        "--key-file", ev.keyFile,
    )

    return cmd.Run()
}
```

**CSI Drivers**:
```go
// Local CSI driver for local storage
type LocalCSIDriver struct {
    storageDir string
}

// NFS CSI driver for NFS storage
type NFSCSIDriver struct {
    server string
    export string
}

// Ceph CSI driver for Ceph RBD
type CephCSIDriver struct {
    monitors []string
    pool     string
}
```

**Usage**:
```bash
# List CSI drivers
containr storage driver ls

# Install CSI driver
containr storage driver install local

# Create encrypted volume
containr volume create mydata \
    --size 10Gi \
    --encrypted \
    --driver local

# Create volume from snapshot
containr volume create mydata-clone \
    --from-snapshot snap-12345 \
    --driver local

# Expand volume
containr volume expand mydata --size 20Gi

# Create snapshot
containr volume snapshot mydata snap-backup

# List volumes
containr volume ls --driver local

# Volume details
containr volume inspect mydata

# Use encrypted volume
containr run -v mydata:/data alpine /bin/sh

# Backup volume
containr volume backup mydata --output backup.tar.gz

# Restore volume
containr volume restore backup.tar.gz --name mydata-restored
```

**Configuration**:
```yaml
# /etc/containr/csi.yaml
csi:
  enabled: true
  socket: /var/run/containr-csi.sock

  drivers:
    local:
      enabled: true
      storage_dir: /var/lib/containr/volumes
      default_size: 10Gi

    nfs:
      enabled: false
      server: nfs.example.com
      export: /exports/containr

    ceph:
      enabled: false
      monitors:
        - mon1.example.com:6789
        - mon2.example.com:6789
      pool: containr

  encryption:
    enabled: true
    default_cipher: aes-xts-plain64
    key_size: 256
    key_file: /etc/containr/volume-keys/%s.key

  snapshots:
    enabled: true
    retention: 7d
```

**API Example**:
```go
import "github.com/therealutkarshpriyadarshi/containr/pkg/storage/csi"

// Create CSI driver
driver := csi.NewLocalDriver("/var/lib/containr/volumes")

// Create encrypted volume
vol, err := driver.CreateVolume(ctx, &csi.CreateVolumeRequest{
    Name:       "mydata",
    Size:       10 * 1024 * 1024 * 1024,
    Encrypted:  true,
    Parameters: map[string]string{
        "cipher":   "aes-xts-plain64",
        "key_size": "256",
    },
})

// Create snapshot
snapshot, err := driver.CreateSnapshot(ctx, vol.ID)

// Clone from snapshot
clone, err := driver.CreateVolume(ctx, &csi.CreateVolumeRequest{
    Name:       "mydata-clone",
    SnapshotID: snapshot.ID,
})
```

#### Educational Value

- **Storage Abstractions**: Learn storage virtualization
- **Encryption**: Understand block device encryption
- **Plugin Systems**: Work with standardized interfaces
- **Data Management**: Practice backup and recovery

---

## 📦 Package Structure

```
pkg/
├── rbac/                      # Role-Based Access Control
│   ├── rbac.go               # RBAC manager
│   ├── user.go               # User management
│   ├── role.go               # Role definitions
│   ├── quota.go              # Resource quotas
│   ├── audit.go              # Audit logging
│   └── rbac_test.go          # RBAC tests
├── observability/            # OpenTelemetry integration
│   ├── tracing.go            # Distributed tracing
│   ├── metrics.go            # Metrics collection
│   ├── logging.go            # Structured logging
│   ├── exporter.go           # Exporters (Prometheus, Jaeger)
│   └── observability_test.go # Observability tests
├── checkpoint/               # Container checkpointing
│   ├── checkpoint.go         # Checkpoint interface
│   ├── criu.go               # CRIU integration
│   ├── migration.go          # Live migration
│   ├── state.go              # State serialization
│   └── checkpoint_test.go    # Checkpoint tests
├── servicemesh/              # Service mesh integration
│   ├── servicemesh.go        # Service mesh manager
│   ├── envoy.go              # Envoy sidecar
│   ├── policy.go             # Traffic policies
│   ├── mtls.go               # mTLS configuration
│   └── servicemesh_test.go   # Service mesh tests
├── security/
│   └── advanced/             # Advanced security
│       ├── opa.go            # OPA policy engine
│       ├── cosign.go         # Image verification
│       ├── runtime.go        # Runtime security
│       ├── scanner.go        # Vulnerability scanning
│       └── advanced_test.go  # Advanced security tests
└── storage/
    └── csi/                  # CSI driver support
        ├── csi.go            # CSI interface
        ├── local.go          # Local CSI driver
        ├── nfs.go            # NFS CSI driver
        ├── encryption.go     # Volume encryption
        └── csi_test.go       # CSI tests
```

## 🎨 CLI Commands

### RBAC Commands
```bash
containr user create <username>         # Create user
containr user ls                        # List users
containr user grant <user> --role <role>  # Grant role
containr quota set <user> [OPTIONS]     # Set quota
containr audit --user <user>            # View audit log
```

### Observability Commands
```bash
containr metrics export                 # Export metrics
containr trace ls                       # List traces
containr trace inspect <id>             # Inspect trace
```

### Checkpoint Commands
```bash
containr checkpoint create <container> <name>  # Create checkpoint
containr checkpoint restore <name>             # Restore checkpoint
containr migrate <container> --to <host>       # Migrate container
```

### Service Mesh Commands
```bash
containr servicemesh policy set <service>   # Set traffic policy
containr servicemesh mtls enable <service>  # Enable mTLS
containr servicemesh split <service>        # Traffic splitting
```

### Security Commands
```bash
containr policy load <dir>              # Load OPA policies
containr security scan <image>          # Scan image
containr security runtime enable        # Enable runtime security
```

### Storage Commands
```bash
containr storage driver install <name>  # Install CSI driver
containr volume create --encrypted      # Create encrypted volume
containr volume snapshot <volume>       # Create snapshot
```

## 🧪 Testing

```bash
# Run all Phase 7 tests
make test-phase7

# Individual component tests
go test ./pkg/rbac/...
go test ./pkg/observability/...
go test ./pkg/checkpoint/...
go test ./pkg/servicemesh/...
go test ./pkg/security/advanced/...
go test ./pkg/storage/csi/...

# Integration tests
make test-rbac-integration
make test-checkpoint-integration
make test-servicemesh-integration

# E2E tests
make test-phase7-e2e
```

## 📚 Documentation

- [RBAC Guide](tutorials/08-rbac-multi-tenancy.md)
- [Observability Guide](tutorials/09-advanced-observability.md)
- [Checkpoint & Migration Guide](tutorials/10-checkpoint-migration.md)
- [Service Mesh Guide](tutorials/11-service-mesh.md)
- [Advanced Security Guide](tutorials/12-advanced-security.md)
- [CSI Storage Guide](tutorials/13-csi-storage.md)

## 🚀 Production Deployment

Phase 7 enables enterprise-ready deployments:

```yaml
# Production deployment example
version: "1.0"
services:
  app:
    image: registry.example.com/myapp:latest
    user: developer
    resources:
      limits:
        cpu: 4
        memory: 8Gi
    volumes:
      - type: csi
        driver: ceph
        name: app-data
        size: 100Gi
        encrypted: true
    security:
      verify_signature: true
      opa_policies:
        - container-security
        - network-policy
      runtime_monitoring: true
    observability:
      tracing: true
      metrics: true
    service_mesh:
      enabled: true
      mtls: true
      traffic_policy:
        retries: 3
        timeout: 10s
```

## 🎓 Educational Value

Phase 7 teaches enterprise container runtime concepts:
- **Multi-Tenancy**: User isolation and resource management
- **Production Observability**: OpenTelemetry and distributed systems
- **Stateful Containers**: Checkpointing and migration
- **Service Mesh**: Modern microservices patterns
- **Enterprise Security**: Policy enforcement and supply chain security
- **Advanced Storage**: CSI and encryption

---

**Phase 7 Status**: ✅ Complete
**Version**: 1.2.0
**Next**: Production hardening and real-world deployments

This phase completes containr's evolution into an enterprise-ready, production-grade educational container runtime! 🚀
