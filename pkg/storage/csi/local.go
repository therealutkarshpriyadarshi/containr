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

// LocalDriver implements CSI driver for local storage
type LocalDriver struct {
	name           string
	version        string
	nodeID         string
	dataDir        string
	maxVolumesPerNode int64
	encryption     *EncryptionManager
	mu             sync.RWMutex
}

// NewLocalDriver creates a new local storage driver
func NewLocalDriver(name, version, nodeID, dataDir string, encryption *EncryptionManager) (*LocalDriver, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &LocalDriver{
		name:              name,
		version:           version,
		nodeID:            nodeID,
		dataDir:           dataDir,
		maxVolumesPerNode: 100,
		encryption:        encryption,
	}, nil
}

// GetPluginInfo returns plugin information
func (d *LocalDriver) GetPluginInfo(ctx context.Context) (*PluginInfo, error) {
	return &PluginInfo{
		Name:          d.name,
		VendorVersion: d.version,
	}, nil
}

// GetPluginCapabilities returns plugin capabilities
func (d *LocalDriver) GetPluginCapabilities(ctx context.Context) (*PluginCapabilities, error) {
	return &PluginCapabilities{
		Service: []ServiceCapability{
			{Type: "CONTROLLER_SERVICE"},
			{Type: "VOLUME_ACCESSIBILITY_CONSTRAINTS"},
		},
	}, nil
}

// CreateVolume creates a new volume
func (d *LocalDriver) CreateVolume(ctx context.Context, req *CreateVolumeRequest) (*Volume, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	volumeID := uuid.New().String()
	volumePath := filepath.Join(d.dataDir, "volumes", volumeID)

	// Create volume directory
	if err := os.MkdirAll(volumePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create volume directory: %w", err)
	}

	volume := &Volume{
		ID:                volumeID,
		Name:              req.Name,
		CapacityBytes:     req.CapacityBytes,
		VolumeContext:     req.Parameters,
		ContentSource:     req.ContentSource,
		AccessibleNodes:   []string{d.nodeID},
		CreatedAt:         time.Now(),
		Path:              volumePath,
		ProvisionerType:   d.name,
		Encrypted:         req.VolumeCapabilities != nil && req.VolumeCapabilities.EncryptionKey != "",
		SnapshotSupported: true,
		Status:            VolumeStatusCreating,
	}

	// Handle content source (clone or restore from snapshot)
	if req.ContentSource != nil {
		if err := d.restoreVolumeContent(ctx, volume, req.ContentSource); err != nil {
			os.RemoveAll(volumePath)
			return nil, fmt.Errorf("failed to restore volume content: %w", err)
		}
	}

	// Setup encryption if requested
	if volume.Encrypted && d.encryption != nil {
		encryptedPath := volumePath + ".encrypted"
		if err := d.encryption.CreateEncryptedVolume(ctx, encryptedPath, req.VolumeCapabilities.EncryptionKey, req.CapacityBytes); err != nil {
			os.RemoveAll(volumePath)
			return nil, fmt.Errorf("failed to create encrypted volume: %w", err)
		}
		volume.Path = encryptedPath
	}

	// Allocate space for the volume
	if req.CapacityBytes > 0 && !volume.Encrypted {
		if err := d.allocateSpace(volumePath, req.CapacityBytes); err != nil {
			os.RemoveAll(volumePath)
			return nil, fmt.Errorf("failed to allocate space: %w", err)
		}
	}

	volume.Status = VolumeStatusReady
	return volume, nil
}

// DeleteVolume deletes a volume
func (d *LocalDriver) DeleteVolume(ctx context.Context, volumeID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	volumePath := filepath.Join(d.dataDir, "volumes", volumeID)

	// Close encrypted volume if needed
	encryptedPath := volumePath + ".encrypted"
	if d.encryption != nil {
		if _, err := os.Stat(encryptedPath); err == nil {
			if err := d.encryption.CloseEncryptedVolume(ctx, volumeID); err != nil {
				return fmt.Errorf("failed to close encrypted volume: %w", err)
			}
		}
	}

	// Remove volume directory
	if err := os.RemoveAll(volumePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove volume directory: %w", err)
	}

	// Remove encrypted volume file
	if err := os.RemoveAll(encryptedPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove encrypted volume: %w", err)
	}

	return nil
}

// ControllerPublishVolume publishes a volume to a node
func (d *LocalDriver) ControllerPublishVolume(ctx context.Context, req *ControllerPublishRequest) error {
	// Local driver doesn't need controller publish
	// Volumes are always accessible on the local node
	if req.NodeID != d.nodeID {
		return fmt.Errorf("volume can only be published to local node %s", d.nodeID)
	}
	return nil
}

// ControllerUnpublishVolume unpublishes a volume from a node
func (d *LocalDriver) ControllerUnpublishVolume(ctx context.Context, req *ControllerUnpublishRequest) error {
	// Local driver doesn't need controller unpublish
	return nil
}

// ValidateVolumeCapabilities validates volume capabilities
func (d *LocalDriver) ValidateVolumeCapabilities(ctx context.Context, volumeID string, caps *VolumeCapabilities) error {
	volumePath := filepath.Join(d.dataDir, "volumes", volumeID)

	if _, err := os.Stat(volumePath); err != nil {
		return fmt.Errorf("volume %s not found: %w", volumeID, err)
	}

	// Local driver supports filesystem mode
	if caps.VolumeMode != VolumeModeFilesystem {
		return fmt.Errorf("unsupported volume mode: %s", caps.VolumeMode)
	}

	// Validate access mode
	switch caps.AccessMode {
	case AccessModeSingleNodeWriter, AccessModeSingleNodeReadOnly:
		// Supported
	default:
		return fmt.Errorf("unsupported access mode: %s", caps.AccessMode)
	}

	return nil
}

// ListVolumes lists all volumes
func (d *LocalDriver) ListVolumes(ctx context.Context) ([]*Volume, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	volumesDir := filepath.Join(d.dataDir, "volumes")
	entries, err := os.ReadDir(volumesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Volume{}, nil
		}
		return nil, fmt.Errorf("failed to read volumes directory: %w", err)
	}

	volumes := make([]*Volume, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		volumePath := filepath.Join(volumesDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		volume := &Volume{
			ID:              entry.Name(),
			Path:            volumePath,
			ProvisionerType: d.name,
			CreatedAt:       info.ModTime(),
			Status:          VolumeStatusReady,
		}

		volumes = append(volumes, volume)
	}

	return volumes, nil
}

// GetCapacity returns available capacity
func (d *LocalDriver) GetCapacity(ctx context.Context) (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(d.dataDir, &stat); err != nil {
		return 0, fmt.Errorf("failed to get filesystem stats: %w", err)
	}

	// Available = free blocks * block size
	available := int64(stat.Bavail) * int64(stat.Bsize)
	return available, nil
}

// CreateSnapshot creates a volume snapshot
func (d *LocalDriver) CreateSnapshot(ctx context.Context, req *CreateSnapshotRequest) (*Snapshot, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	volumePath := filepath.Join(d.dataDir, "volumes", req.VolumeID)
	if _, err := os.Stat(volumePath); err != nil {
		return nil, fmt.Errorf("volume %s not found: %w", req.VolumeID, err)
	}

	snapshotID := uuid.New().String()
	snapshotPath := filepath.Join(d.dataDir, "snapshots", snapshotID)

	// Create snapshot directory
	if err := os.MkdirAll(snapshotPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Copy volume data to snapshot (simple copy for local driver)
	if err := d.copyDir(volumePath, snapshotPath); err != nil {
		os.RemoveAll(snapshotPath)
		return nil, fmt.Errorf("failed to copy volume data: %w", err)
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

	return snapshot, nil
}

// DeleteSnapshot deletes a snapshot
func (d *LocalDriver) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	snapshotPath := filepath.Join(d.dataDir, "snapshots", snapshotID)
	if err := os.RemoveAll(snapshotPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove snapshot: %w", err)
	}

	return nil
}

// ListSnapshots lists all snapshots
func (d *LocalDriver) ListSnapshots(ctx context.Context) ([]*Snapshot, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	snapshotsDir := filepath.Join(d.dataDir, "snapshots")
	entries, err := os.ReadDir(snapshotsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Snapshot{}, nil
		}
		return nil, fmt.Errorf("failed to read snapshots directory: %w", err)
	}

	snapshots := make([]*Snapshot, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		snapshot := &Snapshot{
			ID:            entry.Name(),
			CreatedAt:     info.ModTime(),
			ReadyToUse:    true,
			ProvisionerID: d.name,
		}

		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

// NodeStageVolume stages a volume on the node
func (d *LocalDriver) NodeStageVolume(ctx context.Context, req *NodeStageRequest) error {
	// For encrypted volumes, open the LUKS device
	if d.encryption != nil && req.VolumeCapability.EncryptionKey != "" {
		volumePath := filepath.Join(d.dataDir, "volumes", req.VolumeID) + ".encrypted"
		if _, err := os.Stat(volumePath); err == nil {
			if err := d.encryption.OpenEncryptedVolume(ctx, req.VolumeID, volumePath, req.VolumeCapability.EncryptionKey); err != nil {
				return fmt.Errorf("failed to open encrypted volume: %w", err)
			}
		}
	}

	// Create staging directory
	if err := os.MkdirAll(req.StagingTargetPath, 0755); err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}

	return nil
}

// NodeUnstageVolume unstages a volume from the node
func (d *LocalDriver) NodeUnstageVolume(ctx context.Context, req *NodeUnstageRequest) error {
	// Close encrypted volume if needed
	if d.encryption != nil {
		d.encryption.CloseEncryptedVolume(ctx, req.VolumeID)
	}

	// Remove staging directory
	if err := os.RemoveAll(req.StagingTargetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove staging directory: %w", err)
	}

	return nil
}

// NodePublishVolume publishes a volume on the node
func (d *LocalDriver) NodePublishVolume(ctx context.Context, req *NodePublishRequest) error {
	volumePath := filepath.Join(d.dataDir, "volumes", req.VolumeID)

	// Use encrypted device path if available
	if d.encryption != nil {
		encPath := d.encryption.GetDevicePath(req.VolumeID)
		if encPath != "" {
			volumePath = encPath
		}
	}

	// Create target directory
	if err := os.MkdirAll(req.TargetPath, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Mount bind the volume to target
	mountFlags := syscall.MS_BIND
	if req.Readonly {
		mountFlags |= syscall.MS_RDONLY
	}

	if err := syscall.Mount(volumePath, req.TargetPath, "", uintptr(mountFlags), ""); err != nil {
		return fmt.Errorf("failed to bind mount volume: %w", err)
	}

	return nil
}

// NodeUnpublishVolume unpublishes a volume from the node
func (d *LocalDriver) NodeUnpublishVolume(ctx context.Context, req *NodeUnpublishRequest) error {
	// Unmount the volume
	if err := syscall.Unmount(req.TargetPath, 0); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to unmount volume: %w", err)
	}

	// Remove target directory
	if err := os.RemoveAll(req.TargetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove target directory: %w", err)
	}

	return nil
}

// NodeGetInfo returns node information
func (d *LocalDriver) NodeGetInfo(ctx context.Context) (*NodeInfo, error) {
	return &NodeInfo{
		NodeID:            d.nodeID,
		MaxVolumesPerNode: d.maxVolumesPerNode,
		AccessibleTopology: map[string]string{
			"hostname": d.nodeID,
			"zone":     "local",
		},
	}, nil
}

// NodeGetCapabilities returns node capabilities
func (d *LocalDriver) NodeGetCapabilities(ctx context.Context) (*NodeCapabilities, error) {
	return &NodeCapabilities{
		StageUnstageVolume: true,
		GetVolumeStats:     true,
	}, nil
}

// allocateSpace allocates space for a volume
func (d *LocalDriver) allocateSpace(path string, sizeBytes int64) error {
	// Create a sparse file to reserve space
	sparseFile := filepath.Join(path, ".size")
	f, err := os.Create(sparseFile)
	if err != nil {
		return fmt.Errorf("failed to create sparse file: %w", err)
	}
	defer f.Close()

	if err := f.Truncate(sizeBytes); err != nil {
		return fmt.Errorf("failed to truncate sparse file: %w", err)
	}

	return nil
}

// copyDir recursively copies a directory
func (d *LocalDriver) copyDir(src, dst string) error {
	// Use cp command for efficiency
	cmd := exec.Command("cp", "-a", src+"/.", dst)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to copy directory: %w, output: %s", err, string(output))
	}
	return nil
}

// getDirSize calculates the size of a directory
func (d *LocalDriver) getDirSize(path string) (int64, error) {
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
func (d *LocalDriver) restoreVolumeContent(ctx context.Context, volume *Volume, source *VolumeSource) error {
	var sourcePath string

	switch source.Type {
	case "snapshot":
		sourcePath = filepath.Join(d.dataDir, "snapshots", source.SnapshotID)
	case "volume":
		sourcePath = filepath.Join(d.dataDir, "volumes", source.VolumeID)
	default:
		return fmt.Errorf("unsupported content source type: %s", source.Type)
	}

	if _, err := os.Stat(sourcePath); err != nil {
		return fmt.Errorf("source not found: %w", err)
	}

	// Copy source content to volume
	if err := d.copyDir(sourcePath, volume.Path); err != nil {
		return fmt.Errorf("failed to copy source content: %w", err)
	}

	return nil
}

// GetVolumeStats returns volume statistics
func (d *LocalDriver) GetVolumeStats(ctx context.Context, volumeID string) (int64, int64, error) {
	volumePath := filepath.Join(d.dataDir, "volumes", volumeID)

	var stat syscall.Statfs_t
	if err := syscall.Statfs(volumePath, &stat); err != nil {
		return 0, 0, fmt.Errorf("failed to get volume stats: %w", err)
	}

	total := int64(stat.Blocks) * int64(stat.Bsize)
	available := int64(stat.Bavail) * int64(stat.Bsize)
	used := total - available

	return used, total, nil
}

// IsVolumeAccessible checks if volume is accessible from this node
func (d *LocalDriver) IsVolumeAccessible(volumeID string) bool {
	volumePath := filepath.Join(d.dataDir, "volumes", volumeID)
	_, err := os.Stat(volumePath)
	return err == nil
}

// ExpandVolume expands a volume to a new size
func (d *LocalDriver) ExpandVolume(ctx context.Context, volumeID string, newSizeBytes int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	volumePath := filepath.Join(d.dataDir, "volumes", volumeID)
	if _, err := os.Stat(volumePath); err != nil {
		return fmt.Errorf("volume %s not found: %w", volumeID, err)
	}

	// Update the sparse file size
	sparseFile := filepath.Join(volumePath, ".size")
	f, err := os.OpenFile(sparseFile, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open sparse file: %w", err)
	}
	defer f.Close()

	if err := f.Truncate(newSizeBytes); err != nil {
		return fmt.Errorf("failed to expand volume: %w", err)
	}

	return nil
}

// CleanupOrphanedVolumes removes orphaned volumes
func (d *LocalDriver) CleanupOrphanedVolumes(ctx context.Context, activeVolumeIDs map[string]bool) error {
	volumesDir := filepath.Join(d.dataDir, "volumes")
	entries, err := os.ReadDir(volumesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read volumes directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		volumeID := entry.Name()
		if !activeVolumeIDs[volumeID] {
			volumePath := filepath.Join(volumesDir, volumeID)
			if err := os.RemoveAll(volumePath); err != nil {
				return fmt.Errorf("failed to remove orphaned volume %s: %w", volumeID, err)
			}
		}
	}

	return nil
}

// ValidatePath validates that a path is safe and within bounds
func (d *LocalDriver) ValidatePath(path string) error {
	cleanPath := filepath.Clean(path)

	// Ensure path doesn't escape data directory
	if !strings.HasPrefix(cleanPath, d.dataDir) {
		return fmt.Errorf("path %s is outside data directory", path)
	}

	return nil
}
