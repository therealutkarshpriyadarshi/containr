package profiler

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProfiler(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		OutputDir: tmpDir,
		CPU:       true,
		Memory:    true,
		Trace:     true,
	}

	p, err := New(config)
	if err != nil {
		t.Fatalf("failed to create profiler: %v", err)
	}

	// Test CPU profiling
	if err := p.StartCPUProfile(); err != nil {
		t.Fatalf("failed to start CPU profile: %v", err)
	}

	// Do some work
	time.Sleep(10 * time.Millisecond)

	if err := p.StopCPUProfile(); err != nil {
		t.Fatalf("failed to stop CPU profile: %v", err)
	}

	// Verify CPU profile was created
	cpuProfile := filepath.Join(tmpDir, "cpu.prof")
	if _, err := os.Stat(cpuProfile); os.IsNotExist(err) {
		t.Errorf("CPU profile was not created")
	}

	// Test memory profiling
	if err := p.WriteMemProfile(); err != nil {
		t.Fatalf("failed to write memory profile: %v", err)
	}

	memProfile := filepath.Join(tmpDir, "mem.prof")
	if _, err := os.Stat(memProfile); os.IsNotExist(err) {
		t.Errorf("memory profile was not created")
	}
}

func TestTrace(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		OutputDir: tmpDir,
	}

	p, err := New(config)
	if err != nil {
		t.Fatalf("failed to create profiler: %v", err)
	}

	if err := p.StartTrace(); err != nil {
		t.Fatalf("failed to start trace: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	if err := p.StopTrace(); err != nil {
		t.Fatalf("failed to stop trace: %v", err)
	}

	traceFile := filepath.Join(tmpDir, "trace.out")
	if _, err := os.Stat(traceFile); os.IsNotExist(err) {
		t.Errorf("trace file was not created")
	}
}

func TestMemStats(t *testing.T) {
	stats := MemStats()
	if stats == nil {
		t.Errorf("MemStats returned nil")
	}

	if stats.Alloc == 0 {
		t.Errorf("expected non-zero allocation")
	}
}

func TestGoroutineCount(t *testing.T) {
	count := GoroutineCount()
	if count == 0 {
		t.Errorf("expected non-zero goroutine count")
	}
}

func TestWriteAllProfiles(t *testing.T) {
	tmpDir := t.TempDir()

	config := &Config{
		OutputDir: tmpDir,
	}

	p, err := New(config)
	if err != nil {
		t.Fatalf("failed to create profiler: %v", err)
	}

	if err := p.WriteAllProfiles(); err != nil {
		t.Fatalf("failed to write all profiles: %v", err)
	}

	// Check that at least some profiles were created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to read directory: %v", err)
	}

	if len(files) == 0 {
		t.Errorf("no profile files were created")
	}
}
