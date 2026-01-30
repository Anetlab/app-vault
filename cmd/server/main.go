package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/app-vault/app-vault/internal/api"
	"github.com/app-vault/app-vault/internal/db"
	"github.com/app-vault/app-vault/internal/metrics"
	"github.com/app-vault/app-vault/internal/ratelimit"
	"github.com/app-vault/app-vault/internal/security"
	"github.com/app-vault/app-vault/internal/service"
)

// Config holds application configuration
type Config struct {
	ServerPort       string
	DatabaseURL      string
	JWTSecret        string
	MigrationsPath   string
	EnableTLS        bool
	TLSCertFile      string
	TLSKeyFile       string
	RateLimitEnabled bool
	RateLimitMax     int
	RateLimitWindow  time.Duration
	MaxRequestSize   int64
}

// loadConfig loads configuration from environment variables
func loadConfig() *Config {
	enableTLS := getEnv("ENABLE_TLS", "false") == "true"
	rateLimitEnabled := getEnv("RATE_LIMIT_ENABLED", "true") == "true"
	rateLimitMax, _ := strconv.Atoi(getEnv("RATE_LIMIT_MAX", "100"))
	rateLimitWindowSec, _ := strconv.Atoi(getEnv("RATE_LIMIT_WINDOW_SECONDS", "60"))
	maxRequestSizeMB, _ := strconv.ParseInt(getEnv("MAX_REQUEST_SIZE_MB", "1"), 10, 64)

	return &Config{
		ServerPort:       getEnv("SERVER_PORT", "8888"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/appvault?sslmode=disable"),
		JWTSecret:        getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
		MigrationsPath:   getEnv("MIGRATIONS_PATH", "migrations"),
		EnableTLS:        enableTLS,
		TLSCertFile:      getEnv("TLS_CERT_FILE", "certs/server.crt"),
		TLSKeyFile:       getEnv("TLS_KEY_FILE", "certs/server.key"),
		RateLimitEnabled: rateLimitEnabled,
		RateLimitMax:     rateLimitMax,
		RateLimitWindow:  time.Duration(rateLimitWindowSec) * time.Second,
		MaxRequestSize:   maxRequestSizeMB * 1024 * 1024,
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	log.Println("Starting App Vault server...")

	cfg := loadConfig()

	database, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Println("Connected to database")

	migrationsPath := cfg.MigrationsPath
	if !filepath.IsAbs(migrationsPath) {
		wd, _ := os.Getwd()
		migrationsPath = filepath.Join(wd, migrationsPath)
	}

	if err := database.RunMigrations(migrationsPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed")

	authService := service.NewAuthService(database, cfg.JWTSecret)
	vaultService := service.NewVaultService(database)
	rotationService := service.NewRotationService(database)

	handler := api.NewHandler(authService, vaultService, rotationService)

	// Initialize metrics
	metricsInstance := metrics.GetMetrics()

	// Initialize rate limiter
	var rateLimiter *ratelimit.RateLimiter
	if cfg.RateLimitEnabled {
		rateLimiter = ratelimit.NewRateLimiter(cfg.RateLimitWindow, cfg.RateLimitMax)
		log.Printf("Rate limiting enabled: %d requests per %s", cfg.RateLimitMax, cfg.RateLimitWindow)
	}

	mux := http.NewServeMux()

	// Metrics endpoint (no auth required for monitoring)
	mux.HandleFunc("/metrics", metricsInstance.MetricsHandler())

	// Health check with database status
	mux.HandleFunc("/health", metrics.HealthHandler(func() bool {
		return database.Ping() == nil
	}))

	mux.HandleFunc("/api/v1/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.Register(w, r)
	})

	mux.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.Login(w, r)
	})

	authMiddleware := api.AuthMiddleware(authService)

	mux.Handle("/api/v1/secrets", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.CreateSecret(w, r)
		case http.MethodGet:
			handler.ListSecrets(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/v1/secrets/by-name", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.GetSecretByName(w, r)
	})))

	mux.Handle("/api/v1/secrets/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetSecretByID(w, r)
		case http.MethodDelete:
			handler.DeleteSecret(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/v1/keys/rotate", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.RotateVaultKey(w, r)
	})))

	mux.Handle("/api/v1/keys/status", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.GetKeyStatus(w, r)
	})))

	mux.Handle("/api/v1/auth/change-password", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.ChangePassword(w, r)
	})))

	// Service Principal endpoints
	mux.Handle("/api/v1/service-principals", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.CreateServicePrincipal(w, r)
		case http.MethodGet:
			handler.ListServicePrincipals(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/v1/service-principals/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a regenerate request
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/regenerate") {
			handler.RegenerateServicePrincipalSecret(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			handler.GetServicePrincipal(w, r)
		case http.MethodDelete:
			handler.DeleteServicePrincipal(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Build middleware chain
	finalHandler := api.RecoveryMiddleware(
		api.LoggingMiddleware(
			security.SecurityHeadersMiddleware(
				security.ContentTypeMiddleware(
					security.RequestSizeLimitMiddleware(cfg.MaxRequestSize)(
						metrics.MetricsMiddleware(
							api.CORSMiddleware(mux),
						),
					),
				),
			),
		),
	)

	// Add rate limiting if enabled
	if rateLimiter != nil {
		finalHandler = rateLimiter.Middleware(finalHandler)
	}

	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      finalHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Configure TLS if enabled
	if cfg.EnableTLS {
		server.TLSConfig = security.GetTLSConfig()
		log.Println("TLS enabled with TLS 1.3")
	}

	go func() {
		log.Printf("Server listening on port %s", cfg.ServerPort)
		var err error
		if cfg.EnableTLS {
			log.Printf("Starting HTTPS server with cert: %s", cfg.TLSCertFile)
			err = server.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile)
		} else {
			log.Println("Starting HTTP server (WARNING: TLS disabled)")
			err = server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
