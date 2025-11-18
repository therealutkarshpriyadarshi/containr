package csi

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestManager_CreateVolume(t *testing.T) {
	tmpDir := t.TempDir()
	manager, driver := setupTestManager(t, tmpDir)
	defer manager.Close()

	ctx := context.Background()

	req := &CreateVolumeRequest{
		Name:          "test-volume",
		CapacityBytes: 1024 * 1024 * 100, // 100MB
		VolumeCapabilities: &VolumeCapabilities{
			AccessMode: AccessModeSingleNodeWriter,
			VolumeMode: VolumeModeFilesystem,
			FsType:     "ext4",
		},
		Parameters: map[string]string{
			"type": "local",
		},
	}

	volume, err := manager.CreateVolume(ctx, driver.name, req)
	if err != nil {
		t.Fatalf("CreateVolume failed: %v", err)
	}

	if volume.Name != "test-volume" {
		t.Errorf("expected volume name test-volume, got %s", volume.Name)
	}

	if volume.CapacityBytes != req.CapacityBytes {
		t.Errorf("expected capacity %d, got %d", req.CapacityBytes, volume.CapacityBytes)
	}

	if volume.Status != VolumeStatusReady {
		t.Errorf("expected status ready, got %s", volume.Status)
	}

	// Verify volume was persisted
	retrieved, err := manager.GetVolume(volume.ID)
	if err != nil {
		t.Fatalf("GetVolume failed: %v", err)
	}

	if retrieved.ID != volume.ID {
		t.Errorf("expected volume ID %s, got %s", volume.ID, retrieved.ID)
	}
}

func TestManager_DeleteVolume(t *testing.T) {
	tmpDir := t.TempDir()
	manager, driver := setupTestManager(t, tmpDir)
	defer manager.Close()

	ctx := context.Background()

	// Create volume
	req := &CreateVolumeRequest{
		Name:          "delete-test",
		CapacityBytes: 1024 * 1024 * 10,
		VolumeCapabilities: &VolumeCapabilities{
			AccessMode: AccessModeSingleNodeWriter,
			VolumeMode: VolumeModeFilesystem,
		},
	}

	volume, err := manager.CreateVolume(ctx, driver.name, req)
	if err != nil {
		t.Fatalf("CreateVolume failed: %v", err)
	}

	// Delete volume
	err = manager.DeleteVolume(ctx, volume.ID)
	if err != nil {
		t.Fatalf("DeleteVolume failed: %v", err)
	}

	// Verify volume was deleted
	_, err = manager.GetVolume(volume.ID)
	if err == nil {
		t.Error("expected error when getting deleted volume")
	}
}

func TestManager_ListVolumes(t *testing.T) {
	tmpDir := t.TempDir()
	manager, driver := setupTestManager(t, tmpDir)
	defer manager.Close()

	ctx := context.Background()

	// Create multiple volumes
	volumeNames := []string{"vol-1", "vol-2", "vol-3"}
	for _, name := range volumeNames {
		req := &CreateVolumeRequest{
			Name:          name,
			CapacityBytes: 1024 * 1024 * 10,
			VolumeCapabilities: &VolumeCapabilities{
				AccessMode: AccessModeSingleNodeWriter,
				VolumeMode: VolumeModeFilesystem,
			},
		}
		_, err := manager.CreateVolume(ctx, driver.name, req)
		if err != nil {
			t.Fatalf("CreateVolume failed: %v", err)
		}
	}

	// List volumes
	volumes, err := manager.ListVolumes(ctx)
	if err != nil {
		t.Fatalf("ListVolumes failed: %v", err)
	}

	if len(volumes) != len(volumeNames) {
		t.Errorf("expected %d volumes, got %d", len(volumeNames), len(volumes))
	}
}

func TestManager_CreateSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	manager, driver := setupTestManager(t, tmpDir)
	defer manager.Close()

	ctx := context.Background()

	// Create volume
	volReq := &CreateVolumeRequest{
		Name:          "snap-test-vol",
		CapacityBytes: 1024 * 1024 * 10,
		VolumeCapabilities: &VolumeCapabilities{
			AccessMode: AccessModeSingleNodeWriter,
			VolumeMode: VolumeModeFilesystem,
		},
	}

	volume, err := manager.CreateVolume(ctx, driver.name, volReq)
	if err != nil {
		t.Fatalf("CreateVolume failed: %v", err)
	}

	// Create snapshot
	snapReq := &CreateSnapshotRequest{
		VolumeID: volume.ID,
		Name:     "test-snapshot",
		Parameters: map[string]string{
			"description": "test snapshot",
		},
	}

	snapshot, err := manager.CreateSnapshot(ctx, snapReq)
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	if snapshot.VolumeID != volume.ID {
		t.Errorf("expected volume ID %s, got %s", volume.ID, snapshot.VolumeID)
	}

	if !snapshot.ReadyToUse {
		t.Error("expected snapshot to be ready to use")
	}

	// List snapshots
	snapshots, err := manager.ListSnapshots(ctx)
	if err != nil {
		t.Fatalf("ListSnapshots failed: %v", err)
	}

	if len(snapshots) != 1 {
		t.Errorf("expected 1 snapshot, got %d", len(snapshots))
	}
}

func TestManager_DeleteSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	manager, driver := setupTestManager(t, tmpDir)
	defer manager.Close()

	ctx := context.Background()

	// Create volume and snapshot
	volReq := &CreateVolumeRequest{
		Name:          "snap-delete-vol",
		CapacityBytes: 1024 * 1024 * 10,
		VolumeCapabilities: &VolumeCapabilities{
			AccessMode: AccessModeSingleNodeWriter,
			VolumeMode: VolumeModeFilesystem,
		},
	}

	volume, _ := manager.CreateVolume(ctx, driver.name, volReq)

	snapReq := &CreateSnapshotRequest{
		VolumeID: volume.ID,
		Name:     "delete-test-snap",
	}

	snapshot, err := manager.CreateSnapshot(ctx, snapReq)
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	// Delete snapshot
	err = manager.DeleteSnapshot(ctx, snapshot.ID)
	if err != nil {
		t.Fatalf("DeleteSnapshot failed: %v", err)
	}

	// Verify snapshot was deleted
	snapshots, _ := manager.ListSnapshots(ctx)
	for _, s := range snapshots {
		if s.ID == snapshot.ID {
			t.Error("snapshot still exists after deletion")
		}
	}
}

func TestManager_VolumeFromSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	manager, driver := setupTestManager(t, tmpDir)
	defer manager.Close()

	ctx := context.Background()

	// Create source volume
	volReq := &CreateVolumeRequest{
		Name:          "source-vol",
		CapacityBytes: 1024 * 1024 * 10,
		VolumeCapabilities: &VolumeCapabilities{
			AccessMode: AccessModeSingleNodeWriter,
			VolumeMode: VolumeModeFilesystem,
		},
	}

	sourceVol, _ := manager.CreateVolume(ctx, driver.name, volReq)

	// Write test data to source volume
	testFile := filepath.Join(sourceVol.Path, "test.txt")
	os.WriteFile(testFile, []byte("test data"), 0644)

	// Create snapshot
	snapReq := &CreateSnapshotRequest{
		VolumeID: sourceVol.ID,
		Name:     "restore-snap",
	}

	snapshot, _ := manager.CreateSnapshot(ctx, snapReq)

	// Create volume from snapshot
	restoreReq := &CreateVolumeRequest{
		Name:          "restored-vol",
		CapacityBytes: 1024 * 1024 * 10,
		VolumeCapabilities: &VolumeCapabilities{
			AccessMode: AccessModeSingleNodeWriter,
			VolumeMode: VolumeModeFilesystem,
		},
		ContentSource: &VolumeSource{
			Type:       "snapshot",
			SnapshotID: snapshot.ID,
		},
	}

	restoredVol, err := manager.CreateVolume(ctx, driver.name, restoreReq)
	if err != nil {
		t.Fatalf("CreateVolume from snapshot failed: %v", err)
	}

	// Verify data was restored
	restoredFile := filepath.Join(restoredVol.Path, "test.txt")
	data, err := os.ReadFile(restoredFile)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}

	if string(data) != "test data" {
		t.Errorf("expected 'test data', got '%s'", string(data))
	}
}

func TestLocalDriver_CreateVolume(t *testing.T) {
	tmpDir := t.TempDir()
	driver := setupTestLocalDriver(t, tmpDir)

	ctx := context.Background()

	req := &CreateVolumeRequest{
		Name:          "local-test",
		CapacityBytes: 1024 * 1024 * 50,
		VolumeCapabilities: &VolumeCapabilities{
			AccessMode: AccessModeSingleNodeWriter,
			VolumeMode: VolumeModeFilesystem,
		},
	}

	volume, err := driver.CreateVolume(ctx, req)
	if err != nil {
		t.Fatalf("CreateVolume failed: %v", err)
	}

	// Verify volume directory exists
	if _, err := os.Stat(volume.Path); err != nil {
		t.Errorf("volume directory does not exist: %v", err)
	}

	// Verify sparse file for size reservation
	sparseFile := filepath.Join(volume.Path, ".size")
	info, err := os.Stat(sparseFile)
	if err != nil {
		t.Errorf("sparse file does not exist: %v", err)
	}

	if info.Size() != req.CapacityBytes {
		t.Errorf("expected size %d, got %d", req.CapacityBytes, info.Size())
	}
}

func TestLocalDriver_ValidateVolumeCapabilities(t *testing.T) {
	tmpDir := t.TempDir()
	driver := setupTestLocalDriver(t, tmpDir)

	ctx := context.Background()

	// Create volume
	req := &CreateVolumeRequest{
		Name:          "validate-test",
		CapacityBytes: 1024 * 1024 * 10,
		VolumeCapabilities: &VolumeCapabilities{
			AccessMode: AccessModeSingleNodeWriter,
			VolumeMode: VolumeModeFilesystem,
		},
	}

	volume, _ := driver.CreateVolume(ctx, req)

	tests := []struct {
		name      string
		caps      *VolumeCapabilities
		expectErr bool
	}{
		{
			name: "valid filesystem single node writer",
			caps: &VolumeCapabilities{
				AccessMode: AccessModeSingleNodeWriter,
				VolumeMode: VolumeModeFilesystem,
			},
			expectErr: false,
		},
		{
			name: "valid filesystem single node readonly",
			caps: &VolumeCapabilities{
				AccessMode: AccessModeSingleNodeReadOnly,
				VolumeMode: VolumeModeFilesystem,
			},
			expectErr: false,
		},
		{
			name: "invalid block mode",
			caps: &VolumeCapabilities{
				AccessMode: AccessModeSingleNodeWriter,
				VolumeMode: VolumeModeBlock,
			},
			expectErr: true,
		},
		{
			name: "invalid multi-node access",
			caps: &VolumeCapabilities{
				AccessMode: AccessModeMultiNodeMultiWriter,
				VolumeMode: VolumeModeFilesystem,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := driver.ValidateVolumeCapabilities(ctx, volume.ID, tt.caps)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error %v, got %v", tt.expectErr, err)
			}
		})
	}
}

func TestLocalDriver_GetCapacity(t *testing.T) {
	tmpDir := t.TempDir()
	driver := setupTestLocalDriver(t, tmpDir)

	ctx := context.Background()

	capacity, err := driver.GetCapacity(ctx)
	if err != nil {
		t.Fatalf("GetCapacity failed: %v", err)
	}

	if capacity <= 0 {
		t.Errorf("expected positive capacity, got %d", capacity)
	}
}

func TestLocalDriver_NodePublishUnpublish(t *testing.T) {
	tmpDir := t.TempDir()
	driver := setupTestLocalDriver(t, tmpDir)

	ctx := context.Background()

	// Create volume
	volReq := &CreateVolumeRequest{
		Name:          "publish-test",
		CapacityBytes: 1024 * 1024 * 10,
		VolumeCapabilities: &VolumeCapabilities{
			AccessMode: AccessModeSingleNodeWriter,
			VolumeMode: VolumeModeFilesystem,
		},
	}

	volume, _ := driver.CreateVolume(ctx, volReq)

	// Stage volume
	stagePath := filepath.Join(tmpDir, "stage", volume.ID)
	stageReq := &NodeStageRequest{
		VolumeID:          volume.ID,
		StagingTargetPath: stagePath,
		VolumeCapability: &VolumeCapabilities{
			AccessMode: AccessModeSingleNodeWriter,
			VolumeMode: VolumeModeFilesystem,
		},
	}

	err := driver.NodeStageVolume(ctx, stageReq)
	if err != nil {
		t.Fatalf("NodeStageVolume failed: %v", err)
	}

	// Publish volume
	targetPath := filepath.Join(tmpDir, "target", volume.ID)
	pubReq := &NodePublishRequest{
		VolumeID:          volume.ID,
		StagingTargetPath: stagePath,
		TargetPath:        targetPath,
		VolumeCapability: &VolumeCapabilities{
			AccessMode: AccessModeSingleNodeWriter,
			VolumeMode: VolumeModeFilesystem,
		},
		Readonly: false,
	}

	err = driver.NodePublishVolume(ctx, pubReq)
	if err != nil {
		t.Fatalf("NodePublishVolume failed: %v", err)
	}

	// Verify target is mounted
	if _, err := os.Stat(targetPath); err != nil {
		t.Errorf("target path does not exist: %v", err)
	}

	// Unpublish volume
	unpubReq := &NodeUnpublishRequest{
		VolumeID:   volume.ID,
		TargetPath: targetPath,
	}

	err = driver.NodeUnpublishVolume(ctx, unpubReq)
	if err != nil {
		t.Fatalf("NodeUnpublishVolume failed: %v", err)
	}

	// Unstage volume
	unstageReq := &NodeUnstageRequest{
		VolumeID:          volume.ID,
		StagingTargetPath: stagePath,
	}

	err = driver.NodeUnstageVolume(ctx, unstageReq)
	if err != nil {
		t.Fatalf("NodeUnstageVolume failed: %v", err)
	}
}

func TestLocalDriver_PluginInfo(t *testing.T) {
	tmpDir := t.TempDir()
	driver := setupTestLocalDriver(t, tmpDir)

	ctx := context.Background()

	info, err := driver.GetPluginInfo(ctx)
	if err != nil {
		t.Fatalf("GetPluginInfo failed: %v", err)
	}

	if info.Name != "local.csi.containr.io" {
		t.Errorf("expected plugin name local.csi.containr.io, got %s", info.Name)
	}

	if info.VendorVersion == "" {
		t.Error("vendor version should not be empty")
	}
}

func TestLocalDriver_NodeInfo(t *testing.T) {
	tmpDir := t.TempDir()
	driver := setupTestLocalDriver(t, tmpDir)

	ctx := context.Background()

	info, err := driver.NodeGetInfo(ctx)
	if err != nil {
		t.Fatalf("NodeGetInfo failed: %v", err)
	}

	if info.NodeID == "" {
		t.Error("node ID should not be empty")
	}

	if info.MaxVolumesPerNode <= 0 {
		t.Errorf("max volumes should be positive, got %d", info.MaxVolumesPerNode)
	}
}

func TestEncryptionManager_CreateEncryptedVolume(t *testing.T) {
	// Skip if cryptsetup is not available
	if !isCommandAvailable("cryptsetup") {
		t.Skip("cryptsetup not available")
	}

	tmpDir := t.TempDir()
	em, err := NewEncryptionManager()
	if err != nil {
		t.Fatalf("NewEncryptionManager failed: %v", err)
	}
	defer em.Close()

	ctx := context.Background()

	volumePath := filepath.Join(tmpDir, "test-encrypted.img")
	passphrase := "test-password-123"
	sizeBytes := int64(10 * 1024 * 1024) // 10MB

	err = em.CreateEncryptedVolume(ctx, volumePath, passphrase, sizeBytes)
	if err != nil {
		t.Fatalf("CreateEncryptedVolume failed: %v", err)
	}

	// Verify volume file exists
	if _, err := os.Stat(volumePath); err != nil {
		t.Errorf("encrypted volume file does not exist: %v", err)
	}

	// Verify it's a LUKS volume
	if !em.IsEncrypted(volumePath) {
		t.Error("volume is not LUKS encrypted")
	}
}

func TestEncryptionManager_OpenCloseVolume(t *testing.T) {
	// Skip if cryptsetup is not available
	if !isCommandAvailable("cryptsetup") {
		t.Skip("cryptsetup not available")
	}

	// Skip if not running as root
	if os.Geteuid() != 0 {
		t.Skip("encryption tests require root")
	}

	tmpDir := t.TempDir()
	em, err := NewEncryptionManager()
	if err != nil {
		t.Fatalf("NewEncryptionManager failed: %v", err)
	}
	defer em.Close()

	ctx := context.Background()

	volumeID := "test-volume-123"
	volumePath := filepath.Join(tmpDir, "test-encrypted.img")
	passphrase := "test-password-456"
	sizeBytes := int64(10 * 1024 * 1024)

	// Create encrypted volume
	err = em.CreateEncryptedVolume(ctx, volumePath, passphrase, sizeBytes)
	if err != nil {
		t.Fatalf("CreateEncryptedVolume failed: %v", err)
	}

	// Open encrypted volume
	err = em.OpenEncryptedVolume(ctx, volumeID, volumePath, passphrase)
	if err != nil {
		t.Fatalf("OpenEncryptedVolume failed: %v", err)
	}

	// Get device path
	devicePath := em.GetDevicePath(volumeID)
	if devicePath == "" {
		t.Error("device path should not be empty")
	}

	// Close encrypted volume
	err = em.CloseEncryptedVolume(ctx, volumeID)
	if err != nil {
		t.Fatalf("CloseEncryptedVolume failed: %v", err)
	}

	// Verify device path is cleared
	devicePath = em.GetDevicePath(volumeID)
	if devicePath != "" {
		t.Error("device path should be empty after close")
	}
}

func TestEncryptionManager_GetVolumeInfo(t *testing.T) {
	// Skip if cryptsetup is not available
	if !isCommandAvailable("cryptsetup") {
		t.Skip("cryptsetup not available")
	}

	tmpDir := t.TempDir()
	em, err := NewEncryptionManager()
	if err != nil {
		t.Fatalf("NewEncryptionManager failed: %v", err)
	}

	ctx := context.Background()

	volumePath := filepath.Join(tmpDir, "test-info.img")
	passphrase := "test-password-789"
	sizeBytes := int64(10 * 1024 * 1024)

	// Create encrypted volume
	err = em.CreateEncryptedVolume(ctx, volumePath, passphrase, sizeBytes)
	if err != nil {
		t.Fatalf("CreateEncryptedVolume failed: %v", err)
	}

	// Get volume info
	info, err := em.GetVolumeInfo(volumePath)
	if err != nil {
		t.Fatalf("GetVolumeInfo failed: %v", err)
	}

	if info.Type != "LUKS" {
		t.Errorf("expected type LUKS, got %s", info.Type)
	}

	if info.Cipher == "" {
		t.Error("cipher should not be empty")
	}
}

func TestEncryptionManager_GenerateRandomKey(t *testing.T) {
	em, err := NewEncryptionManager()
	if err != nil {
		t.Skip("encryption manager not available")
	}

	key, err := em.GenerateRandomKey(32)
	if err != nil {
		t.Fatalf("GenerateRandomKey failed: %v", err)
	}

	if key == "" {
		t.Error("key should not be empty")
	}

	if len(key) != 64 { // 32 bytes = 64 hex characters
		t.Errorf("expected key length 64, got %d", len(key))
	}

	// Generate another key and verify they're different
	key2, _ := em.GenerateRandomKey(32)
	if key == key2 {
		t.Error("generated keys should be different")
	}
}

func TestEncryptionManager_GetEncryptionStats(t *testing.T) {
	em, err := NewEncryptionManager()
	if err != nil {
		t.Skip("encryption manager not available")
	}

	stats := em.GetEncryptionStats()

	if stats.EncryptionType != "LUKS2" {
		t.Errorf("expected encryption type LUKS2, got %s", stats.EncryptionType)
	}

	if stats.Cipher != "aes-xts-plain64" {
		t.Errorf("expected cipher aes-xts-plain64, got %s", stats.Cipher)
	}

	if stats.KeySize != 512 {
		t.Errorf("expected key size 512, got %d", stats.KeySize)
	}
}

func TestVolumeStatus(t *testing.T) {
	tests := []struct {
		status VolumeStatus
		valid  bool
	}{
		{VolumeStatusCreating, true},
		{VolumeStatusReady, true},
		{VolumeStatusPublished, true},
		{VolumeStatusDeleting, true},
		{VolumeStatusError, true},
		{VolumeStatus("invalid"), false},
	}

	validStatuses := map[VolumeStatus]bool{
		VolumeStatusCreating:  true,
		VolumeStatusReady:     true,
		VolumeStatusPublished: true,
		VolumeStatusDeleting:  true,
		VolumeStatusError:     true,
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			_, valid := validStatuses[tt.status]
			if valid != tt.valid {
				t.Errorf("expected valid %v, got %v", tt.valid, valid)
			}
		})
	}
}

func TestAccessMode(t *testing.T) {
	modes := []AccessMode{
		AccessModeSingleNodeWriter,
		AccessModeSingleNodeReadOnly,
		AccessModeMultiNodeReadOnly,
		AccessModeMultiNodeSingleWriter,
		AccessModeMultiNodeMultiWriter,
	}

	for _, mode := range modes {
		if mode == "" {
			t.Errorf("access mode should not be empty")
		}
	}
}

func TestManager_Persistence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manager and volume
	manager1, driver := setupTestManager(t, tmpDir)

	ctx := context.Background()
	req := &CreateVolumeRequest{
		Name:          "persist-test",
		CapacityBytes: 1024 * 1024 * 10,
		VolumeCapabilities: &VolumeCapabilities{
			AccessMode: AccessModeSingleNodeWriter,
			VolumeMode: VolumeModeFilesystem,
		},
	}

	volume, _ := manager1.CreateVolume(ctx, driver.name, req)
	volumeID := volume.ID

	manager1.Close()

	// Create new manager and verify volume persisted
	manager2, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager2.Close()

	// Re-register driver
	manager2.RegisterDriver(driver.name, driver)

	retrieved, err := manager2.GetVolume(volumeID)
	if err != nil {
		t.Fatalf("GetVolume failed after restart: %v", err)
	}

	if retrieved.Name != "persist-test" {
		t.Errorf("expected volume name persist-test, got %s", retrieved.Name)
	}
}

// Helper functions

func setupTestManager(t *testing.T, tmpDir string) (*Manager, *LocalDriver) {
	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	driverDir := filepath.Join(tmpDir, "local-driver")
	driver := setupTestLocalDriver(t, driverDir)

	if err := manager.RegisterDriver(driver.name, driver); err != nil {
		t.Fatalf("RegisterDriver failed: %v", err)
	}

	return manager, driver
}

func setupTestLocalDriver(t *testing.T, tmpDir string) *LocalDriver {
	driver, err := NewLocalDriver(
		"local.csi.containr.io",
		"v1.0.0",
		"test-node-1",
		tmpDir,
		nil, // encryption disabled for most tests
	)
	if err != nil {
		t.Fatalf("NewLocalDriver failed: %v", err)
	}

	return driver
}

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func TestManager_RegisterDriver(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Close()

	driver := setupTestLocalDriver(t, tmpDir)

	// Register driver
	err = manager.RegisterDriver("test-driver", driver)
	if err != nil {
		t.Fatalf("RegisterDriver failed: %v", err)
	}

	// Try to register same driver again
	err = manager.RegisterDriver("test-driver", driver)
	if err == nil {
		t.Error("expected error when registering duplicate driver")
	}

	// Get driver
	retrieved, err := manager.GetDriver("test-driver")
	if err != nil {
		t.Fatalf("GetDriver failed: %v", err)
	}

	if retrieved == nil {
		t.Error("expected non-nil driver")
	}
}

func TestCreateVolumeRequest_Validation(t *testing.T) {
	req := &CreateVolumeRequest{
		Name:          "test",
		CapacityBytes: 1024 * 1024,
		VolumeCapabilities: &VolumeCapabilities{
			AccessMode: AccessModeSingleNodeWriter,
			VolumeMode: VolumeModeFilesystem,
		},
	}

	if req.Name == "" {
		t.Error("name should not be empty")
	}

	if req.CapacityBytes <= 0 {
		t.Error("capacity should be positive")
	}

	if req.VolumeCapabilities == nil {
		t.Error("volume capabilities should not be nil")
	}
}

func TestSnapshot_CreationTime(t *testing.T) {
	snapshot := &Snapshot{
		ID:         "snap-123",
		VolumeID:   "vol-456",
		CreatedAt:  time.Now(),
		ReadyToUse: true,
	}

	if snapshot.CreatedAt.IsZero() {
		t.Error("created time should not be zero")
	}

	if time.Since(snapshot.CreatedAt) > time.Minute {
		t.Error("snapshot creation time seems incorrect")
	}
}

func TestVolume_VolumeContext(t *testing.T) {
	volume := &Volume{
		ID:   "vol-123",
		Name: "test-volume",
		VolumeContext: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	if volume.VolumeContext["key1"] != "value1" {
		t.Errorf("expected key1=value1, got %s", volume.VolumeContext["key1"])
	}

	if len(volume.VolumeContext) != 2 {
		t.Errorf("expected 2 context entries, got %d", len(volume.VolumeContext))
	}
}
