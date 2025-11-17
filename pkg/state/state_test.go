package state

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	tmpDir := t.TempDir()

	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	if store.root != tmpDir {
		t.Errorf("Expected root %s, got %s", tmpDir, store.root)
	}

	// Check if directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Store directory was not created")
	}
}

func TestStoreSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	container := &Container{
		ID:       "test-123",
		Name:     "test-container",
		State:    StateRunning,
		PID:      1234,
		Created:  time.Now(),
		Command:  []string{"/bin/sh"},
		Hostname: "test",
	}

	// Save container
	if err := store.Save(container); err != nil {
		t.Fatalf("Failed to save container: %v", err)
	}

	// Load container
	loaded, err := store.Load("test-123")
	if err != nil {
		t.Fatalf("Failed to load container: %v", err)
	}

	if loaded.ID != container.ID {
		t.Errorf("Expected ID %s, got %s", container.ID, loaded.ID)
	}

	if loaded.Name != container.Name {
		t.Errorf("Expected name %s, got %s", container.Name, loaded.Name)
	}

	if loaded.State != container.State {
		t.Errorf("Expected state %s, got %s", container.State, loaded.State)
	}

	if loaded.PID != container.PID {
		t.Errorf("Expected PID %d, got %d", container.PID, loaded.PID)
	}
}

func TestStoreDelete(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	container := &Container{
		ID:      "test-123",
		State:   StateRunning,
		Command: []string{"/bin/sh"},
	}

	// Save container
	if err := store.Save(container); err != nil {
		t.Fatalf("Failed to save container: %v", err)
	}

	// Delete container
	if err := store.Delete("test-123"); err != nil {
		t.Fatalf("Failed to delete container: %v", err)
	}

	// Try to load deleted container
	_, err = store.Load("test-123")
	if err == nil {
		t.Error("Expected error when loading deleted container")
	}

	// Check if directory was removed
	path := filepath.Join(tmpDir, "test-123")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("Container directory was not removed")
	}
}

func TestStoreList(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Save multiple containers
	for i := 0; i < 3; i++ {
		container := &Container{
			ID:      fmt.Sprintf("test-%d", i),
			State:   StateRunning,
			Command: []string{"/bin/sh"},
		}
		if err := store.Save(container); err != nil {
			t.Fatalf("Failed to save container: %v", err)
		}
	}

	// List containers
	containers, err := store.List()
	if err != nil {
		t.Fatalf("Failed to list containers: %v", err)
	}

	if len(containers) != 3 {
		t.Errorf("Expected 3 containers, got %d", len(containers))
	}
}

func TestStoreListByState(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	// Save containers with different states
	states := []ContainerState{StateRunning, StateStopped, StateRunning}
	for i, state := range states {
		container := &Container{
			ID:      fmt.Sprintf("test-%d", i),
			State:   state,
			Command: []string{"/bin/sh"},
		}
		if err := store.Save(container); err != nil {
			t.Fatalf("Failed to save container: %v", err)
		}
	}

	// List running containers
	running, err := store.ListByState(StateRunning)
	if err != nil {
		t.Fatalf("Failed to list running containers: %v", err)
	}

	if len(running) != 2 {
		t.Errorf("Expected 2 running containers, got %d", len(running))
	}

	// List stopped containers
	stopped, err := store.ListByState(StateStopped)
	if err != nil {
		t.Fatalf("Failed to list stopped containers: %v", err)
	}

	if len(stopped) != 1 {
		t.Errorf("Expected 1 stopped container, got %d", len(stopped))
	}
}

func TestStoreExists(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	container := &Container{
		ID:      "test-123",
		State:   StateRunning,
		Command: []string{"/bin/sh"},
	}

	// Check non-existent container
	if store.Exists("test-123") {
		t.Error("Container should not exist")
	}

	// Save container
	if err := store.Save(container); err != nil {
		t.Fatalf("Failed to save container: %v", err)
	}

	// Check existing container
	if !store.Exists("test-123") {
		t.Error("Container should exist")
	}
}

func TestStoreFindByName(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStore(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	container := &Container{
		ID:      "test-123",
		Name:    "my-container",
		State:   StateRunning,
		Command: []string{"/bin/sh"},
	}

	// Save container
	if err := store.Save(container); err != nil {
		t.Fatalf("Failed to save container: %v", err)
	}

	// Find by name
	found, err := store.FindByName("my-container")
	if err != nil {
		t.Fatalf("Failed to find container by name: %v", err)
	}

	if found.ID != container.ID {
		t.Errorf("Expected ID %s, got %s", container.ID, found.ID)
	}

	// Try to find non-existent container
	_, err = store.FindByName("non-existent")
	if err == nil {
		t.Error("Expected error when finding non-existent container")
	}
}
