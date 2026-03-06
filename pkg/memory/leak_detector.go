package memory

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

type LeakDetector struct {
	mu             sync.Mutex
	wg             sync.WaitGroup
	baselineStats  runtime.MemStats
	samples        []runtime.MemStats
	interval       time.Duration
	thresholdRatio float64
	stopCh         chan struct{}
	running        bool
}

type LeakReport struct {
	Timestamp           time.Time
	HeapAlloc           uint64
	HeapSys             uint64
	HeapInUse           uint64
	HeapObjects         uint64
	StackInUse          uint64
	GoroutineCount      int
	GCCount             uint32
	HeapGrowthRatio     float64
	PotentialLeak       bool
	GoroutineGrowthRate float64
	InitialGoroutines   int
}

func NewLeakDetector(interval time.Duration, thresholdRatio float64) *LeakDetector {
	return &LeakDetector{
		interval:       interval,
		thresholdRatio: thresholdRatio,
		samples:        make([]runtime.MemStats, 0),
		stopCh:         make(chan struct{}),
	}
}

func (d *LeakDetector) Start(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running {
		return fmt.Errorf("leak detector already running")
	}

	runtime.ReadMemStats(&d.baselineStats)
	d.samples = append(d.samples, d.baselineStats)
	d.running = true
	d.stopCh = make(chan struct{})

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.monitorLoop(ctx)
	}()

	return nil
}

func (d *LeakDetector) Stop() {
	d.mu.Lock()
	if !d.running {
		d.mu.Unlock()
		return
	}

	close(d.stopCh)
	d.running = false
	d.mu.Unlock()

	d.wg.Wait()
}

func (d *LeakDetector) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	initialGoroutines := runtime.NumGoroutine()

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.collectSample(initialGoroutines)
		}
	}
}

func (d *LeakDetector) collectSample(initialGoroutines int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	d.samples = append(d.samples, stats)

	if len(d.samples) > 100 {
		d.samples = d.samples[1:]
	}
}

func (d *LeakDetector) GetReport() LeakReport {
	d.mu.Lock()
	defer d.mu.Unlock()

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	initialGoroutines := runtime.NumGoroutine()
	if len(d.samples) > 0 {
		initialGoroutines = int(d.samples[0].NumGC) + 1
	}

	heapGrowthRatio := 0.0
	if d.baselineStats.HeapAlloc > 0 {
		heapGrowthRatio = float64(stats.HeapAlloc) / float64(d.baselineStats.HeapAlloc)
	}

	currentGoroutines := runtime.NumGoroutine()
	goroutineGrowthRate := 0.0
	if initialGoroutines > 0 {
		goroutineGrowthRate = float64(currentGoroutines-initialGoroutines) / float64(initialGoroutines)
	}

	potentialLeak := heapGrowthRatio > d.thresholdRatio || goroutineGrowthRate > 0.5

	return LeakReport{
		Timestamp:           time.Now(),
		HeapAlloc:           stats.HeapAlloc,
		HeapSys:             stats.HeapSys,
		HeapInUse:           stats.HeapInuse,
		HeapObjects:         stats.HeapObjects,
		StackInUse:          stats.StackInuse,
		GoroutineCount:      currentGoroutines,
		GCCount:             stats.NumGC,
		HeapGrowthRatio:     heapGrowthRatio,
		PotentialLeak:       potentialLeak,
		GoroutineGrowthRate: goroutineGrowthRate,
		InitialGoroutines:   initialGoroutines,
	}
}

func (d *LeakDetector) GetSamples() []runtime.MemStats {
	d.mu.Lock()
	defer d.mu.Unlock()

	result := make([]runtime.MemStats, len(d.samples))
	copy(result, d.samples)
	return result
}

func WriteHeapProfile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create heap profile file: %w", err)
	}
	defer f.Close()

	runtime.GC()

	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("could not write heap profile: %w", err)
	}

	return nil
}

func WriteGoroutineProfile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create goroutine profile file: %w", err)
	}
	defer f.Close()

	p := pprof.Lookup("goroutine")
	if p == nil {
		return fmt.Errorf("goroutine profile not found")
	}

	if err := p.WriteTo(f, 0); err != nil {
		return fmt.Errorf("could not write goroutine profile: %w", err)
	}

	return nil
}

func ForceGC() {
	runtime.GC()
	runtime.GC()
}

type MemoryMonitor struct {
	detector      *LeakDetector
	wg            sync.WaitGroup
	reportCh      chan LeakReport
	alertCallback func(LeakReport)
}

func NewMemoryMonitor(interval time.Duration, thresholdRatio float64) *MemoryMonitor {
	return &MemoryMonitor{
		detector: NewLeakDetector(interval, thresholdRatio),
		reportCh: make(chan LeakReport, 10),
	}
}

func (m *MemoryMonitor) SetAlertCallback(cb func(LeakReport)) {
	m.alertCallback = cb
}

func (m *MemoryMonitor) Start(ctx context.Context) error {
	if err := m.detector.Start(ctx); err != nil {
		return err
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.monitorReports(ctx)
	}()

	return nil
}

func (m *MemoryMonitor) Stop() {
	m.detector.Stop()
	m.wg.Wait()
}

func (m *MemoryMonitor) monitorReports(ctx context.Context) {
	ticker := time.NewTicker(m.detector.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.detector.stopCh:
			return
		case <-ticker.C:
			report := m.detector.GetReport()
			select {
			case m.reportCh <- report:
			default:
			}

			if report.PotentialLeak && m.alertCallback != nil {
				m.alertCallback(report)
			}
		}
	}
}

func (m *MemoryMonitor) Reports() <-chan LeakReport {
	return m.reportCh
}

func GetCurrentMemoryUsage() string {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	return fmt.Sprintf(
		"HeapAlloc: %d MB, HeapSys: %d MB, HeapInuse: %d MB, HeapObjects: %d, Goroutines: %d",
		stats.HeapAlloc/1024/1024,
		stats.HeapSys/1024/1024,
		stats.HeapInuse/1024/1024,
		stats.HeapObjects,
		runtime.NumGoroutine(),
	)
}
