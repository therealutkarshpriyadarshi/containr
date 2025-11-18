package checkpoint

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// mockCRIUManager is a mock implementation of CRIUManager for testing
type mockCRIUManager struct {
	dumpCalled    bool
	restoreCalled bool
	preDumpCalled bool
	dumpErr       error
	restoreErr    error
	preDumpErr    error
}

func newMockCRIUManager() *mockCRIUManager {
	return &mockCRIUManager{}
}

func (m *mockCRIUManager) Dump(ctx context.Context, containerID string, opts *CheckpointOptions) error {
	m.dumpCalled = true
	if m.dumpErr != nil {
		return m.dumpErr
	}

	// Create mock checkpoint files
	if opts.ImagePath != "" {
		os.MkdirAll(opts.ImagePath, 0700)
		os.WriteFile(filepath.Join(opts.ImagePath, "inventory.img"), []byte(`{"root_pid":1234}`), 0600)
		os.WriteFile(filepath.Join(opts.ImagePath, "pages.img"), []byte("mock pages data"), 0600)
	}

	return nil
}

func (m *mockCRIUManager) Restore(ctx context.Context, containerID string, opts *RestoreOptions) error {
	m.restoreCalled = true
	return m.restoreErr
}

func (m *mockCRIUManager) PreDump(ctx context.Context, containerID string, opts *CheckpointOptions) error {
	m.preDumpCalled = true
	if m.preDumpErr != nil {
		return m.preDumpErr
	}

	// Create mock pre-dump files
	if opts.ImagePath != "" {
		os.MkdirAll(opts.ImagePath, 0700)
		os.WriteFile(filepath.Join(opts.ImagePath, "pages.img"), []byte("mock pre-dump data"), 0600)
	}

	return nil
}

func (m *mockCRIUManager) Check() error {
	return nil
}

func (m *mockCRIUManager) GetVersion() string {
	return "3.17.1"
}

// setupTestManager creates a test checkpoint manager with mock CRIU
func setupTestManager(t *testing.T) (*Manager, *mockCRIUManager, string) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "checkpoints")
	statePath := filepath.Join(tmpDir, "state")

	mockCRIU := newMockCRIUManager()

	stateStore, err := NewStateStore(statePath)
	if err != nil {
		t.Fatalf("Failed to create state store: %v", err)
	}

	mgr := &Manager{
		criuManager: mockCRIU,
		stateStore:  stateStore,
		checkpoints: make(map[string]*Checkpoint),
		storePath:   storePath,
		criuPath:    "/usr/sbin/criu",
	}

	if err := os.MkdirAll(storePath, 0700); err != nil {
		t.Fatalf("Failed to create store path: %v", err)
	}

	return mgr, mockCRIU, tmpDir
}

func TestManager_Checkpoint(t *testing.T) {
	mgr, mockCRIU, _ := setupTestManager(t)
	ctx := context.Background()

	containerID := "test-container-123"
	opts := &CheckpointOptions{
		Name:         "test-checkpoint",
		LeaveRunning: true,
	}

	checkpoint, err := mgr.Checkpoint(ctx, containerID, opts)
	if err != nil {
		t.Fatalf("Checkpoint failed: %v", err)
	}

	// Verify checkpoint was created
	if checkpoint == nil {
		t.Fatal("Expected checkpoint to be created")
	}

	if checkpoint.ContainerID != containerID {
		t.Errorf("Expected container ID %s, got %s", containerID, checkpoint.ContainerID)
	}

	if checkpoint.Name != "test-checkpoint" {
		t.Errorf("Expected name test-checkpoint, got %s", checkpoint.Name)
	}

	if checkpoint.State != CheckpointStateReady {
		t.Errorf("Expected state %s, got %s", CheckpointStateReady, checkpoint.State)
	}

	// Verify CRIU dump was called
	if !mockCRIU.dumpCalled {
		t.Error("Expected CRIU dump to be called")
	}

	// Verify checkpoint files exist
	if _, err := os.Stat(checkpoint.ImagePath); os.IsNotExist(err) {
		t.Errorf("Checkpoint image path does not exist: %s", checkpoint.ImagePath)
	}

	// Verify checkpoint size was calculated
	if checkpoint.Size == 0 {
		t.Error("Expected checkpoint size to be calculated")
	}
}

func TestManager_Checkpoint_DefaultName(t *testing.T) {
	mgr, _, _ := setupTestManager(t)
	ctx := context.Background()

	containerID := "test-container-456"
	opts := &CheckpointOptions{
		LeaveRunning: false,
	}

	checkpoint, err := mgr.Checkpoint(ctx, containerID, opts)
	if err != nil {
		t.Fatalf("Checkpoint failed: %v", err)
	}

	// Verify default name was generated
	if checkpoint.Name == "" {
		t.Error("Expected default name to be generated")
	}

	if len(checkpoint.Name) < 10 {
		t.Errorf("Expected default name to be at least 10 characters, got %d", len(checkpoint.Name))
	}
}

func TestManager_Restore(t *testing.T) {
	mgr, mockCRIU, _ := setupTestManager(t)
	ctx := context.Background()

	// First create a checkpoint
	containerID := "test-container-789"
	checkpoint, err := mgr.Checkpoint(ctx, containerID, &CheckpointOptions{
		Name: "test-checkpoint-restore",
	})
	if err != nil {
		t.Fatalf("Failed to create checkpoint: %v", err)
	}

	// Reset mock
	mockCRIU.restoreCalled = false

	// Now restore it
	restoreOpts := &RestoreOptions{
		Name:   "restored-container",
		Detach: true,
	}

	err = mgr.Restore(ctx, checkpoint.ID, restoreOpts)
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Verify CRIU restore was called
	if !mockCRIU.restoreCalled {
		t.Error("Expected CRIU restore to be called")
	}
}

func TestManager_Restore_NonExistent(t *testing.T) {
	mgr, _, _ := setupTestManager(t)
	ctx := context.Background()

	// Try to restore non-existent checkpoint
	err := mgr.Restore(ctx, "non-existent-checkpoint", &RestoreOptions{})
	if err == nil {
		t.Error("Expected error when restoring non-existent checkpoint")
	}
}

func TestManager_List(t *testing.T) {
	mgr, _, _ := setupTestManager(t)
	ctx := context.Background()

	// Create multiple checkpoints
	checkpoints := []string{"container-1", "container-2", "container-3"}
	for _, containerID := range checkpoints {
		_, err := mgr.Checkpoint(ctx, containerID, &CheckpointOptions{
			Name: "checkpoint-" + containerID,
		})
		if err != nil {
			t.Fatalf("Failed to create checkpoint: %v", err)
		}
	}

	// List checkpoints
	list, err := mgr.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != len(checkpoints) {
		t.Errorf("Expected %d checkpoints, got %d", len(checkpoints), len(list))
	}
}

func TestManager_Get(t *testing.T) {
	mgr, _, _ := setupTestManager(t)
	ctx := context.Background()

	// Create a checkpoint
	containerID := "test-container-get"
	checkpoint, err := mgr.Checkpoint(ctx, containerID, &CheckpointOptions{
		Name: "test-get",
	})
	if err != nil {
		t.Fatalf("Failed to create checkpoint: %v", err)
	}

	// Get the checkpoint
	retrieved, err := mgr.Get(ctx, checkpoint.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.ID != checkpoint.ID {
		t.Errorf("Expected checkpoint ID %s, got %s", checkpoint.ID, retrieved.ID)
	}

	if retrieved.Name != checkpoint.Name {
		t.Errorf("Expected checkpoint name %s, got %s", checkpoint.Name, retrieved.Name)
	}
}

func TestManager_Delete(t *testing.T) {
	mgr, _, _ := setupTestManager(t)
	ctx := context.Background()

	// Create a checkpoint
	containerID := "test-container-delete"
	checkpoint, err := mgr.Checkpoint(ctx, containerID, &CheckpointOptions{
		Name: "test-delete",
	})
	if err != nil {
		t.Fatalf("Failed to create checkpoint: %v", err)
	}

	// Verify checkpoint exists
	if _, err := mgr.Get(ctx, checkpoint.ID); err != nil {
		t.Fatalf("Checkpoint should exist before delete: %v", err)
	}

	// Delete the checkpoint
	err = mgr.Delete(ctx, checkpoint.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify checkpoint is gone
	_, err = mgr.Get(ctx, checkpoint.ID)
	if err == nil {
		t.Error("Expected error when getting deleted checkpoint")
	}

	// Verify checkpoint directory is removed
	checkpointDir := filepath.Dir(checkpoint.ImagePath)
	if _, err := os.Stat(checkpointDir); !os.IsNotExist(err) {
		t.Error("Expected checkpoint directory to be removed")
	}
}

func TestManager_LoadCheckpoints(t *testing.T) {
	mgr, _, tmpDir := setupTestManager(t)
	ctx := context.Background()

	// Create checkpoints
	containerID := "test-container-load"
	checkpoint1, _ := mgr.Checkpoint(ctx, containerID, &CheckpointOptions{Name: "cp1"})
	checkpoint2, _ := mgr.Checkpoint(ctx, containerID, &CheckpointOptions{Name: "cp2"})

	// Create a new manager (simulating restart)
	storePath := filepath.Join(tmpDir, "checkpoints")
	statePath := filepath.Join(tmpDir, "state")

	stateStore, err := NewStateStore(statePath)
	if err != nil {
		t.Fatalf("Failed to create state store: %v", err)
	}

	newMgr := &Manager{
		criuManager: newMockCRIUManager(),
		stateStore:  stateStore,
		checkpoints: make(map[string]*Checkpoint),
		storePath:   storePath,
		criuPath:    "/usr/sbin/criu",
	}

	// Load checkpoints
	err = newMgr.loadCheckpoints()
	if err != nil {
		t.Fatalf("Failed to load checkpoints: %v", err)
	}

	// Verify checkpoints were loaded
	if len(newMgr.checkpoints) != 2 {
		t.Errorf("Expected 2 checkpoints to be loaded, got %d", len(newMgr.checkpoints))
	}

	if _, ok := newMgr.checkpoints[checkpoint1.ID]; !ok {
		t.Error("Expected checkpoint1 to be loaded")
	}

	if _, ok := newMgr.checkpoints[checkpoint2.ID]; !ok {
		t.Error("Expected checkpoint2 to be loaded")
	}
}

func TestCheckpointOptions_TCPEstablished(t *testing.T) {
	mgr, mockCRIU, _ := setupTestManager(t)
	ctx := context.Background()

	containerID := "test-container-tcp"
	opts := &CheckpointOptions{
		Name:           "test-tcp",
		TCPEstablished: true,
	}

	_, err := mgr.Checkpoint(ctx, containerID, opts)
	if err != nil {
		t.Fatalf("Checkpoint failed: %v", err)
	}

	if !mockCRIU.dumpCalled {
		t.Error("Expected CRIU dump to be called")
	}
}

func TestStateStore_Save(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStateStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create state store: %v", err)
	}

	state := &ContainerCheckpointState{
		ID:             "test-state-1",
		ContainerID:    "container-1",
		CheckpointName: "checkpoint-1",
		Created:        time.Now(),
		Status:         "ready",
		ImagePath:      "/tmp/test",
		Size:           1024,
		Metadata:       make(map[string]interface{}),
	}

	err = store.Save(state)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify state file exists
	statePath := filepath.Join(tmpDir, state.ID, "state.json")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("Expected state file to exist")
	}
}

func TestStateStore_Load(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStateStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create state store: %v", err)
	}

	// Save a state
	originalState := &ContainerCheckpointState{
		ID:             "test-state-2",
		ContainerID:    "container-2",
		CheckpointName: "checkpoint-2",
		Created:        time.Now(),
		Status:         "ready",
		ImagePath:      "/tmp/test2",
		Size:           2048,
		Metadata:       make(map[string]interface{}),
	}

	store.Save(originalState)

	// Load the state
	loadedState, err := store.Load(originalState.ID)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loadedState.ID != originalState.ID {
		t.Errorf("Expected ID %s, got %s", originalState.ID, loadedState.ID)
	}

	if loadedState.ContainerID != originalState.ContainerID {
		t.Errorf("Expected container ID %s, got %s", originalState.ContainerID, loadedState.ContainerID)
	}

	if loadedState.Size != originalState.Size {
		t.Errorf("Expected size %d, got %d", originalState.Size, loadedState.Size)
	}
}

func TestStateStore_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStateStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create state store: %v", err)
	}

	// Save a state
	state := &ContainerCheckpointState{
		ID:             "test-state-3",
		ContainerID:    "container-3",
		CheckpointName: "checkpoint-3",
		Created:        time.Now(),
		Status:         "ready",
		ImagePath:      "/tmp/test3",
		Metadata:       make(map[string]interface{}),
	}

	store.Save(state)

	// Delete the state
	err = store.Delete(state.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify state is gone
	_, err = store.Load(state.ID)
	if err == nil {
		t.Error("Expected error when loading deleted state")
	}

	// Verify state directory is removed
	stateDir := filepath.Join(tmpDir, state.ID)
	if _, err := os.Stat(stateDir); !os.IsNotExist(err) {
		t.Error("Expected state directory to be removed")
	}
}

func TestStateStore_ListByContainer(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStateStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create state store: %v", err)
	}

	containerID := "container-multi"

	// Create multiple states for same container
	for i := 0; i < 3; i++ {
		state := &ContainerCheckpointState{
			ID:             generateCheckpointID(containerID),
			ContainerID:    containerID,
			CheckpointName: "checkpoint",
			Created:        time.Now(),
			Status:         "ready",
			Metadata:       make(map[string]interface{}),
		}
		store.Save(state)
	}

	// Create state for different container
	otherState := &ContainerCheckpointState{
		ID:             "other-state",
		ContainerID:    "other-container",
		CheckpointName: "checkpoint",
		Created:        time.Now(),
		Status:         "ready",
		Metadata:       make(map[string]interface{}),
	}
	store.Save(otherState)

	// List by container
	states, err := store.ListByContainer(containerID)
	if err != nil {
		t.Fatalf("ListByContainer failed: %v", err)
	}

	if len(states) != 3 {
		t.Errorf("Expected 3 states, got %d", len(states))
	}

	for _, state := range states {
		if state.ContainerID != containerID {
			t.Errorf("Expected container ID %s, got %s", containerID, state.ContainerID)
		}
	}
}

func TestStateStore_AddMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStateStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create state store: %v", err)
	}

	state := &ContainerCheckpointState{
		ID:             "test-metadata",
		ContainerID:    "container-meta",
		CheckpointName: "checkpoint",
		Created:        time.Now(),
		Status:         "ready",
		Metadata:       make(map[string]interface{}),
	}
	store.Save(state)

	// Add metadata
	err = store.AddMetadata(state.ID, "key1", "value1")
	if err != nil {
		t.Fatalf("AddMetadata failed: %v", err)
	}

	err = store.AddMetadata(state.ID, "key2", 123)
	if err != nil {
		t.Fatalf("AddMetadata failed: %v", err)
	}

	// Load and verify
	loaded, _ := store.Load(state.ID)

	if loaded.Metadata["key1"] != "value1" {
		t.Errorf("Expected key1=value1, got %v", loaded.Metadata["key1"])
	}

	if loaded.Metadata["key2"] != float64(123) { // JSON unmarshals numbers as float64
		t.Errorf("Expected key2=123, got %v", loaded.Metadata["key2"])
	}
}

func TestStateStore_Tags(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStateStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create state store: %v", err)
	}

	state := &ContainerCheckpointState{
		ID:             "test-tags",
		ContainerID:    "container-tags",
		CheckpointName: "checkpoint",
		Created:        time.Now(),
		Status:         "ready",
		Metadata:       make(map[string]interface{}),
	}
	store.Save(state)

	// Add tags
	store.AddTag(state.ID, "production")
	store.AddTag(state.ID, "v1.0")

	// Load and verify
	loaded, _ := store.Load(state.ID)

	if len(loaded.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(loaded.Tags))
	}

	// Remove tag
	store.RemoveTag(state.ID, "v1.0")

	loaded, _ = store.Load(state.ID)

	if len(loaded.Tags) != 1 {
		t.Errorf("Expected 1 tag after removal, got %d", len(loaded.Tags))
	}

	if loaded.Tags[0] != "production" {
		t.Errorf("Expected tag 'production', got %s", loaded.Tags[0])
	}
}

func TestStateStore_FindByTag(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStateStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create state store: %v", err)
	}

	// Create states with tags
	for i := 0; i < 3; i++ {
		state := &ContainerCheckpointState{
			ID:             generateCheckpointID("container"),
			ContainerID:    "container",
			CheckpointName: "checkpoint",
			Created:        time.Now(),
			Status:         "ready",
			Metadata:       make(map[string]interface{}),
		}
		store.Save(state)
		store.AddTag(state.ID, "production")
	}

	// Create state with different tag
	otherState := &ContainerCheckpointState{
		ID:             "other-tag-state",
		ContainerID:    "container",
		CheckpointName: "checkpoint",
		Created:        time.Now(),
		Status:         "ready",
		Metadata:       make(map[string]interface{}),
	}
	store.Save(otherState)
	store.AddTag(otherState.ID, "staging")

	// Find by tag
	states, err := store.FindByTag("production")
	if err != nil {
		t.Fatalf("FindByTag failed: %v", err)
	}

	if len(states) != 3 {
		t.Errorf("Expected 3 states with 'production' tag, got %d", len(states))
	}
}

func TestStateStore_GetStatistics(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStateStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create state store: %v", err)
	}

	// Create states with different statuses
	states := []struct {
		status string
		size   int64
	}{
		{"ready", 1024},
		{"ready", 2048},
		{"creating", 512},
		{"failed", 100},
	}

	for i, s := range states {
		state := &ContainerCheckpointState{
			ID:             generateCheckpointID("container"),
			ContainerID:    "container",
			CheckpointName: "checkpoint",
			Created:        time.Now(),
			Status:         s.status,
			Size:           s.size,
			Metadata:       make(map[string]interface{}),
		}
		store.Save(state)
		// Sleep to ensure unique IDs
		time.Sleep(1 * time.Millisecond)
		_ = i
	}

	// Get statistics
	stats, err := store.GetStatistics()
	if err != nil {
		t.Fatalf("GetStatistics failed: %v", err)
	}

	if stats.Total != 4 {
		t.Errorf("Expected total 4, got %d", stats.Total)
	}

	if stats.ByStatus["ready"] != 2 {
		t.Errorf("Expected 2 ready states, got %d", stats.ByStatus["ready"])
	}

	expectedSize := int64(1024 + 2048 + 512 + 100)
	if stats.TotalSize != expectedSize {
		t.Errorf("Expected total size %d, got %d", expectedSize, stats.TotalSize)
	}
}

func TestStateStore_ExportImport(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStateStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create state store: %v", err)
	}

	// Create a state
	originalState := &ContainerCheckpointState{
		ID:             "export-test",
		ContainerID:    "container-export",
		CheckpointName: "checkpoint",
		Created:        time.Now(),
		Status:         "ready",
		Size:           4096,
		Metadata: map[string]interface{}{
			"key": "value",
		},
		Tags: []string{"tag1", "tag2"},
	}
	store.Save(originalState)

	// Export
	data, err := store.Export(originalState.ID)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected exported data to be non-empty")
	}

	// Create new store
	tmpDir2 := t.TempDir()
	store2, err := NewStateStore(tmpDir2)
	if err != nil {
		t.Fatalf("Failed to create second state store: %v", err)
	}

	// Import
	importedState, err := store2.Import(data)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if importedState.ID != originalState.ID {
		t.Errorf("Expected ID %s, got %s", originalState.ID, importedState.ID)
	}

	if importedState.ContainerID != originalState.ContainerID {
		t.Errorf("Expected container ID %s, got %s", originalState.ContainerID, importedState.ContainerID)
	}

	if len(importedState.Tags) != len(originalState.Tags) {
		t.Errorf("Expected %d tags, got %d", len(originalState.Tags), len(importedState.Tags))
	}
}

func TestGenerateCheckpointID(t *testing.T) {
	containerID := "test-container-123"

	id1 := generateCheckpointID(containerID)
	time.Sleep(1 * time.Millisecond)
	id2 := generateCheckpointID(containerID)

	// IDs should be different
	if id1 == id2 {
		t.Error("Expected checkpoint IDs to be unique")
	}

	// IDs should contain container prefix
	if len(id1) < 10 {
		t.Errorf("Expected checkpoint ID to be at least 10 characters, got %d", len(id1))
	}
}

func TestCalculateDirSize(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some test files
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), make([]byte, 1024), 0600)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), make([]byte, 2048), 0600)

	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0700)
	os.WriteFile(filepath.Join(subDir, "file3.txt"), make([]byte, 512), 0600)

	size, err := calculateDirSize(tmpDir)
	if err != nil {
		t.Fatalf("calculateDirSize failed: %v", err)
	}

	expectedSize := int64(1024 + 2048 + 512)
	if size != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, size)
	}
}
