package service

import (
	"context"
	"fmt"
	"time"

	"github.com/app-vault/app-vault/internal/crypto"
	"github.com/app-vault/app-vault/internal/db"
	"github.com/app-vault/app-vault/internal/models"
	"github.com/google/uuid"
)

// VaultService handles secret management operations
type VaultService struct {
	db *db.Database
}

// NewVaultService creates a new vault service
func NewVaultService(database *db.Database) *VaultService {
	return &VaultService{
		db: database,
	}
}

// CreateSecretRequest contains secret creation data
type CreateSecretRequest struct {
	UserID     uuid.UUID
	Name       string
	Value      string
	SecretType string
	Tags       []string
	ExpiresAt  *time.Time
}

// CreateSecretResponse contains secret creation response
type CreateSecretResponse struct {
	ID        uuid.UUID
	Name      string
	Version   int
	CreatedAt time.Time
}

// CreateSecret creates a new secret
func (s *VaultService) CreateSecret(ctx context.Context, vaultKey []byte, req *CreateSecretRequest) (*CreateSecretResponse, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("secret name is required")
	}
	if req.Value == "" {
		return nil, fmt.Errorf("secret value is required")
	}

	activeKeyVersion, err := s.db.GetActiveKeyVersion(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active key version: %w", err)
	}
	if activeKeyVersion == nil {
		return nil, fmt.Errorf("no active key version found")
	}

	nonce, err := crypto.GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	encryptedData, err := crypto.EncryptData([]byte(req.Value), vaultKey, nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %w", err)
	}

	secretType := req.SecretType
	if secretType == "" {
		secretType = models.SecretTypeGeneric
	}

	existingSecret, err := s.db.GetSecretByName(ctx, req.UserID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing secret: %w", err)
	}

	var previousID *uuid.UUID
	version := 1
	if existingSecret != nil {
		previousID = &existingSecret.ID
		version = existingSecret.Version + 1
	}

	secret := &models.Secret{
		ID:            uuid.New(),
		UserID:        req.UserID,
		KeyVersionID:  activeKeyVersion.ID,
		Name:          req.Name,
		EncryptedData: encryptedData,
		Nonce:         nonce,
		SecretType:    secretType,
		Version:       version,
		PreviousID:    previousID,
		Tags:          req.Tags,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ExpiresAt:     req.ExpiresAt,
		AccessCount:   0,
	}

	if err := s.db.CreateSecret(ctx, secret); err != nil {
		return nil, fmt.Errorf("failed to create secret: %w", err)
	}

	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		UserID:     req.UserID,
		Action:     "secret.created",
		ResourceID: secret.ID.String(),
		Details: map[string]interface{}{
			"name":    req.Name,
			"type":    secretType,
			"version": version,
		},
		CreatedAt: time.Now(),
	}
	_ = s.db.CreateAuditLog(ctx, auditLog)

	return &CreateSecretResponse{
		ID:        secret.ID,
		Name:      req.Name,
		Version:   version,
		CreatedAt: secret.CreatedAt,
	}, nil
}

// GetSecretRequest contains secret retrieval data
type GetSecretRequest struct {
	UserID uuid.UUID
	Name   string
}

// GetSecretResponse contains secret data
type GetSecretResponse struct {
	ID             uuid.UUID
	Name           string
	Value          string
	SecretType     string
	Version        int
	Tags           []string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ExpiresAt      *time.Time
	AccessCount    int
	LastAccessedAt *time.Time
}

// GetSecret retrieves and decrypts a secret by name
func (s *VaultService) GetSecret(ctx context.Context, vaultKey []byte, req *GetSecretRequest) (*GetSecretResponse, error) {
	secret, err := s.db.GetSecretByName(ctx, req.UserID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found")
	}

	if secret.ExpiresAt != nil && time.Now().After(*secret.ExpiresAt) {
		return nil, fmt.Errorf("secret has expired")
	}

	keyVersion, err := s.db.GetKeyVersionByID(ctx, secret.KeyVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get key version: %w", err)
	}
	if keyVersion == nil {
		return nil, fmt.Errorf("key version not found")
	}

	if keyVersion.Status == models.KeyStatusDestroyed {
		return nil, fmt.Errorf("secret encrypted with destroyed key")
	}

	decryptedValue, err := crypto.DecryptData(secret.EncryptedData, vaultKey, secret.Nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret: %w", err)
	}
	defer crypto.ZeroBytes(decryptedValue)

	_ = s.db.UpdateSecretAccess(ctx, secret.ID)

	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		UserID:     req.UserID,
		Action:     "secret.read",
		ResourceID: secret.ID.String(),
		Details: map[string]interface{}{
			"name": req.Name,
		},
		CreatedAt: time.Now(),
	}
	_ = s.db.CreateAuditLog(ctx, auditLog)

	return &GetSecretResponse{
		ID:             secret.ID,
		Name:           secret.Name,
		Value:          string(decryptedValue),
		SecretType:     secret.SecretType,
		Version:        secret.Version,
		Tags:           secret.Tags,
		CreatedAt:      secret.CreatedAt,
		UpdatedAt:      secret.UpdatedAt,
		ExpiresAt:      secret.ExpiresAt,
		AccessCount:    secret.AccessCount + 1,
		LastAccessedAt: secret.LastAccessedAt,
	}, nil
}

// GetSecretByID retrieves and decrypts a secret by ID
func (s *VaultService) GetSecretByID(ctx context.Context, vaultKey []byte, userID, secretID uuid.UUID) (*GetSecretResponse, error) {
	secret, err := s.db.GetSecretByID(ctx, secretID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found")
	}

	if secret.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}

	if secret.ExpiresAt != nil && time.Now().After(*secret.ExpiresAt) {
		return nil, fmt.Errorf("secret has expired")
	}

	keyVersion, err := s.db.GetKeyVersionByID(ctx, secret.KeyVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get key version: %w", err)
	}
	if keyVersion == nil {
		return nil, fmt.Errorf("key version not found")
	}

	if keyVersion.Status == models.KeyStatusDestroyed {
		return nil, fmt.Errorf("secret encrypted with destroyed key")
	}

	decryptedValue, err := crypto.DecryptData(secret.EncryptedData, vaultKey, secret.Nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt secret: %w", err)
	}
	defer crypto.ZeroBytes(decryptedValue)

	_ = s.db.UpdateSecretAccess(ctx, secret.ID)

	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		UserID:     userID,
		Action:     "secret.read",
		ResourceID: secret.ID.String(),
		Details: map[string]interface{}{
			"name": secret.Name,
		},
		CreatedAt: time.Now(),
	}
	_ = s.db.CreateAuditLog(ctx, auditLog)

	return &GetSecretResponse{
		ID:             secret.ID,
		Name:           secret.Name,
		Value:          string(decryptedValue),
		SecretType:     secret.SecretType,
		Version:        secret.Version,
		Tags:           secret.Tags,
		CreatedAt:      secret.CreatedAt,
		UpdatedAt:      secret.UpdatedAt,
		ExpiresAt:      secret.ExpiresAt,
		AccessCount:    secret.AccessCount + 1,
		LastAccessedAt: secret.LastAccessedAt,
	}, nil
}

// ListSecretsRequest contains secret listing parameters
type ListSecretsRequest struct {
	UserID uuid.UUID
}

// SecretListItem represents a secret in the list (without value)
type SecretListItem struct {
	ID             uuid.UUID
	Name           string
	SecretType     string
	Version        int
	Tags           []string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ExpiresAt      *time.Time
	AccessCount    int
	LastAccessedAt *time.Time
}

// ListSecrets lists all secrets for a user without their values
func (s *VaultService) ListSecrets(ctx context.Context, req *ListSecretsRequest) ([]*SecretListItem, error) {
	secrets, err := s.db.ListSecrets(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	items := make([]*SecretListItem, len(secrets))
	for i, secret := range secrets {
		items[i] = &SecretListItem{
			ID:             secret.ID,
			Name:           secret.Name,
			SecretType:     secret.SecretType,
			Version:        secret.Version,
			Tags:           secret.Tags,
			CreatedAt:      secret.CreatedAt,
			UpdatedAt:      secret.UpdatedAt,
			ExpiresAt:      secret.ExpiresAt,
			AccessCount:    secret.AccessCount,
			LastAccessedAt: secret.LastAccessedAt,
		}
	}

	return items, nil
}

// DeleteSecretRequest contains secret deletion parameters
type DeleteSecretRequest struct {
	UserID   uuid.UUID
	SecretID uuid.UUID
}

// DeleteSecret deletes a secret
func (s *VaultService) DeleteSecret(ctx context.Context, req *DeleteSecretRequest) error {
	secret, err := s.db.GetSecretByID(ctx, req.SecretID)
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}
	if secret == nil {
		return fmt.Errorf("secret not found")
	}

	if secret.UserID != req.UserID {
		return fmt.Errorf("access denied")
	}

	if err := s.db.DeleteSecret(ctx, req.SecretID); err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		UserID:     req.UserID,
		Action:     "secret.deleted",
		ResourceID: req.SecretID.String(),
		Details: map[string]interface{}{
			"name": secret.Name,
		},
		CreatedAt: time.Now(),
	}
	_ = s.db.CreateAuditLog(ctx, auditLog)

	return nil
}

// GetVaultKey retrieves and decrypts the user's vault key
func (s *VaultService) GetVaultKey(ctx context.Context, userID uuid.UUID, password, secretKey string) ([]byte, error) {
	user, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	masterKey := crypto.DeriveMasterKey(password, secretKey, user.MasterKeySalt)
	defer crypto.ZeroBytes(masterKey)

	kek, err := crypto.DeriveKeyEncryptionKey(masterKey, "vault-key-wrap")
	if err != nil {
		return nil, fmt.Errorf("failed to derive key encryption key: %w", err)
	}
	defer crypto.ZeroBytes(kek)

	vaultKey, err := crypto.DecryptVaultKey(user.EncryptedVaultKey, kek, user.VaultKeyNonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt vault key: %w", err)
	}

	return vaultKey, nil
}
