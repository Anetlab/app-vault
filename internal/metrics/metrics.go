package metrics

import (
"fmt"
"net/http"
"runtime"
"sync"
"sync/atomic"
"time"
)

type Metrics struct {
requestCount      atomic.Uint64
errorCount        atomic.Uint64
authSuccessCount  atomic.Uint64
authFailureCount  atomic.Uint64
secretCreateCount atomic.Uint64
secretReadCount   atomic.Uint64
secretDeleteCount atomic.Uint64
keyRotationCount  atomic.Uint64

requestDurations sync.Map

startTime time.Time
}

type durationStats struct {
count    atomic.Uint64
totalMs  atomic.Uint64
minMs    atomic.Uint64
maxMs    atomic.Uint64
}

var globalMetrics = &Metrics{
startTime: time.Now(),
}

func GetMetrics() *Metrics {
return globalMetrics
}

func (m *Metrics) RecordRequest() {
m.requestCount.Add(1)
}

func (m *Metrics) RecordError() {
m.errorCount.Add(1)
}

func (m *Metrics) RecordAuthSuccess() {
m.authSuccessCount.Add(1)
}

func (m *Metrics) RecordAuthFailure() {
m.authFailureCount.Add(1)
}

func (m *Metrics) RecordSecretCreate() {
m.secretCreateCount.Add(1)
}

func (m *Metrics) RecordSecretRead() {
m.secretReadCount.Add(1)
}

func (m *Metrics) RecordSecretDelete() {
m.secretDeleteCount.Add(1)
}

func (m *Metrics) RecordKeyRotation() {
m.keyRotationCount.Add(1)
}

func (m *Metrics) RecordRequestDuration(endpoint string, durationMs uint64) {
val, _ := m.requestDurations.LoadOrStore(endpoint, &durationStats{})
stats := val.(*durationStats)

stats.count.Add(1)
stats.totalMs.Add(durationMs)

for {
currentMin := stats.minMs.Load()
if currentMin == 0 || durationMs < currentMin {
if stats.minMs.CompareAndSwap(currentMin, durationMs) {
break
}
} else {
break
}
}

for {
currentMax := stats.maxMs.Load()
if durationMs > currentMax {
if stats.maxMs.CompareAndSwap(currentMax, durationMs) {
break
}
} else {
break
}
}
}

func (m *Metrics) MetricsHandler() http.HandlerFunc {
return func(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet {
http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
return
}

w.Header().Set("Content-Type", "text/plain; version=0.0.4")

uptime := time.Since(m.startTime).Seconds()

var memStats runtime.MemStats
runtime.ReadMemStats(&memStats)

fmt.Fprintf(w, "# HELP appvault_uptime_seconds Server uptime in seconds\n")
fmt.Fprintf(w, "# TYPE appvault_uptime_seconds gauge\n")
fmt.Fprintf(w, "appvault_uptime_seconds %.2f\n", uptime)

fmt.Fprintf(w, "\n# HELP appvault_requests_total Total number of requests\n")
fmt.Fprintf(w, "# TYPE appvault_requests_total counter\n")
fmt.Fprintf(w, "appvault_requests_total %d\n", m.requestCount.Load())

fmt.Fprintf(w, "\n# HELP appvault_errors_total Total number of errors\n")
fmt.Fprintf(w, "# TYPE appvault_errors_total counter\n")
fmt.Fprintf(w, "appvault_errors_total %d\n", m.errorCount.Load())

fmt.Fprintf(w, "\n# HELP appvault_auth_success_total Successful authentications\n")
fmt.Fprintf(w, "# TYPE appvault_auth_success_total counter\n")
fmt.Fprintf(w, "appvault_auth_success_total %d\n", m.authSuccessCount.Load())

fmt.Fprintf(w, "\n# HELP appvault_auth_failure_total Failed authentications\n")
fmt.Fprintf(w, "# TYPE appvault_auth_failure_total counter\n")
fmt.Fprintf(w, "appvault_auth_failure_total %d\n", m.authFailureCount.Load())

fmt.Fprintf(w, "\n# HELP appvault_secrets_created_total Secrets created\n")
fmt.Fprintf(w, "# TYPE appvault_secrets_created_total counter\n")
fmt.Fprintf(w, "appvault_secrets_created_total %d\n", m.secretCreateCount.Load())

fmt.Fprintf(w, "\n# HELP appvault_secrets_read_total Secrets read\n")
fmt.Fprintf(w, "# TYPE appvault_secrets_read_total counter\n")
fmt.Fprintf(w, "appvault_secrets_read_total %d\n", m.secretReadCount.Load())

fmt.Fprintf(w, "\n# HELP appvault_secrets_deleted_total Secrets deleted\n")
fmt.Fprintf(w, "# TYPE appvault_secrets_deleted_total counter\n")
fmt.Fprintf(w, "appvault_secrets_deleted_total %d\n", m.secretDeleteCount.Load())

fmt.Fprintf(w, "\n# HELP appvault_key_rotations_total Key rotations performed\n")
fmt.Fprintf(w, "# TYPE appvault_key_rotations_total counter\n")
fmt.Fprintf(w, "appvault_key_rotations_total %d\n", m.keyRotationCount.Load())

fmt.Fprintf(w, "\n# HELP appvault_memory_alloc_bytes Memory allocated in bytes\n")
fmt.Fprintf(w, "# TYPE appvault_memory_alloc_bytes gauge\n")
fmt.Fprintf(w, "appvault_memory_alloc_bytes %d\n", memStats.Alloc)

fmt.Fprintf(w, "\n# HELP appvault_memory_sys_bytes Total memory from system\n")
fmt.Fprintf(w, "# TYPE appvault_memory_sys_bytes gauge\n")
fmt.Fprintf(w, "appvault_memory_sys_bytes %d\n", memStats.Sys)

fmt.Fprintf(w, "\n# HELP appvault_goroutines Number of goroutines\n")
fmt.Fprintf(w, "# TYPE appvault_goroutines gauge\n")
fmt.Fprintf(w, "appvault_goroutines %d\n", runtime.NumGoroutine())

m.requestDurations.Range(func(key, value interface{}) bool {
endpoint := key.(string)
stats := value.(*durationStats)

count := stats.count.Load()
if count > 0 {
total := stats.totalMs.Load()
avg := float64(total) / float64(count)

fmt.Fprintf(w, "\n# HELP appvault_request_duration_ms Request duration for %s\n", endpoint)
fmt.Fprintf(w, "appvault_request_duration_ms{endpoint=\"%s\",stat=\"avg\"} %.2f\n", endpoint, avg)
fmt.Fprintf(w, "appvault_request_duration_ms{endpoint=\"%s\",stat=\"min\"} %d\n", endpoint, stats.minMs.Load())
fmt.Fprintf(w, "appvault_request_duration_ms{endpoint=\"%s\",stat=\"max\"} %d\n", endpoint, stats.maxMs.Load())
}
return true
})
}
}

func MetricsMiddleware(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
start := time.Now()

globalMetrics.RecordRequest()

wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
next.ServeHTTP(wrapped, r)

duration := time.Since(start).Milliseconds()
globalMetrics.RecordRequestDuration(r.URL.Path, uint64(duration))

if wrapped.statusCode >= 400 {
globalMetrics.RecordError()
}
})
}

type responseWriter struct {
http.ResponseWriter
statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
rw.statusCode = code
rw.ResponseWriter.WriteHeader(code)
}

func HealthHandler(dbHealthy func() bool) http.HandlerFunc {
return func(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet {
http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
return
}

healthy := dbHealthy()

w.Header().Set("Content-Type", "application/json")
if healthy {
w.WriteHeader(http.StatusOK)
w.Write([]byte(`{"status":"healthy","database":"connected"}`))
} else {
w.WriteHeader(http.StatusServiceUnavailable)
w.Write([]byte(`{"status":"unhealthy","database":"disconnected"}`))
}
}
}
