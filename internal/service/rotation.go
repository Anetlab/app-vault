package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/securevault/app-vault/internal/crypto"
	"github.com/securevault/app-vault/internal/db"
	"github.com/securevault/app-vault/internal/models"
)

// RotationService handles key rotation operations
type RotationService struct {
	db *db.Database
}

// NewRotationService creates a new rotation service
func NewRotationService(database *db.Database) *RotationService {
	return &RotationService{
		db: database,
	}
}

// RotateVaultKeyRequest contains vault key rotation parameters
type RotateVaultKeyRequest struct {
	UserID    uuid.UUID
	Password  string
	SecretKey string
	Reason    string
}

// RotateVaultKeyResponse contains rotation result
type RotateVaultKeyResponse struct {
	NewKeyVersionID    uuid.UUID
	SecretsReEncrypted int
	Duration           time.Duration
}

// RotateVaultKey rotates the vault key and re-encrypts all secrets
func (s *RotationService) RotateVaultKey(ctx context.Context, req *RotateVaultKeyRequest) (*RotateVaultKeyResponse, error) {
	startTime := time.Now()

	user, err := s.db.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	masterKey := crypto.DeriveMasterKey(req.Password, req.SecretKey, user.MasterKeySalt)
	defer crypto.ZeroBytes(masterKey)

	kek, err := crypto.DeriveKeyEncryptionKey(masterKey, "vault-key-wrap")
	if err != nil {
		return nil, fmt.Errorf("failed to derive key encryption key: %w", err)
	}
	defer crypto.ZeroBytes(kek)

	oldVaultKey, err := crypto.DecryptVaultKey(user.EncryptedVaultKey, kek, user.VaultKeyNonce)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt current vault key: %w", err)
	}
	defer crypto.ZeroBytes(oldVaultKey)

	newVaultKey, err := crypto.GenerateVaultKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new vault key: %w", err)
	}
	defer crypto.ZeroBytes(newVaultKey)

	newVaultKeyNonce, err := crypto.GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	encryptedNewVaultKey, err := crypto.EncryptVaultKey(newVaultKey, kek, newVaultKeyNonce)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt new vault key: %w", err)
	}

	oldKeyVersion, err := s.db.GetActiveKeyVersion(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current key version: %w", err)
	}
	if oldKeyVersion == nil {
		return nil, fmt.Errorf("no active key version found")
	}

	newVersion := oldKeyVersion.Version + 1
	newKeyVersionID := uuid.New()

	newKeyVersion := &models.KeyVersion{
		ID:                newKeyVersionID,
		UserID:            req.UserID,
		Version:           newVersion,
		EncryptedVaultKey: encryptedNewVaultKey,
		VaultKeyNonce:     newVaultKeyNonce,
		Status:            models.KeyStatusActive,
		CreatedAt:         time.Now(),
		RotationReason:    req.Reason,
	}

	if err := s.db.CreateKeyVersion(ctx, newKeyVersion); err != nil {
		return nil, fmt.Errorf("failed to create new key version: %w", err)
	}

	secrets, err := s.db.GetSecretsByKeyVersion(ctx, oldKeyVersion.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets for re-encryption: %w", err)
	}

	reEncryptedCount := 0
	for _, secret := range secrets {
		decryptedData, err := crypto.DecryptData(secret.EncryptedData, oldVaultKey, secret.Nonce)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt secret %s: %w", secret.Name, err)
		}

		newNonce, err := crypto.GenerateNonce()
		if err != nil {
			crypto.ZeroBytes(decryptedData)
			return nil, fmt.Errorf("failed to generate nonce: %w", err)
		}

		reEncryptedData, err := crypto.EncryptData(decryptedData, newVaultKey, newNonce)
		crypto.ZeroBytes(decryptedData)
		if err != nil {
			return nil, fmt.Errorf("failed to re-encrypt secret %s: %w", secret.Name, err)
		}

		if err := s.db.UpdateSecretKeyVersion(ctx, secret.ID, newKeyVersionID, reEncryptedData, newNonce); err != nil {
			return nil, fmt.Errorf("failed to update secret %s: %w", secret.Name, err)
		}

		reEncryptedCount++
	}

	destroyAt := time.Now().Add(24 * time.Hour)
	if err := s.db.UpdateKeyVersionStatus(ctx, oldKeyVersion.ID, models.KeyStatusDeprecated, &destroyAt); err != nil {
		return nil, fmt.Errorf("failed to deprecate old key version: %w", err)
	}

	if err := s.db.UpdateUserKeys(ctx, req.UserID, encryptedNewVaultKey, newVaultKeyNonce, user.MasterKeySalt, &newKeyVersionID); err != nil {
		return nil, fmt.Errorf("failed to update user keys: %w", err)
	}

	duration := time.Since(startTime)

	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		UserID:     req.UserID,
		Action:     "vault_key.rotated",
		ResourceID: newKeyVersionID.String(),
		Details: map[string]interface{}{
			"old_version":         oldKeyVersion.Version,
			"new_version":         newVersion,
			"secrets_reencrypted": reEncryptedCount,
			"duration_ms":         duration.Milliseconds(),
			"reason":              req.Reason,
		},
		CreatedAt: time.Now(),
	}
	_ = s.db.CreateAuditLog(ctx, auditLog)

	return &RotateVaultKeyResponse{
		NewKeyVersionID:    newKeyVersionID,
		SecretsReEncrypted: reEncryptedCount,
		Duration:           duration,
	}, nil
}

// ChangeMasterKeyRequest contains master key change parameters
type ChangeMasterKeyRequest struct {
	UserID      uuid.UUID
	OldPassword string
	NewPassword string
	SecretKey   string
}

// ChangeMasterKeyResponse contains master key change result
type ChangeMasterKeyResponse struct {
	Success bool
}

// ChangeMasterKey changes the user's master key (password change)
func (s *RotationService) ChangeMasterKey(ctx context.Context, req *ChangeMasterKeyRequest) (*ChangeMasterKeyResponse, error) {
	if len(req.NewPassword) < 12 {
		return nil, fmt.Errorf("new password must be at least 12 characters")
	}

	user, err := s.db.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	oldMasterKey := crypto.DeriveMasterKey(req.OldPassword, req.SecretKey, user.MasterKeySalt)
	defer crypto.ZeroBytes(oldMasterKey)

	oldKek, err := crypto.DeriveKeyEncryptionKey(oldMasterKey, "vault-key-wrap")
	if err != nil {
		return nil, fmt.Errorf("failed to derive old key encryption key: %w", err)
	}
	defer crypto.ZeroBytes(oldKek)

	vaultKey, err := crypto.DecryptVaultKey(user.EncryptedVaultKey, oldKek, user.VaultKeyNonce)
	if err != nil {
		return nil, fmt.Errorf("failed to verify old password: invalid credentials")
	}
	defer crypto.ZeroBytes(vaultKey)

	newMasterKeySalt, err := crypto.GenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new master key salt: %w", err)
	}

	newMasterKey := crypto.DeriveMasterKey(req.NewPassword, req.SecretKey, newMasterKeySalt)
	defer crypto.ZeroBytes(newMasterKey)

	newKek, err := crypto.DeriveKeyEncryptionKey(newMasterKey, "vault-key-wrap")
	if err != nil {
		return nil, fmt.Errorf("failed to derive new key encryption key: %w", err)
	}
	defer crypto.ZeroBytes(newKek)

	newVaultKeyNonce, err := crypto.GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	newEncryptedVaultKey, err := crypto.EncryptVaultKey(vaultKey, newKek, newVaultKeyNonce)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt vault key with new master key: %w", err)
	}

	if err := s.db.UpdateUserKeys(ctx, req.UserID, newEncryptedVaultKey, newVaultKeyNonce, newMasterKeySalt, user.CurrentKeyVersionID); err != nil {
		return nil, fmt.Errorf("failed to update user keys: %w", err)
	}

	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		UserID:     req.UserID,
		Action:     "master_key.changed",
		ResourceID: req.UserID.String(),
		Details: map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
		CreatedAt: time.Now(),
	}
	_ = s.db.CreateAuditLog(ctx, auditLog)

	return &ChangeMasterKeyResponse{
		Success: true,
	}, nil
}

// GetKeyStatusRequest contains key status request parameters
type GetKeyStatusRequest struct {
	UserID uuid.UUID
}

// KeyVersionStatus represents key version status
type KeyVersionStatus struct {
	ID             uuid.UUID
	Version        int
	Status         string
	CreatedAt      time.Time
	DestroyAt      *time.Time
	RotationReason string
}

// GetKeyStatus returns the status of all key versions
func (s *RotationService) GetKeyStatus(ctx context.Context, req *GetKeyStatusRequest) ([]*KeyVersionStatus, error) {
	versions, err := s.db.ListKeyVersions(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to list key versions: %w", err)
	}

	statuses := make([]*KeyVersionStatus, len(versions))
	for i, version := range versions {
		statuses[i] = &KeyVersionStatus{
			ID:             version.ID,
			Version:        version.Version,
			Status:         version.Status,
			CreatedAt:      version.CreatedAt,
			DestroyAt:      version.DestroyAt,
			RotationReason: version.RotationReason,
		}
	}

	return statuses, nil
}
