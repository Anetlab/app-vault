package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/app-vault/app-vault/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// contextKey is a custom type for context keys
type contextKey string

const (
	contextKeyUserID      contextKey = "user_id"
	contextKeyTokenType   contextKey = "token_type"
	contextKeyPermissions contextKey = "permissions"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeError(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			tokenString := parts[1]
			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			userIDStr, ok := claims["user_id"].(string)
			if !ok {
				writeError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid user ID")
				return
			}

			tokenType, _ := claims["type"].(string)

			ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
			ctx = context.WithValue(ctx, contextKeyTokenType, tokenType)

			if permissions, ok := claims["permissions"].([]interface{}); ok {
				perms := make([]string, len(permissions))
				for i, p := range permissions {
					if pStr, ok := p.(string); ok {
						perms[i] = pStr
					}
				}
				ctx = context.WithValue(ctx, contextKeyPermissions, perms)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// getUserIDFromContext extracts user ID from context
func getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(contextKeyUserID).(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// getPermissionsFromContext extracts permissions from context
func getPermissionsFromContext(ctx context.Context) []string {
	permissions, ok := ctx.Value(contextKeyPermissions).([]string)
	if !ok {
		return nil
	}
	return permissions
}

// hasPermission checks if user has a specific permission
func hasPermission(ctx context.Context, required string) bool {
	permissions := getPermissionsFromContext(ctx)
	for _, perm := range permissions {
		if perm == required || perm == "admin" {
			return true
		}
	}
	return false
}

// CORSMiddleware adds CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[%s] %s %s\n", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("panic recovered: %v\n", err)
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Printf("failed to encode JSON response: %v\n", err)
	}
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

// parseJSON parses JSON request body
func parseJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("empty request body")
	}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

// extractClaims extracts JWT claims without validation (for internal use)
func extractClaims(tokenString string) (jwt.MapClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid claims")
}
