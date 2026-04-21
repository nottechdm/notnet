package notnet

import (
	"testing"
	"time"
)

func TestStatsCollector(t *testing.T) {
	sc := InitStatsCollector()
	if sc == nil {
		t.Fatal("expected stats collector to be initialized")
	}

	// Test RegisterRoute
	sc.RegisterRoute("GET", "/test")
	sc.RegisterRoute("GET", "/test") // Duplicate
	sc.RegisterRoute("POST", "/data")

	stats := sc.GetStats()
	if len(stats.RegisteredRoutes) != 2 {
		t.Errorf("expected 2 registered routes, got %d", len(stats.RegisteredRoutes))
	}

	// Test RecordRequest
	sc.RecordRequest(100 * time.Millisecond)
	sc.RecordRequest(200 * time.Millisecond)

	stats = sc.GetStats()
	if stats.RequestCount != 2 {
		t.Errorf("expected request count 2, got %d", stats.RequestCount)
	}

	// Test CollectMetrics
	sc.CollectMetrics()
	stats = sc.GetStats()

	if len(stats.MemoryHistory) == 0 {
		t.Error("expected memory history to be populated")
	}
	if len(stats.CPUHistory) == 0 {
		t.Error("expected CPU history to be populated")
	}
}

func TestGetStatsCollector(t *testing.T) {
	sc1 := GetStatsCollector()
	sc2 := GetStatsCollector()
	if sc1 != sc2 {
		t.Error("expected GetStatsCollector to return the same instance")
	}
}

func TestStatsCollector_MaxHistorySize(t *testing.T) {
	sc := &StatsCollector{
		maxHistorySize: 2,
		requestHistory: make([]RequestMetric, 0),
		memoryMetrics:  make([]MemoryMetric, 0),
		cpuMetrics:     make([]CPUMetric, 0),
	}

	// Test memory history trimming
	for i := 0; i < 5; i++ {
		sc.memoryMetrics = append(sc.memoryMetrics, MemoryMetric{})
		sc.cpuMetrics = append(sc.cpuMetrics, CPUMetric{})
	}

	sc.CollectMetrics()

	stats := sc.GetStats()
	if len(stats.MemoryHistory) > 2 {
		t.Errorf("expected memory history size <= 2, got %d", len(stats.MemoryHistory))
	}
	if len(stats.CPUHistory) > 2 {
		t.Errorf("expected cpu history size <= 2, got %d", len(stats.CPUHistory))
	}
}
