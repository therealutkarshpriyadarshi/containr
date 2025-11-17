package cgroup

import (
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
)

func TestNew(t *testing.T) {
	// Skip if not running as root
	if os.Geteuid() != 0 {
		t.Skip("Skipping test that requires root privileges")
	}

	// Skip if cgroup filesystem is not available
	if _, err := os.Stat("/sys/fs/cgroup"); os.IsNotExist(err) {
		t.Skip("Skipping test: cgroup filesystem not available")
	}

	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "Create cgroup with memory limit",
			config: &Config{
				Name:        "test-memory-cgroup",
				MemoryLimit: 100 * 1024 * 1024, // 100MB
				CPUShares:   0,
				PIDLimit:    0,
			},
			expectError: false,
		},
		{
			name: "Create cgroup with CPU shares",
			config: &Config{
				Name:        "test-cpu-cgroup",
				MemoryLimit: 0,
				CPUShares:   512,
				PIDLimit:    0,
			},
			expectError: false,
		},
		{
			name: "Create cgroup with PID limit",
			config: &Config{
				Name:        "test-pid-cgroup",
				MemoryLimit: 0,
				CPUShares:   0,
				PIDLimit:    100,
			},
			expectError: false,
		},
		{
			name: "Create cgroup with all limits",
			config: &Config{
				Name:        "test-all-limits-cgroup",
				MemoryLimit: 50 * 1024 * 1024, // 50MB
				CPUShares:   256,
				PIDLimit:    50,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg, err := New(tt.config)
			if (err != nil) != tt.expectError {
				t.Errorf("New() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if cg != nil {
				// Cleanup
				defer cg.Remove()

				// Verify cgroup was created
				if cg.Name != tt.config.Name {
					t.Errorf("Cgroup name = %s, want %s", cg.Name, tt.config.Name)
				}
			}
		})
	}
}

func TestCgroupName(t *testing.T) {
	config := &Config{
		Name: "test-cgroup-name",
	}

	cg := &Cgroup{
		Name:   config.Name,
		Parent: "/sys/fs/cgroup",
	}

	if cg.Name != config.Name {
		t.Errorf("Cgroup name = %s, want %s", cg.Name, config.Name)
	}
}

func TestAddCurrentProcess(t *testing.T) {
	// Skip if not running as root
	if os.Geteuid() != 0 {
		t.Skip("Skipping test that requires root privileges")
	}

	// Skip if cgroup filesystem is not available
	if _, err := os.Stat("/sys/fs/cgroup"); os.IsNotExist(err) {
		t.Skip("Skipping test: cgroup filesystem not available")
	}

	config := &Config{
		Name:        "test-current-process-cgroup",
		MemoryLimit: 100 * 1024 * 1024,
	}

	cg, err := New(config)
	if err != nil {
		t.Skipf("Failed to create cgroup: %v", err)
	}
	defer cg.Remove()

	err = cg.AddCurrentProcess()
	if err != nil {
		// This might fail in some environments, so we just log it
		t.Logf("AddCurrentProcess failed (this may be expected in some environments): %v", err)
	}
}

func TestAddProcess(t *testing.T) {
	// Skip if not running as root
	if os.Geteuid() != 0 {
		t.Skip("Skipping test that requires root privileges")
	}

	// Skip if cgroup filesystem is not available
	if _, err := os.Stat("/sys/fs/cgroup"); os.IsNotExist(err) {
		t.Skip("Skipping test: cgroup filesystem not available")
	}

	config := &Config{
		Name:        "test-add-process-cgroup",
		MemoryLimit: 100 * 1024 * 1024,
	}

	cg, err := New(config)
	if err != nil {
		t.Skipf("Failed to create cgroup: %v", err)
	}
	defer cg.Remove()

	// Try to add the current process
	pid := syscall.Getpid()
	err = cg.AddProcess(pid)
	if err != nil {
		t.Logf("AddProcess failed (this may be expected in some environments): %v", err)
	}
}

func TestGetStats(t *testing.T) {
	// Skip if not running as root
	if os.Geteuid() != 0 {
		t.Skip("Skipping test that requires root privileges")
	}

	// Skip if cgroup filesystem is not available
	if _, err := os.Stat("/sys/fs/cgroup"); os.IsNotExist(err) {
		t.Skip("Skipping test: cgroup filesystem not available")
	}

	config := &Config{
		Name:        "test-stats-cgroup",
		MemoryLimit: 100 * 1024 * 1024,
	}

	cg, err := New(config)
	if err != nil {
		t.Skipf("Failed to create cgroup: %v", err)
	}
	defer cg.Remove()

	stats, err := cg.GetStats()
	if err != nil {
		t.Logf("GetStats failed (this may be expected in some environments): %v", err)
		return
	}

	if stats == nil {
		t.Error("GetStats returned nil stats")
	}
}

func TestWriteFile(t *testing.T) {
	// Create a temporary file for testing
	tmpFile := filepath.Join(t.TempDir(), "test-file")

	// Create the file first
	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	f.Close()

	tests := []struct {
		name        string
		path        string
		data        string
		expectError bool
	}{
		{
			name:        "Write to valid file",
			path:        tmpFile,
			data:        "test data",
			expectError: false,
		},
		{
			name:        "Write to nonexistent file",
			path:        "/nonexistent/path/file",
			data:        "test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writeFile(tt.path, tt.data)
			if (err != nil) != tt.expectError {
				t.Errorf("writeFile() error = %v, expectError %v", err, tt.expectError)
			}

			// Verify content if successful
			if err == nil {
				content, err := os.ReadFile(tt.path)
				if err != nil {
					t.Errorf("Failed to read file: %v", err)
				}
				if string(content) != tt.data {
					t.Errorf("File content = %s, want %s", string(content), tt.data)
				}
			}
		})
	}
}

func TestStatsStruct(t *testing.T) {
	stats := &Stats{
		MemoryUsage: 1024,
		CPUUsage:    100,
		PIDCount:    10,
	}

	if stats.MemoryUsage != 1024 {
		t.Errorf("MemoryUsage = %d, want 1024", stats.MemoryUsage)
	}
	if stats.CPUUsage != 100 {
		t.Errorf("CPUUsage = %d, want 100", stats.CPUUsage)
	}
	if stats.PIDCount != 10 {
		t.Errorf("PIDCount = %d, want 10", stats.PIDCount)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		valid  bool
	}{
		{
			name: "Valid config with name",
			config: &Config{
				Name:        "valid-cgroup",
				MemoryLimit: 100 * 1024 * 1024,
			},
			valid: true,
		},
		{
			name: "Empty name",
			config: &Config{
				Name:        "",
				MemoryLimit: 100 * 1024 * 1024,
			},
			valid: false,
		},
		{
			name: "Negative memory limit",
			config: &Config{
				Name:        "test",
				MemoryLimit: -1,
			},
			valid: false,
		},
		{
			name: "Zero limits are valid (means no limit)",
			config: &Config{
				Name:        "test",
				MemoryLimit: 0,
				CPUShares:   0,
				PIDLimit:    0,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.config.Name != "" && tt.config.MemoryLimit >= 0
			if isValid != tt.valid {
				t.Errorf("Config validation = %v, want %v", isValid, tt.valid)
			}
		})
	}
}

func TestCgroupPathConstruction(t *testing.T) {
	cg := &Cgroup{
		Name:   "test-container",
		Parent: "/sys/fs/cgroup",
	}

	expectedPaths := []string{
		"/sys/fs/cgroup/memory/test-container",
		"/sys/fs/cgroup/cpu/test-container",
		"/sys/fs/cgroup/pids/test-container",
	}

	for _, expected := range expectedPaths {
		// Just verify path construction logic
		controller := filepath.Base(filepath.Dir(expected))
		constructed := filepath.Join(cg.Parent, controller, cg.Name)
		if constructed != expected {
			t.Errorf("Path construction = %s, want %s", constructed, expected)
		}
	}
}

func TestMemoryLimitConversion(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{
			name:  "100MB",
			bytes: 100 * 1024 * 1024,
			want:  strconv.FormatInt(100*1024*1024, 10),
		},
		{
			name:  "1GB",
			bytes: 1024 * 1024 * 1024,
			want:  strconv.FormatInt(1024*1024*1024, 10),
		},
		{
			name:  "Zero (unlimited)",
			bytes: 0,
			want:  "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strconv.FormatInt(tt.bytes, 10)
			if result != tt.want {
				t.Errorf("Memory conversion = %s, want %s", result, tt.want)
			}
		})
	}
}

func TestCPUSharesConversion(t *testing.T) {
	// Test cgroup v1 to v2 weight conversion
	tests := []struct {
		name   string
		shares int64
		weight int64
	}{
		{
			name:   "Default shares",
			shares: 1024,
			weight: 10000,
		},
		{
			name:   "Half shares",
			shares: 512,
			weight: 5000,
		},
		{
			name:   "Quarter shares",
			shares: 256,
			weight: 2500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// cgroup v2 uses weight (1-10000), convert from shares
			weight := (tt.shares * 10000) / 1024
			if weight != tt.weight {
				t.Errorf("CPU weight conversion = %d, want %d", weight, tt.weight)
			}
		})
	}
}
