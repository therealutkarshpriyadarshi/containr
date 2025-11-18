# Tutorial: Phase 7 - Enterprise Production Features

This tutorial demonstrates the advanced enterprise features introduced in Phase 7 of containr.

## Prerequisites

- Containr Phase 7 installed
- Root/sudo access
- Basic understanding of containers

## Table of Contents

1. [RBAC & Multi-Tenancy](#1-rbac--multi-tenancy)
2. [Advanced Observability](#2-advanced-observability)
3. [Container Checkpointing & Migration](#3-container-checkpointing--migration)
4. [Service Mesh Integration](#4-service-mesh-integration)
5. [Advanced Security](#5-advanced-security)
6. [CSI Storage & Encryption](#6-csi-storage--encryption)

---

## 1. RBAC & Multi-Tenancy

### Creating Users and Assigning Roles

```bash
# Create a developer user
containr user create alice --role developer --email alice@example.com

# Create an operator user
containr user create bob --role operator

# List all users
containr user ls

# Output:
# USERNAME  ROLES       QUOTA           CREATED
# alice     developer   cpu:2,mem:4Gi   2025-11-17
# bob       operator    cpu:1,mem:2Gi   2025-11-17
```

### Setting Resource Quotas

```bash
# Set quota for alice
containr quota set alice \
    --max-containers 20 \
    --max-cpu 8 \
    --max-memory 16Gi \
    --max-storage 100Gi

# View quota usage
containr quota get alice

# Output:
# RESOURCE      USAGE   LIMIT   PERCENTAGE
# containers    5       20      25%
# cpu           2.5     8.0     31%
# memory        4Gi     16Gi    25%
# storage       15Gi    100Gi   15%
```

### Running Containers as Users

```bash
# Alice runs a container (has developer role)
containr run --user alice --name myapp alpine /bin/sh

# Bob tries to run a container (operator can only start/stop)
containr run --user bob --name test alpine /bin/sh
# Error: permission denied - user 'bob' does not have 'create' permission on 'container'

# Bob can start existing containers
containr start --user bob myapp
```

### Audit Logging

```bash
# View audit log for alice
containr audit --user alice --limit 10

# View all permission denials
containr audit --action permission.denied

# Export audit log for compliance
containr audit export --format json --output audit-report.json
```

---

## 2. Advanced Observability

### Configuring OpenTelemetry

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
        endpoint: http://localhost:14268/api/traces

  metrics:
    enabled: true
    provider: prometheus
    port: 9090

  logging:
    enabled: true
    level: info
    format: json
```

### Running Containers with Tracing

```bash
# Start Jaeger (for trace collection)
docker run -d --name jaeger \
    -p 16686:16686 \
    -p 14268:14268 \
    jaegertracing/all-in-one:latest

# Run container with tracing enabled
containr run --trace alpine /bin/sh -c "echo 'Hello with tracing'"

# View traces in Jaeger UI
# Open http://localhost:16686 in browser
```

### Prometheus Metrics

```bash
# Start Prometheus
docker run -d --name prometheus \
    -p 9090:9090 \
    -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
    prom/prometheus

# Configure Prometheus to scrape containr
cat > prometheus.yml <<EOF
scrape_configs:
  - job_name: 'containr'
    static_configs:
      - targets: ['localhost:9090']
EOF

# View metrics
curl http://localhost:9090/metrics | grep containr

# Example metrics:
# containr_container_created_total{status="success"} 42
# containr_container_cpu_usage_percent{id="abc123"} 15.5
# containr_container_memory_bytes{id="abc123"} 104857600
```

### Viewing Traces

```bash
# List recent traces
containr trace ls --limit 10

# Inspect specific trace
containr trace inspect <trace-id>

# Output:
# Trace ID: 7f3a2b1c4d5e6f7g
# Duration: 125ms
# Spans:
#   - container.create (85ms)
#     - namespace.setup (25ms)
#     - rootfs.prepare (35ms)
#     - network.configure (15ms)
#   - container.start (40ms)
```

---

## 3. Container Checkpointing & Migration

### Basic Checkpointing

```bash
# Start a stateful application
containr run -d --name webapp nginx

# Create checkpoint
containr checkpoint create webapp checkpoint1

# Stop the container
containr stop webapp

# Restore from checkpoint
containr checkpoint restore checkpoint1

# Container resumes exactly where it left off!
```

### Live Migration Between Hosts

```bash
# On source host
containr migrate webapp --to host2.example.com

# Behind the scenes:
# 1. Pre-dump iterations to sync memory
# 2. Final checkpoint
# 3. Transfer to destination
# 4. Restore on destination
# 5. Minimal downtime!

# Check migration status
containr migrate status webapp

# Output:
# CONTAINER  STATUS      DESTINATION       PROGRESS
# webapp     migrating   host2.example.com 75%
```

### Iterative Pre-dump for Fast Migration

```bash
# Create multiple pre-dumps
containr checkpoint create webapp pre1 --pre-dump
sleep 5
containr checkpoint create webapp pre2 --pre-dump --prev pre1
sleep 5
containr checkpoint create webapp final --prev pre2

# Restore with minimal downtime
containr checkpoint restore final

# Memory pages changed since pre-dumps:
# pre1: 100MB
# pre2: 25MB (75% already transferred)
# final: 5MB (95% already transferred)
# Downtime: ~100ms instead of ~2s!
```

### Checkpoint Export/Import

```bash
# Export checkpoint for backup
containr checkpoint export checkpoint1 -o webapp-backup.tar

# On another machine, import and restore
containr checkpoint import webapp-backup.tar
containr checkpoint restore checkpoint1
```

---

## 4. Service Mesh Integration

### Running with Envoy Sidecar

```bash
# Run container with automatic sidecar injection
containr run --sidecar envoy --service-name myapp \
    -p 8080:8080 \
    myapp:latest

# Envoy sidecar handles all traffic
# Inbound: client -> envoy:15001 -> app:8080
# Outbound: app -> envoy:15001 -> destination
```

### Configuring Traffic Policies

```bash
# Set load balancing policy
containr servicemesh policy set myapp \
    --load-balancer round-robin \
    --retries 3 \
    --timeout 10s \
    --circuit-breaker max-connections=1000

# View current policy
containr servicemesh policy get myapp

# Output:
# Service: myapp
# Load Balancer: round_robin
# Retries: 3 attempts
# Timeout: 10s
# Circuit Breaker: max_connections=1000
```

### Enabling mTLS

```bash
# Enable mTLS for service
containr servicemesh mtls enable myapp

# All service-to-service communication now encrypted!
# Certificate auto-generated and rotated

# View certificate info
containr servicemesh mtls info myapp

# Output:
# Service: myapp
# Certificate: /etc/containr/certs/myapp/cert.pem
# Expiration: 2026-11-17
# DNS Names: myapp, myapp.default, myapp.default.svc
```

### Traffic Splitting (Canary Deployment)

```bash
# Deploy v2 alongside v1
containr run --sidecar envoy --service-name myapp-v2 myapp:v2

# Split traffic: 90% v1, 10% v2
containr servicemesh split myapp \
    --v1 90 \
    --v2 10

# Gradually increase v2 traffic
containr servicemesh split myapp --v1 50 --v2 50
containr servicemesh split myapp --v1 10 --v2 90

# Full cutover
containr servicemesh split myapp --v1 0 --v2 100
```

### Fault Injection for Testing

```bash
# Inject 100ms delay for 10% of requests
containr servicemesh fault inject myapp \
    --delay 100ms \
    --percent 10

# Inject 5% failures (HTTP 503)
containr servicemesh fault inject myapp \
    --abort 503 \
    --percent 5

# Test resilience under chaos!
```

---

## 5. Advanced Security

### OPA Policy Enforcement

```bash
# Load security policies
containr policy load /etc/containr/policies/

# List loaded policies
containr policy ls

# Output:
# NAME                  SEVERITY  RULES
# container-security    high      5
# network-policy        medium    3
# image-security        critical  4

# Test policy
cat > test-input.json <<EOF
{
  "container": {
    "privileged": true,
    "user": "root"
  }
}
EOF

containr policy test container-security --input test-input.json

# Output:
# DENY: Privileged containers are not allowed
# DENY: Containers must run as non-root user
```

### Image Signature Verification

```bash
# Generate signing key
cosign generate-key-pair

# Sign image
cosign sign --key cosign.key registry.example.com/myapp:latest

# Configure containr to verify signatures
containr config set security.image_verification.enabled true
containr config set security.image_verification.require_signature true
containr config set security.image_verification.cosign_public_key /path/to/cosign.pub

# Run container - automatically verifies signature
containr run registry.example.com/myapp:latest

# Unsigned images are rejected!
containr run unsigned-image:latest
# Error: image signature verification failed
```

### Runtime Security Monitoring

```bash
# Enable runtime security
containr security runtime enable \
    --rules /etc/containr/security-rules.yaml

# Security rules detect:
# - Privilege escalation attempts
# - Suspicious network connections
# - Crypto mining processes
# - Critical file modifications

# View security events
containr security events --severity high

# Output:
# TIME                 CONTAINER  SEVERITY  THREAT              ACTION
# 2025-11-17 10:15:23  webapp     CRITICAL  privilege_escalation blocked
# 2025-11-17 10:20:45  api        HIGH      suspicious_network   alerted

# Block suspicious container
containr security block <container-id>
```

### Vulnerability Scanning

```bash
# Scan image for vulnerabilities
containr security scan myapp:latest

# Output:
# Image: myapp:latest
# Total Vulnerabilities: 15
# Critical: 2
# High: 5
# Medium: 6
# Low: 2
#
# CVE-2024-12345 | Critical | OpenSSL vulnerability
# CVE-2024-67890 | High     | HTTP library RCE

# Generate compliance report
containr security compliance report \
    --standard cis \
    --output report.pdf

# Standards supported: CIS Docker Benchmark, PCI-DSS
```

---

## 6. CSI Storage & Encryption

### Creating Encrypted Volumes

```bash
# Install CSI driver
containr storage driver install local

# Create encrypted volume
containr volume create mydata \
    --size 10Gi \
    --encrypted \
    --driver local

# Volume is automatically encrypted with LUKS!

# Use encrypted volume in container
containr run -v mydata:/data alpine /bin/sh

# Inside container, write data
echo "sensitive data" > /data/secret.txt

# On host, data is encrypted at rest
cat /var/lib/containr/volumes/mydata/data/secret.txt
# Output: binary encrypted data
```

### Volume Snapshots

```bash
# Create snapshot
containr volume snapshot mydata snap-backup

# List snapshots
containr volume snapshot ls

# Output:
# NAME         VOLUME   SIZE   CREATED
# snap-backup  mydata   10Gi   2025-11-17

# Restore from snapshot (creates new volume)
containr volume create mydata-restored \
    --from-snapshot snap-backup
```

### Volume Cloning

```bash
# Clone volume
containr volume create mydata-clone \
    --from-volume mydata

# Clone is instant (copy-on-write)!
# Only divergent changes consume additional space

# Use cloned volume
containr run -v mydata-clone:/data alpine /bin/sh
```

### NFS Volumes

```bash
# Install NFS CSI driver
containr storage driver install nfs \
    --server nfs.example.com \
    --export /exports/containr

# Create NFS volume
containr volume create shared-data \
    --size 50Gi \
    --driver nfs

# Multiple containers can mount same NFS volume
containr run -v shared-data:/shared alpine /bin/sh
containr run -v shared-data:/shared alpine /bin/sh

# Both containers share the same data!
```

---

## Combined Example: Secure Multi-Tenant Application

Let's put it all together with a realistic scenario:

```bash
# 1. Create users
containr user create alice --role developer
containr user create bob --role developer
containr quota set alice --max-containers 10 --max-cpu 4 --max-memory 8Gi
containr quota set bob --max-containers 10 --max-cpu 4 --max-memory 8Gi

# 2. Load security policies
containr policy load /etc/containr/policies/

# 3. Create encrypted volumes for each user
containr volume create alice-data --encrypted --size 20Gi --user alice
containr volume create bob-data --encrypted --size 20Gi --user bob

# 4. Run applications with observability and service mesh
containr run --user alice \
    --name alice-app \
    --sidecar envoy \
    --service-name alice-app \
    --trace \
    -v alice-data:/data \
    --verify-signature \
    registry.example.com/app:latest

containr run --user bob \
    --name bob-app \
    --sidecar envoy \
    --service-name bob-app \
    --trace \
    -v bob-data:/data \
    --verify-signature \
    registry.example.com/app:latest

# 5. Enable mTLS between services
containr servicemesh mtls enable alice-app
containr servicemesh mtls enable bob-app

# 6. Configure traffic policies
containr servicemesh policy set alice-app \
    --retries 3 \
    --timeout 10s \
    --circuit-breaker max-connections=500

# 7. Enable runtime security
containr security runtime enable

# 8. Monitor with Prometheus and Jaeger
# - Metrics: http://localhost:9090
# - Traces: http://localhost:16686
# - Alerts: containr security events

# 9. Checkpoint for backup
containr checkpoint create alice-app backup1
containr checkpoint create bob-app backup2

# 10. View audit log
containr audit --limit 100 --output audit.json
```

This gives you:
- ✅ Multi-tenancy with RBAC
- ✅ Encrypted data at rest
- ✅ Service mesh with mTLS
- ✅ Full observability
- ✅ Security policies and monitoring
- ✅ Checkpoint/restore capability
- ✅ Complete audit trail

## Next Steps

- Review [Phase 7 Documentation](../PHASE7.md) for detailed API reference
- Explore [Security Guide](../SECURITY.md) for hardening tips
- Check [Performance Guide](../PERFORMANCE.md) for optimization strategies
- Join the community in [Discussions](https://github.com/therealutkarshpriyadarshi/containr/discussions)

## Troubleshooting

### RBAC Issues

```bash
# Check user permissions
containr user inspect alice

# View denied operations
containr audit --user alice --result denied
```

### Observability Issues

```bash
# Test trace endpoint
curl http://localhost:4317

# Check metrics
curl http://localhost:9090/metrics | grep containr
```

### Checkpoint Issues

```bash
# Verify CRIU is installed
criu --version

# Check checkpoint directory permissions
ls -la /var/lib/containr/checkpoints/
```

### Service Mesh Issues

```bash
# Check Envoy sidecar status
containr servicemesh stats myapp

# View Envoy config
containr servicemesh config myapp
```

Happy containerizing with enterprise-grade features! 🚀
