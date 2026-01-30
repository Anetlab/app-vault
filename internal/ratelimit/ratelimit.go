package ratelimit

import (
"fmt"
"net/http"
"sync"
"time"
)

// RateLimiter manages rate limiting
type RateLimiter struct {
cache   map[string]*rateLimitEntry
mu      sync.RWMutex
window  time.Duration
maxReqs int
}

type rateLimitEntry struct {
count      int
windowStart time.Time
mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(window time.Duration, maxReqs int) *RateLimiter {
rl := &RateLimiter{
cache:   make(map[string]*rateLimitEntry),
window:  window,
maxReqs: maxReqs,
}

go rl.cleanupLoop()

return rl
}

// Check checks if a request should be allowed
func (rl *RateLimiter) Check(key string) (bool, int, error) {
rl.mu.Lock()
entry, exists := rl.cache[key]
if !exists {
entry = &rateLimitEntry{
count:      0,
windowStart: time.Now(),
}
rl.cache[key] = entry
}
rl.mu.Unlock()

entry.mu.Lock()
defer entry.mu.Unlock()

now := time.Now()
if now.Sub(entry.windowStart) >= rl.window {
entry.count = 0
entry.windowStart = now
}

if entry.count >= rl.maxReqs {
return false, rl.maxReqs - entry.count, nil
}

entry.count++
return true, rl.maxReqs - entry.count, nil
}

// cleanupLoop periodically cleans up old entries
func (rl *RateLimiter) cleanupLoop() {
ticker := time.NewTicker(5 * time.Minute)
defer ticker.Stop()

for range ticker.C {
rl.mu.Lock()
now := time.Now()
for key, entry := range rl.cache {
entry.mu.Lock()
if now.Sub(entry.windowStart) > rl.window*2 {
delete(rl.cache, key)
}
entry.mu.Unlock()
}
rl.mu.Unlock()
}
}

// Middleware creates a rate limiting middleware
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
key := getClientIP(r)

allowed, remaining, err := rl.Check(key)
if err != nil {
http.Error(w, "internal error", http.StatusInternalServerError)
return
}

w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxReqs))
w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
w.Header().Set("X-RateLimit-Window", rl.window.String())

if !allowed {
w.Header().Set("Retry-After", fmt.Sprintf("%d", int(rl.window.Seconds())))
http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
return
}

next.ServeHTTP(w, r)
})
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
return xff
}

if xri := r.Header.Get("X-Real-IP"); xri != "" {
return xri
}

return r.RemoteAddr
}
