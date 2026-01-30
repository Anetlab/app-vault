package service

import (
	"context"
	"fmt"
	"time"

	"github.com/app-vault/app-vault/internal/crypto"
	"github.com/app-vault/app-vault/internal/db"
	"github.com/app-vault/app-vault/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthService handles authentication operations
type AuthService struct {
	db        *db.Database
	jwtSecret []byte
}

// NewAuthService creates a new authentication service
func NewAuthService(database *db.Database, jwtSecret string) *AuthService {
	return &AuthService{
		db:        database,
		jwtSecret: []byte(jwtSecret),
	}
}

// RegisterRequest contains user registration data
type RegisterRequest struct {
	Email    string
	Password string
}

// RegisterResponse contains registration response data
type RegisterResponse struct {
	UserID    uuid.UUID
	Email     string
	SecretKey string
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	if len(req.Password) < 12 {
		return nil, fmt.Errorf("password must be at least 12 characters")
	}

	existing, err := s.db.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("user already exists")
	}

	secretKey, err := crypto.GenerateSecretKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret key: %w", err)
	}

	masterKeySalt, err := crypto.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate master key salt: %w", err)
	}

	masterKey := crypto.DeriveMasterKey(req.Password, secretKey, masterKeySalt)
	defer crypto.ZeroBytes(masterKey)

	vaultKey, err := crypto.GenerateVaultKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate vault key: %w", err)
	}
	defer crypto.ZeroBytes(vaultKey)

	kek, err := crypto.DeriveKeyEncryptionKey(masterKey, "vault-key-wrap")
	if err != nil {
		return nil, fmt.Errorf("failed to derive key encryption key: %w", err)
	}
	defer crypto.ZeroBytes(kek)

	vaultKeyNonce, err := crypto.GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	encryptedVaultKey, err := crypto.EncryptVaultKey(vaultKey, kek, vaultKeyNonce)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt vault key: %w", err)
	}

	secretKeyHash := crypto.HashSecretKey(secretKey)

	srpVerifier := []byte("placeholder")

	userID := uuid.New()
	user := &models.User{
		ID:                  userID,
		Email:               req.Email,
		SecretKeyHash:       secretKeyHash,
		MasterKeySalt:       masterKeySalt,
		EncryptedVaultKey:   encryptedVaultKey,
		VaultKeyNonce:       vaultKeyNonce,
		SRPVerifier:         srpVerifier,
		CurrentKeyVersionID: nil,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	keyVersionID := uuid.New()
	keyVersion := &models.KeyVersion{
		ID:                keyVersionID,
		UserID:            userID,
		Version:           1,
		EncryptedVaultKey: encryptedVaultKey,
		VaultKeyNonce:     vaultKeyNonce,
		Status:            models.KeyStatusActive,
		CreatedAt:         time.Now(),
		RotationReason:    "initial key",
	}

	user.CurrentKeyVersionID = &keyVersionID

	if err := s.db.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.db.CreateKeyVersion(ctx, keyVersion); err != nil {
		return nil, fmt.Errorf("failed to create key version: %w", err)
	}

	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		UserID:     userID,
		Action:     "user.registered",
		ResourceID: userID.String(),
		CreatedAt:  time.Now(),
	}
	_ = s.db.CreateAuditLog(ctx, auditLog)

	return &RegisterResponse{
		UserID:    userID,
		Email:     req.Email,
		SecretKey: secretKey,
	}, nil
}

// LoginRequest contains user login data
type LoginRequest struct {
	Email     string
	Password  string
	SecretKey string
}

// LoginResponse contains login response data
type LoginResponse struct {
	Token  string
	UserID uuid.UUID
	Email  string
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	user, err := s.db.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("authentication failed: user not found")
	}

	if !crypto.CompareHashConstantTime(user.SecretKeyHash, req.SecretKey) {
		return nil, fmt.Errorf("authentication failed: secret key mismatch")
	}

	masterKey := crypto.DeriveMasterKey(req.Password, req.SecretKey, user.MasterKeySalt)
	defer crypto.ZeroBytes(masterKey)

	kek, err := crypto.DeriveKeyEncryptionKey(masterKey, "vault-key-wrap")
	if err != nil {
		return nil, fmt.Errorf("failed to derive key encryption key: %w", err)
	}
	defer crypto.ZeroBytes(kek)

	vaultKey, err := crypto.DecryptVaultKey(user.EncryptedVaultKey, kek, user.VaultKeyNonce)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: password incorrect")
	}
	crypto.ZeroBytes(vaultKey)

	token, err := s.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		UserID:     user.ID,
		Action:     "user.login",
		ResourceID: user.ID.String(),
		CreatedAt:  time.Now(),
	}
	_ = s.db.CreateAuditLog(ctx, auditLog)

	return &LoginResponse{
		Token:  token,
		UserID: user.ID,
		Email:  user.Email,
	}, nil
}

// ServicePrincipalLoginRequest contains service principal authentication data
type ServicePrincipalLoginRequest struct {
	ClientID     string
	ClientSecret string
}

// ServicePrincipalLoginResponse contains service principal login response
type ServicePrincipalLoginResponse struct {
	Token  string
	UserID uuid.UUID
}

// ServicePrincipalLogin authenticates a service principal
func (s *AuthService) ServicePrincipalLogin(ctx context.Context, req *ServicePrincipalLoginRequest) (*ServicePrincipalLoginResponse, error) {
	sp, err := s.db.GetServicePrincipalByClientID(ctx, req.ClientID)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}
	if sp == nil {
		return nil, fmt.Errorf("authentication failed: invalid credentials")
	}

	if !sp.IsActive {
		return nil, fmt.Errorf("service principal is not active")
	}

	if sp.ExpiresAt != nil && time.Now().After(*sp.ExpiresAt) {
		return nil, fmt.Errorf("service principal has expired")
	}

	if !crypto.CompareHashConstantTime(sp.ClientSecretHash, req.ClientSecret) {
		return nil, fmt.Errorf("authentication failed: invalid credentials")
	}

	token, err := s.GenerateServicePrincipalToken(sp.ID, sp.UserID, sp.Permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	_ = s.db.UpdateServicePrincipalLastUsed(ctx, sp.ID)

	auditLog := &models.AuditLog{
		ID:                 uuid.New(),
		UserID:             sp.UserID,
		ServicePrincipalID: &sp.ID,
		Action:             "service_principal.login",
		ResourceID:         sp.ID.String(),
		CreatedAt:          time.Now(),
	}
	_ = s.db.CreateAuditLog(ctx, auditLog)

	return &ServicePrincipalLoginResponse{
		Token:  token,
		UserID: sp.UserID,
	}, nil
}

// GenerateToken generates a JWT token for a user
func (s *AuthService) GenerateToken(userID uuid.UUID, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"email":   email,
		"type":    "user",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// GenerateServicePrincipalToken generates a JWT token for a service principal
func (s *AuthService) GenerateServicePrincipalToken(spID, userID uuid.UUID, permissions []string) (string, error) {
	claims := jwt.MapClaims{
		"sp_id":       spID.String(),
		"user_id":     userID.String(),
		"permissions": permissions,
		"type":        "service_principal",
		"exp":         time.Now().Add(1 * time.Hour).Unix(),
		"iat":         time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// CreateServicePrincipalRequest contains service principal creation data
type CreateServicePrincipalRequest struct {
	UserID         uuid.UUID
	Name           string
	Description    string
	Permissions    []string
	AllowedSecrets []string
	AllowedTags    []string
	IPWhitelist    []string
	RateLimit      int
	ExpiresAt      *time.Time
}

// CreateServicePrincipalResponse contains service principal creation response
type CreateServicePrincipalResponse struct {
	ID           uuid.UUID
	ClientID     string
	ClientSecret string
}

// CreateServicePrincipal creates a new service principal
func (s *AuthService) CreateServicePrincipal(ctx context.Context, req *CreateServicePrincipalRequest) (*CreateServicePrincipalResponse, error) {
	clientID, err := crypto.GenerateClientID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client ID: %w", err)
	}

	clientSecret, err := crypto.GenerateClientSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client secret: %w", err)
	}

	clientSecretHash := crypto.HashClientSecret(clientSecret)

	sp := &models.ServicePrincipal{
		ID:               uuid.New(),
		UserID:           req.UserID,
		ClientID:         clientID,
		ClientSecretHash: clientSecretHash,
		Name:             req.Name,
		Description:      req.Description,
		Permissions:      req.Permissions,
		AllowedSecrets:   req.AllowedSecrets,
		AllowedTags:      req.AllowedTags,
		IPWhitelist:      req.IPWhitelist,
		RateLimit:        req.RateLimit,
		IsActive:         true,
		ExpiresAt:        req.ExpiresAt,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.db.CreateServicePrincipal(ctx, sp); err != nil {
		return nil, fmt.Errorf("failed to create service principal: %w", err)
	}

	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		UserID:     req.UserID,
		Action:     "service_principal.created",
		ResourceID: sp.ID.String(),
		CreatedAt:  time.Now(),
	}
	_ = s.db.CreateAuditLog(ctx, auditLog)

	return &CreateServicePrincipalResponse{
		ID:           sp.ID,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}

// ListServicePrincipals returns all service principals for a user
func (s *AuthService) ListServicePrincipals(ctx context.Context, userID uuid.UUID) ([]*models.ServicePrincipal, error) {
	return s.db.ListServicePrincipalsByUserID(ctx, userID)
}

// GetServicePrincipal retrieves a specific service principal
func (s *AuthService) GetServicePrincipal(ctx context.Context, userID, spID uuid.UUID) (*models.ServicePrincipal, error) {
	sp, err := s.db.GetServicePrincipalByID(ctx, spID)
	if err != nil {
		return nil, fmt.Errorf("service principal not found: %w", err)
	}

	// Verify ownership
	if sp.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to service principal")
	}

	return sp, nil
}

// DeleteServicePrincipal deletes a service principal
func (s *AuthService) DeleteServicePrincipal(ctx context.Context, userID, spID uuid.UUID) error {
	// First verify ownership
	sp, err := s.GetServicePrincipal(ctx, userID, spID)
	if err != nil {
		return err
	}

	if err := s.db.DeleteServicePrincipal(ctx, spID); err != nil {
		return fmt.Errorf("failed to delete service principal: %w", err)
	}

	// Audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		UserID:     userID,
		Action:     "service_principal.deleted",
		ResourceID: sp.ID.String(),
		CreatedAt:  time.Now(),
	}
	_ = s.db.CreateAuditLog(ctx, auditLog)

	return nil
}

// RegenerateServicePrincipalSecret generates a new client secret for a service principal
func (s *AuthService) RegenerateServicePrincipalSecret(ctx context.Context, userID, spID uuid.UUID) (string, error) {
	// First verify ownership
	_, err := s.GetServicePrincipal(ctx, userID, spID)
	if err != nil {
		return "", err
	}

	// Generate new client secret
	newClientSecret, err := crypto.GenerateClientSecret()
	if err != nil {
		return "", fmt.Errorf("failed to generate client secret: %w", err)
	}

	newClientSecretHash := crypto.HashClientSecret(newClientSecret)

	// Update in database
	if err := s.db.UpdateServicePrincipalSecret(ctx, spID, []byte(newClientSecretHash)); err != nil {
		return "", fmt.Errorf("failed to update service principal secret: %w", err)
	}

	// Audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		UserID:     userID,
		Action:     "service_principal.secret_regenerated",
		ResourceID: spID.String(),
		CreatedAt:  time.Now(),
	}
	_ = s.db.CreateAuditLog(ctx, auditLog)

	return newClientSecret, nil
}
