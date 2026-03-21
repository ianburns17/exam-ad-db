package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type metricsData struct {
	TotalRequests    int64
	RequestsPerRoute map[string]int64
	ErrorCounts      map[string]int64
	TotalLatency     map[string]time.Duration
	mu               sync.Mutex
}

var appMetrics = &metricsData{
	RequestsPerRoute: make(map[string]int64),
	ErrorCounts:      make(map[string]int64),
	TotalLatency:     make(map[string]time.Duration),
}

// metricsMiddleware tracks requests, errors, and latency
func (a *applicationDependencies) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		appMetrics.mu.Lock()
		appMetrics.TotalRequests++
		appMetrics.RequestsPerRoute[r.URL.Path]++
		appMetrics.mu.Unlock()

		rw := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)

		latency := time.Since(start)
		appMetrics.mu.Lock()
		appMetrics.TotalLatency[r.URL.Path] += latency
		if rw.status >= 400 {
			appMetrics.ErrorCounts[r.URL.Path]++
		}
		appMetrics.mu.Unlock()
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// metricsHandler exposes metrics at /metrics
func (a *applicationDependencies) metricsHandler(w http.ResponseWriter, r *http.Request) {
	appMetrics.mu.Lock()
	defer appMetrics.mu.Unlock()
	fmt.Fprintf(w, "total_requests %d\n", appMetrics.TotalRequests)
	for route, count := range appMetrics.RequestsPerRoute {
		fmt.Fprintf(w, "requests_per_route{route=\"%s\"} %d\n", route, count)
	}
	for route, count := range appMetrics.ErrorCounts {
		fmt.Fprintf(w, "error_count{route=\"%s\"} %d\n", route, count)
	}
	for route, total := range appMetrics.TotalLatency {
		avg := float64(total.Milliseconds()) / float64(appMetrics.RequestsPerRoute[route])
		fmt.Fprintf(w, "avg_latency_ms{route=\"%s\"} %.2f\n", route, avg)
	}
}
