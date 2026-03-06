# Lesson 4: Runtime Memory Monitoring

## Objectives

- Use `LeakDetector` to sample heap and goroutine statistics
- Set up `MemoryMonitor` with alert callbacks
- Write heap and goroutine profiles for offline analysis

## Concepts

### LeakDetector

`LeakDetector` samples `runtime.MemStats` at a configurable interval and compares against a baseline. It flags a potential leak when heap growth exceeds the threshold ratio or goroutine count grows by more than 50%.

```go
detector := memory.NewLeakDetector(5*time.Second, 2.0)
detector.Start(ctx)
defer detector.Stop()

report := detector.GetReport()
// report.PotentialLeak, report.HeapGrowthRatio, etc.
```

### MemoryMonitor

`MemoryMonitor` wraps `LeakDetector` and adds a report channel and alert callback:

```go
monitor := memory.NewMemoryMonitor(5*time.Second, 2.0)
monitor.SetAlertCallback(func(r memory.LeakReport) {
    log.Printf("Potential leak: heap ratio %.2f", r.HeapGrowthRatio)
})
monitor.Start(ctx)
```

Reports are available via `monitor.Reports()`.

### Profiling Utilities

Write pprof-compatible profiles to disk:

```go
memory.WriteHeapProfile("heap.prof")
memory.WriteGoroutineProfile("goroutine.prof")
```

Force garbage collection to get accurate readings:

```go
memory.ForceGC()
```

Get a snapshot string of current usage:

```go
fmt.Println(memory.GetCurrentMemoryUsage())
// HeapAlloc: 12 MB, HeapSys: 64 MB, ...
```

## Code Walkthrough

### Continuous monitoring with alerts

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

monitor := memory.NewMemoryMonitor(2*time.Second, 1.5)
monitor.SetAlertCallback(func(r memory.LeakReport) {
    fmt.Printf("ALERT at %s: heap ratio %.2f, goroutines %d\n",
        r.Timestamp.Format(time.RFC3339),
        r.HeapGrowthRatio,
        r.GoroutineCount,
    )
    memory.WriteHeapProfile("leak-" + r.Timestamp.Format("150405") + ".prof")
})

monitor.Start(ctx)
defer monitor.Stop()

// Application runs...
```

### Analyzing the LeakReport

Key fields in `LeakReport`:

| Field | Meaning |
|-------|---------|
| `HeapAlloc` | Bytes allocated and in use |
| `HeapGrowthRatio` | Current heap / baseline heap |
| `PotentialLeak` | True if growth or goroutine ratio exceeds thresholds |
| `GoroutineCount` | Current number of goroutines |
| `GoroutineGrowthRate` | (current - initial) / initial |

## Practice Exercise

1. Create a `LeakDetector` with a 1-second sample interval and 1.5x threshold. Allocate a growing slice in a loop. Check the report and verify `PotentialLeak` becomes true when heap growth exceeds the threshold.
2. Set up a `MemoryMonitor` with an alert callback that records alerts to a slice. Simulate a goroutine leak by starting goroutines without stopping them. Verify the alert fires when goroutine growth exceeds 50%.
3. Use `WriteHeapProfile` and `WriteGoroutineProfile` to capture profiles before and after a workload. Open them with `go tool pprof` and identify the allocation hot spots.
