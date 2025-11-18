# Checkpoint Package

The checkpoint package provides comprehensive container checkpointing and migration capabilities for containr using CRIU (Checkpoint/Restore In Userspace).

## Features

### Core Checkpointing
- **Checkpoint Creation**: Create checkpoints of running containers
- **Checkpoint Restoration**: Restore containers from checkpoints
- **Iterative Checkpointing**: Support for pre-dump optimization
- **Checkpoint Management**: List, get, and delete checkpoints

### Live Migration
- **Pre-Copy Migration**: Iterative pre-dump with minimal downtime
- **Stop-and-Copy Migration**: Simple migration with full checkpoint
- **Post-Copy Migration**: Lazy page migration (planned)

### State Management
- **Detailed State Tracking**: Process, filesystem, network, and memory state
- **State Persistence**: JSON-based state storage
- **Metadata and Tagging**: Organize checkpoints with metadata and tags
- **Statistics**: Track checkpoint usage and sizes

## Components

### 1. checkpoint.go
Main checkpoint interface and manager implementation.

**Key Types:**
- `Checkpointer` - Interface for checkpoint operations
- `Manager` - Main checkpoint manager
- `Checkpoint` - Checkpoint metadata
- `CheckpointOptions` - Options for checkpoint creation
- `RestoreOptions` - Options for checkpoint restoration

**Example Usage:**
```go
// Create checkpoint manager
config := &Config{
    StorePath: "/var/lib/containr/checkpoints",
    CRIUPath:  "/usr/sbin/criu",
}
mgr, err := NewManager(config)
if err != nil {
    log.Fatal(err)
}
defer mgr.Close()

// Create a checkpoint
opts := &CheckpointOptions{
    Name:         "my-checkpoint",
    LeaveRunning: true,
    TCPEstablished: true,
}
checkpoint, err := mgr.Checkpoint(ctx, containerID, opts)
if err != nil {
    log.Fatal(err)
}

// Restore from checkpoint
restoreOpts := &RestoreOptions{
    Name:   "restored-container",
    Detach: true,
}
err = mgr.Restore(ctx, checkpoint.ID, restoreOpts)
if err != nil {
    log.Fatal(err)
}
```

### 2. criu.go
CRIU integration for low-level checkpoint/restore operations.

**Key Features:**
- CRIU binary management
- Dump (checkpoint) operations
- Restore operations
- Pre-dump for iterative checkpointing
- Version checking and feature detection

**Example:**
```go
criuMgr, err := NewCRIUManager("/usr/sbin/criu")
if err != nil {
    log.Fatal(err)
}

// Check CRIU is working
if err := criuMgr.Check(); err != nil {
    log.Fatal(err)
}

// Perform checkpoint
opts := &CheckpointOptions{
    ImagePath: "/tmp/checkpoint",
    LeaveRunning: true,
}
err = criuMgr.Dump(ctx, containerID, opts)
```

### 3. migration.go
Live migration support with multiple strategies.

**Migration Strategies:**
- **Pre-Copy**: Iterative pre-dumps to minimize downtime
- **Stop-and-Copy**: Simple migration with container stop
- **Post-Copy**: Lazy page migration (not yet implemented)

**Example:**
```go
migrationOpts := &MigrationOptions{
    TargetHost: "192.168.1.100",
    TargetPort: 9000,
    Strategy:   MigrationStrategyPreCopy,
    PreDumpIterations: 3,
    PreDumpInterval: 5 * time.Second,
}

err := mgr.Migrate(ctx, containerID, migrationOpts)
if err != nil {
    log.Fatal(err)
}
```

### 4. state.go
State serialization and persistence.

**Key Features:**
- Detailed state tracking (process, filesystem, network, memory)
- JSON-based persistence
- Tagging and metadata
- Search and filtering
- Import/export capabilities

**Example:**
```go
stateStore, err := NewStateStore("/var/lib/containr/state")
if err != nil {
    log.Fatal(err)
}

// Save checkpoint state
state := &ContainerCheckpointState{
    ID:             "checkpoint-1",
    ContainerID:    "container-123",
    CheckpointName: "stable",
    Status:         "ready",
}
err = stateStore.Save(state)

// Add tags
stateStore.AddTag(state.ID, "production")

// Find by tag
states, err := stateStore.FindByTag("production")
```

### 5. checkpoint_test.go
Comprehensive tests with mocks.

**Test Coverage:**
- Manager operations (create, restore, list, delete)
- State store operations
- Mock CRIU integration
- Error handling
- Edge cases

## Requirements

### CRIU Installation
```bash
# Ubuntu/Debian
sudo apt-get install criu

# RHEL/CentOS
sudo yum install criu

# From source
git clone https://github.com/checkpoint-restore/criu
cd criu
make
sudo make install
```

### Kernel Requirements
CRIU requires specific kernel features:
- `CONFIG_CHECKPOINT_RESTORE=y`
- `CONFIG_NAMESPACES=y`
- `CONFIG_UTS_NS=y`
- `CONFIG_IPC_NS=y`
- `CONFIG_PID_NS=y`
- `CONFIG_NET_NS=y`

Check kernel config:
```bash
criu check
```

### Permissions
Checkpoint operations typically require:
- CAP_SYS_ADMIN capability
- CAP_SYS_PTRACE capability
- CAP_SYS_CHROOT capability

## Checkpoint Options

### CheckpointOptions
- `Name` - Checkpoint name
- `LeaveRunning` - Keep container running after checkpoint
- `TCPEstablished` - Include TCP connections
- `ExternalUnixSockets` - Include external UNIX sockets
- `ShellJob` - Include shell jobs
- `FileLocks` - Handle file locks
- `ImagePath` - Path to store checkpoint images
- `PreDump` - Enable iterative checkpointing
- `ParentPath` - Parent checkpoint for iterative dumps

### RestoreOptions
- `Name` - Container name after restore
- `Detach` - Detach after restore
- `TCPEstablished` - Restore TCP connections
- `ExternalUnixSockets` - Restore external sockets
- `ImagePath` - Path to checkpoint images

## Migration Options

### MigrationOptions
- `TargetHost` - Target host address
- `TargetPort` - Target port
- `Strategy` - Migration strategy (pre-copy, stop-and-copy, post-copy)
- `PreDumpIterations` - Number of pre-dump rounds
- `PreDumpInterval` - Interval between pre-dumps
- `Compression` - Enable compression
- `BandwidthLimit` - Bandwidth limit in bytes/sec

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Checkpoint Manager                  │
│  ┌───────────────────────────────────────────────┐  │
│  │              Checkpointer Interface            │  │
│  └───────────────────────────────────────────────┘  │
│         │                    │                       │
│         ▼                    ▼                       │
│  ┌──────────────┐     ┌──────────────┐             │
│  │ CRIU Manager │     │ State Store  │             │
│  └──────────────┘     └──────────────┘             │
│         │                    │                       │
│         ▼                    ▼                       │
│  ┌──────────────┐     ┌──────────────┐             │
│  │     CRIU     │     │  JSON Files  │             │
│  │   (binary)   │     │  (metadata)  │             │
│  └──────────────┘     └──────────────┘             │
└─────────────────────────────────────────────────────┘
           │
           ▼
    ┌──────────────┐
    │  Container   │
    │   Process    │
    └──────────────┘
```

## Error Handling

The package uses the containr error package for consistent error handling:

```go
if err != nil {
    if errors.Is(err, errors.ErrInternal) {
        // Handle internal error
    }
    log.WithError(err).Error("Checkpoint operation failed")
}
```

## Performance Considerations

### Checkpoint Size
Checkpoint size depends on:
- Container memory usage
- Number of open file descriptors
- Network connections
- Shared memory segments

### Downtime
Migration strategies impact downtime:
- **Pre-Copy**: Minimal downtime (seconds)
- **Stop-and-Copy**: Downtime = checkpoint + transfer time
- **Post-Copy**: Very low downtime (not implemented)

### Optimization
- Use pre-dump to reduce final checkpoint size
- Enable compression for network transfers
- Set bandwidth limits to avoid network saturation

## Limitations

1. **CRIU Limitations**:
   - External dependencies may not checkpoint properly
   - Some system calls are not supported
   - Graphics applications may have issues

2. **Current Implementation**:
   - Post-copy migration not yet implemented
   - Network transfer is basic (no encryption by default)
   - No incremental checkpoints cleanup

3. **Compatibility**:
   - Requires CRIU 3.0 or higher
   - Linux kernel 3.11 or higher recommended
   - May not work with all container configurations

## Future Enhancements

- [ ] Implement post-copy migration with page server
- [ ] Add encryption for checkpoint data transfer
- [ ] Implement incremental checkpoint cleanup
- [ ] Add checkpoint compression
- [ ] Support for distributed checkpoints
- [ ] Integration with container runtime (containerd, CRI-O)
- [ ] Checkpoint verification and validation
- [ ] Performance metrics and monitoring

## References

- [CRIU Documentation](https://criu.org)
- [CRIU GitHub](https://github.com/checkpoint-restore/criu)
- [Container Checkpoint/Restore](https://kubernetes.io/blog/2022/12/05/forensic-container-checkpointing-alpha/)
- [Live Migration Paper](https://www.usenix.org/conference/osdi20/presentation/brown)

## License

Copyright (c) 2024 containr project
