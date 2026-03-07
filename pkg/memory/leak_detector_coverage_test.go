package memory

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLeakDetector_CollectSample_SampleTrimming exercises the branch in
// collectSample where len(d.samples) > 100, causing the oldest sample
// to be dropped (line 111).
func TestLeakDetector_CollectSample_SampleTrimming(t *testing.T) {
	d := NewLeakDetector(10*time.Millisecond, 2.0)

	// Manually fill the samples slice to exactly 100 entries.
	d.mu.Lock()
	for i := 0; i < 100; i++ {
		var stats runtime.MemStats
		stats.HeapAlloc = uint64(i) // Tag each sample for identification.
		d.samples = append(d.samples, stats)
	}
	d.mu.Unlock()

	assert.Equal(t, 100, len(d.GetSamples()))

	// Call collectSample — this should add one sample and trim the oldest.
	d.collectSample(runtime.NumGoroutine())

	samples := d.GetSamples()
	assert.Equal(t, 100, len(samples), "samples should be capped at 100")

	// The first sample (HeapAlloc==0) should have been removed.
	assert.NotEqual(t, uint64(0), samples[0].HeapAlloc,
		"oldest sample should have been trimmed")
}

// TestLeakDetector_CollectSample_BelowCap verifies that collectSample does
// not trim when samples count is at or below 100.
func TestLeakDetector_CollectSample_BelowCap(t *testing.T) {
	d := NewLeakDetector(10*time.Millisecond, 2.0)

	d.collectSample(runtime.NumGoroutine())
	assert.Equal(t, 1, len(d.GetSamples()))

	d.collectSample(runtime.NumGoroutine())
	assert.Equal(t, 2, len(d.GetSamples()))
}

// TestWriteHeapProfile_InvalidDir_Coverage tests the error path where
// the directory does not exist.
func TestWriteHeapProfile_InvalidDir_Coverage(t *testing.T) {
	err := WriteHeapProfile("/nonexistent/path/that/does/not/exist/heap.prof")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not create heap profile file")
}

// TestWriteHeapProfile_Success_Coverage confirms that a valid path writes
// a non-empty heap profile.
func TestWriteHeapProfile_Success_Coverage(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "heap_cov.prof")

	err := WriteHeapProfile(filename)
	require.NoError(t, err)

	info, err := os.Stat(filename)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

// TestWriteGoroutineProfile_InvalidDir_Coverage tests the error path where
// the directory does not exist.
func TestWriteGoroutineProfile_InvalidDir_Coverage(t *testing.T) {
	err := WriteGoroutineProfile("/nonexistent/path/goroutine.prof")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not create goroutine profile file")
}

// TestWriteGoroutineProfile_Success_Coverage confirms that a valid path
// writes a non-empty goroutine profile.
func TestWriteGoroutineProfile_Success_Coverage(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "goroutine_cov.prof")

	err := WriteGoroutineProfile(filename)
	require.NoError(t, err)

	info, err := os.Stat(filename)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

// TestMemoryMonitor_Start_DetectorAlreadyRunning_Coverage exercises the
// error path in MemoryMonitor.Start where the underlying detector is
// already running (line 224).
func TestMemoryMonitor_Start_DetectorAlreadyRunning_Coverage(t *testing.T) {
	m := NewMemoryMonitor(50*time.Millisecond, 2.0)
	ctx := context.Background()

	// Start the monitor successfully.
	err := m.Start(ctx)
	require.NoError(t, err)

	// Starting again should fail because the detector is already running.
	err = m.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	m.Stop()
}

// TestMemoryMonitor_MonitorReports_ContextCancel_Coverage exercises the
// context cancellation branch in monitorReports (line 249).
func TestMemoryMonitor_MonitorReports_ContextCancel_Coverage(t *testing.T) {
	m := NewMemoryMonitor(50*time.Millisecond, 2.0)
	ctx, cancel := context.WithCancel(context.Background())

	err := m.Start(ctx)
	require.NoError(t, err)

	// Let it run briefly.
	time.Sleep(80 * time.Millisecond)

	// Cancel context, which should cause monitorReports to exit via ctx.Done().
	cancel()

	// Stop should complete promptly.
	done := make(chan struct{})
	go func() {
		m.Stop()
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(3 * time.Second):
		t.Fatal("Stop did not complete after context cancellation")
	}
}

// TestMemoryMonitor_MonitorReports_StopCh_Coverage exercises the stopCh
// branch in monitorReports (line 250).
func TestMemoryMonitor_MonitorReports_StopCh_Coverage(t *testing.T) {
	m := NewMemoryMonitor(50*time.Millisecond, 2.0)
	ctx := context.Background()

	err := m.Start(ctx)
	require.NoError(t, err)

	// Let it run briefly.
	time.Sleep(80 * time.Millisecond)

	// Stop should trigger the stopCh path.
	m.Stop()
}

// TestMemoryMonitor_MonitorReports_ChannelFull_Coverage exercises the
// default branch in the select when reportCh is full (lines 256-257).
func TestMemoryMonitor_MonitorReports_ChannelFull_Coverage(t *testing.T) {
	m := NewMemoryMonitor(20*time.Millisecond, 2.0)
	ctx := context.Background()

	err := m.Start(ctx)
	require.NoError(t, err)

	// Do NOT read from the channel so it fills up, triggering the default branch.
	time.Sleep(400 * time.Millisecond)

	m.Stop()

	// Drain and count reports. The channel capacity is 10, so at most 10 reports.
	count := 0
	for {
		select {
		case <-m.Reports():
			count++
		default:
			goto done
		}
	}
done:
	assert.LessOrEqual(t, count, 10, "channel should be capped at capacity 10")
}

// TestMemoryMonitor_NoAlertCallback_Coverage exercises the branch in
// monitorReports where alertCallback is nil (line 259 condition false).
func TestMemoryMonitor_NoAlertCallback_Coverage(t *testing.T) {
	m := NewMemoryMonitor(50*time.Millisecond, 0.0001) // Low threshold.
	// Do NOT set alert callback.

	ctx := context.Background()
	err := m.Start(ctx)
	require.NoError(t, err)

	time.Sleep(150 * time.Millisecond)

	// Should not panic even if potentialLeak is true but callback is nil.
	m.Stop()
}

// TestLeakDetector_GetReport_ZeroBaselineHeapAlloc_Coverage exercises the
// branch in GetReport where baselineStats.HeapAlloc is 0 (line 128).
func TestLeakDetector_GetReport_ZeroBaselineHeapAlloc_Coverage(t *testing.T) {
	d := NewLeakDetector(50*time.Millisecond, 2.0)

	// Force baselineStats.HeapAlloc to 0 so the zero-division guard is triggered.
	d.mu.Lock()
	d.baselineStats.HeapAlloc = 0
	d.mu.Unlock()

	report := d.GetReport()
	assert.Equal(t, 0.0, report.HeapGrowthRatio,
		"heapGrowthRatio should be 0 when baseline HeapAlloc is 0")
}

// TestLeakDetector_MonitorLoop_MultipleIntervals_Coverage verifies that
// the monitor loop collects multiple samples over several tick intervals.
func TestLeakDetector_MonitorLoop_MultipleIntervals_Coverage(t *testing.T) {
	d := NewLeakDetector(30*time.Millisecond, 2.0)
	ctx := context.Background()

	err := d.Start(ctx)
	require.NoError(t, err)

	// Wait for multiple intervals.
	time.Sleep(200 * time.Millisecond)
	d.Stop()

	samples := d.GetSamples()
	assert.GreaterOrEqual(t, len(samples), 3,
		"should have collected multiple samples")
}

// TestMemoryMonitor_AlertCallback_ConcurrentAccess_Coverage exercises the
// alert callback under concurrent conditions.
func TestMemoryMonitor_AlertCallback_ConcurrentAccess_Coverage(t *testing.T) {
	m := NewMemoryMonitor(20*time.Millisecond, 0.0001)

	var mu sync.Mutex
	var alerts []LeakReport
	m.SetAlertCallback(func(report LeakReport) {
		mu.Lock()
		defer mu.Unlock()
		alerts = append(alerts, report)
	})

	ctx := context.Background()
	err := m.Start(ctx)
	require.NoError(t, err)

	// Allocate memory to trigger growth detection.
	waste := make([][]byte, 50)
	for i := range waste {
		waste[i] = make([]byte, 4096)
	}

	time.Sleep(200 * time.Millisecond)
	m.Stop()

	mu.Lock()
	alertCount := len(alerts)
	mu.Unlock()

	// Alert callback should have been invoked at least once due to low threshold.
	assert.Greater(t, alertCount, 0)
	_ = waste // Keep alive.
}
