package snapshot

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOverlay2Snapshotter_Prepare(t *testing.T) {
	tmpDir := t.TempDir()
	snap, err := NewOverlay2(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create snapshotter: %v", err)
	}
	defer snap.Close()

	ctx := context.Background()

	// Create a snapshot without parent
	mounts, err := snap.Prepare(ctx, "test-snap", "")
	if err != nil {
		t.Fatalf("Failed to prepare snapshot: %v", err)
	}

	if len(mounts) == 0 {
		t.Fatal("Expected at least one mount")
	}

	// Verify snapshot info
	info, err := snap.Stat(ctx, "test-snap")
	if err != nil {
		t.Fatalf("Failed to get snapshot info: %v", err)
	}

	if info.Name != "test-snap" {
		t.Errorf("Expected name 'test-snap', got '%s'", info.Name)
	}

	if info.Kind != KindActive {
		t.Errorf("Expected KindActive, got %v", info.Kind)
	}

	// Test duplicate preparation
	_, err = snap.Prepare(ctx, "test-snap", "")
	if err == nil {
		t.Fatal("Expected error when preparing duplicate snapshot")
	}
}

func TestOverlay2Snapshotter_PrepareWithLabels(t *testing.T) {
	tmpDir := t.TempDir()
	snap, err := NewOverlay2(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create snapshotter: %v", err)
	}
	defer snap.Close()

	ctx := context.Background()

	labels := map[string]string{
		"app":     "test",
		"version": "1.0",
	}

	_, err = snap.Prepare(ctx, "test-snap", "", WithLabels(labels))
	if err != nil {
		t.Fatalf("Failed to prepare snapshot: %v", err)
	}

	info, err := snap.Stat(ctx, "test-snap")
	if err != nil {
		t.Fatalf("Failed to get snapshot info: %v", err)
	}

	if len(info.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(info.Labels))
	}

	if info.Labels["app"] != "test" {
		t.Errorf("Expected label app='test', got '%s'", info.Labels["app"])
	}
}

func TestOverlay2Snapshotter_Commit(t *testing.T) {
	tmpDir := t.TempDir()
	snap, err := NewOverlay2(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create snapshotter: %v", err)
	}
	defer snap.Close()

	ctx := context.Background()

	// Prepare an active snapshot
	mounts, err := snap.Prepare(ctx, "active-snap", "")
	if err != nil {
		t.Fatalf("Failed to prepare snapshot: %v", err)
	}

	// Write some data to the snapshot
	if len(mounts) > 0 {
		testFile := filepath.Join(tmpDir, "snapshots", "active-snap", "upper", "test.txt")
		os.MkdirAll(filepath.Dir(testFile), 0755)
		os.WriteFile(testFile, []byte("test data"), 0644)
	}

	// Commit the snapshot
	err = snap.Commit(ctx, "committed-snap", "active-snap")
	if err != nil {
		t.Fatalf("Failed to commit snapshot: %v", err)
	}

	// Verify committed snapshot
	info, err := snap.Stat(ctx, "committed-snap")
	if err != nil {
		t.Fatalf("Failed to get committed snapshot info: %v", err)
	}

	if info.Kind != KindCommitted {
		t.Errorf("Expected KindCommitted, got %v", info.Kind)
	}

	// Verify data was copied
	dataFile := filepath.Join(tmpDir, "snapshots", "committed-snap", "data", "test.txt")
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		t.Error("Expected committed data file to exist")
	}
}

func TestOverlay2Snapshotter_View(t *testing.T) {
	tmpDir := t.TempDir()
	snap, err := NewOverlay2(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create snapshotter: %v", err)
	}
	defer snap.Close()

	ctx := context.Background()

	// Create and commit a snapshot
	snap.Prepare(ctx, "active", "")
	snap.Commit(ctx, "committed", "active")

	// Create a read-only view
	_, err = snap.View(ctx, "view-snap", "committed")
	if err != nil {
		t.Fatalf("Failed to create view: %v", err)
	}

	info, err := snap.Stat(ctx, "view-snap")
	if err != nil {
		t.Fatalf("Failed to get view info: %v", err)
	}

	if info.Kind != KindView {
		t.Errorf("Expected KindView, got %v", info.Kind)
	}

	if info.Parent != "committed" {
		t.Errorf("Expected parent 'committed', got '%s'", info.Parent)
	}
}

func TestOverlay2Snapshotter_Remove(t *testing.T) {
	tmpDir := t.TempDir()
	snap, err := NewOverlay2(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create snapshotter: %v", err)
	}
	defer snap.Close()

	ctx := context.Background()

	// Create a snapshot
	snap.Prepare(ctx, "test-snap", "")

	// Remove it
	err = snap.Remove(ctx, "test-snap")
	if err != nil {
		t.Fatalf("Failed to remove snapshot: %v", err)
	}

	// Verify it's gone
	_, err = snap.Stat(ctx, "test-snap")
	if err == nil {
		t.Fatal("Expected error when getting removed snapshot")
	}

	// Verify directory is removed
	snapshotDir := filepath.Join(tmpDir, "snapshots", "test-snap")
	if _, err := os.Stat(snapshotDir); !os.IsNotExist(err) {
		t.Error("Expected snapshot directory to be removed")
	}
}

func TestOverlay2Snapshotter_Walk(t *testing.T) {
	tmpDir := t.TempDir()
	snap, err := NewOverlay2(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create snapshotter: %v", err)
	}
	defer snap.Close()

	ctx := context.Background()

	// Create multiple snapshots
	snap.Prepare(ctx, "snap1", "")
	snap.Prepare(ctx, "snap2", "")
	snap.Prepare(ctx, "snap3", "")

	// Walk through snapshots
	count := 0
	err = snap.Walk(ctx, func(ctx context.Context, info Info) error {
		count++
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk snapshots: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected to walk 3 snapshots, got %d", count)
	}
}

func TestOverlay2Snapshotter_Update(t *testing.T) {
	tmpDir := t.TempDir()
	snap, err := NewOverlay2(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create snapshotter: %v", err)
	}
	defer snap.Close()

	ctx := context.Background()

	// Create a snapshot
	snap.Prepare(ctx, "test-snap", "")

	// Update labels
	newLabels := map[string]string{"updated": "true"}
	updated, err := snap.Update(ctx, Info{Name: "test-snap", Labels: newLabels}, "labels")
	if err != nil {
		t.Fatalf("Failed to update snapshot: %v", err)
	}

	if updated.Labels["updated"] != "true" {
		t.Error("Expected label 'updated' to be set")
	}

	// Verify update was persisted
	info, err := snap.Stat(ctx, "test-snap")
	if err != nil {
		t.Fatalf("Failed to get snapshot info: %v", err)
	}

	if info.Labels["updated"] != "true" {
		t.Error("Expected updated label to be persisted")
	}
}

func TestOverlay2Snapshotter_Usage(t *testing.T) {
	tmpDir := t.TempDir()
	snap, err := NewOverlay2(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create snapshotter: %v", err)
	}
	defer snap.Close()

	ctx := context.Background()

	// Create a snapshot and add some data
	snap.Prepare(ctx, "test-snap", "")

	testFile := filepath.Join(tmpDir, "snapshots", "test-snap", "upper", "test.txt")
	os.MkdirAll(filepath.Dir(testFile), 0755)
	testData := []byte("test data with some content")
	os.WriteFile(testFile, testData, 0644)

	// Get usage
	usage, err := snap.Usage(ctx, "test-snap")
	if err != nil {
		t.Fatalf("Failed to get usage: %v", err)
	}

	if usage.Size == 0 {
		t.Error("Expected non-zero usage size")
	}
}

func TestOverlay2Snapshotter_ParentChain(t *testing.T) {
	tmpDir := t.TempDir()
	snap, err := NewOverlay2(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create snapshotter: %v", err)
	}
	defer snap.Close()

	ctx := context.Background()

	// Create a chain: base -> layer1 -> layer2
	snap.Prepare(ctx, "base", "")
	snap.Commit(ctx, "base-committed", "base")

	snap.Prepare(ctx, "layer1", "base-committed")
	snap.Commit(ctx, "layer1-committed", "layer1")

	snap.Prepare(ctx, "layer2", "layer1-committed")

	// Verify parent relationships
	info, err := snap.Stat(ctx, "layer2")
	if err != nil {
		t.Fatalf("Failed to get layer2 info: %v", err)
	}

	if info.Parent != "layer1-committed" {
		t.Errorf("Expected parent 'layer1-committed', got '%s'", info.Parent)
	}

	// Verify mounts include all layers
	mounts, err := snap.Mounts(ctx, "layer2")
	if err != nil {
		t.Fatalf("Failed to get mounts: %v", err)
	}

	if len(mounts) == 0 {
		t.Fatal("Expected at least one mount")
	}

	// Check that overlay options include lower directories
	hasLowerDir := false
	for _, opt := range mounts[0].Options {
		if contains(opt, "lowerdir") {
			hasLowerDir = true
			break
		}
	}

	if !hasLowerDir {
		t.Error("Expected overlay mount to have lowerdir option for parent chain")
	}
}

func TestMetadata_SaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewMetadata(tmpDir)

	info := Info{
		Name:      "test-snapshot",
		Parent:    "parent-snapshot",
		Kind:      KindActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Labels: map[string]string{
			"key": "value",
		},
		Size: 1024,
	}

	// Save metadata
	err := m.Save(info)
	if err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Load metadata
	loaded, err := m.Load("test-snapshot")
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	if loaded.Name != info.Name {
		t.Errorf("Expected name '%s', got '%s'", info.Name, loaded.Name)
	}

	if loaded.Parent != info.Parent {
		t.Errorf("Expected parent '%s', got '%s'", info.Parent, loaded.Parent)
	}

	if loaded.Kind != info.Kind {
		t.Errorf("Expected kind %v, got %v", info.Kind, loaded.Kind)
	}

	if loaded.Labels["key"] != "value" {
		t.Error("Expected label to be preserved")
	}
}

func TestMetadata_List(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewMetadata(tmpDir)

	// Save multiple metadata entries
	for i := 0; i < 3; i++ {
		info := Info{
			Name:      string(rune('A' + i)),
			Kind:      KindActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		m.Save(info)
	}

	// List all
	infos, err := m.List()
	if err != nil {
		t.Fatalf("Failed to list metadata: %v", err)
	}

	if len(infos) != 3 {
		t.Errorf("Expected 3 metadata entries, got %d", len(infos))
	}
}

func TestMetadata_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewMetadata(tmpDir)

	info := Info{
		Name:      "test",
		Kind:      KindActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	m.Save(info)

	// Delete
	err := m.Delete("test")
	if err != nil {
		t.Fatalf("Failed to delete metadata: %v", err)
	}

	// Verify it's gone
	if m.Exists("test") {
		t.Error("Expected metadata to be deleted")
	}

	// Load should fail
	_, err = m.Load("test")
	if err == nil {
		t.Fatal("Expected error when loading deleted metadata")
	}
}

func TestMetadata_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewMetadata(tmpDir)

	// Should not exist initially
	if m.Exists("test") {
		t.Error("Expected metadata not to exist")
	}

	// Save metadata
	info := Info{
		Name:      "test",
		Kind:      KindActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.Save(info)

	// Should exist now
	if !m.Exists("test") {
		t.Error("Expected metadata to exist")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
	       len(s) > len(substr) && containsHelper(s[1:], substr)
}

func containsHelper(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	if s[:len(substr)] == substr {
		return true
	}
	return containsHelper(s[1:], substr)
}
