package api

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/app-vault/app-vault/internal/metrics"
	"github.com/app-vault/app-vault/internal/service"
	"github.com/google/uuid"
)

// Handler contains all API handlers
type Handler struct {
	authService     *service.AuthService
	vaultService    *service.VaultService
	rotationService *service.RotationService
}

// NewHandler creates a new API handler
func NewHandler(authService *service.AuthService, vaultService *service.VaultService, rotationService *service.RotationService) *Handler {
	return &Handler{
		authService:     authService,
		vaultService:    vaultService,
		rotationService: rotationService,
	}
}

// RegisterRequest is the request body for user registration
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse is the response for user registration
type RegisterResponse struct {
	Token     string       `json:"token"`
	User      UserResponse `json:"user"`
	SecretKey string       `json:"secretKey"`
}

// UserResponse is the response for user data
type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
}

// Register handles user registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.authService.Register(r.Context(), &service.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.authService.GenerateToken(result.UserID, result.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusCreated, RegisterResponse{
		Token: token,
		User: UserResponse{
			ID:        result.UserID.String(),
			Email:     result.Email,
			CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
		},
		SecretKey: result.SecretKey,
	})
}

// LoginRequest is the request body for user login
type LoginRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	SecretKey string `json:"secret_key"`
}

// LoginResponse is the response for user login
type LoginResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// Login handles user login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := parseJSON(r, &req); err != nil {
		metrics.GetMetrics().RecordAuthFailure()
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.authService.Login(r.Context(), &service.LoginRequest{
		Email:     req.Email,
		Password:  req.Password,
		SecretKey: req.SecretKey,
	})
	if err != nil {
		metrics.GetMetrics().RecordAuthFailure()
		// Return more specific error for debugging
		errMsg := err.Error()
		if strings.Contains(errMsg, "user not found") {
			writeError(w, http.StatusUnauthorized, "User not found")
		} else if strings.Contains(errMsg, "secret key mismatch") {
			writeError(w, http.StatusUnauthorized, "Secret key is incorrect")
		} else if strings.Contains(errMsg, "password incorrect") {
			writeError(w, http.StatusUnauthorized, "Password is incorrect")
		} else {
			writeError(w, http.StatusUnauthorized, errMsg)
		}
		return
	}

	metrics.GetMetrics().RecordAuthSuccess()
	writeJSON(w, http.StatusOK, LoginResponse{
		Token:  result.Token,
		UserID: result.UserID.String(),
		Email:  result.Email,
	})
}

// CreateSecretRequest is the request body for creating a secret
type CreateSecretRequest struct {
	Name       string   `json:"name"`
	Value      string   `json:"value"`
	SecretType string   `json:"secret_type,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	ExpiresAt  *string  `json:"expires_at,omitempty"`
}

// CreateSecretResponse is the response for creating a secret
type CreateSecretResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Version   int    `json:"version"`
	CreatedAt string `json:"created_at"`
}

// CreateSecret handles secret creation
func (h *Handler) CreateSecret(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateSecretRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid expires_at format")
			return
		}
		expiresAt = &t
	}

	authHeader := r.Header.Get("Authorization")
	password := r.Header.Get("X-Vault-Password")
	secretKey := r.Header.Get("X-Vault-Secret-Key")

	if password == "" || secretKey == "" {
		writeError(w, http.StatusBadRequest, "X-Vault-Password and X-Vault-Secret-Key headers required")
		return
	}

	vaultKey, err := h.vaultService.GetVaultKey(r.Context(), userID, password, secretKey)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "failed to decrypt vault key")
		return
	}
	defer func() {
		for i := range vaultKey {
			vaultKey[i] = 0
		}
	}()

	_ = authHeader

	result, err := h.vaultService.CreateSecret(r.Context(), vaultKey, &service.CreateSecretRequest{
		UserID:     userID,
		Name:       req.Name,
		Value:      req.Value,
		SecretType: req.SecretType,
		Tags:       req.Tags,
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	metrics.GetMetrics().RecordSecretCreate()
	writeJSON(w, http.StatusCreated, CreateSecretResponse{
		ID:        result.ID.String(),
		Name:      result.Name,
		Version:   result.Version,
		CreatedAt: result.CreatedAt.Format(time.RFC3339),
	})
}

// GetSecretResponse is the response for getting a secret
type GetSecretResponse struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Value          string   `json:"value"`
	PublicKey      *string  `json:"public_key"`
	SecretType     string   `json:"secret_type"`
	Version        int      `json:"version"`
	Tags           []string `json:"tags"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	ExpiresAt      *string  `json:"expires_at,omitempty"`
	AccessCount    int      `json:"access_count"`
	LastAccessedAt *string  `json:"last_accessed_at,omitempty"`
}

// GetSecretByName handles secret retrieval by name
func (h *Handler) GetSecretByName(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name parameter required")
		return
	}

	password := r.Header.Get("X-Vault-Password")
	secretKey := r.Header.Get("X-Vault-Secret-Key")

	if password == "" || secretKey == "" {
		writeError(w, http.StatusBadRequest, "X-Vault-Password and X-Vault-Secret-Key headers required")
		return
	}

	vaultKey, err := h.vaultService.GetVaultKey(r.Context(), userID, password, secretKey)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "failed to decrypt vault key")
		return
	}
	defer func() {
		for i := range vaultKey {
			vaultKey[i] = 0
		}
	}()

	result, err := h.vaultService.GetSecret(r.Context(), vaultKey, &service.GetSecretRequest{
		UserID: userID,
		Name:   name,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	metrics.GetMetrics().RecordSecretRead()
	response := GetSecretResponse{
		ID:          result.ID.String(),
		Name:        result.Name,
		Value:       result.Value,
		SecretType:  result.SecretType,
		Version:     result.Version,
		Tags:        result.Tags,
		CreatedAt:   result.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   result.UpdatedAt.Format(time.RFC3339),
		AccessCount: result.AccessCount,
	}

	if len(result.PublicKey) > 0 {
		publicKeyStr := base64.StdEncoding.EncodeToString(result.PublicKey)
		response.PublicKey = &publicKeyStr
	}
	if result.ExpiresAt != nil {
		expiresAt := result.ExpiresAt.Format(time.RFC3339)
		response.ExpiresAt = &expiresAt
	}
	if result.LastAccessedAt != nil {
		lastAccessedAt := result.LastAccessedAt.Format(time.RFC3339)
		response.LastAccessedAt = &lastAccessedAt
	}

	writeJSON(w, http.StatusOK, response)
}

// GetSecretByID handles secret retrieval by ID
func (h *Handler) GetSecretByID(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	secretIDStr := r.URL.Path[len("/api/v1/secrets/"):]
	secretID, err := uuid.Parse(secretIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid secret ID")
		return
	}

	password := r.Header.Get("X-Vault-Password")
	secretKey := r.Header.Get("X-Vault-Secret-Key")

	if password == "" || secretKey == "" {
		writeError(w, http.StatusBadRequest, "X-Vault-Password and X-Vault-Secret-Key headers required")
		return
	}

	vaultKey, err := h.vaultService.GetVaultKey(r.Context(), userID, password, secretKey)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "failed to decrypt vault key")
		return
	}
	defer func() {
		for i := range vaultKey {
			vaultKey[i] = 0
		}
	}()

	result, err := h.vaultService.GetSecretByID(r.Context(), vaultKey, userID, secretID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	metrics.GetMetrics().RecordSecretRead()
	response := GetSecretResponse{
		ID:          result.ID.String(),
		Name:        result.Name,
		Value:       result.Value,
		SecretType:  result.SecretType,
		Version:     result.Version,
		Tags:        result.Tags,
		CreatedAt:   result.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   result.UpdatedAt.Format(time.RFC3339),
		AccessCount: result.AccessCount,
	}

	if len(result.PublicKey) > 0 {
		publicKeyStr := base64.StdEncoding.EncodeToString(result.PublicKey)
		response.PublicKey = &publicKeyStr
	}
	if result.ExpiresAt != nil {
		expiresAt := result.ExpiresAt.Format(time.RFC3339)
		response.ExpiresAt = &expiresAt
	}
	if result.LastAccessedAt != nil {
		lastAccessedAt := result.LastAccessedAt.Format(time.RFC3339)
		response.LastAccessedAt = &lastAccessedAt
	}

	writeJSON(w, http.StatusOK, response)
}

// SecretListItem represents a secret in the list
type SecretListItem struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	SecretType     string   `json:"secret_type"`
	Version        int      `json:"version"`
	Tags           []string `json:"tags"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	ExpiresAt      *string  `json:"expires_at,omitempty"`
	AccessCount    int      `json:"access_count"`
	LastAccessedAt *string  `json:"last_accessed_at,omitempty"`
}

// ListSecrets handles listing all secrets
func (h *Handler) ListSecrets(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := h.vaultService.ListSecrets(r.Context(), &service.ListSecretsRequest{
		UserID: userID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]SecretListItem, len(result))
	for i, item := range result {
		response[i] = SecretListItem{
			ID:          item.ID.String(),
			Name:        item.Name,
			SecretType:  item.SecretType,
			Version:     item.Version,
			Tags:        item.Tags,
			CreatedAt:   item.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   item.UpdatedAt.Format(time.RFC3339),
			AccessCount: item.AccessCount,
		}

		if item.ExpiresAt != nil {
			expiresAt := item.ExpiresAt.Format(time.RFC3339)
			response[i].ExpiresAt = &expiresAt
		}
		if item.LastAccessedAt != nil {
			lastAccessedAt := item.LastAccessedAt.Format(time.RFC3339)
			response[i].LastAccessedAt = &lastAccessedAt
		}
	}

	writeJSON(w, http.StatusOK, response)
}

// DeleteSecret handles secret deletion
func (h *Handler) DeleteSecret(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	secretIDStr := r.URL.Path[len("/api/v1/secrets/"):]
	secretID, err := uuid.Parse(secretIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid secret ID")
		return
	}

	err = h.vaultService.DeleteSecret(r.Context(), &service.DeleteSecretRequest{
		UserID:   userID,
		SecretID: secretID,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	metrics.GetMetrics().RecordSecretDelete()
	w.WriteHeader(http.StatusNoContent)
}

// RotateKeyRequest is the request body for key rotation
type RotateKeyRequest struct {
	Password  string `json:"password"`
	SecretKey string `json:"secret_key"`
	Reason    string `json:"reason,omitempty"`
}

// RotateKeyResponse is the response for key rotation
type RotateKeyResponse struct {
	NewKeyVersionID    string `json:"new_key_version_id"`
	SecretsReEncrypted int    `json:"secrets_reencrypted"`
	DurationMs         int64  `json:"duration_ms"`
}

// RotateVaultKey handles vault key rotation
func (h *Handler) RotateVaultKey(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req RotateKeyRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.rotationService.RotateVaultKey(r.Context(), &service.RotateVaultKeyRequest{
		UserID:    userID,
		Password:  req.Password,
		SecretKey: req.SecretKey,
		Reason:    req.Reason,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	metrics.GetMetrics().RecordKeyRotation()
	writeJSON(w, http.StatusOK, RotateKeyResponse{
		NewKeyVersionID:    result.NewKeyVersionID.String(),
		SecretsReEncrypted: result.SecretsReEncrypted,
		DurationMs:         result.Duration.Milliseconds(),
	})
}

// ChangePasswordRequest is the request body for password change
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
	SecretKey   string `json:"secret_key"`
}

// ChangePassword handles password change
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req ChangePasswordRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	_, err = h.rotationService.ChangeMasterKey(r.Context(), &service.ChangeMasterKeyRequest{
		UserID:      userID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
		SecretKey:   req.SecretKey,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// GetKeyStatus handles key status retrieval
func (h *Handler) GetKeyStatus(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := h.rotationService.GetKeyStatus(r.Context(), &service.GetKeyStatusRequest{
		UserID: userID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// CreateServicePrincipalRequest is the request body for creating service principals
type CreateServicePrincipalRequest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Permissions    []string `json:"permissions"`
	AllowedSecrets []string `json:"allowed_secrets"`
	AllowedTags    []string `json:"allowed_tags"`
	IPWhitelist    []string `json:"ip_whitelist"`
	RateLimit      int      `json:"rate_limit"`
	ExpiresAt      *string  `json:"expires_at,omitempty"`
}

// CreateServicePrincipal handles service principal creation
func (h *Handler) CreateServicePrincipal(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateServicePrincipalRequest
	if err := parseJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate required fields
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Parse expiration time if provided
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		parsedTime, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid expires_at format, use RFC3339")
			return
		}
		expiresAt = &parsedTime
	}

	// Set default rate limit if not provided
	rateLimit := req.RateLimit
	if rateLimit == 0 {
		rateLimit = 1000
	}

	// Set default permissions if not provided
	permissions := req.Permissions
	if len(permissions) == 0 {
		permissions = []string{"read"}
	}

	result, err := h.authService.CreateServicePrincipal(r.Context(), &service.CreateServicePrincipalRequest{
		UserID:         userID,
		Name:           req.Name,
		Description:    req.Description,
		Permissions:    permissions,
		AllowedSecrets: req.AllowedSecrets,
		AllowedTags:    req.AllowedTags,
		IPWhitelist:    req.IPWhitelist,
		RateLimit:      rateLimit,
		ExpiresAt:      expiresAt,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	metrics.GetMetrics().RecordAuthSuccess()

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":            result.ID,
		"client_id":     result.ClientID,
		"client_secret": result.ClientSecret,
	})
}

// ListServicePrincipals handles listing service principals for a user
func (h *Handler) ListServicePrincipals(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get all service principals for the user from database
	servicePrincipals, err := h.authService.ListServicePrincipals(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, servicePrincipals)
}

// GetServicePrincipal handles retrieving a specific service principal
func (h *Handler) GetServicePrincipal(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Extract service principal ID from URL path
	idStr := r.URL.Path[len("/api/v1/service-principals/"):]
	spID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid service principal ID")
		return
	}

	sp, err := h.authService.GetServicePrincipal(r.Context(), userID, spID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, sp)
}

// DeleteServicePrincipal handles service principal deletion
func (h *Handler) DeleteServicePrincipal(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Extract service principal ID from URL path
	idStr := r.URL.Path[len("/api/v1/service-principals/"):]
	spID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid service principal ID")
		return
	}

	if err := h.authService.DeleteServicePrincipal(r.Context(), userID, spID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	metrics.GetMetrics().RecordAuthSuccess()

	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// RegenerateServicePrincipalSecret handles regenerating a service principal's secret
func (h *Handler) RegenerateServicePrincipalSecret(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Extract service principal ID from URL path
	idStr := r.URL.Path[len("/api/v1/service-principals/"):]
	// Remove "/regenerate" suffix
	idStr = idStr[:len(idStr)-len("/regenerate")]
	spID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid service principal ID")
		return
	}

	newSecret, err := h.authService.RegenerateServicePrincipalSecret(r.Context(), userID, spID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	metrics.GetMetrics().RecordAuthSuccess()

	writeJSON(w, http.StatusOK, map[string]string{
		"client_secret": newSecret,
	})
}
