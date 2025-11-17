# Tutorial: Phase 6 - Cloud-Native Integration & Advanced Runtime

**Level**: Advanced
**Time**: 60 minutes
**Prerequisites**: Completion of previous tutorials or familiarity with container runtimes

## Overview

Phase 6 introduces cloud-native integration features that enable containr to work with Kubernetes and provide extensibility through plugins, snapshots, and a complete build engine.

### What You'll Learn

- How to use the CRI (Container Runtime Interface) for Kubernetes integration
- How to manage and create plugins for extending containr
- How to work with snapshots for fast container operations
- How to build container images from Dockerfiles

## 1. CRI (Container Runtime Interface)

The CRI allows containr to work as a Kubernetes container runtime.

### 1.1 Starting the CRI Server

```bash
# Start CRI server with default settings
sudo containr cri start

# Start with custom socket
sudo containr cri start --listen unix:///var/run/containr.sock

# Start with custom streaming server
sudo containr cri start --stream-addr 0.0.0.0 --stream-port 10010
```

### 1.2 Checking CRI Status

```bash
# Check server status
sudo containr cri status

# Show CRI version
sudo containr cri version
```

### 1.3 Kubernetes Integration

To use containr with Kubernetes:

```bash
# Configure kubelet to use containr
kubelet --container-runtime=remote \
        --container-runtime-endpoint=unix:///var/run/containr.sock \
        --image-service-endpoint=unix:///var/run/containr.sock
```

### 1.4 Understanding CRI Components

The CRI provides two main services:

**RuntimeService**: Manages pods and containers
- `RunPodSandbox` - Create and start a pod sandbox
- `StopPodSandbox` - Stop a running pod
- `CreateContainer` - Create a container in a pod
- `StartContainer` - Start a container
- `StopContainer` - Stop a container

**ImageService**: Manages container images
- `PullImage` - Pull an image from registry
- `ListImages` - List available images
- `RemoveImage` - Remove an image
- `ImageStatus` - Get image information

## 2. Plugin System

Plugins extend containr functionality with custom implementations.

### 2.1 Listing Plugins

```bash
# List all installed plugins
sudo containr plugin ls

# List in JSON format
sudo containr plugin ls --json
```

### 2.2 Managing Plugins

```bash
# Install a plugin
sudo containr plugin install ./my-plugin.so

# Enable a plugin
sudo containr plugin enable prometheus-exporter

# Configure a plugin
# (configuration is passed during enable)

# Disable a plugin
sudo containr plugin disable prometheus-exporter

# Remove a plugin
sudo containr plugin remove prometheus-exporter
```

### 2.3 Plugin Information

```bash
# Show plugin details
sudo containr plugin info prometheus-exporter

# Show in JSON format
sudo containr plugin info prometheus-exporter --json
```

### 2.4 Plugin Types

Containr supports five plugin types:

1. **Runtime Plugins** - Container lifecycle hooks
2. **Network Plugins** - Custom networking (CNI compatible)
3. **Storage Plugins** - Custom volume drivers
4. **Logging Plugins** - Custom log collectors
5. **Metrics Plugins** - Custom metrics exporters

### 2.5 Creating a Custom Plugin

Here's an example metrics plugin:

```go
package main

import (
    "context"
    "github.com/therealutkarshpriyadarshi/containr/pkg/plugin"
)

type MyMetricsPlugin struct {
    *plugin.BasePlugin
}

func NewMyMetricsPlugin() *MyMetricsPlugin {
    return &MyMetricsPlugin{
        BasePlugin: plugin.NewBasePlugin(
            "my-metrics",
            plugin.MetricsPlugin,
            "1.0.0",
        ),
    }
}

func (p *MyMetricsPlugin) Init(config map[string]interface{}) error {
    p.BasePlugin.Init(config)
    // Initialize your plugin
    return nil
}

func (p *MyMetricsPlugin) Start(ctx context.Context) error {
    // Start your plugin
    return nil
}

func (p *MyMetricsPlugin) Stop(ctx context.Context) error {
    // Stop your plugin
    return nil
}

func (p *MyMetricsPlugin) Health(ctx context.Context) error {
    // Health check
    return nil
}
```

## 3. Snapshot Management

Snapshots enable fast container creation and migration.

### 3.1 Creating Snapshots

```bash
# Create snapshot from a container
sudo containr snapshot create mycontainer snapshot1

# Create with labels
sudo containr snapshot create mycontainer snapshot1 \
    --label version=1.0 \
    --label env=production
```

### 3.2 Listing Snapshots

```bash
# List all snapshots
sudo containr snapshot ls

# List in JSON format
sudo containr snapshot ls --json
```

### 3.3 Inspecting Snapshots

```bash
# Show snapshot details
sudo containr snapshot inspect snapshot1

# Show in JSON
sudo containr snapshot inspect snapshot1 --json
```

### 3.4 Snapshot Operations

```bash
# Export snapshot to file
sudo containr snapshot export snapshot1 -o snapshot.tar.gz

# Import snapshot from file
sudo containr snapshot import snapshot.tar.gz

# Show differences between snapshots
sudo containr snapshot diff snapshot1 snapshot2

# Remove a snapshot
sudo containr snapshot rm snapshot1
```

### 3.5 Using Snapshots

Snapshots use overlay2 (or btrfs/zfs) for efficient storage:

```
Snapshot Hierarchy:
├── base (alpine:latest)
│   ├── snapshot1 (app v1.0)
│   │   └── snapshot2 (app v1.1)
│   └── snapshot3 (app v2.0)
```

Each snapshot only stores the differences from its parent, saving space.

## 4. Building Images

The build engine supports full Dockerfile syntax.

### 4.1 Basic Build

```bash
# Build from current directory
sudo containr build -t myapp:latest .

# Build from specific directory
sudo containr build -t myapp:latest /path/to/context

# Use custom Dockerfile
sudo containr build -f Dockerfile.production -t myapp:prod .
```

### 4.2 Build Arguments

```bash
# Pass build arguments
sudo containr build \
    --build-arg VERSION=1.0.0 \
    --build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
    -t myapp:v1 .
```

### 4.3 Multi-Stage Builds

```bash
# Target specific stage
sudo containr build --target production -t myapp:prod .

# Build all stages
sudo containr build -t myapp:latest .
```

Example multi-stage Dockerfile:

```dockerfile
# Stage 1: Build
FROM golang:1.21 AS builder
WORKDIR /src
COPY . .
RUN go build -o app

# Stage 2: Runtime
FROM alpine:latest AS production
COPY --from=builder /src/app /app
ENTRYPOINT ["/app"]
```

### 4.4 Build Cache

```bash
# Build without cache
sudo containr build --no-cache -t myapp:latest .

# Use cache from another image
sudo containr build --cache-from myapp:base -t myapp:latest .
```

### 4.5 Platform Targeting

```bash
# Build for different platform
sudo containr build --platform linux/arm64 -t myapp:arm .
```

### 4.6 Build Labels

```bash
# Add metadata labels
sudo containr build \
    --label version=1.0.0 \
    --label maintainer=team@example.com \
    -t myapp:latest .
```

## 5. Complete Workflow Example

Let's combine all Phase 6 features:

### Step 1: Build an Image

```bash
# Create a Dockerfile
cat > Dockerfile <<EOF
FROM alpine:latest
RUN apk add --no-cache curl
WORKDIR /app
COPY app.sh .
CMD ["/bin/sh", "app.sh"]
EOF

# Create app script
cat > app.sh <<EOF
#!/bin/sh
echo "Application running..."
while true; do
    date
    sleep 30
done
EOF

# Build the image
sudo containr build -t myapp:1.0 .
```

### Step 2: Run Container

```bash
# Create and start container
sudo containr create --name myapp myapp:1.0
sudo containr start myapp
```

### Step 3: Create Snapshot

```bash
# Create snapshot of running container
sudo containr snapshot create myapp myapp-snapshot-v1.0

# Inspect snapshot
sudo containr snapshot inspect myapp-snapshot-v1.0
```

### Step 4: Export and Import

```bash
# Export snapshot
sudo containr snapshot export myapp-snapshot-v1.0 -o myapp.tar.gz

# Import on another system
sudo containr snapshot import myapp.tar.gz
```

### Step 5: Monitor with Plugin

```bash
# Enable metrics plugin
sudo containr plugin enable prometheus-exporter

# Access metrics
curl http://localhost:9090/metrics
```

## 6. Best Practices

### 6.1 CRI Integration

- Use unix sockets for better security
- Configure proper streaming server settings for exec/logs
- Monitor CRI metrics for performance

### 6.2 Plugin Development

- Always implement the Health() method
- Use structured logging
- Handle configuration gracefully
- Clean up resources in Stop()

### 6.3 Snapshot Management

- Label snapshots with metadata
- Regularly prune unused snapshots
- Export important snapshots for backup
- Use overlay2 for best performance

### 6.4 Image Building

- Use multi-stage builds to reduce image size
- Leverage build cache for faster builds
- Pin base image versions for reproducibility
- Use .dockerignore to exclude unnecessary files

## 7. Troubleshooting

### CRI Issues

```bash
# Check CRI server logs
journalctl -u containr-cri

# Test CRI connectivity
crictl --runtime-endpoint unix:///var/run/containr.sock version
```

### Plugin Issues

```bash
# Check plugin status
sudo containr plugin ls

# View plugin logs (if supported)
sudo containr plugin info <plugin-name>

# Restart plugin
sudo containr plugin disable <plugin-name>
sudo containr plugin enable <plugin-name>
```

### Snapshot Issues

```bash
# Verify snapshot integrity
sudo containr snapshot inspect <snapshot>

# Check snapshot driver
ls -la /var/lib/containr/snapshots
```

### Build Issues

```bash
# Build with debug output
sudo containr build --debug -t myapp .

# Check build cache
ls -la /var/lib/containr/build-cache

# Clear build cache
sudo rm -rf /var/lib/containr/build-cache/*
```

## 8. Next Steps

- Integrate containr with Kubernetes
- Develop custom plugins for your use case
- Set up automated image builds in CI/CD
- Implement snapshot-based deployment strategies

## 9. Additional Resources

- [Phase 6 Documentation](../PHASE6.md)
- [CRI Specification](https://github.com/kubernetes/cri-api)
- [containerd Snapshotter](https://github.com/containerd/containerd/tree/main/docs/snapshotters)
- [BuildKit](https://github.com/moby/buildkit)
- [CNI Specification](https://github.com/containernetworking/cni)

## Conclusion

Phase 6 brings enterprise-grade features to containr, enabling cloud-native integration and extensibility. You now have the tools to integrate with Kubernetes, extend functionality through plugins, manage snapshots efficiently, and build container images from Dockerfiles.

Happy cloud-native containerizing! 🚀
