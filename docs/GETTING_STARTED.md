# Getting Started with Containr

This guide will walk you through building and using Containr, a container runtime built from Linux primitives.

## Prerequisites

### System Requirements

- **Operating System**: Linux (kernel 3.8 or later)
  - Check your kernel version: `uname -r`

- **Root Access**: Required for creating namespaces and cgroups
  - Test: `sudo echo "OK"`

- **Go**: Version 1.16 or later
  - Install: `sudo apt install golang` (Ubuntu/Debian)
  - Check: `go version`

### Verify Prerequisites

```bash
# Check kernel version (should be 3.8+)
uname -r

# Check if cgroups are available
ls /sys/fs/cgroup

# Check namespace support
ls /proc/self/ns/
```

## Installation

### Option 1: Build from Source

```bash
# 1. Clone the repository
git clone https://github.com/therealutkarshpriyadarshi/containr.git
cd containr

# 2. Download dependencies
make deps

# 3. Build the binary
make build

# 4. Verify the build
ls bin/containr
```

### Option 2: System-wide Installation

```bash
# Build and install to /usr/local/bin
sudo make install

# Verify installation
which containr
containr help
```

## Your First Container

### Step 1: Simple Command

Run a simple command in an isolated container:

```bash
sudo ./bin/containr run /bin/echo "Hello from container!"
```

You should see:
```
Starting container container-XXXX...
Namespaces: UTS, PID, Mount, IPC, Network
Command: [/bin/echo Hello from container!]
---
Container starting (PID: 1)
Hello from container!

Container container-XXXX exited
```

### Step 2: Interactive Shell

Run an interactive shell:

```bash
sudo ./bin/containr run /bin/sh
```

Inside the container, try:
```bash
# Check your hostname (should be "container")
hostname

# Check process list (isolated PID namespace)
ps aux

# Check your PID (should be 1!)
echo $$

# Exit the container
exit
```

### Step 3: Verify Isolation

Let's prove the container is actually isolated:

```bash
# In one terminal, start a container
sudo ./bin/containr run /bin/sh

# Inside the container, check the hostname
hostname
# Output: container

# Check process list
ps aux
# Should show only processes inside the container

# In another terminal (on host), check hostname
hostname
# Output: your-actual-hostname (different!)

# Check process list on host
ps aux
# Shows all system processes
```

## Understanding Namespaces

### UTS Namespace (Hostname Isolation)

```bash
# Start container and check hostname
sudo ./bin/containr run /bin/bash -c "hostname && cat /etc/hostname"

# Compare with host
hostname
```

### PID Namespace (Process Isolation)

```bash
# Inside container, your shell is PID 1
sudo ./bin/containr run /bin/sh -c "echo 'My PID:' \$\$ && ps aux"

# On host, check the actual PID
ps aux | grep containr
```

### Mount Namespace (Filesystem Isolation)

```bash
# Inside container
sudo ./bin/containr run /bin/sh

# Try creating a mount point
mkdir -p /tmp/test-mount

# This mount only exists inside the container
mount -t tmpfs tmpfs /tmp/test-mount

# Exit and check on host - mount is gone!
```

## Working with Resource Limits

### Running the Example with Cgroups

```bash
# Build and run the example
make run-example
```

This demonstrates:
- Memory limit: 100 MB
- CPU shares: 512 (half of default)
- PID limit: 100 processes

### Create Your Own Limits

Create a file `my-container.go`:

```go
package main

import (
    "fmt"
    "os"
    "github.com/therealutkarshpriyadarshi/containr/pkg/container"
    "github.com/therealutkarshpriyadarshi/containr/pkg/cgroup"
)

func main() {
    // Create resource-limited cgroup
    cg, _ := cgroup.New(&cgroup.Config{
        Name:        "my-container",
        MemoryLimit: 50 * 1024 * 1024,  // 50 MB
        CPUShares:   256,                 // 1/4 CPU
        PIDLimit:    50,                  // Max 50 processes
    })
    defer cg.Remove()
    cg.AddCurrentProcess()

    // Run container
    c := container.New(&container.Config{
        Command:  []string{"/bin/sh", "-c", "echo 'Limited container'; free -m"},
        Hostname: "limited",
    })

    c.RunWithSetup()
}
```

Run it:
```bash
go run my-container.go
```

## Creating a Root Filesystem

### Option 1: Use Existing System

```bash
# Create a minimal rootfs
mkdir -p /tmp/myrootfs/{bin,lib,lib64,proc,sys,dev}

# Copy essential binaries
cp /bin/sh /tmp/myrootfs/bin/
cp /bin/ls /tmp/myrootfs/bin/
cp /bin/echo /tmp/myrootfs/bin/

# Copy required libraries
ldd /bin/sh | grep -o '/lib.*\.[0-9]' | xargs -I {} cp {} /tmp/myrootfs/lib/

# Run container with custom rootfs
sudo ./bin/containr run /bin/sh
```

### Option 2: Use Docker Export

```bash
# Export a Docker container as tarball
docker export $(docker create alpine) > alpine.tar

# Import into containr (future feature)
# containr import alpine.tar alpine:latest
```

### Option 3: Use debootstrap (Debian/Ubuntu)

```bash
# Install debootstrap
sudo apt install debootstrap

# Create Ubuntu rootfs
sudo debootstrap focal /tmp/ubuntu-rootfs

# Use it (when rootfs support is added)
```

## Testing Network Isolation

### Setup (requires additional implementation)

```bash
# Check network namespaces
sudo ./bin/containr run /bin/sh -c "ip addr show"

# Should show only loopback interface
```

## Debugging

### Enable Verbose Output

Modify `cmd/containr/main.go` to add debug output:

```go
fmt.Printf("Debug: Creating container with config: %+v\n", config)
```

### Check Namespaces

```bash
# Run container and check its namespaces
sudo ./bin/containr run /bin/sh &
PID=$!

# Check namespace IDs
ls -la /proc/$PID/ns/
```

### Check Cgroups

```bash
# After starting a container with cgroups
ls /sys/fs/cgroup/memory/my-container/

# View memory usage
cat /sys/fs/cgroup/memory/my-container/memory.usage_in_bytes
```

### Common Issues

**Issue**: "Operation not permitted"
```bash
# Solution: Run with sudo
sudo ./bin/containr run /bin/sh
```

**Issue**: "No such file or directory" when running commands
```bash
# Solution: Ensure the binary exists in the rootfs
ls /bin/sh  # Check if shell exists
```

**Issue**: Cgroups not working
```bash
# Check cgroup version
mount | grep cgroup

# Enable cgroup v2 (if needed)
sudo mkdir -p /sys/fs/cgroup
sudo mount -t cgroup2 none /sys/fs/cgroup
```

## Next Steps

1. **Explore the API**: Check out `examples/simple.go`
2. **Read Architecture**: See [ARCHITECTURE.md](ARCHITECTURE.md)
3. **Customize**: Modify the code to add features
4. **Experiment**: Try different namespace combinations
5. **Build Images**: Create custom root filesystems

## Exercises

### Exercise 1: CPU Limit Test

Create a container that uses maximum CPU and observe the cgroup limit:

```bash
# Inside container
while true; do echo "busy"; done &
# Watch CPU usage - it should be limited by cgroup
```

### Exercise 2: Memory Limit Test

Try to allocate more memory than the limit:

```bash
# This should fail or be killed by OOM killer
stress --vm 1 --vm-bytes 200M
```

### Exercise 3: Custom Hostname

Modify the code to accept hostname as a command-line flag.

### Exercise 4: PID Namespace

Verify PID namespace isolation:
```bash
sudo ./bin/containr run /bin/sh -c "echo \$\$ && sleep 100" &
# Check the PID inside container (should be 1)
# Check the PID outside (should be different)
```

## Resources

- [Linux Namespaces](https://man7.org/linux/man-pages/man7/namespaces.7.html)
- [Cgroups](https://www.kernel.org/doc/Documentation/cgroup-v2.txt)
- [Container Security](https://www.kernel.org/doc/Documentation/security/)

## Getting Help

If you're stuck:
1. Check this guide again
2. Read the [architecture documentation](ARCHITECTURE.md)
3. Look at the example code
4. Open an issue on GitHub

Happy learning! 🚀
