package userns

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestIDMapString(t *testing.T) {
	mapping := IDMap{
		ContainerID: 0,
		HostID:      1000,
		Size:        1,
	}

	expected := "0 1000 1"
	result := formatIDMap(mapping)

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func formatIDMap(mapping IDMap) string {
	return fmt.Sprintf("%d %d %d", mapping.ContainerID, mapping.HostID, mapping.Size)
}

func TestParseSubIDFile(t *testing.T) {
	// Create a temporary subuid file
	tmpDir := t.TempDir()
	subuidPath := filepath.Join(tmpDir, "subuid")

	content := `testuser:100000:65536
otheruser:200000:65536
# comment line
`

	if err := os.WriteFile(subuidPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	mappings, err := parseSubIDFile(subuidPath, "testuser")
	if err != nil {
		t.Fatalf("Failed to parse subuid file: %v", err)
	}

	if len(mappings) != 1 {
		t.Fatalf("Expected 1 mapping, got %d", len(mappings))
	}

	if mappings[0].HostID != 100000 {
		t.Errorf("Expected HostID 100000, got %d", mappings[0].HostID)
	}

	if mappings[0].Size != 65536 {
		t.Errorf("Expected Size 65536, got %d", mappings[0].Size)
	}

	// Test non-existent user
	_, err = parseSubIDFile(subuidPath, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				UIDMappings: []IDMap{
					{ContainerID: 0, HostID: 1000, Size: 1},
				},
				GIDMappings: []IDMap{
					{ContainerID: 0, HostID: 1000, Size: 1},
				},
			},
			expectErr: false,
		},
		{
			name: "no UID mappings",
			config: &Config{
				GIDMappings: []IDMap{
					{ContainerID: 0, HostID: 1000, Size: 1},
				},
			},
			expectErr: true,
		},
		{
			name: "no GID mappings",
			config: &Config{
				UIDMappings: []IDMap{
					{ContainerID: 0, HostID: 1000, Size: 1},
				},
			},
			expectErr: true,
		},
		{
			name: "invalid size",
			config: &Config{
				UIDMappings: []IDMap{
					{ContainerID: 0, HostID: 1000, Size: 0},
				},
				GIDMappings: []IDMap{
					{ContainerID: 0, HostID: 1000, Size: 1},
				},
			},
			expectErr: true,
		},
		{
			name: "negative ID",
			config: &Config{
				UIDMappings: []IDMap{
					{ContainerID: -1, HostID: 1000, Size: 1},
				},
				GIDMappings: []IDMap{
					{ContainerID: 0, HostID: 1000, Size: 1},
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestIsRootless(t *testing.T) {
	// This test will pass different results depending on who runs it
	isRootless := IsRootless()
	uid := os.Getuid()

	if uid == 0 && isRootless {
		t.Error("Running as root but IsRootless returned true")
	}

	if uid != 0 && !isRootless {
		t.Error("Running as non-root but IsRootless returned false")
	}
}

func TestSupportsUserNamespaces(t *testing.T) {
	// This is a simple check - the result will depend on the system
	supports := SupportsUserNamespaces()

	// Just ensure it returns a boolean without panicking
	t.Logf("User namespaces supported: %v", supports)
}
