package debug

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func TestNewDebugger(t *testing.T) {
	config := &Config{
		ContainerID: "test-container",
		EnableTrace: true,
	}

	debugger := NewDebugger(config)
	if debugger == nil {
		t.Fatal("Expected debugger to be created")
	}

	if debugger.containerID != "test-container" {
		t.Errorf("Expected container ID %s, got %s", "test-container", debugger.containerID)
	}
}

func TestDebugger_AddBreakpoint(t *testing.T) {
	debugger := NewDebugger(&Config{ContainerID: "test"})

	bp, err := debugger.AddBreakpoint("syscall:open", BreakpointTypeSyscall)
	if err != nil {
		t.Fatalf("Failed to add breakpoint: %v", err)
	}

	if bp.Location != "syscall:open" {
		t.Errorf("Expected location %s, got %s", "syscall:open", bp.Location)
	}

	if !bp.Enabled {
		t.Error("Expected breakpoint to be enabled")
	}
}

func TestDebugger_RemoveBreakpoint(t *testing.T) {
	debugger := NewDebugger(&Config{ContainerID: "test"})

	bp, _ := debugger.AddBreakpoint("syscall:open", BreakpointTypeSyscall)

	err := debugger.RemoveBreakpoint(bp.ID)
	if err != nil {
		t.Fatalf("Failed to remove breakpoint: %v", err)
	}

	bps := debugger.ListBreakpoints()
	if len(bps) != 0 {
		t.Errorf("Expected 0 breakpoints, got %d", len(bps))
	}
}

func TestDebugger_ProfileCPU(t *testing.T) {
	debugger := NewDebugger(&Config{ContainerID: "test"})
	var buf bytes.Buffer

	err := debugger.ProfileCPU(100*time.Millisecond, &buf)
	if err != nil {
		t.Fatalf("Failed to profile CPU: %v", err)
	}
}

func TestInteractiveSession(t *testing.T) {
	debugger := NewDebugger(&Config{ContainerID: "test"})
	session := NewInteractiveSession(debugger)

	if session == nil {
		t.Fatal("Expected session to be created")
	}

	if session.debugger != debugger {
		t.Error("Expected session to reference debugger")
	}
}

func TestProfiler_New(t *testing.T) {
	config := &ProfilerConfig{
		CPUProfile:   true,
		MemProfile:   true,
		OutputDir:    "/tmp",
	}

	profiler := NewProfiler(config)
	if profiler == nil {
		t.Fatal("Expected profiler to be created")
	}

	if !profiler.config.CPUProfile {
		t.Error("Expected CPU profile to be enabled")
	}
}

func TestProfiler_StartStop(t *testing.T) {
	config := &ProfilerConfig{
		OutputDir: "/tmp/containr-test",
	}

	profiler := NewProfiler(config)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := profiler.StartProfiling(ctx); err != nil {
		t.Fatalf("Failed to start profiling: %v", err)
	}

	<-ctx.Done()

	if err := profiler.StopProfiling(); err != nil {
		t.Fatalf("Failed to stop profiling: %v", err)
	}
}
