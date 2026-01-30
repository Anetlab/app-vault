package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/app-vault/app-vault/internal/models"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// Database wraps the database connection
type Database struct {
	db *sql.DB
}

// New creates a new database connection
func New(connStr string) (*Database, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &Database{db: db}, nil
}

// RunMigrations executes all SQL migration files
func (d *Database) RunMigrations(migrationsPath string) error {
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to list migrations: %w", err)
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		if _, err := d.db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
	}

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// Ping checks the database connection
func (d *Database) Ping() error {
	return d.db.Ping()
}

// User operations

// CreateUser creates a new user in the database
func (d *Database) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, email, secret_key_hash, master_key_salt,
			encrypted_vault_key, vault_key_nonce, srp_verifier,
			current_key_version_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := d.db.ExecContext(ctx, query,
		user.ID, user.Email, user.SecretKeyHash, user.MasterKeySalt,
		user.EncryptedVaultKey, user.VaultKeyNonce, user.SRPVerifier,
		user.CurrentKeyVersionID, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetUserByEmail retrieves a user by email
func (d *Database) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, secret_key_hash, master_key_salt,
			encrypted_vault_key, vault_key_nonce, srp_verifier,
			current_key_version_id, created_at, updated_at
		FROM users WHERE email = $1`

	user := &models.User{}
	var currentKeyVersionID sql.NullString

	err := d.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.SecretKeyHash, &user.MasterKeySalt,
		&user.EncryptedVaultKey, &user.VaultKeyNonce, &user.SRPVerifier,
		&currentKeyVersionID, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if currentKeyVersionID.Valid {
		id, _ := uuid.Parse(currentKeyVersionID.String)
		user.CurrentKeyVersionID = &id
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (d *Database) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, secret_key_hash, master_key_salt,
			encrypted_vault_key, vault_key_nonce, srp_verifier,
			current_key_version_id, created_at, updated_at
		FROM users WHERE id = $1`

	user := &models.User{}
	var currentKeyVersionID sql.NullString

	err := d.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.SecretKeyHash, &user.MasterKeySalt,
		&user.EncryptedVaultKey, &user.VaultKeyNonce, &user.SRPVerifier,
		&currentKeyVersionID, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	if currentKeyVersionID.Valid {
		id, _ := uuid.Parse(currentKeyVersionID.String)
		user.CurrentKeyVersionID = &id
	}

	return user, nil
}

// UpdateUserKeys updates user's encryption keys
func (d *Database) UpdateUserKeys(ctx context.Context, userID uuid.UUID, encryptedVaultKey, vaultKeyNonce, masterKeySalt []byte, currentKeyVersionID *uuid.UUID) error {
	query := `
		UPDATE users
		SET encrypted_vault_key = $1, vault_key_nonce = $2,
			master_key_salt = $3, current_key_version_id = $4, updated_at = $5
		WHERE id = $6`

	_, err := d.db.ExecContext(ctx, query,
		encryptedVaultKey, vaultKeyNonce, masterKeySalt,
		currentKeyVersionID, time.Now(), userID)

	if err != nil {
		return fmt.Errorf("failed to update user keys: %w", err)
	}
	return nil
}

// Secret operations

// CreateSecret creates a new secret
func (d *Database) CreateSecret(ctx context.Context, secret *models.Secret) error {
	query := `
		INSERT INTO secrets (
			id, user_id, key_version_id, name, encrypted_data, nonce,
			secret_type, version, previous_id, tags, created_at, updated_at,
			expires_at, access_count, last_accessed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (user_id, name) DO UPDATE SET
			encrypted_data = EXCLUDED.encrypted_data,
			nonce = EXCLUDED.nonce,
			key_version_id = EXCLUDED.key_version_id,
			secret_type = EXCLUDED.secret_type,
			version = secrets.version + 1,
			previous_id = secrets.id,
			tags = EXCLUDED.tags,
			updated_at = EXCLUDED.updated_at,
			expires_at = EXCLUDED.expires_at`

	_, err := d.db.ExecContext(ctx, query,
		secret.ID, secret.UserID, secret.KeyVersionID, secret.Name,
		secret.EncryptedData, secret.Nonce, secret.SecretType,
		secret.Version, secret.PreviousID, pq.Array(secret.Tags),
		secret.CreatedAt, secret.UpdatedAt, secret.ExpiresAt,
		secret.AccessCount, secret.LastAccessedAt)

	if err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}
	return nil
}

// GetSecretByName retrieves a secret by name
func (d *Database) GetSecretByName(ctx context.Context, userID uuid.UUID, name string) (*models.Secret, error) {
	query := `
		SELECT id, user_id, key_version_id, name, encrypted_data, nonce,
			secret_type, version, previous_id, tags, created_at, updated_at,
			expires_at, access_count, last_accessed_at
		FROM secrets
		WHERE user_id = $1 AND name = $2`

	secret := &models.Secret{}
	var previousID sql.NullString
	var expiresAt sql.NullTime
	var lastAccessedAt sql.NullTime

	err := d.db.QueryRowContext(ctx, query, userID, name).Scan(
		&secret.ID, &secret.UserID, &secret.KeyVersionID, &secret.Name,
		&secret.EncryptedData, &secret.Nonce, &secret.SecretType,
		&secret.Version, &previousID, pq.Array(&secret.Tags),
		&secret.CreatedAt, &secret.UpdatedAt, &expiresAt,
		&secret.AccessCount, &lastAccessedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get secret by name: %w", err)
	}

	if previousID.Valid {
		id, _ := uuid.Parse(previousID.String)
		secret.PreviousID = &id
	}
	if expiresAt.Valid {
		secret.ExpiresAt = &expiresAt.Time
	}
	if lastAccessedAt.Valid {
		secret.LastAccessedAt = &lastAccessedAt.Time
	}

	return secret, nil
}

// GetSecretByID retrieves a secret by ID
func (d *Database) GetSecretByID(ctx context.Context, id uuid.UUID) (*models.Secret, error) {
	query := `
		SELECT id, user_id, key_version_id, name, encrypted_data, nonce,
			secret_type, version, previous_id, tags, created_at, updated_at,
			expires_at, access_count, last_accessed_at
		FROM secrets
		WHERE id = $1`

	secret := &models.Secret{}
	var previousID sql.NullString
	var expiresAt sql.NullTime
	var lastAccessedAt sql.NullTime

	err := d.db.QueryRowContext(ctx, query, id).Scan(
		&secret.ID, &secret.UserID, &secret.KeyVersionID, &secret.Name,
		&secret.EncryptedData, &secret.Nonce, &secret.SecretType,
		&secret.Version, &previousID, pq.Array(&secret.Tags),
		&secret.CreatedAt, &secret.UpdatedAt, &expiresAt,
		&secret.AccessCount, &lastAccessedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get secret by ID: %w", err)
	}

	if previousID.Valid {
		id, _ := uuid.Parse(previousID.String)
		secret.PreviousID = &id
	}
	if expiresAt.Valid {
		secret.ExpiresAt = &expiresAt.Time
	}
	if lastAccessedAt.Valid {
		secret.LastAccessedAt = &lastAccessedAt.Time
	}

	return secret, nil
}

// ListSecrets lists all secrets for a user
func (d *Database) ListSecrets(ctx context.Context, userID uuid.UUID) ([]*models.Secret, error) {
	query := `
		SELECT id, user_id, key_version_id, name, encrypted_data, nonce,
			secret_type, version, previous_id, tags, created_at, updated_at,
			expires_at, access_count, last_accessed_at
		FROM secrets
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := d.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}
	defer rows.Close()

	var secrets []*models.Secret
	for rows.Next() {
		secret := &models.Secret{}
		var previousID sql.NullString
		var expiresAt sql.NullTime
		var lastAccessedAt sql.NullTime

		err := rows.Scan(
			&secret.ID, &secret.UserID, &secret.KeyVersionID, &secret.Name,
			&secret.EncryptedData, &secret.Nonce, &secret.SecretType,
			&secret.Version, &previousID, pq.Array(&secret.Tags),
			&secret.CreatedAt, &secret.UpdatedAt, &expiresAt,
			&secret.AccessCount, &lastAccessedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan secret: %w", err)
		}

		if previousID.Valid {
			id, _ := uuid.Parse(previousID.String)
			secret.PreviousID = &id
		}
		if expiresAt.Valid {
			secret.ExpiresAt = &expiresAt.Time
		}
		if lastAccessedAt.Valid {
			secret.LastAccessedAt = &lastAccessedAt.Time
		}

		secrets = append(secrets, secret)
	}

	return secrets, rows.Err()
}

// DeleteSecret deletes a secret
func (d *Database) DeleteSecret(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM secrets WHERE id = $1`
	result, err := d.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("secret not found")
	}

	return nil
}

// UpdateSecretAccess updates access count and timestamp
func (d *Database) UpdateSecretAccess(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE secrets
		SET access_count = access_count + 1, last_accessed_at = $1
		WHERE id = $2`

	_, err := d.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update secret access: %w", err)
	}
	return nil
}

// GetSecretsByKeyVersion retrieves all secrets encrypted with a specific key version
func (d *Database) GetSecretsByKeyVersion(ctx context.Context, keyVersionID uuid.UUID) ([]*models.Secret, error) {
	query := `
		SELECT id, user_id, key_version_id, name, encrypted_data, nonce,
			secret_type, version, previous_id, tags, created_at, updated_at,
			expires_at, access_count, last_accessed_at
		FROM secrets
		WHERE key_version_id = $1`

	rows, err := d.db.QueryContext(ctx, query, keyVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets by key version: %w", err)
	}
	defer rows.Close()

	var secrets []*models.Secret
	for rows.Next() {
		secret := &models.Secret{}
		var previousID sql.NullString
		var expiresAt sql.NullTime
		var lastAccessedAt sql.NullTime

		err := rows.Scan(
			&secret.ID, &secret.UserID, &secret.KeyVersionID, &secret.Name,
			&secret.EncryptedData, &secret.Nonce, &secret.SecretType,
			&secret.Version, &previousID, pq.Array(&secret.Tags),
			&secret.CreatedAt, &secret.UpdatedAt, &expiresAt,
			&secret.AccessCount, &lastAccessedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan secret: %w", err)
		}

		if previousID.Valid {
			id, _ := uuid.Parse(previousID.String)
			secret.PreviousID = &id
		}
		if expiresAt.Valid {
			secret.ExpiresAt = &expiresAt.Time
		}
		if lastAccessedAt.Valid {
			secret.LastAccessedAt = &lastAccessedAt.Time
		}

		secrets = append(secrets, secret)
	}

	return secrets, rows.Err()
}

// UpdateSecretKeyVersion updates the key version for a secret
func (d *Database) UpdateSecretKeyVersion(ctx context.Context, secretID, newKeyVersionID uuid.UUID, encryptedData, nonce []byte) error {
	query := `
		UPDATE secrets
		SET key_version_id = $1, encrypted_data = $2, nonce = $3, updated_at = $4
		WHERE id = $5`

	_, err := d.db.ExecContext(ctx, query, newKeyVersionID, encryptedData, nonce, time.Now(), secretID)
	if err != nil {
		return fmt.Errorf("failed to update secret key version: %w", err)
	}
	return nil
}

// ServicePrincipal operations

// CreateServicePrincipal creates a new service principal
func (d *Database) CreateServicePrincipal(ctx context.Context, sp *models.ServicePrincipal) error {
	query := `
		INSERT INTO service_principals (
			id, user_id, client_id, client_secret_hash, name, description,
			permissions, allowed_secrets, allowed_tags, ip_whitelist,
			rate_limit, is_active, expires_at, last_used_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`

	_, err := d.db.ExecContext(ctx, query,
		sp.ID, sp.UserID, sp.ClientID, sp.ClientSecretHash, sp.Name, sp.Description,
		sp.Permissions, sp.AllowedSecrets, sp.AllowedTags, sp.IPWhitelist,
		sp.RateLimit, sp.IsActive, sp.ExpiresAt, sp.LastUsedAt,
		sp.CreatedAt, sp.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create service principal: %w", err)
	}
	return nil
}

// GetServicePrincipalByClientID retrieves a service principal by client ID
func (d *Database) GetServicePrincipalByClientID(ctx context.Context, clientID string) (*models.ServicePrincipal, error) {
	query := `
		SELECT id, user_id, client_id, client_secret_hash, name, description,
			permissions, allowed_secrets, allowed_tags, ip_whitelist,
			rate_limit, is_active, expires_at, last_used_at,
			created_at, updated_at
		FROM service_principals
		WHERE client_id = $1`

	sp := &models.ServicePrincipal{}
	var expiresAt sql.NullTime
	var lastUsedAt sql.NullTime

	err := d.db.QueryRowContext(ctx, query, clientID).Scan(
		&sp.ID, &sp.UserID, &sp.ClientID, &sp.ClientSecretHash, &sp.Name, &sp.Description,
		&sp.Permissions, &sp.AllowedSecrets, &sp.AllowedTags, &sp.IPWhitelist,
		&sp.RateLimit, &sp.IsActive, &expiresAt, &lastUsedAt,
		&sp.CreatedAt, &sp.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get service principal: %w", err)
	}

	if expiresAt.Valid {
		sp.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		sp.LastUsedAt = &lastUsedAt.Time
	}

	return sp, nil
}

// UpdateServicePrincipalLastUsed updates the last used timestamp
func (d *Database) UpdateServicePrincipalLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE service_principals SET last_used_at = $1 WHERE id = $2`
	_, err := d.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update service principal last used: %w", err)
	}
	return nil
}

// ListServicePrincipalsByUserID retrieves all service principals for a user
func (d *Database) ListServicePrincipalsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.ServicePrincipal, error) {
	query := `
		SELECT id, user_id, client_id, client_secret_hash, name, description,
			permissions, allowed_secrets, allowed_tags, ip_whitelist,
			rate_limit, is_active, expires_at, last_used_at,
			created_at, updated_at
		FROM service_principals
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := d.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list service principals: %w", err)
	}
	defer rows.Close()

	var servicePrincipals []*models.ServicePrincipal
	for rows.Next() {
		sp := &models.ServicePrincipal{}
		var expiresAt sql.NullTime
		var lastUsedAt sql.NullTime

		err := rows.Scan(
			&sp.ID, &sp.UserID, &sp.ClientID, &sp.ClientSecretHash, &sp.Name, &sp.Description,
			&sp.Permissions, &sp.AllowedSecrets, &sp.AllowedTags, &sp.IPWhitelist,
			&sp.RateLimit, &sp.IsActive, &expiresAt, &lastUsedAt,
			&sp.CreatedAt, &sp.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan service principal: %w", err)
		}

		if expiresAt.Valid {
			sp.ExpiresAt = &expiresAt.Time
		}
		if lastUsedAt.Valid {
			sp.LastUsedAt = &lastUsedAt.Time
		}

		servicePrincipals = append(servicePrincipals, sp)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating service principals: %w", err)
	}

	return servicePrincipals, nil
}

// GetServicePrincipalByID retrieves a service principal by ID
func (d *Database) GetServicePrincipalByID(ctx context.Context, id uuid.UUID) (*models.ServicePrincipal, error) {
	query := `
		SELECT id, user_id, client_id, client_secret_hash, name, description,
			permissions, allowed_secrets, allowed_tags, ip_whitelist,
			rate_limit, is_active, expires_at, last_used_at,
			created_at, updated_at
		FROM service_principals
		WHERE id = $1`

	sp := &models.ServicePrincipal{}
	var expiresAt sql.NullTime
	var lastUsedAt sql.NullTime

	err := d.db.QueryRowContext(ctx, query, id).Scan(
		&sp.ID, &sp.UserID, &sp.ClientID, &sp.ClientSecretHash, &sp.Name, &sp.Description,
		&sp.Permissions, &sp.AllowedSecrets, &sp.AllowedTags, &sp.IPWhitelist,
		&sp.RateLimit, &sp.IsActive, &expiresAt, &lastUsedAt,
		&sp.CreatedAt, &sp.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("service principal not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get service principal: %w", err)
	}

	if expiresAt.Valid {
		sp.ExpiresAt = &expiresAt.Time
	}
	if lastUsedAt.Valid {
		sp.LastUsedAt = &lastUsedAt.Time
	}

	return sp, nil
}

// DeleteServicePrincipal deletes a service principal
func (d *Database) DeleteServicePrincipal(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM service_principals WHERE id = $1`
	result, err := d.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete service principal: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("service principal not found")
	}

	return nil
}

// UpdateServicePrincipalSecret updates the client secret hash for a service principal
func (d *Database) UpdateServicePrincipalSecret(ctx context.Context, id uuid.UUID, clientSecretHash []byte) error {
	query := `UPDATE service_principals SET client_secret_hash = $1, updated_at = $2 WHERE id = $3`
	result, err := d.db.ExecContext(ctx, query, clientSecretHash, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update service principal secret: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("service principal not found")
	}

	return nil
}

// KeyVersion operations

// CreateKeyVersion creates a new key version
func (d *Database) CreateKeyVersion(ctx context.Context, kv *models.KeyVersion) error {
	query := `
		INSERT INTO key_versions (
			id, user_id, version, encrypted_vault_key, vault_key_nonce,
			status, created_at, rotated_at, destroy_at, destroyed_at, rotation_reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := d.db.ExecContext(ctx, query,
		kv.ID, kv.UserID, kv.Version, kv.EncryptedVaultKey, kv.VaultKeyNonce,
		kv.Status, kv.CreatedAt, kv.RotatedAt, kv.DestroyAt, kv.DestroyedAt, kv.RotationReason)

	if err != nil {
		return fmt.Errorf("failed to create key version: %w", err)
	}
	return nil
}

// GetActiveKeyVersion retrieves the active key version for a user
func (d *Database) GetActiveKeyVersion(ctx context.Context, userID uuid.UUID) (*models.KeyVersion, error) {
	query := `
		SELECT id, user_id, version, encrypted_vault_key, vault_key_nonce,
			status, created_at, rotated_at, destroy_at, destroyed_at, rotation_reason
		FROM key_versions
		WHERE user_id = $1 AND status = $2
		ORDER BY version DESC
		LIMIT 1`

	kv := &models.KeyVersion{}
	var rotatedAt, destroyAt, destroyedAt sql.NullTime

	err := d.db.QueryRowContext(ctx, query, userID, models.KeyStatusActive).Scan(
		&kv.ID, &kv.UserID, &kv.Version, &kv.EncryptedVaultKey, &kv.VaultKeyNonce,
		&kv.Status, &kv.CreatedAt, &rotatedAt, &destroyAt, &destroyedAt, &kv.RotationReason)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active key version: %w", err)
	}

	if rotatedAt.Valid {
		kv.RotatedAt = &rotatedAt.Time
	}
	if destroyAt.Valid {
		kv.DestroyAt = &destroyAt.Time
	}
	if destroyedAt.Valid {
		kv.DestroyedAt = &destroyedAt.Time
	}

	return kv, nil
}

// GetKeyVersionByID retrieves a key version by ID
func (d *Database) GetKeyVersionByID(ctx context.Context, id uuid.UUID) (*models.KeyVersion, error) {
	query := `
		SELECT id, user_id, version, encrypted_vault_key, vault_key_nonce,
			status, created_at, rotated_at, destroy_at, destroyed_at, rotation_reason
		FROM key_versions
		WHERE id = $1`

	kv := &models.KeyVersion{}
	var rotatedAt, destroyAt, destroyedAt sql.NullTime

	err := d.db.QueryRowContext(ctx, query, id).Scan(
		&kv.ID, &kv.UserID, &kv.Version, &kv.EncryptedVaultKey, &kv.VaultKeyNonce,
		&kv.Status, &kv.CreatedAt, &rotatedAt, &destroyAt, &destroyedAt, &kv.RotationReason)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get key version: %w", err)
	}

	if rotatedAt.Valid {
		kv.RotatedAt = &rotatedAt.Time
	}
	if destroyAt.Valid {
		kv.DestroyAt = &destroyAt.Time
	}
	if destroyedAt.Valid {
		kv.DestroyedAt = &destroyedAt.Time
	}

	return kv, nil
}

// UpdateKeyVersionStatus updates the status of a key version
func (d *Database) UpdateKeyVersionStatus(ctx context.Context, id uuid.UUID, status string, destroyAt *time.Time) error {
	query := `UPDATE key_versions SET status = $1, destroy_at = $2 WHERE id = $3`
	_, err := d.db.ExecContext(ctx, query, status, destroyAt, id)
	if err != nil {
		return fmt.Errorf("failed to update key version status: %w", err)
	}
	return nil
}

// ListKeyVersions lists all key versions for a user
func (d *Database) ListKeyVersions(ctx context.Context, userID uuid.UUID) ([]*models.KeyVersion, error) {
	query := `
		SELECT id, user_id, version, encrypted_vault_key, vault_key_nonce,
			status, created_at, rotated_at, destroy_at, destroyed_at, rotation_reason
		FROM key_versions
		WHERE user_id = $1
		ORDER BY version DESC`

	rows, err := d.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list key versions: %w", err)
	}
	defer rows.Close()

	var versions []*models.KeyVersion
	for rows.Next() {
		kv := &models.KeyVersion{}
		var rotatedAt, destroyAt, destroyedAt sql.NullTime

		err := rows.Scan(
			&kv.ID, &kv.UserID, &kv.Version, &kv.EncryptedVaultKey, &kv.VaultKeyNonce,
			&kv.Status, &kv.CreatedAt, &rotatedAt, &destroyAt, &destroyedAt, &kv.RotationReason)

		if err != nil {
			return nil, fmt.Errorf("failed to scan key version: %w", err)
		}

		if rotatedAt.Valid {
			kv.RotatedAt = &rotatedAt.Time
		}
		if destroyAt.Valid {
			kv.DestroyAt = &destroyAt.Time
		}
		if destroyedAt.Valid {
			kv.DestroyedAt = &destroyedAt.Time
		}

		versions = append(versions, kv)
	}

	return versions, rows.Err()
}

// AuditLog operations

// CreateAuditLog creates a new audit log entry
func (d *Database) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			id, user_id, service_principal_id, action, resource_id,
			details, ip_address, user_agent, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	var details []byte
	var err error
	if log.Details != nil {
		details, err = json.Marshal(log.Details)
		if err != nil {
			return fmt.Errorf("failed to marshal details: %w", err)
		}
	}

	_, err = d.db.ExecContext(ctx, query,
		log.ID, log.UserID, log.ServicePrincipalID, log.Action, log.ResourceID,
		details, log.IPAddress, log.UserAgent, log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

// BeginTx begins a transaction
func (d *Database) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return d.db.BeginTx(ctx, nil)
}
