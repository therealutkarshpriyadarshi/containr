package benchmark

import (
	"testing"
	"time"
)

func TestBenchmark(t *testing.T) {
	bench := New("test", 10, func() error {
		time.Sleep(1 * time.Millisecond)
		return nil
	})

	result, err := bench.Run()
	if err != nil {
		t.Fatalf("benchmark failed: %v", err)
	}

	if result.Operations != 10 {
		t.Errorf("expected 10 operations, got %d", result.Operations)
	}

	if result.Duration < 10*time.Millisecond {
		t.Errorf("duration too short: %v", result.Duration)
	}
}

func TestSuite(t *testing.T) {
	suite := NewSuite()
	suite.Add("test1", 5, func() error {
		time.Sleep(1 * time.Millisecond)
		return nil
	})
	suite.Add("test2", 5, func() error {
		time.Sleep(2 * time.Millisecond)
		return nil
	})

	results, err := suite.Run()
	if err != nil {
		t.Fatalf("suite failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestTimer(t *testing.T) {
	timer := NewTimer()
	time.Sleep(10 * time.Millisecond)
	elapsed := timer.Elapsed()

	if elapsed < 10*time.Millisecond {
		t.Errorf("elapsed time too short: %v", elapsed)
	}
}

func TestMeasure(t *testing.T) {
	duration, err := Measure("test", func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	if err != nil {
		t.Fatalf("measure failed: %v", err)
	}

	if duration < 10*time.Millisecond {
		t.Errorf("duration too short: %v", duration)
	}
}

// Benchmark tests for the benchmark package itself
func BenchmarkTimer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		timer := NewTimer()
		_ = timer.Elapsed()
	}
}

func BenchmarkMeasure(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Measure("test", func() error {
			return nil
		})
	}
}
