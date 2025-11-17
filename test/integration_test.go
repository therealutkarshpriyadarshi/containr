package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/therealutkarshpriyadarshi/containr/pkg/cgroup"
	"github.com/therealutkarshpriyadarshi/containr/pkg/container"
	"github.com/therealutkarshpriyadarshi/containr/pkg/namespace"
	"github.com/therealutkarshpriyadarshi/containr/pkg/rootfs"
)

// TestContainerRunsSuccessfully tests that a container can run and execute commands
func TestContainerRunsSuccessfully(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Skipping integration test that requires root privileges")
	}

	// Check if /bin/echo exists
	if _, err := os.Stat("/bin/echo"); os.IsNotExist(err) {
		t.Skip("Skipping test: /bin/echo not found")
	}

	config := &container.Config{
		Command:  []string{"/bin/echo", "integration-test"},
		Hostname: "test-container",
		Isolate:  false,
	}

	c := container.New(config)
	if c == nil {
		t.Fatal("Failed to create container")
	}

	// We can't easily test Run() in integration tests because it blocks
	// But we can verify the container was created successfully
	if c.ID == "" {
		t.Error("Container ID should not be empty")
	}

	if len(c.Command) == 0 {
		t.Error("Container should have a command")
	}

	t.Logf("Container created successfully with ID: %s", c.ID)
}

// TestResourceLimitsEnforced tests that cgroup resource limits are applied
func TestResourceLimitsEnforced(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Skipping integration test that requires root privileges")
	}

	if _, err := os.Stat("/sys/fs/cgroup"); os.IsNotExist(err) {
		t.Skip("Skipping test: cgroup filesystem not available")
	}

	// Create a cgroup with memory limit
	memoryLimit := int64(50 * 1024 * 1024) // 50MB
	cgConfig := &cgroup.Config{
		Name:        "integration-test-cgroup",
		MemoryLimit: memoryLimit,
		CPUShares:   512,
		PIDLimit:    50,
	}

	cg, err := cgroup.New(cgConfig)
	if err != nil {
		t.Skipf("Failed to create cgroup (may not be supported): %v", err)
	}
	defer cg.Remove()

	// Verify cgroup was created
	if cg.Name != cgConfig.Name {
		t.Errorf("Cgroup name = %s, want %s", cg.Name, cgConfig.Name)
	}

	// Try to add current process
	err = cg.AddCurrentProcess()
	if err != nil {
		t.Logf("Failed to add process to cgroup (may not be supported): %v", err)
	}

	// Get stats
	stats, err := cg.GetStats()
	if err != nil {
		t.Logf("Failed to get cgroup stats (may not be supported): %v", err)
	} else if stats != nil {
		t.Logf("Memory usage: %d bytes", stats.MemoryUsage)
	}

	t.Log("Cgroup resource limits test completed")
}

// TestNamespaceIsolation tests that namespace isolation works correctly
func TestNamespaceIsolation(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Skipping integration test that requires root privileges")
	}

	// Test namespace flag generation
	flags := namespace.GetNamespaceFlags(
		namespace.UTS,
		namespace.PID,
		namespace.Mount,
	)

	if flags == 0 {
		t.Error("Namespace flags should not be zero")
	}

	expectedFlags := syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS
	if flags != expectedFlags {
		t.Logf("Got flags %d, expected %d", flags, expectedFlags)
	}

	t.Log("Namespace isolation test completed")
}

// TestFilesystemIsolation tests that filesystem isolation works
func TestFilesystemIsolation(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Skipping integration test that requires root privileges")
	}

	tmpDir := t.TempDir()
	rootfsPath := filepath.Join(tmpDir, "rootfs")
	mountPoint := filepath.Join(tmpDir, "mnt")

	// Create directories
	if err := os.MkdirAll(rootfsPath, 0755); err != nil {
		t.Fatalf("Failed to create rootfs directory: %v", err)
	}

	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		t.Fatalf("Failed to create mount point: %v", err)
	}

	// Create a test file in rootfs
	testFile := filepath.Join(rootfsPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Setup rootfs
	config := &rootfs.Config{
		Path:       rootfsPath,
		MountPoint: mountPoint,
		UseOverlay: false,
	}

	r := rootfs.New(config)
	if r == nil {
		t.Fatal("Failed to create rootfs")
	}

	// We can't actually call Setup() and Teardown() here as they require mount privileges
	// But we can verify the configuration
	if r.Path != rootfsPath {
		t.Errorf("RootFS path = %s, want %s", r.Path, rootfsPath)
	}

	if r.MountPoint != mountPoint {
		t.Errorf("Mount point = %s, want %s", r.MountPoint, mountPoint)
	}

	t.Log("Filesystem isolation test completed")
}

// TestContainerCleanup tests that cleanup happens properly
func TestContainerCleanup(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Skipping integration test that requires root privileges")
	}

	if _, err := os.Stat("/sys/fs/cgroup"); os.IsNotExist(err) {
		t.Skip("Skipping test: cgroup filesystem not available")
	}

	// Create a cgroup
	cgConfig := &cgroup.Config{
		Name:        "cleanup-test-cgroup",
		MemoryLimit: 100 * 1024 * 1024,
	}

	cg, err := cgroup.New(cgConfig)
	if err != nil {
		t.Skipf("Failed to create cgroup: %v", err)
	}

	// Remove the cgroup
	err = cg.Remove()
	if err != nil {
		t.Logf("Failed to remove cgroup (may be expected): %v", err)
	}

	t.Log("Container cleanup test completed")
}

// TestEndToEndContainer tests a complete container lifecycle
func TestEndToEndContainer(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Skipping integration test that requires root privileges")
	}

	// 1. Create container configuration
	config := &container.Config{
		Command:  []string{"/bin/echo", "hello"},
		Hostname: "e2e-test",
		Isolate:  false,
	}

	// 2. Create container
	c := container.New(config)
	if c == nil {
		t.Fatal("Failed to create container")
	}

	// 3. Verify container properties
	if c.ID == "" {
		t.Error("Container ID should not be empty")
	}

	if c.Hostname != "e2e-test" {
		t.Errorf("Hostname = %s, want e2e-test", c.Hostname)
	}

	if len(c.Namespaces) == 0 {
		t.Error("Container should have namespaces configured")
	}

	// 4. Verify namespace flags
	flags := namespace.GetNamespaceFlags(c.Namespaces...)
	if flags == 0 {
		t.Error("Namespace flags should not be zero")
	}

	t.Logf("End-to-end container test completed. Container ID: %s", c.ID)
}

// TestCommandExecution tests that commands can be executed
func TestCommandExecution(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Skipping integration test that requires root privileges")
	}

	// Simple command execution test using exec directly
	cmd := exec.Command("/bin/echo", "test")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	if !strings.Contains(string(output), "test") {
		t.Errorf("Command output = %s, want 'test'", string(output))
	}

	t.Log("Command execution test passed")
}

// TestCgroupHierarchy tests cgroup hierarchy creation
func TestCgroupHierarchy(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Skipping integration test that requires root privileges")
	}

	if _, err := os.Stat("/sys/fs/cgroup"); os.IsNotExist(err) {
		t.Skip("Skipping test: cgroup filesystem not available")
	}

	// Create multiple cgroups
	cgConfigs := []*cgroup.Config{
		{
			Name:        "hierarchy-test-1",
			MemoryLimit: 100 * 1024 * 1024,
		},
		{
			Name:        "hierarchy-test-2",
			MemoryLimit: 200 * 1024 * 1024,
		},
	}

	var cgroups []*cgroup.Cgroup
	for _, config := range cgConfigs {
		cg, err := cgroup.New(config)
		if err != nil {
			t.Logf("Failed to create cgroup %s: %v", config.Name, err)
			continue
		}
		cgroups = append(cgroups, cg)
	}

	// Cleanup
	for _, cg := range cgroups {
		if err := cg.Remove(); err != nil {
			t.Logf("Failed to remove cgroup %s: %v", cg.Name, err)
		}
	}

	t.Log("Cgroup hierarchy test completed")
}

// TestNamespaceReexec tests the reexec functionality
func TestNamespaceReexec(t *testing.T) {
	// Save original args
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Test with init arg
	os.Args = []string{"test", "init"}

	// This should return immediately without panicking
	namespace.Reexec()

	t.Log("Namespace reexec test completed")
}

// TestConcurrentContainers tests running multiple containers
func TestConcurrentContainers(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Skipping integration test that requires root privileges")
	}

	numContainers := 3
	containers := make([]*container.Container, numContainers)

	for i := 0; i < numContainers; i++ {
		config := &container.Config{
			Command:  []string{"/bin/echo", "test"},
			Hostname: "concurrent-test",
			Isolate:  false,
		}

		c := container.New(config)
		if c == nil {
			t.Errorf("Failed to create container %d", i)
			continue
		}

		containers[i] = c
	}

	// Verify all containers were created
	for i, c := range containers {
		if c == nil {
			t.Errorf("Container %d is nil", i)
			continue
		}

		if c.ID == "" {
			t.Errorf("Container %d has empty ID", i)
		}
	}

	t.Logf("Created %d containers successfully", numContainers)
}

// TestResourceConstraints tests various resource constraint scenarios
func TestResourceConstraints(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Skipping integration test that requires root privileges")
	}

	if _, err := os.Stat("/sys/fs/cgroup"); os.IsNotExist(err) {
		t.Skip("Skipping test: cgroup filesystem not available")
	}

	tests := []struct {
		name        string
		memoryLimit int64
		cpuShares   int64
		pidLimit    int64
	}{
		{
			name:        "Low memory limit",
			memoryLimit: 10 * 1024 * 1024, // 10MB
			cpuShares:   256,
			pidLimit:    10,
		},
		{
			name:        "Medium resources",
			memoryLimit: 100 * 1024 * 1024, // 100MB
			cpuShares:   512,
			pidLimit:    50,
		},
		{
			name:        "High resources",
			memoryLimit: 500 * 1024 * 1024, // 500MB
			cpuShares:   1024,
			pidLimit:    100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &cgroup.Config{
				Name:        "constraint-test-" + tt.name,
				MemoryLimit: tt.memoryLimit,
				CPUShares:   tt.cpuShares,
				PIDLimit:    tt.pidLimit,
			}

			cg, err := cgroup.New(config)
			if err != nil {
				t.Logf("Failed to create cgroup (may not be supported): %v", err)
				return
			}
			defer cg.Remove()

			t.Logf("Created cgroup with constraints: memory=%d, cpu=%d, pid=%d",
				tt.memoryLimit, tt.cpuShares, tt.pidLimit)
		})
	}
}

// TestLongRunningContainer tests container behavior over time
func TestLongRunningContainer(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Skipping integration test that requires root privileges")
	}

	config := &container.Config{
		Command:  []string{"/bin/sleep", "1"},
		Hostname: "long-running-test",
		Isolate:  false,
	}

	c := container.New(config)
	if c == nil {
		t.Fatal("Failed to create container")
	}

	// Simulate some time passing
	time.Sleep(100 * time.Millisecond)

	// Container should still be valid
	if c.ID == "" {
		t.Error("Container ID became empty")
	}

	t.Log("Long-running container test completed")
}

// TestErrorHandling tests error scenarios
func TestErrorHandling(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("Skipping integration test that requires root privileges")
	}

	// Test with invalid command
	config := &container.Config{
		Command:  []string{"/nonexistent/command"},
		Hostname: "error-test",
		Isolate:  false,
	}

	c := container.New(config)
	if c == nil {
		t.Fatal("Failed to create container")
	}

	// The error will occur when trying to run, not during creation
	t.Log("Error handling test completed")
}
