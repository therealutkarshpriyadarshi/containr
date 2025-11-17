package csi

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// NFSDriver implements CSI driver for NFS storage
type NFSDriver struct {
	name           string
	version        string
	nodeID         string
	server         string
	exportPath     string
	mountOptions   []string
	dataDir        string
	maxVolumesPerNode int64
	mu             sync.RWMutex
}

// NFSConfig holds NFS driver configuration
type NFSConfig struct {
	Server       string   `json:"server"`
	ExportPath   string   `json:"export_path"`
	MountOptions []string `json:"mount_options"`
}

// NewNFSDriver creates a new NFS storage driver
func NewNFSDriver(name, version, nodeID, dataDir string, config *NFSConfig) (*NFSDriver, error) {
	if config.Server == "" {
		return nil, fmt.Errorf("NFS server is required")
	}

	if config.ExportPath == "" {
		return nil, fmt.Errorf("NFS export path is required")
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Default mount options
	if len(config.MountOptions) == 0 {
		config.MountOptions = []string{"vers=4.1", "rsize=1048576", "wsize=1048576", "hard", "timeo=600", "retrans=2"}
	}

	driver := &NFSDriver{
		name:              name,
		version:           version,
		nodeID:            nodeID,
		server:            config.Server,
		exportPath:        config.ExportPath,
		mountOptions:      config.MountOptions,
		dataDir:           dataDir,
		maxVolumesPerNode: 1000, // NFS can handle many volumes
	}

	// Verify NFS server connectivity
	if err := driver.verifyNFSServer(); err != nil {
		return nil, fmt.Errorf("failed to verify NFS server: %w", err)
	}

	return driver, nil
}

// GetPluginInfo returns plugin information
func (d *NFSDriver) GetPluginInfo(ctx context.Context) (*PluginInfo, error) {
	return &PluginInfo{
		Name:          d.name,
		VendorVersion: d.version,
	}, nil
}

// GetPluginCapabilities returns plugin capabilities
func (d *NFSDriver) GetPluginCapabilities(ctx context.Context) (*PluginCapabilities, error) {
	return &PluginCapabilities{
		Service: []ServiceCapability{
			{Type: "CONTROLLER_SERVICE"},
			{Type: "VOLUME_ACCESSIBILITY_CONSTRAINTS"},
		},
	}, nil
}

// CreateVolume creates a new NFS volume
func (d *NFSDriver) CreateVolume(ctx context.Context, req *CreateVolumeRequest) (*Volume, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	volumeID := uuid.New().String()
	volumeName := req.Name
	if volumeName == "" {
		volumeName = volumeID
	}

	// Mount base NFS export temporarily
	tempMount := filepath.Join(d.dataDir, "temp-mount-"+volumeID)
	if err := os.MkdirAll(tempMount, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp mount directory: %w", err)
	}
	defer os.RemoveAll(tempMount)

	nfsSource := fmt.Sprintf("%s:%s", d.server, d.exportPath)
	if err := d.mountNFS(nfsSource, tempMount, d.mountOptions, false); err != nil {
		return nil, fmt.Errorf("failed to mount NFS export: %w", err)
	}
	defer syscall.Unmount(tempMount, 0)

	// Create volume subdirectory on NFS
	volumePath := filepath.Join(tempMount, volumeName)
	if err := os.MkdirAll(volumePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create volume directory: %w", err)
	}

	// Set permissions
	if err := os.Chmod(volumePath, 0777); err != nil {
		return nil, fmt.Errorf("failed to set volume permissions: %w", err)
	}

	// Handle content source (clone or restore from snapshot)
	if req.ContentSource != nil {
		if err := d.restoreVolumeContent(ctx, tempMount, volumeName, req.ContentSource); err != nil {
			os.RemoveAll(volumePath)
			return nil, fmt.Errorf("failed to restore volume content: %w", err)
		}
	}

	// Create volume quota file if size specified
	if req.CapacityBytes > 0 {
		quotaFile := filepath.Join(volumePath, ".quota")
		if err := os.WriteFile(quotaFile, []byte(fmt.Sprintf("%d", req.CapacityBytes)), 0644); err != nil {
			return nil, fmt.Errorf("failed to write quota file: %w", err)
		}
	}

	volume := &Volume{
		ID:                volumeID,
		Name:              volumeName,
		CapacityBytes:     req.CapacityBytes,
		VolumeContext:     req.Parameters,
		ContentSource:     req.ContentSource,
		AccessibleNodes:   []string{}, // Accessible from all nodes
		CreatedAt:         time.Now(),
		Path:              filepath.Join(d.exportPath, volumeName),
		ProvisionerType:   d.name,
		Encrypted:         false, // NFS doesn't support encryption directly
		SnapshotSupported: true,
		Status:            VolumeStatusReady,
	}

	if volume.VolumeContext == nil {
		volume.VolumeContext = make(map[string]string)
	}
	volume.VolumeContext["server"] = d.server
	volume.VolumeContext["exportPath"] = d.exportPath
	volume.VolumeContext["volumeName"] = volumeName

	return volume, nil
}

// DeleteVolume deletes an NFS volume
func (d *NFSDriver) DeleteVolume(ctx context.Context, volumeID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Mount base NFS export temporarily
	tempMount := filepath.Join(d.dataDir, "temp-mount-"+volumeID)
	if err := os.MkdirAll(tempMount, 0755); err != nil {
		return fmt.Errorf("failed to create temp mount directory: %w", err)
	}
	defer os.RemoveAll(tempMount)

	nfsSource := fmt.Sprintf("%s:%s", d.server, d.exportPath)
	if err := d.mountNFS(nfsSource, tempMount, d.mountOptions, false); err != nil {
		return fmt.Errorf("failed to mount NFS export: %w", err)
	}
	defer syscall.Unmount(tempMount, 0)

	// Find and remove volume directory
	entries, err := os.ReadDir(tempMount)
	if err != nil {
		return fmt.Errorf("failed to read NFS export: %w", err)
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), volumeID) || entry.Name() == volumeID {
			volumePath := filepath.Join(tempMount, entry.Name())
			if err := os.RemoveAll(volumePath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove volume directory: %w", err)
			}
			break
		}
	}

	return nil
}

// ControllerPublishVolume publishes a volume to a node
func (d *NFSDriver) ControllerPublishVolume(ctx context.Context, req *ControllerPublishRequest) error {
	// NFS volumes are network-accessible, no controller publish needed
	return nil
}

// ControllerUnpublishVolume unpublishes a volume from a node
func (d *NFSDriver) ControllerUnpublishVolume(ctx context.Context, req *ControllerUnpublishRequest) error {
	// NFS volumes are network-accessible, no controller unpublish needed
	return nil
}

// ValidateVolumeCapabilities validates volume capabilities
func (d *NFSDriver) ValidateVolumeCapabilities(ctx context.Context, volumeID string, caps *VolumeCapabilities) error {
	// NFS driver supports filesystem mode
	if caps.VolumeMode != VolumeModeFilesystem {
		return fmt.Errorf("unsupported volume mode: %s", caps.VolumeMode)
	}

	// Validate access mode - NFS supports multi-node access
	switch caps.AccessMode {
	case AccessModeSingleNodeWriter, AccessModeSingleNodeReadOnly,
		AccessModeMultiNodeReadOnly, AccessModeMultiNodeMultiWriter:
		// Supported
	default:
		return fmt.Errorf("unsupported access mode: %s", caps.AccessMode)
	}

	return nil
}

// ListVolumes lists all NFS volumes
func (d *NFSDriver) ListVolumes(ctx context.Context) ([]*Volume, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Mount base NFS export temporarily
	tempMount := filepath.Join(d.dataDir, "temp-mount-list")
	if err := os.MkdirAll(tempMount, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp mount directory: %w", err)
	}
	defer os.RemoveAll(tempMount)

	nfsSource := fmt.Sprintf("%s:%s", d.server, d.exportPath)
	if err := d.mountNFS(nfsSource, tempMount, d.mountOptions, false); err != nil {
		return nil, fmt.Errorf("failed to mount NFS export: %w", err)
	}
	defer syscall.Unmount(tempMount, 0)

	entries, err := os.ReadDir(tempMount)
	if err != nil {
		return nil, fmt.Errorf("failed to read NFS export: %w", err)
	}

	volumes := make([]*Volume, 0)
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		volumePath := filepath.Join(tempMount, entry.Name())
		var capacityBytes int64

		// Read quota file if exists
		quotaFile := filepath.Join(volumePath, ".quota")
		if data, err := os.ReadFile(quotaFile); err == nil {
			fmt.Sscanf(string(data), "%d", &capacityBytes)
		}

		volume := &Volume{
			ID:              entry.Name(),
			Name:            entry.Name(),
			Path:            filepath.Join(d.exportPath, entry.Name()),
			CapacityBytes:   capacityBytes,
			ProvisionerType: d.name,
			CreatedAt:       info.ModTime(),
			Status:          VolumeStatusReady,
			VolumeContext: map[string]string{
				"server":     d.server,
				"exportPath": d.exportPath,
				"volumeName": entry.Name(),
			},
		}

		volumes = append(volumes, volume)
	}

	return volumes, nil
}

// GetCapacity returns available NFS capacity
func (d *NFSDriver) GetCapacity(ctx context.Context) (int64, error) {
	// Mount base NFS export temporarily
	tempMount := filepath.Join(d.dataDir, "temp-mount-capacity")
	if err := os.MkdirAll(tempMount, 0755); err != nil {
		return 0, fmt.Errorf("failed to create temp mount directory: %w", err)
	}
	defer os.RemoveAll(tempMount)

	nfsSource := fmt.Sprintf("%s:%s", d.server, d.exportPath)
	if err := d.mountNFS(nfsSource, tempMount, d.mountOptions, false); err != nil {
		return 0, fmt.Errorf("failed to mount NFS export: %w", err)
	}
	defer syscall.Unmount(tempMount, 0)

	var stat syscall.Statfs_t
	if err := syscall.Statfs(tempMount, &stat); err != nil {
		return 0, fmt.Errorf("failed to get filesystem stats: %w", err)
	}

	// Available = free blocks * block size
	available := int64(stat.Bavail) * int64(stat.Bsize)
	return available, nil
}

// CreateSnapshot creates an NFS volume snapshot
func (d *NFSDriver) CreateSnapshot(ctx context.Context, req *CreateSnapshotRequest) (*Snapshot, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	snapshotID := uuid.New().String()
	snapshotName := fmt.Sprintf("snapshot-%s", snapshotID)

	// Mount base NFS export temporarily
	tempMount := filepath.Join(d.dataDir, "temp-mount-"+snapshotID)
	if err := os.MkdirAll(tempMount, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp mount directory: %w", err)
	}
	defer os.RemoveAll(tempMount)

	nfsSource := fmt.Sprintf("%s:%s", d.server, d.exportPath)
	if err := d.mountNFS(nfsSource, tempMount, d.mountOptions, false); err != nil {
		return nil, fmt.Errorf("failed to mount NFS export: %w", err)
	}
	defer syscall.Unmount(tempMount, 0)

	volumePath := filepath.Join(tempMount, req.VolumeID)
	if _, err := os.Stat(volumePath); err != nil {
		return nil, fmt.Errorf("volume %s not found: %w", req.VolumeID, err)
	}

	// Create snapshots directory if not exists
	snapshotsDir := filepath.Join(tempMount, ".snapshots")
	if err := os.MkdirAll(snapshotsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	snapshotPath := filepath.Join(snapshotsDir, snapshotName)

	// Copy volume data to snapshot
	if err := d.copyDir(volumePath, snapshotPath); err != nil {
		os.RemoveAll(snapshotPath)
		return nil, fmt.Errorf("failed to copy volume data: %w", err)
	}

	// Write snapshot metadata
	metaFile := filepath.Join(snapshotPath, ".snapshot-meta")
	metadata := fmt.Sprintf("volume_id=%s\ncreated_at=%s\n", req.VolumeID, time.Now().Format(time.RFC3339))
	if err := os.WriteFile(metaFile, []byte(metadata), 0644); err != nil {
		return nil, fmt.Errorf("failed to write snapshot metadata: %w", err)
	}

	// Calculate snapshot size
	size, err := d.getDirSize(snapshotPath)
	if err != nil {
		size = 0
	}

	snapshot := &Snapshot{
		ID:            snapshotID,
		VolumeID:      req.VolumeID,
		CreatedAt:     time.Now(),
		SizeBytes:     size,
		ReadyToUse:    true,
		SnapshotMeta:  req.Parameters,
		ProvisionerID: d.name,
	}

	if snapshot.SnapshotMeta == nil {
		snapshot.SnapshotMeta = make(map[string]string)
	}
	snapshot.SnapshotMeta["snapshotName"] = snapshotName

	return snapshot, nil
}

// DeleteSnapshot deletes an NFS snapshot
func (d *NFSDriver) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	snapshotName := fmt.Sprintf("snapshot-%s", snapshotID)

	// Mount base NFS export temporarily
	tempMount := filepath.Join(d.dataDir, "temp-mount-"+snapshotID)
	if err := os.MkdirAll(tempMount, 0755); err != nil {
		return fmt.Errorf("failed to create temp mount directory: %w", err)
	}
	defer os.RemoveAll(tempMount)

	nfsSource := fmt.Sprintf("%s:%s", d.server, d.exportPath)
	if err := d.mountNFS(nfsSource, tempMount, d.mountOptions, false); err != nil {
		return fmt.Errorf("failed to mount NFS export: %w", err)
	}
	defer syscall.Unmount(tempMount, 0)

	snapshotPath := filepath.Join(tempMount, ".snapshots", snapshotName)
	if err := os.RemoveAll(snapshotPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove snapshot: %w", err)
	}

	return nil
}

// ListSnapshots lists all NFS snapshots
func (d *NFSDriver) ListSnapshots(ctx context.Context) ([]*Snapshot, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Mount base NFS export temporarily
	tempMount := filepath.Join(d.dataDir, "temp-mount-list-snapshots")
	if err := os.MkdirAll(tempMount, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp mount directory: %w", err)
	}
	defer os.RemoveAll(tempMount)

	nfsSource := fmt.Sprintf("%s:%s", d.server, d.exportPath)
	if err := d.mountNFS(nfsSource, tempMount, d.mountOptions, false); err != nil {
		return nil, fmt.Errorf("failed to mount NFS export: %w", err)
	}
	defer syscall.Unmount(tempMount, 0)

	snapshotsDir := filepath.Join(tempMount, ".snapshots")
	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Snapshot{}, nil
		}
		return nil, fmt.Errorf("failed to read snapshots directory: %w", err)
	}

	snapshots := make([]*Snapshot, 0)
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "snapshot-") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		snapshotID := strings.TrimPrefix(entry.Name(), "snapshot-")

		snapshot := &Snapshot{
			ID:            snapshotID,
			CreatedAt:     info.ModTime(),
			ReadyToUse:    true,
			ProvisionerID: d.name,
		}

		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

// NodeStageVolume stages an NFS volume on the node
func (d *NFSDriver) NodeStageVolume(ctx context.Context, req *NodeStageRequest) error {
	// Create staging directory
	if err := os.MkdirAll(req.StagingTargetPath, 0755); err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}

	return nil
}

// NodeUnstageVolume unstages an NFS volume from the node
func (d *NFSDriver) NodeUnstageVolume(ctx context.Context, req *NodeUnstageRequest) error {
	// Remove staging directory
	if err := os.RemoveAll(req.StagingTargetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove staging directory: %w", err)
	}

	return nil
}

// NodePublishVolume publishes an NFS volume on the node
func (d *NFSDriver) NodePublishVolume(ctx context.Context, req *NodePublishRequest) error {
	// Get volume name from publish context
	volumeName := req.PublishContext["volumeName"]
	if volumeName == "" {
		volumeName = req.VolumeID
	}

	// Create target directory
	if err := os.MkdirAll(req.TargetPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Mount NFS volume
	nfsSource := fmt.Sprintf("%s:%s/%s", d.server, d.exportPath, volumeName)
	mountOptions := d.mountOptions

	if err := d.mountNFS(nfsSource, req.TargetPath, mountOptions, req.Readonly); err != nil {
		return fmt.Errorf("failed to mount NFS volume: %w", err)
	}

	return nil
}

// NodeUnpublishVolume unpublishes an NFS volume from the node
func (d *NFSDriver) NodeUnpublishVolume(ctx context.Context, req *NodeUnpublishRequest) error {
	// Unmount the volume
	if err := syscall.Unmount(req.TargetPath, 0); err != nil && !os.IsNotExist(err) {
		// Try lazy unmount if normal unmount fails
		if err := syscall.Unmount(req.TargetPath, syscall.MNT_DETACH); err != nil {
			return fmt.Errorf("failed to unmount volume: %w", err)
		}
	}

	// Remove target directory
	if err := os.RemoveAll(req.TargetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove target directory: %w", err)
	}

	return nil
}

// NodeGetInfo returns node information
func (d *NFSDriver) NodeGetInfo(ctx context.Context) (*NodeInfo, error) {
	return &NodeInfo{
		NodeID:            d.nodeID,
		MaxVolumesPerNode: d.maxVolumesPerNode,
		AccessibleTopology: map[string]string{
			"hostname": d.nodeID,
			"zone":     "any",
		},
	}, nil
}

// NodeGetCapabilities returns node capabilities
func (d *NFSDriver) NodeGetCapabilities(ctx context.Context) (*NodeCapabilities, error) {
	return &NodeCapabilities{
		StageUnstageVolume: true,
		GetVolumeStats:     true,
	}, nil
}

// mountNFS mounts an NFS share
func (d *NFSDriver) mountNFS(source, target string, options []string, readonly bool) error {
	// Check if already mounted
	if d.isMounted(target) {
		return nil
	}

	opts := make([]string, len(options))
	copy(opts, options)

	if readonly {
		opts = append(opts, "ro")
	}

	optString := strings.Join(opts, ",")

	// Use mount command for NFS
	args := []string{"-t", "nfs"}
	if optString != "" {
		args = append(args, "-o", optString)
	}
	args = append(args, source, target)

	cmd := exec.Command("mount", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mount failed: %w, output: %s", err, string(output))
	}

	return nil
}

// isMounted checks if a path is mounted
func (d *NFSDriver) isMounted(target string) bool {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return false
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == target {
			return true
		}
	}

	return false
}

// verifyNFSServer verifies NFS server connectivity
func (d *NFSDriver) verifyNFSServer() error {
	// Try to showmount the NFS server
	cmd := exec.Command("showmount", "-e", d.server)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("cannot connect to NFS server %s: %w, output: %s", d.server, err, string(output))
	}

	return nil
}

// copyDir recursively copies a directory
func (d *NFSDriver) copyDir(src, dst string) error {
	// Use cp command for efficiency
	cmd := exec.Command("cp", "-a", src+"/.", dst)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to copy directory: %w, output: %s", err, string(output))
	}
	return nil
}

// getDirSize calculates the size of a directory
func (d *NFSDriver) getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// restoreVolumeContent restores volume content from source
func (d *NFSDriver) restoreVolumeContent(ctx context.Context, basePath, volumeName string, source *VolumeSource) error {
	var sourcePath string

	switch source.Type {
	case "snapshot":
		snapshotName := fmt.Sprintf("snapshot-%s", source.SnapshotID)
		sourcePath = filepath.Join(basePath, ".snapshots", snapshotName)
	case "volume":
		sourcePath = filepath.Join(basePath, source.VolumeID)
	default:
		return fmt.Errorf("unsupported content source type: %s", source.Type)
	}

	if _, err := os.Stat(sourcePath); err != nil {
		return fmt.Errorf("source not found: %w", err)
	}

	targetPath := filepath.Join(basePath, volumeName)

	// Copy source content to volume
	if err := d.copyDir(sourcePath, targetPath); err != nil {
		return fmt.Errorf("failed to copy source content: %w", err)
	}

	return nil
}

// GetVolumeStats returns NFS volume statistics
func (d *NFSDriver) GetVolumeStats(ctx context.Context, volumeID string) (int64, int64, error) {
	// Mount base NFS export temporarily
	tempMount := filepath.Join(d.dataDir, "temp-mount-stats")
	if err := os.MkdirAll(tempMount, 0755); err != nil {
		return 0, 0, fmt.Errorf("failed to create temp mount directory: %w", err)
	}
	defer os.RemoveAll(tempMount)

	nfsSource := fmt.Sprintf("%s:%s/%s", d.server, d.exportPath, volumeID)
	if err := d.mountNFS(nfsSource, tempMount, d.mountOptions, false); err != nil {
		return 0, 0, fmt.Errorf("failed to mount NFS volume: %w", err)
	}
	defer syscall.Unmount(tempMount, 0)

	var stat syscall.Statfs_t
	if err := syscall.Statfs(tempMount, &stat); err != nil {
		return 0, 0, fmt.Errorf("failed to get volume stats: %w", err)
	}

	total := int64(stat.Blocks) * int64(stat.Bsize)
	available := int64(stat.Bavail) * int64(stat.Bsize)
	used := total - available

	return used, total, nil
}
