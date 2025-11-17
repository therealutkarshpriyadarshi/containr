// Package profiler provides runtime profiling capabilities for containr
package profiler

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
)

// Profiler manages CPU, memory, and trace profiling
type Profiler struct {
	cpuFile   *os.File
	memFile   *os.File
	traceFile *os.File
	outputDir string
}

// Config holds profiler configuration
type Config struct {
	OutputDir string
	CPU       bool
	Memory    bool
	Trace     bool
}

// New creates a new profiler instance
func New(config *Config) (*Profiler, error) {
	if config.OutputDir == "" {
		config.OutputDir = "."
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return &Profiler{
		outputDir: config.OutputDir,
	}, nil
}

// StartCPUProfile starts CPU profiling
func (p *Profiler) StartCPUProfile() error {
	if p.cpuFile != nil {
		return fmt.Errorf("CPU profiling already started")
	}

	filename := filepath.Join(p.outputDir, "cpu.prof")
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CPU profile: %w", err)
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return fmt.Errorf("failed to start CPU profile: %w", err)
	}

	p.cpuFile = f
	return nil
}

// StopCPUProfile stops CPU profiling
func (p *Profiler) StopCPUProfile() error {
	if p.cpuFile == nil {
		return fmt.Errorf("CPU profiling not started")
	}

	pprof.StopCPUProfile()
	if err := p.cpuFile.Close(); err != nil {
		return fmt.Errorf("failed to close CPU profile: %w", err)
	}

	p.cpuFile = nil
	return nil
}

// WriteMemProfile writes a memory profile
func (p *Profiler) WriteMemProfile() error {
	filename := filepath.Join(p.outputDir, "mem.prof")
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create memory profile: %w", err)
	}
	defer f.Close()

	runtime.GC() // Get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("failed to write memory profile: %w", err)
	}

	return nil
}

// StartTrace starts execution tracing
func (p *Profiler) StartTrace() error {
	if p.traceFile != nil {
		return fmt.Errorf("trace already started")
	}

	filename := filepath.Join(p.outputDir, "trace.out")
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create trace file: %w", err)
	}

	if err := trace.Start(f); err != nil {
		f.Close()
		return fmt.Errorf("failed to start trace: %w", err)
	}

	p.traceFile = f
	return nil
}

// StopTrace stops execution tracing
func (p *Profiler) StopTrace() error {
	if p.traceFile == nil {
		return fmt.Errorf("trace not started")
	}

	trace.Stop()
	if err := p.traceFile.Close(); err != nil {
		return fmt.Errorf("failed to close trace file: %w", err)
	}

	p.traceFile = nil
	return nil
}

// WriteAllProfiles writes all available runtime profiles
func (p *Profiler) WriteAllProfiles() error {
	profiles := []string{"goroutine", "heap", "allocs", "threadcreate", "block", "mutex"}

	for _, name := range profiles {
		filename := filepath.Join(p.outputDir, fmt.Sprintf("%s.prof", name))
		f, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("failed to create %s profile: %w", name, err)
		}

		profile := pprof.Lookup(name)
		if profile == nil {
			f.Close()
			continue
		}

		if err := profile.WriteTo(f, 0); err != nil {
			f.Close()
			return fmt.Errorf("failed to write %s profile: %w", name, err)
		}

		f.Close()
	}

	return nil
}

// Stop stops all active profiling
func (p *Profiler) Stop() error {
	var errs []error

	if p.cpuFile != nil {
		if err := p.StopCPUProfile(); err != nil {
			errs = append(errs, err)
		}
	}

	if p.traceFile != nil {
		if err := p.StopTrace(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors stopping profiler: %v", errs)
	}

	return nil
}

// MemStats returns current memory statistics
func MemStats() *runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return &m
}

// PrintMemStats prints formatted memory statistics
func PrintMemStats() {
	m := MemStats()
	fmt.Printf("Memory Statistics:\n")
	fmt.Printf("  Alloc      = %v MB\n", bToMb(m.Alloc))
	fmt.Printf("  TotalAlloc = %v MB\n", bToMb(m.TotalAlloc))
	fmt.Printf("  Sys        = %v MB\n", bToMb(m.Sys))
	fmt.Printf("  NumGC      = %v\n", m.NumGC)
	fmt.Printf("  Goroutines = %v\n", runtime.NumGoroutine())
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// GoroutineCount returns the current number of goroutines
func GoroutineCount() int {
	return runtime.NumGoroutine()
}
