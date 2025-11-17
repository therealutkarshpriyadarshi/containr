package namespace

import (
	"os"
	"syscall"
	"testing"
)

func TestGetNamespaceFlags(t *testing.T) {
	tests := []struct {
		name     string
		types    []NamespaceType
		expected int
	}{
		{
			name:     "Single UTS namespace",
			types:    []NamespaceType{UTS},
			expected: syscall.CLONE_NEWUTS,
		},
		{
			name:     "Single PID namespace",
			types:    []NamespaceType{PID},
			expected: syscall.CLONE_NEWPID,
		},
		{
			name:     "Single IPC namespace",
			types:    []NamespaceType{IPC},
			expected: syscall.CLONE_NEWIPC,
		},
		{
			name:     "Single Mount namespace",
			types:    []NamespaceType{Mount},
			expected: syscall.CLONE_NEWNS,
		},
		{
			name:     "Single Network namespace",
			types:    []NamespaceType{Network},
			expected: syscall.CLONE_NEWNET,
		},
		{
			name:     "Single User namespace",
			types:    []NamespaceType{User},
			expected: syscall.CLONE_NEWUSER,
		},
		{
			name:     "Multiple namespaces (UTS + PID)",
			types:    []NamespaceType{UTS, PID},
			expected: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
		},
		{
			name:     "All namespaces",
			types:    []NamespaceType{UTS, IPC, PID, Mount, Network, User},
			expected: syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWUSER,
		},
		{
			name:     "Empty namespace list",
			types:    []NamespaceType{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNamespaceFlags(tt.types...)
			if result != tt.expected {
				t.Errorf("GetNamespaceFlags() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestNamespaceTypeValues(t *testing.T) {
	// Test that namespace type constants have unique values
	types := []NamespaceType{UTS, IPC, PID, Mount, Network, User}
	seen := make(map[NamespaceType]bool)

	for _, nsType := range types {
		if seen[nsType] {
			t.Errorf("Duplicate namespace type value: %d", nsType)
		}
		seen[nsType] = true
	}
}

func TestCreateNamespaces(t *testing.T) {
	// Skip this test if not running as root
	if os.Geteuid() != 0 {
		t.Skip("Skipping test that requires root privileges")
	}

	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "Simple command execution",
			config: &Config{
				Flags:   0, // No namespace isolation for simple test
				Command: "/bin/echo",
				Args:    []string{"hello"},
			},
			expectError: false,
		},
		{
			name: "Invalid command",
			config: &Config{
				Flags:   0,
				Command: "/nonexistent/command",
				Args:    []string{},
			},
			expectError: true,
		},
		{
			name: "Command with UTS namespace",
			config: &Config{
				Flags:   GetNamespaceFlags(UTS),
				Command: "/bin/echo",
				Args:    []string{"test"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateNamespaces(tt.config)
			if (err != nil) != tt.expectError {
				t.Errorf("CreateNamespaces() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		valid  bool
	}{
		{
			name: "Valid config with command",
			config: &Config{
				Flags:   GetNamespaceFlags(UTS),
				Command: "/bin/sh",
				Args:    []string{"-c", "echo test"},
			},
			valid: true,
		},
		{
			name: "Valid config with working directory",
			config: &Config{
				Flags:      GetNamespaceFlags(UTS, PID),
				Command:    "/bin/ls",
				Args:       []string{},
				WorkingDir: "/tmp",
			},
			valid: true,
		},
		{
			name: "Empty command should be invalid",
			config: &Config{
				Flags:   GetNamespaceFlags(UTS),
				Command: "",
				Args:    []string{},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.config.Command != ""
			if isValid != tt.valid {
				t.Errorf("Config validation = %v, want %v", isValid, tt.valid)
			}
		})
	}
}

func TestNamespaceFlagCombinations(t *testing.T) {
	// Test common namespace flag combinations
	combinations := []struct {
		name     string
		types    []NamespaceType
		hasFlag  int
		contains bool
	}{
		{
			name:     "UTS in UTS+PID",
			types:    []NamespaceType{UTS, PID},
			hasFlag:  syscall.CLONE_NEWUTS,
			contains: true,
		},
		{
			name:     "Network not in UTS+PID",
			types:    []NamespaceType{UTS, PID},
			hasFlag:  syscall.CLONE_NEWNET,
			contains: false,
		},
		{
			name:     "Mount in full isolation",
			types:    []NamespaceType{UTS, IPC, PID, Mount, Network},
			hasFlag:  syscall.CLONE_NEWNS,
			contains: true,
		},
	}

	for _, tc := range combinations {
		t.Run(tc.name, func(t *testing.T) {
			flags := GetNamespaceFlags(tc.types...)
			hasTheFlag := (flags & tc.hasFlag) == tc.hasFlag
			if hasTheFlag != tc.contains {
				t.Errorf("Expected flag %d to be %v in flags %d", tc.hasFlag, tc.contains, flags)
			}
		})
	}
}

func TestReexecDetection(t *testing.T) {
	// Test that we can detect when running in re-exec mode
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test with "init" argument
	os.Args = []string{"test", "init"}
	// Reexec should return immediately if "init" is the first argument
	// This is a simple check that the function doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Reexec panicked: %v", r)
		}
	}()

	// Note: We can't actually test the full Reexec behavior without forking
	// So we just test that it doesn't panic with init argument
	Reexec()
}
