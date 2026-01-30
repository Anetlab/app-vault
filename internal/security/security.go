package security

import (
"crypto/tls"
"net/http"
)

func GetTLSConfig() *tls.Config {
return &tls.Config{
MinVersion:               tls.VersionTLS13,
PreferServerCipherSuites: true,
CurvePreferences: []tls.CurveID{
tls.X25519,
tls.CurveP256,
},
CipherSuites: []uint16{
tls.TLS_AES_256_GCM_SHA384,
tls.TLS_AES_128_GCM_SHA256,
tls.TLS_CHACHA20_POLY1305_SHA256,
},
}
}

func SecurityHeadersMiddleware(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
w.Header().Set("X-Frame-Options", "DENY")
w.Header().Set("X-Content-Type-Options", "nosniff")
w.Header().Set("X-XSS-Protection", "1; mode=block")
w.Header().Set("Content-Security-Policy", "default-src ''self''; script-src ''self''; style-src ''self''; img-src ''self'' data:; font-src ''self''; connect-src ''self''; frame-ancestors ''none''")
w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

next.ServeHTTP(w, r)
})
}

func RequestSizeLimitMiddleware(maxBytes int64) func(http.Handler) http.Handler {
return func(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
next.ServeHTTP(w, r)
})
}
}

func ContentTypeMiddleware(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
ct := r.Header.Get("Content-Type")
if ct != "" && ct != "application/json" {
http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
return
}
}
next.ServeHTTP(w, r)
})
}
