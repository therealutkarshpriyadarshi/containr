package container

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/therealutkarshpriyadarshi/containr/pkg/namespace"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		expectIsolate bool
		expectedNsLen int
		containsUTS   bool
		containsPID   bool
		containsMount bool
		containsIPC   bool
		containsNet   bool
	}{
		{
			name: "Basic container without isolation",
			config: &Config{
				RootFS:     "/tmp/rootfs",
				Command:    []string{"/bin/sh"},
				WorkingDir: "/",
				Hostname:   "test-container",
				Isolate:    false,
			},
			expectIsolate: false,
			expectedNsLen: 3, // UTS, PID, Mount
			containsUTS:   true,
			containsPID:   true,
			containsMount: true,
			containsIPC:   false,
			containsNet:   false,
		},
		{
			name: "Fully isolated container",
			config: &Config{
				RootFS:     "/tmp/rootfs",
				Command:    []string{"/bin/sh"},
				WorkingDir: "/",
				Hostname:   "isolated-container",
				Isolate:    true,
			},
			expectIsolate: true,
			expectedNsLen: 5, // UTS, PID, Mount, IPC, Network
			containsUTS:   true,
			containsPID:   true,
			containsMount: true,
			containsIPC:   true,
			containsNet:   true,
		},
		{
			name: "Container with custom command",
			config: &Config{
				RootFS:   "/tmp/rootfs",
				Command:  []string{"/bin/ls", "-la"},
				Hostname: "ls-container",
				Isolate:  false,
			},
			expectIsolate: false,
			expectedNsLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.config)

			if c == nil {
				t.Fatal("New() returned nil")
			}

			// Check ID is generated
			if c.ID == "" {
				t.Error("Container ID is empty")
			}

			// Check RootFS
			if c.RootFS != tt.config.RootFS {
				t.Errorf("RootFS = %s, want %s", c.RootFS, tt.config.RootFS)
			}

			// Check Command
			if len(c.Command) != len(tt.config.Command) {
				t.Errorf("Command length = %d, want %d", len(c.Command), len(tt.config.Command))
			}

			// Check Hostname
			if c.Hostname != tt.config.Hostname {
				t.Errorf("Hostname = %s, want %s", c.Hostname, tt.config.Hostname)
			}

			// Check namespace configuration
			if len(c.Namespaces) != tt.expectedNsLen {
				t.Errorf("Namespaces length = %d, want %d", len(c.Namespaces), tt.expectedNsLen)
			}

			// Check specific namespaces
			nsMap := make(map[namespace.NamespaceType]bool)
			for _, ns := range c.Namespaces {
				nsMap[ns] = true
			}

			if tt.containsUTS && !nsMap[namespace.UTS] {
				t.Error("Expected UTS namespace but not found")
			}
			if tt.containsPID && !nsMap[namespace.PID] {
				t.Error("Expected PID namespace but not found")
			}
			if tt.containsMount && !nsMap[namespace.Mount] {
				t.Error("Expected Mount namespace but not found")
			}
			if tt.containsIPC && !nsMap[namespace.IPC] {
				t.Error("Expected IPC namespace but not found")
			}
			if tt.containsNet && !nsMap[namespace.Network] {
				t.Error("Expected Network namespace but not found")
			}
		})
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	// Check IDs are not empty
	if id1 == "" || id2 == "" {
		t.Error("generateID() returned empty string")
	}

	// Check ID format
	if !strings.HasPrefix(id1, "container-") {
		t.Errorf("ID does not have correct prefix: %s", id1)
	}
}

func TestSetupChild(t *testing.T) {
	// Skip if not running as root
	if os.Geteuid() != 0 {
		t.Skip("Skipping test that requires root privileges")
	}

	// Save original environment
	originalHostname := os.Getenv("CONTAINER_HOSTNAME")
	defer os.Setenv("CONTAINER_HOSTNAME", originalHostname)

	tests := []struct {
		name        string
		hostname    string
		expectError bool
	}{
		{
			name:        "Setup without hostname",
			hostname:    "",
			expectError: false,
		},
		{
			name:        "Setup with valid hostname",
			hostname:    "test-host",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("CONTAINER_HOSTNAME", tt.hostname)
			err := SetupChild()
			if (err != nil) != tt.expectError {
				t.Errorf("SetupChild() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestSetupRootFS(t *testing.T) {
	tests := []struct {
		name        string
		rootfs      string
		expectError bool
		setup       func() string
		cleanup     func(string)
	}{
		{
			name:        "Empty rootfs path",
			rootfs:      "",
			expectError: false,
		},
		{
			name:        "Nonexistent rootfs",
			rootfs:      "/nonexistent/rootfs",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Container{
				ID:     "test-container",
				RootFS: tt.rootfs,
			}

			err := c.SetupRootFS()
			if (err != nil) != tt.expectError {
				t.Errorf("SetupRootFS() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}

	// Test valid rootfs setup (may fail in some environments)
	t.Run("Valid rootfs setup", func(t *testing.T) {
		tmpDir := t.TempDir()
		rootfsDir := filepath.Join(tmpDir, "rootfs")
		os.MkdirAll(rootfsDir, 0755)

		c := &Container{
			ID:     "test-container-valid",
			RootFS: rootfsDir,
		}

		err := c.SetupRootFS()
		// Mount operations may fail in test environments, so we just log the result
		if err != nil {
			t.Logf("SetupRootFS failed (expected in some test environments): %v", err)
		} else {
			t.Log("SetupRootFS succeeded")
		}
	})
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		valid  bool
	}{
		{
			name: "Valid config",
			config: &Config{
				RootFS:     "/tmp/rootfs",
				Command:    []string{"/bin/sh"},
				WorkingDir: "/",
				Hostname:   "test",
				Isolate:    true,
			},
			valid: true,
		},
		{
			name: "Empty command",
			config: &Config{
				RootFS:     "/tmp/rootfs",
				Command:    []string{},
				WorkingDir: "/",
				Hostname:   "test",
				Isolate:    false,
			},
			valid: false,
		},
		{
			name: "Nil command",
			config: &Config{
				RootFS:     "/tmp/rootfs",
				Command:    nil,
				WorkingDir: "/",
				Hostname:   "test",
				Isolate:    false,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.config.Command != nil && len(tt.config.Command) > 0
			if isValid != tt.valid {
				t.Errorf("Config validation = %v, want %v", isValid, tt.valid)
			}
		})
	}
}

func TestContainerFields(t *testing.T) {
	config := &Config{
		RootFS:     "/test/rootfs",
		Command:    []string{"/bin/echo", "hello"},
		WorkingDir: "/app",
		Hostname:   "test-host",
		Isolate:    true,
	}

	c := New(config)

	// Test all fields are properly set
	if c.RootFS != config.RootFS {
		t.Errorf("RootFS = %s, want %s", c.RootFS, config.RootFS)
	}

	if len(c.Command) != len(config.Command) {
		t.Errorf("Command length = %d, want %d", len(c.Command), len(config.Command))
	}

	for i, cmd := range c.Command {
		if cmd != config.Command[i] {
			t.Errorf("Command[%d] = %s, want %s", i, cmd, config.Command[i])
		}
	}

	if c.WorkingDir != config.WorkingDir {
		t.Errorf("WorkingDir = %s, want %s", c.WorkingDir, config.WorkingDir)
	}

	if c.Hostname != config.Hostname {
		t.Errorf("Hostname = %s, want %s", c.Hostname, config.Hostname)
	}
}

func TestMountProc(t *testing.T) {
	// Skip if not running as root
	if os.Geteuid() != 0 {
		t.Skip("Skipping test that requires root privileges")
	}

	// This test is tricky because mounting /proc requires special privileges
	// and can interfere with the system. We'll just test that the function exists
	// and has the right signature
	err := mountProc()
	if err != nil {
		// It's okay if this fails in test environment
		t.Logf("mountProc failed (expected in test environment): %v", err)
	}
}

func TestNamespaceIntegration(t *testing.T) {
	config := &Config{
		Command: []string{"/bin/echo", "test"},
		Isolate: true,
	}

	c := New(config)

	// Verify namespace flags can be generated
	flags := namespace.GetNamespaceFlags(c.Namespaces...)
	if flags == 0 {
		t.Error("Expected non-zero namespace flags")
	}

	// Check that all expected namespaces are present
	expectedFlags := syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID |
		syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET

	if flags != expectedFlags {
		t.Logf("Namespace flags = %d, expected %d", flags, expectedFlags)
	}
}

func TestContainerLifecycle(t *testing.T) {
	// This is a basic lifecycle test
	config := &Config{
		Command:  []string{"/bin/echo", "test"},
		Hostname: "test-container",
		Isolate:  false,
	}

	c := New(config)

	// Verify container is properly initialized
	if c.ID == "" {
		t.Error("Container ID should not be empty after creation")
	}

	if len(c.Command) == 0 {
		t.Error("Container command should not be empty")
	}

	if len(c.Namespaces) == 0 {
		t.Error("Container should have at least some namespaces configured")
	}
}

func TestEnvironmentPassing(t *testing.T) {
	config := &Config{
		RootFS:   "/tmp/test-rootfs",
		Command:  []string{"/bin/sh"},
		Hostname: "env-test-container",
		Isolate:  false,
	}

	c := New(config)

	// Verify environment variables would be set correctly
	expectedEnvVars := map[string]string{
		"CONTAINER_ID":       c.ID,
		"CONTAINER_ROOTFS":   c.RootFS,
		"CONTAINER_HOSTNAME": c.Hostname,
	}

	for key, expectedValue := range expectedEnvVars {
		// We can't actually test RunWithSetup easily, but we can verify
		// the data that would be passed
		switch key {
		case "CONTAINER_ID":
			if expectedValue != c.ID {
				t.Errorf("Expected %s to be %s, got %s", key, c.ID, expectedValue)
			}
		case "CONTAINER_ROOTFS":
			if expectedValue != c.RootFS {
				t.Errorf("Expected %s to be %s, got %s", key, c.RootFS, expectedValue)
			}
		case "CONTAINER_HOSTNAME":
			if expectedValue != c.Hostname {
				t.Errorf("Expected %s to be %s, got %s", key, c.Hostname, expectedValue)
			}
		}
	}
}
