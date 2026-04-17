package notnet

import (
	"runtime"
	"sync"
	"time"
)

// StatsCollector collects server metrics
type StatsCollector struct {
	mu               sync.RWMutex
	startTime        time.Time
	requestCount     uint64
	requestHistory   []RequestMetric
	maxHistorySize   int
	registeredRoutes []string
	cpuMetrics       []CPUMetric
	memoryMetrics    []MemoryMetric
	lastCPUTime      uint64
	lastCPUTimestamp time.Time
}

// RequestMetric stores request statistics
type RequestMetric struct {
	Timestamp   time.Time     `json:"timestamp"`
	Count       uint64        `json:"count"`
	AvgDuration time.Duration `json:"avg_duration"`
}

// CPUMetric stores CPU usage metrics
type CPUMetric struct {
	Timestamp time.Time `json:"timestamp"`
	Usage     float64   `json:"usage"`
}

// MemoryMetric stores memory usage metrics
type MemoryMetric struct {
	Timestamp  time.Time `json:"timestamp"`
	Alloc      uint64    `json:"alloc"`
	TotalAlloc uint64    `json:"total_alloc"`
	Sys        uint64    `json:"sys"`
	NumGC      uint32    `json:"num_gc"`
}

// StatsData represents the current server statistics
type StatsData struct {
	UpTime           time.Duration   `json:"uptime"`
	RequestCount     uint64          `json:"request_count"`
	RegisteredRoutes []string        `json:"registered_routes"`
	Memory           *MemoryMetric   `json:"memory"`
	RequestHistory   []RequestMetric `json:"request_history"`
	MemoryHistory    []MemoryMetric  `json:"memory_history"`
	CPUHistory       []CPUMetric     `json:"cpu_history"`
	Goroutines       int             `json:"goroutines"`
}

var globalStatsCollector *StatsCollector

// InitStatsCollector initializes the global stats collector
func InitStatsCollector() *StatsCollector {
	globalStatsCollector = &StatsCollector{
		startTime:      time.Now(),
		requestHistory: make([]RequestMetric, 0, 1440), // 24 hours of 1-minute intervals
		maxHistorySize: 1440,
		cpuMetrics:     make([]CPUMetric, 0, 1440),
		memoryMetrics:  make([]MemoryMetric, 0, 1440),
	}

	// Collect initial metrics immediately
	globalStatsCollector.CollectMetrics()

	// Start collecting metrics every 10 seconds
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for range ticker.C {
			globalStatsCollector.CollectMetrics()
		}
	}()

	return globalStatsCollector
}

// GetStatsCollector returns the global stats collector
func GetStatsCollector() *StatsCollector {
	if globalStatsCollector == nil {
		InitStatsCollector()
	}
	return globalStatsCollector
}

// RecordRequest increments the request count
func (sc *StatsCollector) RecordRequest(duration time.Duration) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.requestCount++
}

// RegisterRoute adds a route to the registered routes list
func (sc *StatsCollector) RegisterRoute(method, path string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	routePath := method + " " + path
	for _, r := range sc.registeredRoutes {
		if r == routePath {
			return // Already registered
		}
	}
	sc.registeredRoutes = append(sc.registeredRoutes, routePath)
}

// CollectMetrics collects current server metrics
func (sc *StatsCollector) CollectMetrics() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Memory metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memMetric := MemoryMetric{
		Timestamp:  time.Now(),
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Sys:        m.Sys,
		NumGC:      m.NumGC,
	}
	sc.memoryMetrics = append(sc.memoryMetrics, memMetric)

	// Keep only last N metrics
	if len(sc.memoryMetrics) > sc.maxHistorySize {
		sc.memoryMetrics = sc.memoryMetrics[len(sc.memoryMetrics)-sc.maxHistorySize:]
	}

	// Request history (record every minute)
	if len(sc.requestHistory) == 0 || time.Since(sc.requestHistory[len(sc.requestHistory)-1].Timestamp) >= time.Minute {
		sc.requestHistory = append(sc.requestHistory, RequestMetric{
			Timestamp: time.Now(),
			Count:     sc.requestCount,
		})

		if len(sc.requestHistory) > sc.maxHistorySize {
			sc.requestHistory = sc.requestHistory[len(sc.requestHistory)-sc.maxHistorySize:]
		}
	}

	// CPU metrics (simplified - uses goroutine count as proxy)
	cpuMetric := CPUMetric{
		Timestamp: time.Now(),
		Usage:     float64(runtime.NumGoroutine()) / 10.0, // Simplified metric
	}
	sc.cpuMetrics = append(sc.cpuMetrics, cpuMetric)

	if len(sc.cpuMetrics) > sc.maxHistorySize {
		sc.cpuMetrics = sc.cpuMetrics[len(sc.cpuMetrics)-sc.maxHistorySize:]
	}
}

// GetStats returns the current server statistics
func (sc *StatsCollector) GetStats() *StatsData {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memMetric := &MemoryMetric{
		Timestamp:  time.Now(),
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Sys:        m.Sys,
		NumGC:      m.NumGC,
	}

	// Make copies of history slices
	reqHistory := make([]RequestMetric, len(sc.requestHistory))
	copy(reqHistory, sc.requestHistory)

	memHistory := make([]MemoryMetric, len(sc.memoryMetrics))
	copy(memHistory, sc.memoryMetrics)

	cpuHistory := make([]CPUMetric, len(sc.cpuMetrics))
	copy(cpuHistory, sc.cpuMetrics)

	routes := make([]string, len(sc.registeredRoutes))
	copy(routes, sc.registeredRoutes)

	return &StatsData{
		UpTime:           time.Since(sc.startTime),
		RequestCount:     sc.requestCount,
		RegisteredRoutes: routes,
		Memory:           memMetric,
		RequestHistory:   reqHistory,
		MemoryHistory:    memHistory,
		CPUHistory:       cpuHistory,
		Goroutines:       runtime.NumGoroutine(),
	}
}
