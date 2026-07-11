package middleware

import (
	"runtime"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
)

type MetricsCollector struct {
	mu              sync.RWMutex
	requestCount    map[string]int64
	requestDuration map[string]float64
	statusCounts    map[string]int64
	latencies       []float64
	activeConns     atomic.Int64
	startTime       time.Time
}

var metrics = &MetricsCollector{
	requestCount:    make(map[string]int64),
	requestDuration: make(map[string]float64),
	statusCounts:    make(map[string]int64),
	latencies:       make([]float64, 0, 10000),
	startTime:       time.Now(),
}

// Record updates request count, status category, duration, and latency
// in a single lock acquisition. R50-B6: previously IncRequest and
// ObserveDuration each acquired m.mu independently, doubling lock
// contention per request under high concurrency.
func (m *MetricsCollector) Record(method, path, status string, duration float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := method + " " + path + " " + status
	m.requestCount[key]++

	category := "0xx"
	if len(status) > 0 {
		category = status[:1] + "xx"
	}
	m.statusCounts[category]++

	durKey := method + " " + path
	m.requestDuration[durKey] += duration

	m.latencies = append(m.latencies, duration)
	// Trim to the most recent 10000 samples, but only when we overshoot by
	// 20% (12000). The prior code trimmed on every append past 10000,
	// causing an O(n) copy of 10000 elements on every single request. With
	// this threshold the copy amortizes to ~1 in 2000 requests.
	if len(m.latencies) > 12000 {
		newLat := make([]float64, 10000)
		copy(newLat, m.latencies[len(m.latencies)-10000:])
		m.latencies = newLat
	}
}

func (m *MetricsCollector) IncActive() {
	// R39b-M2: use atomic instead of write lock — IncActive/DecActive are
	// called on every request and contention on m.mu was visible under load.
	m.activeConns.Add(1)
}

func (m *MetricsCollector) DecActive() {
	// R39b-M2: Add(-1) is safe even if it goes negative; the Load() in
	// Snapshot never underflows int64 in practice. Avoid Lock to prevent
	// contention with the request/duration maps.
	m.activeConns.Add(-1)
}

// sortedLatencies returns a sorted copy of the latency samples. Callers must
// hold m.mu (at least RLock) when calling this.
func (m *MetricsCollector) sortedLatencies() []float64 {
	if len(m.latencies) == 0 {
		return nil
	}
	sorted := make([]float64, len(m.latencies))
	copy(sorted, m.latencies)
	sort.Float64s(sorted)
	return sorted
}

func percentileFromSorted(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}

func (m *MetricsCollector) Snapshot() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	counts := make(map[string]int64, len(m.requestCount))
	for k, v := range m.requestCount {
		counts[k] = v
	}
	durations := make(map[string]float64, len(m.requestDuration))
	for k, v := range m.requestDuration {
		durations[k] = v
	}

	var totalReqs int64
	var totalErrors int64
	statusCodes := make(map[string]int64, len(m.statusCounts))
	for cat, cnt := range m.statusCounts {
		totalReqs += cnt
		statusCodes[cat] = cnt
		if cat == "5xx" || cat == "4xx" {
			totalErrors += cnt
		}
	}

	var errorRate float64
	if totalReqs > 0 {
		errorRate = float64(totalErrors) / float64(totalReqs)
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Sort the latency samples once and reuse for all percentiles, instead
	// of sorting (O(n log n)) three separate times under the RLock.
	sorted := m.sortedLatencies()

	return map[string]interface{}{
		"uptime_seconds":       int(time.Since(m.startTime).Seconds()),
		"active_connections":   m.activeConns.Load(),
		"total_requests":       totalReqs,
		"total_errors":         totalErrors,
		"error_rate":           errorRate,
		"status_codes":         statusCodes,
		"request_count":        counts,
		"request_duration":     durations,
		"latency_p50_ms":       percentileFromSorted(sorted, 0.50) * 1000,
		"latency_p95_ms":       percentileFromSorted(sorted, 0.95) * 1000,
		"latency_p99_ms":       percentileFromSorted(sorted, 0.99) * 1000,
		"memory_alloc_mb":      float64(memStats.Alloc) / 1024 / 1024,
		"memory_sys_mb":        float64(memStats.Sys) / 1024 / 1024,
		"gc_cycles":            memStats.NumGC,
		"go_routines":          runtime.NumGoroutine(),
	}
}

func Metrics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		metrics.IncActive()
		// R49-B5: defer the metrics recording so panics in downstream
		// handlers are still counted. Without this, a panicking request
		// would increment active connections (IncActive above) but never
		// record the request count / latency / status — the DecActive
		// defer runs, but the 500 response from Sentry's recovery would
		// not be reflected in status_counts. This ensures panics show up
		// as 5xx in the metrics snapshot.
		defer func() {
			metrics.DecActive()
			duration := time.Since(start).Seconds()
			// R39b-M1: recover() checks if a panic occurred. Sentry middleware
			// (registered earlier) recovers and writes 500, but its defer runs
			// AFTER this one (LIFO), so c.Response().StatusCode() is still 200
			// here. Force 500 for metrics accuracy when a panic happened.
			status := strconv.Itoa(c.Response().StatusCode())
			// Normalize path to the matched route template (e.g. /resumes/:id)
			// so dynamic segments don't create unbounded map keys. For unmatched
			// routes (404), use a fixed "UNMATCHED" key — otherwise attackers
			// can send arbitrary paths (/foo1, /foo2, ...) to grow the maps
			// without bound.
			path := c.Path()
			if r := c.Route(); r.Path != "" {
				path = r.Path
			} else {
				path = "UNMATCHED"
			}
			method := c.Method()

			// R51b-B1: record metrics BEFORE re-panic. The previous code called
			// panic(r) before Record, so panicking requests were never
			// recorded — defeating the purpose of R49-B5's defer.
			if r := recover(); r != nil {
				status = "500"
				metrics.Record(method, path, status, duration)
				panic(r) // re-panic so Sentry middleware can capture it
			}

			metrics.Record(method, path, status, duration)
		}()

		return c.Next()
	}
}

func MetricsHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(metrics.Snapshot())
	}
}
