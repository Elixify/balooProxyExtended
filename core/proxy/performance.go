package proxy

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/shirou/gopsutil/mem"
)

// PerformanceTuning holds calculated performance settings
type PerformanceTuning struct {
	GOMAXPROCS           int
	GCPercent            int
	MemoryLimitMB        int64
	MaxIdleConns         int
	MaxIdleConnsPerHost  int
	MaxConnsPerHost      int
	HTTP2MaxStreams      uint32
	RecommendedRPS       int
}

// CalculateOptimalSettings auto-tunes based on CPU cores and RAM
func CalculateOptimalSettings(cpuCores int, ramMB int) PerformanceTuning {
	tuning := PerformanceTuning{}

	// Auto-detect if not provided
	if cpuCores == 0 {
		cpuCores = runtime.NumCPU()
	}
	if ramMB == 0 {
		if vmem, err := mem.VirtualMemory(); err == nil {
			ramMB = int(vmem.Total / 1024 / 1024)
		} else {
			ramMB = 2048 // Default to 2GB if detection fails
		}
	}

	fmt.Printf("[ Performance Tuning ] Detected: %d CPU cores, %d MB RAM\n", cpuCores, ramMB)

	// GOMAXPROCS: Use all cores, but cap at 8 for efficiency
	tuning.GOMAXPROCS = cpuCores
	if tuning.GOMAXPROCS > 8 {
		tuning.GOMAXPROCS = 8
	}

	// GC tuning based on RAM
	if ramMB <= 2048 {
		// Low memory: aggressive GC
		tuning.GCPercent = 50
		tuning.MemoryLimitMB = int64(ramMB * 75 / 100) // Use 75% of RAM
	} else if ramMB <= 4096 {
		// Medium memory: balanced GC
		tuning.GCPercent = 75
		tuning.MemoryLimitMB = int64(ramMB * 80 / 100) // Use 80% of RAM
	} else {
		// High memory: relaxed GC
		tuning.GCPercent = 100
		tuning.MemoryLimitMB = int64(ramMB * 85 / 100) // Use 85% of RAM
	}

	// Connection pool sizing based on cores and RAM
	if cpuCores == 1 && ramMB <= 2048 {
		// Single core, low memory (your case)
		tuning.MaxIdleConns = 50
		tuning.MaxIdleConnsPerHost = 10
		tuning.MaxConnsPerHost = 100
		tuning.HTTP2MaxStreams = 250
		tuning.RecommendedRPS = 5000
	} else if cpuCores <= 2 && ramMB <= 4096 {
		// Dual core, medium memory
		tuning.MaxIdleConns = 100
		tuning.MaxIdleConnsPerHost = 20
		tuning.MaxConnsPerHost = 200
		tuning.HTTP2MaxStreams = 500
		tuning.RecommendedRPS = 10000
	} else if cpuCores <= 4 && ramMB <= 8192 {
		// Quad core, good memory
		tuning.MaxIdleConns = 200
		tuning.MaxIdleConnsPerHost = 40
		tuning.MaxConnsPerHost = 500
		tuning.HTTP2MaxStreams = 1000
		tuning.RecommendedRPS = 20000
	} else {
		// High-end system
		tuning.MaxIdleConns = 500
		tuning.MaxIdleConnsPerHost = 100
		tuning.MaxConnsPerHost = 1000
		tuning.HTTP2MaxStreams = 2000
		tuning.RecommendedRPS = 50000
	}

	return tuning
}

// ApplySettings applies the calculated tuning to the runtime
func (pt *PerformanceTuning) ApplySettings() {
	runtime.GOMAXPROCS(pt.GOMAXPROCS)
	debug.SetGCPercent(pt.GCPercent)
	debug.SetMemoryLimit(pt.MemoryLimitMB * 1024 * 1024)

	fmt.Printf("[ Performance Tuning ] Applied Settings:\n")
	fmt.Printf("  - GOMAXPROCS: %d\n", pt.GOMAXPROCS)
	fmt.Printf("  - GC Percent: %d%%\n", pt.GCPercent)
	fmt.Printf("  - Memory Limit: %d MB\n", pt.MemoryLimitMB)
	fmt.Printf("  - Max Idle Connections: %d\n", pt.MaxIdleConns)
	fmt.Printf("  - Max Idle Connections Per Host: %d\n", pt.MaxIdleConnsPerHost)
	fmt.Printf("  - Max Connections Per Host: %d\n", pt.MaxConnsPerHost)
	fmt.Printf("  - HTTP/2 Max Streams: %d\n", pt.HTTP2MaxStreams)
	fmt.Printf("  - Recommended Max RPS: ~%d req/s\n", pt.RecommendedRPS)
	fmt.Println()
}

// Global tuning instance
var CurrentTuning PerformanceTuning
