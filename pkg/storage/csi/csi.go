package csi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Driver represents a CSI driver interface
type Driver interface {
	// Identity Service
	GetPluginInfo(ctx context.Context) (*PluginInfo, error)
	GetPluginCapabilities(ctx context.Context) (*PluginCapabilities, error)

	// Controller Service
	CreateVolume(ctx context.Context, req *CreateVolumeRequest) (*Volume, error)
	DeleteVolume(ctx context.Context, volumeID string) error
	ControllerPublishVolume(ctx context.Context, req *ControllerPublishRequest) error
	ControllerUnpublishVolume(ctx context.Context, req *ControllerUnpublishRequest) error
	ValidateVolumeCapabilities(ctx context.Context, volumeID string, caps *VolumeCapabilities) error
	ListVolumes(ctx context.Context) ([]*Volume, error)
	GetCapacity(ctx context.Context) (int64, error)
	CreateSnapshot(ctx context.Context, req *CreateSnapshotRequest) (*Snapshot, error)
	DeleteSnapshot(ctx context.Context, snapshotID string) error
	ListSnapshots(ctx context.Context) ([]*Snapshot, error)

	// Node Service
	NodeStageVolume(ctx context.Context, req *NodeStageRequest) error
	NodeUnstageVolume(ctx context.Context, req *NodeUnstageRequest) error
	NodePublishVolume(ctx context.Context, req *NodePublishRequest) error
	NodeUnpublishVolume(ctx context.Context, req *NodeUnpublishRequest) error
	NodeGetInfo(ctx context.Context) (*NodeInfo, error)
	NodeGetCapabilities(ctx context.Context) (*NodeCapabilities, error)
}

// PluginInfo contains plugin information
type PluginInfo struct {
	Name          string
	VendorVersion string
}

// PluginCapabilities defines plugin capabilities
type PluginCapabilities struct {
	Service []ServiceCapability
}

// ServiceCapability represents a service capability
type ServiceCapability struct {
	Type string
}

// Volume represents a CSI volume
type Volume struct {
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	CapacityBytes     int64             `json:"capacity_bytes"`
	VolumeContext     map[string]string `json:"volume_context"`
	ContentSource     *VolumeSource     `json:"content_source,omitempty"`
	AccessibleNodes   []string          `json:"accessible_nodes,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	Path              string            `json:"path"`
	ProvisionerType   string            `json:"provisioner_type"`
	Encrypted         bool              `json:"encrypted"`
	SnapshotSupported bool              `json:"snapshot_supported"`
	Status            VolumeStatus      `json:"status"`
}

// VolumeStatus represents the status of a volume
type VolumeStatus string

const (
	VolumeStatusCreating  VolumeStatus = "creating"
	VolumeStatusReady     VolumeStatus = "ready"
	VolumeStatusPublished VolumeStatus = "published"
	VolumeStatusDeleting  VolumeStatus = "deleting"
	VolumeStatusError     VolumeStatus = "error"
)

// VolumeSource represents volume content source
type VolumeSource struct {
	Type       string `json:"type"`
	SnapshotID string `json:"snapshot_id,omitempty"`
	VolumeID   string `json:"volume_id,omitempty"`
}

// Snapshot represents a volume snapshot
type Snapshot struct {
	ID            string            `json:"id"`
	VolumeID      string            `json:"volume_id"`
	CreatedAt     time.Time         `json:"created_at"`
	SizeBytes     int64             `json:"size_bytes"`
	ReadyToUse    bool              `json:"ready_to_use"`
	SnapshotMeta  map[string]string `json:"snapshot_meta,omitempty"`
	ProvisionerID string            `json:"provisioner_id"`
}

// CreateVolumeRequest contains volume creation parameters
type CreateVolumeRequest struct {
	Name               string
	CapacityBytes      int64
	VolumeCapabilities *VolumeCapabilities
	Parameters         map[string]string
	ContentSource      *VolumeSource
}

// VolumeCapabilities defines volume capabilities
type VolumeCapabilities struct {
	AccessMode      AccessMode
	MountFlags      []string
	FsType          string
	VolumeMode      VolumeMode
	ReadOnly        bool
	EncryptionKey   string
	SnapshotEnabled bool
}

// AccessMode defines volume access mode
type AccessMode string

const (
	AccessModeSingleNodeWriter     AccessMode = "single_node_writer"
	AccessModeSingleNodeReadOnly   AccessMode = "single_node_readonly"
	AccessModeMultiNodeReadOnly    AccessMode = "multi_node_readonly"
	AccessModeMultiNodeSingleWriter AccessMode = "multi_node_single_writer"
	AccessModeMultiNodeMultiWriter AccessMode = "multi_node_multi_writer"
)

// VolumeMode defines volume mode
type VolumeMode string

const (
	VolumeModeFilesystem VolumeMode = "filesystem"
	VolumeModeBlock      VolumeMode = "block"
)

// ControllerPublishRequest contains controller publish parameters
type ControllerPublishRequest struct {
	VolumeID         string
	NodeID           string
	VolumeCapability *VolumeCapabilities
	Readonly         bool
	VolumeContext    map[string]string
}

// ControllerUnpublishRequest contains controller unpublish parameters
type ControllerUnpublishRequest struct {
	VolumeID string
	NodeID   string
}

// NodeStageRequest contains node stage parameters
type NodeStageRequest struct {
	VolumeID          string
	PublishContext    map[string]string
	StagingTargetPath string
	VolumeCapability  *VolumeCapabilities
}

// NodeUnstageRequest contains node unstage parameters
type NodeUnstageRequest struct {
	VolumeID          string
	StagingTargetPath string
}

// NodePublishRequest contains node publish parameters
type NodePublishRequest struct {
	VolumeID          string
	PublishContext    map[string]string
	StagingTargetPath string
	TargetPath        string
	VolumeCapability  *VolumeCapabilities
	Readonly          bool
}

// NodeUnpublishRequest contains node unpublish parameters
type NodeUnpublishRequest struct {
	VolumeID   string
	TargetPath string
}

// NodeInfo contains node information
type NodeInfo struct {
	NodeID             string
	MaxVolumesPerNode  int64
	AccessibleTopology map[string]string
}

// NodeCapabilities defines node capabilities
type NodeCapabilities struct {
	StageUnstageVolume bool
	GetVolumeStats     bool
}

// CreateSnapshotRequest contains snapshot creation parameters
type CreateSnapshotRequest struct {
	VolumeID   string
	Name       string
	Parameters map[string]string
}

// Manager manages CSI drivers and volumes
type Manager struct {
	drivers      map[string]Driver
	volumes      map[string]*Volume
	snapshots    map[string]*Snapshot
	mu           sync.RWMutex
	dataDir      string
	defaultDriver string
}

// NewManager creates a new CSI manager
func NewManager(dataDir string) (*Manager, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	m := &Manager{
		drivers:   make(map[string]Driver),
		volumes:   make(map[string]*Volume),
		snapshots: make(map[string]*Snapshot),
		dataDir:   dataDir,
	}

	// Load persisted volumes and snapshots
	if err := m.loadState(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return m, nil
}

// RegisterDriver registers a CSI driver
func (m *Manager) RegisterDriver(name string, driver Driver) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.drivers[name]; exists {
		return fmt.Errorf("driver %s already registered", name)
	}

	m.drivers[name] = driver

	// Set first registered driver as default
	if m.defaultDriver == "" {
		m.defaultDriver = name
	}

	return nil
}

// GetDriver retrieves a driver by name
func (m *Manager) GetDriver(name string) (Driver, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if name == "" {
		name = m.defaultDriver
	}

	driver, ok := m.drivers[name]
	if !ok {
		return nil, fmt.Errorf("driver %s not found", name)
	}

	return driver, nil
}

// CreateVolume creates a new volume
func (m *Manager) CreateVolume(ctx context.Context, driverName string, req *CreateVolumeRequest) (*Volume, error) {
	driver, err := m.GetDriver(driverName)
	if err != nil {
		return nil, err
	}

	volume, err := driver.CreateVolume(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	m.mu.Lock()
	m.volumes[volume.ID] = volume
	m.mu.Unlock()

	if err := m.saveState(); err != nil {
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	return volume, nil
}

// DeleteVolume deletes a volume
func (m *Manager) DeleteVolume(ctx context.Context, volumeID string) error {
	m.mu.RLock()
	volume, ok := m.volumes[volumeID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("volume %s not found", volumeID)
	}

	driver, err := m.GetDriver(volume.ProvisionerType)
	if err != nil {
		return err
	}

	if err := driver.DeleteVolume(ctx, volumeID); err != nil {
		return fmt.Errorf("failed to delete volume: %w", err)
	}

	m.mu.Lock()
	delete(m.volumes, volumeID)
	m.mu.Unlock()

	if err := m.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// GetVolume retrieves a volume by ID
func (m *Manager) GetVolume(volumeID string) (*Volume, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	volume, ok := m.volumes[volumeID]
	if !ok {
		return nil, fmt.Errorf("volume %s not found", volumeID)
	}

	return volume, nil
}

// ListVolumes lists all volumes
func (m *Manager) ListVolumes(ctx context.Context) ([]*Volume, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	volumes := make([]*Volume, 0, len(m.volumes))
	for _, vol := range m.volumes {
		volumes = append(volumes, vol)
	}

	return volumes, nil
}

// CreateSnapshot creates a volume snapshot
func (m *Manager) CreateSnapshot(ctx context.Context, req *CreateSnapshotRequest) (*Snapshot, error) {
	m.mu.RLock()
	volume, ok := m.volumes[req.VolumeID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("volume %s not found", req.VolumeID)
	}

	driver, err := m.GetDriver(volume.ProvisionerType)
	if err != nil {
		return nil, err
	}

	snapshot, err := driver.CreateSnapshot(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	m.mu.Lock()
	m.snapshots[snapshot.ID] = snapshot
	m.mu.Unlock()

	if err := m.saveState(); err != nil {
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	return snapshot, nil
}

// DeleteSnapshot deletes a snapshot
func (m *Manager) DeleteSnapshot(ctx context.Context, snapshotID string) error {
	m.mu.RLock()
	snapshot, ok := m.snapshots[snapshotID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("snapshot %s not found", snapshotID)
	}

	driver, err := m.GetDriver(snapshot.ProvisionerID)
	if err != nil {
		return err
	}

	if err := driver.DeleteSnapshot(ctx, snapshotID); err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	m.mu.Lock()
	delete(m.snapshots, snapshotID)
	m.mu.Unlock()

	if err := m.saveState(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// ListSnapshots lists all snapshots
func (m *Manager) ListSnapshots(ctx context.Context) ([]*Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshots := make([]*Snapshot, 0, len(m.snapshots))
	for _, snap := range m.snapshots {
		snapshots = append(snapshots, snap)
	}

	return snapshots, nil
}

// saveState persists volumes and snapshots to disk
func (m *Manager) saveState() error {
	state := struct {
		Volumes   map[string]*Volume   `json:"volumes"`
		Snapshots map[string]*Snapshot `json:"snapshots"`
	}{
		Volumes:   m.volumes,
		Snapshots: m.snapshots,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	statePath := filepath.Join(m.dataDir, "csi-state.json")
	if err := os.WriteFile(statePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// loadState loads volumes and snapshots from disk
func (m *Manager) loadState() error {
	statePath := filepath.Join(m.dataDir, "csi-state.json")

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No state file yet
		}
		return fmt.Errorf("failed to read state file: %w", err)
	}

	var state struct {
		Volumes   map[string]*Volume   `json:"volumes"`
		Snapshots map[string]*Snapshot `json:"snapshots"`
	}

	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	m.volumes = state.Volumes
	m.snapshots = state.Snapshots

	return nil
}

// Close closes the CSI manager
func (m *Manager) Close() error {
	return m.saveState()
}
