package seccomp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultProfile(t *testing.T) {
	profile := DefaultProfile()

	if profile == nil {
		t.Fatal("DefaultProfile returned nil")
	}

	if profile.DefaultAction != ActErrno {
		t.Errorf("DefaultProfile.DefaultAction = %v, want %v", profile.DefaultAction, ActErrno)
	}

	if len(profile.Syscalls) == 0 {
		t.Error("DefaultProfile should have syscall rules")
	}

	if len(profile.Architectures) == 0 {
		t.Error("DefaultProfile should have architectures")
	}

	// Check that dangerous syscalls are blocked
	foundBlocked := false
	for _, syscall := range profile.Syscalls {
		for _, name := range syscall.Names {
			if name == "reboot" || name == "mount" || name == "ptrace" {
				foundBlocked = true
				if syscall.Action != ActErrno {
					t.Errorf("Dangerous syscall %s should be blocked, got action %v", name, syscall.Action)
				}
			}
		}
	}

	if !foundBlocked {
		t.Error("DefaultProfile should block dangerous syscalls")
	}
}

func TestUnconfinedProfile(t *testing.T) {
	profile := UnconfinedProfile()

	if profile == nil {
		t.Fatal("UnconfinedProfile returned nil")
	}

	if profile.DefaultAction != ActAllow {
		t.Errorf("UnconfinedProfile.DefaultAction = %v, want %v", profile.DefaultAction, ActAllow)
	}
}

func TestProfileSaveAndLoad(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "test-profile.json")

	// Create a test profile
	original := &Profile{
		DefaultAction: ActErrno,
		Architectures: []string{"SCMP_ARCH_X86_64"},
		Syscalls: []*Syscall{
			{
				Names:  []string{"read", "write", "open"},
				Action: ActAllow,
			},
			{
				Names:  []string{"reboot", "mount"},
				Action: ActErrno,
			},
		},
	}

	// Save profile
	if err := original.Save(profilePath); err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		t.Fatal("Profile file was not created")
	}

	// Load profile
	loaded, err := LoadProfile(profilePath)
	if err != nil {
		t.Fatalf("Failed to load profile: %v", err)
	}

	// Verify loaded profile matches original
	if loaded.DefaultAction != original.DefaultAction {
		t.Errorf("DefaultAction mismatch: got %v, want %v", loaded.DefaultAction, original.DefaultAction)
	}

	if len(loaded.Architectures) != len(original.Architectures) {
		t.Errorf("Architectures count mismatch: got %d, want %d", len(loaded.Architectures), len(original.Architectures))
	}

	if len(loaded.Syscalls) != len(original.Syscalls) {
		t.Errorf("Syscalls count mismatch: got %d, want %d", len(loaded.Syscalls), len(original.Syscalls))
	}

	// Verify syscall rules
	for i, syscall := range original.Syscalls {
		if i >= len(loaded.Syscalls) {
			break
		}
		loadedSyscall := loaded.Syscalls[i]

		if syscall.Action != loadedSyscall.Action {
			t.Errorf("Syscall[%d] action mismatch: got %v, want %v", i, loadedSyscall.Action, syscall.Action)
		}

		if len(syscall.Names) != len(loadedSyscall.Names) {
			t.Errorf("Syscall[%d] names count mismatch: got %d, want %d", i, len(loadedSyscall.Names), len(syscall.Names))
		}
	}
}

func TestLoadProfileInvalidPath(t *testing.T) {
	_, err := LoadProfile("/nonexistent/path/profile.json")
	if err == nil {
		t.Error("LoadProfile should fail for nonexistent path")
	}
}

func TestLoadProfileInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	if err := os.WriteFile(profilePath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := LoadProfile(profilePath)
	if err == nil {
		t.Error("LoadProfile should fail for invalid JSON")
	}
}

func TestConfigApply(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "default configuration",
			config: &Config{
				Profile: DefaultProfile(),
			},
			wantErr: false,
		},
		{
			name: "unconfined profile",
			config: &Config{
				Profile: UnconfinedProfile(),
			},
			wantErr: false,
		},
		{
			name: "disabled seccomp",
			config: &Config{
				Disabled: true,
			},
			wantErr: false,
		},
		{
			name: "nil profile uses default",
			config: &Config{
				Profile: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Apply()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigApplyWithFile(t *testing.T) {
	tmpDir := t.TempDir()
	profilePath := filepath.Join(tmpDir, "test-profile.json")

	// Save a test profile
	profile := DefaultProfile()
	if err := profile.Save(profilePath); err != nil {
		t.Fatalf("Failed to save test profile: %v", err)
	}

	// Apply with profile path
	config := &Config{
		ProfilePath: profilePath,
	}

	err := config.Apply()
	if err != nil {
		t.Errorf("Config.Apply() with profile path failed: %v", err)
	}
}

func TestConfigApplyInvalidPath(t *testing.T) {
	config := &Config{
		ProfilePath: "/nonexistent/profile.json",
	}

	err := config.Apply()
	if err == nil {
		t.Error("Config.Apply() should fail with invalid profile path")
	}
}

func TestProfileJSONMarshaling(t *testing.T) {
	profile := &Profile{
		DefaultAction: ActErrno,
		Architectures: []string{"SCMP_ARCH_X86_64"},
		Syscalls: []*Syscall{
			{
				Names:  []string{"read", "write"},
				Action: ActAllow,
				Args: []*Arg{
					{
						Index: 0,
						Value: 1,
						Op:    OpEqualTo,
					},
				},
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(profile)
	if err != nil {
		t.Fatalf("Failed to marshal profile: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled Profile
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal profile: %v", err)
	}

	// Verify
	if unmarshaled.DefaultAction != profile.DefaultAction {
		t.Errorf("DefaultAction mismatch after JSON round-trip")
	}

	if len(unmarshaled.Syscalls) != len(profile.Syscalls) {
		t.Errorf("Syscalls count mismatch after JSON round-trip")
	}

	if len(unmarshaled.Syscalls) > 0 && len(unmarshaled.Syscalls[0].Args) > 0 {
		arg := unmarshaled.Syscalls[0].Args[0]
		if arg.Index != 0 || arg.Value != 1 || arg.Op != OpEqualTo {
			t.Errorf("Syscall arg mismatch after JSON round-trip")
		}
	}
}

func TestSyscallStructure(t *testing.T) {
	syscall := &Syscall{
		Names:  []string{"open", "openat"},
		Action: ActAllow,
		Args: []*Arg{
			{
				Index:    1,
				Value:    0x80000,
				ValueTwo: 0,
				Op:       OpMaskedEqual,
			},
		},
	}

	if len(syscall.Names) != 2 {
		t.Errorf("Syscall should have 2 names, got %d", len(syscall.Names))
	}

	if syscall.Action != ActAllow {
		t.Errorf("Syscall action should be ActAllow, got %v", syscall.Action)
	}

	if len(syscall.Args) != 1 {
		t.Fatalf("Syscall should have 1 arg, got %d", len(syscall.Args))
	}

	arg := syscall.Args[0]
	if arg.Index != 1 {
		t.Errorf("Arg index should be 1, got %d", arg.Index)
	}

	if arg.Op != OpMaskedEqual {
		t.Errorf("Arg operator should be OpMaskedEqual, got %v", arg.Op)
	}
}

func TestActionsAndOperators(t *testing.T) {
	// Test action strings
	actions := []Action{ActKill, ActTrap, ActErrno, ActTrace, ActAllow, ActLog}
	for _, action := range actions {
		if string(action) == "" {
			t.Errorf("Action should not be empty: %v", action)
		}
	}

	// Test operator strings
	operators := []Operator{OpNotEqual, OpLessThan, OpLessEqual, OpEqualTo, OpGreaterEqual, OpGreaterThan, OpMaskedEqual}
	for _, op := range operators {
		if string(op) == "" {
			t.Errorf("Operator should not be empty: %v", op)
		}
	}
}

func BenchmarkDefaultProfile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DefaultProfile()
	}
}

func BenchmarkProfileSave(b *testing.B) {
	profile := DefaultProfile()
	tmpDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profilePath := filepath.Join(tmpDir, "bench-profile.json")
		_ = profile.Save(profilePath)
	}
}

func BenchmarkLoadProfile(b *testing.B) {
	tmpDir := b.TempDir()
	profilePath := filepath.Join(tmpDir, "bench-profile.json")

	profile := DefaultProfile()
	if err := profile.Save(profilePath); err != nil {
		b.Fatalf("Failed to save profile: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = LoadProfile(profilePath)
	}
}
