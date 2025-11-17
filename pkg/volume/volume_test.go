package volume

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()

	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager.root != tmpDir {
		t.Errorf("Expected root %s, got %s", tmpDir, manager.root)
	}

	// Check if directory was created
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Manager directory was not created")
	}
}

func TestManagerCreate(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	volume, err := manager.Create("test-volume", nil)
	if err != nil {
		t.Fatalf("Failed to create volume: %v", err)
	}

	if volume.Name != "test-volume" {
		t.Errorf("Expected name 'test-volume', got %s", volume.Name)
	}

	if volume.Type != TypeVolume {
		t.Errorf("Expected type %s, got %s", TypeVolume, volume.Type)
	}

	// Check if volume directory was created
	volumePath := filepath.Join(tmpDir, "test-volume")
	if _, err := os.Stat(volumePath); os.IsNotExist(err) {
		t.Error("Volume directory was not created")
	}

	// Try to create duplicate volume
	_, err = manager.Create("test-volume", nil)
	if err == nil {
		t.Error("Expected error when creating duplicate volume")
	}
}

func TestManagerRemove(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create volume
	_, err = manager.Create("test-volume", nil)
	if err != nil {
		t.Fatalf("Failed to create volume: %v", err)
	}

	// Remove volume
	if err := manager.Remove("test-volume"); err != nil {
		t.Fatalf("Failed to remove volume: %v", err)
	}

	// Check if volume directory was removed
	volumePath := filepath.Join(tmpDir, "test-volume")
	if _, err := os.Stat(volumePath); !os.IsNotExist(err) {
		t.Error("Volume directory was not removed")
	}

	// Try to remove non-existent volume
	if err := manager.Remove("non-existent"); err == nil {
		t.Error("Expected error when removing non-existent volume")
	}
}

func TestManagerGet(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create volume
	created, err := manager.Create("test-volume", nil)
	if err != nil {
		t.Fatalf("Failed to create volume: %v", err)
	}

	// Get volume
	got, err := manager.Get("test-volume")
	if err != nil {
		t.Fatalf("Failed to get volume: %v", err)
	}

	if got.Name != created.Name {
		t.Errorf("Expected name %s, got %s", created.Name, got.Name)
	}

	// Try to get non-existent volume
	_, err = manager.Get("non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent volume")
	}
}

func TestManagerList(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Create multiple volumes
	for i := 0; i < 3; i++ {
		_, err := manager.Create(fmt.Sprintf("volume-%d", i), nil)
		if err != nil {
			t.Fatalf("Failed to create volume: %v", err)
		}
	}

	// List volumes
	volumes, err := manager.List()
	if err != nil {
		t.Fatalf("Failed to list volumes: %v", err)
	}

	if len(volumes) != 3 {
		t.Errorf("Expected 3 volumes, got %d", len(volumes))
	}
}

func TestManagerExists(t *testing.T) {
	tmpDir := t.TempDir()
	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Check non-existent volume
	if manager.Exists("test-volume") {
		t.Error("Volume should not exist")
	}

	// Create volume
	_, err = manager.Create("test-volume", nil)
	if err != nil {
		t.Fatalf("Failed to create volume: %v", err)
	}

	// Check existing volume
	if !manager.Exists("test-volume") {
		t.Error("Volume should exist")
	}
}

func TestParseVolumeString(t *testing.T) {
	tests := []struct {
		name        string
		volumeStr   string
		expectErr   bool
		expectType  VolumeType
		expectRO    bool
		expectDest  string
	}{
		{
			name:       "absolute path bind mount",
			volumeStr:  "/host/path:/container/path",
			expectErr:  false,
			expectType: TypeBind,
			expectRO:   false,
			expectDest: "/container/path",
		},
		{
			name:       "absolute path bind mount read-only",
			volumeStr:  "/host/path:/container/path:ro",
			expectErr:  false,
			expectType: TypeBind,
			expectRO:   true,
			expectDest: "/container/path",
		},
		{
			name:      "invalid format",
			volumeStr: "/host/path",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volume, err := ParseVolumeString(tt.volumeStr, nil)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if volume.Type != tt.expectType {
				t.Errorf("Expected type %s, got %s", tt.expectType, volume.Type)
			}

			if volume.ReadOnly != tt.expectRO {
				t.Errorf("Expected ReadOnly %v, got %v", tt.expectRO, volume.ReadOnly)
			}

			if volume.Destination != tt.expectDest {
				t.Errorf("Expected destination %s, got %s", tt.expectDest, volume.Destination)
			}
		})
	}
}

func TestSplitVolumeString(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "/host/path:/container/path",
			expected: []string{"/host/path", "/container/path"},
		},
		{
			input:    "/host/path:/container/path:ro",
			expected: []string{"/host/path", "/container/path", "ro"},
		},
		{
			input:    "name:/path",
			expected: []string{"name", "/path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := splitVolumeString(tt.input)

			if len(result) != len(tt.expected) {
				t.Fatalf("Expected %d parts, got %d", len(tt.expected), len(result))
			}

			for i, part := range result {
				if part != tt.expected[i] {
					t.Errorf("Part %d: expected %s, got %s", i, tt.expected[i], part)
				}
			}
		})
	}
}
