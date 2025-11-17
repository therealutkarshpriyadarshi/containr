// Package benchmark provides performance benchmarking utilities for containr
package benchmark

import (
	"fmt"
	"runtime"
	"time"
)

// Result represents a benchmark result
type Result struct {
	Name          string
	Duration      time.Duration
	Operations    int64
	BytesPerOp    int64
	AllocsPerOp   int64
	MemAllocBytes int64
}

// Benchmark represents a benchmark test
type Benchmark struct {
	name       string
	iterations int
	fn         func() error
}

// New creates a new benchmark
func New(name string, iterations int, fn func() error) *Benchmark {
	return &Benchmark{
		name:       name,
		iterations: iterations,
		fn:         fn,
	}
}

// Run executes the benchmark and returns results
func (b *Benchmark) Run() (*Result, error) {
	// Warm up
	if err := b.fn(); err != nil {
		return nil, fmt.Errorf("warmup failed: %w", err)
	}

	// Force GC before benchmark
	runtime.GC()

	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	start := time.Now()
	for i := 0; i < b.iterations; i++ {
		if err := b.fn(); err != nil {
			return nil, fmt.Errorf("iteration %d failed: %w", i, err)
		}
	}
	duration := time.Since(start)

	runtime.ReadMemStats(&memStatsAfter)

	result := &Result{
		Name:          b.name,
		Duration:      duration,
		Operations:    int64(b.iterations),
		AllocsPerOp:   int64(memStatsAfter.Mallocs-memStatsBefore.Mallocs) / int64(b.iterations),
		MemAllocBytes: int64(memStatsAfter.Alloc - memStatsBefore.Alloc),
	}

	if result.Operations > 0 {
		result.BytesPerOp = result.MemAllocBytes / result.Operations
	}

	return result, nil
}

// String returns a formatted string representation of the result
func (r *Result) String() string {
	nsPerOp := r.Duration.Nanoseconds() / r.Operations
	return fmt.Sprintf("%s\t%d iterations\t%d ns/op\t%d B/op\t%d allocs/op",
		r.Name, r.Operations, nsPerOp, r.BytesPerOp, r.AllocsPerOp)
}

// Suite represents a collection of benchmarks
type Suite struct {
	benchmarks []*Benchmark
}

// NewSuite creates a new benchmark suite
func NewSuite() *Suite {
	return &Suite{
		benchmarks: make([]*Benchmark, 0),
	}
}

// Add adds a benchmark to the suite
func (s *Suite) Add(name string, iterations int, fn func() error) {
	s.benchmarks = append(s.benchmarks, New(name, iterations, fn))
}

// Run runs all benchmarks in the suite
func (s *Suite) Run() ([]*Result, error) {
	results := make([]*Result, 0, len(s.benchmarks))

	for _, bench := range s.benchmarks {
		result, err := bench.Run()
		if err != nil {
			return nil, fmt.Errorf("benchmark %s failed: %w", bench.name, err)
		}
		results = append(results, result)
	}

	return results, nil
}

// Timer provides simple timing functionality
type Timer struct {
	start time.Time
}

// NewTimer creates and starts a new timer
func NewTimer() *Timer {
	return &Timer{start: time.Now()}
}

// Elapsed returns the elapsed time since timer creation
func (t *Timer) Elapsed() time.Duration {
	return time.Since(t.start)
}

// Reset resets the timer
func (t *Timer) Reset() {
	t.start = time.Now()
}

// Measure measures the execution time of a function
func Measure(name string, fn func() error) (time.Duration, error) {
	timer := NewTimer()
	err := fn()
	duration := timer.Elapsed()

	if err != nil {
		return duration, err
	}

	return duration, nil
}
