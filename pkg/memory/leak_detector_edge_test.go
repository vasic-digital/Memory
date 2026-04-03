package memory_test

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"digital.vasic.memory/pkg/memory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLeakDetector_ConcurrentStartStop(t *testing.T) {
	t.Parallel()

	d := memory.NewLeakDetector(50*time.Millisecond, 2.0)
	ctx := context.Background()

	err := d.Start(ctx)
	require.NoError(t, err)

	var wg sync.WaitGroup
	// Concurrent GetReport calls while running
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = d.GetReport()
		}()
	}

	// Concurrent GetSamples calls while running
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = d.GetSamples()
		}()
	}

	wg.Wait()
	d.Stop()
}

func TestLeakDetector_FalsePositive_HeapOnly(t *testing.T) {
	t.Parallel()

	// Use a very high threshold so normal memory fluctuations do not trigger
	// the heap growth check. Note: GoroutineGrowthRate is unpredictable in
	// test environments (test runner goroutines inflate the count), so we
	// only verify that the heap growth ratio stays under the threshold.
	d := memory.NewLeakDetector(50*time.Millisecond, 100.0)
	ctx := context.Background()

	err := d.Start(ctx)
	require.NoError(t, err)

	// Wait for a couple of samples without allocating significant memory
	time.Sleep(200 * time.Millisecond)

	report := d.GetReport()
	d.Stop()

	// With 100x threshold, heap growth alone should not trigger
	assert.Less(t, report.HeapGrowthRatio, 100.0,
		"heap growth ratio should stay under 100x threshold in normal conditions")
}

func TestLeakDetector_TrackingAfterReset(t *testing.T) {
	t.Parallel()

	d := memory.NewLeakDetector(50*time.Millisecond, 2.0)
	ctx := context.Background()

	// First cycle: start, collect, stop
	err := d.Start(ctx)
	require.NoError(t, err)
	time.Sleep(120 * time.Millisecond)
	d.Stop()

	samplesAfterFirstCycle := d.GetSamples()
	assert.NotEmpty(t, samplesAfterFirstCycle)

	// Second cycle: restart should work and reset samples for the new baseline
	err = d.Start(ctx)
	require.NoError(t, err)

	// The baseline sample from the new Start should be present
	samples := d.GetSamples()
	assert.NotEmpty(t, samples, "restart should have at least a baseline sample")

	time.Sleep(120 * time.Millisecond)
	d.Stop()

	samplesAfterSecondCycle := d.GetSamples()
	assert.GreaterOrEqual(t, len(samplesAfterSecondCycle), 2,
		"second cycle should accumulate new samples")
}

func TestLeakDetector_EmptyTrackerState(t *testing.T) {
	t.Parallel()

	d := memory.NewLeakDetector(time.Hour, 2.0)

	// Before starting: no samples, report still works
	samples := d.GetSamples()
	assert.Empty(t, samples, "new detector should have no samples")

	report := d.GetReport()
	assert.False(t, report.Timestamp.IsZero(), "report timestamp should be set even before start")
	assert.Greater(t, report.GoroutineCount, 0, "goroutine count should be positive")
	assert.Greater(t, report.HeapAlloc, uint64(0), "HeapAlloc should be positive")
}

func TestLeakDetector_VeryLargeAllocationTracking(t *testing.T) {
	t.Parallel()

	d := memory.NewLeakDetector(50*time.Millisecond, 1.5)
	ctx := context.Background()

	err := d.Start(ctx)
	require.NoError(t, err)

	// Wait for baseline
	time.Sleep(80 * time.Millisecond)

	// Allocate a large chunk of memory
	var waste [][]byte
	for i := 0; i < 50; i++ {
		waste = append(waste, make([]byte, 1024*1024)) // 50 MB total
	}

	// Wait for a sample to be collected after the allocation
	time.Sleep(120 * time.Millisecond)

	report := d.GetReport()
	d.Stop()

	// The heap should have grown significantly
	assert.Greater(t, report.HeapAlloc, uint64(10*1024*1024),
		"HeapAlloc should reflect large allocations")

	// Keep waste alive past assertions
	runtime.KeepAlive(waste)
}

func TestMemoryMonitor_ConcurrentReportReading(t *testing.T) {
	t.Parallel()

	m := memory.NewMemoryMonitor(50*time.Millisecond, 2.0)
	ctx := context.Background()

	err := m.Start(ctx)
	require.NoError(t, err)

	var wg sync.WaitGroup
	reportsCh := m.Reports()

	// Multiple goroutines reading from the reports channel concurrently
	var collected []memory.LeakReport
	var mu sync.Mutex

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 3; j++ {
				select {
				case r, ok := <-reportsCh:
					if ok {
						mu.Lock()
						collected = append(collected, r)
						mu.Unlock()
					}
				case <-time.After(500 * time.Millisecond):
					return
				}
			}
		}()
	}

	wg.Wait()
	m.Stop()

	mu.Lock()
	defer mu.Unlock()
	// At least some reports should have been collected across goroutines
	assert.NotEmpty(t, collected, "concurrent readers should have received reports")
}

func TestMemoryMonitor_DoubleStop(t *testing.T) {
	t.Parallel()

	m := memory.NewMemoryMonitor(50*time.Millisecond, 2.0)
	ctx := context.Background()

	err := m.Start(ctx)
	require.NoError(t, err)
	time.Sleep(80 * time.Millisecond)

	// Double stop should not panic
	assert.NotPanics(t, func() {
		m.Stop()
		m.Stop()
	})
}

func TestGetCurrentMemoryUsage_NonEmpty(t *testing.T) {
	t.Parallel()

	usage := memory.GetCurrentMemoryUsage()
	assert.NotEmpty(t, usage)
	assert.Contains(t, usage, "HeapAlloc:")
	assert.Contains(t, usage, "Goroutines:")
}

func TestForceGC_DoesNotPanic(t *testing.T) {
	t.Parallel()

	assert.NotPanics(t, func() {
		memory.ForceGC()
	})
}

func TestLeakDetector_SamplesCapped(t *testing.T) {
	t.Parallel()

	// Very fast interval to generate many samples quickly
	d := memory.NewLeakDetector(5*time.Millisecond, 2.0)
	ctx := context.Background()

	err := d.Start(ctx)
	require.NoError(t, err)

	// Wait long enough for > 100 samples (100 * 5ms = 500ms, give extra)
	time.Sleep(800 * time.Millisecond)

	d.Stop()

	samples := d.GetSamples()
	// Internal cap is 100 samples (plus baseline), should not exceed ~101
	assert.LessOrEqual(t, len(samples), 101,
		"samples should be capped near 100")
}
