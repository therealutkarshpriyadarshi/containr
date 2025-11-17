package runtime

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultSpec(t *testing.T) {
	spec := DefaultSpec()

	if spec.Version == "" {
		t.Error("expected non-empty version")
	}

	if spec.Root == nil {
		t.Error("expected root to be set")
	}

	if spec.Process == nil {
		t.Error("expected process to be set")
	}

	if err := spec.Validate(); err != nil {
		t.Errorf("default spec validation failed: %v", err)
	}
}

func TestSpecSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "config.json")

	spec := DefaultSpec()
	spec.Hostname = "test-container"

	// Save spec
	if err := spec.Save(specPath); err != nil {
		t.Fatalf("failed to save spec: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		t.Error("spec file was not created")
	}

	// Load spec
	loadedSpec, err := LoadSpec(specPath)
	if err != nil {
		t.Fatalf("failed to load spec: %v", err)
	}

	if loadedSpec.Hostname != "test-container" {
		t.Errorf("expected hostname 'test-container', got '%s'", loadedSpec.Hostname)
	}

	if loadedSpec.Version != spec.Version {
		t.Errorf("expected version '%s', got '%s'", spec.Version, loadedSpec.Version)
	}
}

func TestSpecValidation(t *testing.T) {
	tests := []struct {
		name    string
		spec    *Spec
		wantErr bool
	}{
		{
			name:    "valid spec",
			spec:    DefaultSpec(),
			wantErr: false,
		},
		{
			name: "missing version",
			spec: &Spec{
				Root: &Root{Path: "rootfs"},
				Process: &Process{
					Args: []string{"/bin/sh"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing root",
			spec: &Spec{
				Version: "1.0.0",
			},
			wantErr: true,
		},
		{
			name: "missing process",
			spec: &Spec{
				Version: "1.0.0",
				Root:    &Root{Path: "rootfs"},
			},
			wantErr: true,
		},
		{
			name: "empty process args",
			spec: &Spec{
				Version: "1.0.0",
				Root:    &Root{Path: "rootfs"},
				Process: &Process{
					Args: []string{},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "state.json")

	state := &State{
		Version:   "1.0.0",
		ID:        "test-container",
		Status:    StatusRunning,
		Pid:       12345,
		Bundle:    "/path/to/bundle",
		CreatedAt: time.Now(),
		Annotations: map[string]string{
			"key": "value",
		},
	}

	// Save state
	if err := state.Save(statePath); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	// Load state
	loadedState, err := LoadState(statePath)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	if loadedState.ID != state.ID {
		t.Errorf("expected ID '%s', got '%s'", state.ID, loadedState.ID)
	}

	if loadedState.Status != state.Status {
		t.Errorf("expected status '%s', got '%s'", state.Status, loadedState.Status)
	}

	if loadedState.Pid != state.Pid {
		t.Errorf("expected PID %d, got %d", state.Pid, loadedState.Pid)
	}
}

func TestContainerStatus(t *testing.T) {
	statuses := []ContainerStatus{
		StatusCreating,
		StatusCreated,
		StatusRunning,
		StatusStopped,
	}

	for _, status := range statuses {
		if string(status) == "" {
			t.Errorf("status should not be empty")
		}
	}
}

func TestSpecWithResources(t *testing.T) {
	spec := DefaultSpec()

	memLimit := int64(100 * 1024 * 1024) // 100MB
	cpuShares := uint64(512)
	pidLimit := int64(100)

	spec.Linux.Resources = &Resources{
		Memory: &Memory{
			Limit: &memLimit,
		},
		CPU: &CPU{
			Shares: &cpuShares,
		},
		Pids: &Pids{
			Limit: pidLimit,
		},
	}

	if err := spec.Validate(); err != nil {
		t.Errorf("spec with resources validation failed: %v", err)
	}

	if *spec.Linux.Resources.Memory.Limit != memLimit {
		t.Errorf("expected memory limit %d, got %d", memLimit, *spec.Linux.Resources.Memory.Limit)
	}
}
